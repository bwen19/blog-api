-- name: CreateCategory :one
INSERT INTO categories ( name ) VALUES ( $1 )
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE name = $1;

-- name: GetCategory :one
SELECT * FROM categories
WHERE name = $1
LIMIT 1;

-- name: ListCategories :many
SELECT * FROM categories
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET name = @new_name::varchar
WHERE name = @name::varchar
RETURNING *;
