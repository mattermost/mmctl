package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestListLdapGroupsCmd() {
	s.Run("Failure getting Ldap Groups", func() {
		printer.Clean()
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetLdapGroups().
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := listLdapGroupsCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Equal(&mockError, err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("List several groups", func() {
		printer.Clean()
		mockList := []*model.Group{
			&model.Group{DisplayName: "Group1"},
			&model.Group{DisplayName: "Group2"},
			&model.Group{DisplayName: "Group3"},
		}

		s.client.
			EXPECT().
			GetLdapGroups().
			Return(mockList, &model.Response{Error: nil}).
			Times(1)

		err := listLdapGroupsCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().NoError(err)
		s.Require().Len(printer.GetLines(), 3)
		for i, v := range mockList {
			s.Require().Equal(v, printer.GetLines()[i])
		}
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}
