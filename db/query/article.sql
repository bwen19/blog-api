-- name: CreateArticle :one
INSERT INTO articles (
    author,
    category,
    title,
    summary,
    content
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM articles
WHERE id = $1
    AND (@any_author::bool OR author = @author::varchar)
    AND (@any_status::bool OR status = @status::varchar);

-- name: GetArticle :one
SELECT * FROM articles
WHERE id = $1
    AND (@any_author::bool OR author = @author::varchar)
LIMIT 1;

-- name: ListArticles :many
SELECT
    id,
    author,
    category,
    title,
    summary,
    status,
    view_count,
    update_at
FROM articles
WHERE (@any_status::bool OR status = @status::varchar)
    AND (@any_author::bool OR author = @author::varchar)
    AND (@any_category::bool OR category = @category::varchar)
    AND (@any_tag::bool OR id IN (
        SELECT article_id
        FROM article_tags
        WHERE tag = @tag::varchar
    ))
ORDER BY
    CASE WHEN @time_desc::bool THEN update_at END DESC,
    CASE WHEN @count_desc::bool THEN view_count END DESC
LIMIT $1
OFFSET $2;

-- name: ReadArticle :one
UPDATE articles
SET view_count = view_count + 1
WHERE id = $1
    AND status = 'published'
RETURNING *;

-- name: UpdateArticle :one
UPDATE articles
SET
    author = CASE WHEN @set_author::bool
        THEN @author::varchar
        ELSE author END,
    category = CASE WHEN @set_category::bool
        THEN @category::varchar
        ELSE category END,
    title = CASE WHEN @set_title::bool
        THEN @title::varchar
        ELSE title END,
    summary = CASE WHEN @set_summary::bool
        THEN @summary::varchar
        ELSE summary END,
    content = CASE WHEN @set_content::bool
        THEN @content::text
        ELSE content END,
    status = CASE WHEN @set_status::bool
        THEN @status::varchar
        ELSE status END,
    view_count = CASE WHEN @set_view_count::bool
        THEN @view_count::bigint
        ELSE view_count END,
    update_at = now()
WHERE id = $1
    AND (@any_author::bool OR author = @username::varchar)
RETURNING *;
