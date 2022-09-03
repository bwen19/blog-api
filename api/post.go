package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"
	"database/sql"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// CreatePost
func (server *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg1, arg2, err := parseCreatePostRequest(authUser, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	post, err := server.store.CreatePost(ctx, *arg1)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create post")
	}

	categories := []sqlc.Category{}
	if arg2.SetCategory {
		categories, err = server.setPostCategories(ctx, post.ID, arg2.CategoryIDs)
		if err != nil {
			return nil, err
		}
	}

	tags := []sqlc.Tag{}
	if arg2.SetTag {
		tags, err = server.setPostTags(ctx, post.ID, arg2.TagNames)
		if err != nil {
			return nil, err
		}
	}

	rsp := &pb.CreatePostResponse{
		Post: convertPost(post, categories, tags),
	}
	return rsp, nil
}

type SetPostLabels struct {
	SetCategory bool
	CategoryIDs []int64
	SetTag      bool
	TagNames    []string
}

func parseCreatePostRequest(user AuthUser, req *pb.CreatePostRequest) (*sqlc.CreatePostParams, *SetPostLabels, error) {
	title := req.GetTitle()
	if err := util.ValidateString(title, 0, 200); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "title: %s", err.Error())
	}

	abstract := req.GetAbstract()
	if err := util.ValidateString(abstract, 0, 1000); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "abstract: %s", err.Error())
	}

	coverImage := req.GetCoverImage()
	if err := util.ValidateString(coverImage, 0, 200); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "coverImage: %s", err.Error())
	}

	params1 := &sqlc.CreatePostParams{
		AuthorID:   user.ID,
		Title:      title,
		Abstract:   abstract,
		CoverImage: coverImage,
		Content:    req.GetContent(),
	}

	categoryIDs := req.GetCategoryIds()
	if len(categoryIDs) > 2 {
		return nil, nil, status.Errorf(codes.InvalidArgument, "post should not have more than 2 categories")
	}
	for _, categoryID := range categoryIDs {
		if err := util.ValidateID(categoryID); err != nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "categoryID: %s", err.Error())
		}
	}

	tagNames := req.GetTagNames()
	if len(tagNames) > 5 {
		return nil, nil, status.Errorf(codes.InvalidArgument, "post should not have more than 5 tags")
	}
	for _, tagName := range tagNames {
		if err := util.ValidateString(tagName, 1, 50); err != nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "tagName: %s", err.Error())
		}
	}

	params2 := &SetPostLabels{
		SetCategory: len(categoryIDs) > 0,
		CategoryIDs: categoryIDs,
		SetTag:      len(tagNames) > 0,
		TagNames:    tagNames,
	}
	return params1, params2, nil
}

// -------------------------------------------------------------------
// DeletePosts
func (server *Server) DeletePosts(ctx context.Context, req *pb.DeletePostsRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	postIDs := util.RemoveDuplicates(req.GetPostIds())
	for _, postID := range postIDs {
		if err := util.ValidateID(postID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
		}
	}

	arg := sqlc.DeletePostsParams{
		Ids:      postIDs,
		AuthorID: authUser.ID,
	}
	nrows, err := server.store.DeletePosts(ctx, arg)
	if err != nil || int64(len(postIDs)) != nrows {
		return nil, status.Error(codes.Internal, "failed to delete posts")
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// UpdatePost
func (server *Server) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg1, arg2, err := parseUpdatePostRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	newPost, err := server.store.UpdatePost(ctx, *arg1)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update post")
	}

	var categories []sqlc.Category
	if arg2.SetCategory {
		categories, err = server.setPostCategories(ctx, newPost.ID, arg2.CategoryIDs)
		if err != nil {
			return nil, err
		}
	} else {
		categories, err = server.store.GetPostCategories(ctx, newPost.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get post categories")
		}
	}

	var tags []sqlc.Tag
	if arg2.SetTag {
		tags, err = server.setPostTags(ctx, newPost.ID, req.GetPost().GetTagNames())
		if err != nil {
			return nil, err
		}
	} else {
		tags, err = server.store.GetPostTags(ctx, newPost.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get post tags")
		}
	}

	rsp := &pb.UpdatePostResponse{
		Post: convertPost(newPost, categories, tags),
	}
	return rsp, nil
}

func parseUpdatePostRequest(user AuthUser, req *pb.UpdatePostRequest) (*sqlc.UpdatePostParams, *SetPostLabels, error) {
	reqPost := req.GetPost()
	postID := reqPost.GetId()

	if err := util.ValidateID(postID); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}
	params1 := &sqlc.UpdatePostParams{
		ID:       postID,
		AuthorID: user.ID,
	}
	params2 := &SetPostLabels{}

	for _, v := range req.GetUpdateMask().GetPaths() {
		switch v {
		case "title":
			title := reqPost.GetTitle()
			if err := util.ValidateString(title, 1, 200); err != nil {
				return nil, nil, status.Errorf(codes.InvalidArgument, "title: %s", err.Error())
			}
			params1.Title = sql.NullString{String: title, Valid: true}
		case "abstract":
			abstract := reqPost.GetAbstract()
			if abstract == "" {
				return nil, nil, status.Error(codes.InvalidArgument, "abstract: must be a non empty string")
			}
			params1.Abstract = sql.NullString{String: abstract, Valid: true}
		case "cover_image":
			coverImage := reqPost.GetCoverImage()
			if coverImage == "" {
				return nil, nil, status.Error(codes.InvalidArgument, "coverImage: must be a non empty path")
			}
			params1.CoverImage = sql.NullString{String: coverImage, Valid: true}
		case "content":
			content := reqPost.GetContent()
			if content == "" {
				return nil, nil, status.Error(codes.InvalidArgument, "content: must be a non empty content")
			}
			params1.Content = sql.NullString{String: content, Valid: true}
		case "category_ids":
			categoryIDs := reqPost.GetCategoryIds()
			if len(categoryIDs) > 2 {
				return nil, nil, status.Error(codes.InvalidArgument, "post should not have more than 2 categories")
			}
			for _, categoryID := range categoryIDs {
				if err := util.ValidateID(categoryID); err != nil {
					return nil, nil, status.Errorf(codes.InvalidArgument, "categoryId: %s", err.Error())
				}
			}
			params2.SetCategory = true
			params2.CategoryIDs = categoryIDs
		case "tag_names":
			tagNames := reqPost.GetTagNames()
			if len(tagNames) > 5 {
				return nil, nil, status.Error(codes.InvalidArgument, "post should not have more than 5 tags")
			}
			for _, tagName := range tagNames {
				if err := util.ValidateString(tagName, 1, 50); err != nil {
					return nil, nil, status.Errorf(codes.InvalidArgument, "tagName: %s", err.Error())
				}
			}
			params2.SetTag = true
			params2.TagNames = tagNames
		}
	}
	return params1, params2, nil
}

// -------------------------------------------------------------------
// ListPosts
func (server *Server) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	arg, err := parseListPostsRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	posts, err := server.store.ListPosts(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list posts")
	}

	rsp := convertListPosts(posts)
	return rsp, nil
}

func parseListPostsRequest(user AuthUser, req *pb.ListPostsRequest) (*sqlc.ListPostsParams, error) {
	options := []string{"publishAt", "updateAt"}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.ListPostsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		UpdateAtAsc:   req.GetOrderBy() == "updateAt" && req.GetOrder() == "asc",
		UpdateAtDesc:  req.GetOrderBy() == "updateAt" && req.GetOrder() == "desc",
		PublishAtAsc:  req.GetOrderBy() == "publishAt" && req.GetOrder() == "asc",
		PublishAtDesc: req.GetOrderBy() == "publishAt" && req.GetOrder() == "desc",
		AnyStatus:     req.GetStatus() == "",
		Status:        req.GetStatus(),
		AuthorID:      user.ID,
		AnyKeyword:    req.GetKeyword() == "",
		Keyword:       "%" + req.GetKeyword() + "%",
	}
	return params, nil
}

// -------------------------------------------------------------------
// SubmitPost
func (server *Server) SubmitPost(ctx context.Context, req *pb.SubmitPostRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	postIDs := util.RemoveDuplicates(req.GetPostIds())
	for _, postID := range postIDs {
		if err := util.ValidateID(postID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
		}
	}

	arg1 := sqlc.SubmitPostParams{
		Ids:      postIDs,
		AuthorID: authUser.ID,
	}

	submitIDs, err := server.store.SubmitPost(ctx, arg1)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to submit posts")
	}

	for _, postID := range submitIDs {
		arg2 := sqlc.CreateNotificationParams{
			UserID:  authUser.ID,
			Kind:    "admin",
			Title:   "New post submitted",
			Content: fmt.Sprintf("PostID %v has been submitted", postID),
		}

		err = server.store.CreateNotification(ctx, arg2)
		if err != nil {
			log.Println("failed to create new notification for submitting post")
		}
	}

	if len(submitIDs) != len(postIDs) {
		return nil, status.Error(codes.Internal, "some submissions failed")
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// ReviewPost
func (server *Server) ReviewPost(ctx context.Context, req *pb.ReviewPostRequest) (*pb.ReviewPostResponse, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	postID := req.GetPostId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postId: %s", err.Error())
	}

	arg := sqlc.ReviewPostParams{
		PostID:   postID,
		IsAdmin:  authUser.Role == "admin",
		AuthorID: authUser.ID,
	}
	post, err := server.store.ReviewPost(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get post")
	}

	rsp := convertReviewPost(post)
	return rsp, nil
}

// -------------------------------------------------------------------
// PublishPost
func (server *Server) PublishPost(ctx context.Context, req *pb.PublishPostRequest) (*emptypb.Empty, error) {
	postIDs := util.RemoveDuplicates(req.GetPostIds())
	for _, postID := range postIDs {
		if err := util.ValidateID(postID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
		}
	}

	publishPosts, err := server.store.PublishPost(ctx, postIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to publish posts")
	}

	for _, post := range publishPosts {
		arg := sqlc.CreateNotificationParams{
			UserID:  post.AuthorID,
			Kind:    "system",
			Title:   "New post published",
			Content: fmt.Sprintf("Congratulations! PostID %v has been published", post.ID),
		}

		err = server.store.CreateNotification(ctx, arg)
		if err != nil {
			log.Println("failed to create new notification for publishing post")
		}
	}

	if len(publishPosts) != len(postIDs) {
		return nil, status.Error(codes.Internal, "some publications failed")
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// WithdrawPost
func (server *Server) WithdrawPost(ctx context.Context, req *pb.WithdrawPostRequest) (*emptypb.Empty, error) {
	postIDs := util.RemoveDuplicates(req.GetPostIds())
	for _, postID := range postIDs {
		if err := util.ValidateID(postID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
		}
	}

	withdrawPosts, err := server.store.WithdrawPost(ctx, postIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to withdraw posts")
	}

	for _, post := range withdrawPosts {
		arg := sqlc.CreateNotificationParams{
			UserID:  post.AuthorID,
			Kind:    "system",
			Title:   "Post withdrawn",
			Content: fmt.Sprintf("PostID %v has been withdrawn", post.ID),
		}

		err = server.store.CreateNotification(ctx, arg)
		if err != nil {
			log.Println("failed to create new notification for withdrawing post")
		}
	}

	if len(withdrawPosts) != len(postIDs) {
		return nil, status.Error(codes.Internal, "some withdrawal failed")
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// ChangePost
func (server *Server) ChangePost(ctx context.Context, req *pb.ChangePostRequest) (*emptypb.Empty, error) {
	reqPost := req.GetPost()

	postID := reqPost.GetId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	for _, v := range req.GetUpdateMask().GetPaths() {
		switch v {
		case "category_ids":
			categoryIDs := reqPost.GetCategoryIds()
			if len(categoryIDs) > 2 {
				return nil, status.Error(codes.InvalidArgument, "post should not have more than 2 categories")
			}
			for _, categoryID := range categoryIDs {
				if err := util.ValidateID(categoryID); err != nil {
					return nil, status.Errorf(codes.InvalidArgument, "categoryId: %s", err.Error())
				}
			}
			_, err := server.setPostCategories(ctx, postID, categoryIDs)
			if err != nil {
				return nil, err
			}
		case "is_featured":
			arg := sqlc.FeaturePostParams{
				ID:         postID,
				IsFeatured: req.Post.GetIsFeatured(),
			}
			nrows, err := server.store.FeaturePost(ctx, arg)
			if err != nil || nrows != 1 {
				return nil, status.Error(codes.Internal, "failed to update post is_featured")
			}
		}
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// GetPosts
func (server *Server) GetPosts(ctx context.Context, req *pb.GetPostsRequest) (*pb.GetPostsResponse, error) {
	arg, err := parseGetPostsRequest(req)
	if err != nil {
		return nil, err
	}

	posts, err := server.store.GetPosts(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list posts")
	}

	rsp := convertGetPosts(posts)
	return rsp, nil
}

func parseGetPostsRequest(req *pb.GetPostsRequest) (*sqlc.GetPostsParams, error) {
	options := []string{"publishAt", "updateAt"}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.GetPostsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		UpdateAtAsc:   req.GetOrderBy() == "updateAt" && req.GetOrder() == "asc",
		UpdateAtDesc:  req.GetOrderBy() == "updateAt" && req.GetOrder() == "desc",
		PublishAtAsc:  req.GetOrderBy() == "publishAt" && req.GetOrder() == "asc",
		PublishAtDesc: req.GetOrderBy() == "publishAt" && req.GetOrder() == "desc",
		AnyStatus:     req.GetStatus() == "",
		Status:        req.GetStatus(),
		AnyKeyword:    req.GetKeyword() == "",
		Keyword:       "%" + req.GetKeyword() + "%",
	}
	return params, nil
}

// -------------------------------------------------------------------
// StarPost
func (server *Server) StarPost(ctx context.Context, req *pb.StarPostRequest) (*emptypb.Empty, error) {
	authUser, ok := ctx.Value(authUserKey{}).(AuthUser)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get auth user")
	}

	postID := req.GetPostId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	if req.GetIsLike() {
		arg := sqlc.CreatePostStarParams{
			PostID: postID,
			UserID: authUser.ID,
		}
		err := server.store.CreatePostStar(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create post star")
		}
	} else {
		arg := sqlc.DeletePostStarParams{
			PostID: postID,
			UserID: authUser.ID,
		}
		err := server.store.DeletePostStar(ctx, arg)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to delete post star")
		}
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// FetchPosts
func (server *Server) FetchPosts(ctx context.Context, req *pb.FetchPostsRequest) (*pb.FetchPostsResponse, error) {
	var authUser AuthUser
	if user, ok := ctx.Value(authUserKey{}).(AuthUser); ok {
		authUser = user
	}

	arg, err := parseFetchPostsRequest(authUser, req)
	if err != nil {
		return nil, err
	}

	posts, err := server.store.FetchPosts(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list posts")
	}

	rsp := convertFetchPosts(posts)
	return rsp, nil
}

func parseFetchPostsRequest(user AuthUser, req *pb.FetchPostsRequest) (*sqlc.FetchPostsParams, error) {
	options := []string{"viewCount", "publishAt"}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.FetchPostsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		SelfID:        user.ID,
		ViewCountAsc:  req.GetOrderBy() == "viewCount" && req.GetOrder() == "asc",
		ViewCountDesc: req.GetOrderBy() == "viewCount" && req.GetOrder() == "desc",
		PublishAtAsc:  req.GetOrderBy() == "publishAt" && req.GetOrder() == "asc",
		PublishAtDesc: req.GetOrderBy() == "publishAt" && req.GetOrder() == "desc",
		AnyFeatured:   !req.GetIsFeatured(),
		IsFeatured:    req.GetIsFeatured(),
		AnyAuthor:     req.GetAuthorId() == 0,
		AuthorID:      req.GetAuthorId(),
		AnyCategory:   req.GetCategoryId() == 0,
		CategoryID:    req.GetCategoryId(),
		AnyTag:        req.GetTagId() == 0,
		TagID:         req.GetTagId(),
		AnyKeyword:    req.GetKeyword() == "",
		Keyword:       "%" + req.GetKeyword() + "%",
	}
	return params, nil
}

// -------------------------------------------------------------------
// ReadPost
func (server *Server) ReadPost(ctx context.Context, req *pb.ReadPostRequest) (*pb.ReadPostResponse, error) {
	var authUser AuthUser
	if user, ok := ctx.Value(authUserKey{}).(AuthUser); ok {
		authUser = user
	}

	postID := req.GetPostId()
	if err := util.ValidateID(postID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "postID: %s", err.Error())
	}

	arg := sqlc.ReadPostParams{
		PostID: postID,
		SelfID: authUser.ID,
	}
	post, err := server.store.ReadPost(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to read post")
	}

	rsp := convertReadPost(post)
	return rsp, nil
}
