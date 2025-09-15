package testdata

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/card"
)

const (
	POST_KANBAN_CARD_URL_TEMPLATE   = "%s/boards/%d/cards"
	GET_KANBAN_BOARD_CARDS_TEMPLATE = POST_KANBAN_CARD_URL_TEMPLATE
	GET_KANBAN_CARD_URL_TEMPLATE    = "%s/boards/%d/cards/%d"
	PUT_KANBAN_CARD_URL_TEMPLATE    = GET_KANBAN_CARD_URL_TEMPLATE
	DELETE_KANBAN_CARD_URL_TEMPLATE = GET_KANBAN_CARD_URL_TEMPLATE
)

// CreateKanbanCard creates a new Kanban card on the specified board by sending a POST request.
// It takes the board ID and a Kanban card object as parameters, and returns the created card wrapped in a ResponseResult,
// along with an error if the request fails.
func (td *TestData) CreateKanbanCard(boardId int, kanbanCard card.Card) (ResponseResult[card.Card], error) {
	return MakePostRequest[card.Card, card.Card](fmt.Sprintf(POST_KANBAN_CARD_URL_TEMPLATE, td.GetAppUrl(), boardId), kanbanCard)
}

// CreateKanbanCards creates multiple Kanban cards on the specified board.
// It takes a boardId and a slice of card.Card objects to be created.
// Returns a slice of ResponseResult[card.Card] containing the results for each card creation,
// or an error if any card creation fails.
func (td *TestData) CreateKanbanCards(boardId int, kanbanCards []card.Card) ([]ResponseResult[card.Card], error) {
	responseResults := make([]ResponseResult[card.Card], len(kanbanCards))
	for idx, kanbanCard := range kanbanCards {
		responseResult, err := td.CreateKanbanCard(boardId, kanbanCard)
		if err != nil {
			return []ResponseResult[card.Card]{}, err
		}
		responseResults[idx] = responseResult
	}
	return responseResults, nil
}

// GetKanbanCardsForBoard retrieves all Kanban cards for the specified board ID.
// It sends a GET request to the application's Kanban board cards endpoint and returns
// a ResponseResult containing a slice of card.Card objects, along with any error encountered.
//
// Parameters:
//   - boardId: The ID of the Kanban board to fetch cards for.
//
// Returns:
//   - ResponseResult[[]card.Card]: The result containing the list of Kanban cards.
//   - error: Any error encountered during the request.
func (td *TestData) GetKanbanCardsForBoard(boardId int) (ResponseResult[[]card.Card], error) {
	return MakeGetRequest[[]card.Card](fmt.Sprintf(GET_KANBAN_BOARD_CARDS_TEMPLATE, td.GetAppUrl(), boardId))
}

// GetKanbanCardById retrieves a Kanban card by its ID from the specified board.
// It sends a GET request to the application's API using the provided boardId and cardId.
// Returns a ResponseResult containing the card.Card data and an error if the request fails.
func (td *TestData) GetKanbanCardById(boardId int, cardId int) (ResponseResult[card.Card], error) {
	return MakeGetRequest[card.Card](fmt.Sprintf(GET_KANBAN_CARD_URL_TEMPLATE, td.GetAppUrl(), boardId, cardId))
}

// UpdateKanbanCard updates an existing Kanban card on the specified board.
// It sends a PUT request to the Kanban card API endpoint with the updated card data.
// Parameters:
//   - boardId: the ID of the Kanban board containing the card.
//   - cardId: the ID of the card to update.
//   - updatedCard: the updated card data.
//
// Returns:
//   - ResponseResult[struct{}]: the result of the update operation.
//   - error: any error encountered during the request.
func (td *TestData) UpdateKanbanCard(boardId int, cardId int, updatedCard card.Card) (ResponseResult[struct{}], error) {
	return MakePutRequest[card.Card, struct{}](fmt.Sprintf(PUT_KANBAN_CARD_URL_TEMPLATE, td.GetAppUrl(), boardId, cardId), updatedCard)
}

// DeleteKanbanCard deletes a Kanban card identified by cardId from the specified boardId.
// It sends a DELETE request to the application's Kanban card endpoint.
// Returns a ResponseResult containing any response data and an error if the request fails.
func (td *TestData) DeleteKanbanCard(boardId int, cardId int) (ResponseResult[any], error) {
	return MakeDeleteRequest[any](fmt.Sprintf(DELETE_KANBAN_CARD_URL_TEMPLATE, td.GetAppUrl(), boardId, cardId))
}
