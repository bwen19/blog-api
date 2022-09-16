package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomComment(t *testing.T, postID int64, userIDs []int64) {
	maxNum := int64(len(userIDs)) - 1
	for i := 0; i < 10; i++ {
		idx := util.RandomInt(1, maxNum)
		arg := CreateCommentParams{
			PostID:  postID,
			UserID:  userIDs[idx],
			Content: util.RandomString(20),
		}

		cm, err := testStore.CreateComment(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, cm)

		for j := 0; j < 6; j++ {
			idx1 := util.RandomInt(1, maxNum)
			arg1 := CreateCommentParams{
				PostID:      postID,
				ParentID:    sql.NullInt64{Int64: cm.ID, Valid: true},
				UserID:      userIDs[idx1],
				ReplyUserID: sql.NullInt64{Int64: cm.UserID, Valid: true},
				Content:     util.RandomString(20),
			}

			cm1, err := testStore.CreateComment(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, cm1)
		}
	}
}

func TestComment(t *testing.T) {
	post := createRandomPost(t)
	userIDs := []int64{}
	for i := 0; i < 5; i++ {
		user := createRandomUser(t)
		userIDs = append(userIDs, user.ID)
	}

	createRandomComment(t, post.ID, userIDs)

	arg := ListCommentsParams{
		Limit:         5,
		PostID:        post.ID,
		StarCountDesc: true,
	}

	cms, err := testStore.ListComments(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, cms)
	require.Equal(t, 15, len(cms))

	var count1, count2 int
	var parentID int64
	for _, cm := range cms {
		if cm.ParentID.Valid {
			count2++
		} else {
			count1++
			parentID = cm.ID
		}
	}
	require.Equal(t, 5, count1)
	require.Equal(t, 10, count2)

	arg2 := ListRepliesParams{
		Limit:        5,
		CreateAtDesc: true,
		ParentID:     parentID,
	}

	replies, err := testStore.ListReplies(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, replies)
	require.Equal(t, 5, len(replies))
}
