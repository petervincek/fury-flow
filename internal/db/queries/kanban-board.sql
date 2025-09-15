-- name: GetKanbanBoards :many
SELECT * from kanban_boards;

-- name: GetKanbanBoardById :one
SELECT * from kanban_boards WHERE board_id = $1;

-- name: CreateKanbanBoard :one
INSERT INTO kanban_boards (board_name, description, created_by) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateKanbanBoard :execresult
UPDATE kanban_boards SET board_name = $2, description = $3, updated_at = $4, updated_by = $5 WHERE board_id = $1;

-- name: DeleteKanbanBoard :exec
DELETE FROM kanban_boards WHERE board_id = $1;

-- name: DeleteAll :exec
DELETE FROM kanban_boards;

-- name: KanbanBoardExists :one
SELECT EXISTS(SELECT 1 FROM kanban_boards WHERE board_id = $1) AS exists;