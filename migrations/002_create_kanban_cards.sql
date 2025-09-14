-- Kanban Cards System DDL

-- Kanban Cards Table
--------------------
-- This table stores the core information for each Kanban card.

-- +goose Up
CREATE TABLE kanban_cards (
    card_id SERIAL PRIMARY KEY,
    
    -- Link to the kanban board
    board_id INT NOT NULL,

    title VARCHAR(255) NOT NULL,
    description TEXT, -- This field is intended to store content formatted with Markdown.
    status VARCHAR(50) NOT NULL CHECK (status IN ('To Do', 'In Progress', 'In Review', 'Done')),

    -- Additional attributes for more complex systems
    assignee VARCHAR(100),
    story_points SMALLINT CHECK (story_points >= 0),
    acceptance_criteria TEXT, -- This field is intended to store acceptance criteria, often formatted as a list using Markdown.
    time_spent_hours NUMERIC(10, 2) DEFAULT 0.0,

    -- Auditable fields for tracking changes
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    updated_by VARCHAR(100),

    -- Additional details
    priority VARCHAR(20) CHECK (priority IN ('High', 'Medium', 'Low')),
    type VARCHAR(50) CHECK (type IN ('Bug', 'Feature', 'Maintenance')),
    is_blocked BOOLEAN DEFAULT FALSE,
    blocked_reason TEXT,
    
    -- Foreign key constraint linking a card to a board
    FOREIGN KEY (board_id) REFERENCES kanban_boards(board_id) ON DELETE CASCADE
);

-- Note on Calculated Metrics:
-- Lead time and cycle time are calculated metrics and should not be stored in this table.
-- They can be derived by querying the `created_at`, `started_at`, and `completed_at` columns.

-- Indexes for performance
-- +goose StatementBegin
CREATE INDEX idx_kanban_cards_status ON kanban_cards (status);
CREATE INDEX idx_kanban_cards_assignee ON kanban_cards (assignee);
CREATE INDEX idx_kanban_cards_board_id ON kanban_cards (board_id);
-- +goose StatementEnd

-- Card Dependencies Table
---------------------------
-- This table establishes a many-to-many relationship to manage card dependencies.
-- A card can have multiple other cards as prerequisites, and can be a dependency for other cards.

CREATE TABLE card_dependencies (
    -- The card that is dependent on another card.
    card_id INT NOT NULL,
    
    -- The card that is a prerequisite for the other card.
    depends_on_card_id INT NOT NULL,

    PRIMARY KEY (card_id, depends_on_card_id),
    
    FOREIGN KEY (card_id) REFERENCES kanban_cards(card_id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_card_id) REFERENCES kanban_cards(card_id) ON DELETE CASCADE
);


-- +goose Down

-- +goose StatementBegin
DROP INDEX idx_kanban_cards_status;
DROP INDEX idx_kanban_cards_assignee;
DROP INDEX idx_kanban_cards_board_id;
-- +goose StatementEnd

DROP TABLE card_dependencies;
DROP TABLE kanban_cards;