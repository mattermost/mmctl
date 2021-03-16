// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import (
	"io"

	"github.com/mattermost/mattermost-server/v5/model"
)

type OauthProvider interface {
	GetUserFromJson(data io.Reader, tokenUser *model.User) (*model.User, error)
	GetSSOSettings(config *model.Config, service string) (*model.SSOSettings, error)
	GetUserFromIdToken(idToken string) (*model.User, error)
}

var oauthProviders = make(map[string]OauthProvider)

func RegisterOauthProvider(name string, newProvider OauthProvider) {
	oauthProviders[name] = newProvider
}

func GetOauthProvider(name string) OauthProvider {
	provider, ok := oauthProviders[name]
	if ok {
		return provider
	}
	return nil
}
