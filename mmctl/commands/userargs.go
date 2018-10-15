package commands

import (
	"github.com/mattermost/mattermost-server/model"
)

func getUsersFromUserArgs(c *model.Client4, userArgs []string) []*model.User {
	users := make([]*model.User, 0, len(userArgs))
	for _, userArg := range userArgs {
		user := getUserFromUserArg(c, userArg)
		users = append(users, user)
	}
	return users
}

func getUserFromUserArg(c *model.Client4, userArg string) *model.User {
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
