package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ========================// Register //======================== //

func (server *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		Email:          req.GetEmail(),
		HashedPassword: hashedPassword,
		Avatar:         server.config.DefaultAvatar,
		Role:           "user",
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username or email already exists: %s", err.Error())
			}
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	rsp := &pb.RegisterResponse{User: convertUser(user)}
	return rsp, nil
}

func validateRegisterRequest(req *pb.RegisterRequest) error {
	if err := util.ValidateString(req.GetUsername(), 3, 50); err != nil {
		return fmt.Errorf("username: %s", err.Error())
	}
	if err := util.ValidateEmail(req.GetEmail()); err != nil {
		return fmt.Errorf("email: %s", err.Error())
	}
	if err := util.ValidateString(req.GetPassword(), 6, 50); err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}
	return nil
}

// ========================// Login //======================== //

func (server *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var user db.User
	var err error
	switch req.GetPayload().(type) {
	case *(pb.LoginRequest_Username):
		user, err = server.store.GetUserByUsername(ctx, req.GetUsername())
	case *(pb.LoginRequest_Email):
		user, err = server.store.GetUserByEmail(ctx, req.GetEmail())
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if user.Deleted {
		return nil, status.Error(codes.NotFound, "this user is inactive")
	}

	if err = util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		return nil, status.Error(codes.NotFound, "incorrect password")
	}

	accessToken, _, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create access token")
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.ID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create refresh token")
	}

	UserAgent, ClientIp := extractLoginInfo(ctx)
	arg := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       refreshPayload.UserID,
		RefreshToken: refreshToken,
		UserAgent:    UserAgent,
		ClientIp:     ClientIp,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}
	if _, err = server.store.CreateSession(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	unreadCount, err := server.store.GetUnreadCount(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get unread count")
	}

	rsp := &pb.LoginResponse{
		User:         convertUser(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UnreadCount:  unreadCount,
	}
	return rsp, nil
}

func validateLoginRequest(req *pb.LoginRequest) error {
	switch req.GetPayload().(type) {
	case *(pb.LoginRequest_Username):
		if err := util.ValidateString(req.GetUsername(), 3, 50); err != nil {
			return fmt.Errorf("username: %s", err.Error())
		}
	case *(pb.LoginRequest_Email):
		if err := util.ValidateEmail(req.GetEmail()); err != nil {
			return fmt.Errorf("email: %s", err.Error())
		}
	default:
		return fmt.Errorf("username or email is not provided")
	}
	if err := util.ValidateString(req.GetPassword(), 6, 50); err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}
	return nil
}

func extractLoginInfo(ctx context.Context) (string, string) {
	var UserAgent, ClientIp string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get("grpcgateway-user-agent"); len(userAgents) > 0 {
			UserAgent = userAgents[0]
		}
		if clientIPs := md.Get("x-forwarded-for"); len(clientIPs) > 0 {
			ClientIp = clientIPs[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		ClientIp = p.Addr.String()
	}
	return UserAgent, ClientIp
}

// ========================// AutoLogin //======================== //

func (server *Server) AutoLogin(ctx context.Context, req *pb.AutoLoginRequest) (*pb.AutoLoginResponse, error) {
	refreshPayload, user, err := server.checkRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	unreadCount, err := server.store.GetUnreadCount(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get unread count")
	}

	accessToken, _, err := server.tokenMaker.CreateToken(
		refreshPayload.UserID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &pb.AutoLoginResponse{
		User:        convertUser(*user),
		AccessToken: accessToken,
		UnreadCount: unreadCount,
	}
	return rsp, nil
}

// ========================// Refresh //======================== //

func (server *Server) Refresh(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	refreshPayload, _, err := server.checkRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accessToken, _, err := server.tokenMaker.CreateToken(
		refreshPayload.UserID,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create access token")
	}

	rsp := &pb.RefreshTokenResponse{AccessToken: accessToken}
	return rsp, nil
}

// ========================// Logout //======================== //

func (server *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	refreshPayload, _, err := server.checkRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	arg := db.DeleteSessionParams{
		ID:     refreshPayload.ID,
		UserID: refreshPayload.UserID,
	}
	if err = server.store.DeleteSession(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete session")
	}
	return &emptypb.Empty{}, nil
}

// ========================// UTILS //======================== //

// checkRefreshToken
func (server *Server) checkRefreshToken(ctx context.Context, refreshToken string) (*util.Payload, *db.User, error) {
	if refreshToken == "" {
		return nil, nil, fmt.Errorf("refreshToken: must be a non empty string")
	}

	refreshPayload, err := server.tokenMaker.VerifyToker(refreshToken)
	if err != nil {
		return nil, nil, err
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("session not exists")
		}
		return nil, nil, fmt.Errorf("failed to get session")
	}
	if session.UserID != refreshPayload.UserID {
		return nil, nil, fmt.Errorf("mismatched session user")
	}
	if session.RefreshToken != refreshToken {
		return nil, nil, fmt.Errorf("mismatched session token")
	}

	user, err := server.store.GetUser(ctx, refreshPayload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("user not found")
		}
		return nil, nil, fmt.Errorf("failed to get user")
	}

	if user.Deleted {
		return nil, nil, fmt.Errorf("this user is inactive")
	}
	return refreshPayload, &user, nil
}
