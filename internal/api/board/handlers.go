package board

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
	"go.uber.org/zap"
)

const (
	INTERNAL_SERVER_ERROR             = "internal server error"
	ERROR_PARSING_BOARD_ID_MSG        = "Error while parsing boardId, err: %v"
	UNABLE_TO_PARSE_BOARD_ID_MSG      = "unable to parse boardId: %v"
	BOARD_ID_PARAM                    = "boardId"
	KANBAN_BOARD_VALIDATION_ERROR_MSG = "Kanban board validation error"
)

var logger = logging.GetLogger()
var validate = validator.New()

// BoardHandler provides HTTP handlers for operations related to Kanban boards.
// It uses KanbanBoardService to perform business logic for board management.
type BoardHandler struct {
	kbs *service.KanbanBoardService
}

// New creates and returns a new BoardHandler instance using the provided KanbanBoardService.
func New(kbs *service.KanbanBoardService) *BoardHandler {
	return &BoardHandler{
		kbs: kbs,
	}
}

// GetBoards handles the HTTP request to retrieve all Kanban boards.
// It fetches the boards using the Kanban board service and returns them as a JSON response.
// In case of an error during retrieval, it responds with an internal server error message.
func (bh *BoardHandler) GetBoards(c *fiber.Ctx) error {
	boards, err := bh.kbs.GetKanbanBoards(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return c.JSON(TransformBoards(boards))
}

// GetBoardById handles the HTTP request to retrieve a Kanban board by its ID.
// It parses the boardId parameter from the request, validates it, and fetches the corresponding board.
// Returns a JSON response with the board data if found, or an appropriate error message and status code
// if the boardId is invalid, the board is not found, or an internal error occurs.
func (bh *BoardHandler) GetBoardById(c *fiber.Ctx) error {
	boardId, err := c.ParamsInt(BOARD_ID_PARAM)
	if err != nil {
		logger.Error(fmt.Sprintf(ERROR_PARSING_BOARD_ID_MSG, err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(fmt.Sprintf(UNABLE_TO_PARSE_BOARD_ID_MSG, boardId), nil))
	}

	board, err := bh.kbs.GetKanbanBoardById(c.Context(), int32(boardId))
	if err != nil {
		logger.Error(fmt.Sprintf("Error while getting kanban by boardId, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)), zap.Bool("is ErrNoRows", errors.Is(err, sql.ErrNoRows)))
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(common.NewErrorMsg(fmt.Sprintf("no kanban board for boardId: %d", boardId), nil))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return c.JSON(TransformBoard(board))
}

// CreateBoard handles the creation of a new Kanban board.
// It parses and validates the incoming request body, then creates the board in the system.
// Returns a JSON response with the created board on success, or an error message on failure.
//
// Possible responses:
//   - 201 Created: Board successfully created.
//   - 400 Bad Request: Invalid input or validation error.
//   - 500 Internal Server Error: Failed to create board due to server error.
//
// Request body should contain the board details in JSON format.
func (bh *BoardHandler) CreateBoard(c *fiber.Ctx) error {
	// try to parse the input from the request
	board := &Board{}
	if err := c.BodyParser(board); err != nil {
		logger.Error(fmt.Sprintf("Error while parsing board entity, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse board entity", nil))
	}
	// validate the input from the request
	if err := validate.Struct(board); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			logger.Error(fmt.Sprintf("Unexpected validation error type, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
			return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_BOARD_VALIDATION_ERROR_MSG, []string{"invalid input"}))
		}
		errMsgs := make([]string, len(errs))
		for idx, validationErr := range errs {
			errMsgs[idx] = fmt.Sprintf("Field '%s' failed on the '%s' tag", validationErr.Field(), validationErr.Tag())
		}
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_BOARD_VALIDATION_ERROR_MSG, errMsgs))
	}
	// create Kanban board in the system
	boardToCreate := db.CreateKanbanBoardParams{
		BoardName:   board.BoardName,
		Description: pgtype.Text{String: board.Description, Valid: true},
		CreatedBy:   board.CreatedBy,
	}
	createdBoard, err := bh.kbs.CreateKanbanBoard(c.Context(), boardToCreate)
	if err != nil {
		logger.Error(fmt.Sprintf("Error while creating board, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return c.Status(fiber.StatusCreated).JSON(TransformBoard(createdBoard))
}

// UpdateBoardById handles the HTTP request to update a Kanban board by its ID.
// It parses the board ID from the request parameters and the board entity from the request body.
// If parsing fails, it returns a 400 Bad Request error.
// It then constructs the update parameters and calls the service to update the board in the database.
// If the board is not found or not updated, it returns a 404 Not Found error.
// For other errors, it returns a 500 Internal Server Error.
// On success, it returns the updated board information with a 200 OK status.
func (bh *BoardHandler) UpdateBoardById(c *fiber.Ctx) error {
	boardId, err := c.ParamsInt(BOARD_ID_PARAM)
	if err != nil {
		logger.Error(fmt.Sprintf(ERROR_PARSING_BOARD_ID_MSG, err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(fmt.Sprintf(UNABLE_TO_PARSE_BOARD_ID_MSG, boardId), nil))
	}
	board := &Board{}
	if err := c.BodyParser(board); err != nil {
		logger.Error(fmt.Sprintf("Error while parsing board entity, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg("unable to parse board entity", nil))
	}
	// validate the input from the request
	if err := validate.Struct(board); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			logger.Error(fmt.Sprintf("Unexpected validation error type, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
			return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_BOARD_VALIDATION_ERROR_MSG, []string{"invalid input"}))
		}
		errMsgs := make([]string, len(errs))
		for idx, validationErr := range errs {
			errMsgs[idx] = fmt.Sprintf("Field '%s' failed on the '%s' tag", validationErr.Field(), validationErr.Tag())
		}
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(KANBAN_BOARD_VALIDATION_ERROR_MSG, errMsgs))
	}
	boardToUpdate := db.UpdateKanbanBoardParams{
		BoardID:     int32(boardId),
		BoardName:   board.BoardName,
		Description: pgtype.Text{String: board.Description, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedBy:   pgtype.Text{String: board.UpdatedBy, Valid: true},
	}
	err = bh.kbs.UpdateKanbanBoard(c.Context(), boardToUpdate)
	if err != nil {
		if errors.Is(err, service.ErrorKanbanBoardNotUpdated) {
			return c.Status(fiber.StatusNotFound).JSON(common.NewErrorMsg(fmt.Sprintf("no kanban board updated for boardId: %d", boardId), nil))
		}
		logger.Error(fmt.Sprintf("Error while updating board, err: %v", err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return c.Status(fiber.StatusOK).JSON(struct{}{})
}

// DeleteBoardById handles the HTTP request to delete a board by its ID.
// It parses the board ID from the request parameters, attempts to delete the board
// using the Kanban service, and returns appropriate HTTP status codes based on the outcome.
// Returns 400 Bad Request if the board ID cannot be parsed, 500 Internal Server Error if deletion fails,
// and 204 No Content on successful deletion.
func (bh *BoardHandler) DeleteBoardById(c *fiber.Ctx) error {
	boardId, err := c.ParamsInt(BOARD_ID_PARAM)
	if err != nil {
		logger.Error(fmt.Sprintf(ERROR_PARSING_BOARD_ID_MSG, err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusBadRequest).JSON(common.NewErrorMsg(fmt.Sprintf(UNABLE_TO_PARSE_BOARD_ID_MSG, boardId), nil))
	}
	err = bh.kbs.DeleteKanbanBoard(c.Context(), int32(boardId))
	if err != nil {
		logger.Error(fmt.Sprintf("Error while deleting board: %d, err: %v", boardId, err), zap.String("type", fmt.Sprintf("%T", err)))
		return c.Status(fiber.StatusInternalServerError).JSON(common.NewErrorMsg(INTERNAL_SERVER_ERROR, nil))
	}
	return c.SendStatus(fiber.StatusNoContent)
}
