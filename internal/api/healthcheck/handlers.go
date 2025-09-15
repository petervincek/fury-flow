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
//
// It responds with HTTP 200 OK status to indicate that the service is running.
//
// @Summary      Health Check
// @Description  Returns 200 OK if the service is healthy.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func (hch *HealthCheckHandler) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

// Ready handles the readiness check for the service.
// It pings the database connection pool to verify connectivity.
// If the database is reachable, it responds with HTTP 200 OK.
// If the database is not reachable, it responds with HTTP 503 Service Unavailable.
// Ready checks if the application is ready to serve requests by pinging the database.
// If the database is reachable, it returns HTTP 200 OK; otherwise, it returns HTTP 503 Service Unavailable.
//
// @Summary      Readiness check
// @Description  Checks if the application is ready to serve requests by verifying database connectivity.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Failure      503  {string}  string  "Service Unavailable"
// @Router       /health/ready [get]
func (hch *HealthCheckHandler) Ready(ctx *fiber.Ctx) error {
	if err := hch.DbPool.Ping(ctx.Context()); err != nil {
		return ctx.SendStatus(fiber.StatusServiceUnavailable)
	}
	return ctx.SendStatus(fiber.StatusOK)
}
