package api

import (
	db "blog/db/sqlc"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// =============================================================
// createTag
type createTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// @Router /api/admin/tags [post]
func (server *Server) createTag(ctx *gin.Context) {
	var req createTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	tag, err := server.store.CreateTag(ctx, req.Name)
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

	ctx.JSON(http.StatusOK, tag)
}

// =============================================================
// getTag
type getTagRequest struct {
	Name string `uri:"name" binding:"required"`
}

// @Router /api/admin/tags/:name [get]
func (server *Server) getTag(ctx *gin.Context) {
	var req getTagRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	tag, err := server.store.GetTag(ctx, req.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, tag)
}

// =============================================================
// listTags
type listTagsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=20"`
}

// @Router /api/admin/tags [get]
func (server *Server) listTags(ctx *gin.Context) {
	var req listTagsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListTagsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	tags, err := server.store.ListTags(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, tags)
}

// =============================================================
// updateTag
type updateTagUriRequest struct {
	Name string `uri:"name" binding:"required"`
}

type updateTagRequest struct {
	NewName string `json:"new_name"`
	Count   int64  `json:"count" binding:"gte=0"`
}

// @Router /api/admin/tags/:name [patch]
func (server *Server) updateTag(ctx *gin.Context) {
	var req1 updateTagUriRequest
	if err := ctx.ShouldBindUri(&req1); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req2 updateTagRequest
	if err := ctx.ShouldBindJSON(&req2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateTagParams{Name: req1.Name}
	if req2.NewName != "" {
		arg.SetNewName = true
		arg.NewName = req2.NewName
	}
	if req2.Count > 0 {
		arg.SetCount = true
		arg.Count = req2.Count
	}

	tag, err := server.store.UpdateTag(ctx, arg)
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

	ctx.JSON(http.StatusOK, tag)
}

// =============================================================
// deleteTag
type deleteTagRequest struct {
	Name string `uri:"name" binding:"required"`
}

// @Router /api/admin/tags/:name [delete]
func (server *Server) deleteTag(ctx *gin.Context) {
	var req deleteTagRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteTag(ctx, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"tag": req.Name})
}
