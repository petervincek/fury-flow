package healthcheck

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthCheckHandler provides HTTP handlers for health check endpoints.
// It holds a reference to a database connection pool used to verify database connectivity.
type HealthCheckHandler struct {
	DbPool *pgxpool.Pool
}

// New creates and returns a new instance of HealthCheckHandler using the provided pgxpool.Pool.
// It initializes the handler with the given database connection pool.
func New(dbPool *pgxpool.Pool) *HealthCheckHandler {
	return &HealthCheckHandler{
		DbPool: dbPool,
	}
}

// HealthCheck handles the health check endpoint.
// It responds with HTTP 200 OK status to indicate that the service is running.
func (hch *HealthCheckHandler) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

// Ready handles the readiness check for the service.
// It pings the database connection pool to verify connectivity.
// If the database is reachable, it responds with HTTP 200 OK.
// If the database is not reachable, it responds with HTTP 503 Service Unavailable.
func (hch *HealthCheckHandler) Ready(ctx *fiber.Ctx) error {
	if err := hch.DbPool.Ping(ctx.Context()); err != nil {
		return ctx.SendStatus(fiber.StatusServiceUnavailable)
	}
	return ctx.SendStatus(fiber.StatusOK)
}
