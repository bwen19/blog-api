package api

import (
	db "blog/server/db/sqlc"
	"blog/server/token"
	"blog/server/util"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("role", validRole)
		v.RegisterValidation("status", validStatus)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/register", server.registerUser)
		api.POST("/login", server.loginUser)
		api.POST("/renew_access", server.renewAccessToken)
		api.GET("/articles/:id", server.readPublishedArticle)
		api.GET("/articles", server.listPublishedArticles)
		api.GET("/comments", server.listArticleComments)
	}

	roles := []string{"admin", "author", "user"}
	user := router.Group("/api/user").Use(authMiddleware(server, roles))
	{
		user.GET("", server.getUserSelf)
		user.PATCH("", server.updateUserSelf)
		user.POST("/comments", server.createComment)
		user.DELETE("/comments/:id", server.deleteComment)
	}

	roles = []string{"admin", "author"}
	author := router.Group("/api/author").Use(authMiddleware(server, roles))
	{
		author.POST("/articles", server.createArticleByAuthor)
		author.GET("/articles/:id", server.getArticleByAuthor)
		author.GET("/articles", server.listArticlesByAuthor)
		author.PATCH("/articles/:id", server.updateArticleByAuthor)
		author.DELETE("/articles/:id", server.deleteArticleByAuthor)
	}

	roles = []string{"admin"}
	admin := router.Group("/api/admin").Use(authMiddleware(server, roles))
	{
		admin.GET("/users/:username", server.getUser)
		admin.GET("/users", server.listUsers)
		admin.PATCH("/users/:username", server.updateUser)
		admin.DELETE("/users/:username", server.deleteUser)

		admin.GET("/articles/:id", server.getArticle)
		admin.GET("/articles", server.listArticles)
		admin.PATCH("/articles/:id", server.updateArticle)
		admin.DELETE("/articles/:id", server.deleteArticle)

		admin.POST("/categories", server.createCategory)
		admin.GET("/categories/:name", server.getCategory)
		admin.GET("/categories", server.listCategories)
		admin.PATCH("/categories/:name", server.updateCategory)
		admin.DELETE("/categories/:name", server.deleteCategory)

		admin.POST("/tags", server.createTag)
		admin.GET("/tags/:name", server.getTag)
		admin.GET("/tags", server.listTags)
		admin.PATCH("/tags/:name", server.updateTag)
		admin.DELETE("/tags/:name", server.deleteTag)
	}

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
