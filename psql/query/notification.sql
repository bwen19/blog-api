-- name: CreateNotification :exec
INSERT INTO notifications (user_id, kind, title, content)
VALUES ($1, $2, $3, $4);

-- name: DeleteNotifications :execrows
DELETE FROM notifications
WHERE id = ANY(@ids::bigint[])
  AND user_id = @user_id::bigint
  AND kind <> 'admin';

-- name: MarkAllRead :exec
UPDATE notifications SET unread = false
WHERE user_id = @user_id::bigint
  AND unread = true
  AND kind <> 'admin';

-- name: MarkNotifications :execrows
UPDATE notifications SET unread = @unread::bool
WHERE id = ANY(@ids::bigint[]);

-- name: GetUnreadCount :one
SELECT count(*) FROM notifications
WHERE user_id = @user_id::bigint
  AND unread = true
  AND kind <> 'admin';

-- name: ListNotifications :many
WITH Data_CTE AS (
  SELECT id, kind, title, content, unread, create_at
  FROM notifications
  WHERE user_id = @user_id::bigint AND kind = @kind::varchar
),
Count_CTE AS (
  SELECT count(*) total,
    count(*) filter(WHERE unread = true) unread_count,
    count(*) filter(WHERE unread = true AND kind = 'system') system_count,
    count(*) filter(WHERE unread = true AND kind = 'reply') reply_count
  FROM Data_CTE
)
SELECT dc.*, cnt.*
FROM Data_CTE dc
CROSS JOIN Count_CTE cnt
ORDER BY create_at DESC
LIMIT $1
OFFSET $2;

-- name: ListMessages :many
WITH Data_CTE AS (
  SELECT *
  FROM notifications WHERE kind = 'admin'
),
Count_CTE AS (
  SELECT count(*) total,
    count(*) filter(WHERE unread = true) unread_count
  FROM Data_CTE
)
SELECT dc.*, cc.total, u.username, u.email, u.avatar
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
JOIN users u ON u.id = dc.user_id
ORDER BY create_at DESC
LIMIT $1
OFFSET $2;
