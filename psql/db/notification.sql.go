// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: notification.sql

package db

import (
	"context"
	"time"

	"github.com/lib/pq"
)

const checkMessage = `-- name: CheckMessage :exec
UPDATE notifications
SET unread = false
WHERE id = ANY($1::bigint[])
`

func (q *Queries) CheckMessage(ctx context.Context, ids []int64) error {
	_, err := q.db.ExecContext(ctx, checkMessage, pq.Array(ids))
	return err
}

const createNotification = `-- name: CreateNotification :exec
INSERT INTO notifications (user_id, kind, title, content)
VALUES ($1, $2, $3, $4)
`

type CreateNotificationParams struct {
	UserID  int64  `json:"user_id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) error {
	_, err := q.db.ExecContext(ctx, createNotification,
		arg.UserID,
		arg.Kind,
		arg.Title,
		arg.Content,
	)
	return err
}

const deleteNotifications = `-- name: DeleteNotifications :execrows
DELETE FROM notifications
WHERE id = ANY($1::bigint[])
  AND user_id = $2::bigint
  AND kind <> 'admin'
`

type DeleteNotificationsParams struct {
	Ids    []int64 `json:"ids"`
	UserID int64   `json:"user_id"`
}

func (q *Queries) DeleteNotifications(ctx context.Context, arg DeleteNotificationsParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteNotifications, pq.Array(arg.Ids), arg.UserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getUnreadCount = `-- name: GetUnreadCount :one
SELECT count(*) FROM notifications
WHERE user_id = $1::bigint AND unread = true
`

func (q *Queries) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, getUnreadCount, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listMessages = `-- name: ListMessages :many
WITH Data_CTE AS (
  SELECT id, user_id, kind, title, content, unread, create_at
  FROM notifications WHERE kind = 'admin'
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.id, dc.user_id, dc.kind, dc.title, dc.content, dc.unread, dc.create_at, cc.total, u.username, u.email, u.avatar
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
JOIN users u ON u.id = dc.user_id
ORDER BY create_at DESC
LIMIT $1
OFFSET $2
`

type ListMessagesParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListMessagesRow struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	Kind     string    `json:"kind"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Unread   bool      `json:"unread"`
	CreateAt time.Time `json:"create_at"`
	Total    int64     `json:"total"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
}

func (q *Queries) ListMessages(ctx context.Context, arg ListMessagesParams) ([]ListMessagesRow, error) {
	rows, err := q.db.QueryContext(ctx, listMessages, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListMessagesRow{}
	for rows.Next() {
		var i ListMessagesRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Kind,
			&i.Title,
			&i.Content,
			&i.Unread,
			&i.CreateAt,
			&i.Total,
			&i.Username,
			&i.Email,
			&i.Avatar,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listNotifications = `-- name: ListNotifications :many
WITH Data_CTE AS (
  SELECT id, kind, title, content, unread, create_at
  FROM notifications
  WHERE user_id = $3::bigint AND kind = $4::varchar
),
Count_CTE AS (
  SELECT count(*) filter(WHERE unread = true) unread_count,
    count(*) total
  FROM Data_CTE
),
Notifs_CTE AS (
  SELECT id, kind, title, content, unread, create_at FROM Data_CTE
  ORDER BY create_at DESC
  LIMIT $1
  OFFSET $2
),
Read_CTE AS (
  UPDATE notifications
  SET unread = false
  WHERE id = ANY(SELECT id FROM Notifs_CTE) AND unread = true
  RETURNING id
),
ReadCount_CTE AS (
  SELECT count(*) read_count FROM Read_CTE
)
SELECT nc.id, nc.kind, nc.title, nc.content, nc.unread, nc.create_at, (cc.unread_count - rc.read_count)::bigint unread_count, cc.total
FROM Notifs_CTE nc
CROSS JOIN Count_CTE cc
CROSS JOIN ReadCount_CTE rc
`

type ListNotificationsParams struct {
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
	UserID int64  `json:"user_id"`
	Kind   string `json:"kind"`
}

type ListNotificationsRow struct {
	ID          int64     `json:"id"`
	Kind        string    `json:"kind"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Unread      bool      `json:"unread"`
	CreateAt    time.Time `json:"create_at"`
	UnreadCount int64     `json:"unread_count"`
	Total       int64     `json:"total"`
}

func (q *Queries) ListNotifications(ctx context.Context, arg ListNotificationsParams) ([]ListNotificationsRow, error) {
	rows, err := q.db.QueryContext(ctx, listNotifications,
		arg.Limit,
		arg.Offset,
		arg.UserID,
		arg.Kind,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListNotificationsRow{}
	for rows.Next() {
		var i ListNotificationsRow
		if err := rows.Scan(
			&i.ID,
			&i.Kind,
			&i.Title,
			&i.Content,
			&i.Unread,
			&i.CreateAt,
			&i.UnreadCount,
			&i.Total,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markAllRead = `-- name: MarkAllRead :exec
UPDATE notifications SET unread = false
WHERE user_id = $1::bigint
  AND unread = true
  AND kind <> 'admin'
`

func (q *Queries) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := q.db.ExecContext(ctx, markAllRead, userID)
	return err
}