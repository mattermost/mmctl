package commands

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestPostCreateCmdF() {
	s.Run("create a post with empty text", func() {
		cmd := &cobra.Command{}
		cmd.Flags().String("message", "", "")

		err := postCreateCmdF(s.client, cmd, nil)
		s.Require().NotNil(err)
	})

	s.Run("no channel specified", func() {
		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		err := postCreateCmdF(s.client, cmd, []string{msgArg})
		s.Require().NotNil(err)
	})

	s.Run("wrong reply msg", func() {
		msgArg := "some text"
		replyToArg := "a-non-existing-post"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")
		cmd.Flags().String("reply-to", replyToArg, "")

		s.client.
			EXPECT().
			GetPost(replyToArg, "").
			Return(nil, &model.Response{Error: &model.AppError{Message: "some-error"}}).
			Times(1)

		err := postCreateCmdF(s.client, cmd, []string{msgArg})
		s.Require().NotNil(err)
	})

	s.Run("create a post", func() {
		msgArg := "some text"
		channelArg := "example-channel"
		mockChannel := model.Channel{Name: channelArg}
		mockPost := model.Post{Message: msgArg}

		cmd := &cobra.Command{}
		cmd.Flags().String("channel", channelArg, "")
		cmd.Flags().String("message", msgArg, "")

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreatePost(&mockPost).
			Return(&mockPost, &model.Response{Error: nil}).
			Times(1)

		err := postCreateCmdF(s.client, cmd, []string{channelArg, msgArg})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("reply to an existing post", func() {
		msgArg := "some text"
		replyToArg := "an-existing-post"
		rootID := "some-root-id"
		channelArg := "example-channel"
		mockChannel := model.Channel{Name: channelArg}
		mockReplyTo := model.Post{RootId: rootID}
		mockPost := model.Post{Message: msgArg, RootId: rootID}

		cmd := &cobra.Command{}
		cmd.Flags().String("channel", channelArg, "")
		cmd.Flags().String("reply-to", replyToArg, "")
		cmd.Flags().String("message", msgArg, "")

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetPost(replyToArg, "").
			Return(&mockReplyTo, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreatePost(&mockPost).
			Return(&mockPost, &model.Response{Error: nil}).
			Times(1)

		err := postCreateCmdF(s.client, cmd, []string{channelArg, msgArg})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})
}
