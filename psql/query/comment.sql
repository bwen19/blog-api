-- name: CreateComment :one
WITH Ins_CTE AS (
  INSERT INTO comments (
      post_id, user_id, parent_id, reply_user_id, content
  )
  VALUES ($1, $2, $3, $4, $5) RETURNING *
)
SELECT ic.*, p.author_id, ru.username r_username,
    ru.avatar r_avatar, ru.intro r_intro, fu.user_id r_followed,
    (SELECT count(*) FROM follows f
      WHERE f.user_id = ic.reply_user_id) r_follower_count,
    (SELECT count(*) FROM follows f
      WHERE f.follower_id = ic.reply_user_id) r_following_count,
    (SELECT count(*) FROM follows f
      WHERE f.user_id = ic.user_id) follower_count,
    (SELECT count(*) FROM follows f
      WHERE f.follower_id = ic.user_id ) following_count
FROM Ins_CTE ic
JOIN posts p ON p.id = ic.post_id
LEFT JOIN users ru ON ru.id = ic.reply_user_id
LEFT JOIN follows fu
  ON fu.user_id = ic.reply_user_id AND fu.follower_id = ic.user_id;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1 AND (@is_admin::bool OR user_id = @user_id::bigint);

-- name: ListComments :many
WITH Data_CTE AS (
  SELECT cm.*, count(cms.user_id) star_count
  FROM comments cm
  LEFT JOIN comment_stars cms ON cms.comment_id = cm.id
  WHERE post_id = @post_id::bigint
  GROUP BY id, post_id, cm.user_id, parent_id, reply_user_id,
      content, create_at
),
Count_CTE AS (
  SELECT sum(CASE WHEN parent_id IS NULL THEN 1 ELSE 0 END) total,
      count(*) comment_count
  FROM Data_CTE
),
Comment_CTE AS (
  SELECT id, parent_id, user_id, content, star_count, create_at
  FROM Data_CTE
  WHERE parent_id IS NULL
  ORDER BY
    CASE WHEN @star_count_asc::bool THEN star_count END ASC,
    CASE WHEN @star_count_desc::bool THEN star_count END DESC,
    CASE WHEN @create_at_asc::bool THEN create_at END ASC,
    CASE WHEN @create_at_desc::bool THEN create_at END DESC
  LIMIT $1
  OFFSET $2
),
Reply_CTE AS (
  SELECT *,
      row_number() OVER (PARTITION BY parent_id ORDER BY star_count DESC) AS gr
  FROM Data_CTE
  WHERE parent_id IN (SELECT id FROM Comment_CTE)
)
SELECT cc.id, cc.parent_id, cc.content, cc.star_count, cc.create_at,
    cnt.total, cnt.comment_count,
    (SELECT count(*) FROM Reply_CTE rc WHERE rc.parent_id = cc.id) reply_count,
    u.id user_id, u.username, u.avatar, u.intro, fu.follower_id followed,
    (SELECT count(*) FROM follows f WHERE f.user_id = u.id) follower_count,
    (SELECT count(*) FROM follows f WHERE f.follower_id = u.id) following_count,
    NULL::bigint r_user_id, NULL::varchar r_username, NULL::varchar r_avatar,
    NULL::varchar r_intro, NULL::bigint r_followed, 0::bigint r_follower_count,
    0::bigint r_following_count
FROM Comment_CTE cc
CROSS JOIN Count_CTE cnt
JOIN users u ON u.id = cc.user_id
LEFT JOIN follows fu
  ON fu.user_id = cc.user_id AND fu.follower_id = @self_id::bigint
UNION
SELECT rc.id, rc.parent_id, rc.content, rc.star_count, rc.create_at,
    cnt.total, cnt.comment_count, 0::bigint reply_count,
    u.id user_id, u.username, u.avatar, u.intro, fu.follower_id followed,
    (SELECT count(*) FROM follows f WHERE f.user_id = u.id) follower_count,
    (SELECT count(*) FROM follows f WHERE f.follower_id = u.id) following_count,
    ru.id r_user_id, ru.username r_username, ru.avatar r_avatar, ru.intro r_intro,
    fr.follower_id r_followed,
    (SELECT count(*) FROM follows f WHERE f.user_id = ru.id) r_follower_count,
    (SELECT count(*) FROM follows f WHERE f.follower_id = ru.id) r_following_count
FROM Reply_CTE rc
CROSS JOIN Count_CTE cnt
JOIN users u ON u.id = rc.user_id
LEFT JOIN users ru ON ru.id = rc.reply_user_id
LEFT JOIN follows fu
  ON fu.user_id = rc.user_id AND fu.follower_id = @self_id::bigint
LEFT JOIN follows fr
  ON fr.user_id = rc.user_id AND fr.follower_id = @self_id::bigint
WHERE rc.gr < 3;

-- name: ListReplies :many
WITH Data_CTE AS (
  SELECT cm.*, count(cms.user_id) star_count
  FROM comments cm
  LEFT JOIN comment_stars cms ON cms.comment_id = cm.id
  WHERE cm.parent_id = @parent_id::bigint
  GROUP BY id, post_id, cm.user_id, parent_id, reply_user_id, content, create_at
),
Count_CTE AS (
  SELECT count(*) total FROM Data_CTE
)
SELECT dc.id, dc.content, dc.star_count, dc.create_at, cc.total,
    u.id user_id, u.username, u.avatar, u.intro, fu.follower_id followed,
    (SELECT count(*) FROM follows f WHERE f.user_id = u.id) follower_count,
    (SELECT count(*) FROM follows f WHERE f.follower_id = u.id) following_count,
    ru.id r_user_id, ru.username r_username, ru.avatar r_avatar, ru.intro r_intro,
    fr.follower_id r_followed,
    (SELECT count(*) FROM follows f WHERE f.user_id = ru.id) r_follower_count,
    (SELECT count(*) FROM follows f WHERE f.follower_id = ru.id) r_following_count
FROM Data_CTE dc
CROSS JOIN Count_CTE cc
JOIN users u ON u.id = dc.user_id
LEFT JOIN users ru ON ru.id = dc.reply_user_id
LEFT JOIN follows fu
  ON fu.user_id = dc.user_id AND fu.follower_id = @self_id::bigint
LEFT JOIN follows fr
  ON fr.user_id = dc.reply_user_id AND fr.follower_id = @self_id::bigint
ORDER BY
  CASE WHEN @star_count_asc::bool THEN star_count END ASC,
  CASE WHEN @star_count_desc::bool THEN star_count END DESC,
  CASE WHEN @create_at_asc::bool THEN dc.create_at END ASC,
  CASE WHEN @create_at_desc::bool THEN dc.create_at END DESC
LIMIT $1
OFFSET $2;

-- name: CreateCommentStar :exec
INSERT INTO comment_stars (comment_id, user_id)
VALUES ($1, $2);

-- name: DeleteCommentStar :exec
DELETE FROM comment_stars
WHERE comment_id = $1 AND user_id = $2;