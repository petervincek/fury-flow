package app

import (
	"fmt"

	"github.com/petervincek/fury-flow/internal/api/board"
	"github.com/petervincek/fury-flow/internal/api/healthcheck"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/service"
)

const (
	BOARD_ID_PARAM = "boardId"
)

var (
	boardIdParamRouteContext = fmt.Sprintf("/:%s", BOARD_ID_PARAM)
)

// setupRoutes configures the HTTP routes for the FuryFlow application.
// It sets up endpoints for health checks, readiness checks, and Kanban board management.
// The routes include:
//   - GET /health: Returns the application's health status.
//   - GET /ready: Returns the readiness status based on dependencies.
//   - /boards group: Provides CRUD operations for Kanban boards, including:
//   - GET /boards: List all boards.
//   - GET /boards/:id: Retrieve a board by ID.
//   - POST /boards: Create a new board.
//   - PUT /boards/:id: Update a board by ID.
//   - DELETE /boards/:id: Delete a board by ID.
func setupRoutes(ff *FuryFlow) {

	// define the routes for healtcheck and ready(dependency readiness) endpoints
	hch := healthcheck.New(ff.pool)
	ff.app.Get("/health", hch.HealthCheck)
	ff.app.Get("/ready", hch.Ready)

	// define routes to manage the Kanban Boards (the top level namespace for everything else)
	bh := board.New(service.NewKanbanBoardService(db.New(ff.pool)))
	boardsGroup := ff.app.Group("/boards")
	boardsGroup.Get("/", bh.GetBoards)
	boardsGroup.Get(boardIdParamRouteContext, bh.GetBoardById)
	boardsGroup.Post("/", bh.CreateBoard)
	boardsGroup.Put(boardIdParamRouteContext, bh.UpdateBoardById)
	boardsGroup.Delete(boardIdParamRouteContext, bh.DeleteBoardById)
}
