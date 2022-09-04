package test

import (
	"blog/server/db/sqlc"
	"blog/server/util"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createPost(t *testing.T) sqlc.Post {
	user := createRandomUser(t)

	arg := sqlc.CreatePostParams{
		AuthorID:   user.ID,
		Title:      util.RandomString(10),
		Abstract:   util.RandomString(30),
		CoverImage: util.RandomString(20),
		Content:    util.RandomString(80),
	}

	post, err := testQueries.CreatePost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post)

	require.Equal(t, arg.AuthorID, post.AuthorID)
	require.Equal(t, arg.Title, post.Title)
	require.Equal(t, arg.Abstract, post.Abstract)
	require.Equal(t, arg.CoverImage, post.CoverImage)
	require.Equal(t, arg.Content, post.Content)
	require.Equal(t, "draft", post.Status)
	require.NotZero(t, post.ID)
	require.NotZero(t, post.UpdateAt)

	return post
}

func TestCreatePost(t *testing.T) {
	createPost(t)
}

func TestDeletePosts(t *testing.T) {
	post1 := createPost(t)
	postIDs := []int64{post1.ID}

	arg := sqlc.DeletePostsParams{
		Ids:      postIDs,
		AuthorID: post1.AuthorID,
	}
	nrows, err := testQueries.DeletePosts(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(len(postIDs)), nrows)
}
