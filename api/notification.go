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
// MarkAllRead
func (server *Server) MarkAllRead(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	err := server.store.MarkAllRead(ctx, authUser.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark all notifications as read")
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// DeleteNotifs
func (server *Server) DeleteNotifs(ctx context.Context, req *pb.DeleteNotifsRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	notifIDs := util.RemoveDuplicates(req.GetNotificationIds())
	for _, notifID := range notifIDs {
		if err := util.ValidateID(notifID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "notificationID: %s", err.Error())
		}
	}
	arg := sqlc.DeleteNotificationsParams{
		Ids:    notifIDs,
		UserID: authUser.ID,
	}
	nrows, err := server.store.DeleteNotifications(ctx, arg)
	if err != nil || int64(len(notifIDs)) != nrows {
		return nil, status.Error(codes.Internal, "failed to delete notifications")
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// ListNotifs
func (server *Server) ListNotifs(ctx context.Context, req *pb.ListNotifsRequest) (*pb.ListNotifsResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg1, err := parseListNotifsRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	notifs, err := server.store.ListNotifications(ctx, *arg1)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}

	notifIDs := []int64{}
	for _, notif := range notifs {
		if notif.Unread {
			notifIDs = append(notifIDs, notif.ID)
		}
	}

	arg2 := sqlc.MarkReadByIDsParams{
		UserID: authUser.ID,
		Ids:    notifIDs,
	}
	numRead, err := server.store.MarkReadByIDs(ctx, arg2)
	if err != nil || int64(len(notifIDs)) != numRead {
		return nil, status.Error(codes.Internal, "failed to update notifications")
	}

	rsp := convertListNotifs(notifs, numRead)
	return rsp, nil
}

func parseListNotifsRequest(user AuthUser, req *pb.ListNotifsRequest) (*sqlc.ListNotificationsParams, error) {
	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := util.ValidateOneOf(req.GetKind(), []string{"system", "reply"}); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "kind: %s", err.Error())
	}

	params := &sqlc.ListNotificationsParams{
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
		UserID: user.ID,
		Kind:   req.GetKind(),
	}
	return params, nil
}
