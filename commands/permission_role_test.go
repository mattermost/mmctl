// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestAssignUsersCmd() {
	s.Run("Assigning a user to a role", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit"},
		}

		mockUser := &model.User{
			Id:       model.NewId(),
			Username: "user1",
			Roles:    "system_user",
		}

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Username, "").
			Return(mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(mockUser.Id, fmt.Sprintf("%s %s", mockUser.Roles, mockRole.Name)).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, mockUser.Username}
		err := assignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Assigning multiple users to a role", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit"},
		}

		mockUser1 := &model.User{
			Id:       model.NewId(),
			Username: "user1",
			Roles:    "system_user",
		}

		mockUser2 := &model.User{
			Id:       model.NewId(),
			Username: "user2",
			Roles:    "system_user system_admin",
		}

		notFoundUser := &model.User{
			Id:       model.NewId(),
			Username: "notfound",
			Roles:    "system_user system_admin",
		}

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		for _, user := range []*model.User{mockUser1, mockUser2} {
			s.client.
				EXPECT().
				GetUserByEmail(user.Username, "").
				Return(nil, &model.Response{Error: nil}).
				Times(1)

			s.client.
				EXPECT().
				GetUserByUsername(user.Username, "").
				Return(user, &model.Response{Error: nil}).
				Times(1)

			s.client.
				EXPECT().
				UpdateUserRoles(user.Id, fmt.Sprintf("%s %s", user.Roles, mockRole.Name)).
				Return(true, &model.Response{Error: nil}).
				Times(1)
		}

		s.client.
			EXPECT().
			GetUserByEmail(notFoundUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(notFoundUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(notFoundUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, mockUser1.Username, notFoundUser.Username, mockUser2.Username}
		err := assignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Assigning to a non-existent role", func() {
		expectedError := model.NewAppError("Role", "role_not_found", nil, "", http.StatusNotFound)

		s.client.
			EXPECT().
			GetRoleByName("non-existent").
			Return(nil, &model.Response{Error: expectedError}).
			Times(1)

		args := []string{"non-existent", "user1"}
		err := assignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().NotNil(err)
		s.Require().Equal(expectedError, err)
	})

	s.Run("Assigning a user to a role that is already assigned", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit"},
		}

		mockUser := &model.User{
			Id:       model.NewId(),
			Username: "user1",
			Roles:    "system_user mock-role",
		}

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Username, "").
			Return(mockUser, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, mockUser.Username}
		err := assignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Assigning a user that is not found", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit"},
		}

		requestedUser := "user99"

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(requestedUser, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(requestedUser, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(requestedUser, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, requestedUser}
		err := assignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})
}

func (s *MmctlUnitTestSuite) TestUnassignUsersCmd() {
	s.Run("Unassigning a user from a role", func() {
		roleName := "mock-role"

		mockUser := &model.User{
			Id:       model.NewId(),
			Username: "user1",
			Roles:    fmt.Sprintf("system_user %s team_admin", roleName),
		}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Username, "").
			Return(mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateUserRoles(mockUser.Id, "system_user team_admin").
			Return(true, &model.Response{Error: nil}).
			Times(1)

		args := []string{roleName, mockUser.Username}
		err := unassignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Unassign multiple users from a role", func() {
		roleName := "mock-role"

		mockUser1 := &model.User{
			Id:       model.NewId(),
			Username: "user1",
			Roles:    "system_user mock-role",
		}

		mockUser2 := &model.User{
			Id:       model.NewId(),
			Username: "user2",
			Roles:    "system_user system_admin mock-role",
		}

		notFoundUser := &model.User{
			Id:       model.NewId(),
			Username: "notfound",
			Roles:    "system_user system_admin",
		}

		for _, user := range []*model.User{mockUser1, mockUser2} {
			s.client.
				EXPECT().
				GetUserByEmail(user.Username, "").
				Return(nil, &model.Response{Error: nil}).
				Times(1)

			s.client.
				EXPECT().
				GetUserByUsername(user.Username, "").
				Return(user, &model.Response{Error: nil}).
				Times(1)

			s.client.
				EXPECT().
				UpdateUserRoles(user.Id, strings.TrimSpace(strings.ReplaceAll(user.Roles, roleName, ""))).
				Return(true, &model.Response{Error: nil}).
				Times(1)
		}

		s.client.
			EXPECT().
			GetUserByEmail(notFoundUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(notFoundUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(notFoundUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		args := []string{roleName, mockUser1.Username, notFoundUser.Username, mockUser2.Username}
		err := unassignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Unassign from a non-assigned or role", func() {
		roleName := "mock-role"

		mockUser := &model.User{
			Id:       model.NewId(),
			Username: "user1",
			Roles:    "system_user",
		}

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Username, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Username, "").
			Return(mockUser, &model.Response{Error: nil}).
			Times(1)

		args := []string{roleName, mockUser.Username}
		err := unassignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Unassigning a user that is not found", func() {
		requestedUser := "user99"

		s.client.
			EXPECT().
			GetUserByEmail(requestedUser, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(requestedUser, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser(requestedUser, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		args := []string{"mock-role-id", requestedUser}
		err := unassignUsersCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})
}

func (s *MmctlUnitTestSuite) TestShowRoleCmd() {
	s.Run("Show custom role", func() {
		printer.Clean()
		printer.SetFormat(printer.FormatPlain)
		defer printer.SetFormat(printer.FormatJSON)

		commandArg := "example-role-name"
		mockRole := &model.Role{
			Id:   "example-mock-id",
			Name: commandArg,
		}

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		err := showRoleCmdF(s.client, &cobra.Command{}, []string{commandArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Equal(`
Property      Value
--------      -----
Name          example-role-name
DisplayName   
BuiltIn       false
SchemeManaged false
`, printer.GetLines()[0])
	})

	s.Run("Show a role with a sysconsole_* permission", func() {
		printer.Clean()
		printer.SetFormat(printer.FormatPlain)
		defer printer.SetFormat(printer.FormatJSON)

		commandArg := "example-role-name"
		mockRole := &model.Role{
			Id:          "example-mock-id",
			Name:        commandArg,
			Permissions: []string{"sysconsole_write_site", "edit_brand"},
		}

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		err := showRoleCmdF(s.client, &cobra.Command{}, []string{commandArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Require().Len(printer.GetErrorLines(), 0)
		s.Equal(`
Property      Value                 Used by
--------      -----                 -------
Name          example-role-name     
DisplayName                         
BuiltIn       false                 
SchemeManaged false                 
Permissions   edit_brand            
              sysconsole_write_site 
`, printer.GetLines()[0])
	})

	s.Run("Show custom role with invalid name", func() {
		printer.Clean()

		expectedError := model.NewAppError("Role", "role_not_found", nil, "", http.StatusNotFound)

		commandArgBogus := "bogus-role-name"

		// showRoleCmdF will look up role by name
		s.client.
			EXPECT().
			GetRoleByName(commandArgBogus).
			Return(nil, &model.Response{Error: expectedError}).
			Times(1)

		err := showRoleCmdF(s.client, &cobra.Command{}, []string{commandArgBogus})
		s.Require().NotNil(err)
		s.Require().Equal(expectedError, err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}
