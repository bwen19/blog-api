package api

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

// Authorization guard for GRPC APIs
func (server *Server) grpcGuard(ctx context.Context, roleRank int) (*db.User, *GuardError) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md[authorizationHeaderKey]; len(values) > 0 {
			return server.authorize(ctx, values[0], roleRank)
		}
	}

	// Some GRPC APIs can be accessed without authorization
	if roleRank == roleGhost {
		return &db.User{}, nil
	}
	return nil, NewGuardError("authorization token is not provided", false)
}

// Authorization guard for HTTP APIs
func (server *Server) httpGuard(r *http.Request, roleRank int) (*db.User, *GuardError) {
	if values := r.Header.Get(authorizationHeaderKey); len(values) > 0 {
		return server.authorize(r.Context(), values, roleRank)
	}
	return nil, NewGuardError("authorization token is not provided", false)
}

// Perform authentication for input token with special role rank
func (server *Server) authorize(ctx context.Context, token string, roleRank int) (*db.User, *GuardError) {
	fields := strings.Fields(token)
	if len(fields) != 2 {
		return nil, NewGuardError("invalid authorization header format", false)
	}

	authorizationType := strings.ToLower(fields[0])
	if authorizationType != authorizationTypeBearer {
		return nil, NewGuardError("unsupported authorization type", false)
	}

	accessPayload, err := server.tokenMaker.VerifyToker(fields[1])
	if err != nil {
		if err == util.ErrExpiredToken {
			return nil, NewGuardError(err.Error(), true)
		}
		return nil, NewGuardError(err.Error(), false)
	}

	user, err := server.store.GetUser(ctx, accessPayload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewGuardError("user not found", false)
		}
		return nil, NewGuardError("failed to get user", false)
	}

	if user.Deleted {
		return nil, NewGuardError("this user is inactive", false)
	}

	if roleToRank(user.Role) < roleRank {
		return nil, NewGuardError("no permission to access", false)
	}

	return &user, nil
}

// Definition of role rank
const (
	roleGhost  = 0
	roleUser   = 1
	roleAuthor = 2
	roleAdmin  = 3
)

func roleToRank(role string) int {
	switch role {
	case "user":
		return roleUser
	case "author":
		return roleAuthor
	case "admin":
		return roleAdmin
	default:
		return roleGhost
	}
}

// ========================// GuardError //======================== //

type GuardError struct {
	Message     string `json:"message,omitempty"`
	Refreshable bool   `json:"refreshable,omitempty"`
}

func (x *GuardError) GrpcErr() error {
	if x.Refreshable {
		st := status.New(codes.Unauthenticated, x.Message)
		stDetails, err := st.WithDetails(&pb.RefreshInfo{Refreshable: true})
		if err != nil {
			return st.Err()
		}
		return stDetails.Err()
	}
	return status.Errorf(codes.Unauthenticated, x.Message)
}

func (x *GuardError) HttpErr(w http.ResponseWriter) {
	err := &ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: x.Message,
	}
	if x.Refreshable {
		err.Details = append(err.Details, &pb.RefreshInfo{Refreshable: true})
	}
	writeResponse(w, err.Code, err)
}

func NewGuardError(err string, refreshable bool) *GuardError {
	return &GuardError{
		Message:     err,
		Refreshable: refreshable,
	}
}
