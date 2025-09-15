package app

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/attachment"
	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/card"
	"github.com/petervincek/fury-flow/internal/api/comment"
	"github.com/petervincek/fury-flow/internal/api/healthcheck"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/service"
)

const (
	BOARD_ID_PARAM           = "boardId"
	CARD_ID_PARAM            = "cardId"
	CARD_DEPENDENCY_ID_PARAM = "cardDependencyId"
	COMMENT_ID_PARAM         = "commentId"
	ATTACHMENT_ID_PARAM      = "attachmentId"
)

var (
	boardIdParamRouteContext          = fmt.Sprintf("/:%s", BOARD_ID_PARAM)
	cardIdParamRouteContext           = fmt.Sprintf("/:%s", CARD_ID_PARAM)
	cardIdDependencyParamRouteContext = fmt.Sprintf("/:%s", CARD_DEPENDENCY_ID_PARAM)
	commentIdParamRouteContext        = fmt.Sprintf("/:%s", COMMENT_ID_PARAM)
	attachmentIdParamRouteContext     = fmt.Sprintf("/:%s", ATTACHMENT_ID_PARAM)
)

// setupRoutes configures all HTTP routes for the FuryFlow application.
//
// It sets up endpoints for:
//   - Health and readiness checks.
//   - Kanban board management (CRUD operations).
//   - Kanban card management within boards (CRUD operations).
//   - Managing dependencies between cards (create, retrieve, delete).
//   - Managing comments for individual cards (create, retrieve, delete).
//   - Managing attachments for individual cards (create, retrieve, delete).
//
// Each route is grouped logically under relevant namespaces (e.g., /boards, /boards/:boardId/cards).
// Handlers are constructed using services for boards and cards, which interact with the database pool.
// Some comment-related endpoints are placeholders and require implementation.
//
// Routes configured:
//   - /health: Health check endpoint.
//   - /ready: Readiness check endpoint.
//   - /boards: CRUD operations for Kanban boards.
//   - /boards/:boardId/cards: CRUD operations for cards within a board.
//   - /boards/:boardId/cards/:cardId/dependencies: Manage dependencies between cards.
//   - /boards/:boardId/cards/:cardId/comments: Manage comments for cards.
//   - /boards/:boardId/cards/:cardId/attachments: Manage attachments for cards.
//
// Parameters:
//
//	ff *FuryFlow: The main application context containing dependencies and the router.
func setupRoutes(ff *FuryFlow) {

	// define the routes for healtcheck and ready(dependency readiness) endpoints
	hch := healthcheck.New(ff.pool)
	ff.app.Get("/health", hch.HealthCheck)
	ff.app.Get("/ready", hch.Ready)

	// create integrity service
	kis := service.NewKanbanIntegrityService(db.New(ff.pool))

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

	// define additional routes to manage comments for individual card in the board
	commentService := service.NewKanbanCommentService(db.New(ff.pool))
	commentHandler := comment.New(kbs, kcs, commentService, kis)
	cardCommentsGroup := cardsGroup.Group(fmt.Sprintf("%s/comments", cardIdParamRouteContext))
	cardCommentsGroup.Get("/", commentHandler.GetCommentsForCard)                      // retrieve all comments for a given card
	cardCommentsGroup.Get(commentIdParamRouteContext, commentHandler.GetCommentById)   // retrieve comment by id
	cardCommentsGroup.Post("/", commentHandler.CreateComment)                          // create comment for individual card
	cardCommentsGroup.Delete("/", commentHandler.DeleteCommentsForCard)                // delete/remove all comments for a given card
	cardCommentsGroup.Delete(commentIdParamRouteContext, commentHandler.DeleteComment) // delete/remove specific comment for a given card

	// define additional routes to manage attachments for individual card in the board
	attachmentService := service.NewKanbanAttachmentService(db.New(ff.pool))
	attachmentHandler := attachment.New(kbs, kcs, attachmentService, kis)
	cardAttachmentsGroup := cardsGroup.Group(fmt.Sprintf("%s/attachments", cardIdParamRouteContext))
	cardAttachmentsGroup.Get("/", attachmentHandler.GetAttachmentsForCard)                         // retrieve all attachments for a given card
	cardAttachmentsGroup.Get(attachmentIdParamRouteContext, attachmentHandler.GetAttachmentById)   // retrieve attachment by id
	cardAttachmentsGroup.Post("/", attachmentHandler.CreateAttachment)                             // create attachment for individual card
	cardAttachmentsGroup.Delete("/", attachmentHandler.DeleteAttachmentsForCard)                   // delete/remove all attachments for a given card
	cardAttachmentsGroup.Delete(attachmentIdParamRouteContext, attachmentHandler.DeleteAttachment) // delete/remove specific attachment for a given card
}
