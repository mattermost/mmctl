package telemetry

import (
	"github.com/splitio/go-split-commons/v4/dtos"
)

type NoOp struct{}

func (n *NoOp) SynchronizeConfig(cfg InitConfig, timedUntilReady int64, factoryInstances map[string]int64, tags []string) {
}

func (n *NoOp) SynchronizeStats() error {
	return nil
}

func (n *NoOp) SynchronizeUniqueKeys(uniques dtos.Uniques) error {
	return nil
}
