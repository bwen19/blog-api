-- name: CreateTag :one
INSERT INTO tags (name) VALUES ($1) RETURNING *;

-- name: DeleteTags :execrows
DELETE FROM tags WHERE id = ANY(@ids::bigint[]);

-- name: UpdateTag :one
UPDATE tags SET name = @name::varchar
WHERE id = $1 RETURNING *;

-- name: ListTags :many
WITH Data_CTE AS (
  SELECT t.id, t.name, count(pt.post_id) post_count
  FROM tags t
  LEFT JOIN post_tags pt ON pt.tag_id = t.id
  GROUP BY t.id, t.name
),
Count_CTE AS (
  SELECT count(*) AS total FROM Data_CTE
)
SELECT *
FROM Data_CTE
CROSS JOIN Count_CTE
ORDER BY
  CASE WHEN @name_asc::bool THEN name END ASC,
  CASE WHEN @name_desc::bool THEN name END DESC,
  CASE WHEN @post_count_asc::bool THEN post_count END ASC,
  CASE WHEN @post_count_desc::bool THEN post_count END DESC,
  id ASC
LIMIT $1
OFFSET $2;

-- name: GetTagsByName :one
SELECT * FROM tags
WHERE name = @name::varchar LIMIT 1;

-- name: CreatePostTags :many
WITH Tag_CTE AS (
  SELECT * FROM tags
  WHERE id = ANY(@tag_ids::bigint[])
),
Values_CTE AS (
  SELECT p.post_id, tc.id tag_id FROM (
    SELECT id post_id FROM posts WHERE id = @post_id::bigint
  ) p
  CROSS JOIN Tag_CTE tc
),
Ins_CTE AS (
  INSERT INTO post_tags (post_id, tag_id)
  SELECT * FROM Values_CTE
  ON CONFLICT (post_id, tag_id) DO NOTHING
)
SELECT * FROM Tag_CTE;

-- name: DeletePostTags :exec
DELETE FROM post_tags
WHERE post_id = @post_id::bigint
  AND tag_id <> ALL(@tag_ids::bigint[]);

-- name: GetPostTags :many
SELECT * FROM tags
WHERE id = ANY(
  SELECT tag_id FROM post_tags
  WHERE post_id = $1
);
