package card

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	INTERNAL_SERVER_ERROR            = "internal server error"
	BOARD_ID_PARAM                   = "boardId"
	CARD_ID_PARAM                    = "cardId"
	CARD_DEPENDENCY_ID_PARAM         = "cardDependencyId"
	KANBAN_CARD_VALIDATION_ERROR_MSG = "Kanban card validation error"
	CARD_NOT_PART_OF_BOARD_ERROR_MSG = "requested kanban card is not part of provided kanban board"
)

var logger = logging.GetLogger()
var validate = validator.New()

// CardHandler provides HTTP handlers for operations related to Kanban boards and cards.
// It uses KanbanBoardService for board-related logic and KanbanCardService for card-related logic.
type CardHandler struct {
	kbs *service.KanbanBoardService
	kcs *service.KanbanCardService
}

// New creates and returns a new instance of CardHandler with the provided KanbanBoardService and KanbanCardService.
// It initializes the CardHandler with the given services for board and card operations.
func New(kbs *service.KanbanBoardService, kcs *service.KanbanCardService) *CardHandler {
	return &CardHandler{
		kbs: kbs,
		kcs: kcs,
	}
}

// GetCards handles the HTTP request to retrieve all Kanban cards.
// It fetches the cards using the KanbanCardService and returns them as a JSON response.
// In case of an error during retrieval, it logs the error and responds with an internal server error message.
func (ch *CardHandler) GetCards(ctx *fiber.Ctx) error {
	kanbanCards, err := ch.kcs.GetKanbanCards(ctx.Context())
	if err != nil {
		logger.Error("error while getting all the kanban cards", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(200).JSON(TransformCards(kanbanCards))
}

// GetCardsForBoard handles the HTTP request to retrieve all kanban cards for a specific board.
// It parses the board ID from the request path, checks if the board exists, and then fetches the cards.
// Returns a JSON response containing the cards if successful, or an appropriate error message otherwise.
func (ch *CardHandler) GetCardsForBoard(ctx *fiber.Ctx) error {
	// parse the board id param (path variable)
	result := common.ParseBoardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	boardId := result.Right()

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// get/fetch the cards for the given board id
	cards, err := ch.kcs.GetKanbanCardsForBoard(ctx.Context(), int32(boardId))
	if err != nil {
		logger.Error(fmt.Sprintf("error while getting kanban cards for boardId, err: %v", err), zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformCards(cards))
}

// GetCardById handles the HTTP request to retrieve a kanban card by its ID and associated board ID.
// It parses the board and card IDs from the request context, checks for their validity, and ensures
// the card belongs to the specified board. If the card is found and belongs to the board, it returns
// the card data as JSON. Otherwise, it responds with appropriate error messages and status codes.
//
// Possible responses:
//   - 200 OK: Returns the card data in JSON format.
//   - 400 Bad Request: If the card does not belong to the specified board.
//   - 404 Not Found: If no card is found for the given card ID.
//   - 500 Internal Server Error: For unexpected errors during retrieval.
//
// Parameters:
//   - ctx: *fiber.Ctx - The Fiber context containing request information.
//
// Returns:
//   - error: An error if the request could not be processed.
func (ch *CardHandler) GetCardById(ctx *fiber.Ctx) error {
	// we are going to parse board id along card id, there is an ownership relationship between board and card
	result := common.ParseBoardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	boardId := result.Right()

	// parse and handle cardId
	result = common.ParseCardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	cardId := result.Right()

	kanbanCard, err := ch.kcs.GetKanbanCardById(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error(fmt.Sprintf("Error while getting kanban card by cardId, err: %v", err),
			zap.String("type", fmt.Sprintf("%T", err)), zap.Bool("is ErrNoRows", errors.Is(err, sql.ErrNoRows)),
			zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return ctx.Status(fiber.StatusNotFound).JSON(common.NewErrorMsg(fmt.Sprintf("no kanban card for cardId: %d", cardId), nil))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	// check if the card belongs to the board based on the ids from the request (board and card ids are part of path)
	if kanbanCard.BoardID != int32(boardId) {
		logger.Error(CARD_NOT_PART_OF_BOARD_ERROR_MSG,
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(CARD_NOT_PART_OF_BOARD_ERROR_MSG, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformCard(kanbanCard))
}

// CreateCard handles the creation of a new Kanban card under a specified board.
// It performs the following steps:
//  1. Parses the board ID from the request context.
//  2. Checks if the specified board exists.
//  3. Parses and validates the card data from the request body.
//  4. Attempts to create the card in the database.
//  5. Returns the created card as a JSON response with HTTP 201 status on success.
//  6. Handles and returns appropriate errors for invalid input, non-existent boards, or database failures.
//
// Returns:
//   - HTTP 201 Created with the created card on success.
//   - HTTP 400 Bad Request or HTTP 404 Not Found for invalid input or non-existent board.
//   - HTTP 500 Internal Server Error for database or unexpected errors.
func (ch *CardHandler) CreateCard(ctx *fiber.Ctx) error {
	// we need to know under which board to create a new card
	result := common.ParseBoardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	boardId := result.Right()

	// we need to check the existence of this board, we can create cards just for already existing boards
	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// parse and validate the card from the request body
	card := &Card{}
	resultCheck := parseAndValidateCard(ctx, card)
	if resultCheck.IsLeft() {
		return resultCheck.Left()
	}

	// try to create the card
	timeSpentHours := pgtype.Numeric{}
	if card.TimeSpentHours != 0 {
		timeSpentHours.Scan(card.TimeSpentHours)
	}
	cardToCreate := db.CreateKanbanCardParams{
		BoardID:            int32(boardId),
		Title:              card.Title,
		Description:        pgtype.Text{String: card.Description, Valid: true},
		Status:             string(card.Status),
		Assignee:           pgtype.Text{String: card.Assignee, Valid: true},
		StoryPoints:        pgtype.Int2{Int16: int16(card.StoryPoints), Valid: true},
		AcceptanceCriteria: pgtype.Text{String: card.AcceptanceCriteria, Valid: true},
		TimeSpentHours:     timeSpentHours,
		CreatedBy:          card.CreatedBy,
		Priority:           pgtype.Text{String: string(card.Priority), Valid: true},
		Type:               pgtype.Text{String: string(card.Type), Valid: true},
		IsBlocked:          pgtype.Bool{Bool: card.IsBlocked, Valid: true},
		BlockedReason:      pgtype.Text{String: card.BlockedReason, Valid: true},
	}
	createdKanbanCard, err := ch.kcs.CreateKanbanCard(ctx.Context(), cardToCreate)
	if err != nil {
		// log and handle the error
		logger.Error("error while creating kanban card", zap.Error(err), zap.Int(BOARD_ID_PARAM, boardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusCreated).JSON(TransformCard(createdKanbanCard))
}

// UpdateCardById handles the HTTP request to update a kanban card by its ID.
// It performs the following steps:
//  1. Parses the board ID and card ID from the request context.
//  2. Checks the existence of the specified kanban board.
//  3. Fetches the existing card and verifies it belongs to the board.
//  4. Parses and validates the updated card data from the request body.
//  5. Prepares the update parameters and updates the card in the database.
//  6. Returns appropriate HTTP responses based on success or error conditions.
//
// Returns:
//   - 200 OK on successful update.
//   - 400 Bad Request if the card does not belong to the board or validation fails.
//   - 404 Not Found if the card does not exist.
//   - 500 Internal Server Error for unexpected errors.
//
// Fiber context is used for request handling and response generation.
func (ch *CardHandler) UpdateCardById(ctx *fiber.Ctx) error {
	// parse board id and card id
	result := common.ParseBoardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	boardId := result.Right()

	// parse and handle boardId
	result = common.ParseCardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	cardId := result.Right()

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// fetch the existing card
	existingCard, err := ch.kcs.GetKanbanCardById(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error(fmt.Sprintf("Error while getting kanban card by cardId, err: %v", err),
			zap.String("type", fmt.Sprintf("%T", err)), zap.Bool("is ErrNoRows", errors.Is(err, sql.ErrNoRows)),
			zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			return ctx.Status(fiber.StatusNotFound).JSON(common.NewErrorMsg(fmt.Sprintf("no kanban card for cardId: %d", cardId), nil))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	if existingCard.BoardID != int32(boardId) {
		logger.Error(CARD_NOT_PART_OF_BOARD_ERROR_MSG,
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(CARD_NOT_PART_OF_BOARD_ERROR_MSG, nil))
	}

	// parse and validate the card from the request body
	card := &Card{}
	resultCheck := parseAndValidateCard(ctx, card)
	if resultCheck.IsLeft() {
		return resultCheck.Left()
	}

	// prepare update params
	timeSpentHours := pgtype.Numeric{}
	if card.TimeSpentHours != 0 {
		timeSpentHours.Scan(card.TimeSpentHours)
	}
	updateParams := db.UpdateKanbanCardParams{
		CardID:             int32(cardId),
		Title:              card.Title,
		Description:        pgtype.Text{String: card.Description, Valid: true},
		Status:             string(card.Status),
		Assignee:           pgtype.Text{String: card.Assignee, Valid: true},
		StoryPoints:        pgtype.Int2{Int16: int16(card.StoryPoints), Valid: true},
		AcceptanceCriteria: pgtype.Text{String: card.AcceptanceCriteria, Valid: true},
		TimeSpentHours:     timeSpentHours,
		Priority:           pgtype.Text{String: string(card.Priority), Valid: true},
		Type:               pgtype.Text{String: string(card.Type), Valid: true},
		IsBlocked:          pgtype.Bool{Bool: card.IsBlocked, Valid: true},
		BlockedReason:      pgtype.Text{String: card.BlockedReason, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedBy:          pgtype.Text{String: card.UpdatedBy, Valid: true},
	}

	// update kanban card
	err = ch.kcs.UpdateKanbanCard(ctx.Context(), updateParams)
	if err != nil {
		logger.Error("error while updating kanban card", zap.Error(err), zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(struct{}{})
}

// DeleteCardById handles the HTTP request to delete a kanban card by its ID.
// It parses the board and card IDs from the request context, checks for the existence
// of the specified kanban board, and deletes the card if all validations pass.
// Returns a 204 No Content status on success, or an appropriate error response on failure.
func (ch *CardHandler) DeleteCardById(ctx *fiber.Ctx) error {
	// parse board id and card id
	result := common.ParseBoardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	boardId := result.Right()

	// parse and handle boardId
	result = common.ParseCardId(ctx)
	if result.IsLeft() {
		return result.Left()
	}
	cardId := result.Right()

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// delete kanban card
	err := ch.kcs.DeleteKanbanCard(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error(fmt.Sprintf("Error while deleting card: %d, err: %v", cardId, err), zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// parseAndValidateCard parses the request body into the provided Card struct and validates its fields.
// It returns a monad.Result containing an empty struct on success, or an error result if parsing or validation fails.
// In case of error, an appropriate HTTP response is set on the fiber context with a descriptive error message.
// Validation errors are reported with details about which fields failed and the validation tags involved.
func parseAndValidateCard(ctx *fiber.Ctx, card *Card) monad.Result[struct{}] {
	// parse and validate the card from the request body
	if err := ctx.BodyParser(card); err != nil {
		logger.Error(fmt.Sprintf("error while parsing card entity, err: %v", err),
			zap.String("type", fmt.Sprintf("%T", err)), zap.Error(err))
		e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse card entity", nil))
		return monad.NewResultError[struct{}](&e)
	}
	//validate the card
	if err := validate.Struct(card); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			logger.Error(fmt.Sprintf("Unexpected validation error type, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
			e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_CARD_VALIDATION_ERROR_MSG, []string{"invalid input"}))
			return monad.NewResultError[struct{}](&e)
		}
		errMsgs := make([]string, len(errs))
		for idx, validationErr := range errs {
			errMsgs[idx] = fmt.Sprintf("Field '%s' failed on the '%s' tag", validationErr.Field(), validationErr.Tag())
		}
		e := ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_CARD_VALIDATION_ERROR_MSG, errMsgs))
		return monad.NewResultError[struct{}](&e)
	}
	return monad.NewResult(&struct{}{})
}
