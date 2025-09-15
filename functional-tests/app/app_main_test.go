//go:build functional
// +build functional

package app_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petervincek/fury-flow/app"
	teststartupmodel "github.com/petervincek/fury-flow/functional-tests/app/test-startup-model"
	"github.com/petervincek/fury-flow/functional-tests/utils"
	"github.com/petervincek/fury-flow/functional-tests/utils/testdata"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const (
	POSTGRES_DOCKER_IMAGE = "postgres:16"
	DB_DRIVER             = "postgres"
	POSTGRES_USER         = "admin"
	POSTGRES_PASSWORD     = "password"
	POSTGRES_DB_NAME      = "furryflow"
)

var (
	// TODO: find a better way how to get and share the dynamic port for app instance started for functional tests
	TEST_APP_SERVER_PORT = 0
	Pool                 *pgxpool.Pool
	TestDataCreator      testdata.TestData
	logger               = logging.GetLogger()
)

// startPostgresqlContainer starts a PostgreSQL Docker container for testing purposes using testcontainers,
// sets up environment variables for database access and migrations, runs Goose migrations, and returns the container instance.
// It panics on any error encountered during setup or migration execution.
func startPostgresqlContainer() testcontainers.Container {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        POSTGRES_DOCKER_IMAGE,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     POSTGRES_USER,
			"POSTGRES_PASSWORD": POSTGRES_PASSWORD,
			"POSTGRES_DB":       POSTGRES_DB_NAME,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	utils.PanicOnError(err)

	// expose the postgresql service
	host, err := postgresContainer.Host(ctx)
	utils.PanicOnError(err)
	port, err := postgresContainer.MappedPort(ctx, nat.Port("5432"))
	utils.PanicOnError(err)
	fmt.Printf("[PostgreSQL] Host: %s, Port: %d\n", host, port.Int())
	// expose env variables related to PostgreSQL docker container
	utils.PanicOnError(os.Setenv(app.DB_DRIVER, DB_DRIVER))
	utils.PanicOnError(os.Setenv(app.DB_HOST, host))
	utils.PanicOnError(os.Setenv(app.DB_PORT, strconv.Itoa(port.Int())))
	utils.PanicOnError(os.Setenv(app.DB_USERNAME, POSTGRES_USER))
	utils.PanicOnError(os.Setenv(app.DB_PASSWORD, POSTGRES_PASSWORD))
	utils.PanicOnError(os.Setenv(app.DB_NAME, POSTGRES_DB_NAME))
	// expose env variables related to PostgreSQL and the Goose migrations
	utils.PanicOnError(os.Setenv(app.GOOSE_DRIVER, DB_DRIVER))
	utils.PanicOnError(os.Setenv(app.GOOSE_DBSTRING,
		fmt.Sprintf("%s://%s:%s@%s:%d/%s", DB_DRIVER, POSTGRES_USER, POSTGRES_PASSWORD, host, port.Int(), POSTGRES_DB_NAME)))
	utils.PanicOnError(os.Setenv(app.GOOSE_MIGRATION_DIR, "../../migrations"))
	utils.PanicOnError(os.Setenv(app.GOOSE_TABLE, "goose_migrations"))

	// run the DB migrations via Goose
	cmd := exec.Command("../../bin/goose", "up") // the location of the goose command is relative to the location of this file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	utils.PanicOnError(err)

	logger.Info("[Functional Test]: PostgreSQL container started successfully")
	return postgresContainer
}

// startFuryFlowApp initializes and starts a FuryFlow application instance in a separate goroutine.
// It waits until the application is healthy and ready before returning the instance.
// The function also sets the global TEST_APP_SERVER_PORT variable to the application's dynamic port.
// If the application fails to start or become healthy, the function logs a fatal error and terminates the process.
func startFuryFlowApp() *app.FuryFlow {
	// create a instance of the application
	app := app.NewApp()
	// start the application in separate goroutine
	go func() {
		err := app.StartApp()
		if err != nil {
			logger.Fatal("error while starting FuryFlow app in functioal tests", zap.Error(err))
		}
	}()
	// get the dynamic port of the started application
	TEST_APP_SERVER_PORT = app.GetAppPort()
	Pool = app.GetPool()
	TestDataCreator = testdata.New(app)
	// wait until healthy and ready (with retry and some definite timeout)
	err := utils.WaitUntilHealthyAndReady(context.Background(), func() int { return app.GetAppPort() })
	if err != nil {
		logger.Fatal("error while waiting for health check in functional tests", zap.Error(err))
	}
	logger.Info("[Functional Test]: FuryFlow app instance started successfully on port", zap.Int("port", TEST_APP_SERVER_PORT))
	return app
}

// startLocalInfra initializes and starts the required local infrastructure for testing.
// It first starts necessary containers (e.g., PostgreSQL) that the application depends on,
// then creates and returns a TestInfra instance with the running application and containers.
func startLocalInfra() *teststartupmodel.TestInfra {
	// we need to start containers first, because app depends on the containers (services)
	containers := []testcontainers.Container{startPostgresqlContainer()}
	return teststartupmodel.New(startFuryFlowApp(), containers)
}

// stopLocalInfra gracefully shuts down the local testing infrastructure.
// It first shuts down the FuryFlow application, then iterates over all test containers,
// attempting to terminate each one. If any container fails to terminate, an error is logged
// and the process exits with a non-zero status code.
func stopLocalInfra(testInfra *teststartupmodel.TestInfra) {
	// first shutdown the furryflow app
	testInfra.FuryFlowApp.ShutdownApp(5 * time.Second)

	// then release the containers
	containers := testInfra.Containers
	ctx := context.Background()
	var terminatingErr error
	for _, container := range containers {
		var containerName string
		containerId := container.GetContainerID()
		inspect, err := container.Inspect(ctx)
		if err != nil {
			logger.Error("error while inspecting container with ID", zap.Error(err), zap.String("containerId", containerId))
		} else {
			containerName = fmt.Sprintf("'%s:%s'", inspect.Config.Image, inspect.Name)
		}
		err = container.Terminate(ctx)
		if err != nil {
			logger.Error("error while terminating container", zap.Error(err), zap.String("container", utils.GetFirstNonEmpty(containerName, containerId)))
			terminatingErr = err
		}
		logger.Info("[Functional Test]: Container stopped successfully", zap.String("container", utils.GetFirstNonEmpty(containerName, containerId)))
	}
	if terminatingErr != nil {
		logger.Error("unable to stop local testing infrastructure", zap.Error(terminatingErr))
		os.Exit(1)
	}
}

// TestMain sets up and tears down local infrastructure for all package-level tests.
// It uses utils.AroundWrapperFunction to ensure startLocalInfra is called before tests
// and stopLocalInfra is called after tests, wrapping the execution of m.Run().
func TestMain(m *testing.M) {
	// run all the package level tests
	utils.AroundWrapperFunction(startLocalInfra, stopLocalInfra, func() int {
		code := m.Run()
		return code
	})()
}

// setupTest prepares the test environment by performing database cleanup before each test,
// and registers a cleanup function to run after the test completes. It ensures that the
// database is in a clean state for each test case and logs the cleanup actions.
func setupTest(t *testing.T) {
	// Before Hook - Setup Code
	logger.Debug("Running db cleanup before test")
	err := db.New(Pool).DeleteAll(t.Context())
	assert.NoError(t, err)

	// After Hook - Clean up
	t.Cleanup(func() {
		// Clean up code
		logger.Debug("Running db cleanup after test")
	})
}
