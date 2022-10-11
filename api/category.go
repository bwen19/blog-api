package api

import (
	"context"
	"fmt"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ========================// CreateCategory //======================== //

func (server *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := util.ValidateString(req.GetName(), 1, 50); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "name: %s", err.Error())
	}

	category, err := server.store.CreateCategory(ctx, req.GetName())
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_voilation":
				return nil, status.Errorf(codes.AlreadyExists, "category name already exists: %s", req.GetName())
			}
		}
		return nil, status.Error(codes.Internal, "failed to create category")
	}

	rsp := &pb.CreateCategoryResponse{Category: convertCategory(category)}
	return rsp, nil
}

// ========================// DeleteCategories //======================== //

func (server *Server) DeleteCategories(ctx context.Context, req *pb.DeleteCategoriesRequest) (*emptypb.Empty, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	categoryIDs, err := util.ValidateRepeatedIDs(req.GetCategoryIds())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "categoryId: %s", err.Error())
	}

	nrows, err := server.store.DeleteCategories(ctx, categoryIDs)
	if err != nil || len(categoryIDs) != int(nrows) {
		return nil, status.Error(codes.Internal, "failed to delete categories")
	}

	return &emptypb.Empty{}, nil
}

// ========================// UpdateCategory //======================== //

func (server *Server) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	if err := validateUpdateCategoryRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.UpdateCategoryParams{
		ID:   req.GetCategoryId(),
		Name: req.GetName(),
	}

	newCategory, err := server.store.UpdateCategory(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_voilation":
				return nil, status.Errorf(codes.AlreadyExists, "category name already exists: %s", arg.Name)
			}
		}
		return nil, status.Error(codes.Internal, "failed to update category")
	}

	rsp := &pb.UpdateCategoryResponse{Category: convertCategory(newCategory)}
	return rsp, nil
}

func validateUpdateCategoryRequest(req *pb.UpdateCategoryRequest) error {
	if err := util.ValidateID(req.GetCategoryId()); err != nil {
		return fmt.Errorf("categoryId: %s", err.Error())
	}
	if err := util.ValidateString(req.GetName(), 1, 50); err != nil {
		return fmt.Errorf("name: %s", err.Error())
	}
	return nil
}

// ========================// ListCategories //======================== //

func (server *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	if _, gErr := server.grpcGuard(ctx, roleAdmin); gErr != nil {
		return nil, gErr.GrpcErr()
	}

	options := []string{"name", ""}
	if err := util.ValidateOrder(req, options); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	arg := db.ListCategoriesParams{
		NameAsc:  req.GetOrderBy() == "name" && req.GetOrder() == "asc",
		NameDesc: req.GetOrderBy() == "name" && req.GetOrder() == "desc",
	}

	categories, err := server.store.ListCategories(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list categories")
	}
	return convertListCategories(categories), nil
}

// ========================// GetCategories //======================== //

func (server *Server) GetCategories(ctx context.Context, req *emptypb.Empty) (*pb.GetCategoriesResponse, error) {
	categories, err := server.store.GetCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get categories")
	}

	rsp := &pb.GetCategoriesResponse{Categories: convertCategories(categories)}
	return rsp, nil
}
