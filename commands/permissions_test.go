package commands

import (
	"net/http"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/mocks"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestAddPermissionsCmd(t *testing.T) {
	t.Run("Adding a new permission to an existing role", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

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

		c := mocks.NewMockClient(mockCtrl)
		c.
			EXPECT().
			GetRoleByName(gomock.Eq(mockRole.Name)).
			Return(mockRole, &model.Response{Error: nil}).
			Times(1)

		c.
			EXPECT().
			PatchRole(gomock.Eq(mockRole.Id), gomock.Eq(expectedPatch)).
			Return(&model.Role{}, &model.Response{Error: nil}).
			Times(1)

		args := []string{mockRole.Name, newPermission}
		err := addPermissionsCmdF(c, &cobra.Command{}, args)
		require.Nil(t, err)
	})

	t.Run("Trying to add a new permission to a non existing role", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		expectedError := model.NewAppError("Role", "role_not_found", nil, "", http.StatusNotFound)

		c := mocks.NewMockClient(mockCtrl)
		c.
			EXPECT().
			GetRoleByName(gomock.Any()).
			Return(nil, &model.Response{Error: expectedError}).
			Times(1)

		args := []string{"mockRole", "newPermission"}
		err := addPermissionsCmdF(c, &cobra.Command{}, args)
		require.Equal(t, expectedError, err)
	})
}
