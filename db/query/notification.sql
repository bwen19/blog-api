-- name: CreateNotification :exec
INSERT INTO notifications (
  user_id, kind, title, content
) VALUES (
  $1, $2, $3, $4
);

-- name: MarkAllRead :exec
UPDATE notifications
SET unread = false
WHERE user_id = @user_id::bigint AND unread = true;

-- name: DeleteNotifications :execrows
DELETE FROM notifications
WHERE id = ANY(@ids::bigint[]) AND user_id = @user_id::bigint;

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
)
SELECT * FROM Data_CTE
CROSS JOIN Count_CTE
ORDER BY create_at DESC
LIMIT $1
OFFSET $2;

-- name: MarkReadByIDs :execrows
UPDATE notifications
SET unread = false
WHERE user_id = @user_id::bigint AND id = ANY(@ids::bigint[])
  AND unread = true;

-- name: GetUnreadCount :one
SELECT count(*) FROM notifications
WHERE user_id = @user_id::bigint AND unread = true;

