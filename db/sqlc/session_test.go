package db

import (
	"blog/util"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomSession(t *testing.T, user User) Session {
	id, err := uuid.NewRandom()
	require.NoError(t, err)

	arg := CreateSessionParams{
		ID:           id,
		Username:     user.Username,
		RefreshToken: util.RandomString(12),
		UserAgent:    util.RandomString(9),
		ClientIp:     util.RandomString(11),
		IsBlocked:    false,
		ExpiresAt:    time.Now(),
	}

	session, err := testQueries.CreateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	require.Equal(t, arg.ID, session.ID)
	require.Equal(t, arg.Username, session.Username)
	require.Equal(t, arg.RefreshToken, session.RefreshToken)
	require.Equal(t, arg.UserAgent, session.UserAgent)
	require.Equal(t, arg.ClientIp, session.ClientIp)
	require.Equal(t, arg.IsBlocked, session.IsBlocked)
	require.WithinDuration(t, arg.ExpiresAt, session.ExpiresAt, time.Second)

	return session
}

func TestCreateSession(t *testing.T) {
	user := createRandomUser(t)
	createRandomSession(t, user)
}

func TestDeleteSession(t *testing.T) {
	user := createRandomUser(t)
	session1 := createRandomSession(t, user)

	err := testQueries.DeleteSession(context.Background(), session1.ID)
	require.NoError(t, err)

	session2, err := testQueries.GetSession(context.Background(), session1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, session2)
}

func TestGetSession(t *testing.T) {
	user := createRandomUser(t)
	session1 := createRandomSession(t, user)

	session2, err := testQueries.GetSession(context.Background(), session1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, session2)

	require.Equal(t, session1.ID, session2.ID)
	require.Equal(t, session1.Username, session2.Username)
	require.Equal(t, session1.RefreshToken, session2.RefreshToken)
	require.Equal(t, session1.UserAgent, session2.UserAgent)
	require.Equal(t, session1.ClientIp, session2.ClientIp)
	require.Equal(t, session1.IsBlocked, session2.IsBlocked)
	require.Equal(t, session1.ExpiresAt, session2.ExpiresAt)
}
