package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/petervincek/fury-flow/internal/utils/retry"
	"github.com/stretchr/testify/assert"
)

func TestRetryRetryTimeoutExceeded(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	functionToRetry := func() error {
		callCount++
		return errors.New("unable to complete")
	}

	// exercise
	err := retry.RetryWithBackoff(ctx, functionToRetry, 10*time.Millisecond, 20*time.Millisecond,
		retry.ExponentialBackoffStrategy, 40*time.Millisecond, 5)
	// verify
	assert.ErrorIs(t, err, retry.ErrorRetryTimeoutExceeded)
	assert.Equal(t, 3, callCount, "expecting function to be retried")
}

// TestRetrySuccessOnFirstAttempt verifies that the RetryWithBackoff function succeeds on the first attempt
// and does not perform unnecessary retries. It checks that the function to retry is called only once
// when it returns nil (no error), and that no error is returned from RetryWithBackoff.
func TestRetrySuccessOnFirstAttempt(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	functionToRetry := func() error {
		callCount++
		return nil
	}

	// exercise
	err := retry.RetryWithBackoff(ctx, functionToRetry, 100*time.Millisecond, 2*time.Second,
		retry.ExponentialBackoffStrategy, 10*time.Second, 5)
	// verify
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "should only call once if successful")
}

// TestRetryMaxAttemptsExceeded verifies that RetryWithBackoff returns the ErrorMaxRetryAttemptsExceeded error
// after the maximum number of retry attempts is reached, and that the function to retry is called exactly
// maxAttempts times.
func TestRetryMaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	functionToRetry := func() error {
		callCount++
		return errors.New("fail")
	}

	// exercise
	err := retry.RetryWithBackoff(ctx, functionToRetry, 10*time.Millisecond, 20*time.Millisecond,
		retry.ConstantBackoffStrategy(10*time.Millisecond), 1*time.Second, 3)
	// verify
	assert.ErrorIs(t, err, retry.ErrorMaxRetryAttemptsExceeded)
	assert.Equal(t, 3, callCount, "should call function maxAttempts times")
}

// TestRetryContextCancelled verifies that RetryWithBackoff correctly handles context cancellation.
// It sets up a retry function that cancels the context on the second invocation and always returns an error.
// The test asserts that the returned error is context.Canceled and that the function was called at least twice.
func TestRetryContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	functionToRetry := func() error {
		callCount++
		if callCount == 2 {
			cancel()
		}
		return errors.New("fail")
	}

	// exercise
	err := retry.RetryWithBackoff(ctx, functionToRetry, 10*time.Millisecond, 20*time.Millisecond,
		retry.ConstantBackoffStrategy(10*time.Millisecond), 1*time.Second, 5)
	// verify
	assert.ErrorIs(t, err, context.Canceled)
	assert.GreaterOrEqual(t, callCount, 2)
}

// TestRetryBuilderInvalidSettings verifies that the retry builder returns an error
// when provided with invalid retry settings. It checks various scenarios such as:
// - Initial backoff set to zero
// - Initial backoff greater than max backoff
// - Missing backoff strategy
// - Timeout set to zero
// - Max attempts set to zero
// In each case, the builder should fail with ErrorInvalidRetrySettings.
func TestRetryBuilderInvalidSettings(t *testing.T) {
	rb := retry.NewBuilder().
		SetInitialBackoff(0).
		SetMaxBackoff(1 * time.Second).
		SetBackoffStrategy(retry.ExponentialBackoffStrategy).
		SetTimeout(1 * time.Second).
		SetMaxAttempts(1)
	_, err := rb.Build()
	assert.ErrorIs(t, err, retry.ErrorInvalidRetrySettings)

	rb = retry.NewBuilder().
		SetInitialBackoff(100 * time.Millisecond).
		SetMaxBackoff(50 * time.Millisecond).
		SetBackoffStrategy(retry.ExponentialBackoffStrategy).
		SetTimeout(1 * time.Second).
		SetMaxAttempts(1)
	_, err = rb.Build()
	assert.ErrorIs(t, err, retry.ErrorInvalidRetrySettings)

	rb = retry.NewBuilder().
		SetInitialBackoff(100 * time.Millisecond).
		SetMaxBackoff(200 * time.Millisecond).
		SetTimeout(1 * time.Second).
		SetMaxAttempts(1)
	_, err = rb.Build()
	assert.ErrorIs(t, err, retry.ErrorInvalidRetrySettings)

	rb = retry.NewBuilder().
		SetInitialBackoff(100 * time.Millisecond).
		SetMaxBackoff(200 * time.Millisecond).
		SetBackoffStrategy(retry.ExponentialBackoffStrategy).
		SetTimeout(0).
		SetMaxAttempts(1)
	_, err = rb.Build()
	assert.ErrorIs(t, err, retry.ErrorInvalidRetrySettings)

	rb = retry.NewBuilder().
		SetInitialBackoff(100 * time.Millisecond).
		SetMaxBackoff(200 * time.Millisecond).
		SetBackoffStrategy(retry.ExponentialBackoffStrategy).
		SetTimeout(1 * time.Second).
		SetMaxAttempts(0)
	_, err = rb.Build()
	assert.ErrorIs(t, err, retry.ErrorInvalidRetrySettings)
}

// TestRetryBuilderSuccess verifies that the retry builder correctly constructs a retry function
// with the specified configuration. It checks that the function retries the operation the expected
// number of times and succeeds when the operation eventually returns no error. The test ensures
// that the retry logic respects the initial backoff, maximum backoff, linear backoff strategy,
// timeout, and maximum attempts settings.
func TestRetryBuilderSuccess(t *testing.T) {
	rb := retry.NewBuilder().
		SetInitialBackoff(10 * time.Millisecond).
		SetMaxBackoff(100 * time.Millisecond).
		SetBackoffStrategy(retry.LinearBackoffStrategy(10 * time.Millisecond)).
		SetTimeout(1 * time.Second).
		SetMaxAttempts(3)
	retryFunc, err := rb.Build()
	assert.NoError(t, err)
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 2 {
			return errors.New("fail")
		}
		return nil
	}
	err = retryFunc(context.Background(), fn)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

// TestDefaultRetry verifies that the Default retry function retries the provided
// function until it succeeds or the maximum number of attempts is reached.
// It checks that the function is called the expected number of times and that
// no error is returned when the function eventually succeeds.
func TestDefaultRetry(t *testing.T) {
	retryFunc := retry.Default()
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("fail")
		}
		return nil
	}
	err := retryFunc(context.Background(), fn)
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}
