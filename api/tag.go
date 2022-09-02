package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"

	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// -------------------------------------------------------------------
// CreateTag
func (server *Server) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.CreateTagResponse, error) {
	name := req.GetName()
	if err := util.ValidateString(name, 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name: %s", err.Error())
	}

	tag, err := server.store.CreateTag(ctx, name)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "tags_name_key":
				return nil, status.Errorf(codes.AlreadyExists, "tag name already exists: %s", req.GetName())
			}
		}
		return nil, status.Error(codes.Internal, "failed to create tag")
	}

	rsp := &pb.CreateTagResponse{Tag: convertTag(tag)}
	return rsp, nil
}

// -------------------------------------------------------------------
// DeleteTags
func (server *Server) DeleteTags(ctx context.Context, req *pb.DeleteTagsRequest) (*emptypb.Empty, error) {
	tagIDs := util.RemoveDuplicates(req.GetTagIds())
	for _, tagID := range tagIDs {
		if err := util.ValidateID(tagID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "tagId: %s", err.Error())
		}
	}

	nrows, err := server.store.DeleteTags(ctx, tagIDs)
	if err != nil || int64(len(tagIDs)) != nrows {
		return nil, status.Error(codes.Internal, "failed to delete tags")
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// UpdateTag
func (server *Server) UpdateTag(ctx context.Context, req *pb.UpdateTagRequest) (*pb.UpdateTagResponse, error) {
	arg, err := parseUpdateTagRequest(req)
	if err != nil {
		return nil, err
	}

	newTag, err := server.store.UpdateTag(ctx, *arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "tags_name_key":
				return nil, status.Errorf(codes.AlreadyExists, "tag name already exists: %s", arg.Name)
			}
		}
		return nil, status.Error(codes.Internal, "failed to update tag")
	}

	rsp := &pb.UpdateTagResponse{Tag: convertTag(newTag)}
	return rsp, nil
}

func parseUpdateTagRequest(req *pb.UpdateTagRequest) (*sqlc.UpdateTagParams, error) {
	tagID := req.GetTagId()
	if err := util.ValidateID(tagID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "tagId: %s", err.Error())
	}

	name := req.GetName()
	if err := util.ValidateString(name, 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name: %s", err.Error())
	}

	arg := &sqlc.UpdateTagParams{
		ID:   tagID,
		Name: name,
	}
	return arg, nil
}

// -------------------------------------------------------------------
// ListTags
func (server *Server) ListTags(ctx context.Context, req *pb.ListTagsRequest) (*pb.ListTagsResponse, error) {
	arg, err := parseListTagsRequest(req)
	if err != nil {
		return nil, err
	}

	tags, err := server.store.ListTags(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list tags")
	}

	rsp := convertListTags(tags)
	return rsp, nil
}

func parseListTagsRequest(req *pb.ListTagsRequest) (*sqlc.ListTagsParams, error) {
	options := []string{"name", "postCount", ""}
	err := util.ValidatePageOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.ListTagsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		NameAsc:       req.GetOrderBy() == "name" && req.GetOrder() == "asc",
		NameDesc:      req.GetOrderBy() == "name" && req.GetOrder() == "desc",
		PostCountAsc:  req.GetOrderBy() == "postCount" && req.GetOrder() == "asc",
		PostCountDesc: req.GetOrderBy() == "postCount" && req.GetOrder() == "desc",
	}
	return params, nil
}

// -------------------------------------------------------------------
// Utils for post-tag

// setPostTags
func (server *Server) setPostTags(ctx context.Context, postID int64, tagNames []string) ([]sqlc.Tag, error) {
	names := util.RemoveDuplicates(tagNames)
	oldTags, err := server.store.GetTagsByNames(ctx, names)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tags by names")
	}

	newNames := []string{}
	for _, name := range names {
		flag := true
		for _, tag := range oldTags {
			if tag.Name == name {
				flag = false
				break
			}
		}
		if flag {
			newNames = append(newNames, name)
		}
	}

	// create tags that not exist in database
	if len(newNames) > 0 {
		_, err = server.store.CreateTags(ctx, newNames)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create new tags")
		}
	}

	arg := sqlc.SetPostTagsParams{
		PostID:   postID,
		TagNames: names,
	}
	tags, err := server.store.SetPostTags(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get post tags")
	}

	res := []sqlc.Tag{}
	for _, tag := range tags {
		res = append(res, sqlc.Tag(tag))
	}
	return res, nil
}
