-- name: CreateUser :one
INSERT INTO users (
  username, email, hashed_password, avatar, role
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: DeleteUsers :execrows
DELETE FROM users WHERE id = ANY(@ids::bigint[]);

-- name: UpdateUser :one
UPDATE users
SET
  username = CASE WHEN @set_username::bool
    THEN @username::varchar
    ELSE username END,
  email = CASE WHEN @set_email::bool
    THEN @email::varchar
    ELSE email END,
  hashed_password = CASE WHEN @set_password::bool
    THEN @hashed_password::varchar
    ELSE hashed_password END,
  avatar = CASE WHEN @set_avatar::bool
    THEN @avatar::varchar
    ELSE avatar END,
  info = CASE WHEN @set_info::bool
    THEN @info::varchar
    ELSE info END,
  role = CASE WHEN @set_role::bool
    THEN @role::varchar
    ELSE role END,
  is_deleted = CASE WHEN @set_deleted::bool
    THEN @is_deleted::bool
    ELSE is_deleted END
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
WITH Data_CTE AS (
  SELECT id, username, email, avatar, role, is_deleted, create_at
  FROM users
  WHERE @any_keyword::bool
    OR username LIKE @keyword::varchar
    OR email LIKE @keyword::varchar
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.*, cc.total, (
    SELECT count(*) FROM posts WHERE posts.author_id = dc.id
  ) post_count
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
ORDER BY
  CASE WHEN @username_asc::bool THEN username END ASC,
  CASE WHEN @username_desc::bool THEN username END DESC,
  CASE WHEN @role_asc::bool THEN role END ASC,
  CASE WHEN @role_desc::bool THEN role END DESC,
  CASE WHEN @deleted_asc::bool THEN is_deleted END ASC,
  CASE WHEN @deleted_desc::bool THEN is_deleted END DESC,
  CASE WHEN @create_at_asc::bool THEN create_at END ASC,
  CASE WHEN @create_at_desc::bool THEN create_at END DESC
LIMIT $1
OFFSET $2;

-- name: GetUserProfile :one
WITH SUM_CTE AS (
  SELECT coalesce(sum(view_count), 0)::bigint view_count,
    coalesce(sum(star_count), 0)::bigint star_count
  FROM (
    SELECT p.id, p.view_count, count(ps.user_id) star_count
    FROM posts p
    LEFT JOIN post_stars ps ON ps.post_id = p.id
    WHERE p.author_id = @user_id::bigint
    GROUP BY p.id, p.view_count
  ) pc
)
SELECT u.id, u.username, u.avatar, u.info, sc.view_count,
  sc.star_count, fu.follower_id followed,
  (SELECT count(*) FROM follows f
    WHERE f.user_id = @user_id::bigint) follower_count,
  (SELECT count(*) FROM follows f
    WHERE f.follower_id = @user_id::bigint) following_count
FROM users u
CROSS JOIN SUM_CTE sc
LEFT JOIN follows fu
  ON fu.user_id = u.id AND fu.follower_id = @self_id::bigint
WHERE u.id = @user_id::bigint LIMIT 1;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;
