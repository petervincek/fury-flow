package testdata

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/board"
)

const (
	APP_URL_TEMPLATE   = "%s%s/%d"
	BOARDS_URL_CONTEXT = "/boards"
)

// CreateKanbanBoard creates a new Kanban board by sending a POST request to the application's boards endpoint.
// It takes a board.Board object as input and returns a ResponseResult containing the created board and an error, if any.
func (td *TestData) CreateKanbanBoard(kanbanBoard board.Board) (ResponseResult[board.Board], error) {
	return MakePostRequest[board.Board, board.Board](fmt.Sprintf("%s%s", td.GetAppUrl(), BOARDS_URL_CONTEXT), kanbanBoard)
}

// CreateKanbanBoards creates multiple Kanban boards by calling CreateKanbanBoard for each board in the provided slice.
// It returns a slice of ResponseResult containing the results for each created board, or an error if any creation fails.
// If an error occurs during the creation of any board, the function returns the results accumulated so far along with the error.
func (td *TestData) CreateKanbanBoards(kanbanBoards []board.Board) ([]ResponseResult[board.Board], error) {
	responseResults := make([]ResponseResult[board.Board], len(kanbanBoards))
	for idx, kanbanBoard := range kanbanBoards {
		responseResult, err := td.CreateKanbanBoard(kanbanBoard)
		if err != nil {
			return []ResponseResult[board.Board]{}, err
		}
		responseResults[idx] = responseResult
	}
	return responseResults, nil
}

// GetKanbanBoardById retrieves a Kanban board by its unique identifier.
// It constructs the request URL using the provided board ID and sends a GET request.
// Returns a ResponseResult containing the board.Board object and an error if the request fails.
func (td *TestData) GetKanbanBoardById(id int) (ResponseResult[board.Board], error) {
	url := fmt.Sprintf(APP_URL_TEMPLATE, td.GetAppUrl(), BOARDS_URL_CONTEXT, id)
	return MakeGetRequest[board.Board](url)
}

// UpdateKanbanBoardById updates a Kanban board identified by the given ID with the provided updatedBoard data.
// It sends a PUT request to the appropriate endpoint and returns the updated board wrapped in a ResponseResult,
// along with any error encountered during the request.
//
// Parameters:
//   - id: The unique identifier of the Kanban board to update.
//   - updatedBoard: The board.Board object containing the updated board data.
//
// Returns:
//   - ResponseResult[board.Board]: The response containing the updated board.
//   - error: Any error encountered during the update operation.
func (td *TestData) UpdateKanbanBoardById(id int, updatedBoard board.Board) (ResponseResult[board.Board], error) {
	url := fmt.Sprintf(APP_URL_TEMPLATE, td.GetAppUrl(), BOARDS_URL_CONTEXT, id)
	return MakePutRequest[board.Board, board.Board](url, updatedBoard)
}

// DeleteKanbanBoardById deletes a Kanban board identified by the given ID.
// It constructs the appropriate URL using the application URL and board context,
// then sends a DELETE request to remove the board.
// Returns a ResponseResult containing any response data and an error if the request fails.
func (td *TestData) DeleteKanbBoardById(id int) (ResponseResult[any], error) {
	url := fmt.Sprintf(APP_URL_TEMPLATE, td.GetAppUrl(), BOARDS_URL_CONTEXT, id)
	return MakeDeleteRequest[any](url)
}
