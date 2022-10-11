package api

import (
	"context"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ========================// DeleteSessions //======================== //

func (server *Server) DeleteSessions(ctx context.Context, req *pb.DeleteSessionsRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	dict := map[string]byte{}
	ids := []uuid.UUID{}
	for _, v := range req.GetSessionIds() {
		if _, ok := dict[v]; ok {
			continue
		}
		dict[v] = 0

		sessionID, err := uuid.Parse(v)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to parse session ID")
		}
		ids = append(ids, sessionID)
	}

	nrows, err := server.store.DeleteSessions(ctx, ids)
	if err != nil || len(ids) != int(nrows) {
		return nil, status.Error(codes.Internal, "failed to delete sessions")
	}

	return &emptypb.Empty{}, nil
}

// ========================// DeleteExpiredSessions //======================== //

func (server *Server) DeleteExpiredSessions(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := server.store.DeleteExpiredSessions(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete expired sessions: %s", err.Error())
	}
	return &emptypb.Empty{}, nil
}

// ========================// ListSessions //======================== //

func (server *Server) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	options := []string{"clientIp", "createAt", "expiresAt"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListSessionsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		ClientIpAsc:   req.GetOrderBy() == "clientIp" && req.GetOrder() == "asc",
		ClientIpDesc:  req.GetOrderBy() == "clientIp" && req.GetOrder() == "desc",
		CreateAtAsc:   req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
		ExpiresAtAsc:  req.GetOrderBy() == "expiresAt" && req.GetOrder() == "asc",
		ExpiresAtDesc: req.GetOrderBy() == "expiresAt" && req.GetOrder() == "desc",
	}

	sessions, err := server.store.ListSessions(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get sessions")
	}

	return convertListSessions(sessions), nil
}
