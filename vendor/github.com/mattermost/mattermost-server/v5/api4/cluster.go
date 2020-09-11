// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (api *API) InitCluster() {
	api.BaseRoutes.Cluster.Handle("/status", api.ApiSessionRequired(getClusterStatus)).Methods("GET")
}

func getClusterStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT) {
		c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT)
		return
	}

	if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin {
		c.Err = model.NewAppError("getClusterStatus", "api.restricted_system_admin", nil, "", http.StatusForbidden)
		return
	}

	infos := c.App.GetClusterStatus()
	w.Write([]byte(model.ClusterInfosToJson(infos)))
}
