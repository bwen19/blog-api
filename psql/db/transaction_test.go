package db

import (
	"context"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomPost(t *testing.T, user User) CreateNewPostRow {
	arg := CreateNewPostParams{
		AuthorID:   user.ID,
		Title:      util.RandomText(10),
		CoverImage: "/image/post/default",
		Content:    util.RandomText(125),
	}

	post, err := testStore.CreateNewPost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post)

	require.Equal(t, arg.AuthorID, post.AuthorID)
	require.Equal(t, arg.Title, post.Title)
	require.Equal(t, arg.CoverImage, post.CoverImage)
	require.Equal(t, arg.Content, post.Content)
	require.Equal(t, "draft", post.Status)
	require.NotZero(t, post.ID)
	require.NotZero(t, post.UpdateAt)

	return post
}

func TestCreatePost(t *testing.T) {
	user := createRandomUser(t)
	createRandomPost(t, user)
}

func TestSetPostCategories(t *testing.T) {
	user := createRandomUser(t)
	post := createRandomPost(t, user)

	cateIDs := []int64{}
	for i := 0; i < 2; i++ {
		cate := createRandomCategory(t)
		cateIDs = append(cateIDs, cate.ID)
	}

	arg := SetPostCategoriesParams{
		PostID:      post.ID,
		CategoryIDs: cateIDs,
	}

	cates, err := testStore.SetPostCategories(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, cates)
	require.Equal(t, len(cates), 2)

	cateIDs = []int64{}
	for i := 0; i < 3; i++ {
		cate := createRandomCategory(t)
		cateIDs = append(cateIDs, cate.ID)
	}

	arg = SetPostCategoriesParams{
		PostID:      post.ID,
		CategoryIDs: cateIDs,
	}

	cates, err = testStore.SetPostCategories(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, cates)
	require.Equal(t, len(cates), 3)
}

func TestAll(t *testing.T) {
	cateIDs := []int64{}
	tagIDs := []int64{}
	for i := 0; i < 10; i++ {
		cate := createRandomCategory(t)
		cateIDs = append(cateIDs, cate.ID)
		tag := createRandomTag(t)
		tagIDs = append(tagIDs, tag.ID)
	}

	postIDs := []int64{}
	for i := 0; i < 6; i++ {
		user := createRandomUser(t)
		for j := 0; j < 5; j++ {
			post := createRandomPost(t, user)
			postIDs = append(postIDs, post.ID)

			k := i*5 + j

			arg1 := SetPostCategoriesParams{
				PostID:      post.ID,
				CategoryIDs: []int64{cateIDs[k%10], cateIDs[(k+1)%10]},
			}
			cates, err := testStore.SetPostCategories(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, cates)

			arg2 := SetPostTagsParams{
				PostID: post.ID,
				TagIDs: []int64{tagIDs[k%10], tagIDs[(k+1)%10], tagIDs[(k+2)%10]},
			}
			tags, err := testStore.SetPostTags(context.Background(), arg2)
			require.NoError(t, err)
			require.NotEmpty(t, tags)
		}

	}

	arg := UpdatePostStatusParams{
		Status:    "publish",
		Ids:       postIDs,
		OldStatus: []string{"draft"},
		IsAdmin:   true,
	}
	posts, err := testStore.UpdatePostStatus(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, posts)
}
