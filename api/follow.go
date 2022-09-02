package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// FollowUser
func (server *Server) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	userID := req.GetUserId()
	if err := util.ValidateID(userID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}
	if userID == authUser.ID {
		return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
	}

	if req.GetIsLike() {
		arg1 := sqlc.CreateFollowParams{
			UserID:     userID,
			FollowerID: authUser.ID,
		}
		err := server.store.CreateFollow(ctx, arg1)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create follow")
		}
	} else {
		arg := sqlc.DeleteFollowParams{
			UserID:     userID,
			FollowerID: authUser.ID,
		}
		err := server.store.DeleteFollow(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to delete follow")
		}
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// ListFollowers
func (server *Server) ListFollows(ctx context.Context, req *pb.ListFollowsRequest) (*pb.ListFollowsResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg1, err := parseListFollowsRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	var rsp *pb.ListFollowsResponse
	if req.GetIsFollower() {
		followers, err := server.store.ListFollowers(ctx, *arg1)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list follows", err)
		}
		rsp = convertListFollowers(followers)
	} else {
		arg2 := sqlc.ListFollowingsParams(*arg1)
		followings, err := server.store.ListFollowings(ctx, arg2)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list follows", err)
		}
		rsp = convertListFollowings(followings)
	}

	return rsp, nil
}

func parseListFollowsRequest(user AuthUser, req *pb.ListFollowsRequest) (*sqlc.ListFollowersParams, error) {
	err := util.ValidatePage(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := util.ValidateID(req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}

	params := &sqlc.ListFollowersParams{
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
		UserID: req.GetUserId(),
		SelfID: user.ID,
	}
	return params, nil
}
