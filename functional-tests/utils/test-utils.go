package utils

import (
	"os"
)

// AroundWrapperFunction is a generic utility function that wraps a given function
// with "before" and "after" hooks. It allows you to execute setup and teardown
// logic around the execution of a core function.
//
// Type Parameters:
//   - T: The type of the resource returned by the "before" function and consumed
//     by the "after" function.
//
// Parameters:
//   - before: A function that performs setup logic and returns a resource of type T.
//   - after: A function that performs teardown logic, accepting the resource of type T
//     returned by the "before" function.
//   - wrappedFunction: The core function to be executed, which returns an integer exit code.
//
// Returns:
//   - A function that, when called, executes the "before" function, the core function,
//     and the "after" function in sequence. The program exits with the exit code
//     returned by the core function.
func AroundWrapperFunction[T any](before func() T, after func(T), wrappedFunction func() int) func() {
	return func() {
		resource := before()
		code := wrappedFunction()
		after(resource)
		os.Exit(code)
	}
}

// PanicOnError panics if the provided error is not nil.
// This function is useful for simplifying error handling in tests or initialization code
// where errors should immediately halt execution.
func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// GetFirstNonEmpty returns the first non-empty string from the provided arguments.
// If all arguments are empty, it returns an empty string.
func GetFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
