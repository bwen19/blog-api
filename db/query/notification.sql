-- name: CreateNotification :exec
INSERT INTO notifications (
  user_id, kind, title, content
) VALUES (
  $1, $2, $3, $4
);

-- name: DeleteNotifications :execrows
DELETE FROM notifications
WHERE id = ANY(@ids::bigint[]) AND user_id = @user_id::bigint;

-- name: MarkAllRead :exec
UPDATE notifications
SET unread = false
WHERE user_id = @user_id::bigint AND unread = true;

-- name: GetUnreadCount :one
SELECT count(*) FROM notifications
WHERE user_id = @user_id::bigint AND unread = true;

-- name: ListNotifications :many
WITH Data_CTE AS (
  SELECT id, kind, title, content, unread, create_at
  FROM notifications
  WHERE user_id = @user_id::bigint AND kind = @kind::varchar
),
Count_CTE AS (
  SELECT count(*) filter(WHERE unread = true) unread_count,
    count(*) total
  FROM Data_CTE
),
Notifs_CTE AS (
  SELECT * FROM Data_CTE
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
SELECT nc.*, (cc.unread_count - rc.read_count)::bigint unread_count, cc.total
FROM Notifs_CTE nc
CROSS JOIN Count_CTE cc
CROSS JOIN ReadCount_CTE rc;

-- name: ListMessages :many
WITH Data_CTE AS (
  SELECT *
  FROM notifications WHERE kind = 'admin'
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.*, cc.total, u.username, u.email, u.avatar
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
JOIN users u ON u.id = dc.user_id
ORDER BY create_at DESC
LIMIT $1
OFFSET $2;

-- name: CheckMessage :exec
UPDATE notifications
SET unread = false
WHERE id = ANY(@ids::bigint[]);
