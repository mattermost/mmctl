package commands

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
)

func (s *MmctlUnitTestSuite) TestGetUserFromArgs() {

	s.Run("user not found", func() {
		notFoundEmail := "emailNotfound@notfound.com"
		notFoundErr := &model.AppError{Message: "user not found", StatusCode: http.StatusNotFound}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(notFoundEmail, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(notFoundEmail, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUser(notFoundEmail, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)

		users, summary := getUsersFromArgs(s.client, []string{notFoundEmail})
		s.Require().Len(users, 0)
		s.Require().NotNil(summary)
		s.Require().Len(summary.Errors, 1)
		s.Require().Equal(fmt.Sprintf("user %v not found", notFoundEmail), summary.Errors[0].Error())
	})

	s.Run("bad request don't throw unexpected error", func() {
		badRequestEmail := "emailbadrequest@badrequest.com"
		badRequestErr := &model.AppError{Message: "bad request", StatusCode: http.StatusBadRequest}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(badRequestEmail, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUserByUsername(badRequestEmail, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)
		s.client.
			EXPECT().
			GetUser(badRequestEmail, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)

		users, summary := getUsersFromArgs(s.client, []string{badRequestEmail})
		s.Require().Len(users, 0)
		s.Require().NotNil(summary)
		s.Require().Len(summary.Errors, 1)
		s.Require().Equal(fmt.Sprintf("user %v not found", badRequestEmail), summary.Errors[0].Error())
	})

	s.Run("unexpected error throws according error", func() {
		unexpectedErrEmail := "emailunexpected@unexpected.com"
		unexpectedErr := &model.AppError{Message: "internal server error", StatusCode: http.StatusInternalServerError}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(unexpectedErrEmail, "").
			Return(nil, &model.Response{Error: unexpectedErr}).
			Times(1)
		users, summary := getUsersFromArgs(s.client, []string{unexpectedErrEmail})
		s.Require().Len(users, 0)
		s.Require().NotNil(summary)
		s.Require().Len(summary.Errors, 1)
		s.Require().Equal(": internal server error, ", summary.Errors[0].Error())
	})
	s.Run("forbidden error stops searching", func() {
		forbiddenErrEmail := "forbidden@forbidden.com"
		forbiddenErr := &model.AppError{Message: "forbidden", StatusCode: http.StatusForbidden}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(forbiddenErrEmail, "").
			Return(nil, &model.Response{Error: forbiddenErr}).
			Times(1)
		users, summary := getUsersFromArgs(s.client, []string{forbiddenErrEmail})
		s.Require().Len(users, 0)
		s.Require().NotNil(summary)
		s.Require().Len(summary.Errors, 1)
		s.Require().Equal(": forbidden, ", summary.Errors[0].Error())
	})
	s.Run("success", func() {
		successEmail := "success@success.com"
		successUser := &model.User{Email: successEmail}
		printer.Clean()
		s.client.
			EXPECT().
			GetUserByEmail(successEmail, "").
			Return(successUser, nil).
			Times(1)
		users, summary := getUsersFromArgs(s.client, []string{successEmail})
		s.Require().Nil(summary)
		s.Require().Len(users, 1)
		s.Require().Equal(successUser, users[0])
	})
}
