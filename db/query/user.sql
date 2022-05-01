-- name: CreateUser :one
INSERT INTO users (
    username,
    hashed_password,
    email,
    avatar_src
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE username = $1;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1
  OR email = $2
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET
  username = CASE WHEN @set_new_name::bool
    THEN @new_name::varchar
    ELSE username END,
  hashed_password = CASE WHEN @set_hashed_password::bool
    THEN @hashed_password::varchar
    ELSE hashed_password END,
  email = CASE WHEN @set_email::bool
    THEN @email::varchar
    ELSE email END,
  role = CASE WHEN @set_role::bool
    THEN @role::varchar
    ELSE role END,
  avatar_src = CASE WHEN @set_avatar_src::bool
    THEN @avatar_src::varchar
    ELSE avatar_src END
WHERE username = $1
RETURNING *;
