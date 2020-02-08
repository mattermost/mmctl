// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"net/http"

	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestAddPermissionsCmd() {
	s.Run("Adding a new permission to an existing role", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit"},
		}
		newPermission := "delete"

		expectedPermissions := append(mockRole.Permissions, newPermission)
		expectedPatch := &model.RolePatch{
			Permissions: &expectedPermissions,
		}

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchRole(mockRole.Id, expectedPatch).
			Return(&model.Role{}, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, newPermission}
		err := addPermissionsCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Trying to add a new permission to a non existing role", func() {
		expectedError := model.NewAppError("Role", "role_not_found", nil, "", http.StatusNotFound)

		s.client.
			EXPECT().
			GetRoleByName(gomock.Any()).
			Return(nil, &model.Response{Error: expectedError}).
			Times(1)

		args := []string{"mockRole", "newPermission"}
		err := addPermissionsCmdF(s.client, &cobra.Command{}, args)
		s.Require().Equal(expectedError, err)
	})
}

func (s *MmctlUnitTestSuite) TestRemovePermissionsCmd() {
	s.Run("Removing a permission from an existing role", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit", "delete"},
		}

		expectedPatch := &model.RolePatch{
			Permissions: &[]string{"view", "edit"},
		}
		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchRole(mockRole.Id, expectedPatch).
			Return(&model.Role{}, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, "delete"}
		err := removePermissionsCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Removing multiple permissions from an existing role", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit", "delete"},
		}

		expectedPatch := &model.RolePatch{
			Permissions: &[]string{"edit"},
		}
		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchRole(mockRole.Id, expectedPatch).
			Return(&model.Role{}, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, "view", "delete"}
		err := removePermissionsCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Removing a non-existing permission from an existing role", func() {
		mockRole := &model.Role{
			Id:          "mock-id",
			Name:        "mock-role",
			Permissions: []string{"view", "edit"},
		}

		expectedPatch := &model.RolePatch{
			Permissions: &[]string{"view", "edit"},
		}
		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)
		s.client.
			EXPECT().
			PatchRole(mockRole.Id, expectedPatch).
			Return(&model.Role{}, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, "delete"}
		err := removePermissionsCmdF(s.client, &cobra.Command{}, args)
		s.Require().Nil(err)
	})

	s.Run("Removing a permission from a non-existing role", func() {
		mockRole := model.Role{
			Name:        "exampleName",
			Permissions: []string{"view", "edit", "delete"},
		}

		mockError := model.NewAppError("Role", "role_not_found", nil, "", http.StatusNotFound)
		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		args := []string{mockRole.Name, "delete"}
		err := removePermissionsCmdF(s.client, &cobra.Command{}, args)
		s.Require().EqualError(err, "Role: role_not_found, ")
	})
}

func (s *MmctlUnitTestSuite) TestShowRoleCmd() {
	s.Run("Show custom role", func() {
		commandArg := "example-role-name"
		mockRole := &model.Role{
			Id:   "example-mock-id",
			Name: commandArg,
		}

		printer.Clean()

		s.client.
			EXPECT().
			GetRoleByName(mockRole.Name).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		err := showRoleCmdF(s.client, &cobra.Command{}, []string{commandArg})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 1)
		s.Equal(mockRole, printer.GetLines()[0])
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Show custom role with invalid name", func() {
		expectedError := model.NewAppError("Role", "role_not_found", nil, "", http.StatusNotFound)

		commandArgBogus := "bogus-role-name"

		printer.Clean()

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
