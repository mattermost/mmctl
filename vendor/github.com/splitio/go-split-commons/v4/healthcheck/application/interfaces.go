package application

const (
	// Splits monitor type
	Splits = iota
	// Segments monitor type
	Segments
	// Storage monitor type
	Storage
	// SyncErros monitor type
	SyncErros
)

// MonitorProducerInterface application monitor producer interface
type MonitorProducerInterface interface {
	NotifyEvent(monitorType int)
	Reset(monitorType int, value int)
}
