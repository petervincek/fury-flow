package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petervincek/fury-flow/internal/utils/retry"
)

const (
	DB_DRIVER   = "DB_DRIVER"
	DB_HOST     = "DB_HOST"
	DB_PORT     = "DB_PORT"
	DB_USERNAME = "DB_USERNAME"
	DB_PASSWORD = "DB_PASSWORD"
	DB_NAME     = "DB_NAME"

	GOOSE_DRIVER        = "GOOSE_DRIVER"
	GOOSE_DBSTRING      = "GOOSE_DBSTRING"
	GOOSE_MIGRATION_DIR = "GOOSE_MIGRATION_DIR"
	GOOSE_TABLE         = "GOOSE_TABLE"

	LOG_FILE   = "LOG_FILE"
	LOG_FORMAT = "LOG_FORMAT"
	LOG_LEVEL  = "LOG_LEVEL"

	SERVER_PORT = "SERVER_PORT"
)

var (
	Retry                     = retry.Default()
	ErrorInvalidListenerState = errors.New("invalid state of listener, listener is nil")
)

// FuryFlow represents the main application structure, encapsulating the Fiber web server instance,
// a PostgreSQL connection pool, and an optional network listener for advanced server configurations.
type FuryFlow struct {
	app      *fiber.App
	pool     *pgxpool.Pool
	listener net.Listener
}

// NewApp creates and returns a new instance of FuryFlow with a Fiber app initialized.
func NewApp() *FuryFlow {
	return &FuryFlow{
		app: fiber.New(),
	}
}

// StartApp initializes the FuryFlow application by setting up a TCP listener and a PostgreSQL connection pool.
// It configures the application's routes and starts listening for incoming requests on the specified port.
// Returns an error if any step in the initialization fails.
func (ff *FuryFlow) StartApp() error {
	// create a listener
	var err error
	ff.listener, err = net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")))
	if err != nil {
		return err
	}
	// create a connection pool to PostgreSQL
	pool, err := createConnectionPool()
	if err != nil {
		return err
	}
	ff.pool = pool

	setupRoutes(ff)
	// return ff.app.Listen(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")))
	return ff.app.Listener(ff.listener)
}

// ShutdownApp gracefully shuts down the application by closing the resource pool
// and invoking the application's shutdown procedure with the specified timeout.
// It returns an error if the shutdown process fails.
//
// Parameters:
//
//	timeout - the maximum duration to wait for the shutdown to complete.
//
// Returns:
//
//	error - an error if the shutdown fails, otherwise nil.
func (ff *FuryFlow) ShutdownApp(timeout time.Duration) error {
	defer ff.pool.Close()
	return ff.app.ShutdownWithTimeout(timeout)
}

// GetAppPort returns the port number on which the application is listening.
// It ensures that the listener is properly initialized before retrieving the port.
// If the listener is not in a valid state, the function panics with the encountered error.
func (ff *FuryFlow) GetAppPort() int {
	// TODO: implement better way to make sure app is started and listener is initialized
	err := Retry(context.Background(), func() error {
		if ff.listener == nil {
			return ErrorInvalidListenerState
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return ff.listener.Addr().(*net.TCPAddr).Port
}

// GetAppHost returns the IP address of the application's TCP listener as a string.
// It retries the operation using the Retry function to ensure the listener is in a valid state.
// If the listener is not properly initialized, it panics with the encountered error.
func (ff *FuryFlow) GetAppHost() string {
	err := Retry(context.Background(), func() error {
		if ff.listener == nil {
			return ErrorInvalidListenerState
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return ff.listener.Addr().(*net.TCPAddr).IP.String()
}

// GetPool returns the pgxpool.Pool instance associated with the FuryFlow.
// This can be used to perform database operations using the connection pool.
func (ff *FuryFlow) GetPool() *pgxpool.Pool {
	return ff.pool
}

// createConnectionPool initializes and returns a new pgxpool.Pool for PostgreSQL connections.
// It reads database configuration from environment variables: DB_DRIVER, DB_USERNAME, DB_PASSWORD,
// DB_HOST, DB_PORT, and DB_NAME. The connection pool is created using these parameters.
// Returns the connection pool and any error encountered during initialization.
func createConnectionPool() (*pgxpool.Pool, error) {
	dbDriver := os.Getenv(DB_DRIVER)
	dbUsername := os.Getenv(DB_USERNAME)
	dbPassword := os.Getenv(DB_PASSWORD)
	dbHost := os.Getenv(DB_HOST)
	dbPort, err := strconv.Atoi(os.Getenv(DB_PORT))
	if err != nil {
		return nil, fmt.Errorf("unable to parse db port, err: %w", err)
	}
	dbName := os.Getenv(DB_NAME)
	dbUrl := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable", dbDriver, dbUsername, dbPassword, dbHost, dbPort, dbName)
	pool, err := pgxpool.New(context.Background(), dbUrl)
	return pool, err
}
