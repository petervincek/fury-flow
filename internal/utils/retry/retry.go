package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrorRetryTimeoutExceeded     = errors.New("retry timeout exceeded")
	ErrorMaxRetryAttemptsExceeded = errors.New("max retry attempts exceeded")
	ErrorInvalidRetrySettings     = errors.New("retry settings are invalid")
)

// BackoffStrategy defines a function type that takes the current delay duration
// and returns the next delay duration to be used in a retry mechanism.
// This allows for customizable backoff strategies such as exponential, linear, or constant delays.
type BackoffStrategy func(time.Duration) time.Duration

// ExponentialBackoffStrategy is a BackoffStrategy implementation that doubles the current duration
// on each retry attempt, resulting in exponential backoff. This strategy helps to progressively
// increase the wait time between retries, which can be useful for handling transient errors or
// rate limiting scenarios.
var ExponentialBackoffStrategy BackoffStrategy = func(current time.Duration) time.Duration {
	return current * 2
}

// Linear backoff: increases by a fixed step each time
func LinearBackoffStrategy(step time.Duration) BackoffStrategy {
	return func(current time.Duration) time.Duration {
		return current + step
	}
}

// Constant backoff: always returns the same duration
func ConstantBackoffStrategy(constant time.Duration) BackoffStrategy {
	return func(current time.Duration) time.Duration {
		return constant
	}
}

// Fibonacci backoff: increases according to the Fibonacci sequence
func FibonacciBackoffStrategy() BackoffStrategy {
	var prev time.Duration = 0
	return func(current time.Duration) time.Duration {
		next := current + prev
		if next == 0 {
			next = current
		}
		prev = current
		return next
	}
}

// Retry defines a function type that executes the provided function fn with retry logic.
// It accepts a context for cancellation and returns an error if all retry attempts fail.
type Retry func(ctx context.Context, fn func() error) error

// RetryWithBackoff attempts to execute the provided function fn up to maxAttempts times,
// applying an exponential backoff strategy between attempts. The initial backoff duration
// is specified by initialBackoff, and will not exceed maxBackoff. The backoff duration
// for each attempt is determined by the provided backoffStrategy function.
//
// The retry process will not exceed the specified timeout duration, and will be cancelled
// if the provided context is done. If fn succeeds (returns nil), RetryWithBackoff returns nil.
// If the timeout is exceeded, ErrorRetryTimeoutExceeded is returned. If the maximum number
// of attempts is reached without success, ErrorMaxRetryAttemptsExceeded is returned, wrapping
// the last error returned by fn.
//
// Parameters:
//   - ctx: Context for cancellation.
//   - fn: Function to execute and retry on error.
//   - initialBackoff: Initial backoff duration.
//   - maxBackoff: Maximum backoff duration.
//   - backoffStrategy: Function to calculate next backoff duration.
//   - timeout: Maximum total duration for all attempts.
//   - maxAttempts: Maximum number of retry attempts.
//
// Returns:
//   - error: nil if fn succeeds, or an error indicating timeout or max attempts exceeded.
func RetryWithBackoff(ctx context.Context, fn func() error, initialBackoff, maxBackoff time.Duration,
	backoffStrategy BackoffStrategy, timeout time.Duration, maxAttempts int) error {

	deadline := time.Now().Add(timeout)
	backoff := initialBackoff

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if time.Now().Add(backoff).After(deadline) {
			return ErrorRetryTimeoutExceeded
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// continue to next attempt
		}

		backoff = backoffStrategy(backoff)
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	return fmt.Errorf("%w: %v", ErrorMaxRetryAttemptsExceeded, lastErr)
}

// RetryBuilder is a configuration struct for building retry logic.
// It allows customization of initial and maximum backoff durations,
// selection of a backoff strategy, overall timeout, and the maximum
// number of retry attempts.
type RetryBuilder struct {
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	backoffStrategy BackoffStrategy
	timeout         time.Duration
	maxAttempts     int
}

// NewBuilder creates and returns a new instance of RetryBuilder.
// This function is typically used to initialize a retry builder for configuring retry logic.
func NewBuilder() *RetryBuilder {
	return &RetryBuilder{}
}

// SetInitialBackoff sets the initial backoff duration for the retry mechanism.
// This value determines the amount of time to wait before the first retry attempt.
// Returns the RetryBuilder to allow for method chaining.
func (rb *RetryBuilder) SetInitialBackoff(initialBackoff time.Duration) *RetryBuilder {
	rb.initialBackoff = initialBackoff
	return rb
}

// SetMaxBackoff sets the maximum backoff duration for retries.
// This limits the exponential backoff to the specified maxBackoff value.
// Returns the RetryBuilder to allow method chaining.
func (rb *RetryBuilder) SetMaxBackoff(maxBackoff time.Duration) *RetryBuilder {
	rb.maxBackoff = maxBackoff
	return rb
}

// SetBackoffStrategy sets the backoff strategy to be used by the RetryBuilder.
// It allows customizing how retries are spaced out in case of failures.
// Returns the updated RetryBuilder for method chaining.
func (rb *RetryBuilder) SetBackoffStrategy(backoffStrategy BackoffStrategy) *RetryBuilder {
	rb.backoffStrategy = backoffStrategy
	return rb
}

// SetTimeout sets the timeout duration for the retry operation.
// It returns the RetryBuilder to allow for method chaining.
func (rb *RetryBuilder) SetTimeout(timeout time.Duration) *RetryBuilder {
	rb.timeout = timeout
	return rb
}

// SetMaxAttempts sets the maximum number of retry attempts for the RetryBuilder.
// It returns the RetryBuilder to allow for method chaining.
func (rb *RetryBuilder) SetMaxAttempts(maxAttempts int) *RetryBuilder {
	rb.maxAttempts = maxAttempts
	return rb
}

// Build validates the RetryBuilder configuration and constructs a Retry function.
// It returns an error if any of the required settings are invalid, such as non-positive
// initialBackoff, maxBackoff less than initialBackoff, unset backoffStrategy, non-positive
// timeout, or non-positive maxAttempts. On success, it returns a Retry function that
// executes the provided operation with the configured retry and backoff strategy.
func (rb *RetryBuilder) Build() (Retry, error) {
	// validate
	if rb.initialBackoff <= 0 {
		return nil, fmt.Errorf("%w: initialBackoff must be greater than zero", ErrorInvalidRetrySettings)
	}
	if rb.maxBackoff < rb.initialBackoff {
		return nil, fmt.Errorf("%w: maxBackoff must be greater than or equal to initialBackoff", ErrorInvalidRetrySettings)
	}
	if rb.backoffStrategy == nil {
		return nil, fmt.Errorf("%w: backoffStrategy must be set", ErrorInvalidRetrySettings)
	}
	if rb.timeout <= 0 {
		return nil, fmt.Errorf("%w: timeout must be greater than zero", ErrorInvalidRetrySettings)
	}
	if rb.maxAttempts <= 0 {
		return nil, fmt.Errorf("%w: maxAttempts must be greater than zero", ErrorInvalidRetrySettings)
	}
	return func(ctx context.Context, fn func() error) error {
		return RetryWithBackoff(ctx, fn, rb.initialBackoff, rb.maxBackoff,
			rb.backoffStrategy, rb.timeout, rb.maxAttempts)
	}, nil
}

// Default returns a Retry instance configured with sensible default values:
// initial backoff of 100ms, maximum backoff of 2s, exponential backoff strategy,
// timeout of 10s, and a maximum of 5 attempts.
func Default() Retry {
	retry, _ := NewBuilder().
		SetInitialBackoff(100 * time.Millisecond).
		SetMaxBackoff(2 * time.Second).
		SetBackoffStrategy(ExponentialBackoffStrategy).
		SetTimeout(10 * time.Second).
		SetMaxAttempts(5).
		Build()
	return retry
}
