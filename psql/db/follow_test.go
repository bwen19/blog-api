package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomFollow(t *testing.T, userID int64, followerID int64) {
	arg := CreateFollowParams{
		UserID:     userID,
		FollowerID: followerID,
	}

	err := testStore.CreateFollow(context.Background(), arg)
	require.NoError(t, err)
}

func TestFollow(t *testing.T) {
	userIDs := []int64{}
	for i := 0; i < 6; i++ {
		user := createRandomUser(t)
		userIDs = append(userIDs, user.ID)
	}

	for _, u1 := range userIDs {
		for _, u2 := range userIDs {
			if u1 == u2 {
				continue
			}
			createRandomFollow(t, u1, u2)
		}
	}

	arg1 := ListFollowersParams{
		Limit:  5,
		SelfID: userIDs[1],
		UserID: userIDs[0],
	}

	f1, err := testStore.ListFollowers(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, f1)
	require.Equal(t, 5, len(f1))

	for _, v := range f1 {
		if v.UserID == userIDs[1] {
			continue
		}
		require.Equal(t, true, v.Followed.Valid)
	}

	arg2 := ListFollowingsParams{
		Limit:  5,
		SelfID: 0,
		UserID: userIDs[1],
	}

	f2, err := testStore.ListFollowings(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, f2)
	require.Equal(t, 5, len(f2))

	for _, v := range f2 {
		require.Equal(t, false, v.Followed.Valid)
	}
}
