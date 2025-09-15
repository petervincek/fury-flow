package attachment

import (
	"time"

	"github.com/petervincek/fury-flow/internal/db"
)

// Attachment represents a file attached to a card, including metadata and binary data.
// It contains information such as the attachment's unique ID, associated card ID,
// filename, file type, file size, binary data, uploader's identity, and upload timestamp.
type Attachment struct {
	AttachmentId int `json:"attachmentId"`
	// should this be at this level just id or the whole card document
	CardId         int       `json:"cardId"`
	Filename       string    `json:"filename" validate:"required,min=3"`
	FileType       string    `json:"fileType" validate:"required,min=3"`
	FileSizeBytes  int       `json:"fileSizeBytes" validate:"required,gt=0"`
	AttachmentData []byte    `json:"attachmentData" validate:"required"`
	UploadedBy     string    `json:"uploadedBy" validate:"required,min=3"`
	UploadedAt     time.Time `json:"uploadedAt"`
}

// Builder provides a convenient way to construct and configure an Attachment instance.
type Builder struct {
	attachment Attachment
}

// NewBuilder creates and returns a new Builder instance with the UploadedAt field
// of the underlying Attachment initialized to the current time.
func NewBuilder() *Builder {
	return &Builder{attachment: Attachment{
		UploadedAt: time.Now(),
	}}
}

// SetAttachmentId sets the AttachmentId field of the attachment to the provided id.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetAttachmentId(id int) *Builder {
	b.attachment.AttachmentId = id
	return b
}

// SetCardId sets the CardId field of the attachment to the provided cardId.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetCardId(cardId int) *Builder {
	b.attachment.CardId = cardId
	return b
}

// SetFilename sets the filename for the attachment and returns the Builder for chaining.
func (b *Builder) SetFilename(filename string) *Builder {
	b.attachment.Filename = filename
	return b
}

// SetFileType sets the file type of the attachment.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetFileType(fileType string) *Builder {
	b.attachment.FileType = fileType
	return b
}

// SetFileSizeBytes sets the file size in bytes for the attachment.
// It returns the Builder to allow for method chaining.
func (b *Builder) SetFileSizeBytes(size int) *Builder {
	b.attachment.FileSizeBytes = size
	return b
}

// SetAttachmentData sets the attachment data for the builder.
// It assigns the provided byte slice to the AttachmentData field of the attachment.
// Returns the builder to allow for method chaining.
func (b *Builder) SetAttachmentData(data []byte) *Builder {
	b.attachment.AttachmentData = data
	return b
}

// SetUploadedBy sets the UploadedBy field of the attachment to the provided value.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetUploadedBy(uploadedBy string) *Builder {
	b.attachment.UploadedBy = uploadedBy
	return b
}

// SetUploadedAt sets the UploadedAt field of the attachment to the specified time.
// Returns the Builder to allow for method chaining.
func (b *Builder) SetUploadedAt(uploadedAt time.Time) *Builder {
	b.attachment.UploadedAt = uploadedAt
	return b
}

// Build validates the Attachment being constructed by the Builder.
// If validation passes, it returns the built Attachment and a nil error.
// If validation fails, it returns an empty Attachment and the validation error.
func (b *Builder) Build() (Attachment, error) {
	err := validate.Struct(b.attachment)
	if err != nil {
		return Attachment{}, err
	}
	return b.attachment, nil
}

// TransformAttachment converts a db.CardAttachment object to an Attachment object,
// mapping all relevant fields and performing necessary type conversions.
// It is typically used to transform database models into API response models.
func TransformAttachment(ca db.CardAttachment) Attachment {
	return Attachment{
		AttachmentId:   int(ca.AttachmentID),
		CardId:         int(ca.CardID),
		Filename:       ca.Filename,
		FileType:       ca.FileType.String,
		FileSizeBytes:  int(ca.FileSizeBytes.Int64),
		AttachmentData: ca.AttachmentData,
		UploadedBy:     ca.UploadedBy,
		UploadedAt:     ca.UploadedAt.Time,
	}
}

// TransformAttachments converts a slice of db.CardAttachment objects into a slice of Attachment objects
// by applying the TransformAttachment function to each element.
// It returns the resulting slice of Attachment.
func TransformAttachments(cas []db.CardAttachment) []Attachment {
	attachments := make([]Attachment, len(cas))
	for idx, ca := range cas {
		attachments[idx] = TransformAttachment(ca)
	}
	return attachments
}
