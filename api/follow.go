package api

import (
	"context"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ========================// FollowUser //======================== //

func (server *Server) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	userID := req.GetUserId()
	if err := util.ValidateID(userID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}
	if userID == authUser.ID {
		return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
	}

	if req.GetLike() {
		arg := db.CreateFollowParams{
			UserID:     userID,
			FollowerID: authUser.ID,
		}
		err := server.store.CreateFollow(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create follow")
		}
	} else {
		arg := db.DeleteFollowParams{
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

// ========================// ListFollows //======================== //

func (server *Server) ListFollows(ctx context.Context, req *pb.ListFollowsRequest) (*pb.ListFollowsResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := util.ValidateID(req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "userId: %s", err.Error())
	}

	arg := db.ListFollowersParams{
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
		UserID: req.GetUserId(),
		SelfID: authUser.ID,
	}

	if req.GetFollower() {
		followers, err := server.store.ListFollowers(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to list follows")
		}
		return convertListFollowers(followers), nil
	} else {
		followings, err := server.store.ListFollowings(ctx, db.ListFollowingsParams(arg))
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to list follows")
		}
		return convertListFollowings(followings), nil
	}
}
