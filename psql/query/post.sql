-- name: CreatePost :one
INSERT INTO posts (author_id, title, cover_image)
VALUES ($1, $2, $3) RETURNING *;

-- name: CreatePostContent :one
INSERT INTO post_contents (id, content)
VALUES ($1, $2) RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = @id::bigint AND author_id = @author_id::bigint
  AND status = ANY('{draft, revise}'::varchar[]);

-- name: UpdatePost :one
UPDATE posts
SET
  title = coalesce(sqlc.narg('title'), title),
  cover_image = coalesce(sqlc.narg('cover_image'), cover_image),
  update_at = now()
WHERE id = $1 AND author_id = @author_id::bigint
  AND status = ANY('{draft, revise}'::varchar[])
RETURNING *;

-- name: UpdatePostContent :one
UPDATE post_contents
SET content = $2
WHERE id = (
  SELECT p.id FROM posts p
  WHERE p.id = $1
    AND author_id = @author_id::bigint
    AND status = ANY('{draft, revise}'::varchar[])
) RETURNING *;

-- name: UpdatePostStatus :many
UPDATE posts SET status = @status::varchar
WHERE id = ANY(@ids::bigint[]) AND status = ANY(@old_status::varchar[])
  AND (@is_admin::bool OR author_id = @author_id::bigint)
RETURNING *;

-- name: UpdatePostFeature :exec
UPDATE posts SET featured = @featured::bool
WHERE id = @id::bigint;

-- name: ListPosts :many
WITH Data_CTE AS (
  SELECT * FROM posts
  WHERE (@is_admin::bool OR author_id = @author_id::bigint)
    AND (@any_status::bool OR status = @status::varchar)
    AND (@any_keyword::bool OR title LIKE @keyword::varchar)
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
),
Post_CTE AS (
  SELECT * FROM Data_CTE
  ORDER BY
    CASE WHEN @update_at_asc::bool THEN update_at END ASC,
    CASE WHEN @update_at_desc::bool THEN update_at END DESC,
    CASE WHEN @publish_at_asc::bool THEN publish_at END ASC,
    CASE WHEN @publish_at_desc::bool THEN publish_at END DESC,
    CASE WHEN @view_count_asc::bool THEN view_count END ASC,
    CASE WHEN @view_count_desc::bool THEN view_count END DESC,
    id ASC
  LIMIT $1
  OFFSET $2
),
Category_CTE AS (
  SELECT pc.post_id,
      array_agg(c.id)::bigint[] category_ids,
      array_agg(c.name)::varchar[] category_names
  FROM post_categories pc
  JOIN categories c
    ON pc.category_id = c.id
    AND pc.post_id = ANY(SELECT id FROM Post_CTE)
  GROUP BY pc.post_id
),
Tag_CTE AS (
  SELECT pt.post_id,
      array_agg(t.id)::bigint[] tag_ids,
      array_agg(t.name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t
    ON pt.tag_id = t.id
    AND pt.post_id = ANY(SELECT id FROM Post_CTE)
  GROUP BY pt.post_id
)
SELECT p.*, cnt.total, u.username, u.avatar,
      cc.category_ids, cc.category_names,
      tc.tag_ids, tc.tag_names
FROM Post_CTE p
CROSS JOIN Count_CTE cnt
JOIN users u ON u.id = p.author_id
LEFT JOIN Category_CTE cc ON cc.post_id = p.id
LEFT JOIN Tag_CTE tc ON tc.post_id = p.id;

-- name: GetPost :one
WITH Category_CTE AS (
  SELECT pc.post_id,
      array_agg(pc.category_id)::bigint[] category_ids,
      array_agg(c.name)::varchar[] category_names
  FROM post_categories pc
  JOIN categories c
    ON pc.category_id = c.id AND pc.post_id = @post_id::bigint
  GROUP BY pc.post_id
),
Tag_CTE AS (
  SELECT pt.post_id,
      array_agg(pt.tag_id)::bigint[] tag_ids,
      array_agg(t.name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t
    ON pt.tag_id = t.id AND pt.post_id = @post_id::bigint
  GROUP BY pt.post_id
)
SELECT p.*, pc.content, cc.category_ids, cc.category_names,
    tc.tag_ids, tc.tag_names
FROM posts p
JOIN post_contents pc ON pc.id = p.id
LEFT JOIN Category_CTE cc ON cc.post_id = p.id
LEFT JOIN Tag_CTE tc ON tc.post_id = p.id
WHERE p.id = @post_id::bigint
  AND (@is_admin::bool OR author_id = @author_id::bigint)
LIMIT 1;

-- name: GetFeaturedPosts :many
SELECT p.id, p.title, p.cover_image, p.view_count,
    p.publish_at, p.author_id, u.username, u.avatar,
    (SELECT count(*) FROM comments cm
      WHERE cm.post_id = p.id) comment_count,
    (SELECT count(*) FROM post_stars ps
      WHERE ps.post_id = p.id) star_count
FROM posts p
JOIN users u ON u.id = p.author_id
WHERE featured = true AND status = 'publish'
ORDER BY random()
LIMIT $1;

-- name: GetPosts :many
WITH Data_CTE AS (
  SELECT id, title, author_id, cover_image,
      featured, view_count, publish_at
  FROM posts
  WHERE status = 'publish'
    AND (@any_featured::bool OR featured = @featured::bool)
    AND (@any_author::bool OR author_id = @author_id::bigint)
    AND (@any_category::bool OR id = ANY(
      SELECT post_id FROM post_categories
      WHERE category_id = @category_id::bigint
    ))
    AND (@any_tag::bool OR id = ANY(
      SELECT post_id FROM post_tags
      WHERE tag_id = @tag_id::bigint
    ))
    AND (@any_keyword::bool OR title LIKE @keyword::varchar)
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
),
Post_CTE AS (
  SELECT * FROM Data_CTE
  ORDER BY
    CASE WHEN @publish_at_asc::bool THEN publish_at END ASC,
    CASE WHEN @publish_at_desc::bool THEN publish_at END DESC,
    CASE WHEN @view_count_asc::bool THEN view_count END ASC,
    CASE WHEN @view_count_desc::bool THEN view_count END DESC,
    id ASC
  LIMIT $1
  OFFSET $2
),
Tag_CTE AS (
  SELECT pt.post_id,
      array_agg(t.id)::bigint[] tag_ids,
      array_agg(t.name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t
    ON pt.tag_id = t.id
    AND pt.post_id = ANY(SELECT id FROM Post_CTE)
  GROUP BY pt.post_id
)
SELECT p.*, cnt.total, u.username, u.avatar,
    tc.tag_ids, tc.tag_names,
    (SELECT count(*) FROM comments cm
      WHERE cm.post_id = p.id) comment_count,
    (SELECT count(*) FROM post_stars ps
      WHERE ps.post_id = p.id) star_count
FROM Post_CTE p
CROSS JOIN Count_CTE cnt
JOIN users u ON u.id = p.author_id
LEFT JOIN Tag_CTE tc ON tc.post_id = p.id;

-- name: ReadPost :one
WITH Post_CTE AS (
  UPDATE posts SET view_count = view_count + 1
  WHERE id = @post_id::bigint AND status = 'publish'
  RETURNING *
),
Category_CTE AS (
  SELECT pc.post_id,
      array_agg(c.id)::bigint[] category_ids,
      array_agg(c.name)::varchar[] category_names
  FROM post_categories pc
  JOIN categories c
    ON pc.category_id = c.id AND pc.post_id = @post_id::bigint
  GROUP BY pc.post_id
),
Tag_CTE AS (
  SELECT pt.post_id,
      array_agg(t.id)::bigint[] tag_ids,
      array_agg(t.name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t
    ON pt.tag_id = t.id AND pt.post_id = @post_id::bigint
  GROUP BY pt.post_id
)
SELECT p.id, p.title, pc.content, p.view_count, p.publish_at,
    p.author_id, u.username, u.avatar, u.intro,
    cc.category_ids, cc.category_names, tc.tag_ids,
    tc.tag_names, fu.follower_id followed,
    (SELECT count(*) FROM follows f
      WHERE f.user_id = p.author_id) follower_count,
    (SELECT count(*) FROM follows f
      WHERE f.follower_id = p.author_id) following_count,
    (SELECT count(*) FROM post_stars ps
      WHERE ps.post_id = @post_id::bigint) star_count
FROM Post_CTE p
JOIN post_contents pc ON pc.id = p.id
JOIN users u ON u.id = p.author_id
LEFT JOIN Category_CTE cc ON cc.post_id = p.id
LEFT JOIN Tag_CTE tc ON tc.post_id = p.id
LEFT JOIN follows fu
  ON fu.user_id = p.author_id AND fu.follower_id = @self_id::bigint;

-- name: CreatePostStar :exec
INSERT INTO post_stars (post_id, user_id)
VALUES ($1, $2);

-- name: DeletePostStar :exec
DELETE FROM post_stars
WHERE post_id = $1 AND user_id = $2;