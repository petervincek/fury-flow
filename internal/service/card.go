package service

import (
	"context"
	"errors"

	"github.com/petervincek/fury-flow/internal/db"
)

var (
	ErrorKanbanCardNotUpdated           = errors.New("kanban card not updated")
	ErrorKanbanCardDependencyNotCreated = errors.New("kanban card dependency not created")
)

type KanbanCardService struct {
	q *db.Queries
}

// NewKanbanCardService creates and returns a new instance of KanbanCardService using the provided database queries.
// It initializes the service with the given db.Queries, allowing interaction with the card-related database operations.
func NewKanbanCardService(q *db.Queries) *KanbanCardService {
	return &KanbanCardService{
		q: q,
	}
}

// GetKanbanCards retrieves all Kanban cards from the database.
// It returns a slice of KanbanCard objects and an error, if any occurs during retrieval.
func (kcs *KanbanCardService) GetKanbanCards(ctx context.Context) ([]db.KanbanCard, error) {
	return kcs.q.GetKanbanCards(ctx)
}

// GetKanbanCardsForBoard retrieves all Kanban cards associated with the specified board ID.
// It returns a slice of KanbanCard objects and an error if the operation fails.
//
// Parameters:
//
//	ctx     - The context for controlling cancellation and deadlines.
//	boardId - The ID of the Kanban board to fetch cards for.
//
// Returns:
//
//	([]db.KanbanCard, error) - A slice of KanbanCard and an error, if any.
func (kcs *KanbanCardService) GetKanbanCardsForBoard(ctx context.Context, boardId int32) ([]db.KanbanCard, error) {
	return kcs.q.GetKanbanCardsForBoard(ctx, boardId)
}

// GetKanbanCardById retrieves a Kanban card from the database by its unique identifier.
// It takes a context for request-scoped values and cancellation, and the card's ID as an int32.
// Returns the corresponding KanbanCard and an error if the operation fails.
func (kcs *KanbanCardService) GetKanbanCardById(ctx context.Context, cardId int32) (db.KanbanCard, error) {
	return kcs.q.GetKanbanCardById(ctx, cardId)
}

// CreateKanbanCard creates a new Kanban card in the database using the provided parameters.
// It returns the created KanbanCard and any error encountered during the operation.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - card: The parameters required to create a new Kanban card.
//
// Returns:
//   - db.KanbanCard: The newly created Kanban card.
//   - error: An error if the creation fails, otherwise nil.
func (kcs *KanbanCardService) CreateKanbanCard(ctx context.Context, card db.CreateKanbanCardParams) (db.KanbanCard, error) {
	return kcs.q.CreateKanbanCard(ctx, card)
}

// UpdateKanbanCard updates a Kanban card in the database using the provided parameters.
// It returns an error if the update operation fails or if no rows were affected.
// If the update is successful, it returns nil.
//
// Parameters:
//   - ctx: The context for controlling cancellation and timeouts.
//   - card: The parameters required to update the Kanban card.
//
// Returns:
//   - error: An error if the update fails or no card was updated; nil otherwise.
func (kcs *KanbanCardService) UpdateKanbanCard(ctx context.Context, card db.UpdateKanbanCardParams) error {
	commandTag, err := kcs.q.UpdateKanbanCard(ctx, card)
	if err != nil {
		return err
	}
	rowsAffected := commandTag.RowsAffected()
	if rowsAffected != 1 {
		return ErrorKanbanCardNotUpdated
	}
	return nil
}

// DeleteKanbanCard deletes a Kanban card identified by the given cardId.
// It returns an error if the deletion fails.
//
// Parameters:
//
//	ctx    - The context for controlling cancellation and deadlines.
//	cardId - The unique identifier of the Kanban card to be deleted.
//
// Returns:
//
//	error - An error if the card could not be deleted, otherwise nil.
func (kcs *KanbanCardService) DeleteKanbanCard(ctx context.Context, cardId int32) error {
	return kcs.q.DeleteKanbanCard(ctx, cardId)
}

// KanbanCardExists checks if a Kanban card with the specified cardId exists in the database.
// It returns true if the card exists, false otherwise, along with any error encountered during the operation.
func (kcs *KanbanCardService) KanbanCardExists(ctx context.Context, cardId int32) (bool, error) {
	return kcs.q.KanbanCardExists(ctx, cardId)
}

// CreateCardDependency creates a dependency between two Kanban cards.
// It associates the card identified by cardId to depend on the card identified by dependencyCardId.
// Returns an error if the dependency could not be created or if the database operation fails.
func (kcs *KanbanCardService) CreateCardDependency(ctx context.Context, cardId int32, dependencyCardId int32) error {
	commandTag, err := kcs.q.CreateCardDependency(ctx, db.CreateCardDependencyParams{
		CardID:          cardId,
		DependsOnCardID: dependencyCardId,
	})
	if err != nil {
		return err
	}
	rowsAffected := commandTag.RowsAffected()
	if rowsAffected != 1 {
		return ErrorKanbanCardDependencyNotCreated
	}
	return nil
}

// CreateCardDependencies creates dependencies for a given Kanban card by associating it with other card IDs.
// It takes a context, the card ID, and a slice of dependency card IDs.
// Returns an error if the operation fails or if not all dependencies are created successfully.
func (kcs *KanbanCardService) CreateCardDependencies(ctx context.Context, cardId int32, dependencyCardIds []int32) error {
	commandTag, err := kcs.q.CreateCardDependencies(ctx, db.CreateCardDependenciesParams{
		CardID:  cardId,
		Column2: dependencyCardIds,
	})
	if err != nil {
		return err
	}
	rowsAffected := commandTag.RowsAffected()
	if rowsAffected != int64(len(dependencyCardIds)) {
		return ErrorKanbanCardDependencyNotCreated
	}
	return nil
}

// GetCardDependencyIds retrieves the IDs of cards that the specified card depends on.
// It takes a context and the card's ID as parameters, and returns a slice of dependent card IDs or an error.
func (kcs *KanbanCardService) GetCardDependencyIds(ctx context.Context, cardId int32) ([]int32, error) {
	return kcs.q.GetCardDependencyIds(ctx, cardId)
}

// GetCardDependencies retrieves the dependencies of a Kanban card specified by cardId.
// It returns a slice of KanbanCard objects representing the dependent cards,
// or an error if the operation fails.
func (kcs *KanbanCardService) GetCardDependencies(ctx context.Context, cardId int32) ([]db.KanbanCard, error) {
	return kcs.q.GetCardDependencies(ctx, cardId)
}

// DeleteCardDependency removes the dependency relationship between a card and another card in the Kanban system.
// It takes the context, the ID of the card, and the ID of the dependency card to be removed.
// Returns an error if the operation fails.
func (kcs *KanbanCardService) DeleteCardDependency(ctx context.Context, cardId int32, dependencyCardId int32) error {
	return kcs.q.DeleteCardDependency(ctx, db.DeleteCardDependencyParams{
		CardID:          cardId,
		DependsOnCardID: dependencyCardId,
	})
}

// DeleteCardDependencies removes all dependencies associated with the specified card.
// It takes a context for cancellation and timeout control, and the card's unique identifier.
// Returns an error if the operation fails.
func (kcs *KanbanCardService) DeleteCardDependencies(ctx context.Context, cardId int32) error {
	return kcs.q.DeleteCardDependencies(ctx, cardId)
}
