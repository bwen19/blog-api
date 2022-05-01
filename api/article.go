package api

import (
	db "blog/server/db/sqlc"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleList = db.ListArticlesRow

type articleResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	ViewCount int64     `json:"view_count"`
	UpdateAt  time.Time `json:"update_at"`
	CreateAt  time.Time `json:"create_at"`
	Author    string    `json:"author"`
	AvatarSrc string    `json:"avatar_src"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
}

func newArticleResponse(article *db.Article, author *db.User, tags []string) articleResponse {
	return articleResponse{
		ID:        article.ID,
		Title:     article.Title,
		Summary:   article.Summary,
		Content:   article.Content,
		Status:    article.Status,
		ViewCount: article.ViewCount,
		UpdateAt:  article.UpdateAt,
		CreateAt:  article.CreateAt,
		Author:    author.Username,
		AvatarSrc: author.AvatarSrc,
		Category:  article.Category,
		Tags:      tags,
	}
}

type articleListResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Status    string    `json:"status"`
	ViewCount int64     `json:"view_count"`
	UpdateAt  time.Time `json:"update_at"`
	Author    string    `json:"author"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
}

func newArticleListResponse(article *ArticleList, tags []string) articleListResponse {
	return articleListResponse{
		ID:        article.ID,
		Title:     article.Title,
		Summary:   article.Summary,
		ViewCount: article.ViewCount,
		UpdateAt:  article.UpdateAt,
		Author:    article.Author,
		Category:  article.Category,
		Tags:      tags,
	}
}

// setArticleTags
func (server *Server) setArticleTags(ctx *gin.Context, articleID int64, tags []string) ([]string, error) {
	oldTags, err := server.store.ListArticleTags(ctx, articleID)
	if err != nil {
		return []string{}, err
	}

	tagsMap := make(map[string]int)
	for _, tag := range tags {
		tagsMap[tag] = 1
	}
	for _, oldTag := range oldTags {
		if _, ok := tagsMap[oldTag]; ok {
			tagsMap[oldTag] = 0
		} else {
			tagsMap[oldTag] = -1
		}
	}

	tags = []string{}
	for tag, value := range tagsMap {
		if value == 1 {
			tags = append(tags, tag)
			// check if tag exists in database, if not then create it
			_, err1 := server.store.GetTag(ctx, tag)
			if err1 != nil {
				if err1 == sql.ErrNoRows {
					_, err2 := server.store.CreateTag(ctx, tag)
					if err2 != nil {
						return tags, err2
					}
				} else {
					return tags, err1
				}
			}

			arg1 := db.CreateArticleTagParams{
				ArticleID: articleID,
				Tag:       tag,
			}
			_, err := server.store.CreateArticleTag(ctx, arg1)
			if err != nil {
				return tags, err
			}

			arg2 := db.UpdateTagParams{
				Name:     tag,
				AddCount: true,
			}
			_, err = server.store.UpdateTag(ctx, arg2)
			if err != nil {
				return tags, err
			}
		} else if value == -1 {
			arg1 := db.DeleteArticleTagParams{
				ArticleID: articleID,
				Tag:       tag,
			}
			err := server.store.DeleteArticleTag(ctx, arg1)
			if err != nil {
				return tags, err
			}

			arg2 := db.UpdateTagParams{
				Name:       tag,
				MinusCount: true,
			}
			_, err = server.store.UpdateTag(ctx, arg2)
			if err != nil {
				return tags, err
			}
		} else {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}

// =============================================================
// readPublishedArticle
type readPublishedArticleRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// @Router /api/articles/:id [get]
func (server *Server) readPublishedArticle(ctx *gin.Context) {
	var req readPublishedArticleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	article, err := server.store.ReadArticle(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.GetUserParams{
		Username: article.Author,
	}
	author, err := server.store.GetUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tags, err := server.store.ListArticleTags(ctx, article.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newArticleResponse(&article, &author, tags)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// listPublishedArticles
type listPublishedArticlesRequest struct {
	PageID   int32  `form:"page_id" binding:"required,min=1"`
	PageSize int32  `form:"page_size" binding:"required,min=5,max=20"`
	SortBy   string `form:"sort_by" binding:"required,oneof=time count"`
	Author   string `form:"author"`
	Category string `form:"category"`
	Tag      string `form:"tag"`
}

// @Router /api/articles [get]
func (server *Server) listPublishedArticles(ctx *gin.Context) {
	var req listPublishedArticlesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListArticlesParams{
		Limit:       req.PageSize,
		Offset:      (req.PageID - 1) * req.PageSize,
		Status:      "published",
		AnyAuthor:   true,
		AnyCategory: true,
		AnyTag:      true,
		TimeDesc:    false,
		CountDesc:   false,
	}

	if req.Author != "" {
		arg.AnyAuthor = false
		arg.Author = req.Author
	} else if req.Category != "" {
		arg.AnyCategory = false
		arg.Category = req.Category
	} else if req.Tag != "" {
		arg.AnyTag = false
		arg.Tag = req.Tag
	}

	if req.SortBy == "time" {
		arg.TimeDesc = true
	} else {
		arg.CountDesc = true
	}

	articles, err := server.store.ListArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var rsp []articleListResponse
	for _, article := range articles {
		tags, err := server.store.ListArticleTags(ctx, article.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		rsp = append(rsp, newArticleListResponse(&article, tags))
	}
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// createArticleByAuthor
type createArticleByAuthorRequest struct {
	Title    string   `json:"title" binding:"required"`
	Summary  string   `json:"summary"`
	Content  string   `json:"content" binding:"required"`
	Category string   `json:"category"`
	Tags     []string `json:"tags" binding:"max=5"`
}

// @Router /api/author/articles [post]
func (server *Server) createArticleByAuthor(ctx *gin.Context) {
	var req createArticleByAuthorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	if req.Category == "" {
		req.Category = "default"
	}
	_, err := server.store.GetCategory(ctx, req.Category)
	if err != nil {
		if err == sql.ErrNoRows && req.Category == "default" {
			_, err2 := server.store.CreateCategory(ctx, req.Category)
			if err2 != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err2))
				return
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	if req.Summary == "" {
		contentRune := []rune(req.Content)
		req.Summary = string(contentRune[:50])
	}

	arg := db.CreateArticleParams{
		Author:   authUser.Username,
		Category: req.Category,
		Title:    req.Title,
		Summary:  req.Summary,
		Content:  req.Content,
	}

	article, err := server.store.CreateArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tags, err := server.setArticleTags(ctx, article.ID, req.Tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newArticleResponse(&article, authUser, tags)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// deleteArticleByAuthor
type deleteArticleByAuthorRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteArticleByAuthor(ctx *gin.Context) {
	var req deleteArticleByAuthorRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.DeleteArticleParams{
		ID:     req.ID,
		Author: authUser.Username,
		Status: "draft",
	}

	err := server.store.DeleteArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
}

// =============================================================
// getArticleByAuthor
type getArticleByAuthorRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// @Router /api/author/articles/:id [get]
func (server *Server) getArticleByAuthor(ctx *gin.Context) {
	var req getArticleByAuthorRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.GetArticleParams{
		ID:     req.ID,
		Author: authUser.Username,
	}

	article, err := server.store.GetArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tags, err := server.store.ListArticleTags(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newArticleResponse(&article, authUser, tags)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// listArticlesByAuthor
type listArticlesByAuthorRequest struct {
	PageID   int32  `form:"page_id" binding:"required,min=1"`
	PageSize int32  `form:"page_size" binding:"required,min=5,max=20"`
	SortBy   string `form:"sort_by" binding:"required,oneof=time count"`
	Status   string `form:"status" binding:"omitempty,status"`
	Category string `form:"category"`
	Tag      string `form:"tag"`
}

// @Router /api/author/articles [get]
func (server *Server) listArticlesByAuthor(ctx *gin.Context) {
	var req listArticlesByAuthorRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.ListArticlesParams{
		Limit:       req.PageSize,
		Offset:      (req.PageID - 1) * req.PageSize,
		Author:      authUser.Username,
		AnyStatus:   true,
		AnyCategory: true,
		AnyTag:      true,
		TimeDesc:    false,
		CountDesc:   false,
	}

	if req.Status != "" {
		arg.AnyStatus = false
		arg.Status = req.Status
	}
	if req.Category != "" {
		arg.AnyCategory = false
		arg.Category = req.Category
	}
	if req.Tag != "" {
		arg.AnyTag = false
		arg.Tag = req.Tag
	}

	if req.SortBy == "time" {
		arg.TimeDesc = true
	} else {
		arg.CountDesc = true
	}

	articles, err := server.store.ListArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var rsp []articleListResponse
	for _, article := range articles {
		tags, err := server.store.ListArticleTags(ctx, article.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		rsp = append(rsp, newArticleListResponse(&article, tags))
	}
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// updateArticleByAuthor
type updateArticleByAuthorUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateArticleByAuthorRequest struct {
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Content  string   `json:"content"`
	Submit   bool     `json:"submit"`
	Category string   `json:"category"`
	Tags     []string `json:"tags" binding:"omitempty,max=5"`
}

// @Router /api/author/articles/:id [patch]
func (server *Server) updateArticleByAuthor(ctx *gin.Context) {
	var req1 updateArticleByAuthorUriRequest
	if err := ctx.ShouldBindUri(&req1); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req2 updateArticleByAuthorRequest
	if err := ctx.ShouldBindJSON(&req2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authUser := ctx.MustGet(authorizationUserKey).(*db.User)

	arg := db.UpdateArticleParams{
		ID:       req1.ID,
		Username: authUser.Username,
	}
	if req2.Category != "" {
		arg.SetCategory = true
		arg.Category = req2.Category
	}
	if req2.Title != "" {
		arg.SetTitle = true
		arg.Title = req2.Title
	}
	if req2.Summary != "" {
		arg.SetSummary = true
		arg.Summary = req2.Summary
	}
	if req2.Content != "" {
		arg.SetContent = true
		arg.Content = req2.Content
	}
	if req2.Submit {
		arg.SetStatus = true
		arg.Status = "review"
	}

	article, err := server.store.UpdateArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tags, err := server.setArticleTags(ctx, req1.ID, req2.Tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newArticleResponse(&article, authUser, tags)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// getArticle
type getArticleRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// @Router /api/admin/articles/:id [get]
func (server *Server) getArticle(ctx *gin.Context) {
	var req getArticleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg1 := db.GetArticleParams{
		ID:        req.ID,
		AnyAuthor: true,
	}

	article, err := server.store.GetArticle(ctx, arg1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg2 := db.GetUserParams{
		Username: article.Author,
	}
	author, err := server.store.GetUser(ctx, arg2)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tags, err := server.store.ListArticleTags(ctx, article.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newArticleResponse(&article, &author, tags)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// listArticles
type listArticlesRequest struct {
	PageID   int32  `form:"page_id" binding:"required,min=1"`
	PageSize int32  `form:"page_size" binding:"required,min=5,max=20"`
	SortBy   string `form:"sort_by" binding:"required,oneof=time count"`
	Author   string `form:"author"`
	Status   string `form:"status" binding:"omitempty,status"`
	Category string `form:"category"`
	Tag      string `form:"tag"`
}

// @Router /api/admin/articles [get]
func (server *Server) listArticles(ctx *gin.Context) {
	var req listArticlesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListArticlesParams{
		Limit:       req.PageSize,
		Offset:      (req.PageID - 1) * req.PageSize,
		AnyStatus:   true,
		AnyAuthor:   true,
		AnyCategory: true,
		AnyTag:      true,
		TimeDesc:    false,
		CountDesc:   false,
	}

	if req.Status != "" {
		arg.AnyStatus = false
		arg.Status = req.Status
	}
	if req.Author != "" {
		arg.AnyAuthor = false
		arg.Author = req.Author
	}
	if req.Category != "" {
		arg.AnyCategory = false
		arg.Category = req.Category
	}
	if req.Tag != "" {
		arg.AnyTag = false
		arg.Tag = req.Tag
	}

	if req.SortBy == "time" {
		arg.TimeDesc = true
	} else {
		arg.CountDesc = true
	}

	articles, err := server.store.ListArticles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var rsp []articleListResponse
	for _, article := range articles {
		tags, err := server.store.ListArticleTags(ctx, article.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		rsp = append(rsp, newArticleListResponse(&article, tags))
	}
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// updateArticle
type updateArticleUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateArticleRequest struct {
	Author   string   `json:"author"`
	Status   string   `json:"status" binding:"omitempty,status"`
	Category string   `json:"category"`
	Tags     []string `json:"tags" binding:"omitempty,max=5"`
}

// @Router /api/admin/articles/:id [patch]
func (server *Server) updateArticle(ctx *gin.Context) {
	var req1 updateArticleUriRequest
	if err := ctx.ShouldBindUri(&req1); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req2 updateArticleRequest
	if err := ctx.ShouldBindJSON(&req2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateArticleParams{ID: req1.ID}
	if req2.Author != "" {
		arg.SetAuthor = true
		arg.Author = req2.Author
	}
	if req2.Status != "" {
		arg.SetStatus = true
		arg.Status = req2.Status
	}
	if req2.Category != "" {
		arg.SetCategory = true
		arg.Category = req2.Category
	}

	article, err := server.store.UpdateArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg2 := db.GetUserParams{
		Username: article.Author,
	}
	author, err := server.store.GetUser(ctx, arg2)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tags, err := server.setArticleTags(ctx, article.ID, req2.Tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newArticleResponse(&article, &author, tags)
	ctx.JSON(http.StatusOK, rsp)
}

// =============================================================
// deleteArticle
type deleteArticleRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// @Router /api/admin/articles/:id [delete]
func (server *Server) deleteArticle(ctx *gin.Context) {
	var req deleteArticleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteArticleParams{
		ID:        req.ID,
		AnyAuthor: true,
		AnyStatus: true,
	}
	err := server.store.DeleteArticle(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": req.ID})
}
