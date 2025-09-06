package common

// ErrorMsg represents a structured error message containing a main message and a list of detailed errors.
// It is typically used for API responses to provide both a summary and specific error details.
type ErrorMsg struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// NewErrorMsg creates a new ErrorMsg instance with the provided message and a slice of error strings.
// It returns an ErrorMsg containing the specified message and errors.
func NewErrorMsg(message string, errs []string) ErrorMsg {
	return ErrorMsg{
		Message: message,
		Errors:  errs,
	}
}
