// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"net/url"
	"strings"

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
	if !checkDots(userArg) {
		user, _ = c.GetUserByEmail(userArg, "")
	}

	if !checkSlash(userArg) {
		if user == nil {
			user, _ = c.GetUserByUsername(userArg, "")
		}

		if user == nil {
			user, _ = c.GetUser(userArg, "")
		}
	}

	return user
}

// returns true if slash is found in the arg
func checkSlash(arg string) bool {
	unescapedArg, _ := url.PathUnescape(arg)
	return strings.Contains(unescapedArg, "/")
}

// returns true if double dot is found in the arg
func checkDots(arg string) bool {
	unescapedArg, _ := url.PathUnescape(arg)
	return strings.Contains(unescapedArg, "..")
}

func getUsersFromArgs(c client.Client, userArgs []string) ([]*model.User, *FindEntitySummary) {
	users := make([]*model.User, 0, len(userArgs))
	errors := make([]error, 0)
	for _, userArg := range userArgs {
		user, err := getUserFromArg(c, userArg)
		if err != nil {
			errors = append(errors, err)
		} else {
			users = append(users, user)
		}
	}
	if len(errors) > 0 {
		summary := &FindEntitySummary{
			Errors: errors,
		}
		return users, summary
	}
	return users, nil
}

func getUserFromArg(c client.Client, userArg string) (*model.User, error) {
	var (
		user     *model.User
		response *model.Response
	)
	if !checkDots(userArg) {
		user, response = c.GetUserByEmail(userArg, "")
		if isErrorSevere(response) {
			return nil, response.Error
		}
	}

	if !checkSlash(userArg) {
		if user == nil {
			user, response = c.GetUserByUsername(userArg, "")
			if isErrorSevere(response) {
				return nil, response.Error
			}
		}

		if user == nil {
			user, response = c.GetUser(userArg, "")
			if isErrorSevere(response) {
				return nil, response.Error
			}
		}
	}

	if user == nil {
		return nil, ErrEntityNotFound{Type: "user", ID: userArg}
	}

	return user, nil
}
