-- +goose Up
CREATE TABLE feed_follows(
    id BIGSERIAL PRIMARY KEY,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE feed_follows ADD CONSTRAINT unique_feed_follow UNIQUE(feed_id, user_id);

-- +goose Down
DROP TABLE feed_follows;
DROP INDEX IF EXISTS unique_feed_follow;
