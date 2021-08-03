// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import (
	"github.com/mattermost/mattermost-server/v6/app/request"
	"github.com/mattermost/mattermost-server/v6/model"
)

type SamlInterface interface {
	ConfigureSP() error
	BuildRequest(relayState string) (*model.SamlAuthRequest, *model.AppError)
	DoLogin(c *request.Context, encodedXML string, relayState map[string]string) (*model.User, *model.AppError)
	GetMetadata() (string, *model.AppError)
	CheckProviderAttributes(SS *model.SamlSettings, ouser *model.User, patch *model.UserPatch) string
}
