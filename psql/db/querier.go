// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CreateCategory(ctx context.Context, name string) (Category, error)
	CreateComment(ctx context.Context, arg CreateCommentParams) (CreateCommentRow, error)
	CreateCommentStar(ctx context.Context, arg CreateCommentStarParams) error
	CreateFollow(ctx context.Context, arg CreateFollowParams) error
	CreateNotification(ctx context.Context, arg CreateNotificationParams) error
	CreatePost(ctx context.Context, arg CreatePostParams) (Post, error)
	CreatePostCategories(ctx context.Context, arg CreatePostCategoriesParams) ([]CreatePostCategoriesRow, error)
	CreatePostContent(ctx context.Context, arg CreatePostContentParams) (PostContent, error)
	CreatePostStar(ctx context.Context, arg CreatePostStarParams) error
	CreatePostTags(ctx context.Context, arg CreatePostTagsParams) ([]CreatePostTagsRow, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateTag(ctx context.Context, name string) (Tag, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteCategories(ctx context.Context, ids []int64) (int64, error)
	DeleteComment(ctx context.Context, arg DeleteCommentParams) error
	DeleteCommentStar(ctx context.Context, arg DeleteCommentStarParams) error
	DeleteExpiredSessions(ctx context.Context) error
	DeleteFollow(ctx context.Context, arg DeleteFollowParams) error
	DeleteMessages(ctx context.Context, ids []int64) (int64, error)
	DeleteNotifications(ctx context.Context, arg DeleteNotificationsParams) (int64, error)
	DeletePost(ctx context.Context, arg DeletePostParams) error
	DeletePostCategories(ctx context.Context, arg DeletePostCategoriesParams) error
	DeletePostStar(ctx context.Context, arg DeletePostStarParams) error
	DeletePostTags(ctx context.Context, arg DeletePostTagsParams) error
	DeleteSession(ctx context.Context, arg DeleteSessionParams) error
	DeleteSessions(ctx context.Context, arg DeleteSessionsParams) (int64, error)
	DeleteTags(ctx context.Context, ids []int64) (int64, error)
	DeleteUsers(ctx context.Context, ids []int64) (int64, error)
	GetCategories(ctx context.Context) ([]Category, error)
	GetFeaturedPosts(ctx context.Context, limit int32) ([]GetFeaturedPostsRow, error)
	GetPost(ctx context.Context, arg GetPostParams) (GetPostRow, error)
	GetPostCategories(ctx context.Context, postID int64) ([]Category, error)
	GetPostTags(ctx context.Context, postID int64) ([]Tag, error)
	GetPosts(ctx context.Context, arg GetPostsParams) ([]GetPostsRow, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	GetTagsByName(ctx context.Context, name string) (Tag, error)
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	GetUser(ctx context.Context, id int64) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	GetUserProfile(ctx context.Context, arg GetUserProfileParams) (GetUserProfileRow, error)
	ListCategories(ctx context.Context, arg ListCategoriesParams) ([]ListCategoriesRow, error)
	ListComments(ctx context.Context, arg ListCommentsParams) ([]ListCommentsRow, error)
	ListFollowers(ctx context.Context, arg ListFollowersParams) ([]ListFollowersRow, error)
	ListFollowings(ctx context.Context, arg ListFollowingsParams) ([]ListFollowingsRow, error)
	ListMessages(ctx context.Context, arg ListMessagesParams) ([]ListMessagesRow, error)
	ListNotifications(ctx context.Context, arg ListNotificationsParams) ([]ListNotificationsRow, error)
	ListPosts(ctx context.Context, arg ListPostsParams) ([]ListPostsRow, error)
	ListReplies(ctx context.Context, arg ListRepliesParams) ([]ListRepliesRow, error)
	ListSessions(ctx context.Context, arg ListSessionsParams) ([]ListSessionsRow, error)
	ListTags(ctx context.Context, arg ListTagsParams) ([]ListTagsRow, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error)
	MarkAllRead(ctx context.Context, userID int64) error
	MarkNotifications(ctx context.Context, arg MarkNotificationsParams) (int64, error)
	ReadPost(ctx context.Context, arg ReadPostParams) (ReadPostRow, error)
	UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error)
	UpdatePost(ctx context.Context, arg UpdatePostParams) (Post, error)
	UpdatePostContent(ctx context.Context, arg UpdatePostContentParams) (PostContent, error)
	UpdatePostFeature(ctx context.Context, arg UpdatePostFeatureParams) error
	UpdatePostStatus(ctx context.Context, arg UpdatePostStatusParams) ([]Post, error)
	UpdateTag(ctx context.Context, arg UpdateTagParams) (Tag, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}

var _ Querier = (*Queries)(nil)
