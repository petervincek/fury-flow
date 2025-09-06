package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/petervincek/fury-flow/internal/logging"
	"github.com/petervincek/fury-flow/internal/utils/retry"
)

var (
	Retry                  = retry.Default()
	ErrorServiceNotHealthy = errors.New("service is not healthy")
	ErrorServiceNotReady   = errors.New("service is not ready")
	logger                 = logging.GetLogger()
)

// WaitUntilHealthyAndReady repeatedly checks the health and readiness endpoints of a local service
// running on the port provided by getPort. It first verifies the "/health" endpoint returns a 200 OK
// status, then checks the "/ready" endpoint for the same. Both checks are retried using the Retry
// function until successful or the context is cancelled. Returns an error if either endpoint does not
// return a healthy status or if any request fails.
func WaitUntilHealthyAndReady(ctx context.Context, getPort func() int) error {
	logger.Info("Waiting until app is healthy and ready to server requests")
	// healthcheck request
	err := Retry(ctx, func() error {
		logger.Info("Checking Health status\n")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", getPort()))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		logger.Info(fmt.Sprintf("Health response status code: %d\n", resp.StatusCode))
		if resp.StatusCode != 200 {
			return ErrorServiceNotHealthy
		}
		return nil
	})
	if err != nil {
		return err
	}

	// readiness check request
	err = Retry(ctx, func() error {
		logger.Info("Checking Ready status\n")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ready", getPort()))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		logger.Info(fmt.Sprintf("Ready response status code: %d\n", resp.StatusCode))
		if resp.StatusCode != 200 {
			return ErrorServiceNotReady
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
