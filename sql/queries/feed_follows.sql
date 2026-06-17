-- name: CreateFeedFollow :one
WITH feed_follow AS (
    INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, $2)
    RETURNING *
)
SELECT feed_follow.*, feeds.name as feed_name, users.name AS username
FROM feed_follow
JOIN feeds ON feeds.id = feed_follow.feed_id
JOIN users ON users.id = feed_follow.user_id;


-- name: GetFeedFollowsByUser :many
SELECT feeds.*, users.name AS username
FROM feed_follows
JOIN feeds ON feeds.id = feed_follows.feed_id
JOIN users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1;
