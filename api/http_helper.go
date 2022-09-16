package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"

	"google.golang.org/grpc/codes"
)

// -------------------------------------------------------------------
// Define of HttpError
type HttpError struct {
	Code    int32         `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Details []interface{} `json:"details,omitempty"`
}

func (h *HttpError) Error() string {
	return h.Message
}

func (h *HttpError) Status() int {
	switch h.Code {
	case int32(codes.InvalidArgument):
		return http.StatusBadRequest
	case int32(codes.NotFound):
		return http.StatusNotFound
	case int32(codes.Internal):
		return http.StatusInternalServerError
	case int32(codes.Unimplemented):
		return http.StatusNotImplemented
	case int32(codes.Unauthenticated):
		return http.StatusUnauthorized
	case int32(codes.PermissionDenied):
		return http.StatusForbidden
	default:
		return 200
	}
}

func (h *HttpError) Marshal() ([]byte, error) {
	return json.Marshal(h)
}

func NewHttpError(code codes.Code, err string, detail ...interface{}) *HttpError {
	return &HttpError{
		Code:    int32(code),
		Message: err,
		Details: detail,
	}
}

// -------------------------------------------------------------------
// replies to the request with the specified error message and HTTP code.
func writeHttpError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(*HttpError); ok {
		jsonRsp, marshalErr := httpErr.Marshal()
		if marshalErr == nil {
			status := httpErr.Status()
			w.WriteHeader(status)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Write(jsonRsp)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}
		http.Error(w, httpErr.Message, httpErr.Status())
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// writes the data to the response as part of an HTTP reply.
func writeHttpResponse(w http.ResponseWriter, data interface{}) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	jsonResp, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error happened in json marshal")
	}
	w.Write(jsonResp)
	return nil
}

// -------------------------------------------------------------------
// Extract access token from request header
func (server *Server) extractTokenFromHeader(r *http.Request) (string, error) {
	values := r.Header.Get(authorizationHeaderKey)
	if len(values) == 0 {
		return "", NewHttpError(codes.Unauthenticated, "authorization token is not provided")
	}

	fields := strings.Split(values, " ")
	if len(fields) != 2 {
		return "", NewHttpError(codes.Unauthenticated, "invalid authorization header format")
	}

	authorizationType := strings.ToLower(fields[0])
	if authorizationType != authorizationTypeBearer {
		return "", NewHttpError(codes.Unauthenticated, "unsupported authorization type")
	}

	return fields[1], nil
}

// authorization interceptor for http
func (server *Server) authorize(r *http.Request, method string) (*db.User, error) {
	log.Println("call authorize: ", method)

	allowedRoles, ok := server.allowedRoles[method]
	if !ok {
		return nil, NewHttpError(codes.Unimplemented, "method not implemented")
	}

	accessToken, err := server.extractTokenFromHeader(r)
	if err != nil {
		return nil, err
	}

	accessPayload, err := server.tokenMaker.VerifyToker(accessToken)
	if err != nil {
		if err == util.ErrExpiredToken {
			return nil, NewHttpError(codes.Unauthenticated, err.Error(), &pb.RefreshInfo{Refreshable: true})
		}
		return nil, NewHttpError(codes.Unauthenticated, err.Error())
	}

	currUser, err := server.store.GetUser(r.Context(), accessPayload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewHttpError(codes.NotFound, "user not found")
		}
		return nil, NewHttpError(codes.Internal, "failed to get user")
	}

	if currUser.Deleted {
		return nil, NewHttpError(codes.NotFound, "this user is inactive")
	}

	for _, role := range allowedRoles {
		if role == currUser.Role {
			return &currUser, nil
		}
	}

	return nil, NewHttpError(codes.PermissionDenied, "no permission to access")
}
