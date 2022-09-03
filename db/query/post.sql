-- name: CreatePost :one
INSERT INTO posts (
  author_id, title, abstract, cover_image, content
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: DeletePosts :execrows
DELETE FROM posts
WHERE id = ANY(@ids::bigint[]) AND author_id = @author_id::bigint
  AND status = ANY('{draft, revise}'::varchar[]);

-- name: UpdatePost :one
UPDATE posts
SET
  title = coalesce(sqlc.narg('title'), title),
  abstract = coalesce(sqlc.narg('abstract'), abstract),
  cover_image = coalesce(sqlc.narg('cover_image'), cover_image),
  content = coalesce(sqlc.narg('content'), content),
  update_at = now()
WHERE id = $1 AND author_id = @author_id::bigint
  AND status = ANY('{draft, revise}'::varchar[])
RETURNING *;

-- name: ListPosts :many
WITH Data_CTE AS (
  SELECT id, title, status, view_count, update_at, publish_at
  FROM posts
  WHERE author_id = @author_id::bigint
    AND (@any_status::bool OR status = @status::varchar)
    AND (@any_keyword::bool OR title LIKE @keyword::varchar)
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT * FROM Data_CTE
CROSS JOIN Count_CTE
ORDER BY
  CASE WHEN @update_at_asc::bool THEN update_at END ASC,
  CASE WHEN @update_at_desc::bool THEN update_at END DESC,
  CASE WHEN @publish_at_asc::bool THEN publish_at END ASC,
  CASE WHEN @publish_at_desc::bool THEN publish_at END DESC
LIMIT $1
OFFSET $2;

-- name: SubmitPost :many
UPDATE posts SET status = 'review'
WHERE id = ANY(@ids::bigint[]) AND author_id = @author_id::bigint
  AND status = ANY('{draft, revise}'::varchar[])
RETURNING id;

-- name: ReviewPost :one
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
  SELECT post_id,
      array_agg(tag_id)::bigint[] tag_ids,
      array_agg(name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t
    ON pt.tag_id = t.id AND pt.post_id = @post_id::bigint
  GROUP BY pt.post_id
)
SELECT p.id, title, abstract, cover_image, content, is_featured, status,
    cc.category_ids, cc.category_names, tc.tag_ids, tc.tag_names
FROM posts p
LEFT JOIN Category_CTE cc ON cc.post_id = p.id
LEFT JOIN Tag_CTE tc ON tc.post_id = p.id
WHERE p.id = @post_id::bigint
  AND (@is_admin::bool OR author_id = @author_id::bigint)
LIMIT 1;

-- name: PublishPost :many
UPDATE posts SET status = 'publish'
WHERE id = ANY(@ids::bigint[]) AND status = 'review'
RETURNING id, author_id;

-- name: WithdrawPost :many
UPDATE posts SET status = 'revise'
WHERE id = ANY(@ids::bigint[])
  AND status = ANY('{review, publish}'::varchar[])
RETURNING id, author_id;

-- name: FeaturePost :execrows
UPDATE posts SET is_featured = @is_featured::bool
WHERE id = @id::bigint;

-- name: GetPosts :many
WITH Data_CTE AS (
  SELECT id, title, author_id, status, is_featured,
      view_count, update_at, publish_at
  FROM posts
  WHERE (@any_status::bool OR status = @status::varchar)
    AND (@any_keyword::bool OR title LIKE @keyword::varchar)
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
),
Category_CTE AS (
  SELECT pc.post_id post_id,
      array_agg(pc.category_id)::bigint[] category_ids,
      array_agg(c.name)::varchar[] category_names
  FROM post_categories pc
  JOIN categories c ON c.id = pc.category_id
    AND pc.post_id = ANY(SELECT id FROM Data_CTE)
  GROUP BY pc.post_id
)
SELECT dc.*, cnt.total, u.username, u.email, u.avatar,
    cc.category_ids, cc.category_names
FROM Data_CTE dc
CROSS JOIN Count_CTE cnt
JOIN users u ON u.id = dc.author_id
LEFT JOIN Category_CTE cc ON cc.post_id = dc.id
ORDER BY
  CASE WHEN @update_at_asc::bool THEN update_at END ASC,
  CASE WHEN @update_at_desc::bool THEN update_at END DESC,
  CASE WHEN @publish_at_asc::bool THEN publish_at END ASC,
  CASE WHEN @publish_at_desc::bool THEN publish_at END DESC
LIMIT $1
OFFSET $2;

-- name: FetchPosts :many
WITH Data_CTE AS (
  SELECT id, title, author_id, abstract, cover_image,
      view_count, publish_at
  FROM posts
  WHERE status = 'publish'
    AND (@any_featured::bool OR is_featured = @is_featured::bool)
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
Tag_CTE AS (
  SELECT pt.post_id post_id,
      array_agg(pt.tag_id)::bigint[] tag_ids,
      array_agg(t.name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t ON t.id = pt.tag_id
    AND pt.post_id = ANY(SELECT id FROM Data_CTE)
  GROUP BY pt.post_id
)
SELECT dc.*, cnt.total, u.username, u.info, u.avatar,
    tc.tag_ids, tc.tag_names, fu.follower_id followed,
    (SELECT count(*) FROM follows f
      WHERE f.user_id = dc.author_id) follower_count,
    (SELECT count(*) FROM follows f
      WHERE f.follower_id = dc.author_id) following_count,
    (SELECT count(*) FROM comments cm
      WHERE cm.post_id = dc.id) comment_count,
    (SELECT count(*) FROM post_stars ps
      WHERE ps.post_id = dc.id) star_count
FROM Data_CTE dc
CROSS JOIN Count_CTE cnt
JOIN users u ON u.id = dc.author_id
LEFT JOIN Tag_CTE tc ON tc.post_id = dc.id
LEFT JOIN follows fu
  ON fu.user_id = dc.author_id AND fu.follower_id = @self_id::bigint
ORDER BY
  CASE WHEN @view_count_asc::bool THEN view_count END ASC,
  CASE WHEN @view_count_desc::bool THEN view_count END DESC,
  CASE WHEN @publish_at_asc::bool THEN publish_at END ASC,
  CASE WHEN @publish_at_desc::bool THEN publish_at END DESC
LIMIT $1
OFFSET $2;

-- name: ReadPost :one
WITH Post_CTE AS (
  UPDATE posts
  SET view_count = view_count + 1
  WHERE id = @post_id::bigint  AND status = 'publish'
  RETURNING *
),
Category_CTE AS (
  SELECT post_id,
      array_agg(category_id)::bigint[] category_ids,
      array_agg(name)::varchar[] category_names
  FROM post_categories pc
  JOIN categories c
    ON pc.category_id = c.id AND pc.post_id = @post_id::bigint
  GROUP BY pc.post_id
),
Tag_CTE AS (
  SELECT post_id,
      array_agg(tag_id)::bigint[] tag_ids,
      array_agg(name)::varchar[] tag_names
  FROM post_tags pt
  JOIN tags t
    ON pt.tag_id = t.id AND pt.post_id = @post_id::bigint
  GROUP BY pt.post_id
)
SELECT p.id, p.title, p.cover_image, p.content, p.view_count,
    p.publish_at, p.author_id, u.username, u.avatar, u.info,
    cc.category_ids, cc.category_names, tc.tag_ids,
    tc.tag_names, fu.follower_id followed,
    (SELECT count(*) FROM follows f
      WHERE f.user_id = p.author_id) follower_count,
    (SELECT count(*) FROM follows f
      WHERE f.follower_id = p.author_id) following_count,
    (SELECT count(*) FROM post_stars ps
      WHERE ps.post_id = @post_id::bigint) star_count
FROM Post_CTE p
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