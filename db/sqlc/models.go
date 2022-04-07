// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package db

import (
	"database/sql"
	"time"
)

type Article struct {
	ID            int32         `json:"id"`
	Author        sql.NullInt32 `json:"author"`
	Title         string        `json:"title"`
	Summary       string        `json:"summary"`
	Content       string        `json:"content"`
	ArticleStatus string        `json:"article_status"`
	PublishAt     time.Time     `json:"publish_at"`
}

type User struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Nickname  string    `json:"nickname"`
	AvatarSrc string    `json:"avatar_src"`
	CreateAt  time.Time `json:"create_at"`
}
