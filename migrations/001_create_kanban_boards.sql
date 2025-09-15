-- Kanban Cards System DDL

-- Kanban Boards Table
---------------------
-- This table represents different Kanban boards for various projects or teams.

-- +goose Up
CREATE TABLE kanban_boards (
    board_id SERIAL PRIMARY KEY,
    board_name VARCHAR(255) NOT NULL,
    description TEXT, -- This field is intended to store content formatted with Markdown.
    
    -- Auditable fields for the board
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by VARCHAR(100)
);

-- +goose Down
DROP TABLE kanban_boards;