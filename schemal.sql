-- Remember to run this command each time you connect to your database!
PRAGMA foreign_keys = ON;

--------------------------------------------------

CREATE TABLE IF NOT EXISTS "users" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "username" VARCHAR NOT NULL UNIQUE,
    "email" VARCHAR NOT NULL UNIQUE,
    "password" VARCHAR NOT NULL,
    "avatar" VARCHAR,
    "role" INTEGER NOT NULL DEFAULT 0,
    "session_id" VARCHAR UNIQUE,
    "session_created_at" TIMESTAMP
);

--------------------------------------------------

CREATE TABLE IF NOT EXISTS "categories" (
    "id" INTEGER NOT NULL PRIMARY KEY,
    "name" VARCHAR NOT NULL UNIQUE
);

--------------------------------------------------

CREATE TABLE IF NOT EXISTS "posts" (
    "id" INTEGER NOT NULL PRIMARY KEY,
    "title" VARCHAR NOT NULL,
    "content" TEXT NOT NULL,
    "author_id" TEXT NOT NULL,
    "category_id" INTEGER,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified" TIMESTAMP,
    FOREIGN KEY ("author_id") REFERENCES "users"("id") ON DELETE CASCADE,
    FOREIGN KEY ("category_id") REFERENCES "categories"("id") ON DELETE SET NULL
);

--------------------------------------------------

CREATE TABLE IF NOT EXISTS "comments" (
    "id" INTEGER NOT NULL PRIMARY KEY,
    "content" TEXT NOT NULL,
    "author_id" TEXT NOT NULL,
    "post_id" INTEGER NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_modified" TIMESTAMP,
    FOREIGN KEY ("author_id") REFERENCES "users"("id") ON DELETE CASCADE,
    FOREIGN KEY ("post_id") REFERENCES "posts"("id") ON DELETE CASCADE
);



CREATE TABLE IF NOT EXISTS "reactions" (
    "id" INTEGER NOT NULL PRIMARY KEY,
    "type" INTEGER NOT NULL,
    "user_id" TEXT NOT NULL,
    "post_id" INTEGER,
    "comment_id" INTEGER,
    FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE,
    FOREIGN KEY ("post_id") REFERENCES "posts"("id") ON DELETE CASCADE,
    FOREIGN KEY ("comment_id") REFERENCES "comments"("id") ON DELETE CASCADE,
    CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL) OR
        (post_id IS NULL AND comment_id IS NOT NULL)
    )
);


CREATE UNIQUE INDEX IF NOT EXISTS "idx_unique_post_reaction"
ON "reactions" ("user_id", "post_id")
WHERE "post_id" IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS "idx_unique_comment_reaction"
ON "reactions" ("user_id", "comment_id")
WHERE "comment_id" IS NOT NULL;