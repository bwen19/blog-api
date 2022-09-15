-- name: CreateFollow :exec
INSERT INTO follows (user_id, follower_id)
VALUES ($1, $2);

-- name: DeleteFollow :exec
DELETE FROM follows
WHERE user_id = $1 AND follower_id = $2;

-- name: ListFollowers :many
WITH Data_CTE AS (
  SELECT user_id, follower_id, create_at FROM follows
  WHERE user_id = @user_id::bigint
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.follower_id user_id, u.username, u.avatar,
    u.intro, cc.total, f.user_id followed
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
JOIN users u ON u.id = dc.follower_id
LEFT JOIN follows f
  ON f.user_id = dc.follower_id AND f.follower_id = @self_id::bigint
ORDER BY dc.create_at DESC
LIMIT $1 OFFSET $2;

-- name: ListFollowings :many
WITH Data_CTE AS (
  SELECT user_id, follower_id, create_at FROM follows
  WHERE follower_id = @user_id::bigint
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.user_id user_id, u.username, u.avatar,
    u.intro, cc.total, f.user_id followed
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
JOIN users u ON u.id = dc.user_id
LEFT JOIN follows f
  ON f.user_id = dc.user_id AND f.follower_id = @self_id::bigint
ORDER BY dc.create_at DESC
LIMIT $1 OFFSET $2;
