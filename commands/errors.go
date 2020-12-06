package commands

import (
	"fmt"
	"strings"
)

// ErrEntitySummary is thrown when an error is found
// when finding entities.
type ErrEntitySummary struct {
	Errors []string
}

func (e ErrEntitySummary) Error() string {
	var str strings.Builder
	str.WriteString("summary when finding entities:\n")
	for _, e := range e.Errors {
		str.WriteString("- " + e)
	}
	return str.String()
}

// ErrEntityNotFound is thrown when an entity (user, team, etc.)
// is not found, returning the id sent by arguments
type ErrEntityNotFound struct {
	ID string
}

func (e ErrEntityNotFound) Error() string {
	return fmt.Sprintf("entity %s not found", e.ID)
}
