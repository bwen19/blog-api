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
)

// Handler of file upload
func (server *Server) HandleFileUpload(w http.ResponseWriter, r *http.Request, params map[string]string) {
	param, ok := params["suffix"]
	if !ok {
		httpError(w, http.StatusBadRequest, "invalid file upload url")
		return
	}

	switch param {
	case "avatar":
		server.UploadAvatar(w, r)
	case "post-image":
		server.UploadPostImage(w, r)
	default:
		httpError(w, http.StatusBadRequest, "invalid file upload suffix")
	}
}

// -------------------------------------------------------------------
// UploadAvatar
type UploadAvatarResponse struct {
	User *pb.User `json:"user,omitempty"`
}

// UploadAvatar
func (server *Server) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	authUser, gErr := server.httpGuard(r, roleUser)
	if gErr != nil {
		gErr.HttpErr(w)
		return
	}

	// maxMemory set as 1M for avatar
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		httpError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	f, _, err := r.FormFile("avatar")
	if err != nil {
		httpError(w, http.StatusBadRequest, "failed to get avatar from multipart form")
		return
	}
	defer f.Close()

	avatarSrc := path.Join(server.config.AvatarPath, util.RandomImageName())
	fileName := path.Join(server.config.PublicPath, avatarSrc)

	fn, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to save the new avatar file")
		return
	}
	defer fn.Close()
	io.Copy(fn, f)

	// update avatar src of user in database
	arg := db.UpdateUserParams{
		ID:     authUser.ID,
		Avatar: sql.NullString{String: avatarSrc, Valid: true},
	}

	user, err := server.store.UpdateUser(r.Context(), arg)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to update user avatar in db")
		return
	}

	if authUser.Avatar != server.config.DefaultAvatar {
		oldFile := path.Join(server.config.PublicPath, authUser.Avatar)
		if err = os.Remove(oldFile); err != nil {
			log.Println("failed to remove old avatar file")
		}
	}

	rsp := &UploadAvatarResponse{User: convertUser(user)}
	httpResponse(w, rsp)
}

// -------------------------------------------------------------------
// UploadPostImage
type UploadPostImageResponse struct {
	Image string `json:"image,omitempty"`
}

// UploadPostImage
func (server *Server) UploadPostImage(w http.ResponseWriter, r *http.Request) {
	if _, gErr := server.httpGuard(r, roleAuthor); gErr != nil {
		gErr.HttpErr(w)
		return
	}

	// maxMemory set as 5M for post image
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		httpError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	f, _, err := r.FormFile("image")
	if err != nil {
		httpError(w, http.StatusBadRequest, "failed to get image from multipart form")
		return
	}
	defer f.Close()

	imageSrc := path.Join(server.config.PostPath, util.RandomImageName())
	fileName := path.Join(server.config.PublicPath, imageSrc)

	fn, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to save the new image file")
		return
	}
	defer fn.Close()
	io.Copy(fn, f)

	rsp := &UploadPostImageResponse{Image: imageSrc}
	httpResponse(w, rsp)
}
