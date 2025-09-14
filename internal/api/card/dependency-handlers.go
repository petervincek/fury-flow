package card

import (
	"github.com/gofiber/fiber/v2"
	"github.com/petervincek/fury-flow/internal/api/common"
	"go.uber.org/zap"
)

// GetCardDependencies handles the HTTP request to retrieve dependencies for a specific kanban card.
// It parses and validates the board and card IDs from the request context, checks their existence,
// and then fetches the card's dependencies from the service layer. On success, it returns the dependencies
// as a JSON response; on failure, it returns an appropriate error response.
//
// Expected URL parameters:
//   - boardId: the ID of the kanban board
//   - cardId: the ID of the kanban card
//
// Responses:
//   - 200 OK: JSON array of card dependencies
//   - 400 Bad Request: if parameters are invalid
//   - 404 Not Found: if board or card does not exist
//   - 500 Internal Server Error: on unexpected errors
func (ch *CardHandler) GetCardDependencies(ctx *fiber.Ctx) error {
	// parse and handle boardId
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

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// check the existence of kanban card
	checkResult = common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.kcs.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// get the card dependencies
	cardDependencies, err := ch.kcs.GetCardDependencies(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error("error while getting card dependencies", zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusOK).JSON(TransformCards(cardDependencies))
}

// CreateCardDependencies handles the creation of dependencies for a kanban card.
// It parses the board ID and card ID from the request context, validates their existence,
// parses the list of dependency card IDs from the request body, and creates the dependencies.
// Returns appropriate HTTP error responses if parsing or validation fails, or if dependency creation encounters an error.
// On success, returns HTTP 201 Created with an empty JSON object.
func (ch *CardHandler) CreateCardDependencies(ctx *fiber.Ctx) error {
	// parse and handle boardId
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

	dependencyCardIds := []int32{}
	if err := ctx.BodyParser(&dependencyCardIds); err != nil {
		logger.Error("error parsing dependency card ids", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse dependency card ids", nil))
	}

	if len(dependencyCardIds) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("at least one dependency card ID must be provided", nil))
	}

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// check the existence of kanban card
	checkResult = common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.kcs.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// create card dependencies
	err := ch.kcs.CreateCardDependencies(ctx.Context(), int32(cardId), dependencyCardIds)
	if err != nil {
		logger.Error("error while creating card dependencies", zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.Status(fiber.StatusCreated).JSON(struct{}{})
}

// DeleteCardDependencies handles the HTTP request to delete all dependencies of a specific card within a kanban board.
// It parses and validates the board and card IDs from the request context, checks for their existence,
// and then deletes all dependencies associated with the specified card.
// Returns a 204 No Content status on success, or an appropriate error response if validation or deletion fails.
func (ch *CardHandler) DeleteCardDependencies(ctx *fiber.Ctx) error {
	// parse and handle boardId
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

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// check the existence of kanban card
	checkResult = common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.kcs.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// delete all card dependencies
	err := ch.kcs.DeleteCardDependencies(ctx.Context(), int32(cardId))
	if err != nil {
		logger.Error("error while deleting card dependencies", zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// DeleteCardDependency handles the HTTP request to delete a card dependency from a kanban board.
// It parses and validates the board ID, card ID, and card dependency ID from the request context.
// The handler checks for the existence of the specified board, card, and card dependency.
// If all validations pass, it deletes the card dependency and returns a 204 No Content status.
// In case of errors (parsing, validation, or deletion), it responds with the appropriate HTTP status and error message.
func (ch *CardHandler) DeleteCardDependency(ctx *fiber.Ctx) error {
	// parse and handle boardId
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

	cardDependencyId, err := ctx.ParamsInt(CARD_DEPENDENCY_ID_PARAM)
	if err != nil {
		logger.Error("error parsing card dependency id value", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse card dependency id", nil))
	}

	// check the existence of kanban board
	checkResult := common.CheckAndHandleExistenceOfBoard(ctx, boardId, func() (bool, error) {
		return ch.kbs.KanbanBoardExists(ctx.Context(), int32(boardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// check the existence of kanban card
	checkResult = common.CheckAndHandleExistenceOfCard(ctx, boardId, cardId, func() (bool, error) {
		return ch.kcs.KanbanCardExists(ctx.Context(), int32(cardId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// check the existence of kanban card dependency
	checkResult = common.CheckAndHandleExistenceOfCardDependency(ctx, boardId, cardId, cardDependencyId, func() (bool, error) {
		return ch.kcs.KanbanCardExists(ctx.Context(), int32(cardDependencyId))
	})
	if checkResult.IsLeft() {
		return checkResult.Left()
	}

	// delete card dependency
	err = ch.kcs.DeleteCardDependency(ctx.Context(), int32(cardId), int32(cardDependencyId))
	if err != nil {
		logger.Error("error while deleting card dependency", zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId),
			zap.Int(CARD_DEPENDENCY_ID_PARAM, cardDependencyId))
		return ctx.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
