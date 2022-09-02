package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// CreateComment
func (server *Server) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg, err := parseCreateCommentRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	comment, err := server.store.CreateComment(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create comment")
	}

	rsp := convertCreateComment(comment, authUser.User)
	return rsp, nil
}

func parseCreateCommentRequest(user AuthUser, req *pb.CreateCommentRequest) (*sqlc.CreateCommentParams, error) {
	postID := req.GetPostId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postId: %s", err.Error())
	}

	parentID := req.GetParentId()
	replyUserID := req.GetReplyUserId()

	content := req.GetContent()
	if err := util.ValidateString(content, 1, 500); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "content: %s", err.Error())
	}

	params := &sqlc.CreateCommentParams{
		PostID:      postID,
		UserID:      user.ID,
		ParentID:    sql.NullInt64{Int64: parentID, Valid: parentID > 0},
		ReplyUserID: sql.NullInt64{Int64: replyUserID, Valid: replyUserID > 0},
		Content:     content,
	}
	return params, nil
}

// -------------------------------------------------------------------
// DeleteComment
func (server *Server) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	commentID := req.GetCommentId()
	if err := util.ValidateID(commentID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "commentId: %s", err.Error())
	}

	arg := sqlc.DeleteCommentParams{
		ID:      commentID,
		IsAdmin: authUser.Role == "admin",
		UserID:  authUser.ID,
	}
	err := server.store.DeleteComment(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete comment")
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// ListComments
func (server *Server) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	arg, err := parseListCommentsRequest(req)
	if err != nil {
		return nil, err
	}

	comments, err := server.store.ListComments(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get comment list")
	}

	rsp := convertListComments(comments)
	return rsp, nil
}

func parseListCommentsRequest(req *pb.ListCommentsRequest) (*sqlc.ListCommentsParams, error) {
	options := []string{"createAt", "starCount"}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	postID := req.GetPostId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postId: %s", err.Error())
	}

	params := &sqlc.ListCommentsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		PostID:        postID,
		SelfID:        req.GetSelfId(),
		StarCountAsc:  req.GetOrderBy() == "starCount" && req.GetOrder() == "asc",
		StarCountDesc: req.GetOrderBy() == "starCount" && req.GetOrder() == "desc",
		CreateAtAsc:   req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
	}
	return params, nil
}

// -------------------------------------------------------------------
// ListReplies
func (server *Server) ListReplies(ctx context.Context, req *pb.ListRepliesRequest) (*pb.ListRepliesResponse, error) {
	arg, err := parseListRepliesRequest(req)
	if err != nil {
		return nil, err
	}

	replies, err := server.store.ListReplies(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get comment list")
	}

	rsp := convertListReplies(replies)
	return rsp, nil
}

func parseListRepliesRequest(req *pb.ListRepliesRequest) (*sqlc.ListRepliesParams, error) {
	options := []string{"createAt", "starCount"}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	commentID := req.GetCommentId()
	if err := util.ValidateID(commentID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postId: %s", err.Error())
	}

	params := &sqlc.ListRepliesParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		ParentID:      commentID,
		SelfID:        req.GetSelfId(),
		StarCountAsc:  req.GetOrderBy() == "starCount" && req.GetOrder() == "asc",
		StarCountDesc: req.GetOrderBy() == "starCount" && req.GetOrder() == "desc",
		CreateAtAsc:   req.GetOrderBy() == "createAt" && req.GetOrder() == "asc",
		CreateAtDesc:  req.GetOrderBy() == "createAt" && req.GetOrder() == "desc",
	}
	return params, nil
}

// -------------------------------------------------------------------
// StarComment
func (server *Server) StarComment(ctx context.Context, req *pb.StarCommentRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	commentID := req.GetCommentId()
	if err := util.ValidateID(commentID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "commentId: %s", err.Error())
	}

	if req.GetIsLike() {
		arg := sqlc.CreateCommentStarParams{
			CommentID: commentID,
			UserID:    authUser.ID,
		}
		err := server.store.CreateCommentStar(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create comment star")
		}
	} else {
		arg := sqlc.DeleteCommentStarParams{
			CommentID: commentID,
			UserID:    authUser.ID,
		}
		err := server.store.DeleteCommentStar(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to delete comment star")
		}
	}
	return &emptypb.Empty{}, nil
}
