package api

import (
	"context"
	"database/sql"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"github.com/lib/pq"
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
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_voilation":
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
	tagIDs, err := util.ValidateRepeatedIDs(req.GetTagIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "tagId: %s", err.Error())
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
	if err := util.ValidateID(req.GetTagId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "tagId: %s", err.Error())
	}

	if err := util.ValidateString(req.GetName(), 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name: %s", err.Error())
	}

	arg := db.UpdateTagParams{
		ID:   req.GetTagId(),
		Name: req.GetName(),
	}

	newTag, err := server.store.UpdateTag(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_voilation":
				return nil, status.Errorf(codes.AlreadyExists, "tag name already exists: %s", arg.Name)
			}
		}
		return nil, status.Error(codes.Internal, "failed to update tag")
	}

	rsp := &pb.UpdateTagResponse{Tag: convertTag(newTag)}
	return rsp, nil
}

// -------------------------------------------------------------------
// ListTags
func (server *Server) ListTags(ctx context.Context, req *pb.ListTagsRequest) (*pb.ListTagsResponse, error) {
	options := []string{"name", "postCount", ""}
	if err := util.ValidatePageOrder(req, options); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListTagsParams{
		Limit:         req.GetPageSize(),
		Offset:        (req.GetPageId() - 1) * req.GetPageSize(),
		NameAsc:       req.GetOrderBy() == "name" && req.GetOrder() == "asc",
		NameDesc:      req.GetOrderBy() == "name" && req.GetOrder() == "desc",
		PostCountAsc:  req.GetOrderBy() == "postCount" && req.GetOrder() == "asc",
		PostCountDesc: req.GetOrderBy() == "postCount" && req.GetOrder() == "desc",
	}

	tags, err := server.store.ListTags(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list tags")
	}

	return convertListTags(tags), nil
}

// -------------------------------------------------------------------
// GetTag
func (server *Server) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.GetTagResponse, error) {
	if err := util.ValidateString(req.GetTagName(), 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "tagName: %s", err.Error())
	}

	var tag db.Tag
	tag, err := server.store.GetTagsByName(ctx, req.GetTagName())
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, "failed to get tag")
		}

		tag, err = server.store.CreateTag(ctx, req.GetTagName())
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create new tag")
		}
	}

	rsp := &pb.GetTagResponse{Tag: convertTag(tag)}
	return rsp, nil
}
