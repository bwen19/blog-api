package db

import (
	"blog/util"
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomCategory(t *testing.T) string {
	name := util.RandomString(7)
	category, err := testQueries.CreateCategory(context.Background(), name)
	require.NoError(t, err)
	require.NotEmpty(t, category)
	require.Equal(t, name, category)

	return category
}

func TestCreateCategory(t *testing.T) {
	createRandomCategory(t)
}

func TestDeleteCategory(t *testing.T) {
	category1 := createRandomCategory(t)
	err := testQueries.DeleteCategory(context.Background(), category1)
	require.NoError(t, err)

	category2, err := testQueries.GetCategory(context.Background(), category1)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Equal(t, category1, category2)
}

func TestGetCategory(t *testing.T) {
	category1 := createRandomCategory(t)

	category2, err := testQueries.GetCategory(context.Background(), category1)
	require.NoError(t, err)
	require.NotEmpty(t, category2)
	require.Equal(t, category1, category2)
}

func TestListCategories(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomCategory(t)
	}

	categories, err := testQueries.ListCategories(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(categories), 10)

	for _, category := range categories {
		require.NotEmpty(t, category)
	}
}

func TestUpdateCategory(t *testing.T) {
	category1 := createRandomCategory(t)

	arg := UpdateCategoryParams{
		NewName: util.RandomString(8),
		Name:    category1,
	}

	category2, err := testQueries.UpdateCategory(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, category2)
	require.Equal(t, arg.NewName, category2)
}
