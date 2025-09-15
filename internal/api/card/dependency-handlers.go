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
//
// GetCardDependencies handles the HTTP request to retrieve dependencies of a specific kanban card.
// It parses the boardId and cardId from the request context, checks their existence,
// and returns the list of card dependencies in JSON format.
//
// @Summary      Get card dependencies
// @Description  Retrieves the dependencies of a specific kanban card within a board.
// @Tags         cards
// @Produce      json
// @Param        boardId  path      int  true  "Board ID"
// @Param        cardId   path      int  true  "Card ID"
// @Success      200      {array}   card.Card                 "List of card dependencies"
// @Failure      400      {object}  common.ErrorMsg           "Invalid board or card ID"
// @Failure      404      {object}  common.ErrorMsg           "Board or card not found"
// @Failure      500      {object}  common.ErrorMsg           "Internal server error"
// @Router       /boards/{boardId}/cards/{cardId}/dependencies [get]
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
//
// It parses the boardId and cardId from the request context, validates their existence,
// and expects a JSON array of dependency card IDs in the request body. At least one dependency
// card ID must be provided. If all validations pass, it creates the card dependencies.
//
// @Summary      Create card dependencies
// @Description  Creates dependencies for a kanban card by specifying dependent card IDs.
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        boardId  path      int     true  "Board ID"
// @Param        cardId   path      int     true  "Card ID"
// @Param        body     body      []int32 true  "Array of dependency card IDs"
// @Success      201      "No Content"
// @Failure      400      {object}  common.ErrorMsg "Invalid input or missing dependency card IDs"
// @Failure      404      {object}  common.ErrorMsg "Board or card not found"
// @Failure      500      {object}  common.ErrorMsg "Internal server error"
// @Router       /boards/{boardId}/cards/{cardId}/dependencies [post]
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
	return ctx.Status(fiber.StatusCreated).JSON(nil)
}

// DeleteCardDependencies handles the HTTP request to delete all dependencies of a specific card within a kanban board.
// It parses and validates the board and card IDs from the request context, checks for their existence,
// and then deletes all dependencies associated with the specified card.
// Returns a 204 No Content status on success, or an appropriate error response if validation or deletion fails.
//
// @Summary Delete all dependencies of a kanban card
// @Description Deletes all dependencies for the specified card in the given board.
// @Tags cards
// @Param boardId path int true "Board ID"
// @Param cardId path int true "Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} common.ErrorMsg "Invalid board or card ID"
// @Failure 404 {object} common.ErrorMsg "Board or card not found"
// @Failure 500 {object} common.ErrorMsg "Internal server error"
// @Router /boards/{boardId}/cards/{cardId}/dependencies [delete]
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
// DeleteCardDependency handles the deletion of a card dependency from a kanban card.
//
// It parses and validates the board ID, card ID, and card dependency ID from the request context,
// checks the existence of the board, card, and card dependency, and then deletes the specified card dependency.
// Returns appropriate HTTP status codes and error messages for invalid input or internal errors.
//
// @Summary Delete a card dependency
// @Description Deletes a dependency from a kanban card by its ID.
// @Tags cards
// @Param boardId path int true "Board ID"
// @Param cardId path int true "Card ID"
// @Param cardDependencyId path int true "Card Dependency ID"
// @Success 204 "No Content"
// @Failure 400 {object} common.ErrorMsg "Invalid input"
// @Failure 404 {object} common.ErrorMsg "Not found"
// @Failure 500 {object} common.ErrorMsg "Internal server error"
// @Router /boards/{boardId}/cards/{cardId}/dependencies/{cardDependencyId} [delete]
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
