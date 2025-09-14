//go:build functional
// +build functional

package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/petervincek/fury-flow/functional-tests/utils/testdata"
	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/card"
	"github.com/petervincek/fury-flow/internal/api/comment"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/stretchr/testify/assert"
)

// createBoardAndCard creates a new Kanban board and a card associated with it for testing purposes.
// It returns the IDs of the created board and card.
// The function asserts that all creation steps succeed and that the card creation returns a 201 status code.
//
// Parameters:
//
//	t - The testing context.
//
// Returns:
//
//	boardId - The ID of the created Kanban board.
//	cardId  - The ID of the created Kanban card.
func createBoardAndCard(t *testing.T) (int, int) {
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Comment Test Board").
		SetDescription("Board for comment tests").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	kanbanCard, err := card.NewBuilder().
		SetBoardId(boardId).
		SetTitle("Comment Test Card").
		SetDescription("desc").
		SetStatus(card.StatusTodo).
		SetAcceptanceCriteria("criteria").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityMedium).
		SetStoryPoints(1).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	cardResp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
	assert.NoError(t, err)
	assert.Equal(t, 201, cardResp.StatusCode())
	cardId := cardResp.Entity().CardId

	return boardId, cardId
}

// TestGetKanbanCommentsForInvalidBoardId verifies that requesting kanban comments with an invalid board ID
// returns an appropriate error response (HTTP 400 or 404). It sets up a test environment, creates a valid card,
// and then attempts to fetch comments using a non-numeric board ID, asserting that the response status code
// indicates a client or not found error.
func TestGetKanbanCommentsForInvalidBoardId(t *testing.T) {
	setupTest(t)
	_, cardId := createBoardAndCard(t)
	invalidBoardId := NOT_A_NUMBER

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%s/cards/%d/comments", TestDataCreator.GetAppUrl(), invalidBoardId, cardId))
	assert.NoError(t, err)
	assert.True(t, getResp.StatusCode() == 400 || getResp.StatusCode() == 404)
}

// TestGetKanbanCommentsForInvalidCardId verifies that requesting comments for a card with an invalid ID
// returns an appropriate error response (HTTP 400 or 404). It sets up a test board and card, then attempts
// to fetch comments using a non-numeric card ID, asserting that the response status code indicates a client error.
func TestGetKanbanCommentsForInvalidCardId(t *testing.T) {
	setupTest(t)
	boardId, _ := createBoardAndCard(t)
	invalidCardId := NOT_A_NUMBER

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%d/cards/%s/comments", TestDataCreator.GetAppUrl(), boardId, invalidCardId))
	assert.NoError(t, err)
	assert.True(t, getResp.StatusCode() == 400 || getResp.StatusCode() == 404)
}

// TestGetKanbanCommentsForNonExistingKanbanBoard verifies that attempting to retrieve comments for a card
// on a non-existing Kanban board returns a 404 Not Found status code. It sets up the test environment,
// creates a valid card, and then sends a GET request using an invalid board ID, asserting that the response
// status code is 404 and no error occurred during the request.
func TestGetKanbanCommentsForNonExistingKanbanBoard(t *testing.T) {
	setupTest(t)
	_, cardId := createBoardAndCard(t)
	nonExistingBoardId := 999999

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestGetKanbanCommentsForNonExistingKanbanCard verifies that requesting comments for a non-existing Kanban card
// returns a 404 Not Found status code. It sets up a test board and card, then attempts to retrieve comments for
// a card ID that does not exist, asserting that the response status code is 404.
func TestGetKanbanCommentsForNonExistingKanbanCard(t *testing.T) {
	setupTest(t)
	boardId, _ := createBoardAndCard(t)
	nonExistingCardId := 999999

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId))
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestGetKanbanCommentsForCardWithNoComments verifies that retrieving comments for a Kanban card
// with no comments returns an empty list and a 200 OK status code.
func TestGetKanbanCommentsForCardWithNoComments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	getResp, err := testdata.MakeGetRequest[[]comment.Comment](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	assert.Len(t, *getResp.Entity(), 0)
}

// TestGetKanbanCommentsForCardWithMultipleComments verifies that multiple comments can be added to a kanban card
// and that all added comments are correctly retrieved via the API. It creates a board and a card, adds several
// comments to the card, and then fetches the comments to ensure they match the expected values.
func TestGetKanbanCommentsForCardWithMultipleComments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	commentTexts := []string{"Alpha", "Beta", "Gamma"}
	for _, text := range commentTexts {
		kanbanComment, err := comment.NewBuilder().
			SetCardId(cardId).
			SetCommentText(text).
			SetCommentBy("Peter").
			SetCreatedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanComment(boardId, cardId, kanbanComment)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
	}

	getResp, err := testdata.MakeGetRequest[[]comment.Comment](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	comments := *getResp.Entity()
	assert.Len(t, comments, len(commentTexts))
	foundTexts := make(map[string]bool)
	for _, c := range comments {
		foundTexts[c.CommentText] = true
	}
	for _, text := range commentTexts {
		assert.True(t, foundTexts[text], "expecting comment with text %s to be present", text)
	}
}

// TestTryToCreateKanbanCommentWithInvalidBoardId verifies that attempting to create a kanban comment
// with an invalid board ID (non-numeric value) results in an appropriate error response from the API.
// The test expects the API to return either a 400 Bad Request or 404 Not Found status code.
func TestTryToCreateKanbanCommentWithInvalidBoardId(t *testing.T) {
	setupTest(t)
	_, cardId := createBoardAndCard(t)
	invalidBoardId := NOT_A_NUMBER

	kanbanComment, err := comment.NewBuilder().
		SetCardId(cardId).
		SetCommentText("Invalid board id test").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[comment.Comment, common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%s/cards/%d/comments", TestDataCreator.GetAppUrl(), invalidBoardId, cardId), kanbanComment)
	assert.NoError(t, err)
	assert.True(t, resp.StatusCode() == 400 || resp.StatusCode() == 404)
}

// TestTryToCreateKanbanCommentWithInvalidCardId verifies that attempting to create a kanban comment
// with an invalid card ID (non-numeric value) results in an appropriate error response from the API.
// The test expects the response status code to be either 400 (Bad Request) or 404 (Not Found).
func TestTryToCreateKanbanCommentWithInvalidCardId(t *testing.T) {
	setupTest(t)
	boardId, _ := createBoardAndCard(t)
	invalidCardId := NOT_A_NUMBER

	kanbanComment, err := comment.NewBuilder().
		SetCommentText("Invalid card id test").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[comment.Comment, common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%d/cards/%s/comments", TestDataCreator.GetAppUrl(), boardId, invalidCardId), kanbanComment)
	assert.NoError(t, err)
	assert.True(t, resp.StatusCode() == 400 || resp.StatusCode() == 404)
}

// TestTryToCreateKanbanCommentForNonExistingBoard verifies that attempting to create a kanban comment
// for a non-existing board returns a 404 Not Found status code. It sets up a test environment, creates
// a card, and then tries to post a comment to a board ID that does not exist, asserting that the
// response status code is 404.
func TestTryToCreateKanbanCommentForNonExistingBoard(t *testing.T) {
	setupTest(t)
	_, cardId := createBoardAndCard(t)
	nonExistingBoardId := 999999

	kanbanComment, err := comment.NewBuilder().
		SetCardId(cardId).
		SetCommentText("Non-existing board test").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[comment.Comment, common.ErrorMsg](fmt.Sprintf(
		testdata.POST_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId), kanbanComment)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode())
}

// TestTryToCreateKanbanCommentForNonExistingCard verifies that attempting to create a comment
// for a non-existing Kanban card returns a 404 Not Found status code.
// It sets up the test environment, creates a board and a valid card, then tries to create
// a comment for a card ID that does not exist. The test asserts that no error occurs during
// comment building and request execution, and that the response status code is 404.
func TestTryToCreateKanbanCommentForNonExistingCard(t *testing.T) {
	setupTest(t)
	boardId, _ := createBoardAndCard(t)
	nonExistingCardId := 999999

	kanbanComment, err := comment.NewBuilder().
		SetCardId(nonExistingCardId).
		SetCommentText("Non-existing card test").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[comment.Comment, common.ErrorMsg](fmt.Sprintf(
		testdata.POST_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId), kanbanComment)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode())
}

// TestCreateKanbanComment verifies that a kanban comment can be successfully created via the API.
// It sets up the test environment, creates a board and card, constructs a new comment, and sends a POST request.
// The test asserts that the request completes without error, returns a 201 status code, and that the response contains
// the expected comment data including a non-zero CommentId and matching CommentText.
func TestCreateKanbanComment(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	kanbanComment, err := comment.NewBuilder().
		SetCardId(cardId).
		SetCommentText("This is a test comment").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[comment.Comment, comment.Comment](fmt.Sprintf(
		testdata.POST_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId), kanbanComment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	assert.NotZero(t, resp.Entity().CommentId)
	assert.Equal(t, kanbanComment.CommentText, resp.Entity().CommentText)
}

// TestCreateKanbanCommentValidationError verifies that creating a kanban comment with missing required fields
// returns a validation error. It ensures that the API responds with a 400 status code when the comment text
// and commenter fields are empty, indicating proper validation handling for invalid input.
func TestCreateKanbanCommentValidationError(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	// Missing required fields
	invalidComment := comment.Comment{
		CardId:      cardId,
		CommentText: "",
		CommentBy:   "",
	}
	resp, err := testdata.MakePostRequest[comment.Comment, comment.Comment](fmt.Sprintf(
		testdata.POST_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId), invalidComment)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
	assert.Contains(t, resp.Entity().CommentText, "")
}

// TestGetKanbanCardComments verifies that multiple comments can be created for a Kanban card
// and that all created comments can be retrieved successfully. It checks that the correct
// number of comments is returned and that each comment's text matches the expected values.
func TestGetKanbanCardComments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	// Create multiple comments
	commentTexts := []string{"First comment", "Second comment", "Third comment"}
	for _, text := range commentTexts {
		kanbanComment, err := comment.NewBuilder().
			SetCardId(cardId).
			SetCommentText(text).
			SetCommentBy("Peter").
			SetCreatedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanComment(boardId, cardId, kanbanComment)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
	}

	// Get all comments
	getResp, err := testdata.MakeGetRequest[[]comment.Comment](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_COMMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	// getResp, err := TestDataCreator.GetKanbanCardComments(boardId, cardId)
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	comments := *getResp.Entity()
	assert.Len(t, comments, len(commentTexts))
	foundTexts := make(map[string]bool)
	for _, c := range comments {
		foundTexts[c.CommentText] = true
	}
	for _, text := range commentTexts {
		assert.True(t, foundTexts[text], "expecting comment with text %s to be present", text)
	}
}

// TestGetKanbanCardCommentById verifies that a Kanban card comment can be created and retrieved by its ID.
// The test performs the following steps:
// 1. Sets up the test environment.
// 2. Creates a new board and card.
// 3. Builds a unique comment for the card.
// 4. Creates the comment via the API and checks for successful creation.
// 5. Retrieves the comment by its ID and asserts that the returned comment matches the expected text.
func TestGetKanbanCardCommentById(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	kanbanComment, err := comment.NewBuilder().
		SetCardId(cardId).
		SetCommentText("Unique comment").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := TestDataCreator.CreateKanbanComment(boardId, cardId, kanbanComment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	commentId := resp.Entity().CommentId

	getResp, err := testdata.MakeGetRequest[comment.Comment](fmt.Sprintf(
		testdata.GET_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId, commentId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	assert.Equal(t, "Unique comment", getResp.Entity().CommentText)
}

// TestGetKanbanCardCommentByNonExistingId verifies that attempting to retrieve a kanban card comment
// using a non-existing comment ID returns a 404 Not Found status code. It sets up the test environment,
// creates a board and card, and then sends a GET request for a comment ID that does not exist.
func TestGetKanbanCardCommentByNonExistingId(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	getResp, err := testdata.MakeGetRequest[comment.Comment](fmt.Sprintf(
		testdata.GET_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId, 9999))
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestDeleteKanbanCardComment verifies that a kanban card comment can be successfully deleted.
// The test performs the following steps:
// 1. Sets up the test environment and creates a board and card.
// 2. Adds a new comment to the card.
// 3. Deletes the created comment.
// 4. Confirms that the deletion was successful by checking for a 204 status code.
// 5. Attempts to retrieve the deleted comment and expects a 404 status code, confirming the comment no longer exists.
func TestDeleteKanbanCardComment(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	kanbanComment, err := comment.NewBuilder().
		SetCardId(cardId).
		SetCommentText("Comment to delete").
		SetCommentBy("Peter").
		SetCreatedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := TestDataCreator.CreateKanbanComment(boardId, cardId, kanbanComment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	commentId := resp.Entity().CommentId

	deleteResp, err := testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
		testdata.DELETE_KANBAN_COMMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId, commentId))
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteResp.StatusCode())

	// Verify deletion
	getResp, err := TestDataCreator.GetKanbanCardComment(boardId, cardId, commentId)
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestDeleteAllKanbanCardComments verifies that all comments associated with a Kanban card
// can be successfully deleted. It creates multiple comments for a card, deletes them using
// the appropriate API endpoint, and then checks that no comments remain for the card.
func TestDeleteAllKanbanCardComments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	// Create multiple comments
	for i := 0; i < 3; i++ {
		kanbanComment, err := comment.NewBuilder().
			SetCardId(cardId).
			SetCommentText(fmt.Sprintf("Comment %d", i+1)).
			SetCommentBy("Peter").
			SetCreatedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanComment(boardId, cardId, kanbanComment)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
	}

	// Delete all comments
	deleteResp, err := testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
		testdata.DELETE_KANBAN_COMMENTS_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteResp.StatusCode())

	// Verify all comments are deleted
	getResp, err := TestDataCreator.GetKanbanCardComments(boardId, cardId)
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	assert.Len(t, *getResp.Entity(), 0)
}

// TestTryToDeleteKanbanCommentWithInvalidIds verifies that attempting to delete kanban comments
// with invalid board or card IDs results in a 400 Bad Request response. It covers scenarios where
// either the board ID or card ID is not a valid number, for both single comment deletion and
// deletion of all comments for a card. The test ensures that the API correctly handles invalid
// path parameters by returning the appropriate error status code.
func TestTryToDeleteKanbanCommentWithInvalidIds(t *testing.T) {
	testCases := []struct {
		name     string
		boardId  string
		cardId   string
		exercise func(boardId string, cardId string, commentId string) (testdata.ResponseResult[struct{}], error)
	}{
		{
			name:    "single comment delete with invalid board id",
			boardId: NOT_A_NUMBER,
			cardId:  "1",
			exercise: func(boardId, cardId, commentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf("%s/boards/%s/cards/%s/comments/%s", TestDataCreator.GetAppUrl(), boardId, cardId, commentId))
			},
		},
		{
			name:    "all comments for card delete with invalid board id",
			boardId: NOT_A_NUMBER,
			cardId:  "1",
			exercise: func(boardId, cardId, commentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf("%s/boards/%s/cards/%s/comments", TestDataCreator.GetAppUrl(), boardId, cardId))
			},
		},
		{
			name:    "single comment delete with invalid card id",
			boardId: "1",
			cardId:  NOT_A_NUMBER,
			exercise: func(boardId, cardId, commentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf("%s/boards/%s/cards/%s/comments/%s", TestDataCreator.GetAppUrl(), boardId, cardId, commentId))
			},
		},
		{
			name:    "all comments for card delete with invalid card id",
			boardId: "1",
			cardId:  NOT_A_NUMBER,
			exercise: func(boardId, cardId, commentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf("%s/boards/%s/cards/%s/comments", TestDataCreator.GetAppUrl(), boardId, cardId))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			setupTest(t)

			deleteResp, err := testCase.exercise(testCase.boardId, testCase.cardId, "1")
			assert.NoError(t, err)
			assert.True(t, deleteResp.StatusCode() == 400)
		})
	}
}

// TestTryToDeleteKanbanCommentForNonExistingBoard verifies that attempting to delete a kanban comment
// (either a single comment or all comments for a card) on a non-existing board returns a 404 Not Found status.
// It tests both the deletion of a specific comment and the deletion of all comments for a card using invalid board IDs.
func TestTryToDeleteKanbanCommentForNonExistingBoard(t *testing.T) {
	boardId := 999999 // non-existing board ID
	cardId := 1       // arbitrary card ID
	commentId := 1    // arbitrary comment ID

	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "single comment delete for non-existing board",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/comments/%d", TestDataCreator.GetAppUrl(), boardId, cardId, commentId),
		},
		{
			name: "all comments for card delete for non-existing board",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/comments", TestDataCreator.GetAppUrl(), boardId, cardId),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTest(t)
			deleteResp, err := testdata.MakeDeleteRequest[struct{}](tc.url)
			assert.NoError(t, err)
			assert.Equal(t, 404, deleteResp.StatusCode())
		})
	}
}

// TestTryToDeleteKanbanCommentForNonExistingCard verifies that attempting to delete a comment
// (either a single comment or all comments) from a non-existing Kanban card returns a 404 Not Found status.
// It tests both the endpoint for deleting a specific comment and for deleting all comments on a card,
// ensuring proper error handling for invalid card IDs.
func TestTryToDeleteKanbanCommentForNonExistingCard(t *testing.T) {
	setupTest(t)
	boardId, _ := createBoardAndCard(t)
	nonExistingCardId := 999999
	commentId := 1 // arbitrary comment ID

	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "single comment delete for non-existing card",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/comments/%d", TestDataCreator.GetAppUrl(), boardId, nonExistingCardId, commentId),
		},
		{
			name: "all comments for card delete for non-existing card",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/comments", TestDataCreator.GetAppUrl(), boardId, nonExistingCardId),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteResp, err := testdata.MakeDeleteRequest[struct{}](tc.url)
			assert.NoError(t, err)
			assert.Equal(t, 404, deleteResp.StatusCode())
		})
	}
}
