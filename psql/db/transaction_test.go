package db

import (
	"context"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomPost(t *testing.T) CreateNewPostRow {
	store := NewStore(testDB)
	user := createRandomUser(t)

	arg := CreateNewPostParams{
		AuthorID:   user.ID,
		Title:      util.RandomString(10),
		Abstract:   util.RandomString(30),
		CoverImage: util.RandomString(20),
		Content:    util.RandomString(100),
	}

	post, err := store.CreateNewPost(context.Background(), arg)
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
	createRandomPost(t)
}

func TestDeletePosts(t *testing.T) {
	post1 := createRandomPost(t)

	arg := DeletePostParams{
		ID:       post1.ID,
		AuthorID: post1.AuthorID,
	}
	err := testQueries.DeletePost(context.Background(), arg)
	require.NoError(t, err)
}
