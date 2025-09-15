-- name: GetKanbanCardComments :many
SELECT * FROM card_comments WHERE card_id = $1;

-- name: GetKanbanCommentById :one
SELECT * FROM card_comments WHERE comment_id = $1;

-- name: CreateKanbanComment :one
INSERT INTO card_comments (card_id, comment_text, comment_by, created_at) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteAllKanbanComments :exec
DELETE FROM card_comments;

-- name: DeleteKanbanComment :exec
DELETE FROM card_comments WHERE comment_id = $1;

-- name: DeleteKanbanCommentsForCard :exec
DELETE FROM card_comments WHERE card_id = $1;