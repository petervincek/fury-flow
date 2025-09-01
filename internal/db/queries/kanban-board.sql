-- name: GetKanbanBoards :many
SELECT * from kanban_boards;

-- name: CreateKanbanBoard :one
INSERT INTO kanban_boards (board_name, description, created_by) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateKanbanBoard :exec
UPDATE kanban_boards SET board_name = $2, description = $3, updated_at = $4, updated_by = $5 WHERE board_id = $1;

-- name: DeleteKanbanBoard :exec
DELETE FROM kanban_boards WHERE board_id = $1;