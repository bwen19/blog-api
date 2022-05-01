package api

import (
	db "blog/server/db/sqlc"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type commentList = db.ListCommentsByArticleRow

type commentResponse struct {
	ID        int64     `json:"id"`
	ParentID  int64     `json:"parent_id"`
	ArticleID int64     `json:"article_id"`
	Commenter string    `json:"commenter"`
	AvatarSrc string    `json:"avatar_src"`
	Content   string    `json:"content"`
	CommentAt time.Time `json:"comment_at"`
}

func newCommentResponse1(comment *db.Comment, commenter *db.User) commentResponse {
	return commentResponse{
		ID:        comment.ID,
		ParentID:  comment.ParentID.Int64,
		ArticleID: comment.ArticleID,
		Commenter: commenter.Username,
		AvatarSrc: commenter.AvatarSrc,
		Content:   comment.Content,
		CommentAt: comment.CommentAt,
	}
}

func newCommentResponse2(comment *commentList) commentResponse {
	return commentResponse{
		ID:        comment.ID,
		ParentID:  comment.ParentID.Int64,
		ArticleID: comment.ArticleID,
		Commenter: comment.Commenter,
		AvatarSrc: comment.AvatarSrc,
		Content:   comment.Content,
		CommentAt: comment.CommentAt,
	}
}

// =============================================================
// listComments
type listCommentsRequest struct {
	PageID    int32 `form:"page_id" binding:"required,min=1"`
	PageSize  int32 `form:"page_size" binding:"required,min=5,max=20"`
	ArticleID int64 `form:"article_id" binding:"required,min=1"`
}

// @Router /api/comments [get]
func (server *Server) listArticleComments(ctx *gin.Context) {
	var req listCommentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCommentsByArticleParams{
		Limit:     req.PageSize,
		Offset:    (req.PageID - 1) * req.PageSize,
		ArticleID: req.ArticleID,
	}

	comments, err := server.store.ListCommentsByArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var rsp []commentResponse
	var commentIDs []int64
	for _, comment := range comments {
		commentIDs = append(commentIDs, comment.ID)
		rsp = append(rsp, newCommentResponse2(&comment))
	}

	childComments, err := server.store.ListChildComments(ctx, commentIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	for _, comment := range childComments {
		cm := commentList(comment)
		rsp = append(rsp, newCommentResponse2(&cm))
	}

	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// createComment
type createCommentRequest struct {
	ParentID  int64  `json:"parent_id" binding:"omitempty,min=1"`
	ArticleID int64  `json:"article_id" binding:"required,min=1"`
	Content   string `json:"content" binding:"required,min=1"`
}

// @Router /api/user/comments [post]
func (server *Server) createComment(ctx *gin.Context) {
	var req createCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.CreateCommentParams{
		ArticleID: req.ArticleID,
		Commenter: authUser.Username,
		Content:   req.Content,
	}
	if req.ParentID > 0 {
		arg.SetParentID = true
		arg.ParentID = req.ParentID
	}

	comment, err := server.store.CreateComment(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newCommentResponse1(&comment, authUser)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// deleteComment
type deleteCommentRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// @Router /api/user/comments/:id [delete]
func (server *Server) deleteComment(ctx *gin.Context) {
	var req deleteCommentRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.DeleteCommentParams{ID: req.ID}
	if authUser.Role == "admin" {
		arg.AnyCommenter = true
	} else {
		arg.Commenter = authUser.Username
	}

	err := server.store.DeleteComment(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": req.ID})
}
