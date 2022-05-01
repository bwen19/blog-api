package db

import (
	"blog/server/util"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomComment(t *testing.T, arg CreateCommentParams) Comment {
	comment, err := testQueries.CreateComment(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, comment)
	require.NotZero(t, comment.ID)
	require.Equal(t, arg.ParentID, comment.ParentID.Int64)
	require.Equal(t, arg.Commenter, comment.Commenter)
	require.Equal(t, arg.ArticleID, comment.ArticleID)
	require.Equal(t, arg.Content, comment.Content)
	require.NotZero(t, comment.CommentAt)

	return comment
}

func createRandomComments(t *testing.T) []Comment {
	var comments []Comment
	article := createRandomArticle(t)
	user := createRandomUser(t)
	arg := CreateCommentParams{
		SetParentID: false,
		ArticleID:   article.ID,
		Commenter:   user.Username,
		Content:     util.RandomString(15),
	}
	comment := createRandomComment(t, arg)
	comments = append(comments, comment)

	for i := 0; i < 5; i++ {
		arg = CreateCommentParams{
			SetParentID: true,
			ParentID:    comment.ID,
			ArticleID:   article.ID,
			Commenter:   user.Username,
			Content:     util.RandomString(15),
		}
		childComment := createRandomComment(t, arg)
		comments = append(comments, childComment)
	}
	return comments
}

func TestCreateComment(t *testing.T) {
	createRandomComments(t)
}

func TestDeleteComment(t *testing.T) {
	comments := createRandomComments(t)
	arg := DeleteCommentParams{
		ID:           comments[0].ID,
		AnyCommenter: true,
	}
	err := testQueries.DeleteComment(context.Background(), arg)
	require.NoError(t, err)

	for _, cm := range comments {
		comment, err := testQueries.GetComment(context.Background(), cm.ID)
		require.Error(t, err)
		require.EqualError(t, err, sql.ErrNoRows.Error())
		require.Empty(t, comment)
	}
}

func TestGetComment(t *testing.T) {
	comments := createRandomComments(t)
	comment1 := comments[0]

	comment2, err := testQueries.GetComment(context.Background(), comment1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, comment2)
	require.Equal(t, comment1.ID, comment2.ID)
	require.Equal(t, comment1.ParentID, comment2.ParentID)
	require.Equal(t, comment1.ArticleID, comment2.ArticleID)
	require.Equal(t, comment1.Commenter, comment2.Commenter)
	require.Equal(t, comment1.Content, comment2.Content)
	require.WithinDuration(t, comment1.CommentAt, comment2.CommentAt, time.Second)
}

func TestListChildComments(t *testing.T) {
	comments1 := createRandomComments(t)
	pcomment := comments1[0]

	comments2, err := testQueries.ListChildComments(context.Background(), []int64{pcomment.ID})
	require.NoError(t, err)
	require.Len(t, comments2, 5)
	for _, comment := range comments2 {
		require.Equal(t, comment.ParentID.Int64, pcomment.ID)
		require.NotEmpty(t, comment)
	}
}

func TestListArticleComments(t *testing.T) {
	comments := createRandomComments(t)
	comment1 := comments[0]

	arg := ListCommentsByArticleParams{
		ArticleID: comment1.ArticleID,
		Limit:     5,
		Offset:    0,
	}

	comments2, err := testQueries.ListCommentsByArticle(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, comments2, 1)

	for _, comment := range comments2 {
		require.NotEmpty(t, comment)
		require.Zero(t, comment.ParentID)
	}
}
