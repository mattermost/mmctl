package application

type Dummy struct{}

func (d *Dummy) NotifyEvent(monitorType int)      {}
func (d *Dummy) Reset(monitorType int, value int) {}

var _ MonitorProducerInterface = (*Dummy)(nil)
