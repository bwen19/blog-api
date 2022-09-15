// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: comment.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const createComment = `-- name: CreateComment :one
WITH Ins_CTE AS (
  INSERT INTO comments (
      post_id, user_id, parent_id, reply_user_id, content
  )
  VALUES ($1, $2, $3, $4, $5) RETURNING id, post_id, user_id, parent_id, reply_user_id, content, create_at
)
SELECT ic.id, ic.post_id, ic.user_id, ic.parent_id, ic.reply_user_id, ic.content, ic.create_at, p.author_id, ru.username r_username,
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
  ON fu.user_id = ic.reply_user_id AND fu.follower_id = ic.user_id
`

type CreateCommentParams struct {
	PostID      int64         `json:"post_id"`
	UserID      int64         `json:"user_id"`
	ParentID    sql.NullInt64 `json:"parent_id"`
	ReplyUserID sql.NullInt64 `json:"reply_user_id"`
	Content     string        `json:"content"`
}

type CreateCommentRow struct {
	ID              int64          `json:"id"`
	PostID          int64          `json:"post_id"`
	UserID          int64          `json:"user_id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	ReplyUserID     sql.NullInt64  `json:"reply_user_id"`
	Content         string         `json:"content"`
	CreateAt        time.Time      `json:"create_at"`
	AuthorID        int64          `json:"author_id"`
	RUsername       sql.NullString `json:"r_username"`
	RAvatar         sql.NullString `json:"r_avatar"`
	RIntro          sql.NullString `json:"r_intro"`
	RFollowed       sql.NullInt64  `json:"r_followed"`
	RFollowerCount  int64          `json:"r_follower_count"`
	RFollowingCount int64          `json:"r_following_count"`
	FollowerCount   int64          `json:"follower_count"`
	FollowingCount  int64          `json:"following_count"`
}

func (q *Queries) CreateComment(ctx context.Context, arg CreateCommentParams) (CreateCommentRow, error) {
	row := q.db.QueryRowContext(ctx, createComment,
		arg.PostID,
		arg.UserID,
		arg.ParentID,
		arg.ReplyUserID,
		arg.Content,
	)
	var i CreateCommentRow
	err := row.Scan(
		&i.ID,
		&i.PostID,
		&i.UserID,
		&i.ParentID,
		&i.ReplyUserID,
		&i.Content,
		&i.CreateAt,
		&i.AuthorID,
		&i.RUsername,
		&i.RAvatar,
		&i.RIntro,
		&i.RFollowed,
		&i.RFollowerCount,
		&i.RFollowingCount,
		&i.FollowerCount,
		&i.FollowingCount,
	)
	return i, err
}

const createCommentStar = `-- name: CreateCommentStar :exec
INSERT INTO comment_stars (comment_id, user_id)
VALUES ($1, $2)
`

type CreateCommentStarParams struct {
	CommentID int64 `json:"comment_id"`
	UserID    int64 `json:"user_id"`
}

func (q *Queries) CreateCommentStar(ctx context.Context, arg CreateCommentStarParams) error {
	_, err := q.db.ExecContext(ctx, createCommentStar, arg.CommentID, arg.UserID)
	return err
}

const deleteComment = `-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1 AND ($2::bool OR user_id = $3::bigint)
`

type DeleteCommentParams struct {
	ID      int64 `json:"id"`
	IsAdmin bool  `json:"is_admin"`
	UserID  int64 `json:"user_id"`
}

func (q *Queries) DeleteComment(ctx context.Context, arg DeleteCommentParams) error {
	_, err := q.db.ExecContext(ctx, deleteComment, arg.ID, arg.IsAdmin, arg.UserID)
	return err
}

const deleteCommentStar = `-- name: DeleteCommentStar :exec
DELETE FROM comment_stars
WHERE comment_id = $1 AND user_id = $2
`

type DeleteCommentStarParams struct {
	CommentID int64 `json:"comment_id"`
	UserID    int64 `json:"user_id"`
}

func (q *Queries) DeleteCommentStar(ctx context.Context, arg DeleteCommentStarParams) error {
	_, err := q.db.ExecContext(ctx, deleteCommentStar, arg.CommentID, arg.UserID)
	return err
}

const listComments = `-- name: ListComments :many
WITH Data_CTE AS (
  SELECT cm.id, cm.post_id, cm.user_id, cm.parent_id, cm.reply_user_id, cm.content, cm.create_at, count(cms.user_id) star_count
  FROM comments cm
  LEFT JOIN comment_stars cms ON cms.comment_id = cm.id
  WHERE post_id = $3::bigint
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
    CASE WHEN $4::bool THEN star_count END ASC,
    CASE WHEN $5::bool THEN star_count END DESC,
    CASE WHEN $6::bool THEN create_at END ASC,
    CASE WHEN $7::bool THEN create_at END DESC
  LIMIT $1
  OFFSET $2
),
Reply_CTE AS (
  SELECT id, post_id, user_id, parent_id, reply_user_id, content, create_at, star_count,
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
  ON fu.user_id = cc.user_id AND fu.follower_id = $8::bigint
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
  ON fu.user_id = rc.user_id AND fu.follower_id = $8::bigint
LEFT JOIN follows fr
  ON fr.user_id = rc.user_id AND fr.follower_id = $8::bigint
WHERE rc.gr < 3
`

type ListCommentsParams struct {
	Limit         int32 `json:"limit"`
	Offset        int32 `json:"offset"`
	PostID        int64 `json:"post_id"`
	StarCountAsc  bool  `json:"star_count_asc"`
	StarCountDesc bool  `json:"star_count_desc"`
	CreateAtAsc   bool  `json:"create_at_asc"`
	CreateAtDesc  bool  `json:"create_at_desc"`
	SelfID        int64 `json:"self_id"`
}

type ListCommentsRow struct {
	ID              int64          `json:"id"`
	ParentID        sql.NullInt64  `json:"parent_id"`
	Content         string         `json:"content"`
	StarCount       int64          `json:"star_count"`
	CreateAt        time.Time      `json:"create_at"`
	Total           int64          `json:"total"`
	CommentCount    int64          `json:"comment_count"`
	ReplyCount      int64          `json:"reply_count"`
	UserID          int64          `json:"user_id"`
	Username        string         `json:"username"`
	Avatar          string         `json:"avatar"`
	Intro           string         `json:"intro"`
	Followed        sql.NullInt64  `json:"followed"`
	FollowerCount   int64          `json:"follower_count"`
	FollowingCount  int64          `json:"following_count"`
	RUserID         sql.NullInt64  `json:"r_user_id"`
	RUsername       sql.NullString `json:"r_username"`
	RAvatar         sql.NullString `json:"r_avatar"`
	RIntro          sql.NullString `json:"r_intro"`
	RFollowed       sql.NullInt64  `json:"r_followed"`
	RFollowerCount  int64          `json:"r_follower_count"`
	RFollowingCount int64          `json:"r_following_count"`
}

func (q *Queries) ListComments(ctx context.Context, arg ListCommentsParams) ([]ListCommentsRow, error) {
	rows, err := q.db.QueryContext(ctx, listComments,
		arg.Limit,
		arg.Offset,
		arg.PostID,
		arg.StarCountAsc,
		arg.StarCountDesc,
		arg.CreateAtAsc,
		arg.CreateAtDesc,
		arg.SelfID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListCommentsRow{}
	for rows.Next() {
		var i ListCommentsRow
		if err := rows.Scan(
			&i.ID,
			&i.ParentID,
			&i.Content,
			&i.StarCount,
			&i.CreateAt,
			&i.Total,
			&i.CommentCount,
			&i.ReplyCount,
			&i.UserID,
			&i.Username,
			&i.Avatar,
			&i.Intro,
			&i.Followed,
			&i.FollowerCount,
			&i.FollowingCount,
			&i.RUserID,
			&i.RUsername,
			&i.RAvatar,
			&i.RIntro,
			&i.RFollowed,
			&i.RFollowerCount,
			&i.RFollowingCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listReplies = `-- name: ListReplies :many
WITH Data_CTE AS (
  SELECT cm.id, cm.post_id, cm.user_id, cm.parent_id, cm.reply_user_id, cm.content, cm.create_at, count(cms.user_id) star_count
  FROM comments cm
  LEFT JOIN comment_stars cms ON cms.comment_id = cm.id
  WHERE cm.parent_id = $8::bigint
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
  ON fu.user_id = dc.user_id AND fu.follower_id = $3::bigint
LEFT JOIN follows fr
  ON fr.user_id = dc.reply_user_id AND fr.follower_id = $3::bigint
ORDER BY
  CASE WHEN $4::bool THEN star_count END ASC,
  CASE WHEN $5::bool THEN star_count END DESC,
  CASE WHEN $6::bool THEN dc.create_at END ASC,
  CASE WHEN $7::bool THEN dc.create_at END DESC
LIMIT $1
OFFSET $2
`

type ListRepliesParams struct {
	Limit         int32 `json:"limit"`
	Offset        int32 `json:"offset"`
	SelfID        int64 `json:"self_id"`
	StarCountAsc  bool  `json:"star_count_asc"`
	StarCountDesc bool  `json:"star_count_desc"`
	CreateAtAsc   bool  `json:"create_at_asc"`
	CreateAtDesc  bool  `json:"create_at_desc"`
	ParentID      int64 `json:"parent_id"`
}

type ListRepliesRow struct {
	ID              int64          `json:"id"`
	Content         string         `json:"content"`
	StarCount       int64          `json:"star_count"`
	CreateAt        time.Time      `json:"create_at"`
	Total           int64          `json:"total"`
	UserID          int64          `json:"user_id"`
	Username        string         `json:"username"`
	Avatar          string         `json:"avatar"`
	Intro           string         `json:"intro"`
	Followed        sql.NullInt64  `json:"followed"`
	FollowerCount   int64          `json:"follower_count"`
	FollowingCount  int64          `json:"following_count"`
	RUserID         sql.NullInt64  `json:"r_user_id"`
	RUsername       sql.NullString `json:"r_username"`
	RAvatar         sql.NullString `json:"r_avatar"`
	RIntro          sql.NullString `json:"r_intro"`
	RFollowed       sql.NullInt64  `json:"r_followed"`
	RFollowerCount  int64          `json:"r_follower_count"`
	RFollowingCount int64          `json:"r_following_count"`
}

func (q *Queries) ListReplies(ctx context.Context, arg ListRepliesParams) ([]ListRepliesRow, error) {
	rows, err := q.db.QueryContext(ctx, listReplies,
		arg.Limit,
		arg.Offset,
		arg.SelfID,
		arg.StarCountAsc,
		arg.StarCountDesc,
		arg.CreateAtAsc,
		arg.CreateAtDesc,
		arg.ParentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListRepliesRow{}
	for rows.Next() {
		var i ListRepliesRow
		if err := rows.Scan(
			&i.ID,
			&i.Content,
			&i.StarCount,
			&i.CreateAt,
			&i.Total,
			&i.UserID,
			&i.Username,
			&i.Avatar,
			&i.Intro,
			&i.Followed,
			&i.FollowerCount,
			&i.FollowingCount,
			&i.RUserID,
			&i.RUsername,
			&i.RAvatar,
			&i.RIntro,
			&i.RFollowed,
			&i.RFollowerCount,
			&i.RFollowingCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}