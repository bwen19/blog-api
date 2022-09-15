package api

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
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
	case "post-image":
		return "UploadPostImage", nil
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

	switch method {
	case "UploadAvatar":
		server.UploadAvatar(w, r, authUser)
	case "UploadPostImage":
		server.UploadPostImage(w, r, authUser)
	default:
		writeHttpError(w, NewHttpError(codes.Unimplemented, "method not implemented"))
	}
}

// -------------------------------------------------------------------
// UploadAvatar
type UploadAvatarResponse struct {
	User *pb.User `json:"user,omitempty"`
}

// UploadAvatar
func (server *Server) UploadAvatar(w http.ResponseWriter, r *http.Request, authUser *db.User) {
	// maxMemory set as 1M for avatar
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeHttpError(w, NewHttpError(codes.InvalidArgument, "failed to parse multipart form"))
		return
	}

	f, _, err := r.FormFile("avatar")
	if err != nil {
		writeHttpError(w, NewHttpError(codes.InvalidArgument, "failed to get avatar from multipart form"))
		return
	}
	defer f.Close()

	filename := util.RandomImageName(authUser.ID)
	avatar := path.Join(server.config.AvatarPath, filename)
	fullname := path.Join(server.config.PublicPath, avatar)

	fn, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, "failed to save the new avatar file"))
		return
	}
	defer fn.Close()
	io.Copy(fn, f)

	// update avatar src of user in database
	arg := db.UpdateUserParams{
		ID:     authUser.ID,
		Avatar: sql.NullString{String: avatar, Valid: true},
	}
	user, err := server.store.UpdateUser(r.Context(), arg)
	if err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, "failed to update user avatar in db"))
		return
	}

	if authUser.Avatar != server.config.DefaultAvatar {
		oldFile := path.Join(server.config.PublicPath, authUser.Avatar)
		if err = os.Remove(oldFile); err != nil {
			log.Println("failed to remove old avatar file")
		}
	}

	if err = writeHttpResponse(w, &UploadAvatarResponse{User: convertUser(user)}); err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, err.Error()))
	}
}

// -------------------------------------------------------------------
// UploadPostImage
type UploadPostImageResponse struct {
	Image string `json:"image,omitempty"`
}

// UploadPostImage
func (server *Server) UploadPostImage(w http.ResponseWriter, r *http.Request, authUser *db.User) {
	// maxMemory set as 5M for post image
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeHttpError(w, NewHttpError(codes.InvalidArgument, "failed to parse multipart form"))
		return
	}

	f, _, err := r.FormFile("image")
	if err != nil {
		writeHttpError(w, NewHttpError(codes.InvalidArgument, "failed to get image from multipart form"))
		return
	}
	defer f.Close()

	filename := util.RandomImageName(authUser.ID)
	image := path.Join(server.config.PostPath, filename)
	fullname := path.Join(server.config.PublicPath, image)

	fn, err := os.OpenFile(fullname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, "failed to save the new image file"))
		return
	}
	defer fn.Close()
	io.Copy(fn, f)

	if err = writeHttpResponse(w, &UploadPostImageResponse{Image: image}); err != nil {
		writeHttpError(w, NewHttpError(codes.Internal, err.Error()))
	}
}
