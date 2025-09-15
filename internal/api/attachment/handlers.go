package attachment

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/logging"
	"github.com/petervincek/fury-flow/internal/service"
	"github.com/petervincek/fury-flow/internal/utils/monad"
	"go.uber.org/zap"
)

const (
	REQUESTED_ATTACHMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG = "requested attachment does not belong to request card"
	KANBAN_ATTACHMENT_VALIDATION_ERROR_MSG                 = "kanban attachment validation error"
)

var logger = logging.GetLogger()
var validate = validator.New()

// AttachmentHandler provides HTTP handlers for managing attachments within Kanban boards and cards.
// It coordinates operations between board, card, attachment, and integrity services to ensure
// proper attachment management and data consistency.
type AttachmentHandler struct {
	boardService      *service.KanbanBoardService
	cardService       *service.KanbanCardService
	attachmentService *service.KanbanAttachmentService
	integrityService  *service.KanbanIntegrityService
}

// New creates a new instance of AttachmentHandler with the provided services.
// It initializes the handler with Kanban board, card, attachment, and integrity services,
// enabling operations related to attachments in the Kanban workflow.
//
// Parameters:
//   - boardService: Service for managing Kanban boards.
//   - cardService: Service for managing Kanban cards.
//   - attachmentService: Service for managing Kanban attachments.
//   - integrityService: Service for ensuring data integrity.
//
// Returns:
//   - *AttachmentHandler: A pointer to the initialized AttachmentHandler.
func New(boardService *service.KanbanBoardService,
	cardService *service.KanbanCardService,
	attachmentService *service.KanbanAttachmentService,
	integrityService *service.KanbanIntegrityService) *AttachmentHandler {
	return &AttachmentHandler{
		boardService:      boardService,
		cardService:       cardService,
		attachmentService: attachmentService,
		integrityService:  integrityService,
	}
}

// GetAttachmentsForCard handles the HTTP request to retrieve all attachments for a specific card.
// It performs the following steps:
//  1. Parses the board ID and card ID from the request context.
//  2. Checks the existence of the board and card.
//  3. Verifies that the card belongs to the specified board.
//  4. Retrieves the attachments associated with the card.
//  5. Returns the attachments as a JSON response.
//
// Returns an appropriate error response if any validation or retrieval step fails.
//
// @Summary      Get attachments for a card
// @Description  Retrieves all attachments associated with a specific card in a board.
// @Tags         attachments
// @Produce      json
// @Param        boardId   path      int  true  "Board ID"
// @Param        cardId    path      int  true  "Card ID"
// @Success      200  {array}   attachment.Attachment  "List of attachments"
// @Failure      400  {object}  common.ErrorMsg        "Invalid board or card ID"
// @Failure      404  {object}  common.ErrorMsg        "Board or card not found"
// @Failure      500  {object}  common.ErrorMsg        "Internal server error"
// @Router       /boards/{boardId}/cards/{cardId}/attachments [get]
func (ah *AttachmentHandler) GetAttachmentsForCard(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ah.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ah.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// check the integrity between board and card entity
	integrityResult := common.CheckIntegrityThatCardBelongsToBoard(ctx, boardId, cardId, func() (bool, error) {
		return ah.integrityService.CardBelongsToBoard(ctx.Context(), boardId, cardId)
	})
	if integrityResult.IsLeft() {
		return integrityResult.Left()
	}

	// try to get the attachments for a given card (card id)
	attachments, err := ah.attachmentService.GetKanbanCardAttachments(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error("error while getting card's attachments", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformAttachments(attachments))
}

// GetAttachmentById handles the HTTP request to retrieve a specific attachment by its ID.
// It performs the following steps:
//  1. Parses the board, card, and attachment IDs from the request context.
//  2. Checks the existence of the board and card.
//  3. Verifies that the card belongs to the specified board.
//  4. Retrieves the attachment by its ID.
//  5. Ensures the attachment belongs to the specified card.
//  6. Returns the attachment as a JSON response if all checks pass.
//
// Returns appropriate HTTP error responses for invalid input, non-existent resources,
// integrity violations, or internal server errors.
//
// @Summary      Get Kanban Card Attachment by ID
// @Description  Retrieves a specific attachment for a kanban card by its ID, ensuring board and card integrity.
// @Tags         attachments
// @Produce      json
// @Param        boardId      path      int     true  "Board ID"
// @Param        cardId       path      int     true  "Card ID"
// @Param        attachmentId path      int     true  "Attachment ID"
// @Success      200  {object}  attachment.Attachment  "Attachment found"
// @Failure      400  {object}  common.ErrorMsg        "Invalid request or integrity error"
// @Failure      404  {object}  common.ErrorMsg        "Attachment not found"
// @Failure      500  {object}  common.ErrorMsg        "Internal server error"
// @Router       /boards/{boardId}/cards/{cardId}/attachments/{attachmentId} [get]
func (ah *AttachmentHandler) GetAttachmentById(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// try to parse attachment id
	attachmentIdResult := common.ParseAttachmentId(ctx)
	if attachmentIdResult.IsLeft() {
		return attachmentIdResult.Left()
	}
	attachmentId := attachmentIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ah.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ah.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// check the integrity between board and card entity
	integrityResult := common.CheckIntegrityThatCardBelongsToBoard(ctx, boardId, cardId, func() (bool, error) {
		return ah.integrityService.CardBelongsToBoard(ctx.Context(), boardId, cardId)
	})
	if integrityResult.IsLeft() {
		return integrityResult.Left()
	}

	// try to get the attachment (by attachment id)
	attachment, err := ah.attachmentService.GetKanbanAttachmentById(ctx.Context(), int32(attachmentId))
	if err != nil {
		logger.Error("error while getting card attachment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.ATTACHMENT_ID_PARAM, attachmentId))
		if errors.Is(err, sql.ErrNoRows) {
			return ctx.Status(fiber.StatusNotFound).JSON(common.NewErrorMsg(fmt.Sprintf("no kanban attachment for attachmentId: %d", attachmentId), nil))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}

	// check that the attachment belong to the card
	if attachment.CardID != int32(cardId) {
		logger.Error(REQUESTED_ATTACHMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.ATTACHMENT_ID_PARAM, attachmentId))
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewErrorMsg(REQUESTED_ATTACHMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformAttachment(attachment))
}

// CreateAttachment handles the creation of a new attachment for a specific card within a board.
// It performs the following steps:
//  1. Parses and validates the board ID from the request context.
//  2. Parses and validates the card ID from the request context.
//  3. Parses and validates the attachment data from the request body.
//  4. Checks the existence of the specified board.
//  5. Checks the existence of the specified card.
//  6. Verifies that the card belongs to the specified board.
//  7. Creates the attachment using the provided attachment service.
//  8. Returns the created attachment as a JSON response with HTTP 201 status.
//
// Returns an appropriate error response if any validation or integrity check fails,
// or if the attachment creation encounters an error.
//
// @Summary Create a new attachment for a card
// @Description Creates a new attachment for the specified card within a board.
// @Tags attachments
// @Accept  json
// @Produce json
// @Param boardId path int true "Board ID"
// @Param cardId  path int true "Card ID"
// @Param attachment body attachment.Attachment true "Attachment object"
// @Success 201 {object} attachment.Attachment "Created attachment"
// @Failure 400 {object} common.ErrorMsg       "Invalid input or entity not found"
// @Failure 500 {object} common.ErrorMsg       "Internal server error"
// @Router /boards/{boardId}/cards/{cardId}/attachments [post]
func (ah *AttachmentHandler) CreateAttachment(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// try to parse the attachment
	// parse and validate the attachment from the request body
	attachment := &Attachment{}
	resultCheck := parseAndValidateAttachment(ctx, attachment)
	if resultCheck.IsLeft() {
		return resultCheck.Left()
	}

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ah.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ah.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// check the integrity between board and card entity
	integrityResult := common.CheckIntegrityThatCardBelongsToBoard(ctx, boardId, cardId, func() (bool, error) {
		return ah.integrityService.CardBelongsToBoard(ctx.Context(), boardId, cardId)
	})
	if integrityResult.IsLeft() {
		return integrityResult.Left()
	}

	// try to create card's attachment
	createdAttachment, err := ah.attachmentService.CreateKanbanAttachment(ctx.Context(), db.CreateKanbanAttachmentParams{
		CardID:         int32(attachment.CardId),
		Filename:       attachment.Filename,
		FileType:       pgtype.Text{String: attachment.FileType, Valid: true},
		FileSizeBytes:  pgtype.Int8{Int64: int64(attachment.FileSizeBytes), Valid: true},
		AttachmentData: attachment.AttachmentData,
		UploadedBy:     attachment.UploadedBy,
		UploadedAt:     pgtype.Timestamptz{Time: attachment.UploadedAt, Valid: true},
	})
	if err != nil {
		logger.Error("error while creating new attachment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}

	return ctx.Status(fiber.StatusCreated).JSON(TransformAttachment(createdAttachment))
}

// DeleteAttachmentsForCard handles the HTTP request to delete all attachments associated with a specific card.
// It performs the following steps:
//  1. Parses the board ID and card ID from the request context.
//  2. Checks the existence of the board and card.
//  3. Verifies that the card belongs to the specified board.
//  4. Deletes all attachments linked to the card.
//
// If any validation or deletion fails, it returns an appropriate error response.
// On success, it returns a 204 No Content status.
//
// @Summary Delete all attachments for a card
// @Description Deletes all attachments associated with the specified card in the given board.
// @Tags attachments
// @Param boardId path int true "Board ID"
// @Param cardId path int true "Card ID"
// @Produce json
// @Success 204 "No Content"
// @Failure 400 {object} common.ErrorMsg "Invalid board or card ID"
// @Failure 404 {object} common.ErrorMsg "Board or card not found"
// @Failure 500 {object} common.ErrorMsg "Internal server error"
// @Router /boards/{boardId}/cards/{cardId}/attachments [delete]
func (ah *AttachmentHandler) DeleteAttachmentsForCard(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ah.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ah.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// check the integrity between board and card entity
	integrityResult := common.CheckIntegrityThatCardBelongsToBoard(ctx, boardId, cardId, func() (bool, error) {
		return ah.integrityService.CardBelongsToBoard(ctx.Context(), boardId, cardId)
	})
	if integrityResult.IsLeft() {
		return integrityResult.Left()
	}

	// try to delete all attachments associated with the given card
	err := ah.attachmentService.DeleteKanbanAttachmentsForCard(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error("error while deleting all card's attachments", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// DeleteAttachment handles the deletion of a specific attachment from a card within a board.
// It performs the following steps:
//  1. Parses the board, card, and attachment IDs from the request context.
//  2. Checks the existence of the board and card.
//  3. Verifies that the card belongs to the specified board.
//  4. Retrieves the attachment and ensures it belongs to the specified card.
//  5. Deletes the attachment if all checks pass.
//
// Returns appropriate HTTP status codes and error messages for validation, integrity, and internal errors.
//
// @Summary Delete an attachment from a kanban card
// @Description Deletes a specific attachment from a card in a kanban board after validating all relationships and existence.
// @Tags attachments
// @Param boardId path int true "Board ID"
// @Param cardId path int true "Card ID"
// @Param attachmentId path int true "Attachment ID"
// @Success 204 "No Content"
// @Failure 400 {object} common.ErrorMsg "Bad Request"
// @Failure 404 {object} common.ErrorMsg "Not Found"
// @Failure 500 {object} common.ErrorMsg "Internal Server Error"
// @Router /boards/{boardId}/cards/{cardId}/attachments/{attachmentId} [delete]
func (ah *AttachmentHandler) DeleteAttachment(ctx *fiber.Ctx) error {
	// try to parse board id
	boardIdResult := common.ParseBoardId(ctx)
	if boardIdResult.IsLeft() {
		return boardIdResult.Left()
	}
	boardId := boardIdResult.Right()

	// try to parse card id
	cardIdResult := common.ParseCardId(ctx)
	if cardIdResult.IsLeft() {
		return cardIdResult.Left()
	}
	cardId := cardIdResult.Right()

	// try to parse attachment id
	attachmentIdResult := common.ParseAttachmentId(ctx)
	if attachmentIdResult.IsLeft() {
		return attachmentIdResult.Left()
	}
	attachmentId := attachmentIdResult.Right()

	// check the existence of the board
	boardExistenceResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ah.boardService.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if boardExistenceResult.IsLeft() {
		return boardExistenceResult.Left()
	}

	// check the existence of the card
	cardExistenceResult := common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ah.cardService.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if cardExistenceResult.IsLeft() {
		return cardExistenceResult.Left()
	}

	// check the integrity between board and card entity
	integrityResult := common.CheckIntegrityThatCardBelongsToBoard(ctx, boardId, cardId, func() (bool, error) {
		return ah.integrityService.CardBelongsToBoard(ctx.Context(), boardId, cardId)
	})
	if integrityResult.IsLeft() {
		return integrityResult.Left()
	}

	// try to get the attachments for a given card (card id)
	attachment, err := ah.attachmentService.GetKanbanAttachmentById(ctx.Context(), int32(attachmentId))
	if err != nil {
		logger.Error("error while getting card attachment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.ATTACHMENT_ID_PARAM, attachmentId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}

	// check that the attachment belong to the card
	if attachment.CardID != int32(cardId) {
		logger.Error(REQUESTED_ATTACHMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.ATTACHMENT_ID_PARAM, attachmentId))
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewErrorMsg(REQUESTED_ATTACHMENT_DOES_NOT_BELONG_TO_CARD_ERROR_MSG, nil))
	}

	// try to delete that one specific/requested attachment
	err = ah.attachmentService.DeleteKanbanAttachment(ctx.Context(), int32(attachmentId))
	if err != nil {
		logger.Error("error while deleting card attachment", zap.Error(err), zap.Int(common.BOARD_ID_PARAM, boardId),
			zap.Int(common.CARD_ID_PARAM, cardId), zap.Int(common.ATTACHMENT_ID_PARAM, attachmentId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(common.INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// parseAndValidateAttachment parses the request body into the provided Attachment struct and validates its fields.
// It returns a monad.Result containing an empty struct on success, or an error result with a JSON error response on failure.
// Parsing errors and validation errors are logged and returned as HTTP 400 Bad Request responses.
// Validation errors include detailed field and tag information in the error message.
func parseAndValidateAttachment(ctx *fiber.Ctx, attachment *Attachment) monad.Result[struct{}] {
	// parse and validate the attachment from the request body
	if err := ctx.BodyParser(attachment); err != nil {
		logger.Error(fmt.Sprintf("error while parsing attachment entity, err: %v", err),
			zap.String("type", fmt.Sprintf("%T", err)), zap.Error(err))
		e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse attachment entity", nil))
		return monad.NewResultError[struct{}](&e)
	}
	//validate the card
	if err := validate.Struct(attachment); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			logger.Error(fmt.Sprintf("Unexpected validation error type, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
			e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_ATTACHMENT_VALIDATION_ERROR_MSG, []string{"invalid input"}))
			return monad.NewResultError[struct{}](&e)
		}
		errMsgs := make([]string, len(errs))
		for idx, validationErr := range errs {
			errMsgs[idx] = fmt.Sprintf("Field '%s' failed on the '%s' tag", validationErr.Field(), validationErr.Tag())
		}
		e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_ATTACHMENT_VALIDATION_ERROR_MSG, errMsgs))
		return monad.NewResultError[struct{}](&e)
	}
	return monad.NewResult(&struct{}{})
}
