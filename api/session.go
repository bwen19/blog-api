package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// DeleteSessions
func (server *Server) DeleteSessions(ctx context.Context, req *pb.DeleteSessionsRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	sessionIDs := util.RemoveDuplicates(req.GetSessionIds())
	ids := []uuid.UUID{}
	for _, v := range sessionIDs {
		sessionID, err := uuid.Parse(v)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to parse session ID")
		}
		ids = append(ids, sessionID)
	}

	arg := sqlc.DeleteSessionsParams{
		Ids:    ids,
		UserID: authUser.ID,
	}
	nrows, err := server.store.DeleteSessions(ctx, arg)
	if err != nil || int64(len(ids)) != nrows {
		return nil, status.Error(codes.Internal, "failed to delete sessions")
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// DeleteExpiredSessions
func (server *Server) DeleteExpiredSessions(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	err := server.store.DeleteExpiredSessions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete expired sessions: %s", err.Error())
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// ListSessions
func (server *Server) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg, err := parseListSessionsRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	sessions, err := server.store.ListSessions(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get sessions")
	}

	rsp := convertListSessions(sessions)
	return rsp, nil
}

func parseListSessionsRequest(user AuthUser, req *pb.ListSessionsRequest) (*sqlc.ListSessionsParams, error) {
	options := []string{"clientIp", "createAt", "expiresAt"}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.ListSessionsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		ClientIpAsc:   req.GetOrderBy() == "clientIp" && req.GetOrder() == "asc",
		ClientIpDesc:  req.GetOrderBy() == "clientIp" && req.GetOrder() == "desc",
		CreateAtAsc:   req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
		ExpiresAtAsc:  req.GetOrderBy() == "expiresAt" && req.GetOrder() == "asc",
		ExpiresAtDesc: req.GetOrderBy() == "expiresAt" && req.GetOrder() == "desc",
		UserID:        user.ID,
	}
	return params, nil
}
