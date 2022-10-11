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

// ========================// CreatePost //======================== //

func (server *Server) CreatePost(ctx context.Context, req *emptypb.Empty) (*pb.CreatePostResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleAuthor)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	arg := db.CreateNewPostParams{
		AuthorID:   authUser.ID,
		Title:      "",
		CoverImage: server.config.DefaultCover,
		Content:    "",
	}

	post, err := server.store.CreateNewPost(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create post")
	}

	rsp := &pb.CreatePostResponse{Post: convertNewPost(post)}
	return rsp, nil
}

// ========================// DeletePost //======================== //

func (server *Server) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleAuthor)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidateID(req.GetPostId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.DeletePostParams{
		ID:       req.PostId,
		AuthorID: authUser.ID,
	}

	if err := server.store.DeletePost(ctx, arg); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete post")
	}
	return &emptypb.Empty{}, nil
}

// ========================// UpdatePost //======================== //

func (server *Server) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleAuthor)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := validateUpdatePostRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.UpdatePostParams{
		ID:         req.GetPostId(),
		AuthorID:   authUser.ID,
		Title:      sql.NullString{String: req.GetTitle(), Valid: req.Title != nil},
		CoverImage: sql.NullString{String: req.GetCoverImage(), Valid: req.CoverImage != nil},
	}

	newPost, err := server.store.UpdatePost(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update post")
	}

	var newContent db.PostContent
	if req.Content != nil {
		arg := db.UpdatePostContentParams{
			ID:       req.GetPostId(),
			Content:  req.GetContent(),
			AuthorID: authUser.ID,
		}
		if newContent, err = server.store.UpdatePostContent(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to update post content")
		}
	}

	var categories []db.Category
	if req.CategoryIds != nil {
		arg := db.SetPostCategoriesParams{
			PostID:      newPost.ID,
			CategoryIDs: req.GetCategoryIds(),
		}
		if categories, err = server.store.SetPostCategories(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to set post categories")
		}
	} else {
		if categories, err = server.store.GetPostCategories(ctx, newPost.ID); err != nil {
			return nil, status.Error(codes.Internal, "failed to get post categories")
		}
	}

	var tags []db.Tag
	if req.TagIds != nil {
		arg := db.SetPostTagsParams{
			PostID: newPost.ID,
			TagIDs: req.GetTagIds(),
		}
		if tags, err = server.store.SetPostTags(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to set post tags")
		}
	} else {
		if tags, err = server.store.GetPostTags(ctx, newPost.ID); err != nil {
			return nil, status.Error(codes.Internal, "failed to get post tags")
		}
	}

	rsp := convertUpdatePost(newPost, newContent, categories, tags)
	return rsp, nil
}

func validateUpdatePostRequest(req *pb.UpdatePostRequest) error {
	if err := util.ValidateID(req.GetPostId()); err != nil {
		return fmt.Errorf("postID: %s", err.Error())
	}
	if req.Title != nil {
		if err := util.ValidateString(req.GetTitle(), 1, 200); err != nil {
			return fmt.Errorf("title: %s", err.Error())
		}
	}
	if req.CoverImage != nil {
		if err := util.ValidateString(req.GetCoverImage(), 1, 100); err != nil {
			return fmt.Errorf("coverImage: %s", err.Error())
		}
	}
	return nil
}

// ========================// SubmitPost //======================== //

func (server *Server) SubmitPost(ctx context.Context, req *pb.SubmitPostRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleAuthor)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	postIDs, err := util.ValidateRepeatedIDs(req.GetPostIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	arg := db.UpdatePostStatusParams{
		Ids:       postIDs,
		Status:    "review",
		OldStatus: []string{"draft", "revise"},
		IsAdmin:   authUser.Role == "admin",
		AuthorID:  authUser.ID,
	}

	newPosts, err := server.store.UpdatePostStatus(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to submit posts")
	}

	for _, post := range newPosts {
		arg := db.CreateNotificationParams{
			UserID:  authUser.ID,
			Kind:    "admin",
			Title:   "New post submitted",
			Content: fmt.Sprintf("Post entitled \"%s\" has been submitted", post.Title),
		}

		if err = server.store.CreateNotification(ctx, arg); err != nil {
			log.Println("failed to create new notification for submitting post")
		}
	}

	if len(newPosts) != len(postIDs) {
		return nil, status.Error(codes.Internal, "some submissions failed")
	}
	return &emptypb.Empty{}, nil
}

// ========================// PublishPost //======================== //

func (server *Server) PublishPost(ctx context.Context, req *pb.PublishPostRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	postIDs, err := util.ValidateRepeatedIDs(req.GetPostIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	arg := db.UpdatePostStatusParams{
		Ids:       postIDs,
		Status:    "publish",
		OldStatus: []string{"review"},
		IsAdmin:   true,
	}

	newPosts, err := server.store.UpdatePostStatus(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to publish posts")
	}

	for _, post := range newPosts {
		arg := db.CreateNotificationParams{
			UserID:  post.AuthorID,
			Kind:    "system",
			Title:   "New post published",
			Content: fmt.Sprintf("Congratulations! Post entitled \"%s\" has been published", post.Title),
		}

		if err = server.store.CreateNotification(ctx, arg); err != nil {
			log.Println("failed to create new notification for publishing post")
		}
	}

	if len(newPosts) != len(postIDs) {
		return nil, status.Error(codes.Internal, "some publications failed")
	}
	return &emptypb.Empty{}, nil
}

// ========================// WithdrawPost //======================== //

func (server *Server) WithdrawPost(ctx context.Context, req *pb.WithdrawPostRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	postIDs, err := util.ValidateRepeatedIDs(req.GetPostIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	arg := db.UpdatePostStatusParams{
		Ids:       postIDs,
		Status:    "revise",
		OldStatus: []string{"publish", "review"},
		IsAdmin:   true,
	}

	withdrawPosts, err := server.store.UpdatePostStatus(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to withdraw posts")
	}

	for _, post := range withdrawPosts {
		arg := db.CreateNotificationParams{
			UserID:  post.AuthorID,
			Kind:    "system",
			Title:   "Post withdrawn",
			Content: fmt.Sprintf("Post \"%s\" has been withdrawn", post.Title),
		}

		if err = server.store.CreateNotification(ctx, arg); err != nil {
			log.Println("failed to create new notification for withdrawing post")
		}
	}

	if len(withdrawPosts) != len(postIDs) {
		return nil, status.Error(codes.Internal, "some withdrawal failed")
	}
	return &emptypb.Empty{}, nil
}

// ========================// UpdatePostLabel //======================== //

func (server *Server) UpdatePostLabel(ctx context.Context, req *pb.UpdatePostLabelRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	categoryIDs, tagIDs, err := validateUpdatePostLabelRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Featured != nil {
		arg := db.UpdatePostFeatureParams{
			ID:       req.GetPostId(),
			Featured: req.GetFeatured(),
		}
		if err := server.store.UpdatePostFeature(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to update post feature")
		}
	}

	if req.CategoryIds != nil {
		arg := db.SetPostCategoriesParams{
			PostID:      req.GetPostId(),
			CategoryIDs: categoryIDs,
		}
		if _, err := server.store.SetPostCategories(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to update post categories")
		}
	}

	if req.TagIds != nil {
		arg := db.SetPostTagsParams{
			PostID: req.GetPostId(),
			TagIDs: tagIDs,
		}
		if _, err := server.store.SetPostTags(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to update post tags")
		}
	}
	return &emptypb.Empty{}, nil
}

func validateUpdatePostLabelRequest(req *pb.UpdatePostLabelRequest) ([]int64, []int64, error) {
	var err error
	if err = util.ValidateID(req.GetPostId()); err != nil {
		return nil, nil, fmt.Errorf("postID: %s", err.Error())
	}

	var categoryIDs []int64
	if req.CategoryIds != nil {
		if categoryIDs, err = util.ValidateRepeatedIDs(req.GetCategoryIds()); err != nil {
			return nil, nil, fmt.Errorf("categoryID: %s", err.Error())
		}
		if len(categoryIDs) > 2 {
			return nil, nil, fmt.Errorf("post should not have more than 2 categories")
		}
	}

	var tagIDs []int64
	if req.TagIds != nil {
		if tagIDs, err = util.ValidateRepeatedIDs(req.GetTagIds()); err != nil {
			return nil, nil, fmt.Errorf("tagID: %s", err.Error())
		}
		if len(tagIDs) > 3 {
			return nil, nil, fmt.Errorf("post should not have more than 3 tags")
		}
	}
	return categoryIDs, tagIDs, nil
}

// ========================// ListPosts //======================== //

func (server *Server) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleAuthor)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := validateListPostsRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListPostsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		UpdateAtAsc:   req.GetOrderBy() == "updateAt" && req.GetOrder() == "asc",
		UpdateAtDesc:  req.GetOrderBy() == "updateAt" && req.GetOrder() == "desc",
		PublishAtAsc:  req.GetOrderBy() == "publishAt" && req.GetOrder() == "asc",
		PublishAtDesc: req.GetOrderBy() == "publishAt" && req.GetOrder() == "desc",
		ViewCountAsc:  req.GetOrderBy() == "viewCount" && req.GetOrder() == "asc",
		ViewCountDesc: req.GetOrderBy() == "viewCount" && req.GetOrder() == "desc",
		IsAdmin:       authUser.Role == "admin",
		AuthorID:      authUser.ID,
		AnyStatus:     req.Status == nil,
		Status:        req.GetStatus(),
		AnyKeyword:    req.Keyword == nil,
		Keyword:       "%" + req.GetKeyword() + "%",
	}

	posts, err := server.store.ListPosts(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list posts")
	}
	return convertListPosts(posts), nil
}

func validateListPostsRequest(req *pb.ListPostsRequest) error {
	options := []string{"publishAt", "updateAt", "viewCount"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return err
	}
	if req.Status != nil {
		options = []string{"publish", "review", "revise", "draft"}
		if err := util.ValidateOneOf(req.GetStatus(), options); err != nil {
			return fmt.Errorf("status: %s", err.Error())
		}
	}
	if req.Keyword != nil {
		if err := util.ValidateString(req.GetKeyword(), 1, 50); err != nil {
			return fmt.Errorf("keyword: %s", err.Error())
		}
	}
	return nil
}

// ========================// GetPost //======================== //

func (server *Server) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleAuthor)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidateID(req.GetPostId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postId: %s", err.Error())
	}

	arg := db.GetPostParams{
		PostID:   req.GetPostId(),
		IsAdmin:  authUser.Role == "admin",
		AuthorID: authUser.ID,
	}

	post, err := server.store.GetPost(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get post")
	}
	return convertGetPost(post), nil
}

// ========================// GetFeaturedPosts //======================== //

func (server *Server) GetFeaturedPosts(ctx context.Context, req *pb.GetFeaturedPostsRequest) (*pb.GetFeaturedPostsResponse, error) {
	if err := util.ValidateNumber(int64(req.GetNum()), 1, 8); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "num: %s", err.Error())
	}

	posts, err := server.store.GetFeaturedPosts(ctx, req.GetNum())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get featured posts")
	}
	return convertFeaturedPosts(posts), nil
}

// ========================// GetPosts //======================== //

func (server *Server) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	if err := validateGetPostsRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.GetPostsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		PublishAtAsc:  req.GetOrderBy() == "publishAt" && req.GetOrder() == "asc",
		PublishAtDesc: req.GetOrderBy() == "publishAt" && req.GetOrder() == "desc",
		ViewCountAsc:  req.GetOrderBy() == "viewCount" && req.GetOrder() == "asc",
		ViewCountDesc: req.GetOrderBy() == "viewCount" && req.GetOrder() == "desc",
		AnyFeatured:   req.Featured == nil,
		Featured:      req.GetFeatured(),
		AnyAuthor:     req.AuthorId == nil,
		AuthorID:      req.GetAuthorId(),
		AnyCategory:   req.CategoryId == nil,
		CategoryID:    req.GetCategoryId(),
		AnyTag:        req.TagId == nil,
		TagID:         req.GetTagId(),
		AnyKeyword:    req.Keyword == nil,
		Keyword:       "%" + req.GetKeyword() + "%",
	}

	posts, err := server.store.GetPosts(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list posts")
	}
	return convertGetPosts(posts), nil
}

func validateGetPostsRequest(req *pb.GetPostsRequest) error {
	options := []string{"publishAt", "viewCount"}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return err
	}
	if req.AuthorId != nil {
		if err := util.ValidateID(req.GetAuthorId()); err != nil {
			return fmt.Errorf("authorId: %s", err.Error())
		}
	}
	if req.CategoryId != nil {
		if err := util.ValidateID(req.GetCategoryId()); err != nil {
			return fmt.Errorf("categoryId: %s", err.Error())
		}
	}
	if req.TagId != nil {
		if err := util.ValidateID(req.GetTagId()); err != nil {
			return fmt.Errorf("tagId: %s", err.Error())
		}
	}
	if req.Keyword != nil {
		if err := util.ValidateString(req.GetKeyword(), 1, 50); err != nil {
			return fmt.Errorf("keyword: %s", err.Error())
		}
	}
	return nil
}

// ========================// ReadPost //======================== //

func (server *Server) ReadPost(ctx context.Context, req *pb.ReadPostRequest) (*pb.ReadPostResponse, error) {
	authUser, gErr := server.grpcGuard(ctx, roleGhost)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidateID(req.GetPostId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	arg := db.ReadPostParams{
		PostID: req.GetPostId(),
		SelfID: authUser.ID,
	}

	post, err := server.store.ReadPost(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to read post")
	}
	return convertReadPost(post), nil
}

// ========================// StarPost //======================== //

func (server *Server) StarPost(ctx context.Context, req *pb.StarPostRequest) (*emptypb.Empty, error) {
	authUser, gErr := server.grpcGuard(ctx, roleUser)
	if gErr != nil {
		return nil, gErr.GrpcErr()
	}

	postID := req.GetPostId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	if req.GetLike() {
		arg := db.CreatePostStarParams{
			PostID: postID,
			UserID: authUser.ID,
		}
		if err := server.store.CreatePostStar(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to create post star")
		}
	} else {
		arg := db.DeletePostStarParams{
			PostID: postID,
			UserID: authUser.ID,
		}
		if err := server.store.DeletePostStar(ctx, arg); err != nil {
			return nil, status.Error(codes.Internal, "failed to delete post star")
		}
	}
	return &emptypb.Empty{}, nil
}
