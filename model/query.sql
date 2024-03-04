-- name: ListStoriesBeginningAt :many
SELECT * FROM stories WHERE id >= $1 ORDER BY id desc;

-- name: UpdateStory :one
UPDATE stories SET 
  title = $1, 
  url = $2, 
  score = $3, 
  num_comments = $4,
  type = $5
WHERE id = $6
RETURNING *;

-- name: FindByScraperAndRef :many
SELECT * FROM stories WHERE scraper=$1 AND ref_id=$2;


-- name: CreateStory :one
INSERT INTO stories (
  ref_id, url, by, created_at, scraped_at, title, type, score, num_comments, scraper
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;
