-- name: GetKanbanCards :many
SELECT * from kanban_cards;

-- name: GetKanbanCardsForBoard :many
SELECT * from kanban_cards WHERE board_id = $1;

-- name: GetKanbanCardById :one
SELECT * from kanban_cards WHERE card_id = $1;

-- name: CreateKanbanCard :one
INSERT INTO kanban_cards (board_id, title, description, status, assignee, story_points, acceptance_criteria, time_spent_hours,
 created_by, priority, type, is_blocked, blocked_reason) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING *;

-- name: UpdateKanbanCard :execresult
UPDATE kanban_cards SET title = $2, description = $3, status = $4, assignee = $5, story_points = $6, acceptance_criteria = $7, time_spent_hours = $8,
 priority = $9, type = $10, is_blocked = $11, blocked_reason = $12, updated_at = $13, updated_by = $14 WHERE card_id = $1;

-- name: DeleteKanbanCard :exec
DELETE FROM kanban_cards WHERE card_id = $1;

-- name: DeleteAllKanbanCards :exec
DELETE FROM kanban_cards;

-- name: KanbanCardExists :one
SELECT EXISTS(SELECT 1 FROM kanban_cards WHERE card_id = $1);

-- name: CreateCardDependency :execresult
-- Purpose: Establish a dependency relationship between two cards.
-- Parameters: $1 = card_id, $2 = depends_on_card_id
INSERT INTO card_dependencies (card_id, depends_on_card_id) VALUES ($1, $2);

-- name: CreateCardDependencies :execresult
-- Purpose: Establish multiple dependency relationships for one card in a single statement.
-- Parameters: $1 = card_id (int), $2 = depends_on_card_id (int[])
INSERT INTO card_dependencies (card_id, depends_on_card_id)
SELECT $1 as card_id, unnest($2::int[]) as dependency_card_id;

-- name: GetCardDependencyIds :many
-- Purpose: Retrieve all cards that the specified card depends on.
-- Parameter: $1 = card_id
SELECT depends_on_card_id FROM card_dependencies WHERE card_id = $1;

-- name: GetCardDependencies :many
-- Purpose: Retrieve detailed information about all dependency cards for a given card.
-- Parameter: $1 = card_id
SELECT kc.* 
FROM card_dependencies cd JOIN kanban_cards kc ON cd.depends_on_card_id = kc.card_id 
WHERE cd.card_id = $1;

-- name: DeleteCardDependency :exec
-- Purpose: Remove a specific dependency relationship between two cards.
-- Parameters: $1 = card_id, $2 = depends_on_card_id
DELETE FROM card_dependencies WHERE card_id = $1 AND depends_on_card_id = $2;

-- name: DeleteCardDependencies :exec
-- Purpose: Remove all dependency relationships for provided card.
-- Parameters: $1 = card_id
DELETE FROM card_dependencies WHERE card_id = $1;