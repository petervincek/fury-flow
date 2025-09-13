package monad

// Result is a type alias for Either[error, T], representing a computation that may either result in a value of type T or an error.
// It is commonly used to handle operations that can fail, encapsulating both success and failure cases in a single type.
type Result[T any] = Either[error, T]

// NewResult creates a new Result containing the provided value as the right (success) value.
// The value should be passed as a pointer to allow for nil (empty) results.
// T represents the type of the value contained in the Result.
func NewResult[T any](value *T) Result[T] {
	return Result[T]{
		right: value,
	}
}

// NewResultError creates a new Result of type T containing an error.
// The error is stored in the left field of the Result.
// This function is typically used to represent a failed computation or operation.
func NewResultError[T any](err *error) Result[T] {
	return Result[T]{
		left: err,
	}
}
