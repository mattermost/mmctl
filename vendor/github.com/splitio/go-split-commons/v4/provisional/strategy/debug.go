package strategy

import (
	"github.com/splitio/go-split-commons/v4/dtos"
)

// DebugImpl struct for debug impression mode strategy.
type DebugImpl struct {
	impressionObserver ImpressionObserver
	listenerEnabled    bool
}

// NewDebugImpl creates new DebugImpl.
func NewDebugImpl(impressionObserver ImpressionObserver, listenerEnabled bool) ProcessStrategyInterface {
	return &DebugImpl{
		impressionObserver: impressionObserver,
		listenerEnabled:    listenerEnabled,
	}
}

func (s *DebugImpl) apply(impression *dtos.Impression) bool {
	impression.Pt, _ = s.impressionObserver.TestAndSet(impression.FeatureName, impression)

	return true
}

// Apply calculate the pt and return the impression.
func (s *DebugImpl) Apply(impressions []dtos.Impression) ([]dtos.Impression, []dtos.Impression) {
	forLog := make([]dtos.Impression, 0, len(impressions))
	forListener := make([]dtos.Impression, 0, len(impressions))

	for index := range impressions {
		s.apply(&impressions[index])
	}

	forLog = impressions

	if s.listenerEnabled {
		forListener = impressions
	}

	return forLog, forListener
}

// ApplySingle description
func (s *DebugImpl) ApplySingle(impression *dtos.Impression) bool {
	return s.apply(impression)
}
