package common

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/petervincek/fury-flow/internal/logging"
	"github.com/petervincek/fury-flow/internal/utils/monad"
	"go.uber.org/zap"
)

const (
	INTERNAL_SERVER_ERROR                = "internal server error"
	BOARD_ID_PARAM                       = "boardId"
	CARD_ID_PARAM                        = "cardId"
	COMMENT_ID_PARAM                     = "commentId"
	CARD_DEPENDENCY_ID_PARAM             = "cardDependencyId"
	PARSING_BOARD_ID_ERROR_MSG           = "error parsing board id value"
	PARSING_CARD_ID_ERROR_MSG            = "error parsing card id value"
	PARSING_COMMENT_ID_ERROR_MSG         = "error parsing comment id value"
	EXISTENCE_OF_KANBAN_BOARD_ERROR_MSG  = "error while checking existence of kanban board"
	EXISTENCE_OF_KANBAN_CARD_ERROR_MSG   = "error while checking existence of kanban card"
	BOARD_DOES_NOT_EXIST_ERROR_MSG       = "board does not exist"
	BOARD_WITH_ID_DOES_NOT_EXIST_TEMPATE = "board with id: %d does not exist"
	CARD_DOES_NOT_EXIST_ERROR_MSG        = "card does not exist"
	CARD_WITH_ID_DOES_NOT_EXIST_TEMPATE  = "card with id: %d does not exist"
	UNABLE_TO_PARSE_BOARD_ID_ERROR_MSG   = "unable to parse board id"
	UNABLE_TO_PARSE_CARD_ID_ERROR_MSG    = "unable to parse card id"
	UNABLE_TO_PARSE_COMMENT_ID_ERROR_MSG = "unable to parse comment id"
)

var logger = logging.GetLogger()

// ParseBoardId attempts to parse the board ID from the request context parameters.
// If parsing fails, it logs the error, sets the response status to BadRequest, and returns a Result containing the error.
// On success, it returns a Result containing the parsed board ID.
func ParseBoardId(ctx *fiber.Ctx) monad.Result[int] {
	boardId, err := ctx.ParamsInt(BOARD_ID_PARAM)
	if err != nil {
		logger.Error(PARSING_BOARD_ID_ERROR_MSG, zap.Error(err))
		e := ctx.Status(fiber.StatusBadRequest).JSON(NewErrorMsg(UNABLE_TO_PARSE_BOARD_ID_ERROR_MSG, nil))
		return monad.NewResultError[int](&e)
	}
	return monad.NewResult(&boardId)
}

// ParseCardId attempts to parse the card ID from the request context parameters.
// If parsing fails, it logs the error, sets the response status to BadRequest,
// and returns a Result containing the error. On success, it returns a Result containing the parsed card ID.
func ParseCardId(ctx *fiber.Ctx) monad.Result[int] {
	cardId, err := ctx.ParamsInt(CARD_ID_PARAM)
	if err != nil {
		logger.Error(PARSING_CARD_ID_ERROR_MSG, zap.Error(err))
		e := ctx.Status(fiber.StatusBadRequest).JSON(NewErrorMsg(UNABLE_TO_PARSE_CARD_ID_ERROR_MSG, nil))
		return monad.NewResultError[int](&e)
	}
	return monad.NewResult(&cardId)
}

// ParseCommentId extracts and parses the comment ID parameter from the Fiber context.
// It returns a monad.Result containing the parsed integer ID on success, or an error result
// if the parameter is missing or cannot be parsed. In case of error, it logs the issue and
// responds with a Bad Request status and a JSON error message.
func ParseCommentId(ctx *fiber.Ctx) monad.Result[int] {
	commentId, err := ctx.ParamsInt(COMMENT_ID_PARAM)
	if err != nil {
		logger.Error(PARSING_COMMENT_ID_ERROR_MSG, zap.Error(err))
		e := ctx.Status(fiber.StatusBadRequest).JSON(NewErrorMsg(UNABLE_TO_PARSE_COMMENT_ID_ERROR_MSG, nil))
		return monad.NewResultError[int](&e)
	}
	return monad.NewResult(&commentId)
}

// CheckAndHandleExistenceOfBoard checks if a board with the given boardId exists using the provided exists function.
// If an error occurs during the existence check, it logs the error and returns a Result containing an internal server error response.
// If the board does not exist, it logs the error and returns a Result containing a not found error response.
// If the board exists, it returns a successful Result.
// Parameters:
//   - ctx: Fiber context for handling HTTP responses.
//   - boardId: The ID of the board to check existence for.
//   - exists: A function that returns a boolean indicating existence and an error if any.
//
// Returns:
//   - monad.Result[struct{}]: A result containing either a success or an error response.
func CheckAndHandleExistenceOfBoard(ctx *fiber.Ctx, boardId int, exists func() (bool, error)) monad.Result[struct{}] {
	boardExists, err := exists()
	if err != nil {
		logger.Error(EXISTENCE_OF_KANBAN_BOARD_ERROR_MSG, zap.Error(err), zap.Int(BOARD_ID_PARAM, boardId))
		e := ctx.Status(fiber.StatusInternalServerError).JSON(NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
		return monad.NewResultError[struct{}](&e)
	}
	if !boardExists {
		logger.Error(BOARD_DOES_NOT_EXIST_ERROR_MSG, zap.Int(BOARD_ID_PARAM, boardId))
		e := ctx.Status(fiber.StatusNotFound).JSON(NewErrorMsg(fmt.Sprintf(BOARD_WITH_ID_DOES_NOT_EXIST_TEMPATE, boardId), nil))
		return monad.NewResultError[struct{}](&e)
	}
	return monad.NewResult(&struct{}{})
}

// CheckAndHandleExistenceOfCard checks if a card exists for the given board and card IDs using the provided exists function.
// If an error occurs during the existence check, it logs the error and returns a monad.Result containing an internal server error response.
// If the card does not exist, it logs the error and returns a monad.Result containing a not found error response.
// If the card exists, it returns a successful monad.Result.
// Parameters:
//   - ctx: Fiber context for handling HTTP responses.
//   - boardId: The ID of the board.
//   - cardId: The ID of the card.
//   - exists: A function that returns whether the card exists and an error if any.
//
// Returns:
//   - monad.Result[struct{}]: Result indicating success or containing an error response.
func CheckAndHandleExistenceOfCard(ctx *fiber.Ctx, boardId int, cardId int, exists func() (bool, error)) monad.Result[struct{}] {
	cardExists, err := exists()
	if err != nil {
		logger.Error(EXISTENCE_OF_KANBAN_CARD_ERROR_MSG, zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		e := ctx.Status(fiber.StatusInternalServerError).JSON(NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
		return monad.NewResultError[struct{}](&e)
	}
	if !cardExists {
		logger.Error(CARD_DOES_NOT_EXIST_ERROR_MSG,
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId))
		e := ctx.Status(fiber.StatusNotFound).JSON(NewErrorMsg(fmt.Sprintf(CARD_WITH_ID_DOES_NOT_EXIST_TEMPATE, cardId), nil))
		return monad.NewResultError[struct{}](&e)
	}
	return monad.NewResult(&struct{}{})
}

// CheckAndHandleExistenceOfCardDependency checks if a card dependency exists for a given board and card.
// If an error occurs during the existence check, it logs the error and returns a Result containing an internal server error response.
// If the dependency card does not exist, it logs the error and returns a Result containing a not found error response.
// On success, it returns an empty Result.
// Parameters:
//   - ctx: Fiber context for handling HTTP responses.
//   - boardId: ID of the board.
//   - cardId: ID of the card.
//   - cardDependencyId: ID of the dependency card to check.
//   - exists: Function that checks existence and returns (bool, error).
//
// Returns:
//   - monad.Result[struct{}]: Result containing either an error response or success.
func CheckAndHandleExistenceOfCardDependency(ctx *fiber.Ctx, boardId int, cardId int, cardDependencyId int, exists func() (bool, error)) monad.Result[struct{}] {
	dependencyCardExists, err := exists()
	if err != nil {
		logger.Error("error while checking existence of kanban dependency card", zap.Error(err),
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId), zap.Int(CARD_DEPENDENCY_ID_PARAM, cardDependencyId))
		e := ctx.Status(fiber.StatusInternalServerError).JSON(NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
		return monad.NewResultError[struct{}](&e)
	}
	if !dependencyCardExists {
		logger.Error("dependency card does not exist",
			zap.Int(BOARD_ID_PARAM, boardId), zap.Int(CARD_ID_PARAM, cardId), zap.Int(CARD_DEPENDENCY_ID_PARAM, cardDependencyId))
		e := ctx.Status(fiber.StatusNotFound).JSON(NewErrorMsg(fmt.Sprintf("dependency card with id: %d does not exist", cardDependencyId), nil))
		return monad.NewResultError[struct{}](&e)
	}
	return monad.NewResult(&struct{}{})
}
