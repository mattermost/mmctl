// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost-server/v6/model"
)

func (s *Server) License() *model.License {
	license, _ := s.licenseValue.Load().(*model.License)
	return license
}
