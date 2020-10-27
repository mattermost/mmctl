package commands

import (
	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

const (
	userID   = "userID"
	email    = "example@example.org"
	userName = "ExampleUser"
	roles    = ""
)

func (s *MmctlUnitTestSuite) TestMakeAdminCmd() {

	s.Run("Add admin priveleges to user", func() {
		printer.Clean()
		mockUser := model.User{Id: userID, Username: userName, Email: email, Roles: roles}
		newRoles := model.SYSTEM_ADMIN_ROLE_ID
		updatedUser := model.User{Id: userID, Username: userName, Email: email, Roles: newRoles}

		s.client.
			EXPECT().
			GetUserByEmail(email, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(gomock.Eq(userID), gomock.Eq(model.SYSTEM_ADMIN_ROLE_ID)).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(userID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userID, "").
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		err := makeSystemAdminCmdF(s.client, &cobra.Command{}, []string{email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])

	})

	s.Run("Adding admin privileges to existing admin", func() {
		printer.Clean()
		rolesArg := model.SYSTEM_ADMIN_ROLE_ID
		mockUser := model.User{Id: userID, Username: userName, Email: email, Roles: rolesArg}

		s.client.
			EXPECT().
			GetUserByEmail(email, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		err := makeSystemAdminCmdF(s.client, &cobra.Command{}, []string{email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)

	})

	s.Run("Add admin to non existing user", func() {
		printer.Clean()
		emailArg := "doesnotexist@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := makeSystemAdminCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().Error(err)
		s.Require().EqualErrorf(err, err.Error(), "Unable to find user '%s'", emailArg)
	})

	s.Run("Error while updating admin role", func() {
		printer.Clean()
		mockUser := model.User{Id: userID, Username: userName, Email: email, Roles: roles}
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetUserByEmail(email, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(gomock.Eq(userID), gomock.Eq(model.SYSTEM_ADMIN_ROLE_ID)).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := makeSystemAdminCmdF(s.client, &cobra.Command{}, []string{email})
		s.Require().Error(err)
		s.Require().EqualError(err, "Unable to update user roles. Error: : , ")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})
}

func (s *MmctlUnitTestSuite) TestMakeMemberCmd() {
	s.Run("Remove admin privileges for admin", func() {

		printer.Clean()
		rolesArg := model.SYSTEM_ADMIN_ROLE_ID
		mockUser := model.User{Id: userID, Username: userName, Email: email, Roles: rolesArg}
		newRoles := model.SYSTEM_USER_ROLE_ID
		updatedUser := model.User{Id: userID, Username: userName, Email: email, Roles: newRoles}

		s.client.
			EXPECT().
			GetUserByEmail(email, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(gomock.Eq(userID), gomock.Eq(model.SYSTEM_USER_ROLE_ID)).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(userID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userID, "").
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		err := makeMemberCmdF(s.client, &cobra.Command{}, []string{email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])
	})

	s.Run("Remove admin privileges from non admin user", func() {
		printer.Clean()
		mockUser := model.User{Id: userID, Username: userName, Email: email, Roles: roles}
		newRoles := model.SYSTEM_USER_ROLE_ID
		updatedUser := model.User{Id: userID, Username: userName, Email: email, Roles: newRoles}

		s.client.
			EXPECT().
			GetUserByEmail(email, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(gomock.Eq(userID), gomock.Eq(model.SYSTEM_USER_ROLE_ID)).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(userID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(userID, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(userID, "").
			Return(&updatedUser, &model.Response{Error: nil}).
			Times(1)

		err := makeMemberCmdF(s.client, &cobra.Command{}, []string{email})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&updatedUser, printer.GetLines()[0])

	})

	s.Run("Error while updating non admin role", func() {
		printer.Clean()
		rolesArg := model.SYSTEM_ADMIN_ROLE_ID
		mockUser := model.User{Id: userID, Username: userName, Email: email, Roles: rolesArg}
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetUserByEmail(email, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(gomock.Eq(userID), gomock.Eq(model.SYSTEM_USER_ROLE_ID)).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := makeMemberCmdF(s.client, &cobra.Command{}, []string{email})
		s.Require().Error(err)
		s.Require().EqualError(err, "Unable to update user roles. Error: : , ")
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove admin from non existing user", func() {
		printer.Clean()
		emailArg := "doesnotexist@example.com"

		s.client.
			EXPECT().
			GetUserByEmail(emailArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(emailArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(emailArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := makeMemberCmdF(s.client, &cobra.Command{}, []string{emailArg})
		s.Require().Error(err)
		s.Require().EqualError(err, "Unable to find user 'doesnotexist@example.com'")
	})
}
