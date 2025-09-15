package monad

// Either represents a value of one of two possible types (a disjoint union).
// An Either is either Left containing a value of type LEFT, or Right containing a value of type RIGHT.
// This can be used to model computations that may result in two different types of outcomes,
// such as success (Right) or failure (Left).
type Either[LEFT any, RIGHT any] struct {
	left  *LEFT
	right *RIGHT
}

// Left creates an Either instance representing a left value.
// The left value is typically used to indicate an error or alternative computation path.
// The generic parameters LEFT and RIGHT specify the types of the left and right values, respectively.
func Left[LEFT any, RIGHT any](left LEFT) Either[LEFT, RIGHT] {
	return Either[LEFT, RIGHT]{
		left: &left,
	}
}

// Right creates an Either instance representing a successful value of type RIGHT.
// The LEFT type parameter represents the error or alternative type, while RIGHT is the success type.
// The returned Either will have its right value set, indicating a successful computation.
func Right[LEFT any, RIGHT any](right RIGHT) Either[LEFT, RIGHT] {
	return Either[LEFT, RIGHT]{
		right: &right,
	}
}

// IsLeft returns true if the Either instance contains a value of type LEFT.
// It indicates that the Either is in the "Left" state.
func (e Either[LEFT, RIGHT]) IsLeft() bool {
	return e.left != nil
}

// IsRight returns true if the Either contains a value of type RIGHT, indicating a successful result.
// It returns false if the Either contains a value of type LEFT, indicating an error or alternative outcome.
func (e Either[LEFT, RIGHT]) IsRight() bool {
	return e.right != nil
}

// Bind applies the provided function 'f' to the value contained in the Either if it is a Right,
// returning the result of 'f'. If the Either is a Left, it returns itself unchanged.
// This enables chaining operations that may each return an Either, following the monadic pattern.
//
// Example usage:
//
//	result := either.Bind(func(val RIGHT) Either[LEFT, RIGHT] { ... })
//
// Parameters:
//
//	f - a function that takes a RIGHT value and returns an Either[LEFT, RIGHT].
//
// Returns:
//
//	Either[LEFT, RIGHT] - the result of applying 'f' if the Either is a Right, or the original Either if it is a Left.
func (e Either[LEFT, RIGHT]) Bind(f func(RIGHT) Either[LEFT, RIGHT]) Either[LEFT, RIGHT] {
	if e.IsRight() {
		return f(*e.right)
	}
	return e
}

// Right returns the value contained in the Either as a RIGHT type.
// It assumes that the Either instance holds a RIGHT value.
// If the Either contains a LEFT value, dereferencing e.right may cause a runtime panic.
func (e Either[LEFT, RIGHT]) Right() RIGHT {
	return *e.right
}

// Left returns the value contained in the Left side of the Either.
// It panics if the Either does not contain a Left value.
func (e Either[LEFT, RIGHT]) Left() LEFT {
	return *e.left
}
