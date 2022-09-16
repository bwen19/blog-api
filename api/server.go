package api

import (
	"fmt"

	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"github.com/bwen19/blog/util"
)

// Server serves gRPC requests for blog service
type Server struct {
	pb.UnimplementedBlogServer
	config       util.Config
	store        db.Store
	tokenMaker   util.TokenMaker
	allowedRoles map[string][]string
}

// Create a new gRPC server
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := util.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:       config,
		store:        store,
		tokenMaker:   tokenMaker,
		allowedRoles: allowedRoles(),
	}
	return server, nil
}

func allowedRoles() map[string][]string {
	const blogBasePath = "/pb.Blog/"

	return map[string][]string{
		"UploadAvatar":                         {"admin", "author", "user"},
		"UploadPostImage":                      {"admin", "author"},
		blogBasePath + "Register":              {"any"},
		blogBasePath + "Login":                 {"any"},
		blogBasePath + "AutoLogin":             {"any"},
		blogBasePath + "Refresh":               {"any"},
		blogBasePath + "Logout":                {"any"},
		blogBasePath + "DeleteSessions":        {"admin", "author", "user"},
		blogBasePath + "DeleteExpiredSessions": {"admin"},
		blogBasePath + "ListSessions":          {"admin", "author", "user"},
		blogBasePath + "CreateUser":            {"admin"},
		blogBasePath + "DeleteUsers":           {"admin"},
		blogBasePath + "UpdateUser":            {"admin"},
		blogBasePath + "ListUsers":             {"admin"},
		blogBasePath + "ChangeProfile":         {"admin", "author", "user"},
		blogBasePath + "ChangePassword":        {"admin", "author", "user"},
		blogBasePath + "GetUserProfile":        {"any", "admin", "author", "user"},
		blogBasePath + "MarkAllRead":           {"admin", "author", "user"},
		blogBasePath + "LeaveMessage":          {"admin", "author", "user"},
		blogBasePath + "DeleteNotifs":          {"admin", "author", "user"},
		blogBasePath + "ListNotifs":            {"admin", "author", "user"},
		blogBasePath + "ListMessages":          {"admin"},
		blogBasePath + "CheckMessages":         {"admin"},
		blogBasePath + "FollowUser":            {"admin", "author", "user"},
		blogBasePath + "ListFollows":           {"admin", "author", "user"},
		blogBasePath + "CreatePost":            {"admin", "author"},
		blogBasePath + "DeletePost":            {"admin", "author"},
		blogBasePath + "UpdatePost":            {"admin", "author"},
		blogBasePath + "SubmitPost":            {"admin", "author"},
		blogBasePath + "PublishPost":           {"admin"},
		blogBasePath + "WithdrawPost":          {"admin"},
		blogBasePath + "UpdatePostLabel":       {"admin"},
		blogBasePath + "ListPosts":             {"admin", "author"},
		blogBasePath + "GetPost":               {"admin", "author"},
		blogBasePath + "GetPosts":              {"any", "admin", "author", "user"},
		blogBasePath + "ReadPost":              {"any", "admin", "author", "user"},
		blogBasePath + "StarPost":              {"admin", "author", "user"},
		blogBasePath + "CreateCategory":        {"admin"},
		blogBasePath + "DeleteCategories":      {"admin"},
		blogBasePath + "UpdateCategory":        {"admin"},
		blogBasePath + "ListCategories":        {"admin"},
		blogBasePath + "GetCategories":         {"any"},
		blogBasePath + "CreateTag":             {"admin"},
		blogBasePath + "DeleteTags":            {"admin"},
		blogBasePath + "UpdateTag":             {"admin"},
		blogBasePath + "ListTags":              {"admin"},
		blogBasePath + "GetTag":                {"admin", "author"},
		blogBasePath + "CreateComment":         {"admin", "author", "user"},
		blogBasePath + "DeleteComment":         {"admin", "author", "user"},
		blogBasePath + "ListComments":          {"any", "admin", "author", "user"},
		blogBasePath + "ListReplies":           {"any", "admin", "author", "user"},
		blogBasePath + "StarComment":           {"admin", "author", "user"},
	}
}
