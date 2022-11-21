package provisional

import (
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/provisional/strategy"
)

// ImpressionManager interface
type ImpressionManager interface {
	ProcessImpressions(impressions []dtos.Impression) ([]dtos.Impression, []dtos.Impression)
	ProcessSingle(impression *dtos.Impression) bool
}

// ImpressionManagerImpl implements
type ImpressionManagerImpl struct {
	processStrategy strategy.ProcessStrategyInterface
}

// NewImpressionManager creates new ImpManager
func NewImpressionManager(processStrategy strategy.ProcessStrategyInterface) ImpressionManager {
	return &ImpressionManagerImpl{
		processStrategy: processStrategy,
	}
}

// ProcessImpressions bulk processes
func (i *ImpressionManagerImpl) ProcessImpressions(impressions []dtos.Impression) ([]dtos.Impression, []dtos.Impression) {
	return i.processStrategy.Apply(impressions)
}

// ProcessSingle accepts a pointer to an impression, updates it's PT accordingly,
// and returns whether it should be sent to the BE and to the lister
func (i *ImpressionManagerImpl) ProcessSingle(impression *dtos.Impression) bool {
	return i.processStrategy.ApplySingle(impression)
}
