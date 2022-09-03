package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"
	"database/sql"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// CreateUser
func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*emptypb.Empty, error) {
	arg, err := server.parseCreateUserRequest(req)
	if err != nil {
		return nil, err
	}

	_, err = server.store.CreateUser(ctx, *arg)
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

	return &emptypb.Empty{}, nil
}

func (server *Server) parseCreateUserRequest(req *pb.CreateUserRequest) (*sqlc.CreateUserParams, error) {
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

	role := req.GetRole()
	if err := util.ValidateOneOf(role, []string{"admin", "author", "user"}); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "role: %s", err.Error())
	}

	params := &sqlc.CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: hashedPassword,
		Avatar:         server.config.AvatarPath + "/default",
		Role:           role,
	}
	return params, nil
}

// -------------------------------------------------------------------
// DeleteUsers
func (server *Server) DeleteUsers(ctx context.Context, req *pb.DeleteUsersRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	userIDs := util.RemoveDuplicates(req.GetUserIds())
	for _, userID := range userIDs {
		if err := util.ValidateID(userID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
		}
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
	arg, err := parseUpdateUserRequest(req)
	if err != nil {
		return nil, err
	}

	_, err = server.store.UpdateUser(ctx, *arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", arg.Username.String)
			case "users_email_key":
				return nil, status.Errorf(codes.AlreadyExists, "email already exists: %s", arg.Email.String)
			}
		}
		return nil, status.Error(codes.Internal, "failed to update user")
	}
	return &emptypb.Empty{}, nil
}

func parseUpdateUserRequest(req *pb.UpdateUserRequest) (*sqlc.UpdateUserParams, error) {
	reqUser := req.GetUser()

	userID := reqUser.GetId()
	if err := util.ValidateID(userID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}

	params := &sqlc.UpdateUserParams{ID: userID}
	for _, v := range req.GetUpdateMask().GetPaths() {
		switch v {
		case "username":
			username := reqUser.GetUsername()
			if err := util.ValidateString(username, 3, 50); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "username: %s", err.Error())
			}
			params.Username = sql.NullString{String: username, Valid: true}
		case "email":
			email := reqUser.GetEmail()
			if err := util.ValidateEmail(email); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "email: %s", err.Error())
			}
			params.Email = sql.NullString{String: email, Valid: true}
		case "password":
			password := reqUser.GetPassword()
			if err := util.ValidateString(password, 6, 50); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "password: %s", err.Error())
			}
			hashedPassword, err := util.HashPassword(password)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			params.HashedPassword = sql.NullString{String: hashedPassword, Valid: true}
		case "role":
			role := reqUser.GetRole()
			if err := util.ValidateOneOf(role, []string{"admin", "author", "user"}); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "role: %s", err.Error())
			}
			params.Role = sql.NullString{String: role, Valid: true}
		case "is_deleted":
			params.IsDeleted = sql.NullBool{Bool: req.User.GetIsDeleted(), Valid: true}
		}
	}
	return params, nil
}

// -------------------------------------------------------------------
// ListUsers
func (server *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	arg, err := parseListUsersRequest(req)
	if err != nil {
		return nil, err
	}

	users, err := server.store.ListUsers(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	rsp := convertListUsers(users)
	return rsp, nil
}

func parseListUsersRequest(req *pb.ListUsersRequest) (*sqlc.ListUsersParams, error) {
	options := []string{"username", "role", "idDeleted", "createAt"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.ListUsersParams{
		Limit:        req.GetPageSize(),
		Offset:       (req.GetPageId() - 1) * req.GetPageSize(),
		UsernameAsc:  req.GetOrderBy() == "username" && req.GetOrder() == "asc",
		UsernameDesc: req.GetOrderBy() == "username" && req.GetOrder() == "desc",
		RoleAsc:      req.GetOrderBy() == "role" && req.GetOrder() == "asc",
		RoleDesc:     req.GetOrderBy() == "role" && req.GetOrder() == "desc",
		DeletedAsc:   req.GetOrderBy() == "idDeleted" && req.GetOrder() == "asc",
		DeletedDesc:  req.GetOrderBy() == "idDeleted" && req.GetOrder() == "desc",
		CreateAtAsc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc: req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
		AnyKeyword:   req.GetKeyword() == "",
		Keyword:      "%" + req.GetKeyword() + "%",
	}
	return params, nil
}

// -------------------------------------------------------------------
// ChangeProfile
func (server *Server) ChangeProfile(ctx context.Context, req *pb.ChangeProfileRequest) (*pb.ChangeProfileResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg, err := parseChangeProfileRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	user, err := server.store.UpdateUser(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	rsp := &pb.ChangeProfileResponse{User: convertUser(user)}
	return rsp, nil
}

func parseChangeProfileRequest(user AuthUser, req *pb.ChangeProfileRequest) (*sqlc.UpdateUserParams, error) {
	reqUser := req.GetUser()

	userID := reqUser.GetId()
	if userID != user.ID {
		return nil, status.Error(codes.PermissionDenied, "no permission to change this user")
	}
	params := &sqlc.UpdateUserParams{ID: userID}
	for _, v := range req.GetUpdateMask().GetPaths() {
		switch v {
		case "username":
			username := reqUser.GetUsername()
			if err := util.ValidateString(username, 3, 50); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "username: %s", err.Error())
			}
			params.Username = sql.NullString{String: username, Valid: true}
		case "email":
			email := reqUser.GetEmail()
			if err := util.ValidateEmail(email); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "email: %s", err.Error())
			}
			params.Email = sql.NullString{String: email, Valid: true}
		case "info":
			info := reqUser.GetInfo()
			if err := util.ValidateString(info, 1, 150); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "role: %s", err.Error())
			}
			params.Info = sql.NullString{String: info, Valid: true}
		}
	}
	return params, nil
}

// -------------------------------------------------------------------
// ChangePassword
func (server *Server) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg, err := parseChangePasswordRequest(authUser, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = server.store.UpdateUser(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}
	return &emptypb.Empty{}, nil
}

func parseChangePasswordRequest(user AuthUser, req *pb.ChangePasswordRequest) (*sqlc.UpdateUserParams, error) {
	userID := req.GetUserId()
	if userID != user.ID {
		return nil, status.Error(codes.PermissionDenied, "no permission to change this user's password")
	}

	oldPassword := req.GetOldPassword()
	if err := util.ValidateString(oldPassword, 6, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "oldPassword: %s", err.Error())
	}
	err := util.CheckPassword(oldPassword, user.HashedPassword)
	if err != nil {
		return nil, status.Error(codes.NotFound, "incorrect old password")
	}

	newPassword := req.GetNewPassword()
	if err := util.ValidateString(newPassword, 6, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "newPassword: %s", err.Error())
	}
	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	params := &sqlc.UpdateUserParams{
		ID:             userID,
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
	}
	return params, nil
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

	arg := sqlc.GetUserProfileParams{
		UserID: req.GetUserId(),
		SelfID: authUser.ID,
	}
	user, err := server.store.GetUserProfile(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	rsp := &pb.GetUserProfileResponse{
		User: &pb.GetUserProfileResponse_User{
			Id:             user.ID,
			Username:       user.Username,
			Avatar:         user.Avatar,
			Info:           user.Info,
			StarCount:      user.StarCount,
			ViewCount:      user.ViewCount,
			FollowerCount:  user.FollowerCount,
			FollowingCount: user.FollowingCount,
			IsFollowed:     user.Followed.Valid,
		},
	}
	return rsp, nil
}
