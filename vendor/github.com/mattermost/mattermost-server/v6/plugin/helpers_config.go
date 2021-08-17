// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/utils"
)

// CheckRequiredServerConfiguration implements Helpers.CheckRequiredServerConfiguration
func (p *HelpersImpl) CheckRequiredServerConfiguration(req *model.Config) (bool, error) {
	if req == nil {
		return true, nil
	}

	cfg := p.API.GetConfig()

	mc, err := utils.Merge(cfg, req, nil)
	if err != nil {
		return false, errors.Wrap(err, "could not merge configurations")
	}

	mergedCfg := mc.(model.Config)
	cfgBuf, err := json.Marshal(cfg)
	if err != nil {
		return false, fmt.Errorf("failed to marshal config: %v", err)
	}

	mergedCfgBuf, err := json.Marshal(mergedCfg)
	if err != nil {
		return false, fmt.Errorf("failed to marshal merged config: %v", err)
	}

	if !bytes.Equal(cfgBuf, mergedCfgBuf) {
		return false, nil
	}

	return true, nil
}
