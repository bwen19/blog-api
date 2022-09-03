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

	arg := sqlc.CreateNotificationParams{
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

	arg, err := parseListNotifsRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	notifs, err := server.store.ListNotifications(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}

	rsp := convertListNotifs(notifs)
	return rsp, nil
}

func parseListNotifsRequest(user AuthUser, req *pb.ListNotifsRequest) (*sqlc.ListNotificationsParams, error) {
	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := util.ValidateOneOf(req.GetKind(), []string{"system", "reply", "private"}); err != nil {
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

// -------------------------------------------------------------------
// ListMessages
func (server *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	arg, err := parseListMessagesRequest(req)
	if err != nil {
		return nil, err
	}

	messages, err := server.store.ListMessages(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}

	rsp := convertLisMessages(messages)
	return rsp, nil
}

func parseListMessagesRequest(req *pb.ListMessagesRequest) (*sqlc.ListMessagesParams, error) {
	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.ListMessagesParams{
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
	}
	return params, nil
}

// -------------------------------------------------------------------
// CheckMessages
func (server *Server) CheckMessages(ctx context.Context, req *pb.CheckMessagesRequest) (*emptypb.Empty, error) {
	messageIDs := util.RemoveDuplicates(req.GetMessageIds())
	for _, messageID := range messageIDs {
		if err := util.ValidateID(messageID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "messageId: %s", err.Error())
		}
	}

	err := server.store.CheckMessage(ctx, messageIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check messages")
	}
	return &emptypb.Empty{}, nil
}
