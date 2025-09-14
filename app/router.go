package app

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/card"
	"github.com/petervincek/fury-flow/internal/api/healthcheck"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/service"
)

const (
	BOARD_ID_PARAM           = "boardId"
	CARD_ID_PARAM            = "cardId"
	CARD_DEPENDENCY_ID_PARAM = "cardDependencyId"
)

var (
	boardIdParamRouteContext          = fmt.Sprintf("/:%s", BOARD_ID_PARAM)
	cardIdParamRouteContext           = fmt.Sprintf("/:%s", CARD_ID_PARAM)
	cardIdDependencyParamRouteContext = fmt.Sprintf("/:%s", CARD_DEPENDENCY_ID_PARAM)
)

// setupRoutes configures all HTTP routes for the FuryFlow application.
// It sets up endpoints for health checks, Kanban board management, Kanban card management,
// and card dependency management. The routes are organized into groups for boards, cards,
// and card dependencies, providing RESTful operations for each resource.
//
// Routes configured:
//   - /health: Health check endpoint.
//   - /ready: Readiness check endpoint.
//   - /boards: CRUD operations for Kanban boards.
//   - /boards/:boardId/cards: CRUD operations for cards within a board.
//   - /boards/:boardId/cards/:cardId/dependencies: Manage dependencies between cards.
//
// Parameters:
//
//	ff *FuryFlow: The main application context containing dependencies and the router.
func setupRoutes(ff *FuryFlow) {

	// define the routes for healtcheck and ready(dependency readiness) endpoints
	hch := healthcheck.New(ff.pool)
	ff.app.Get("/health", hch.HealthCheck)
	ff.app.Get("/ready", hch.Ready)

	// define routes to manage the Kanban Boards (the top level namespace for everything else)
	kbs := service.NewKanbanBoardService(db.New(ff.pool))
	bh := board.New(kbs)
	boardsGroup := ff.app.Group("/boards")
	boardsGroup.Get("/", bh.GetBoards)
	boardsGroup.Get(boardIdParamRouteContext, bh.GetBoardById)
	boardsGroup.Post("/", bh.CreateBoard)
	boardsGroup.Put(boardIdParamRouteContext, bh.UpdateBoardById)
	boardsGroup.Delete(boardIdParamRouteContext, bh.DeleteBoardById)

	// define routes to manage the Kanban Cards (individual tasks and unit of work in Kanban board)
	kcs := service.NewKanbanCardService(db.New(ff.pool))
	ch := card.New(kbs, kcs)
	cardsGroup := boardsGroup.Group(fmt.Sprintf("%s/cards", boardIdParamRouteContext))
	cardsGroup.Get("/", ch.GetCardsForBoard)
	cardsGroup.Get(cardIdParamRouteContext, ch.GetCardById)
	cardsGroup.Post("/", ch.CreateCard)
	cardsGroup.Put(cardIdParamRouteContext, ch.UpdateCardById)
	cardsGroup.Delete(cardIdParamRouteContext, ch.DeleteCardById)

	// define additional routes to manage dependencies between individual card in the board
	cardDependenciesGroup := cardsGroup.Group(fmt.Sprintf("%s/dependencies", cardIdParamRouteContext))
	cardDependenciesGroup.Get("/", ch.GetCardDependencies)                                   // retrieve all dependencies for a given card
	cardDependenciesGroup.Post("/", ch.CreateCardDependencies)                               // create dependency relationships between cards
	cardDependenciesGroup.Delete("/", ch.DeleteCardDependencies)                             // delete/remove all dependencies for a given card
	cardDependenciesGroup.Delete(cardIdDependencyParamRouteContext, ch.DeleteCardDependency) // delete/remove specific provided dependency
}
