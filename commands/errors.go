package commands

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
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

func isErrorSevere(r *model.Response) bool {
	return r != nil && r.Error != nil && (r.Error.StatusCode != http.StatusNotFound && r.Error.StatusCode != http.StatusBadRequest)
}
