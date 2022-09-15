-- name: CreateUser :one
INSERT INTO users (username, email, hashed_password, avatar, role)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteUsers :execrows
DELETE FROM users WHERE id = ANY(@ids::bigint[]);

-- name: UpdateUser :one
UPDATE users
SET
  username = coalesce(sqlc.narg('username'), username),
  email = coalesce(sqlc.narg('email'), email),
  hashed_password = coalesce(sqlc.narg('hashed_password'), hashed_password),
  avatar = coalesce(sqlc.narg('avatar'), avatar),
  intro = coalesce(sqlc.narg('intro'), intro),
  role = coalesce(sqlc.narg('role'), role),
  deleted = coalesce(sqlc.narg('deleted'), deleted)
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
WITH Data_CTE AS (
  SELECT id, username, email, avatar, role, deleted, create_at
  FROM users
  WHERE @any_keyword::bool
    OR username LIKE @keyword::varchar
    OR email LIKE @keyword::varchar
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.*, cc.total,
    (SELECT count(*) FROM posts WHERE posts.author_id = dc.id) post_count
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
ORDER BY
  CASE WHEN @username_asc::bool THEN username END ASC,
  CASE WHEN @username_desc::bool THEN username END DESC,
  CASE WHEN @role_asc::bool THEN role END ASC,
  CASE WHEN @role_desc::bool THEN role END DESC,
  CASE WHEN @deleted_asc::bool THEN deleted END ASC,
  CASE WHEN @deleted_desc::bool THEN deleted END DESC,
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
SELECT u.id, u.username, u.avatar, u.intro, sc.view_count,
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
