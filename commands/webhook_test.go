// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestListWebhookCmd() {
	teamID := "teamID"
	incomingWebhookID := "incomingWebhookID"
	incomingWebhookDisplayName := "incomingWebhookDisplayName"
	outgoingWebhookID := "outgoingWebhookID"
	outgoingWebhookDisplayName := "outgoingWebhookDisplayName"

	s.Run("Listing all webhooks", func() {
		printer.Clean()

		mockTeam := model.Team{
			Id: teamID,
		}
		mockIncomingWebhook := model.IncomingWebhook{
			Id:          incomingWebhookID,
			DisplayName: incomingWebhookDisplayName,
		}
		mockOutgoingWebhook := model.OutgoingWebhook{
			Id:          outgoingWebhookID,
			DisplayName: outgoingWebhookDisplayName,
		}

		s.client.
			EXPECT().
			GetAllTeams("", 0, 100000000).
			Return([]*model.Team{&mockTeam}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetIncomingWebhooksForTeam(teamID, 0, 100000000, "").
			Return([]*model.IncomingWebhook{&mockIncomingWebhook}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhooksForTeam(teamID, 0, 100000000, "").
			Return([]*model.OutgoingWebhook{&mockOutgoingWebhook}, &model.Response{Error: nil}).
			Times(1)

		err := listWebhookCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 2)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&mockIncomingWebhook, printer.GetLines()[0])
		s.Require().Equal(&mockOutgoingWebhook, printer.GetLines()[1])
	})

	s.Run("List webhooks by team", func() {
		printer.Clean()

		mockTeam := model.Team{
			Id: teamID,
		}
		mockIncomingWebhook := model.IncomingWebhook{
			Id:          incomingWebhookID,
			DisplayName: incomingWebhookDisplayName,
		}
		mockOutgoingWebhook := model.OutgoingWebhook{
			Id:          outgoingWebhookID,
			DisplayName: outgoingWebhookDisplayName,
		}
		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetIncomingWebhooksForTeam(teamID, 0, 100000000, "").
			Return([]*model.IncomingWebhook{&mockIncomingWebhook}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhooksForTeam(teamID, 0, 100000000, "").
			Return([]*model.OutgoingWebhook{&mockOutgoingWebhook}, &model.Response{Error: nil}).
			Times(1)

		err := listWebhookCmdF(s.client, &cobra.Command{}, []string{teamID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 2)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&mockIncomingWebhook, printer.GetLines()[0])
		s.Require().Equal(&mockOutgoingWebhook, printer.GetLines()[1])
	})

	s.Run("Unable to list webhooks", func() {
		printer.Clean()

		mockTeam := model.Team{
			Id: teamID,
		}
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetAllTeams("", 0, 100000000).
			Return([]*model.Team{&mockTeam}, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetIncomingWebhooksForTeam(teamID, 0, 100000000, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhooksForTeam(teamID, 0, 100000000, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := listWebhookCmdF(s.client, &cobra.Command{}, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 2)
		s.Require().Equal("Unable to list incoming webhooks for '"+teamID+"'", printer.GetErrorLines()[0])
		s.Require().Equal("Unable to list outgoing webhooks for '"+teamID+"'", printer.GetErrorLines()[1])
	})
}

func (s *MmctlUnitTestSuite) TestCreateIncomingWebhookCmd() {
	incomingWebhookID := "incomingWebhookID"
	channelID := "channelID"
	userID := "userID"
	emailID := "emailID"
	userName := "userName"
	displayName := "displayName"

	cmd := &cobra.Command{}
	cmd.Flags().String("channel", channelID, "")
	cmd.Flags().String("user", emailID, "")
	cmd.Flags().String("display-name", displayName, "")

	s.Run("Successfully create new incoming webhook", func() {
		printer.Clean()

		mockChannel := model.Channel{
			Id: channelID,
		}
		mockUser := model.User{
			Id:       userID,
			Email:    emailID,
			Username: userName,
		}
		mockIncomingWebhook := model.IncomingWebhook{
			ChannelId:   channelID,
			Username:    userName,
			DisplayName: displayName,
		}
		returnedIncomingWebhook := mockIncomingWebhook
		returnedIncomingWebhook.Id = incomingWebhookID

		s.client.
			EXPECT().
			GetChannel(channelID, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailID, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateIncomingWebhook(&mockIncomingWebhook).
			Return(&returnedIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		err := createIncomingWebhookCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&returnedIncomingWebhook, printer.GetLines()[0])
	})

	s.Run("Incoming webhook creation error", func() {
		printer.Clean()

		mockChannel := model.Channel{
			Id: channelID,
		}
		mockUser := model.User{
			Id:       userID,
			Email:    emailID,
			Username: userName,
		}
		mockIncomingWebhook := model.IncomingWebhook{
			ChannelId:   channelID,
			Username:    userName,
			DisplayName: displayName,
		}
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetChannel(channelID, "").
			Return(&mockChannel, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailID, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateIncomingWebhook(&mockIncomingWebhook).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := createIncomingWebhookCmdF(s.client, cmd, []string{})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to create webhook", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestModifyIncomingWebhookCmd() {
	incomingWebhookID := "incomingWebhookID"
	channelID := "channelID"
	userName := "userName"
	displayName := "displayName"

	s.Run("Successfully modify incoming webhook", func() {
		printer.Clean()

		mockIncomingWebhook := model.IncomingWebhook{
			Id:            incomingWebhookID,
			ChannelId:     channelID,
			Username:      userName,
			DisplayName:   displayName,
			ChannelLocked: false,
		}

		lockToChannel := true
		updatedIncomingWebhook := mockIncomingWebhook
		updatedIncomingWebhook.ChannelLocked = lockToChannel

		cmd := &cobra.Command{}

		_ = cmd.Flags().Set("lock-to-channel", strconv.FormatBool(lockToChannel))

		s.client.
			EXPECT().
			GetIncomingWebhook(incomingWebhookID, "").
			Return(&mockIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateIncomingWebhook(&mockIncomingWebhook).
			Return(&updatedIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		err := modifyIncomingWebhookCmdF(s.client, cmd, []string{incomingWebhookID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&updatedIncomingWebhook, printer.GetLines()[0])
	})

	s.Run("modify incoming webhook errored", func() {
		printer.Clean()

		mockIncomingWebhook := model.IncomingWebhook{
			Id:            incomingWebhookID,
			ChannelId:     channelID,
			Username:      userName,
			DisplayName:   displayName,
			ChannelLocked: false,
		}

		lockToChannel := true

		mockError := model.AppError{Id: "Mock Error"}

		cmd := &cobra.Command{}

		_ = cmd.Flags().Set("lock-to-channel", strconv.FormatBool(lockToChannel))

		s.client.
			EXPECT().
			GetIncomingWebhook(incomingWebhookID, "").
			Return(&mockIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateIncomingWebhook(&mockIncomingWebhook).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := modifyIncomingWebhookCmdF(s.client, cmd, []string{incomingWebhookID})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to modify incoming webhook", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestCreateOutgoingWebhookCmd() {
	teamID := "teamID"
	outgoingWebhookID := "outgoingWebhookID"
	userID := "userID"
	emailID := "emailID"
	userName := "userName"
	triggerWhen := "exact"

	cmd := &cobra.Command{}
	cmd.Flags().String("team", teamID, "")
	cmd.Flags().String("user", emailID, "")
	cmd.Flags().String("trigger-when", triggerWhen, "")

	s.Run("Successfully create outgoing webhook", func() {
		printer.Clean()

		mockTeam := model.Team{
			Id: teamID,
		}
		mockUser := model.User{
			Id:       userID,
			Email:    emailID,
			Username: userName,
		}
		mockOutgoingWebhook := model.OutgoingWebhook{
			CreatorId:    userID,
			Username:     userName,
			TeamId:       teamID,
			TriggerWords: []string{},
			TriggerWhen:  0,
			CallbackURLs: []string{},
		}

		createdOutgoingWebhook := mockOutgoingWebhook
		createdOutgoingWebhook.Id = outgoingWebhookID

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailID, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateOutgoingWebhook(&mockOutgoingWebhook).
			Return(&createdOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		err := createOutgoingWebhookCmdF(s.client, cmd, []string{})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&createdOutgoingWebhook, printer.GetLines()[0])
	})

	s.Run("Create outgoing webhook error", func() {
		printer.Clean()

		mockTeam := model.Team{
			Id: teamID,
		}
		mockUser := model.User{
			Id:       userID,
			Email:    emailID,
			Username: userName,
		}
		mockOutgoingWebhook := model.OutgoingWebhook{
			CreatorId:    userID,
			Username:     userName,
			TeamId:       teamID,
			TriggerWords: []string{},
			TriggerWhen:  0,
			CallbackURLs: []string{},
		}
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(emailID, "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			CreateOutgoingWebhook(&mockOutgoingWebhook).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := createOutgoingWebhookCmdF(s.client, cmd, []string{})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to create outgoing webhook", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestModifyOutgoingWebhookCmd() {
	outgoingWebhookID := "outgoingWebhookID"

	s.Run("Successfully modify outgoing webhook", func() {
		printer.Clean()

		mockOutgoingWebhook := model.OutgoingWebhook{
			Id:           outgoingWebhookID,
			TriggerWords: []string{},
			CallbackURLs: []string{},
			TriggerWhen:  0,
		}

		updatedOutgoingWebhook := mockOutgoingWebhook
		updatedOutgoingWebhook.TriggerWhen = 1

		cmd := &cobra.Command{}
		cmd.Flags().StringArray("url", []string{}, "")
		cmd.Flags().StringArray("trigger-word", []string{}, "")
		cmd.Flags().String("trigger-when", "start", "")

		s.client.
			EXPECT().
			GetOutgoingWebhook(outgoingWebhookID).
			Return(&mockOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateOutgoingWebhook(&mockOutgoingWebhook).
			Return(&updatedOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		err := modifyOutgoingWebhookCmdF(s.client, cmd, []string{outgoingWebhookID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&updatedOutgoingWebhook, printer.GetLines()[0])
	})

	s.Run("Modify outgoing webhook error", func() {
		printer.Clean()

		mockOutgoingWebhook := model.OutgoingWebhook{
			Id:           outgoingWebhookID,
			TriggerWords: []string{},
			CallbackURLs: []string{},
			TriggerWhen:  0,
		}
		mockError := model.AppError{Id: "Mock Error"}

		cmd := &cobra.Command{}
		cmd.Flags().StringArray("url", []string{}, "")
		cmd.Flags().StringArray("trigger-word", []string{}, "")
		cmd.Flags().String("trigger-when", "start", "")

		s.client.
			EXPECT().
			GetOutgoingWebhook(outgoingWebhookID).
			Return(&mockOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			UpdateOutgoingWebhook(&mockOutgoingWebhook).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := modifyOutgoingWebhookCmdF(s.client, cmd, []string{outgoingWebhookID})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to modify outgoing webhook", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestDeleteWebhookCmd() {
	incomingWebhookID := "incomingWebhookID"
	outgoingWebhookID := "outgoingWebhookID"

	s.Run("Successfully delete incoming webhook", func() {
		printer.Clean()

		mockIncomingWebhook := model.IncomingWebhook{Id: incomingWebhookID}

		s.client.
			EXPECT().
			GetIncomingWebhook(incomingWebhookID, "").
			Return(&mockIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteIncomingWebhook(incomingWebhookID).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := deleteWebhookCmdF(s.client, &cobra.Command{}, []string{incomingWebhookID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&mockIncomingWebhook, printer.GetLines()[0])
	})

	s.Run("Successfully delete outgoing webhook", func() {
		printer.Clean()

		mockError := model.AppError{Id: "Mock Error"}
		mockOutgoingWebhook := model.OutgoingWebhook{Id: outgoingWebhookID}

		s.client.
			EXPECT().
			GetIncomingWebhook(outgoingWebhookID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhook(outgoingWebhookID).
			Return(&mockOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteOutgoingWebhook(outgoingWebhookID).
			Return(true, &model.Response{Error: nil}).
			Times(1)

		err := deleteWebhookCmdF(s.client, &cobra.Command{}, []string{outgoingWebhookID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(&mockOutgoingWebhook, printer.GetLines()[0])
	})

	s.Run("delete incoming webhook error", func() {
		printer.Clean()

		mockIncomingWebhook := model.IncomingWebhook{Id: incomingWebhookID}
		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetIncomingWebhook(incomingWebhookID, "").
			Return(&mockIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteIncomingWebhook(incomingWebhookID).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := deleteWebhookCmdF(s.client, &cobra.Command{}, []string{incomingWebhookID})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to delete webhook '"+incomingWebhookID+"'", printer.GetErrorLines()[0])
	})

	s.Run("delete outgoing webhook error", func() {
		printer.Clean()

		mockError := model.AppError{Id: "Mock Error"}
		mockOutgoingWebhook := model.OutgoingWebhook{Id: outgoingWebhookID}

		s.client.
			EXPECT().
			GetIncomingWebhook(outgoingWebhookID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhook(outgoingWebhookID).
			Return(&mockOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			DeleteOutgoingWebhook(outgoingWebhookID).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := deleteWebhookCmdF(s.client, &cobra.Command{}, []string{outgoingWebhookID})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 1)
		s.Require().Equal("Unable to delete webhook '"+outgoingWebhookID+"'", printer.GetErrorLines()[0])
	})
}

func (s *MmctlUnitTestSuite) TestShowWebhookCmd() {
	incomingWebhookID := "incomingWebhookID"
	outgoingWebhookID := "outgoingWebhookID"
	nonExistentID := "nonExistentID"

	s.Run("Successfully show incoming webhook", func() {
		printer.Clean()

		mockIncomingWebhook := model.IncomingWebhook{Id: incomingWebhookID}

		s.client.
			EXPECT().
			GetIncomingWebhook(incomingWebhookID, "").
			Return(&mockIncomingWebhook, &model.Response{Error: nil}).
			Times(1)

		err := showWebhookCmdF(s.client, &cobra.Command{}, []string{incomingWebhookID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(mockIncomingWebhook, printer.GetLines()[0])
	})

	s.Run("Successfully show outgoing webhook", func() {
		printer.Clean()

		mockError := model.AppError{Id: "Mock Error"}
		mockOutgoingWebhook := model.OutgoingWebhook{Id: outgoingWebhookID}

		s.client.
			EXPECT().
			GetIncomingWebhook(outgoingWebhookID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhook(outgoingWebhookID).
			Return(&mockOutgoingWebhook, &model.Response{Error: nil}).
			Times(1)

		err := showWebhookCmdF(s.client, &cobra.Command{}, []string{outgoingWebhookID})
		s.Require().Nil(err)
		s.Len(printer.GetLines(), 1)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal(mockOutgoingWebhook, printer.GetLines()[0])
	})

	s.Run("Error in show webhook", func() {
		printer.Clean()

		mockError := model.AppError{Id: "Mock Error"}

		s.client.
			EXPECT().
			GetIncomingWebhook(nonExistentID, "").
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		s.client.
			EXPECT().
			GetOutgoingWebhook(nonExistentID).
			Return(nil, &model.Response{Error: &mockError}).
			Times(1)

		err := showWebhookCmdF(s.client, &cobra.Command{}, []string{nonExistentID})
		s.Require().Error(err)
		s.Len(printer.GetLines(), 0)
		s.Len(printer.GetErrorLines(), 0)
		s.Require().Equal("Webhook with id '"+nonExistentID+"' not found", err.Error())
	})
}
