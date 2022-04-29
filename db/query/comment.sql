-- name: CreateComment :one
INSERT INTO comments (
    parent_id,
    article_id,
    commenter,
    content
) VALUES (
    CASE WHEN @set_parent_id::bool
        THEN @parent_id::bigint ELSE NULL END,
    @article_id::bigint,
    @commenter::varchar,
    @content::varchar
) RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1
    AND (@any_commenter::bool OR commenter = @commenter::varchar);

-- name: GetComment :one
SELECT * FROM comments
WHERE id = $1
LIMIT 1;

-- name: ListCommentsByArticle :many
SELECT
    id,
    parent_id,
    article_id,
    commenter,
    avatar_src,
    content,
    comment_at
FROM comments AS c
    JOIN users AS u
    ON c.commenter = u.username
WHERE article_id = @article_id::bigint
    AND parent_id IS NULL
ORDER BY comment_at
LIMIT $1
OFFSET $2;

-- name: ListChildComments :many
SELECT
    id,
    parent_id,
    article_id,
    commenter,
    avatar_src,
    content,
    comment_at
FROM comments AS c
    JOIN users AS u
    ON c.commenter = u.username
WHERE parent_id = ANY(@comment_ids::bigint[]);
