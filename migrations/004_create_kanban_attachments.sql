-- Kanban Cards System DDL

-- Card Attachments Table
--------------------------
-- This table stores attachments related to Kanban cards.
-- It's best practice to store file metadata here and the files themselves in a separate system,
-- but for this example, we'll store the bytes directly.

-- +goose Up
CREATE TABLE card_attachments (
    attachment_id SERIAL PRIMARY KEY,
    card_id INT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_type VARCHAR(50),
    file_size_bytes BIGINT,
    attachment_data BYTEA,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    uploaded_by VARCHAR(100) NOT NULL,

    FOREIGN KEY (card_id) REFERENCES kanban_cards(card_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE card_attachments;
