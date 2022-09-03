package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// Register
func (server *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	arg, err := server.parseRegisterRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := server.store.CreateUser(ctx, *arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", arg.Username)
			case "users_email_key":
				return nil, status.Errorf(codes.AlreadyExists, "email already exists: %s", arg.Email)
			}
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	rsp := &pb.RegisterResponse{User: convertUser(user)}
	return rsp, nil
}

func (server *Server) parseRegisterRequest(req *pb.RegisterRequest) (*sqlc.CreateUserParams, error) {
	username := req.GetUsername()
	if err := util.ValidateString(username, 3, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "username: %s", err.Error())
	}

	email := req.GetEmail()
	if err := util.ValidateEmail(email); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "email: %s", err.Error())
	}

	password := req.GetPassword()
	if err := util.ValidateString(password, 6, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "password: %s", err.Error())
	}
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	params := &sqlc.CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: hashedPassword,
		Avatar:         server.config.AvatarPath + "/default",
		Role:           "user",
	}
	return params, nil
}

// -------------------------------------------------------------------
// Login
func (server *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	err := validateLoginRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var user sqlc.User

	switch req.GetPayload().(type) {
	case *(pb.LoginRequest_Username):
		user, err = server.store.GetUserByUsername(ctx, req.GetUsername())
	case *(pb.LoginRequest_Email):
		user, err = server.store.GetUserByEmail(ctx, req.GetEmail())
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if user.IsDeleted {
		return nil, status.Error(codes.NotFound, "this user is inactive")
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return nil, status.Error(codes.NotFound, "incorrect password")
	}

	unreadCount, err := server.store.GetUnreadCount(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get unread count")
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

	mtdt := server.extractMetadata(ctx)
	_, err = server.store.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       refreshPayload.UserID,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIp,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create session")
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

// -------------------------------------------------------------------
// AutoLogin
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

// -------------------------------------------------------------------
// Refresh
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

	rsp := &pb.RefreshTokenResponse{
		AccessToken: accessToken,
	}
	return rsp, nil
}

// -------------------------------------------------------------------
// Logout
func (server *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	refreshPayload, _, err := server.checkRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	arg := sqlc.DeleteSessionParams{
		ID:     refreshPayload.ID,
		UserID: refreshPayload.UserID,
	}
	err = server.store.DeleteSession(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete session")
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// Utils
// -------------------------------------------------------------------

// Check refresh token
func (server *Server) checkRefreshToken(ctx context.Context, refreshToken string) (*util.Payload, *sqlc.User, error) {
	if refreshToken == "" {
		return nil, nil, fmt.Errorf("refreshToken: must be a non empty string")
	}

	refreshPayload, err := server.tokenMaker.VerifyToker(refreshToken)
	if err != nil {
		return nil, nil, err
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
			return nil, nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, nil, status.Error(codes.Internal, "failed to get user")
	}

	if user.IsDeleted {
		return nil, nil, status.Error(codes.NotFound, "this user is inactive")
	}

	return refreshPayload, &user, nil
}
