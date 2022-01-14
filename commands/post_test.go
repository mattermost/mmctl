// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/mmctl/v6/printer"
)

func (s *MmctlUnitTestSuite) TestPostCreateCmdF() {
	s.Run("create a post with empty text", func() {
		cmd := &cobra.Command{}

		err := postCreateCmdF(s.client, cmd, []string{"some-channel", ""})
		s.Require().EqualError(err, "message cannot be empty")
	})

	s.Run("no channel specified", func() {
		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		err := postCreateCmdF(s.client, cmd, []string{"", msgArg})
		s.Require().EqualError(err, "Unable to find channel ''")
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
			Return(nil, &model.Response{}, errors.New("some-error")).
			Times(1)

		err := postCreateCmdF(s.client, cmd, []string{msgArg})
		s.Require().Contains(err.Error(), "some-error")
	})

	s.Run("error when creating a post", func() {
		msgArg := "some text"
		channelArg := "example-channel"
		mockChannel := model.Channel{Name: channelArg}
		mockPost := &model.Post{Message: msgArg}
		data, err := mockPost.ToJSON()
		s.Require().NoError(err)

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{}, nil).
			Times(1)

		s.client.
			EXPECT().
			DoAPIPost("/posts?set_online=false", data).
			Return(nil, errors.New("some-error")).
			Times(1)

		err = postCreateCmdF(s.client, cmd, []string{channelArg, msgArg})
		s.Require().Contains(err.Error(), "could not create post")
	})

	s.Run("create a post", func() {
		msgArg := "some text"
		channelArg := "example-channel"
		mockChannel := model.Channel{Name: channelArg}
		mockPost := model.Post{Message: msgArg}
		data, err := mockPost.ToJSON()
		s.Require().NoError(err)

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{}, nil).
			Times(1)

		s.client.
			EXPECT().
			DoAPIPost("/posts?set_online=false", data).
			Return(nil, nil).
			Times(1)

		err = postCreateCmdF(s.client, cmd, []string{channelArg, msgArg})
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
		data, err := mockPost.ToJSON()
		s.Require().NoError(err)

		cmd := &cobra.Command{}
		cmd.Flags().String("reply-to", replyToArg, "")
		cmd.Flags().String("message", msgArg, "")

		s.client.
			EXPECT().
			GetChannel(channelArg, "").
			Return(&mockChannel, &model.Response{}, nil).
			Times(1)

		s.client.
			EXPECT().
			GetPost(replyToArg, "").
			Return(&mockReplyTo, &model.Response{}, nil).
			Times(1)

		s.client.
			EXPECT().
			DoAPIPost("/posts?set_online=false", data).
			Return(nil, nil).
			Times(1)

		err = postCreateCmdF(s.client, cmd, []string{channelArg, msgArg})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})
}
