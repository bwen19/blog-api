-- name: CreateSession :one
INSERT INTO sessions (
  id, user_id, refresh_token, user_agent, client_ip, expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = @id::uuid AND user_id = @user_id::bigint;

-- name: DeleteSessions :execrows
DELETE FROM sessions
WHERE id = ANY(@ids::uuid[]) AND user_id = @user_id::bigint;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < now() - interval '30 days';

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: ListSessions :many
WITH Data_CTE AS (
  SELECT id, user_agent, client_ip, expires_at, create_at
  FROM sessions
  WHERE user_id = @user_id::bigint
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT * FROM Data_CTE
CROSS JOIN Count_CTE
ORDER BY
  CASE WHEN @client_ip_asc::bool THEN client_ip END ASC,
  CASE WHEN @client_ip_desc::bool THEN client_ip END DESC,
  CASE WHEN @create_at_asc::bool THEN create_at END ASC,
  CASE WHEN @create_at_desc::bool THEN create_at END DESC,
  CASE WHEN @expires_at_asc::bool THEN expires_at END ASC,
  CASE WHEN @expires_at_desc::bool THEN expires_at END DESC
LIMIT $1
OFFSET $2;
