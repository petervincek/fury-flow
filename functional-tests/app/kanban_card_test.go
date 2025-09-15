//go:build functional
// +build functional

package app_test

import (
	"fmt"
	"testing"

	"github.com/petervincek/fury-flow/functional-tests/utils/testdata"
	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/card"
	"github.com/petervincek/fury-flow/internal/api/common"
	"github.com/stretchr/testify/assert"
)

const (
	EXPECTING_NO_ERROR_KANBAN_BOARD_TEST_DATA_CREATOR_MSG = "expecting no error while creating Kanban Board through TestDataCreator"
	EXPECTING_NO_ERROR_KANBAN_BOARD_BUILDER_MSG           = "expecting no error while creating Kanban Card through Builder"
	EXPECTING_NO_ERROR_KANBAN_BOARD_API_MSG               = "expecting no error while creating Kanban Card through API"
)

func TestCreateKanbanCard(t *testing.T) {
	setupTest(t)

	// create test data
	// create a kanban board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Project Hockey News").
		SetDescription("Kanban Board for project related to Hockey News").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_MSG)
	createdKanbanBoard, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_TEST_DATA_CREATOR_MSG)
	_, err = TestDataCreator.GetKanbanBoardById(createdKanbanBoard.Entity().BoardId)
	assert.NoError(t, err, "expecting no error here")

	// exercise
	kanbanCard, err := card.NewBuilder().
		SetBoardId(createdKanbanBoard.Entity().BoardId).
		SetTitle("title").
		SetDescription("desc").
		SetStatus(card.StatusDone).
		SetAcceptanceCriteria("something").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityHigh).
		SetStoryPoints(1).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_BUILDER_MSG)
	response, err := testdata.MakePostRequest[card.Card, card.Card](
		fmt.Sprintf(testdata.POST_KANBAN_CARD_URL_TEMPLATE, TestDataCreator.GetAppUrl(),
			createdKanbanBoard.Entity().BoardId), kanbanCard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_API_MSG)
	assert.Equal(t, 201, response.StatusCode(), EXPECTING_STATUS_CODE_201_MSG)
	assert.NotZero(t, response.Entity().CardId, "expecting non zero card id")
}

func TestGetKanbanCardsForExistingKanbanBoardWithZeroCards(t *testing.T) {
	setupTest(t)

	// create a kanban board with no cards
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Empty Kanban Board").
		SetDescription("Kanban Board with zero cards").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_BUILDER_MSG)
	createdKanbanBoard, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_TEST_DATA_CREATOR_MSG)

	// exercise
	response, err := testdata.MakeGetRequest[[]card.Card](
		fmt.Sprintf(testdata.GET_KANBAN_BOARD_CARDS_TEMPLATE, TestDataCreator.GetAppUrl(), createdKanbanBoard.Entity().BoardId))
	assert.NoError(t, err, "expecting no error while getting cards for kanban board")
	assert.Equal(t, 200, response.StatusCode(), "expecting status code 200 - OK")
	assert.Empty(t, response.Entity(), "expecting zero cards for newly created board")
}

func TestGetKanbanCardsForExistingKanbanBoardWithMultipleCards(t *testing.T) {
	setupTest(t)

	// create a kanban board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Kanban Board With Multiple Cards").
		SetDescription("Kanban Board for testing multiple cards").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err, "expecting no error while creating Kanban Board through Builder")
	createdKanbanBoard, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_TEST_DATA_CREATOR_MSG)

	// create multiple cards
	cardTitles := []string{"Card 1", "Card 2", "Card 3"}
	createdCards := make([]card.Card, 0, len(cardTitles))
	for _, title := range cardTitles {
		kanbanCard, err := card.NewBuilder().
			SetBoardId(createdKanbanBoard.Entity().BoardId).
			SetTitle(title).
			SetDescription("desc for " + title).
			SetStatus(card.StatusTodo).
			SetAcceptanceCriteria("criteria for " + title).
			SetType(card.TypeFeature).
			SetPriority(card.PriorityMedium).
			SetStoryPoints(2).
			SetCreatedBy("Peter").
			Build()
		assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_BUILDER_MSG)
		response, err := TestDataCreator.CreateKanbanCard(createdKanbanBoard.Entity().BoardId, kanbanCard)
		assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_API_MSG)
		assert.Equal(t, 201, response.StatusCode(), "expecting status code 201 - CREATED")
		createdCards = append(createdCards, *response.Entity())
	}

	// exercise
	response, err := testdata.MakeGetRequest[[]card.Card](
		fmt.Sprintf(testdata.GET_KANBAN_BOARD_CARDS_TEMPLATE, TestDataCreator.GetAppUrl(), createdKanbanBoard.Entity().BoardId))
	assert.NoError(t, err, "expecting no error while getting cards for kanban board")
	assert.Equal(t, 200, response.StatusCode(), "expecting status code 200 - OK")
	assert.Len(t, *response.Entity(), len(cardTitles), "expecting number of cards to match created cards")

	// verify that all created cards are present
	foundTitles := make(map[string]bool)
	for _, c := range *response.Entity() {
		foundTitles[c.Title] = true
	}
	for _, title := range cardTitles {
		assert.True(t, foundTitles[title], "expecting card with title %s to be present", title)
	}
}

func TestGetKanbanCardsForNonExistingKanbanBoard(t *testing.T) {
	setupTest(t)

	// exercise
	response, err := testdata.MakeGetRequest[common.ErrorMsg](
		fmt.Sprintf(testdata.GET_KANBAN_BOARD_CARDS_TEMPLATE, TestDataCreator.GetAppUrl(), 9999))
	// verify
	assert.NoError(t, err, "expecting no error while trying to get cards for non-existing kanban board")
	assert.Equal(t, 404, response.StatusCode(), "expecting status code 404 - NOT FOUND")
	assert.Equal(t, response.Entity().Message, "board with id: 9999 does not exist", "expecting specific error message")
}

func TestDeleteExistingCardInExistingKanbanBoard(t *testing.T) {
	setupTest(t)
	// create a kanban board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Kanban Board For Delete Card").
		SetDescription("Kanban Board for testing card deletion").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err, "expecting no error while creating Kanban Board through Builder")
	createdKanbanBoard, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_TEST_DATA_CREATOR_MSG)

	// create a card
	kanbanCard, err := card.NewBuilder().
		SetBoardId(createdKanbanBoard.Entity().BoardId).
		SetTitle("Card To Delete").
		SetDescription("desc").
		SetStatus(card.StatusTodo).
		SetAcceptanceCriteria("criteria").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityLow).
		SetStoryPoints(1).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_BUILDER_MSG)
	response, err := TestDataCreator.CreateKanbanCard(createdKanbanBoard.Entity().BoardId, kanbanCard)
	assert.NoError(t, err, EXPECTING_NO_ERROR_KANBAN_BOARD_API_MSG)
	assert.Equal(t, 201, response.StatusCode(), "expecting status code 201 - CREATED")
	createdCard := response.Entity()

	// delete the card
	deleteResponse, err := testdata.MakeDeleteRequest[struct{}](
		fmt.Sprintf(testdata.DELETE_KANBAN_CARD_URL_TEMPLATE, TestDataCreator.GetAppUrl(), createdKanbanBoard.Entity().BoardId, createdCard.CardId))
	assert.NoError(t, err, "expecting no error while deleting Kanban Card")
	assert.Equal(t, 204, deleteResponse.StatusCode(), "expecting status code 204 - NO CONTENT")

	// verify the card is deleted
	getResponse, err := testdata.MakeGetRequest[common.ErrorMsg](
		fmt.Sprintf(testdata.GET_KANBAN_CARD_URL_TEMPLATE, TestDataCreator.GetAppUrl(), createdKanbanBoard.Entity().BoardId, createdCard.CardId))
	assert.NoError(t, err, "expecting no error while getting deleted card")
	assert.Equal(t, 404, getResponse.StatusCode(), "expecting status code 404 - NOT FOUND")
	assert.Contains(t, getResponse.Entity().Message, "no kanban card for cardId:", "expecting not found error message")
}
