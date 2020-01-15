// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/printer"
	"github.com/spf13/cobra"
)

func (s *MmctlUnitTestSuite) TestTeamUsersArchiveCmd() {
	teamArg := "example-team-id"
	userArg := "example-user-id"

	s.Run("Remove users from team with a non-existent team returns an error", func() {
		printer.Clean()

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{teamArg, userArg})
		s.Require().Equal(err.Error(), "Unable to find team '"+teamArg+"'")
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Remove users from team with a non-existent user returns an error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{teamArg, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Can't find user '"+userArg+"'")
	})

	s.Run("Remove users from team by email and get team by name should not return an error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team by email and get team should not return an error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team by username and get team should not return an error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team by user and get team should not return an error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg}
		mockUser := &model.User{Id: userArg}
		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername(mockUser.Id, "").
			Return(nil, nil).
			Times(1)

		s.client.
			EXPECT().
			GetUser(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})

	s.Run("Remove users from team with an erroneous RemoveTeamMember should return an error", func() {
		printer.Clean()
		mockTeam := &model.Team{Id: teamArg, Name: "example-name"}
		mockUser := &model.User{Id: userArg}
		mockError := model.AppError{Message: "Mock error"}

		s.client.
			EXPECT().
			GetTeam(teamArg, "").
			Return(mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail(mockUser.Id, "").
			Return(mockUser, nil).
			Times(1)

		s.client.
			EXPECT().
			RemoveTeamMember(mockTeam.Id, mockUser.Id).
			Return(false, &model.Response{Error: &mockError}).
			Times(1)

		err := teamUsersRemoveCmdF(s.client, &cobra.Command{}, []string{mockTeam.Id, mockUser.Id})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Unable to remove '"+mockUser.Id+"' from "+mockTeam.Name+". Error: "+mockError.Error())
	})

}

func (s *MmctlUnitTestSuite) TestAddUsersCmd() {
	mockTeam := model.Team{
		Id:          "TeamId",
		Name:        "team1",
		DisplayName: "DisplayName",
	}
	mockUser := model.User{
		Id:       "UserID",
		Username: "ExampleUser",
		Email:    "example@example.com",
	}

	s.Run("Add users with a team that cannot be found returns error", func() {

		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetTeamByName("team1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersAddCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Equal(err.Error(), "Unable to find team 'team1'")
		s.Require().Len(printer.GetLines(), 0)
	})

	s.Run("Add users with nonexistent user in arguments prints error", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail("user1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByUsername("user1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUser("user1", "").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersAddCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0], "Can't find user 'user1'")
	})

	s.Run("Add users should print error when cannot add team member", func() {
		printer.Clean()
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail("user1", "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		mockError := &model.AppError{
			Message:       "Cannot add team member",
			DetailedError: "This user was banned in this team",
			Where:         "Team.AddTeamMember",
		}

		s.client.
			EXPECT().
			AddTeamMember("TeamId", "UserID").
			Return(nil, &model.Response{Error: mockError}).
			Times(1)

		err := teamUsersAddCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetErrorLines(), 1)
		s.Require().Equal(printer.GetErrorLines()[0],
			"Unable to add 'user1' to team1. Error: Team.AddTeamMember: Cannot add team member, This user was banned in this team")
	})

	s.Run("Add users should not print in console anything on success", func() {
		printer.Clean()

		cmd := &cobra.Command{}
		s.client.
			EXPECT().
			GetTeam("team1", "").
			Return(&mockTeam, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			GetUserByEmail("user1", "").
			Return(&mockUser, &model.Response{Error: nil}).
			Times(1)

		s.client.
			EXPECT().
			AddTeamMember("TeamId", "UserID").
			Return(nil, &model.Response{Error: nil}).
			Times(1)

		err := teamUsersAddCmdF(s.client, cmd, []string{"team1", "user1"})
		s.Require().Nil(err)
		s.Require().Len(printer.GetLines(), 0)
		s.Require().Len(printer.GetErrorLines(), 0)
	})
}
