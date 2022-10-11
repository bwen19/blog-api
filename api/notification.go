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

// ========================// MarkAllRead //======================== //

func (server *Server) MarkAllRead(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := server.store.MarkAllRead(ctx, authUser.ID); err != nil {
		return nil, status.Error(codes.Internal, "failed to mark all notifications as read")
	}

	return &emptypb.Empty{}, nil
}

// ========================// DeleteNotifs //======================== //

func (server *Server) DeleteNotifs(ctx context.Context, req *pb.DeleteNotifsRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
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
	if err != nil || len(notifIDs) != int(nrows) {
		return nil, status.Error(codes.Internal, "failed to delete notifications")
	}

	return &emptypb.Empty{}, nil
}

// ========================// ListNotifs //======================== //

func (server *Server) ListNotifs(ctx context.Context, req *pb.ListNotifsRequest) (*pb.ListNotifsResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidatePage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	options := []string{"system", "reply"}
	if err := util.ValidateOneOf(req.GetKind(), options); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "kind: %s", err.Error())
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

	notifIDs := []int64{}
	for _, notif := range notifs {
		if notif.Unread {
			notifIDs = append(notifIDs, notif.ID)
		}
	}

	mArg := db.MarkNotificationsParams{
		Ids:    notifIDs,
		Unread: false,
	}
	nrows, err := server.store.MarkNotifications(ctx, mArg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark notifications")
	}

	return convertListNotifs(notifs, nrows), nil
}

// ========================// LeaveMessage //======================== //

func (server *Server) LeaveMessage(ctx context.Context, req *pb.LeaveMessageRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title should not be empty")
	}

	if req.GetContent() == "" {
		return nil, status.Error(codes.InvalidArgument, "content should not be empty")
	}

	arg := db.CreateNotificationParams{
		UserID:  authUser.ID,
		Kind:    "admin",
		Title:   req.GetTitle(),
		Content: req.GetContent(),
	}
	if err := server.store.CreateNotification(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to leave message to admin")
	}

	return &emptypb.Empty{}, nil
}

// ========================// ListMessages //======================== //

func (server *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

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

	return convertLisMessages(messages), nil
}

// ========================// CheckMessages //======================== //

func (server *Server) CheckMessages(ctx context.Context, req *pb.CheckMessagesRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	messageIDs, err := util.ValidateRepeatedIDs(req.GetMessageIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "messageId: %s", err.Error())
	}

	arg := db.MarkNotificationsParams{
		Unread: req.GetCheck(),
		Ids:    messageIDs,
	}

	if _, err = server.store.MarkNotifications(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to check messages")
	}

	return &emptypb.Empty{}, nil
}

// ========================// DeleteMessages //======================== //

func (server *Server) DeleteMessages(ctx context.Context, req *pb.DeleteMessagesRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	messageIDs, err := util.ValidateRepeatedIDs(req.GetMessageIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "messageId: %s", err.Error())
	}

	nrows, err := server.store.DeleteMessages(ctx, messageIDs)
	if err != nil || int(nrows) != len(messageIDs) {
		return nil, status.Error(codes.Internal, "failed to delete messages")
	}

	return &emptypb.Empty{}, nil
}
