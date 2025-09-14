package card

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/petervincek/fury-flow/internal/db"
	"go.uber.org/zap"
)

// Status represents the current state of a card in the system.
type Status string

// Priority represents the priority level of a card as a string value.
type Priority string

// Type represents the type of a card as a string value.
type Type string

const (
	StatusTodo       Status = "To Do"
	StatusInProgress Status = "In Progress"
	StatusInReview   Status = "In Review"
	StatusDone       Status = "Done"

	PriorityHigh   Priority = "High"
	PriorityMedium Priority = "Medium"
	PriorityLow    Priority = "Low"

	TypeBug         Type = "Bug"
	TypeFeature     Type = "Feature"
	TypeMaintenance Type = "Maintenance"
)

// Card represents a task or item within a board, containing details such as title, description, status, assignee, story points, acceptance criteria, time spent, priority, type, and blocking information.
// It also embeds AuditableFields for tracking creation and modification metadata.
type Card struct {
	CardId int `json:"cardId"`
	// should this be at this level just id or the whole board document
	BoardId            int      `json:"boardId"`
	Title              string   `json:"title" validate:"required,min=3"`
	Description        string   `json:"description" validate:"required,min=3"`
	Status             Status   `json:"status" validate:"required,min=3"`
	Assignee           string   `json:"assignee"`
	StoryPoints        int      `json:"storyPoints" validate:"required,gt=0"`
	AcceptanceCriteria string   `json:"acceptanceCriteria" validate:"required,min=3"`
	TimeSpentHours     float64  `json:"timeSpentHours"`
	Priority           Priority `json:"priority"`
	Type               Type     `json:"type" validate:"required,min=3"`
	IsBlocked          bool     `json:"isBlocked"`
	BlockedReason      string   `json:"blockedReason"`
	common.AuditableFields
}

// Builder provides a convenient way to construct and configure Card instances.
type Builder struct {
	card Card
}

// NewBuilder creates and returns a new instance of Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// SetCardId sets the CardId field of the card to the provided cardId value.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCardId(cardId int) *Builder {
	b.card.CardId = cardId
	return b
}

// SetBoardId sets the BoardId field of the card to the provided boardId value.
// It returns the Builder instance to allow for method chaining.
func (b *Builder) SetBoardId(boardId int) *Builder {
	b.card.BoardId = boardId
	return b
}

// SetTitle sets the title of the card and returns the Builder for method chaining.
func (b *Builder) SetTitle(title string) *Builder {
	b.card.Title = title
	return b
}

// SetDescription sets the description of the card in the Builder.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetDescription(description string) *Builder {
	b.card.Description = description
	return b
}

// SetStatus sets the status of the card in the Builder and returns the Builder for chaining.
func (b *Builder) SetStatus(status Status) *Builder {
	b.card.Status = status
	return b
}

// SetAssignee sets the assignee of the card to the specified value.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetAssignee(assignee string) *Builder {
	b.card.Assignee = assignee
	return b
}

// SetStoryPoints sets the story points for the card and returns the Builder for chaining.
func (b *Builder) SetStoryPoints(storyPoints int) *Builder {
	b.card.StoryPoints = storyPoints
	return b
}

// SetAcceptanceCriteria sets the acceptance criteria for the card being built.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetAcceptanceCriteria(acceptanceCriteria string) *Builder {
	b.card.AcceptanceCriteria = acceptanceCriteria
	return b
}

// SetTimeSpentHours sets the time spent in hours on the card.
// It updates the TimeSpentHours field of the card and returns the Builder for chaining.
func (b *Builder) SetTimeSpentHours(timeSpentHours float64) *Builder {
	b.card.TimeSpentHours = timeSpentHours
	return b
}

// SetPriority sets the priority of the card in the Builder.
// It takes a Priority value and returns the Builder for method chaining.
func (b *Builder) SetPriority(priority Priority) *Builder {
	b.card.Priority = priority
	return b
}

// SetType sets the type of the card in the Builder and returns the Builder for chaining.
func (b *Builder) SetType(cardType Type) *Builder {
	b.card.Type = cardType
	return b
}

// SetIsBlocked sets the IsBlocked field of the card to the specified value.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetIsBlocked(isBlocked bool) *Builder {
	b.card.IsBlocked = isBlocked
	return b
}

// SetBlockedReason sets the reason why the card is blocked.
// It updates the BlockedReason field of the card and returns the Builder for chaining.
func (b *Builder) SetBlockedReason(blockedReason string) *Builder {
	b.card.BlockedReason = blockedReason
	return b
}

// SetCreatedAt sets the CreatedAt field of the card to the specified time.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCreatedAt(createdAt time.Time) *Builder {
	b.card.CreatedAt = createdAt
	return b
}

// SetCreatedBy sets the CreatedBy field of the card to the specified value.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCreatedBy(createdBy string) *Builder {
	b.card.CreatedBy = createdBy
	return b
}

// SetUpdatedAt sets the UpdatedAt field of the card to the specified time.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetUpdatedAt(updatedAt time.Time) *Builder {
	b.card.UpdatedAt = updatedAt
	return b
}

// SetUpdatedBy sets the UpdatedBy field of the card to the specified value.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetUpdatedBy(updatedBy string) *Builder {
	b.card.UpdatedBy = updatedBy
	return b
}

// Build validates the Card being constructed by the Builder.
// If validation passes, it returns the built Card and a nil error.
// If validation fails, it returns an empty Card and the validation error.
func (b *Builder) Build() (Card, error) {
	err := validate.Struct(b.card)
	if err != nil {
		return Card{}, err
	}
	return b.card, nil
}

// TransformCard converts a db.KanbanCard object to a Card object.
// It maps all relevant fields from the KanbanCard database model to the Card API model,
// including handling nullable types and custom type conversions.
// This function is typically used to prepare card data for API responses.
func TransformCard(kc db.KanbanCard) Card {
	return Card{
		CardId:             int(kc.CardID),
		BoardId:            int(kc.BoardID),
		Title:              kc.Title,
		Description:        kc.Description.String,
		Status:             Status(kc.Status),
		Assignee:           kc.Assignee.String,
		StoryPoints:        int(kc.StoryPoints.Int16),
		AcceptanceCriteria: kc.AcceptanceCriteria.String,
		TimeSpentHours:     getFloat(kc.TimeSpentHours),
		Priority:           Priority(kc.Priority.String),
		Type:               Type(kc.Type.String),
		IsBlocked:          kc.IsBlocked.Bool,
		BlockedReason:      kc.BlockedReason.String,
		AuditableFields: common.AuditableFields{
			CreatedAt: kc.CreatedAt.Time,
			CreatedBy: kc.CreatedBy,
			UpdatedAt: kc.UpdatedAt.Time,
			UpdatedBy: kc.UpdatedBy.String,
		},
	}
}

// TransformCards converts a slice of db.KanbanCard objects into a slice of Card objects
// by applying the TransformCard function to each element.
// It returns the resulting slice of Card.
func TransformCards(kcs []db.KanbanCard) []Card {
	cards := make([]Card, len(kcs))
	for idx, kc := range kcs {
		cards[idx] = TransformCard(kc)
	}
	return cards
}

// getFloat converts a pgtype.Numeric value to a float64.
// If the conversion fails, it logs the error and returns 0.0.
func getFloat(num pgtype.Numeric) float64 {
	value, err := num.Float64Value()
	if err != nil {
		logger.Error("error while trying to get float number from pgtype.Numeric", zap.Error(err))
		return 0.0
	}
	return value.Float64
}
