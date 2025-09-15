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
	NOT_A_NUMBER                                     = "not-a-number"
	UNABLE_TO_PARSE_BOARD_ID_ERROR_MSG               = "unable to parse board id"
	UNABLE_TO_PARSE_CARD_ID_ERROR_MSG                = "unable to parse card id"
	APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE         = "%s/boards/%d/cards/%d/dependencies"
	CARD_WITH_ID_DOES_NOT_EXIST_ERROR_MSG_TEMPLATE   = "card with id: %d does not exist"
	APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE = "%s/boards/%d/cards/%d/dependencies/%d"
)

func TestGetKanbanCardDependenciesForInvalidBoardId(t *testing.T) {
	cardId := 1
	invalidBoardId := NOT_A_NUMBER // boardId should be a number

	getDepsURL := fmt.Sprintf("%s/boards/%s/cards/%d/dependencies", TestDataCreator.GetAppUrl(), invalidBoardId, cardId)
	getDepsResp, err := testdata.MakeGetRequest[common.ErrorMsg](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 400, getDepsResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(UNABLE_TO_PARSE_BOARD_ID_ERROR_MSG, nil), *getDepsResp.Entity())
}

func TestGetKanbanCardDependenciesForInvalidCardId(t *testing.T) {
	boardId := 1
	invalidCardId := NOT_A_NUMBER // cardId should be a number

	getDepsURL := fmt.Sprintf("%s/boards/%d/cards/%s/dependencies", TestDataCreator.GetAppUrl(), boardId, invalidCardId)
	getDepsResp, err := testdata.MakeGetRequest[common.ErrorMsg](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 400, getDepsResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(UNABLE_TO_PARSE_CARD_ID_ERROR_MSG, nil), *getDepsResp.Entity())
}

func TestGetKanbanCardDependenciesForNonExistingKanbanBoard(t *testing.T) {
	nonExistingBoardId := 999999 // assuming this board ID does not exist
	cardId := 1                  // arbitrary card ID

	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId)
	getDepsResp, err := testdata.MakeGetRequest[common.ErrorMsg](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 404, getDepsResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(fmt.Sprintf("board with id: %d does not exist", nonExistingBoardId), nil), *getDepsResp.Entity())
}

func TestGetKanbanCardDependenciesForNonExistingKanbanCard(t *testing.T) {
	setupTest(t)

	// Create a board to use a valid boardId
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Non-existing Card Get Dependencies Board").
		SetDescription("Board for non-existing card get dependencies test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	nonExistingCardId := 999999 // assuming this card ID does not exist

	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId)
	getDepsResp, err := testdata.MakeGetRequest[common.ErrorMsg](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 404, getDepsResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(fmt.Sprintf(CARD_WITH_ID_DOES_NOT_EXIST_ERROR_MSG_TEMPLATE, nonExistingCardId), nil), *getDepsResp.Entity())
}

func TestGetKanbanCardDependenciesForCardWithNoDependencies(t *testing.T) {
	setupTest(t)

	// Create a board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("No Dependencies Board").
		SetDescription("Board for card with no dependencies test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// Create a card with no dependencies
	kanbanCard, err := card.NewBuilder().
		SetBoardId(boardId).
		SetTitle("Solo Card").
		SetDescription("Card with no dependencies").
		SetStatus(card.StatusTodo).
		SetAcceptanceCriteria("criteria for Solo Card").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityMedium).
		SetStoryPoints(2).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	cardResp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
	assert.NoError(t, err)
	assert.Equal(t, 201, cardResp.StatusCode())
	cardId := cardResp.Entity().CardId

	// Get dependencies for the card (should be empty)
	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId)
	getDepsResp, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp.StatusCode())
	depCards := *getDepsResp.Entity()
	assert.Len(t, depCards, 0)
}

func TestGetKanbanCardDependenciesForCardWithMutipleDependencies(t *testing.T) {
	setupTest(t)

	// Create a board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Multiple Dependencies Board").
		SetDescription("Board for card with multiple dependencies test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// Create cards
	cardTitles := []string{"Main Card 0", "Dep Card 1", "Dep Card 2", "Dep Card 3"}
	createdCards := make([]card.Card, 0, len(cardTitles))
	for _, title := range cardTitles {
		kanbanCard, err := card.NewBuilder().
			SetBoardId(boardId).
			SetTitle(title).
			SetDescription("description for " + title).
			SetStatus(card.StatusTodo).
			SetAcceptanceCriteria("acceptance criteria for " + title).
			SetType(card.TypeFeature).
			SetPriority(card.PriorityMedium).
			SetStoryPoints(2).
			SetCreatedBy("Peter").
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		createdCards = append(createdCards, *resp.Entity())
	}

	mainCard := createdCards[0]
	depCard1 := createdCards[1]
	depCard2 := createdCards[2]
	depCard3 := createdCards[3]

	// Add multiple dependencies to Main Card
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, mainCard.CardId)
	dependencyPayload := []int{depCard1.CardId, depCard2.CardId, depCard3.CardId}
	depResp, err := testdata.MakePostRequest[[]int, struct{}](dependencyURL, dependencyPayload)
	assert.NoError(t, err)
	assert.Equal(t, 201, depResp.StatusCode())

	// Get dependencies for Main Card
	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, mainCard.CardId)
	getDepsResp, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp.StatusCode())
	depCards := *getDepsResp.Entity()
	assert.Len(t, depCards, 3)

	found := map[int]bool{}
	for _, c := range depCards {
		found[c.CardId] = true
	}
	assert.True(t, found[depCard1.CardId])
	assert.True(t, found[depCard2.CardId])
	assert.True(t, found[depCard3.CardId])
}

func TestCreateKanbanCardDependencies(t *testing.T) {
	setupTest(t)

	// 1. Create a board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Dependency Test Board").
		SetDescription("Board for card dependency tests").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// 2. Create cards
	cardTitles := []string{"Card A", "Card B", "Card C"}
	createdCards := make([]card.Card, 0, len(cardTitles))
	for _, title := range cardTitles {
		kanbanCard, err := card.NewBuilder().
			SetBoardId(boardId).
			SetTitle(title).
			SetDescription("some desc for " + title).
			SetStatus(card.StatusTodo).
			SetAcceptanceCriteria("some criteria for " + title).
			SetType(card.TypeFeature).
			SetPriority(card.PriorityMedium).
			SetStoryPoints(2).
			SetCreatedBy("Peter").
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		createdCards = append(createdCards, *resp.Entity())
	}

	// 3. Create dependencies: Card A depends on Card B and Card C
	cardA := createdCards[0]
	cardB := createdCards[1]
	cardC := createdCards[2]

	// Assuming you have an endpoint like /boards/{boardId}/cards/{cardId}/dependencies
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardA.CardId)
	dependencyPayload := []int{cardB.CardId, cardC.CardId}
	depResp, err := testdata.MakePostRequest[[]int, struct{}](dependencyURL, dependencyPayload)
	assert.NoError(t, err)
	assert.Equal(t, 201, depResp.StatusCode())

	// 4. Verify dependencies
	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardA.CardId)
	getDepsResp, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp.StatusCode())
	depCards := *getDepsResp.Entity()
	assert.Len(t, depCards, 2)
	found := map[int]bool{}
	for _, c := range depCards {
		found[c.CardId] = true
	}
	assert.True(t, found[cardB.CardId])
	assert.True(t, found[cardC.CardId])
}

func TestTryToCreateKanbanCardDependencyWithInvalidBoardId(t *testing.T) {
	boardId := NOT_A_NUMBER // boardId should be a number
	cardId := 1
	// Assuming you have an endpoint like /boards/{boardId}/cards/{cardId}/dependencies
	dependencyURL := fmt.Sprintf("%s/boards/%s/cards/%d/dependencies", TestDataCreator.GetAppUrl(), boardId, cardId)
	depResp, err := testdata.MakePostRequest[[]int, common.ErrorMsg](dependencyURL, []int{})
	assert.NoError(t, err)
	assert.Equal(t, 400, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(UNABLE_TO_PARSE_BOARD_ID_ERROR_MSG, nil), *depResp.Entity())
}

func TestTryToCreateKanbanCardDependencyWithInvalidCardId(t *testing.T) {
	boardId := 1
	cardId := NOT_A_NUMBER // cardId should be a number
	// Assuming you have an endpoint like /boards/{boardId}/cards/{cardId}/dependencies
	dependencyURL := fmt.Sprintf("%s/boards/%d/cards/%s/dependencies", TestDataCreator.GetAppUrl(), boardId, cardId)
	depResp, err := testdata.MakePostRequest[[]int, common.ErrorMsg](dependencyURL, []int{})
	assert.NoError(t, err)
	assert.Equal(t, 400, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(UNABLE_TO_PARSE_CARD_ID_ERROR_MSG, nil), *depResp.Entity())
}

func TestTryToCreateKanbanCardDependenciesWithInvalidCardDependencyId(t *testing.T) {
	boardId := 1
	cardId := 1
	dependencyCardIds := []string{NOT_A_NUMBER} // should be an array of numbers
	// Assuming you have an endpoint like /boards/{boardId}/cards/{cardId}/dependencies
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardId)
	depResp, err := testdata.MakePostRequest[[]string, common.ErrorMsg](dependencyURL, dependencyCardIds)
	assert.NoError(t, err)
	assert.Equal(t, 400, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg("unable to parse dependency card ids", nil), *depResp.Entity())
}

func TestTryToCreateKanbanCardDependencyForNonExistingKanbanBoard(t *testing.T) {
	nonExistingBoardId := 999999     // assuming this board ID does not exist
	cardId := 1                      // arbitrary card ID
	dependencyPayload := []int{2, 3} // arbitrary dependency card IDs

	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId)
	depResp, err := testdata.MakePostRequest[[]int, common.ErrorMsg](dependencyURL, dependencyPayload)
	assert.NoError(t, err)
	assert.Equal(t, 404, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg("board with id: 999999 does not exist", nil), *depResp.Entity())
}

func TestTryToCreateKanbanCardDependencyForNonExistingKanbanCard(t *testing.T) {
	setupTest(t)

	// Create a board to use a valid boardId
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Non-existing Card Dependency Board").
		SetDescription("Board for non-existing card dependency test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	nonExistingCardId := 999999      // assuming this card ID does not exist
	dependencyPayload := []int{1, 2} // arbitrary dependency card IDs

	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId)
	depResp, err := testdata.MakePostRequest[[]int, common.ErrorMsg](dependencyURL, dependencyPayload)
	assert.NoError(t, err)
	assert.Equal(t, 404, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(fmt.Sprintf(CARD_WITH_ID_DOES_NOT_EXIST_ERROR_MSG_TEMPLATE, nonExistingCardId), nil), *depResp.Entity())
}

func TestTryToCreateKanbanCardDependencyForDepedencyCardsThatDoesNotExists(t *testing.T) {
	setupTest(t)

	// Create a board to use a valid boardId
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Dependency Cards Not Exist Board").
		SetDescription("Board for dependency cards not exist test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// Create a card to use a valid cardId
	kanbanCard, err := card.NewBuilder().
		SetBoardId(boardId).
		SetTitle("Main Card").
		SetDescription("desc for Main Card").
		SetStatus(card.StatusTodo).
		SetAcceptanceCriteria("criteria for Main Card").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityMedium).
		SetStoryPoints(2).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	cardResp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
	assert.NoError(t, err)
	assert.Equal(t, 201, cardResp.StatusCode())
	mainCardId := cardResp.Entity().CardId

	// Try to create dependencies with non-existing card IDs
	nonExistingDepIds := []int{888888, 999999}
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, mainCardId)
	depResp, err := testdata.MakePostRequest[[]int, common.ErrorMsg](dependencyURL, nonExistingDepIds)
	assert.NoError(t, err)
	assert.Equal(t, 500, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg("internal server error", nil), *depResp.Entity())
}

func TestTryToCreateKanbanCardDependencyWithEmptyListOfDependencyIds(t *testing.T) {
	setupTest(t)

	// Create a board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Empty Dependency List Board").
		SetDescription("Board for empty dependency list test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// Create a card
	kanbanCard, err := card.NewBuilder().
		SetBoardId(boardId).
		SetTitle("Main Card - no dependencies").
		SetDescription("desc for Main Card with no dependencies").
		SetStatus(card.StatusTodo).
		SetAcceptanceCriteria("criteria for Main Card with no dependencies").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityMedium).
		SetStoryPoints(2).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	cardResp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
	assert.NoError(t, err)
	assert.Equal(t, 201, cardResp.StatusCode())
	mainCardId := cardResp.Entity().CardId

	// Try to create dependencies with an empty list
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, mainCardId)
	emptyDepIds := []int{}
	depResp, err := testdata.MakePostRequest[[]int, common.ErrorMsg](dependencyURL, emptyDepIds)
	assert.NoError(t, err)
	assert.Equal(t, 400, depResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg("at least one dependency card ID must be provided", nil), *depResp.Entity())
}

func TestTryToDeleteCardDependenciesForInvalidBoardId(t *testing.T) {
	cardId := 1
	depCardId := 2
	invalidBoardId := NOT_A_NUMBER
	deleteDepURL := fmt.Sprintf("%s/boards/%s/cards/%d/dependencies/%d", TestDataCreator.GetAppUrl(), invalidBoardId, cardId, depCardId)
	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 400, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(UNABLE_TO_PARSE_BOARD_ID_ERROR_MSG, nil), *deleteResp.Entity())
}

func TestTryToDeleteCardDependenciesForInvalidCardId(t *testing.T) {
	boardId := 1
	depCardId := 2
	invalidCardId := NOT_A_NUMBER
	deleteDepURL := fmt.Sprintf("%s/boards/%d/cards/%s/dependencies/%d", TestDataCreator.GetAppUrl(), boardId, invalidCardId, depCardId)
	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 400, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(UNABLE_TO_PARSE_CARD_ID_ERROR_MSG, nil), *deleteResp.Entity())
}

func TestTryToDeleteCardDependenciesForInvalidDependencyCardId(t *testing.T) {
	boardId := 1
	cardId := 1
	invalidDepCardId := NOT_A_NUMBER
	deleteDepURL := fmt.Sprintf("%s/boards/%d/cards/%d/dependencies/%s", TestDataCreator.GetAppUrl(), boardId, cardId, invalidDepCardId)
	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 400, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg("unable to parse card dependency id", nil), *deleteResp.Entity())
}

func TestTryToDeleteCardDependenciesForNonExistingKanbanBoard(t *testing.T) {
	nonExistingBoardId := 999999
	cardId := 1
	depCardId := 2
	deleteDepURL := fmt.Sprintf(APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), nonExistingBoardId, cardId, depCardId)
	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 404, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(fmt.Sprintf("board with id: %d does not exist", nonExistingBoardId), nil), *deleteResp.Entity())
}

func TestTryToDeleteCardDependenciesForNonExistingKanbanCard(t *testing.T) {
	setupTest(t)

	// Create a board to use a valid boardId
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Delete Non-existing Card Dependency Board").
		SetDescription("Board for non-existing card dependency delete test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	nonExistingCardId := 999999
	depCardId := 2
	deleteDepURL := fmt.Sprintf(APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, nonExistingCardId, depCardId)
	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 404, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(fmt.Sprintf(CARD_WITH_ID_DOES_NOT_EXIST_ERROR_MSG_TEMPLATE, nonExistingCardId), nil), *deleteResp.Entity())
}

func TestTryToDeleteCardDependenciesForNonExistingDependencyCard(t *testing.T) {
	setupTest(t)

	// Create a board and a card to use valid boardId and cardId
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Delete Non-existing Dependency Card Board").
		SetDescription("Board for non-existing dependency card delete test").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	kanbanCard, err := card.NewBuilder().
		SetBoardId(boardId).
		SetTitle("Main Card").
		SetDescription("desc for Main Card").
		SetStatus(card.StatusTodo).
		SetAcceptanceCriteria("criteria for Main Card").
		SetType(card.TypeFeature).
		SetPriority(card.PriorityMedium).
		SetStoryPoints(2).
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	cardResp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
	assert.NoError(t, err)
	assert.Equal(t, 201, cardResp.StatusCode())
	mainCardId := cardResp.Entity().CardId

	nonExistingDepCardId := 9999
	deleteDepURL := fmt.Sprintf(APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, mainCardId, nonExistingDepCardId)
	deleteResp, err := testdata.MakeDeleteRequest[common.ErrorMsg](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 404, deleteResp.StatusCode())
	assert.Equal(t, common.NewErrorMsg(fmt.Sprintf("dependency card with id: %d does not exist", nonExistingDepCardId), nil), *deleteResp.Entity())
}

func TestDeleteSomeOfCardDependencies(t *testing.T) {
	setupTest(t)

	// Create board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Delete Some Dependencies Board").
		SetDescription("Board for deleting some dependencies").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// Create cards
	cardTitles := []string{"Card X", "Card Y", "Card Z"}
	createdCards := make([]card.Card, 0, len(cardTitles))
	for _, title := range cardTitles {
		kanbanCard, err := card.NewBuilder().
			SetBoardId(boardId).
			SetTitle(title).
			SetDescription("desc for " + title).
			SetStatus(card.StatusTodo).
			SetAcceptanceCriteria("criteria for " + title).
			SetType(card.TypeFeature).
			SetPriority(card.PriorityMedium).
			SetStoryPoints(2).
			SetCreatedBy("Peter").
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		createdCards = append(createdCards, *resp.Entity())
	}

	cardX := createdCards[0]
	cardY := createdCards[1]
	cardZ := createdCards[2]

	// Establish dependencies: Card X depends on Card Y and Card Z
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardX.CardId)
	dependencyPayload := []int{cardY.CardId, cardZ.CardId}
	depResp, err := testdata.MakePostRequest[[]int, struct{}](dependencyURL, dependencyPayload)
	assert.NoError(t, err)
	assert.Equal(t, 201, depResp.StatusCode())

	// Verify initial dependencies
	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardX.CardId)
	getDepsResp, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp.StatusCode())
	depCards := *getDepsResp.Entity()
	assert.Len(t, depCards, 2)

	// Remove one dependency: Card X no longer depends on Card Y
	deleteDepURL := fmt.Sprintf(APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardX.CardId, cardY.CardId)
	deleteResp, err := testdata.MakeDeleteRequest[struct{}](deleteDepURL)
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteResp.StatusCode())

	// Verify new state: only Card Z should remain as a dependency
	getDepsResp2, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp2.StatusCode())
	depCards2 := *getDepsResp2.Entity()
	assert.Len(t, depCards2, 1)
	assert.Equal(t, cardZ.CardId, depCards2[0].CardId)
}

func TestDeleteAllCardDependencies(t *testing.T) {
	setupTest(t)

	// Create board
	kanbanBoard, err := board.NewBuilder().
		SetBoardName("Delete All Dependencies Board").
		SetDescription("Board for deleting all dependencies").
		SetCreatedBy("Peter").
		Build()
	assert.NoError(t, err)
	boardResp, err := TestDataCreator.CreateKanbanBoard(kanbanBoard)
	assert.NoError(t, err)
	boardId := boardResp.Entity().BoardId

	// Create cards
	cardTitles := []string{"Card M", "Card N", "Card O"}
	createdCards := make([]card.Card, 0, len(cardTitles))
	for _, title := range cardTitles {
		kanbanCard, err := card.NewBuilder().
			SetBoardId(boardId).
			SetTitle(title).
			SetDescription("desc for " + title).
			SetStatus(card.StatusTodo).
			SetAcceptanceCriteria("criteria for " + title).
			SetType(card.TypeFeature).
			SetPriority(card.PriorityMedium).
			SetStoryPoints(2).
			SetCreatedBy("Peter").
			Build()
		assert.NoError(t, err)
		resp, err := TestDataCreator.CreateKanbanCard(boardId, kanbanCard)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode())
		createdCards = append(createdCards, *resp.Entity())
	}

	cardM := createdCards[0]
	cardN := createdCards[1]
	cardO := createdCards[2]

	// Establish dependencies: Card M depends on Card N and Card O
	dependencyURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardM.CardId)
	dependencyPayload := []int{cardN.CardId, cardO.CardId}
	depResp, err := testdata.MakePostRequest[[]int, struct{}](dependencyURL, dependencyPayload)
	assert.NoError(t, err)
	assert.Equal(t, 201, depResp.StatusCode())

	// Verify initial dependencies
	getDepsURL := fmt.Sprintf(APP_CARD_DEPENDENCY_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardM.CardId)
	getDepsResp, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp.StatusCode())
	depCards := *getDepsResp.Entity()
	assert.Len(t, depCards, 2)

	// Remove all dependencies: Card M no longer depends on Card N and Card O
	deleteDepNURL := fmt.Sprintf(APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardM.CardId, cardN.CardId)
	deleteDepOURL := fmt.Sprintf(APP_CARD_DEPENDENCY_WITH_ID_URL_CONTEXT_TEMPLATE, TestDataCreator.GetAppUrl(), boardId, cardM.CardId, cardO.CardId)
	deleteRespN, err := testdata.MakeDeleteRequest[struct{}](deleteDepNURL)
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteRespN.StatusCode())
	deleteRespO, err := testdata.MakeDeleteRequest[struct{}](deleteDepOURL)
	assert.NoError(t, err)
	assert.Equal(t, 204, deleteRespO.StatusCode())

	// Verify new state: no dependencies should remain
	getDepsResp2, err := testdata.MakeGetRequest[[]card.Card](getDepsURL)
	assert.NoError(t, err)
	assert.Equal(t, 200, getDepsResp2.StatusCode())
	depCards2 := *getDepsResp2.Entity()
	assert.Len(t, depCards2, 0)
}
