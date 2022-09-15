package db

import (
	"context"
	"time"
)

type CreateNewPostParams struct {
	AuthorID   int64  `json:"author_id"`
	Title      string `json:"title"`
	Abstract   string `json:"abstract"`
	CoverImage string `json:"cover_image"`
	Content    string `json:"content"`
}

type CreateNewPostRow struct {
	ID         int64     `json:"id"`
	AuthorID   int64     `json:"author_id"`
	Title      string    `json:"title"`
	Abstract   string    `json:"abstract"`
	CoverImage string    `json:"cover_image"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
	Featured   bool      `json:"featured"`
	ViewCount  int64     `json:"view_count"`
	UpdateAt   time.Time `json:"update_at"`
	PublishAt  time.Time `json:"publish_at"`
}

func (store *SqlStore) CreateNewPost(ctx context.Context, arg CreateNewPostParams) (CreateNewPostRow, error) {
	var result CreateNewPostRow

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		arg1 := CreatePostParams{
			AuthorID:   arg.AuthorID,
			Title:      arg.Title,
			Abstract:   arg.Abstract,
			CoverImage: arg.CoverImage,
		}
		post, err := q.CreatePost(ctx, arg1)
		if err != nil {
			return err
		}

		arg2 := CreatePostContentParams{
			ID:      post.ID,
			Content: arg.Content,
		}
		content, err := q.CreatePostContent(ctx, arg2)
		if err != nil {
			return err
		}

		result.ID = post.ID
		result.Title = post.Title
		result.AuthorID = post.AuthorID
		result.Abstract = post.Abstract
		result.CoverImage = post.CoverImage
		result.Content = content.Content
		result.Status = post.Status
		result.Featured = post.Featured
		result.ViewCount = post.ViewCount
		result.PublishAt = post.PublishAt
		result.UpdateAt = post.UpdateAt
		return err
	})
	return result, err
}

// -------------------------------------------------------------------

type SetPostCategoriesParams struct {
	PostID      int64   `json:"post_id"`
	CategoryIDs []int64 `json:"category_ids"`
}

func (store *SqlStore) SetPostCategories(ctx context.Context, arg SetPostCategoriesParams) ([]Category, error) {
	var result []Category

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		arg1 := DeletePostCategoriesParams{
			PostID:      arg.PostID,
			CategoryIds: arg.CategoryIDs,
		}
		if err = q.DeletePostCategories(ctx, arg1); err != nil {
			return err
		}

		arg2 := CreatePostCategoriesParams{
			PostID:      arg.PostID,
			CategoryIds: arg.CategoryIDs,
		}
		categories, err := q.CreatePostCategories(ctx, arg2)
		if err != nil {
			return err
		}

		for _, category := range categories {
			result = append(result, Category(category))
		}
		return err
	})
	return result, err
}

// -------------------------------------------------------------------

type SetPostTagsParams struct {
	PostID int64   `json:"post_id"`
	TagIDs []int64 `json:"tag_ids"`
}

func (store *SqlStore) SetPostTags(ctx context.Context, arg SetPostTagsParams) ([]Tag, error) {
	var result []Tag

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		arg1 := DeletePostTagsParams{
			PostID: arg.PostID,
			TagIds: arg.TagIDs,
		}
		if err = q.DeletePostTags(ctx, arg1); err != nil {
			return err
		}

		arg2 := CreatePostTagsParams{
			PostID: arg.PostID,
			TagIds: arg.TagIDs,
		}
		tags, err := q.CreatePostTags(ctx, arg2)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			result = append(result, Tag(tag))
		}
		return err
	})
	return result, err
}
