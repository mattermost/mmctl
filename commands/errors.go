package commands

import (
	"fmt"
)

// FindEntitySummary is thrown when at least, one error is found
// when finding entities.
type FindEntitySummary struct {
	Errors []error
}

// ErrEntityNotFound is thrown when an entity (user, team, etc.)
// is not found, returning the id sent by arguments
type ErrEntityNotFound struct {
	Type string
	ID   string
}

func (e ErrEntityNotFound) Error() string {
	return fmt.Sprintf("%s %s not found", e.Type, e.ID)
}
