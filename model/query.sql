-- name: ListStoriesBeginningAt :many
SELECT * FROM stories WHERE id >= ? ORDER BY id desc LIMIT 1000;

-- name: UpdateStory :one
UPDATE stories SET 
  title = ?, 
  url = ?, 
  score = ?, 
  num_comments = ?,
  type = ?,
  last_seen_fp = ?
WHERE id = ?
RETURNING *;

-- name: MarkStoryDeleted :one
UPDATE stories SET deleted = true WHERE id = ? RETURNING *;

-- name: FindByScraperAndRef :many
SELECT * FROM stories WHERE scraper=? AND ref_id=?;

-- name: FindRecentForUpdate :many
SELECT * FROM stories WHERE scraper=? AND created_at > ? AND updated_at < ?  AND deleted=false;

-- name: CreateStory :one
INSERT INTO stories (
  ref_id, url, by, published_at, title, type, score, num_comments, scraper
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: RecordStoryChange :one
INSERT INTO story_history(story_id, field, old_val, new_val) VALUES(?, ?, ?, ?) RETURNING *;
