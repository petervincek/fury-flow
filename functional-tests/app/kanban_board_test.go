//go:build functional
// +build functional

package app_test

import (
	"fmt"
	"testing"

	"github.com/petervincek/fury-flow/functional-tests/utils/testdata"
	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/stretchr/testify/assert"
)

const (
	APP_URL_WITH_CONTEXT_TEMPLATE          = "http://localhost:%d%s"
	APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE = "http://localhost:%d%s/%d"
	EXPECTING_STATUS_CODE_201_MSG          = "expecting status code 201 - CREATED"
	EXPECTING_STATUS_CODE_200_MSG          = "expecting status code 200 - OK"
	EXPECTING_NON_ZERO_BOARD_ID_MSG        = "expecting non-zero boardId"
	EXPECTING_NO_ERROR_KANBAN_BOARD_MSG    = "expecting no error while creating kanban board"
)

// TestCreateKanbanBoard verifies that a new Kanban board can be successfully created via the API.
// It sets up valid test data, sends a POST request to create the board, and asserts that:
//   - No error occurs during the request.
//   - The response status code is 201 (Created).
//   - The returned board entity has a non-zero BoardId.
func TestCreateKanbanBoard(t *testing.T) {
	setupTest(t)

	// valid test data
	kanbanBoard := board.Board{
		BoardName:   "Project Hockey News",
		Description: "Kanban Board to manage the workflow/tasks/dependencies for project 'Hockey News'",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Peter",
		},
	}

	// exercise
	responseResult, err := testdata.MakePostRequest[board.Board, board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT), kanbanBoard)
	// verify
	assert.NoError(t, err, "expecting no error while making request to create new kanban board")
	assert.Equal(t, 201, responseResult.StatusCode(), EXPECTING_STATUS_CODE_201_MSG)
	assert.NotZero(t, responseResult.Entity().BoardId, EXPECTING_NON_ZERO_BOARD_ID_MSG)
}

// TestCreateKanbanBoardValidationError verifies that creating a Kanban board with missing required fields
// (specifically, an empty BoardName) results in a validation error. It checks that the API responds with
// a 400 Bad Request status code, returns an appropriate error message, and includes validation error details
// indicating the missing BoardName field.
func TestCreateKanbanBoardValidationError(t *testing.T) {
	setupTest(t)

	// invalid test data: missing required BoardName
	kanbanBoard := board.Board{
		BoardName:   "", // BoardName is required, so this should fail validation
		Description: "Kanban Board with missing name",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Peter",
		},
	}

	// exercise
	responseResult, err := testdata.MakePostRequest[board.Board, common.ErrorMsg](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT), kanbanBoard)

	// verify
	assert.NoError(t, err, "expecting no error while making request to create new kanban board with invalid data")
	assert.Equal(t, 400, responseResult.StatusCode(), "expecting status code 400 - BAD REQUEST due to validation error")
	errorEntity := responseResult.Entity()
	assert.NotNil(t, errorEntity, "expecting error response entity")
	assert.Contains(t, errorEntity.Message, "Kanban board validation error", "expecting error message to mention missing BoardName")
	assert.Equal(t, errorEntity.Errors, []string{"Field 'BoardName' failed on the 'required' tag"}, "expecting validation error details")
}

// TestGetKanbanBoardByExistingId verifies that a Kanban board can be successfully retrieved by its existing ID.
// The test performs the following steps:
// 1. Sets up the test environment.
// 2. Creates a test Kanban board using a helper.
// 3. Asserts that the board is created without errors and has a valid non-zero ID.
// 4. Retrieves the board by its ID via a GET request.
// 5. Asserts that the retrieval is successful and the returned board's fields match the original input.
func TestGetKanbanBoardByExistingId(t *testing.T) {
	setupTest(t)

	// create test kanban board
	kanbanBoard := board.Board{
		BoardName:   "Test Board",
		Description: "Test Kanban Board for retrieval by ID",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Tester",
		},
	}

	// create it with helper TestDataCreator
	createResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_MSG)
	assert.Equal(t, 201, createResp.StatusCode(), EXPECTING_STATUS_CODE_201_MSG)
	boardId := createResp.Entity().BoardId
	assert.NotZero(t, boardId, EXPECTING_NON_ZERO_BOARD_ID_MSG)

	// exercise: get the board by ID
	getResp, err := testdata.MakeGetRequest[board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, boardId))
	assert.NoError(t, err, "expecting no error while getting kanban board by ID")
	assert.Equal(t, 200, getResp.StatusCode(), EXPECTING_STATUS_CODE_200_MSG)
	assert.Equal(t, boardId, getResp.Entity().BoardId, "expecting boardId to match")
	assert.Equal(t, kanbanBoard.BoardName, getResp.Entity().BoardName, "expecting board name to match")
	assert.Equal(t, kanbanBoard.Description, getResp.Entity().Description, "expecting description to match")
}

// TestGetKanbanBoardByNonExistingId verifies that attempting to retrieve a kanban board
// using a non-existing board ID returns a 404 Not Found status code. It ensures that
// the API correctly handles requests for resources that do not exist.
func TestGetKanbanBoardByNonExistingId(t *testing.T) {
	setupTest(t)

	// Use a boardId that is unlikely to exist (e.g., a very large number)
	nonExistingBoardId := int64(99999999)

	// exercise: attempt to get the non-existing board
	getResp, err := testdata.MakeGetRequest[board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, nonExistingBoardId))
	assert.NoError(t, err, "expecting no error while attempting to get non-existing kanban board")
	assert.Equal(t, 404, getResp.StatusCode(), "expecting status code 404 - NOT FOUND for non-existing board")
}

// TestGetAllKanbanBoards verifies that multiple kanban boards can be created and subsequently
// retrieved via the API. It creates several boards, sends a GET request to fetch all boards,
// and asserts that all created boards are present in the response.
func TestGetAllKanbanBoards(t *testing.T) {
	setupTest(t)

	// Create several kanban boards
	boardsToCreate := []board.Board{
		{
			BoardName:   "Board One",
			Description: "First test board",
			AuditableFields: common.AuditableFields{
				CreatedBy: "User1",
			},
		},
		{
			BoardName:   "Board Two",
			Description: "Second test board",
			AuditableFields: common.AuditableFields{
				CreatedBy: "User2",
			},
		},
		{
			BoardName:   "Board Three",
			Description: "Third test board",
			AuditableFields: common.AuditableFields{
				CreatedBy: "User3",
			},
		},
	}

	responseResults, err := TestDataCreator.CreateKanbanBoards(boardsToCreate)
	assert.NoError(t, err, "expecting no error while creating kanban boards")

	// Exercise: get all kanban boards
	getAllResp, err := testdata.MakeGetRequest[[]board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT))
	assert.NoError(t, err, "expecting no error while getting all kanban boards")
	assert.Equal(t, 200, getAllResp.StatusCode(), EXPECTING_STATUS_CODE_200_MSG)

	// Verify: check that all created boards are present in the response
	allBoards := getAllResp.Entity()
	found := 0
	for _, responseResult := range responseResults {
		for _, b := range *allBoards {
			if b.BoardId == responseResult.Entity().BoardId {
				found++
				break
			}
		}
	}
	assert.Equal(t, len(boardsToCreate), found, "expecting all created boards to be present in the response")
}

// TestGetAllKanbanBoardsEmptyDatabase verifies that retrieving all kanban boards from an empty database
// returns a successful response with an empty list of boards. It ensures no error occurs during the request,
// the response status code is 200, and the returned entity is not nil but contains zero boards.
func TestGetAllKanbanBoardsEmptyDatabase(t *testing.T) {
	setupTest(t)

	// Ensure the database is empty (no boards created)
	// Exercise: get all kanban boards
	getAllResp, err := testdata.MakeGetRequest[[]board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT))
	// verify
	assert.NoError(t, err, "expecting no error while getting all kanban boards from empty database")
	assert.Equal(t, 200, getAllResp.StatusCode(), EXPECTING_STATUS_CODE_200_MSG)
	allBoards := getAllResp.Entity()
	assert.NotNil(t, allBoards, "expecting response entity to be not nil")
	assert.Equal(t, 0, len(*allBoards), "expecting zero boards in response when database is empty")
}

// TestUpdateExistingKanbanBoard verifies that an existing Kanban board can be updated successfully.
// The test performs the following steps:
// 1. Creates a new Kanban board using the TestDataCreator helper.
// 2. Updates the created board with new name and description via a PUT request.
// 3. Retrieves the updated board and asserts that the changes are reflected correctly.
// The test asserts proper status codes and checks for errors at each step.
func TestUpdateExistingKanbanBoard(t *testing.T) {
	setupTest(t)

	// create test kanban board
	kanbanBoard := board.Board{
		BoardName:   "Update Test Board",
		Description: "Kanban Board for update test",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Updater",
		},
	}

	// create it with helper TestDataCreator
	createResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_MSG)
	assert.Equal(t, 201, createResp.StatusCode(), EXPECTING_STATUS_CODE_201_MSG)
	boardId := createResp.Entity().BoardId
	assert.NotZero(t, boardId, EXPECTING_NON_ZERO_BOARD_ID_MSG)

	// prepare updated board data
	updatedBoard := board.Board{
		BoardId:     boardId,
		BoardName:   "Updated Board Name",
		Description: "Updated description for the board",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Updater",
		},
	}

	// exercise: update the board by ID
	updateResp, err := testdata.MakePutRequest[board.Board, struct{}](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, boardId), updatedBoard)
	assert.NoError(t, err, "expecting no error while updating kanban board")
	assert.Equal(t, 200, updateResp.StatusCode(), EXPECTING_STATUS_CODE_200_MSG)

	// verify: get the updated board and check fields
	getResp, err := testdata.MakeGetRequest[board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, boardId))
	assert.NoError(t, err, "expecting no error while getting updated kanban board by ID")
	assert.Equal(t, 200, getResp.StatusCode(), "expecting status code 200 - OK after update")
	assert.Equal(t, updatedBoard.BoardName, getResp.Entity().BoardName, "expecting updated board name")
	assert.Equal(t, updatedBoard.Description, getResp.Entity().Description, "expecting updated description")
}

// TestUpdateNonExistingKanbanBoard verifies that attempting to update a kanban board
// with a non-existent board ID returns a 404 Not Found status code. It prepares an
// updated board payload with an unlikely board ID, sends a PUT request to update the board,
// and asserts that no error occurs during the request and that the response status code is 404.
func TestUpdateNonExistingKanbanBoard(t *testing.T) {
	setupTest(t)

	// Use a boardId that is unlikely to exist
	nonExistingBoardId := 9999

	// prepare updated board data
	updatedBoard := board.Board{
		BoardId:     nonExistingBoardId,
		BoardName:   "Non-Existing Board",
		Description: "Trying to update a board that does not exist",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Updater",
		},
	}

	// exercise: attempt to update the non-existing board
	updateResp, err := testdata.MakePutRequest[board.Board, board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, nonExistingBoardId), updatedBoard)
	assert.NoError(t, err, "expecting no error while attempting to update non-existing kanban board")
	assert.Equal(t, 404, updateResp.StatusCode(), "expecting status code 404 - NOT FOUND for non-existing board")
}

// TestDeleteExistingKanbanBoard verifies that a kanban board can be successfully deleted.
// The test performs the following steps:
// 1. Creates a new kanban board using the TestDataCreator helper.
// 2. Deletes the created board by its ID.
// 3. Asserts that the deletion returns a 204 No Content status code.
// 4. Attempts to retrieve the deleted board and expects a 404 Not Found status code.
// This ensures that the delete operation removes the board and subsequent retrieval fails as expected.
func TestDeleteExistingKanbanBoard(t *testing.T) {
	setupTest(t)

	// create test kanban board
	kanbanBoard := board.Board{
		BoardName:   "Delete Test Board",
		Description: "Kanban Board for delete test",
		AuditableFields: common.AuditableFields{
			CreatedBy: "Deleter",
		},
	}

	// create it with helper TestDataCreator
	createResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_MSG)
	assert.Equal(t, 201, createResp.StatusCode(), EXPECTING_STATUS_CODE_201_MSG)
	boardId := createResp.Entity().BoardId
	assert.NotZero(t, boardId, EXPECTING_NON_ZERO_BOARD_ID_MSG)

	// exercise: delete the board by ID
	deleteResp, err := testdata.MakeDeleteRequest[struct{}](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, boardId))
	assert.NoError(t, err, "expecting no error while deleting kanban board")
	assert.Equal(t, 204, deleteResp.StatusCode(), "expecting status code 204 - NO CONTENT")

	// verify: try to get the deleted board
	getResp, err := testdata.MakeGetRequest[board.Board](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, boardId))
	assert.NoError(t, err, "expecting no error while getting kanban board by ID")
	assert.Equal(t, 404, getResp.StatusCode(), "expecting status code 404 - NOT FOUND after deletion")
}

// TestDeleteNonExistingKanbanBoard verifies that attempting to delete a kanban board
// with a non-existing board ID returns a 204 NO CONTENT status code and does not produce an error.
// This ensures the delete endpoint behaves idempotently when the resource does not exist.
func TestDeleteNonExistingKanbanBoard(t *testing.T) {
	setupTest(t)

	// Use a boardId that is unlikely to exist (e.g., a very large number)
	nonExistingBoardId := int64(99999999)

	// exercise: attempt to delete the non-existing board
	deleteResp, err := testdata.MakeDeleteRequest[struct{}](
		fmt.Sprintf(APP_URL_WITH_CONTEXT_BOARD_ID_TEMPLATE, TEST_APP_SERVER_PORT, testdata.BOARDS_URL_CONTEXT, nonExistingBoardId))
	assert.NoError(t, err, "expecting no error while attempting to delete non-existing kanban board")
	assert.Equal(t, 204, deleteResp.StatusCode(), "expecting status code 204 - NO CONTENT for non-existing board")
}
