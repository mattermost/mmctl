package telemetry

type NoOp struct{}

func (n *NoOp) SynchronizeConfig(cfg InitConfig, timedUntilReady int64, factoryInstances map[string]int64, tags []string) {
}

func (n *NoOp) SynchronizeStats() error {
	return nil
}
