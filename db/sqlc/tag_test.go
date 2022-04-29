package db

import (
	"blog/util"
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomTag(t *testing.T) Tag {
	name := util.RandomString(7)
	tag, err := testQueries.CreateTag(context.Background(), name)
	require.NoError(t, err)
	require.NotEmpty(t, tag)
	require.Equal(t, name, tag.Name)
	require.Equal(t, int64(0), tag.Count)

	return tag
}

func TestCreateTag(t *testing.T) {
	createRandomTag(t)
}

func TestDeleteTag(t *testing.T) {
	tag1 := createRandomTag(t)
	err := testQueries.DeleteTag(context.Background(), tag1.Name)
	require.NoError(t, err)

	tag2, err := testQueries.GetTag(context.Background(), tag1.Name)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, tag2)
}

func TestGetTag(t *testing.T) {
	tag1 := createRandomTag(t)

	tag2, err := testQueries.GetTag(context.Background(), tag1.Name)
	require.NoError(t, err)
	require.NotEmpty(t, tag2)
	require.Equal(t, tag1.Name, tag2.Name)
	require.Equal(t, tag1.Count, tag2.Count)
}

func TestListTags(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTag(t)
	}

	arg := ListTagsParams{
		Limit:  5,
		Offset: 5,
	}

	tags, err := testQueries.ListTags(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, tags, 5)

	for _, tag := range tags {
		require.NotEmpty(t, tag)
	}
}

func TestUpdateTag(t *testing.T) {
	tag1 := createRandomTag(t)

	arg := UpdateTagParams{
		Name:       tag1.Name,
		SetNewName: true,
		NewName:    util.RandomString(8),
		SetCount:   true,
		Count:      9,
	}

	tag2, err := testQueries.UpdateTag(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, tag2)

	require.Equal(t, arg.NewName, tag2.Name)
	require.Equal(t, arg.Count, tag2.Count)
}

func createRandomArticleTag(t *testing.T, article Article, tag Tag) ArticleTag {
	arg := CreateArticleTagParams{
		ArticleID: article.ID,
		Tag:       tag.Name,
	}

	articleTag, err := testQueries.CreateArticleTag(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, articleTag)
	require.Equal(t, arg.ArticleID, articleTag.ArticleID)
	require.Equal(t, arg.Tag, articleTag.Tag)

	return articleTag
}

func TestCreateArticleTag(t *testing.T) {
	article := createRandomArticle(t)
	tag := createRandomTag(t)
	createRandomArticleTag(t, article, tag)
}

func TestDeleteArticleTag(t *testing.T) {
	article := createRandomArticle(t)
	tag := createRandomTag(t)
	articleTag1 := createRandomArticleTag(t, article, tag)

	arg := DeleteArticleTagParams(articleTag1)
	err := testQueries.DeleteArticleTag(context.Background(), arg)
	require.NoError(t, err)
}

func TestListArticleTags(t *testing.T) {
	article := createRandomArticle(t)
	for i := 0; i < 5; i++ {
		tag := createRandomTag(t)
		createRandomArticleTag(t, article, tag)
	}

	tags, err := testQueries.ListArticleTags(context.Background(), article.ID)
	require.NoError(t, err)
	require.Len(t, tags, 5)

}
