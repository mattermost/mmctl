// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlE2ETestSuite) TestPostListCmd() {
	s.SetupTestHelper().InitBasic()

	s.RunForAllClients("List all posts for a channel", func(c client.Client) {
		printer.Clean()

		cmd := &cobra.Command{}
		cmd.Flags().Int("number", 1, "")
		err := postListCmdF(c, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().Nil(err)
		s.Equal(1, len(printer.GetLines()))
		s.Len(printer.GetErrorLines(), 0)
	})

	s.RunForAllClients("List all posts for a channel with since flag", func(c client.Client) {
		printer.Clean()

		ISO8601ValidString := "2006-01-02T15:04:05-07:00"

		cmd := &cobra.Command{}
		cmd.Flags().Int("number", 1, "")
		cmd.Flags().String("since", ISO8601ValidString, "")
		err := postListCmdF(c, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().Nil(err)
		s.Equal(2, len(printer.GetLines()))
		s.Len(printer.GetErrorLines(), 0)
	})

}

func (s *MmctlE2ETestSuite) TestPostCreateCmd() {
	s.SetupTestHelper().InitBasic()

	s.Run("Create a post for System Admin Client", func() {
		printer.Clean()

		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		err := postCreateCmdF(s.th.SystemAdminClient, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a post for Client", func() {
		printer.Clean()

		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		err := postCreateCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Create a post for Local Client should fail", func() {
		printer.Clean()

		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")

		err := postCreateCmdF(s.th.LocalClient, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().NotNil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Reply to a an existing post for System Admin Client", func() {
		printer.Clean()

		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")
		cmd.Flags().String("reply-to", s.th.BasicPost.Id, "")

		err := postCreateCmdF(s.th.SystemAdminClient, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Reply to a an existing post for Client", func() {
		printer.Clean()

		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")
		cmd.Flags().String("reply-to", s.th.BasicPost.Id, "")

		err := postCreateCmdF(s.th.Client, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().Nil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

	s.Run("Reply to a an existing post for Local Client should fail", func() {
		printer.Clean()

		msgArg := "some text"

		cmd := &cobra.Command{}
		cmd.Flags().String("message", msgArg, "")
		cmd.Flags().String("reply-to", s.th.BasicPost.Id, "")

		err := postCreateCmdF(s.th.LocalClient, cmd, []string{s.th.BasicTeam.Name + ":" + s.th.BasicChannel.Name})
		s.Require().NotNil(err)
		s.Len(printer.GetErrorLines(), 0)
	})

}
