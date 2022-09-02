package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"google.golang.org/grpc/codes"
)

func parseUploadMethod(params map[string]string) (string, error) {
	param, ok := params["suffix"]
	if !ok {
		return "", NewHttpError(codes.NotFound, "invalid file upload url")
	}

	switch param {
	case "avatar":
		return "UploadAvatar", nil
	default:
		return "", NewHttpError(codes.NotFound, "invalid file upload suffix")
	}
}

// Handler of file upload
func (server *Server) HandleFileUpload(w http.ResponseWriter, r *http.Request, params map[string]string) {
	method, err := parseUploadMethod(params)
	if err != nil {
		writeHttpError(w, err)
		return
	}

	authUser, err := server.authorize(r, method)
	if err != nil {
		writeHttpError(w, err)
		return
	}

	if method == "UploadAvatar" {
		server.UploadAvatar(w, r, authUser)
	}
}

// -------------------------------------------------------------------
// UploadAvatar response
type UploadAvatarResponse struct {
	User *pb.User `json:"user,omitempty"`
}

// UploadAvatar
func (server *Server) UploadAvatar(w http.ResponseWriter, r *http.Request, authUser *sqlc.User) {
	// maxMemory set as 1M for avatar
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		writeHttpError(w, NewHttpError(codes.InvalidArgument, "failed to parse multipart form"))
		return
	}

	f, _, err := r.FormFile("avatar")
	if err != nil {
		writeHttpError(w, NewHttpError(codes.InvalidArgument, "failed to get avatar from multipart form"))
		return
	}
	defer f.Close()

	filename := fmt.Sprintf("%s/%d%d", server.config.AvatarPath, authUser.ID, time.Now().Unix())
	fullname := path.Join(server.config.PublicPath, filename)
	fn, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, "failed to save the new avatar file"))
		return
	}
	defer fn.Close()
	io.Copy(fn, f)

	// update avatar src of user in database
	arg := sqlc.UpdateUserParams{
		ID:        authUser.ID,
		SetAvatar: true,
		Avatar:    filename,
	}
	user, err := server.store.UpdateUser(r.Context(), arg)
	if err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, "failed to update user avatar in db"))
		return
	}

	if authUser.Avatar != server.config.AvatarPath+"/default" {
		oldFile := path.Join(server.config.PublicPath, authUser.Avatar)
		err = os.Remove(oldFile)
		if err != nil {
			log.Println("failed to remove old avatar file")
		}
	}

	err = writeHttpResponse(w, &UploadAvatarResponse{User: convertUser(user)})
	if err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, err.Error()))
	}
}
