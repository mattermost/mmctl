package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestConfigShowCmd() {
	s.Run("Should show config", func() {
		printer.Clean()
		configShowArg := "config-show"
		mockConfig := &model.Config{}

		s.client.
			EXPECT().
			GetConfig().
			Return(mockConfig, &model.Response{Error: nil}).
			Times(1)

		err := configShowCmdF(s.client, &cobra.Command{}, []string{configShowArg})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Equal(mockConfig, printer.GetLines()[0])
		s.Len(printer.GetErrorLines(), 0)
	})

}
