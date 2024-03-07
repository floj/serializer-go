-- name: ListStoriesBeginningAt :many
SELECT * FROM stories WHERE id >= $1 ORDER BY id desc LIMIT 1000;

-- name: UpdateStory :one
UPDATE stories SET 
  title = $1, 
  url = $2, 
  score = $3, 
  num_comments = $4,
  updated_at = $5,
  type = $6
WHERE id = $7
RETURNING *;

-- name: FindByScraperAndRef :many
SELECT * FROM stories WHERE scraper=$1 AND ref_id=$2;

-- name: FindRecentForUpdate :many
SELECT * FROM stories WHERE scraper=$1 AND updated_at < $2 and created_at > $3;

-- name: CreateStory :one
INSERT INTO stories (
  ref_id, url, by, created_at, updated_at, scraped_at, title, type, score, num_comments, scraper
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: CreateStoryHistory :one
INSERT INTO story_history (
  story_id, field, old_val, new_val, created_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;
