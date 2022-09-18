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
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// CreateUser
func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*emptypb.Empty, error) {
	if err := validateCreateUserRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		Email:          req.GetEmail(),
		Role:           req.GetRole(),
		HashedPassword: hashedPassword,
		Avatar:         server.config.AvatarPath + "/default",
	}

	_, err = server.store.CreateUser(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username or email already exists: %s", err.Error())
			}
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &emptypb.Empty{}, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) error {
	if err := util.ValidateString(req.GetUsername(), 3, 50); err != nil {
		return fmt.Errorf("username: %s", err.Error())
	}

	if err := util.ValidateEmail(req.GetEmail()); err != nil {
		return fmt.Errorf("email: %s", err.Error())
	}

	if err := util.ValidateString(req.GetPassword(), 6, 50); err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}

	if err := util.ValidateOneOf(req.GetRole(), []string{"admin", "author", "user"}); err != nil {
		return fmt.Errorf("role: %s", err.Error())
	}

	return nil
}

// -------------------------------------------------------------------
// DeleteUsers
func (server *Server) DeleteUsers(ctx context.Context, req *pb.DeleteUsersRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	userIDs, err := util.ValidateRepeatedIDs(req.GetUserIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}

	for _, userID := range userIDs {
		if userID == authUser.ID {
			return nil, status.Errorf(codes.InvalidArgument, "cannot delete yourself")
		}
	}

	nrows, err := server.store.DeleteUsers(ctx, userIDs)
	if err != nil || int64(len(userIDs)) != nrows {
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// UpdateUser
func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	if err := validateUpdateUserRequest(req); err != nil {
		return nil, err
	}

	arg := db.UpdateUserParams{
		ID:       req.GetUserId(),
		Username: sql.NullString{String: req.GetUsername(), Valid: req.Username != nil},
		Email:    sql.NullString{String: req.GetEmail(), Valid: req.Email != nil},
		Role:     sql.NullString{String: req.GetRole(), Valid: req.Role != nil},
	}

	if req.Password != nil {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		arg.HashedPassword = sql.NullString{String: hashedPassword, Valid: true}
	}

	if _, err := server.store.UpdateUser(ctx, arg); err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_voilation":
				return nil, status.Errorf(codes.AlreadyExists, "username or email already exists: %s", err.Error())
			}
		}
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &emptypb.Empty{}, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) error {
	if err := util.ValidateID(req.GetUserId()); err != nil {
		return fmt.Errorf("userId: %s", err.Error())
	}

	if req.Username != nil {
		if err := util.ValidateString(req.GetUsername(), 3, 50); err != nil {
			return fmt.Errorf("username: %s", err.Error())
		}
	}

	if req.Email != nil {
		if err := util.ValidateEmail(req.GetEmail()); err != nil {
			return fmt.Errorf("email: %s", err.Error())
		}
	}

	if req.Password != nil {
		if err := util.ValidateString(req.GetPassword(), 6, 50); err != nil {
			return fmt.Errorf("password: %s", err.Error())
		}
	}

	if req.Role != nil {
		if err := util.ValidateOneOf(req.GetRole(), []string{"admin", "author", "user"}); err != nil {
			return fmt.Errorf("role: %s", err.Error())
		}
	}

	return nil
}

// -------------------------------------------------------------------
// ListUsers
func (server *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	options := []string{"username", "role", "deleted", "createAt"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Keyword != nil {
		if err := util.ValidateString(req.GetKeyword(), 1, 50); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "keyword: %s", err.Error())
		}
	}

	arg := db.ListUsersParams{
		Limit:        req.GetPageSize(),
		Offset:       (req.GetPageId() - 1) * req.GetPageSize(),
		UsernameAsc:  req.GetOrderBy() == "username" && req.GetOrder() == "asc",
		UsernameDesc: req.GetOrderBy() == "username" && req.GetOrder() == "desc",
		RoleAsc:      req.GetOrderBy() == "role" && req.GetOrder() == "asc",
		RoleDesc:     req.GetOrderBy() == "role" && req.GetOrder() == "desc",
		DeletedAsc:   req.GetOrderBy() == "deleted" && req.GetOrder() == "asc",
		DeletedDesc:  req.GetOrderBy() == "deleted" && req.GetOrder() == "desc",
		CreateAtAsc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc: req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
		AnyKeyword:   req.Keyword == nil,
		Keyword:      "%" + req.GetKeyword() + "%",
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	return convertListUsers(users), nil
}

// -------------------------------------------------------------------
// ChangeProfile
func (server *Server) ChangeProfile(ctx context.Context, req *pb.ChangeProfileRequest) (*pb.ChangeProfileResponse, error) {
	if err := validateChangeProfileRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}
	if authUser.ID != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "no permission to change this user")
	}

	arg := db.UpdateUserParams{
		ID:       req.GetUserId(),
		Username: sql.NullString{String: req.GetUsername(), Valid: req.Username != nil},
		Email:    sql.NullString{String: req.GetEmail(), Valid: req.Email != nil},
		Intro:    sql.NullString{String: req.GetIntro(), Valid: req.Intro != nil},
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	rsp := &pb.ChangeProfileResponse{User: convertUser(user)}
	return rsp, nil
}

func validateChangeProfileRequest(req *pb.ChangeProfileRequest) error {
	if err := util.ValidateID(req.GetUserId()); err != nil {
		return fmt.Errorf("userId: %s", err.Error())
	}

	if req.Username != nil {
		if err := util.ValidateString(req.GetUsername(), 3, 50); err != nil {
			return fmt.Errorf("username: %s", err.Error())
		}
	}

	if req.Email != nil {
		if err := util.ValidateEmail(req.GetEmail()); err != nil {
			return fmt.Errorf("email: %s", err.Error())
		}
	}

	if req.Intro != nil {
		if err := util.ValidateString(req.GetIntro(), 1, 150); err != nil {
			return fmt.Errorf("intro: %s", err.Error())
		}
	}

	return nil
}

// -------------------------------------------------------------------
// ChangePassword
func (server *Server) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := validateChangePasswordRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	if authUser.ID != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "no permission to change this user's password")
	}

	if err := util.CheckPassword(req.GetOldPassword(), authUser.HashedPassword); err != nil {
		return nil, status.Error(codes.NotFound, "incorrect old password")
	}

	hashedPassword, err := util.HashPassword(req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	arg := db.UpdateUserParams{
		ID:             req.GetUserId(),
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
	}

	if _, err = server.store.UpdateUser(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &emptypb.Empty{}, nil
}

func validateChangePasswordRequest(req *pb.ChangePasswordRequest) error {
	if err := util.ValidateID(req.GetUserId()); err != nil {
		return fmt.Errorf("userId: %s", err.Error())
	}

	if err := util.ValidateString(req.GetOldPassword(), 6, 50); err != nil {
		return fmt.Errorf("oldPassword: %s", err.Error())
	}

	if err := util.ValidateString(req.GetNewPassword(), 6, 50); err != nil {
		return fmt.Errorf("newPassword: %s", err.Error())
	}

	return nil
}

// -------------------------------------------------------------------
// GetUserProfile
func (server *Server) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	var authUser AuthUser
	if user, ok := ctx.Value(authUserKey{}).(AuthUser); ok {
		authUser = user
	}

	if err := util.ValidateID(req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}

	arg := db.GetUserProfileParams{
		UserID: req.GetUserId(),
		SelfID: authUser.ID,
	}

	user, err := server.store.GetUserProfile(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	rsp := &pb.GetUserProfileResponse{
		User: &pb.GetUserProfileResponse_User{
			Id:             user.ID,
			Username:       user.Username,
			Avatar:         user.Avatar,
			Intro:          user.Intro,
			StarCount:      user.StarCount,
			ViewCount:      user.ViewCount,
			FollowerCount:  user.FollowerCount,
			FollowingCount: user.FollowingCount,
			Followed:       user.Followed.Valid,
		},
	}
	return rsp, nil
}
