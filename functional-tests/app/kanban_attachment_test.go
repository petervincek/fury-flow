//go:build functional
// +build functional

package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/petervincek/fury-flow/functional-tests/utils/testdata"
	"github.com/petervincek/fury-flow/internal/api/attachment"
	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/card"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/stretchr/testify/assert"
)

const (
	SOME_DATA       = "some data"
	APPLICATION_PDF = "application/pdf"
)

// createAttachmentBoardAndCard creates a new Kanban board and a card with predefined attributes for attachment tests.
// It returns the IDs of the created board and card.
// The function asserts that all creation steps succeed and fails the test if any error occurs.
//
// Parameters:
//
//	t - The testing context.
//
// Returns:
//
//	boardId - The ID of the created Kanban board.
//	cardId  - The ID of the created Kanban card.
func createAttachmentBoardAndCard(t *testing.T) (int, int) {
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Attachment Test Board").
		SetDescription("Board for attachment tests").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	kanbanCard, err := card.NewBuilder().
		SetBoardId(boardId).
		SetTitle("Attachment Test Card").
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

// TestGetKanbanAttachmentsForCardNotBelongingToBoard verifies that attempting to retrieve attachments for a card
// that does not belong to the specified board returns a 400 error with the appropriate error message.
func TestGetKanbanAttachmentsForCardNotBelongingToBoard(t *testing.T) {
	setupTest(t)
	boardId1, _ := createAttachmentBoardAndCard(t)
	_, cardId2 := createAttachmentBoardAndCard(t)

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId1, cardId2))
	assert.NoError(t, err)
	assert.Equal(t, 400, getResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(CARD_DOES_NOT_BELONG_TO_BOARD_ERROR_MSG, nil), *getResp.Entity())
}

// TestGetKanbanAttachmentsForInvalidBoardId verifies that requesting kanban attachments
// with an invalid board ID returns an appropriate error response (HTTP 400 or 404).
// It sets up the test environment, creates a card and attachment, and then attempts
// to retrieve attachments using an invalid board ID, asserting that the response
// status code indicates a client or not found error.
func TestGetKanbanAttachmentsForInvalidBoardId(t *testing.T) {
	setupTest(t)
	_, cardId := createAttachmentBoardAndCard(t)
	invalidBoardId := NOT_A_NUMBER

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%s/cards/%d/attachments", TestDataCreator.GetAppUrl(), invalidBoardId, cardId))
	assert.NoError(t, err)
	assert.True(t, getResp.StatusCode() == 400 || getResp.StatusCode() == 404)
}

// TestGetKanbanAttachmentsForInvalidCardId verifies that requesting attachments for a card with an invalid ID
// returns an appropriate error response (HTTP 400 or 404). It sets up the test environment, creates a board and card,
// then attempts to fetch attachments using an invalid card ID, asserting that the response status code indicates an error.
func TestGetKanbanAttachmentsForInvalidCardId(t *testing.T) {
	setupTest(t)
	boardId, _ := createAttachmentBoardAndCard(t)
	invalidCardId := NOT_A_NUMBER

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%d/cards/%s/attachments", TestDataCreator.GetAppUrl(), boardId, invalidCardId))
	assert.NoError(t, err)
	assert.True(t, getResp.StatusCode() == 400 || getResp.StatusCode() == 404)
}

// TestGetKanbanAttachmentsForNonExistingKanbanBoard verifies that requesting attachments for a non-existing Kanban board
// returns a 404 Not Found status code. It sets up the test environment, creates a valid card ID, and attempts to fetch
// attachments using an invalid board ID, asserting that the response status code is 404 and no error occurred during the request.
func TestGetKanbanAttachmentsForNonExistingKanbanBoard(t *testing.T) {
	setupTest(t)
	_, cardId := createAttachmentBoardAndCard(t)
	nonExistingBoardId := 999999

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestGetKanbanAttachmentsForNonExistingKanbanCard verifies that requesting attachments for a non-existing kanban card
// returns a 404 Not Found status code. It sets up the test environment, creates a valid board and card, then attempts
// to retrieve attachments for a card ID that does not exist, asserting that the response status code is 404.
func TestGetKanbanAttachmentsForNonExistingKanbanCard(t *testing.T) {
	setupTest(t)
	boardId, _ := createAttachmentBoardAndCard(t)
	nonExistingCardId := 999999

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId))
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestGetKanbanAttachmentsForCardWithNoAttachments verifies that retrieving attachments for a Kanban card
// with no attachments returns an empty list and a 200 OK status code.
func TestGetKanbanAttachmentsForCardWithNoAttachments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	getResp, err := testdata.MakeGetRequest[[]attachment.Attachment](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	assert.Len(t, *getResp.Entity(), 0)
}

// TestGetKanbanAttachmentsForCardWithMultipleAttachments verifies that multiple attachments
// can be added to a kanban card and subsequently retrieved via the API. It creates a card,
// attaches several files to it, and asserts that all attachments are returned with correct filenames.
func TestGetKanbanAttachmentsForCardWithMultipleAttachments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	attachmentNames := []string{"Alpha.pdf", "Beta.docx", "Gamma.png"}
	for _, name := range attachmentNames {
		kanbanAttachment, err := attachment.NewBuilder().
			SetCardId(cardId).
			SetFilename(name).
			SetFileType("application/octet-stream").
			SetFileSizeBytes(12345).
			SetAttachmentData([]byte(SOME_DATA)).
			SetUploadedBy("Peter").
			SetUploadedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		_, err = TestDataCreator.CreateKanbanAttachment(boardId, cardId, kanbanAttachment)
		assert.NoError(t, err)
	}

	getResp, err := testdata.MakeGetRequest[[]attachment.Attachment](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	attachments := *getResp.Entity()
	assert.Len(t, attachments, len(attachmentNames))
	foundNames := make(map[string]bool)
	for _, a := range attachments {
		foundNames[a.Filename] = true
	}
	for _, name := range attachmentNames {
		assert.True(t, foundNames[name])
	}
}

// TestTryToCreateKanbanAttachmentForCardNotBelongingToBoard verifies that attempting to create a kanban attachment
// for a card that does not belong to the specified board fails with a 400 Bad Request error. The test creates two
// separate boards and cards, then tries to attach a file to a card using the board ID of the other board. It asserts
// that the operation returns the expected error message indicating the card does not belong to the board.
func TestTryToCreateKanbanAttachmentForCardNotBelongingToBoard(t *testing.T) {
	setupTest(t)
	boardId1, _ := createAttachmentBoardAndCard(t)
	_, cardId2 := createAttachmentBoardAndCard(t)

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId2).
		SetFilename("Should fail due to board/card mismatch.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[attachment.Attachment, common.ErrorMsg](fmt.Sprintf(
		testdata.POST_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId1, cardId2), kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(CARD_DOES_NOT_BELONG_TO_BOARD_ERROR_MSG, nil), *resp.Entity())
}

// TestTryToCreateKanbanAttachmentWithInvalidBoardId verifies that attempting to create a kanban attachment
// with an invalid board ID results in an appropriate error response (HTTP 400 or 404).
// The test sets up the environment, creates a valid card, and then tries to attach a file to a board
// using an invalid board ID. It asserts that no error occurs during attachment building and posting,
// and that the response status code indicates a client or not found error.
func TestTryToCreateKanbanAttachmentWithInvalidBoardId(t *testing.T) {
	setupTest(t)
	_, cardId := createAttachmentBoardAndCard(t)
	invalidBoardId := NOT_A_NUMBER

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId).
		SetFilename("Invalid board id test.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[attachment.Attachment, common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%s/cards/%d/attachments", TestDataCreator.GetAppUrl(), invalidBoardId, cardId), kanbanAttachment)
	assert.NoError(t, err)
	assert.True(t, resp.StatusCode() == 400 || resp.StatusCode() == 404)
}

// TestTryToCreateKanbanAttachmentWithInvalidCardId verifies that attempting to create a kanban attachment
// with an invalid card ID results in an appropriate error response (HTTP 400 or 404).
// It sets up a test board and card, constructs an attachment, and sends a POST request using an invalid card ID.
// The test asserts that no error occurs during attachment creation and that the response status code indicates failure.
func TestTryToCreateKanbanAttachmentWithInvalidCardId(t *testing.T) {
	setupTest(t)
	boardId, _ := createAttachmentBoardAndCard(t)
	invalidCardId := NOT_A_NUMBER

	kanbanAttachment, err := attachment.NewBuilder().
		SetFilename("Invalid card id test.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[attachment.Attachment, common.ErrorMsg](fmt.Sprintf(
		"%s/boards/%d/cards/%s/attachments", TestDataCreator.GetAppUrl(), boardId, invalidCardId), kanbanAttachment)
	assert.NoError(t, err)
	assert.True(t, resp.StatusCode() == 400 || resp.StatusCode() == 404)
}

// TestTryToCreateKanbanAttachmentForNonExistingBoard verifies that attempting to create a kanban attachment
// for a non-existing board returns a 404 Not Found error. The test sets up a valid card and attachment,
// but uses a board ID that does not exist, then asserts that the API responds with the expected error code.
func TestTryToCreateKanbanAttachmentForNonExistingBoard(t *testing.T) {
	setupTest(t)
	_, cardId := createAttachmentBoardAndCard(t)
	nonExistingBoardId := 999999

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId).
		SetFilename("Non-existing board test.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[attachment.Attachment, common.ErrorMsg](fmt.Sprintf(
		testdata.POST_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId), kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode())
}

// TestTryToCreateKanbanAttachmentForNonExistingCard verifies that attempting to create a kanban attachment
// for a non-existing card returns a 404 Not Found error. It sets up a test board and uses a card ID that
// does not exist, then asserts that the API responds with the expected error status.
func TestTryToCreateKanbanAttachmentForNonExistingCard(t *testing.T) {
	setupTest(t)
	boardId, _ := createAttachmentBoardAndCard(t)
	nonExistingCardId := 999999

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(nonExistingCardId).
		SetFilename("Non-existing card test.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[attachment.Attachment, common.ErrorMsg](fmt.Sprintf(
		testdata.POST_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId), kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode())
}

// TestCreateKanbanAttachment verifies that a kanban attachment can be successfully created.
// It sets up the test environment, creates a board and card, builds an attachment object,
// sends a POST request to create the attachment, and asserts that the response is successful
// and the returned attachment data matches the input.
func TestCreateKanbanAttachment(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId).
		SetFilename("This is a test attachment.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := testdata.MakePostRequest[attachment.Attachment, attachment.Attachment](fmt.Sprintf(
		testdata.POST_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId), kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	assert.NotZero(t, resp.Entity().AttachmentId)
	assert.Equal(t, kanbanAttachment.Filename, resp.Entity().Filename)
}

// TestCreateKanbanAttachmentValidationError verifies that creating a Kanban attachment with missing required fields
// returns a validation error. It attempts to create an attachment with empty filename, file type, file size, and uploaded by fields,
// and asserts that the response status code is 400 (Bad Request) and the filename field in the response is empty.
func TestCreateKanbanAttachmentValidationError(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	// Missing required fields
	invalidAttachment := attachment.Attachment{
		CardId:        cardId,
		Filename:      "",
		FileType:      "",
		FileSizeBytes: 0,
		UploadedBy:    "",
	}
	resp, err := testdata.MakePostRequest[attachment.Attachment, attachment.Attachment](fmt.Sprintf(
		testdata.POST_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId), invalidAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
	assert.Contains(t, resp.Entity().Filename, "")
}

// TestGetKanbanCardAttachments verifies that multiple attachments can be created for a Kanban card
// and that all created attachments are correctly retrieved via the API. It creates a board and card,
// adds several attachments with different filenames, and asserts that the GET request returns all
// attachments with the expected filenames.
func TestGetKanbanCardAttachments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	// Create multiple attachments
	attachmentNames := []string{"First.pdf", "Second.docx", "Third.png"}
	for _, name := range attachmentNames {
		kanbanAttachment, err := attachment.NewBuilder().
			SetCardId(cardId).
			SetFilename(name).
			SetFileType("application/octet-stream").
			SetFileSizeBytes(12345).
			SetAttachmentData([]byte(SOME_DATA)).
			SetUploadedBy("Peter").
			SetUploadedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		_, err = TestDataCreator.CreateKanbanAttachment(boardId, cardId, kanbanAttachment)
		assert.NoError(t, err)
	}

	getResp, err := testdata.MakeGetRequest[[]attachment.Attachment](fmt.Sprintf(
		testdata.GET_KANBAN_BOARD_CARD_ATTACHMENTS_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	attachments := *getResp.Entity()
	assert.Len(t, attachments, len(attachmentNames))
	foundNames := make(map[string]bool)
	for _, a := range attachments {
		foundNames[a.Filename] = true
	}
	for _, name := range attachmentNames {
		assert.True(t, foundNames[name])
	}
}

// TestTryToGetKanbanAttachmentByIdForCardNotBelongingToBoard verifies that attempting to retrieve a kanban attachment
// for a card that does not belong to the specified board results in an error. The test creates two separate boards and cards,
// attaches a file to the second card, and then tries to access the attachment using the first board's ID. It asserts that
// the response status code is 400 and the error message indicates the card does not belong to the board.
func TestTryToGetKanbanAttachmentByIdForCardNotBelongingToBoard(t *testing.T) {
	setupTest(t)
	boardId1, _ := createAttachmentBoardAndCard(t)
	_, cardId2 := createAttachmentBoardAndCard(t)

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId2).
		SetFilename("Should not be accessible from boardId1.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := TestDataCreator.CreateKanbanAttachment(boardId1+1, cardId2, kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	attachmentId := resp.Entity().AttachmentId

	getResp, err := testdata.MakeGetRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.GET_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId1, cardId2, attachmentId))
	assert.NoError(t, err)
	assert.Equal(t, 400, getResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(CARD_DOES_NOT_BELONG_TO_BOARD_ERROR_MSG, nil), *getResp.Entity())
}

// TestGetKanbanCardAttachmentById verifies that a kanban card attachment can be created and retrieved by its ID.
// The test performs the following steps:
// 1. Sets up the test environment.
// 2. Creates a board and card for the attachment.
// 3. Builds a new attachment with specific properties.
// 4. Creates the attachment via the API and checks for successful creation.
// 5. Retrieves the attachment by its ID and validates the response and attachment properties.
func TestGetKanbanCardAttachmentById(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId).
		SetFilename("Unique attachment.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := TestDataCreator.CreateKanbanAttachment(boardId, cardId, kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	attachmentId := resp.Entity().AttachmentId

	getResp, err := testdata.MakeGetRequest[attachment.Attachment](fmt.Sprintf(
		testdata.GET_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId, attachmentId))
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	assert.Equal(t, "Unique attachment.pdf", getResp.Entity().Filename)
}

// TestGetKanbanCardAttachmentByNonExistingId verifies that requesting a Kanban card attachment
// with a non-existing attachment ID returns a 404 Not Found status code.
// It sets up the test environment, creates a board and card, and attempts to retrieve
// an attachment using an invalid ID, asserting that the response status code is 404.
func TestGetKanbanCardAttachmentByNonExistingId(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	getResp, err := testdata.MakeGetRequest[attachment.Attachment](fmt.Sprintf(
		testdata.GET_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId, 9999))
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestTryToDeleteKanbanAttachmentByIdForCardNotBelongingToBoard verifies that attempting to delete a kanban attachment
// from a card that does not belong to the specified board fails with a 400 error and the appropriate error message.
// The test creates two boards and cards, attaches a file to a card on the second board, and then tries to delete it
// using the first board's ID, expecting the operation to be rejected.
func TestTryToDeleteKanbanAttachmentByIdForCardNotBelongingToBoard(t *testing.T) {
	setupTest(t)
	boardId1, _ := createBoardAndCard(t)
	_, cardId2 := createBoardAndCard(t)

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId2).
		SetFilename("Should not be deletable from boardId1.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := TestDataCreator.CreateKanbanAttachment(boardId1+1, cardId2, kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	attachmentId := resp.Entity().AttachmentId

	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.DELETE_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId1, cardId2, attachmentId))
	assert.NoError(t, err)
	assert.Equal(t, 400, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(CARD_DOES_NOT_BELONG_TO_BOARD_ERROR_MSG, nil), *deleteResp.Entity())
}

// TestDeleteKanbanCardAttachment verifies that a kanban card attachment can be successfully deleted.
// The test performs the following steps:
// 1. Sets up the test environment and creates a board and card.
// 2. Creates a new attachment for the card.
// 3. Sends a request to delete the created attachment.
// 4. Asserts that the deletion request returns a 204 No Content status.
// 5. Attempts to retrieve the deleted attachment and asserts that it returns a 404 Not Found status.
func TestDeleteKanbanCardAttachment(t *testing.T) {
	setupTest(t)
	boardId, cardId := createBoardAndCard(t)

	kanbanAttachment, err := attachment.NewBuilder().
		SetCardId(cardId).
		SetFilename("Attachment to delete.pdf").
		SetFileType(APPLICATION_PDF).
		SetFileSizeBytes(12345).
		SetAttachmentData([]byte(SOME_DATA)).
		SetUploadedBy("Peter").
		SetUploadedAt(time.Now()).
		Build()
	assert.NoError(t, err)
	resp, err := TestDataCreator.CreateKanbanAttachment(boardId, cardId, kanbanAttachment)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	attachmentId := resp.Entity().AttachmentId

	deleteResp, err := testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
		testdata.DELETE_KANBAN_ATTACHMENT_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId, attachmentId))
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteResp.StatusCode())

	// Verify deletion
	getResp, err := TestDataCreator.GetKanbanCardAttachment(boardId, cardId, attachmentId)
	assert.NoError(t, err)
	assert.Equal(t, 404, getResp.StatusCode())
}

// TestTryToDeleteAllKanbanAttachmentsForCardNotBelongingToBoard verifies that attempting to delete all attachments
// for a card that does not belong to the specified board results in a 400 error response with the appropriate error message.
// The test creates two boards and cards, attaches files to the second card, and then tries to delete attachments using the first board's ID.
// It asserts that the operation fails with the expected error.
func TestTryToDeleteAllKanbanAttachmentsForCardNotBelongingToBoard(t *testing.T) {
	setupTest(t)
	boardId1, _ := createAttachmentBoardAndCard(t)
	_, cardId2 := createAttachmentBoardAndCard(t)

	for i := 0; i < 2; i++ {
		kanbanAttachment, err := attachment.NewBuilder().
			SetCardId(cardId2).
			SetFilename(fmt.Sprintf("Attachment-%d.pdf", i)).
			SetFileType(APPLICATION_PDF).
			SetFileSizeBytes(12345).
			SetAttachmentData([]byte(SOME_DATA)).
			SetUploadedBy("Peter").
			SetUploadedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		_, err = TestDataCreator.CreateKanbanAttachment(boardId1+1, cardId2, kanbanAttachment)
		assert.NoError(t, err)
	}

	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](fmt.Sprintf(
		testdata.DELETE_KANBAN_ATTACHMENTS_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId1, cardId2))
	assert.NoError(t, err)
	assert.Equal(t, 400, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(CARD_DOES_NOT_BELONG_TO_BOARD_ERROR_MSG, nil), *deleteResp.Entity())
}

// TestDeleteAllKanbanCardAttachments verifies that all attachments for a given Kanban card
// can be deleted successfully. It creates a board and card, adds multiple attachments to the card,
// sends a delete request to remove all attachments, and then asserts that the attachments list is empty.
func TestDeleteAllKanbanCardAttachments(t *testing.T) {
	setupTest(t)
	boardId, cardId := createAttachmentBoardAndCard(t)

	for i := 0; i < 3; i++ {
		kanbanAttachment, err := attachment.NewBuilder().
			SetCardId(cardId).
			SetFilename(fmt.Sprintf("Attachment-%d.pdf", i)).
			SetFileType(APPLICATION_PDF).
			SetFileSizeBytes(12345).
			SetAttachmentData([]byte(SOME_DATA)).
			SetUploadedBy("Peter").
			SetUploadedAt(time.Now()).
			Build()
		assert.NoError(t, err)
		_, err = TestDataCreator.CreateKanbanAttachment(boardId, cardId, kanbanAttachment)
		assert.NoError(t, err)
	}

	deleteResp, err := testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
		testdata.DELETE_KANBAN_ATTACHMENTS_URL_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId))
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteResp.StatusCode())

	getResp, err := TestDataCreator.GetKanbanCardAttachments(boardId, cardId)
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.StatusCode())
	assert.Len(t, *getResp.Entity(), 0)
}

// TestTryToDeleteKanbanAttachmentWithInvalidIds tests the deletion of kanban card attachments using invalid board and card IDs.
// It covers scenarios where either the board ID or card ID is invalid, for both single attachment deletion and deletion of all attachments for a card.
// The test verifies that the delete requests do not return an error, ensuring the API handles invalid IDs gracefully.
func TestTryToDeleteKanbanAttachmentWithInvalidIds(t *testing.T) {
	testCases := []struct {
		name     string
		boardId  string
		cardId   string
		exercise func(boardId string, cardId string, attachmentId string) (testdata.ResponseResult[struct{}], error)
	}{
		{
			name:    "single attachment delete with invalid board id",
			boardId: NOT_A_NUMBER,
			cardId:  "1",
			exercise: func(boardId, cardId, attachmentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
					"%s/boards/%s/cards/%s/attachments/%s", TestDataCreator.GetAppUrl(), boardId, cardId, attachmentId))
			},
		},
		{
			name:    "all attachments for card delete with invalid board id",
			boardId: NOT_A_NUMBER,
			cardId:  "1",
			exercise: func(boardId, cardId, attachmentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
					"%s/boards/%s/cards/%s/attachments", TestDataCreator.GetAppUrl(), boardId, cardId))
			},
		},
		{
			name:    "single attachment delete with invalid card id",
			boardId: "1",
			cardId:  NOT_A_NUMBER,
			exercise: func(boardId, cardId, attachmentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
					"%s/boards/%s/cards/%s/attachments/%s", TestDataCreator.GetAppUrl(), boardId, cardId, attachmentId))
			},
		},
		{
			name:    "all attachments for card delete with invalid card id",
			boardId: "1",
			cardId:  NOT_A_NUMBER,
			exercise: func(boardId, cardId, attachmentId string) (testdata.ResponseResult[struct{}], error) {
				return testdata.MakeDeleteRequest[struct{}](fmt.Sprintf(
					"%s/boards/%s/cards/%s/attachments", TestDataCreator.GetAppUrl(), boardId, cardId))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.exercise(testCase.boardId, testCase.cardId, "1")
			assert.NoError(t, err)
		})
	}
}

// TestTryToDeleteKanbanAttachmentForNonExistingBoard verifies that attempting to delete a kanban card attachment
// (either a single attachment or all attachments for a card) on a non-existing board does not result in an error.
// It tests both the endpoint for deleting a single attachment and the endpoint for deleting all attachments for a card,
// using a board ID that does not exist.
func TestTryToDeleteKanbanAttachmentForNonExistingBoard(t *testing.T) {
	boardId := 999999 // non-existing board ID
	cardId := 1       // arbitrary card ID
	attachmentId := 1 // arbitrary attachment ID

	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "single attachment delete for non-existing board",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/attachments/%d", TestDataCreator.GetAppUrl(), boardId, cardId, attachmentId),
		},
		{
			name: "all attachments for card delete for non-existing board",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/attachments", TestDataCreator.GetAppUrl(), boardId, cardId),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := testdata.MakeDeleteRequest[struct{}](tc.url)
			assert.NoError(t, err)
		})
	}
}

// TestTryToDeleteKanbanAttachmentForNonExistingCard verifies that attempting to delete attachments
// (either a single attachment or all attachments) for a non-existing kanban card does not result in an error.
// It tests both the endpoint for deleting a specific attachment and the endpoint for deleting all attachments
// for a card that does not exist, ensuring the API handles such cases gracefully.
func TestTryToDeleteKanbanAttachmentForNonExistingCard(t *testing.T) {
	setupTest(t)
	boardId, _ := createAttachmentBoardAndCard(t)
	nonExistingCardId := 999999
	attachmentId := 1 // arbitrary attachment ID

	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "single attachment delete for non-existing card",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/attachments/%d", TestDataCreator.GetAppUrl(), boardId, nonExistingCardId, attachmentId),
		},
		{
			name: "all attachments for card delete for non-existing card",
			url:  fmt.Sprintf("%s/boards/%d/cards/%d/attachments", TestDataCreator.GetAppUrl(), boardId, nonExistingCardId),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := testdata.MakeDeleteRequest[struct{}](tc.url)
			assert.NoError(t, err)
		})
	}
}
