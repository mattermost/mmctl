package commands

import (
	"net/http"

	"github.com/mattermost/mattermost-server/model"

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
			GetRoleByName(gomock.Eq(mockRole.Name)).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			PatchRole(gomock.Eq(mockRole.Id), gomock.Eq(expectedPatch)).
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
