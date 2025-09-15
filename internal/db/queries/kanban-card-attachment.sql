-- name: GetKanbanCardAttachments :many
SELECT * FROM card_attachments WHERE card_id = $1;

-- name: GetKanbanAttachmentById :one
SELECT * FROM card_attachments WHERE attachment_id = $1;

-- name: CreateKanbanAttachment :one
INSERT INTO card_attachments (card_id, filename, file_type, file_size_bytes, attachment_data, uploaded_by, uploaded_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: DeleteAllKanbanAttachments :exec
DELETE FROM card_attachments;

-- name: DeleteKanbanAttachment :exec
DELETE FROM card_attachments WHERE attachment_id = $1;

-- name: DeleteKanbanAttachmentsForCard :exec
DELETE FROM card_attachments WHERE card_id = $1;