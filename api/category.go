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
// CreateCategory
func (server *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	name := req.GetName()
	if err := util.ValidateString(name, 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name: %s", err.Error())
	}

	category, err := server.store.CreateCategory(ctx, name)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "categories_name_key":
				return nil, status.Errorf(codes.AlreadyExists, "category name already exists: %s", req.GetName())
			}
		}
		return nil, status.Error(codes.Internal, "failed to create category")
	}

	rsp := &pb.CreateCategoryResponse{Category: convertCategory(category)}
	return rsp, nil
}

// -------------------------------------------------------------------
// DeleteCategories
func (server *Server) DeleteCategories(ctx context.Context, req *pb.DeleteCategoriesRequest) (*emptypb.Empty, error) {
	categoryIDs := util.RemoveDuplicates(req.GetCategoryIds())
	for _, categoryID := range categoryIDs {
		if err := util.ValidateID(categoryID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "categoryId: %s", err.Error())
		}
	}

	nrows, err := server.store.DeleteCategories(ctx, categoryIDs)
	if err != nil || int64(len(categoryIDs)) != nrows {
		return nil, status.Error(codes.Internal, "failed to delete categories")
	}
	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------
// UpdateCategory
func (server *Server) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	arg, err := parseUpdateCategoryRequest(req)
	if err != nil {
		return nil, err
	}

	newCategory, err := server.store.UpdateCategory(ctx, *arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "categories_name_key":
				return nil, status.Errorf(codes.AlreadyExists, "category name already exists: %s", arg.Name)
			}
		}
		return nil, status.Error(codes.Internal, "failed to update category")
	}

	rsp := &pb.UpdateCategoryResponse{Category: convertCategory(newCategory)}
	return rsp, nil
}

func parseUpdateCategoryRequest(req *pb.UpdateCategoryRequest) (*sqlc.UpdateCategoryParams, error) {
	categoryID := req.GetCategoryId()
	if err := util.ValidateID(categoryID); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "categoryId: %s", err.Error())
	}

	name := req.GetName()
	if err := util.ValidateString(name, 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name: %s", err.Error())
	}

	arg := &sqlc.UpdateCategoryParams{
		ID:   categoryID,
		Name: name,
	}
	return arg, nil
}

// -------------------------------------------------------------------
// ListCategories
func (server *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	arg, err := parseListCategoriesRequest(req)
	if err != nil {
		return nil, err
	}

	categories, err := server.store.ListCategories(ctx, *arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	rsp := convertListCategories(categories)
	return rsp, nil
}

func parseListCategoriesRequest(req *pb.ListCategoriesRequest) (*sqlc.ListCategoriesParams, error) {
	options := []string{"name", ""}
	err := util.ValidateOrder(req, options)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := &sqlc.ListCategoriesParams{
		NameAsc:  req.GetOrderBy() == "name" && req.GetOrder() == "asc",
		NameDesc: req.GetOrderBy() == "name" && req.GetOrder() == "desc",
	}
	return params, nil
}

// -------------------------------------------------------------------
// GetCategories
func (server *Server) GetCategories(ctx context.Context, req *emptypb.Empty) (*pb.GetCategoriesResponse, error) {
	categories, err := server.store.GetCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get categories")
	}

	rsp := &pb.GetCategoriesResponse{
		Categories: convertCategories(categories),
	}
	return rsp, nil
}

// -------------------------------------------------------------------
// Utils for post-category

// setPostCategories
func (server *Server) setPostCategories(ctx context.Context, postID int64, categoryIDs []int64) ([]sqlc.Category, error) {
	arg := sqlc.SetPostCategoriesParams{
		PostID:      postID,
		CategoryIds: categoryIDs,
	}
	categories, err := server.store.SetPostCategories(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set post categories")
	}

	res := []sqlc.Category{}
	for _, category := range categories {
		res = append(res, sqlc.Category(category))
	}
	return res, nil
}
