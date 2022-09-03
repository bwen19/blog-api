package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"blog/server/util"
	"context"
	"database/sql"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// -------------------------------------------------------------------
// WrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type WrappedServerStream struct {
	grpc.ServerStream
	// WrappedContext is the wrapper's own Context. You can assign it.
	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *WrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

// WrapServerStream returns a ServerStream that has the ability to overwrite context.
func WrapServerStream(stream grpc.ServerStream) *WrappedServerStream {
	if existing, ok := stream.(*WrappedServerStream); ok {
		return existing
	}
	return &WrappedServerStream{ServerStream: stream, WrappedContext: stream.Context()}
}

// -------------------------------------------------------------------
// UnaryInterceptor returns a new unary server interceptors that performs per-request auth
func (server *Server) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		newCtx, err := server.authFunc(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// StreamInterceptor returns a new stream server interceptors that performs per-request auth
func (server *Server) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, err := server.authFunc(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		wrapped := WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

// -------------------------------------------------------------------

type authUserKey struct{}

type AuthUser struct {
	sqlc.User
}

// authFunc is the pluggable function that performs authentication
func (server *Server) authFunc(ctx context.Context, method string) (context.Context, error) {
	log.Println("call authorize: ", method)

	allowedRoles, ok := server.allowedRoles[method]
	if !ok {
		return ctx, status.Errorf(codes.Unimplemented, "method %s not implemented", method)
	}

	accessToken, err := server.extractTokenFromMeta(ctx)
	if err != nil {
		if allowedRoles[0] == "any" {
			return ctx, nil
		}
		return ctx, status.Error(codes.Unauthenticated, err.Error())
	}

	accessPayload, err := server.tokenMaker.VerifyToker(accessToken)
	if err != nil {
		if err == util.ErrExpiredToken {
			st := status.New(codes.Unauthenticated, err.Error())
			stDetail, err := st.WithDetails(&pb.RefreshInfo{Refreshable: true})
			if err != nil {
				return ctx, st.Err()
			}
			return ctx, stDetail.Err()
		}
		return ctx, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := server.store.GetUser(ctx, accessPayload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx, status.Error(codes.NotFound, "user not found")
		}
		return ctx, status.Error(codes.Internal, "failed to get user")
	}

	if user.IsDeleted {
		return ctx, status.Error(codes.NotFound, "this user is inactive")
	}

	for _, role := range allowedRoles {
		if role == user.Role {
			newCtx := context.WithValue(ctx, authUserKey{}, AuthUser{user})
			return newCtx, nil
		}
	}
	return ctx, status.Error(codes.PermissionDenied, "no permission to access")
}
