package rule

import "strings"

// IdentifiedError is a map of errors associated with a unique identifier.
type IdentifiedError struct {
	errors map[string]error
}

// Error implements the error interface for IdentifiedError by combining all errors into a string.
func (ie *IdentifiedError) Error() string {
	if ie == nil || ie.errors == nil {
		return ""
	}

	var sb strings.Builder

	// Combined each error with its identifier.
	for id, err := range ie.errors {
		sb.WriteString(id)
		sb.WriteString(": ")
		sb.WriteString(err.Error())
		sb.WriteString("; ")
	}

	// Remove the last semicolon and space.
	errStr := sb.String()
	errStr = errStr[:len(errStr)-2]

	return errStr
}

// Add adds a new error to the identified error map.
func (ie *IdentifiedError) Add(id string, err error) {
	if err == nil {
		return
	}

	// Check if the error map is nil and initialize it if necessary.
	if ie.errors == nil {
		ie.errors = make(map[string]error)
	}

	ie.errors[id] = err
}

// Get retrieves an error by its identifier.
func (ie *IdentifiedError) Get(id string) error {
	if ie == nil || ie.errors == nil {
		return nil
	}

	if err, found := ie.errors[id]; found {
		return err
	}

	return nil
}

// Size returns the number of errors in the identified error map.
func (ie *IdentifiedError) Size() int {
	if ie == nil {
		return 0
	}

	return len(ie.errors)
}
