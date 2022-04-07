CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "username" varchar NOT NULL,
  "password" varchar NOT NULL,
  "nickname" varchar NOT NULL,
  "avatar_src" varchar NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "articles" (
  "id" serial PRIMARY KEY,
  "author" integer,
  "title" varchar NOT NULL,
  "summary" varchar NOT NULL,
  "content" text NOT NULL,
  "article_status" varchar NOT NULL,
  "publish_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "articles" ADD FOREIGN KEY ("author") REFERENCES "users" ("id");
