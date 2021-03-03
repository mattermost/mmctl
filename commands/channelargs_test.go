// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (s *MmctlUnitTestSuite) TestGetChannelArgs() {
	s.Run("channel not found", func() {
		notFoundChannel := "notfoundchannel"
		notFoundErr := &model.AppError{Message: "channel not found", StatusCode: http.StatusNotFound}

		s.client.
			EXPECT().
			GetChannel(notFoundChannel, "").
			Return(nil, &model.Response{Error: notFoundErr}).
			Times(1)

		channels, err := getChannelsFromArgs(s.client, []string{notFoundChannel})
		s.Require().Len(channels, 0)
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* channel %s not found\n\n", notFoundChannel), err.Error())
	})
	s.Run("bad request", func() {
		badRequestChannel := "badrequest"
		badRequestErr := &model.AppError{Message: "channel bad request", StatusCode: http.StatusBadRequest}

		s.client.
			EXPECT().
			GetChannel(badRequestChannel, "").
			Return(nil, &model.Response{Error: badRequestErr}).
			Times(1)

		channels, err := getChannelsFromArgs(s.client, []string{badRequestChannel})
		s.Require().Len(channels, 0)
		s.Require().NotNil(err)
		s.Require().Equal(fmt.Sprintf("1 error occurred:\n\t* channel %s not found\n\n", badRequestChannel), err.Error())
	})
	s.Run("forbidden", func() {
		forbidden := "forbidden"
		forbiddenErr := &model.AppError{Message: "channel forbidden", StatusCode: http.StatusForbidden}

		s.client.
			EXPECT().
			GetChannel(forbidden, "").
			Return(nil, &model.Response{Error: forbiddenErr}).
			Times(1)

		channels, err := getChannelsFromArgs(s.client, []string{forbidden})
		s.Require().Len(channels, 0)
		s.Require().NotNil(err)
		s.Require().Equal("1 error occurred:\n\t* : channel forbidden, \n\n", err.Error())
	})
	s.Run("internal server error", func() {
		errChannel := "internalServerError"
		internalServerErrorErr := &model.AppError{Message: "channel internalServerError", StatusCode: http.StatusInternalServerError}

		s.client.
			EXPECT().
			GetChannel(errChannel, "").
			Return(nil, &model.Response{Error: internalServerErrorErr}).
			Times(1)

		channels, err := getChannelsFromArgs(s.client, []string{errChannel})
		s.Require().Len(channels, 0)
		s.Require().NotNil(err)
		s.Require().Equal("1 error occurred:\n\t* : channel internalServerError, \n\n", err.Error())
	})
	s.Run("success", func() {
		successID := "success"
		successChannel := &model.Channel{Id: successID}

		s.client.
			EXPECT().
			GetChannel(successID, "").
			Return(successChannel, nil).
			Times(1)

		channels, summary := getChannelsFromArgs(s.client, []string{successID})
		s.Require().Nil(summary)
		s.Require().Len(channels, 1)
		s.Require().Equal(successChannel, channels[0])
	})

	s.Run("success with team on channel", func() {
		channelID := "success"
		teamID := "myTeamID"
		successTeam := &model.Team{Id: teamID}
		successChannel := &model.Channel{Id: channelID}

		s.client.
			EXPECT().
			GetTeam(teamID, "").
			Return(successTeam, nil).
			Times(1)
		s.client.
			EXPECT().
			GetChannelByNameIncludeDeleted(channelID, teamID, "").
			Return(successChannel, nil).
			Times(1)

		channels, summary := getChannelsFromArgs(s.client, []string{fmt.Sprintf("%v:%v", teamID, channelID)})
		s.Require().Nil(summary)
		s.Require().Len(channels, 1)
		s.Require().Equal(successChannel, channels[0])
	})
}
