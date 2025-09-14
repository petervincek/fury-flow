package comment

import (
	"time"

	"github.com/petervincek/fury-flow/internal/db"
)

// Comment represents a comment made on a card within the system.
// It contains the unique identifier for the comment, the associated card's ID,
// the text of the comment, the author of the comment, and the timestamp when
// the comment was created.
type Comment struct {
	CommentId int `json:"commentId"`
	// should this be at this level just id or the whole card document
	CardId      int       `json:"cardId"`
	CommentText string    `json:"commentText" validate:"required,min=3"`
	CommentBy   string    `json:"commentBy" validate:"required,min=3"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Builder provides a convenient way to construct and configure Comment instances.
type Builder struct {
	comment Comment
}

// NewBuilder creates and returns a new Builder instance with the Comment's CreatedAt field set to the current time.
func NewBuilder() *Builder {
	return &Builder{
		comment: Comment{
			CreatedAt: time.Now(),
		},
	}
}

// SetCommentId sets the CommentId field of the comment to the specified id.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCommentId(id int) *Builder {
	b.comment.CommentId = id
	return b
}

// SetCardId sets the CardId field of the comment to the provided cardId value.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCardId(cardId int) *Builder {
	b.comment.CardId = cardId
	return b
}

// SetCommentText sets the text of the comment in the Builder.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetCommentText(text string) *Builder {
	b.comment.CommentText = text
	return b
}

// SetCommentBy sets the author of the comment.
// It assigns the provided 'by' string to the CommentBy field of the comment
// and returns the Builder for method chaining.
func (b *Builder) SetCommentBy(by string) *Builder {
	b.comment.CommentBy = by
	return b
}

// SetCreatedAt sets the CreatedAt field of the comment to the specified time.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCreatedAt(t time.Time) *Builder {
	b.comment.CreatedAt = t
	return b
}

// Build validates the Comment being constructed by the Builder.
// If validation passes, it returns the built Comment and a nil error.
// If validation fails, it returns an empty Comment and the validation error.
func (b *Builder) Build() (Comment, error) {
	err := validate.Struct(b.comment)
	if err != nil {
		return Comment{}, err
	}
	return b.comment, nil
}

// TransformComment converts a db.CardComment object to a Comment object.
// It maps the fields from the database model to the API model, including
// converting ID fields to int and extracting the time from CreatedAt.
func TransformComment(cc db.CardComment) Comment {
	return Comment{
		CommentId:   int(cc.CommentID),
		CardId:      int(cc.CardID),
		CommentText: cc.CommentText,
		CommentBy:   cc.CommentBy,
		CreatedAt:   cc.CreatedAt.Time,
	}
}

// TransformComments takes a slice of db.CardComment and returns a slice of Comment,
// transforming each db.CardComment into a Comment using the TransformComment function.
func TransformComments(ccs []db.CardComment) []Comment {
	comments := make([]Comment, len(ccs))
	for idx, cc := range ccs {
		comments[idx] = TransformComment(cc)
	}
	return comments
}
