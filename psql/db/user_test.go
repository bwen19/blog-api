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
		Avatar:         util.RandomString(10),
		Email:          util.RandomEmail(),
		Role:           "user",
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
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

	nrows, err := testQueries.DeleteUsers(context.Background(), userIDs)
	require.NoError(t, err)
	require.Equal(t, int64(len(userIDs)), nrows)

	user2, err := testQueries.GetUser(context.Background(), user1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user2)
}
