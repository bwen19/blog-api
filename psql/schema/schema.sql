-- SQL dump generated using DBML (dbml-lang.org)
-- Database: PostgreSQL
-- Generated at: 2022-09-18T07:15:58.768Z

CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "avatar" varchar NOT NULL,
  "intro" varchar NOT NULL DEFAULT '',
  "role" varchar NOT NULL DEFAULT 'user',
  "deleted" boolean NOT NULL DEFAULT false,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "notifications" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "kind" varchar NOT NULL,
  "title" varchar NOT NULL,
  "content" text NOT NULL,
  "unread" boolean NOT NULL DEFAULT true,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "follows" (
  "user_id" bigint,
  "follower_id" bigint,
  "create_at" timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY ("user_id", "follower_id")
);

CREATE TABLE "posts" (
  "id" bigserial PRIMARY KEY,
  "author_id" bigint NOT NULL,
  "title" varchar NOT NULL,
  "cover_image" varchar NOT NULL,
  "status" varchar NOT NULL DEFAULT 'draft',
  "featured" boolean NOT NULL DEFAULT false,
  "view_count" bigint NOT NULL DEFAULT 0,
  "update_at" timestamptz NOT NULL DEFAULT (now()),
  "publish_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "post_contents" (
  "id" bigint PRIMARY KEY,
  "content" text NOT NULL
);

CREATE TABLE "post_stars" (
  "post_id" bigint,
  "user_id" bigint,
  "create_at" timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY ("post_id", "user_id")
);

CREATE TABLE "categories" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL
);

CREATE TABLE "post_categories" (
  "post_id" bigint,
  "category_id" bigint,
  PRIMARY KEY ("post_id", "category_id")
);

CREATE TABLE "tags" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL
);

CREATE TABLE "post_tags" (
  "post_id" bigint,
  "tag_id" bigint,
  PRIMARY KEY ("post_id", "tag_id")
);

CREATE TABLE "comments" (
  "id" bigserial PRIMARY KEY,
  "post_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "parent_id" bigint,
  "reply_user_id" bigint,
  "content" text NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "comment_stars" (
  "comment_id" bigint,
  "user_id" bigint,
  PRIMARY KEY ("comment_id", "user_id")
);

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "notifications" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "follows" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "follows" ADD FOREIGN KEY ("follower_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "posts" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE "post_contents" ADD FOREIGN KEY ("id") REFERENCES "posts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "post_stars" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "post_stars" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "post_categories" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "post_categories" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "post_tags" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "post_tags" ADD FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "comments" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "comments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "comments" ADD FOREIGN KEY ("parent_id") REFERENCES "comments" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "comments" ADD FOREIGN KEY ("reply_user_id") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION;

ALTER TABLE "comment_stars" ADD FOREIGN KEY ("comment_id") REFERENCES "comments" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "comment_stars" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
