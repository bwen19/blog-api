// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: notification.sql

package sqlc

import (
	"context"
	"time"
)

const createNotification = `-- name: CreateNotification :exec
INSERT INTO notifications (
  user_id, kind, title, content
) VALUES (
  $1, $2, $3, $4
)
`

type CreateNotificationParams struct {
	UserID  int64  `json:"user_id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) error {
	_, err := q.db.Exec(ctx, createNotification,
		arg.UserID,
		arg.Kind,
		arg.Title,
		arg.Content,
	)
	return err
}

const deleteNotifications = `-- name: DeleteNotifications :execrows
DELETE FROM notifications
WHERE id = ANY($1::bigint[]) AND user_id = $2::bigint
`

type DeleteNotificationsParams struct {
	Ids    []int64 `json:"ids"`
	UserID int64   `json:"user_id"`
}

func (q *Queries) DeleteNotifications(ctx context.Context, arg DeleteNotificationsParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteNotifications, arg.Ids, arg.UserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getUnreadCount = `-- name: GetUnreadCount :one
SELECT count(*) FROM notifications
WHERE user_id = $1::bigint AND unread = true
`

func (q *Queries) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	row := q.db.QueryRow(ctx, getUnreadCount, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
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
)
SELECT id, kind, title, content, unread, create_at, unread_count, total FROM Data_CTE
CROSS JOIN Count_CTE
ORDER BY create_at DESC
LIMIT $1
OFFSET $2
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
	rows, err := q.db.Query(ctx, listNotifications,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markAllRead = `-- name: MarkAllRead :exec
UPDATE notifications
SET unread = false
WHERE user_id = $1::bigint AND unread = true
`

func (q *Queries) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := q.db.Exec(ctx, markAllRead, userID)
	return err
}

const markReadByIDs = `-- name: MarkReadByIDs :execrows
UPDATE notifications
SET unread = false
WHERE user_id = $1::bigint AND id = ANY($2::bigint[])
  AND unread = true
`

type MarkReadByIDsParams struct {
	UserID int64   `json:"user_id"`
	Ids    []int64 `json:"ids"`
}

func (q *Queries) MarkReadByIDs(ctx context.Context, arg MarkReadByIDsParams) (int64, error) {
	result, err := q.db.Exec(ctx, markReadByIDs, arg.UserID, arg.Ids)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}