package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomString(7),
		HashedPassword: hashedPassword,
		Avatar:         "/image/avatar/default",
		Email:          util.RandomEmail(),
		Role:           "user",
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Avatar, user.Avatar)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, "user", user.Role)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestDeleteUsers(t *testing.T) {
	user1 := createRandomUser(t)
	userIDs := []int64{user1.ID}

	nrows, err := testStore.DeleteUsers(context.Background(), userIDs)
	require.NoError(t, err)
	require.Equal(t, int64(len(userIDs)), nrows)

	user2, err := testStore.GetUser(context.Background(), user1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user2)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomUser(t)
	}

	arg := ListUsersParams{
		Limit:        5,
		Offset:       0,
		CreateAtDesc: true,
		AnyKeyword:   true,
	}

	users, err := testStore.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, users)

	require.Equal(t, len(users), 5)
}

func TestGetUserProfile(t *testing.T) {
	user := createRandomUser(t)

	arg := GetUserProfileParams{
		UserID: user.ID,
		SelfID: 0,
	}
	userP, err := testStore.GetUserProfile(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userP)

	require.Equal(t, user.ID, userP.ID)
	require.Equal(t, user.Username, userP.Username)
}
