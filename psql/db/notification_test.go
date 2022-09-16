package db

import (
	"context"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomNotif(t *testing.T, user User) {
	arg := CreateNotificationParams{
		UserID:  user.ID,
		Kind:    "system",
		Title:   util.RandomString(10),
		Content: util.RandomString(30),
	}

	err := testStore.CreateNotification(context.Background(), arg)
	require.NoError(t, err)
}

func TestCreateNotif(t *testing.T) {
	user := createRandomUser(t)
	createRandomNotif(t, user)
}

func TestDeleteNotifs(t *testing.T) {
	user := createRandomUser(t)

	for i := 0; i < 5; i++ {
		createRandomNotif(t, user)
	}

	err := testStore.MarkAllRead(context.Background(), user.ID)
	require.NoError(t, err)

	arg := ListNotificationsParams{
		Limit:  5,
		UserID: user.ID,
		Kind:   "system",
	}

	notifs, err := testStore.ListNotifications(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notifs)
	require.Equal(t, 5, len(notifs))

	notifIDs := []int64{}
	for _, notif := range notifs {
		require.Equal(t, false, notif.Unread)
		require.Equal(t, int64(0), notif.UnreadCount)
		require.Equal(t, int64(0), notif.SystemCount)
		notifIDs = append(notifIDs, notif.ID)
	}

	arg2 := DeleteNotificationsParams{
		Ids:    notifIDs,
		UserID: user.ID,
	}
	nrows, err := testStore.DeleteNotifications(context.Background(), arg2)
	require.NoError(t, err)
	require.Equal(t, len(notifIDs), int(nrows))
}

func TestListNotifications(t *testing.T) {
	user := createRandomUser(t)
	for i := 0; i < 5; i++ {
		createRandomNotif(t, user)
	}

	arg := ListNotificationsParams{
		Limit:  5,
		UserID: user.ID,
		Kind:   "system",
	}

	notifs, err := testStore.ListNotifications(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notifs)
	require.Equal(t, 5, len(notifs))

	notifIDs := []int64{}
	for _, notif := range notifs {
		require.Equal(t, true, notif.Unread)
		require.Equal(t, int64(5), notif.UnreadCount)
		require.Equal(t, int64(5), notif.SystemCount)
		notifIDs = append(notifIDs, notif.ID)
	}

	arg2 := MarkNotificationsParams{
		Ids:    notifIDs,
		Unread: false,
	}
	nrows, err := testStore.MarkNotifications(context.Background(), arg2)
	require.NoError(t, err)
	require.Equal(t, len(notifIDs), int(nrows))
}
