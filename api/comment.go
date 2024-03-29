package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ========================// CreateComment //======================== //

func (server *Server) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := validateCreateCommentRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.CreateCommentParams{
		PostID:      req.GetPostId(),
		UserID:      authUser.ID,
		ParentID:    sql.NullInt64{Int64: req.GetParentId(), Valid: req.ParentId != nil},
		ReplyUserID: sql.NullInt64{Int64: req.GetReplyUserId(), Valid: req.ReplyUserId != nil},
		Content:     req.GetContent(),
	}

	comment, err := server.store.CreateComment(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create comment")
	}

	userID := comment.AuthorID
	if comment.ReplyUserID.Valid {
		userID = comment.ReplyUserID.Int64
	}
	arg2 := db.CreateNotificationParams{
		UserID:  userID,
		Kind:    "reply",
		Title:   "New Comment at You",
		Content: comment.Content,
	}

	if err = server.store.CreateNotification(ctx, arg2); err != nil {
		log.Println("failed to create new notification for post comment")
	}

	rsp := convertCreateComment(comment, authUser)
	return rsp, nil
}

func validateCreateCommentRequest(req *pb.CreateCommentRequest) error {
	if err := util.ValidateID(req.GetPostId()); err != nil {
		return fmt.Errorf("postId: %s", err.Error())
	}
	if req.ParentId != nil {
		if err := util.ValidateID(req.GetParentId()); err != nil {
			return fmt.Errorf("parentId: %s", err.Error())
		}
	}
	if req.ReplyUserId != nil {
		if err := util.ValidateID(req.GetReplyUserId()); err != nil {
			return fmt.Errorf("replyUserId: %s", err.Error())
		}
	}
	if err := util.ValidateString(req.GetContent(), 1, 500); err != nil {
		return fmt.Errorf("content: %s", err.Error())
	}
	return nil
}

// ========================// DeleteComment //======================== //

func (server *Server) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidateID(req.GetCommentId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "commentId: %s", err.Error())
	}

	arg := db.DeleteCommentParams{
		ID:      req.GetCommentId(),
		IsAdmin: authUser.Role == "admin",
		UserID:  authUser.ID,
	}

	if err := server.store.DeleteComment(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete comment")
	}
	return &emptypb.Empty{}, nil
}

// ========================// ListComments //======================== //

func (server *Server) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleGhost)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := validateListCommentsRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListCommentsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		PostID:        req.GetPostId(),
		SelfID:        authUser.ID,
		StarCountAsc:  req.GetOrderBy() == "starCount" && req.GetOrder() == "asc",
		StarCountDesc: req.GetOrderBy() == "starCount" && req.GetOrder() == "desc",
		CreateAtAsc:   req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
	}

	comments, err := server.store.ListComments(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get comment list")
	}
	return convertListComments(comments), nil
}

func validateListCommentsRequest(req *pb.ListCommentsRequest) error {
	options := []string{"createAt", "starCount"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return err
	}
	if err := util.ValidateID(req.GetPostId()); err != nil {
		return fmt.Errorf("postId: %s", err.Error())
	}
	return nil
}

// ========================// ListReplies //======================== //

func (server *Server) ListReplies(ctx context.Context, req *pb.ListRepliesRequest) (*pb.ListRepliesResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleGhost)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := validateListRepliesRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListRepliesParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		ParentID:      req.GetCommentId(),
		SelfID:        authUser.ID,
		StarCountAsc:  req.GetOrderBy() == "starCount" && req.GetOrder() == "asc",
		StarCountDesc: req.GetOrderBy() == "starCount" && req.GetOrder() == "desc",
		CreateAtAsc:   req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
	}

	replies, err := server.store.ListReplies(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get comment list")
	}
	return convertListReplies(replies), nil
}

func validateListRepliesRequest(req *pb.ListRepliesRequest) error {
	options := []string{"createAt", "starCount"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return err
	}
	if err := util.ValidateID(req.GetCommentId()); err != nil {
		return fmt.Errorf("postId: %s", err.Error())
	}
	return nil
}

// ========================// StarComment //======================== //

func (server *Server) StarComment(ctx context.Context, req *pb.StarCommentRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidateID(req.GetCommentId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "commentId: %s", err.Error())
	}

	if req.GetLike() {
		arg := db.CreateCommentStarParams{
			CommentID: req.GetCommentId(),
			UserID:    authUser.ID,
		}
		if err := server.store.CreateCommentStar(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to create comment star")
		}
	} else {
		arg := db.DeleteCommentStarParams{
			CommentID: req.GetCommentId(),
			UserID:    authUser.ID,
		}
		if err := server.store.DeleteCommentStar(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to delete comment star")
		}
	}
	return &emptypb.Empty{}, nil
}
