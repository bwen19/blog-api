package db

import (
	"context"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomCategory(t *testing.T) Category {
	name := util.RandomString(6)

	category, err := testStore.CreateCategory(context.Background(), name)
	require.NoError(t, err)
	require.NotEmpty(t, category)

	require.Equal(t, name, category.Name)
	require.NotZero(t, category.ID)

	return category
}

func TestCreateCategory(t *testing.T) {
	createRandomCategory(t)
}

func TestDeleteCategories(t *testing.T) {
	cateIDs := []int64{}
	count := 5
	for i := 0; i < count; i++ {
		cate := createRandomCategory(t)
		cateIDs = append(cateIDs, cate.ID)
	}

	nrows, err := testStore.DeleteCategories(context.Background(), cateIDs)
	require.NoError(t, err)
	require.Equal(t, nrows, int64(count))
}

func TestUpdateCategory(t *testing.T) {
	cate := createRandomCategory(t)
	newName := util.RandomString(7)

	arg := UpdateCategoryParams{
		ID:   cate.ID,
		Name: newName,
	}
	newCate, err := testStore.UpdateCategory(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, newCate)

	require.Equal(t, newName, newCate.Name)
}

func TestListCategories(t *testing.T) {
	count := 5
	for i := 0; i < count; i++ {
		createRandomCategory(t)
	}

	arg := ListCategoriesParams{
		NameAsc: true,
	}
	cates, err := testStore.ListCategories(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, cates)
}
