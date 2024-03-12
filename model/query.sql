-- name: ListStoriesBeginningAt :many
SELECT * FROM stories WHERE id >= $1 ORDER BY id desc LIMIT 1000;

-- name: UpdateStory :one
UPDATE stories SET 
  title = $1, 
  url = $2, 
  score = $3, 
  num_comments = $4,
  type = $5,
  last_seen_fp = $6
WHERE id = $7
RETURNING *;

-- name: MarkStoryDeleted :one
UPDATE stories SET deleted = true WHERE id = $1 RETURNING *;

-- name: FindByScraperAndRef :many
SELECT * FROM stories WHERE scraper=$1 AND ref_id=$2;

-- name: FindRecentForUpdate :many
SELECT * FROM stories WHERE scraper=$1 AND created_at > $2 AND updated_at < $3  AND deleted=false;

-- name: CreateStory :one
INSERT INTO stories (
  ref_id, url, by, published_at, title, type, score, num_comments, scraper
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;