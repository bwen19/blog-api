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
// LeaveMessage
func (server *Server) LeaveMessage(ctx context.Context, req *pb.LeaveMessageRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	title := req.GetTitle()
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "title should not be empty")
	}
	content := req.GetContent()
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "content should not be empty")
	}

	arg := db.CreateNotificationParams{
		UserID:  authUser.ID,
		Kind:    "admin",
		Title:   title,
		Content: content,
	}
	err := server.store.CreateNotification(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to leave message to admin")
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

	notifIDs, err := util.ValidateRepeatedIDs(req.GetNotificationIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "notificationID: %s", err.Error())
	}

	arg := db.DeleteNotificationsParams{
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
	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := util.ValidateOneOf(req.GetKind(), []string{"system", "reply", "private"}); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "kind: %s", err.Error())
	}

	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg := db.ListNotificationsParams{
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
		UserID: authUser.ID,
		Kind:   req.GetKind(),
	}

	notifs, err := server.store.ListNotifications(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}

	rsp := convertListNotifs(notifs)
	return rsp, nil
}

// -------------------------------------------------------------------
// ListMessages
func (server *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListMessagesParams{
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
	}

	messages, err := server.store.ListMessages(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}

	rsp := convertLisMessages(messages)
	return rsp, nil
}

// -------------------------------------------------------------------
// CheckMessages
func (server *Server) CheckMessages(ctx context.Context, req *pb.CheckMessagesRequest) (*emptypb.Empty, error) {
	messageIDs, err := util.ValidateRepeatedIDs(req.GetMessageIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "messageId: %s", err.Error())
	}

	if err = server.store.CheckMessage(ctx, messageIDs); err != nil {
		return nil, status.Error(codes.Internal, "failed to check messages")
	}
	return &emptypb.Empty{}, nil
}
