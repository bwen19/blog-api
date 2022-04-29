-- name: CreateTag :one
INSERT INTO tags ( name ) VALUES ( $1 )
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE name = $1;

-- name: GetTag :one
SELECT * FROM tags
WHERE name = $1
LIMIT 1;

-- name: ListTags :many
SELECT * FROM tags
ORDER BY count DESC
LIMIT $1
OFFSET $2;

-- name: UpdateTag :one
UPDATE tags
SET
    name = CASE WHEN @set_new_name::bool
        THEN @new_name::varchar
        ELSE name END,
    count = CASE WHEN @set_count::bool THEN @count::bigint
        WHEN @add_count::bool THEN count + 1
        WHEN @minus_count::bool THEN count - 1
        ELSE count END
WHERE name = @name::varchar
RETURNING *;

-- name: CreateArticleTag :one
INSERT INTO article_tags (
    article_id,
    tag
) VALUES (
    $1, $2
) RETURNING *;

-- name: DeleteArticleTag :exec
DELETE FROM article_tags
WHERE article_id = $1
    AND tag = $2;

-- name: ListArticleTags :many
SELECT tag FROM article_tags
WHERE article_id = $1;

