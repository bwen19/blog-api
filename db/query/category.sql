-- name: CreateCategory :one
INSERT INTO categories (name) VALUES ($1) RETURNING *;

-- name: DeleteCategories :execrows
DELETE FROM categories WHERE id = ANY(@ids::bigint[]);

-- name: UpdateCategory :one
UPDATE categories SET name = @name::varchar
WHERE id = $1 RETURNING *;

-- name: ListCategories :many
SELECT c.id, c.name,
  (SELECT count(*) FROM post_categories
    WHERE category_id = c.id) post_count
FROM categories c
ORDER BY
  CASE WHEN @name_asc::bool THEN name END ASC,
  CASE WHEN @name_desc::bool THEN name END DESC,
  id ASC;

-- name: GetCategories :many
SELECT * FROM categories
ORDER BY id ASC;

-- name: SetPostCategories :many
WITH Category_CTE AS (
  SELECT DISTINCT * FROM categories
  WHERE id = ANY(@category_ids::bigint[])
), Values_CTE AS (
  SELECT p.post_id, cc.id category_id FROM (
    SELECT id post_id FROM posts
    WHERE id = @post_id::bigint
  ) p
  CROSS JOIN Category_CTE cc
), Del_CTE AS (
  DELETE FROM post_categories
  WHERE post_id = @post_id::bigint
    AND category_id NOT IN (SELECT id FROM Category_CTE)
), Ins_CTE AS (
  INSERT INTO post_categories (post_id, category_id)
  SELECT * FROM Values_CTE
  ON CONFLICT (post_id, category_id) DO NOTHING
)
SELECT * FROM Category_CTE;

-- name: GetPostCategories :many
SELECT * FROM categories
WHERE id = ANY(
  SELECT category_id FROM post_categories
  WHERE post_id = $1
);
