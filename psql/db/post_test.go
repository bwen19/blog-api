package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func TestDeletePosts(t *testing.T) {
	post1 := createRandomPost(t)

	arg := DeletePostParams{
		ID:       post1.ID,
		AuthorID: post1.AuthorID,
	}
	err := testStore.DeletePost(context.Background(), arg)
	require.NoError(t, err)

	getArg := GetPostParams{
		PostID:   post1.ID,
		IsAdmin:  true,
		AuthorID: post1.AuthorID,
	}

	post2, err := testStore.GetPost(context.Background(), getArg)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, post2)
}

func TestUpdatePost(t *testing.T) {
	post1 := createRandomPost(t)

	arg := UpdatePostParams{
		ID:       post1.ID,
		Title:    sql.NullString{String: "test", Valid: true},
		AuthorID: post1.AuthorID,
	}

	post2, err := testStore.UpdatePost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post2)

	require.Equal(t, arg.Title.String, post2.Title)

	conArg := UpdatePostContentParams{
		ID:       post1.ID,
		Content:  util.RandomString(50),
		AuthorID: post1.AuthorID,
	}
	content, err := testStore.UpdatePostContent(context.Background(), conArg)
	require.NoError(t, err)
	require.NotEmpty(t, content)
	require.Equal(t, conArg.Content, content.Content)
}

func TestNotUpdatePost(t *testing.T) {
	post1 := createRandomPost(t)

	arg := UpdatePostParams{
		ID:       post1.ID,
		Title:    sql.NullString{String: "test", Valid: false},
		AuthorID: post1.AuthorID,
	}

	post2, err := testStore.UpdatePost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post2)

	require.NotEqual(t, arg.Title.String, post2.Title)
	require.Equal(t, post1.Title, post2.Title)
}

func TestUpdatePostStatus(t *testing.T) {
	post1 := createRandomPost(t)

	arg := UpdatePostStatusParams{
		Ids:       []int64{post1.ID},
		Status:    "publish",
		OldStatus: []string{"draft"},
		IsAdmin:   true,
	}

	post2, err := testStore.UpdatePostStatus(context.TODO(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post2)
	require.Equal(t, post1.ID, post2[0].ID)
	require.Equal(t, post1.Status, "draft")
	require.Equal(t, post2[0].Status, "publish")

	arg = UpdatePostStatusParams{
		Ids:       []int64{post1.ID},
		Status:    "review",
		OldStatus: []string{"publish"},
		AuthorID:  0,
	}
	post3, err := testStore.UpdatePostStatus(context.TODO(), arg)
	require.NoError(t, err)
	require.Empty(t, post3)
}

func TestListPosts(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomPost(t)
	}

	arg := ListPostsParams{
		Limit:      5,
		Offset:     0,
		IsAdmin:    true,
		AnyStatus:  true,
		AnyKeyword: true,
	}
	posts, err := testStore.ListPosts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, posts)
	require.Equal(t, len(posts), 5)
}

func TestGetFeaturedPosts(t *testing.T) {
	postIDs := []int64{}
	for i := 0; i < 10; i++ {
		post := createRandomPost(t)
		postIDs = append(postIDs, post.ID)
	}

	arg1 := UpdatePostStatusParams{
		Ids:       postIDs,
		OldStatus: []string{"draft"},
		Status:    "publish",
		IsAdmin:   true,
	}
	p, err := testStore.UpdatePostStatus(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, p)

	for _, postID := range postIDs {
		arg := UpdatePostFeatureParams{
			ID:       postID,
			Featured: true,
		}
		err := testStore.UpdatePostFeature(context.Background(), arg)
		require.NoError(t, err)
	}

	posts, err := testStore.GetFeaturedPosts(context.Background(), 4)
	require.NoError(t, err)
	require.NotEmpty(t, posts)
	require.Equal(t, 4, len(posts))
}

func TestGetPosts(t *testing.T) {
	postIDs := []int64{}
	for i := 0; i < 10; i++ {
		post := createRandomPost(t)
		postIDs = append(postIDs, post.ID)
	}

	arg1 := UpdatePostStatusParams{
		Ids:       postIDs,
		OldStatus: []string{"draft"},
		Status:    "publish",
		IsAdmin:   true,
	}
	p, err := testStore.UpdatePostStatus(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, p)

	arg := GetPostsParams{
		Limit:        5,
		AnyFeatured:  true,
		AnyAuthor:    true,
		AnyCategory:  true,
		AnyTag:       true,
		AnyKeyword:   true,
		PublishAtAsc: true,
	}

	posts, err := testStore.GetPosts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, posts)
	require.Equal(t, len(posts), 5)

	count := 0
	for _, postID := range postIDs {
		if count > 2 {
			break
		}
		arg := UpdatePostFeatureParams{
			ID:       postID,
			Featured: true,
		}
		testStore.UpdatePostFeature(context.Background(), arg)
		count++
	}

	arg = GetPostsParams{
		Limit:        3,
		Featured:     true,
		AnyAuthor:    true,
		AnyCategory:  true,
		AnyTag:       true,
		AnyKeyword:   true,
		PublishAtAsc: true,
	}

	posts, err = testStore.GetPosts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, posts)
	require.Equal(t, len(posts), 3)
}

func TestReadPost(t *testing.T) {
	post := createRandomPost(t)
	arg1 := UpdatePostStatusParams{
		Ids:       []int64{post.ID},
		OldStatus: []string{"draft"},
		Status:    "publish",
		IsAdmin:   true,
	}
	p, err := testStore.UpdatePostStatus(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, p)

	arg := ReadPostParams{
		PostID: post.ID,
		SelfID: 0,
	}
	p1, err := testStore.ReadPost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, p1)
	require.Equal(t, post.ID, p1.ID)
	require.Equal(t, post.ViewCount+1, p1.ViewCount)
	require.Equal(t, false, p1.Followed.Valid)

	arg2 := CreatePostStarParams{
		PostID: post.ID,
		UserID: post.AuthorID,
	}
	err = testStore.CreatePostStar(context.Background(), arg2)
	require.NoError(t, err)

	p1, err = testStore.ReadPost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, p1)
	require.Equal(t, post.ID, p1.ID)
	require.Equal(t, post.ViewCount+2, p1.ViewCount)
	require.Equal(t, false, p1.Followed.Valid)
	require.Equal(t, int64(1), p1.StarCount)

	arg3 := DeletePostStarParams{
		PostID: post.ID,
		UserID: post.AuthorID,
	}
	err = testStore.DeletePostStar(context.Background(), arg3)
	require.NoError(t, err)

	p1, err = testStore.ReadPost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, p1)
	require.Equal(t, post.ID, p1.ID)
	require.Equal(t, post.ViewCount+3, p1.ViewCount)
	require.Equal(t, false, p1.Followed.Valid)
	require.Equal(t, int64(0), p1.StarCount)
}
