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

type NotFoundError struct {
	Msg string
}

func (e *NotFoundError) Error() string {
	return e.Msg
}

type BadRequestError struct {
	Msg string
}

func (e *BadRequestError) Error() string {
	return e.Msg
}

// ExtractErrorFromResponse extracts the error from the response,
// encapsulating it if matches the common cases, such as when it's
// not found, and when we've made a bad request
func ExtractErrorFromResponse(r *model.Response) error {
	switch r.Error.StatusCode {
	case http.StatusNotFound:
		return &NotFoundError{Msg: r.Error.Error()}
	case http.StatusBadRequest:
		return &BadRequestError{Msg: r.Error.Error()}
	default:
		return r.Error
	}
}
