// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
)

func getUsersFromUserArgs(c client.Client, userArgs []string) []*model.User {
	users := make([]*model.User, 0, len(userArgs))
	for _, userArg := range userArgs {
		user := getUserFromUserArg(c, userArg)
		users = append(users, user)
	}
	return users
}

func getUserFromUserArg(c client.Client, userArg string) *model.User {
	var user *model.User
	user, _ = c.GetUserByEmail(userArg, "")

	if user == nil {
		user, _ = c.GetUserByUsername(userArg, "")
	}

	if user == nil {
		user, _ = c.GetUser(userArg, "")
	}

	return user
}
