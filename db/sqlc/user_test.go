package db

import (
	"blog/server/util"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomString(6),
		HashedPassword: hashedPassword,
		Email:          util.RandomEmail(),
		AvatarSrc:      util.RandomString(6),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, "user", user.Role)
	require.Equal(t, arg.AvatarSrc, user.AvatarSrc)
	require.NotZero(t, user.CreateAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestDeleteUser(t *testing.T) {
	user1 := createRandomUser(t)
	err := testQueries.DeleteUser(context.Background(), user1.Username)
	require.NoError(t, err)

	arg := GetUserParams{
		Username: user1.Username,
		Email:    "",
	}
	user2, err := testQueries.GetUser(context.Background(), arg)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user2)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	arg := GetUserParams{
		Username: user1.Username,
		Email:    "",
	}
	user2, err := testQueries.GetUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Role, user2.Role)
	require.Equal(t, user1.AvatarSrc, user2.AvatarSrc)
	require.WithinDuration(t, user1.CreateAt, user2.CreateAt, time.Second)

	arg = GetUserParams{
		Username: "",
		Email:    user1.Email,
	}
	user3, err := testQueries.GetUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user3)
	require.Equal(t, user1.Username, user3.Username)
	require.Equal(t, user1.HashedPassword, user3.HashedPassword)
	require.Equal(t, user1.Email, user3.Email)
	require.Equal(t, user1.Role, user3.Role)
	require.Equal(t, user1.AvatarSrc, user3.AvatarSrc)
	require.WithinDuration(t, user1.CreateAt, user3.CreateAt, time.Second)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}

	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}

	users, err := testQueries.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, users, 5)
	for _, user := range users {
		require.NotEmpty(t, user)
	}
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := UpdateUserParams{
		Username:          user1.Username,
		SetNewName:        true,
		NewName:           util.RandomString(8),
		SetHashedPassword: true,
		HashedPassword:    hashedPassword,
		SetEmail:          true,
		Email:             util.RandomEmail(),
		SetRole:           true,
		Role:              "author",
		SetAvatarSrc:      true,
		AvatarSrc:         util.RandomString(10),
	}

	user2, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, arg.NewName, user2.Username)
	require.Equal(t, arg.HashedPassword, user2.HashedPassword)
	require.Equal(t, arg.Email, user2.Email)
	require.Equal(t, arg.Role, user2.Role)
	require.Equal(t, arg.AvatarSrc, user2.AvatarSrc)
	require.WithinDuration(t, user1.CreateAt, user2.CreateAt, time.Second)
}
