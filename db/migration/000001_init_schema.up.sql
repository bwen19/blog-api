CREATE TABLE "users" (
  "username" varchar NOT NULL PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "avatar_src" varchar NOT NULL,
  "role" varchar NOT NULL DEFAULT 'user',
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "articles" (
  "id" bigserial PRIMARY KEY,
  "author" varchar NOT NULL,
  "category" varchar NOT NULL,
  "title" varchar NOT NULL,
  "summary" varchar NOT NULL,
  "content" text NOT NULL,
  "status" varchar NOT NULL DEFAULT 'draft',
  "view_count" bigint NOT NULL DEFAULT 0,
  "update_at" timestamptz NOT NULL DEFAULT (now()),
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "categories" (
  "name" varchar PRIMARY KEY
);

CREATE TABLE "tags" (
  "name" varchar PRIMARY KEY,
  "count" bigint NOT NULL DEFAULT 0
);

CREATE TABLE "article_tags" (
  "article_id" bigint,
  "tag" varchar,
  PRIMARY KEY ("article_id", "tag")
);

CREATE TABLE "comments" (
  "id" bigserial PRIMARY KEY,
  "parent_id" bigint,
  "article_id" bigint NOT NULL,
  "commenter" varchar NOT NULL,
  "content" varchar NOT NULL,
  "comment_at" timestamptz NOT NULL DEFAULT (now())
);


ALTER TABLE "sessions" ADD FOREIGN KEY ("username")
REFERENCES "users" ("username")
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "articles" ADD FOREIGN KEY ("author")
REFERENCES "users" ("username")
ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE "articles" ADD FOREIGN KEY ("category")
REFERENCES "categories" ("name")
ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE "article_tags" ADD FOREIGN KEY ("article_id")
REFERENCES "articles" ("id")
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "article_tags" ADD FOREIGN KEY ("tag")
REFERENCES "tags" ("name")
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "comments" ADD FOREIGN KEY ("parent_id")
REFERENCES "comments" ("id")
ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "comments" ADD FOREIGN KEY ("commenter")
REFERENCES "users" ("username")
ON DELETE CASCADE ON UPDATE CASCADE;;

ALTER TABLE "comments" ADD FOREIGN KEY ("article_id")
REFERENCES "articles" ("id")
ON DELETE CASCADE ON UPDATE CASCADE;;
