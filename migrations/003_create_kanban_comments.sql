-- Kanban Cards System DDL

-- Card Comments Table
-----------------------
-- This table stores comments related to specific Kanban cards.

-- +goose Up
CREATE TABLE card_comments (
    comment_id SERIAL PRIMARY KEY,
    card_id INT NOT NULL,
    comment_text TEXT NOT NULL, -- This field is intended to store comments formatted with Markdown.
    comment_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (card_id) REFERENCES kanban_cards(card_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE card_comments;