package db

import (
	"blog/util"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomArticleBy(t *testing.T, user User, category string) Article {
	arg := CreateArticleParams{
		Author:   user.Username,
		Category: category,
		Title:    util.RandomString(10),
		Summary:  util.RandomString(6),
		Content:  util.RandomString(20),
	}

	article, err := testQueries.CreateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, article)
	require.NotZero(t, article.ID)
	require.Equal(t, arg.Author, article.Author)
	require.Equal(t, arg.Category, article.Category)
	require.Equal(t, arg.Title, article.Title)
	require.Equal(t, arg.Summary, article.Summary)
	require.Equal(t, arg.Content, article.Content)
	require.Equal(t, "draft", article.Status)
	require.Equal(t, int64(0), article.ViewCount)
	require.NotZero(t, article.UpdateAt)
	require.NotZero(t, article.CreateAt)

	return article
}

func createRandomArticle(t *testing.T) Article {
	user := createRandomUser(t)
	category := createRandomCategory(t)
	return createRandomArticleBy(t, user, category)
}

func TestCreateArticle(t *testing.T) {
	createRandomArticle(t)
}

func TestDeleteArticle(t *testing.T) {
	article1 := createRandomArticle(t)
	arg1 := DeleteArticleParams{
		ID:        article1.ID,
		AnyAuthor: true,
		AnyStatus: true,
	}
	err := testQueries.DeleteArticle(context.Background(), arg1)
	require.NoError(t, err)

	arg := GetArticleParams{
		ID:        article1.ID,
		AnyAuthor: true,
	}
	article2, err := testQueries.GetArticle(context.Background(), arg)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, article2)
}

func TestGetArticle(t *testing.T) {
	article1 := createRandomArticle(t)

	arg := GetArticleParams{
		ID:        article1.ID,
		AnyAuthor: true,
	}
	article2, err := testQueries.GetArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, article2)
	require.Equal(t, article1.ID, article2.ID)
	require.Equal(t, article1.Author, article2.Author)
	require.Equal(t, article1.Category, article2.Category)
	require.Equal(t, article1.Title, article2.Title)
	require.Equal(t, article1.Summary, article2.Summary)
	require.Equal(t, article1.Content, article2.Content)
	require.Equal(t, article1.Status, article2.Status)
	require.Equal(t, article1.ViewCount, article2.ViewCount)
	require.WithinDuration(t, article1.UpdateAt, article2.UpdateAt, time.Second)
	require.WithinDuration(t, article1.CreateAt, article2.CreateAt, time.Second)

	arg = GetArticleParams{
		ID:        article1.ID,
		AnyAuthor: false,
		Author:    article1.Author,
	}
	article3, err := testQueries.GetArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, article3)
}

func TestListArticles(t *testing.T) {
	user := createRandomUser(t)
	category := createRandomCategory(t)
	tag := createRandomTag(t)
	for i := 0; i < 10; i++ {
		article := createRandomArticleBy(t, user, category)
		if i > 7 {
			createRandomArticleTag(t, article, tag)
		}
	}

	arg := ListArticlesParams{
		Limit:       5,
		Offset:      5,
		Status:      "draft",
		AnyAuthor:   true,
		AnyCategory: true,
		AnyTag:      true,
		TimeDesc:    true,
		CountDesc:   false,
	}

	articles, err := testQueries.ListArticles(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, articles, 5)

	for _, article := range articles {
		require.NotEmpty(t, article)
		require.Equal(t, article.Status, "draft")
	}

	arg = ListArticlesParams{
		Limit:       5,
		Offset:      5,
		AnyStatus:   true,
		Status:      "review",
		AnyAuthor:   false,
		Author:      user.Username,
		AnyCategory: true,
		AnyTag:      true,
		TimeDesc:    true,
		CountDesc:   false,
	}

	articles, err = testQueries.ListArticles(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, articles, 5)

	for _, article := range articles {
		require.NotEmpty(t, article)
		require.Equal(t, article.Author, user.Username)
	}

	arg = ListArticlesParams{
		Limit:       5,
		Offset:      5,
		AnyStatus:   true,
		Status:      "published",
		AnyAuthor:   true,
		AnyCategory: false,
		Category:    category,
		AnyTag:      true,
		TimeDesc:    true,
		CountDesc:   false,
	}

	articles, err = testQueries.ListArticles(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, articles, 5)

	for _, article := range articles {
		require.NotEmpty(t, article)
		require.Equal(t, article.Category, category)
	}

	arg = ListArticlesParams{
		Limit:       5,
		Offset:      0,
		AnyStatus:   true,
		Status:      "draft",
		AnyAuthor:   true,
		AnyCategory: true,
		AnyTag:      false,
		Tag:         tag.Name,
		TimeDesc:    true,
		CountDesc:   false,
	}

	articles, err = testQueries.ListArticles(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, articles, 2)

	for _, article := range articles {
		require.NotEmpty(t, article)
	}
}

func TestReadArticle(t *testing.T) {
	article1 := createRandomArticle(t)
	arg := UpdateArticleParams{
		ID:        article1.ID,
		AnyAuthor: true,
		SetStatus: true,
		Status:    "published",
	}
	article, err := testQueries.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, article)

	article2, err := testQueries.ReadArticle(context.Background(), article1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, article2)
	require.Equal(t, article1.ID, article2.ID)
	require.Equal(t, article1.Title, article2.Title)
	require.Equal(t, article1.Author, article2.Author)
	require.Equal(t, article1.Category, article2.Category)
	require.Equal(t, article1.Summary, article2.Summary)
	require.Equal(t, article1.Content, article2.Content)
	require.Equal(t, article.Status, article2.Status)
	require.Equal(t, article1.ViewCount+1, article2.ViewCount)
	require.WithinDuration(t, article1.UpdateAt, article2.UpdateAt, time.Second)
	require.WithinDuration(t, article1.CreateAt, article2.CreateAt, time.Second)
}

func TestUpdateArticle(t *testing.T) {
	article1 := createRandomArticle(t)
	newUser := createRandomUser(t)
	newCategory := createRandomCategory(t)

	arg := UpdateArticleParams{
		ID:           article1.ID,
		AnyAuthor:    true,
		SetTitle:     true,
		Title:        util.RandomString(7),
		SetAuthor:    true,
		Author:       newUser.Username,
		SetCategory:  true,
		Category:     newCategory,
		SetSummary:   true,
		Summary:      util.RandomString(10),
		SetContent:   true,
		Content:      util.RandomString(22),
		SetStatus:    true,
		Status:       "review",
		SetViewCount: true,
		ViewCount:    8,
	}

	article2, err := testQueries.UpdateArticle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, article2)
	require.Equal(t, arg.ID, article2.ID)
	require.Equal(t, arg.Title, article2.Title)
	require.Equal(t, arg.Author, article2.Author)
	require.Equal(t, arg.Category, article2.Category)
	require.Equal(t, arg.Summary, article2.Summary)
	require.Equal(t, arg.Content, article2.Content)
	require.Equal(t, arg.Status, article2.Status)
	require.Equal(t, arg.ViewCount, article2.ViewCount)
	require.WithinDuration(t, article1.UpdateAt, article2.UpdateAt, time.Second)
	require.WithinDuration(t, article1.CreateAt, article2.CreateAt, time.Second)
}
