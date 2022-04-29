package api

import (
	db "blog/db/sqlc"
	"blog/util"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Response types and utils
type userResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	AvatarSrc string    `json:"avatar_src"`
	CreateAt  time.Time `json:"create_at"`
}

func newUserResponse(user *db.User) userResponse {
	return userResponse{
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		AvatarSrc: user.AvatarSrc,
		CreateAt:  user.CreateAt,
	}
}

// =============================================================
// registerUser

type registerUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// @Router /api/register [post]
func (server *Server) registerUser(ctx *gin.Context) {
	var req registerUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		Email:          req.Email,
		AvatarSrc:      server.config.DefaultAvatarSrc,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(&user)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// loginUser
type loginUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

// @Router /api/login [post]
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.Username == "" && req.Email == "" {
		err := fmt.Errorf("invalid username or email")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg1 := db.GetUserParams{
		Username: req.Username,
		Email:    req.Email,
	}

	user, err := server.store.GetUser(ctx, arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg2 := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}
	session, err := server.store.CreateSession(ctx, arg2)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(&user),
	}
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// getUserSelf

// @Router /api/user [get]
func (server *Server) getUserSelf(ctx *gin.Context) {
	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	rsp := newUserResponse(authUser)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// updateUserSelf
type updateUserSelfRequest struct {
	Password  string `json:"password" binding:"omitempty,min=6"`
	Email     string `json:"email" binding:"omitempty,email"`
	AvatarSrc string `json:"avatar_src"`
}

// @Router /api/user [patch]
func (server *Server) updateUserSelf(ctx *gin.Context) {
	var req updateUserSelfRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.UpdateUserParams{Username: authUser.Username}
	if req.Password != "" {
		hashedPassword, err := util.HashPassword(req.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		arg.SetHashedPassword = true
		arg.HashedPassword = hashedPassword
	}
	if req.Email != "" {
		arg.SetEmail = true
		arg.Email = req.Email
	}
	if req.AvatarSrc != "" {
		arg.SetAvatarSrc = true
		arg.AvatarSrc = req.AvatarSrc
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(&user)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// getUser
type getUserRequest struct {
	Username string `uri:"username" binding:"required"`
}

// @Router /api/admin/users/:username [get]
func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetUserParams{
		Username: req.Username,
	}
	user, err := server.store.GetUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(&user)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// listUsers
type listUsersRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=20"`
}

// @Router /api/admin/users [get]
func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := []userResponse{}
	for _, user := range users {
		rsp = append(rsp, newUserResponse(&user))
	}
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// updateUser
type updateUserUriRequest struct {
	Username string `uri:"username" binding:"required"`
}

type updateUserRequest struct {
	NewName string `json:"new_name"`
	Role    string `json:"role" binding:"omitempty,role"`
}

// @Router /api/admin/users/:username [patch]
func (server *Server) updateUser(ctx *gin.Context) {
	var req1 updateUserUriRequest
	if err := ctx.ShouldBindUri(&req1); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req2 updateUserRequest
	if err := ctx.ShouldBindJSON(&req2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateUserParams{Username: req1.Username}
	if req2.NewName != "" {
		arg.SetNewName = true
		arg.NewName = req2.NewName
	}
	if req2.Role != "" {
		arg.SetRole = true
		arg.Role = req2.Role
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(&user)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// deleteUser
type deleteUserRequest struct {
	Username string `uri:"username" binding:"required"`
}

// deleteUser
// @Router /api/admin/users/:username [delete]
func (server *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteUser(ctx, req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": req.Username})
}
