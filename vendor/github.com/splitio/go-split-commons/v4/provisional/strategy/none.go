package strategy

import (
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
)

// NoneImpl struct for none impression mode strategy.
type NoneImpl struct {
	impressionsCounter *ImpressionsCounter
	uniqueKeysTracker  UniqueKeysTracker
	listenerEnabled    bool
}

// NewNoneImpl creates new NoneImpl.
func NewNoneImpl(impressionCounter *ImpressionsCounter, uniqueKeysTracker UniqueKeysTracker, listenerEnabled bool) ProcessStrategyInterface {
	return &NoneImpl{
		impressionsCounter: impressionCounter,
		uniqueKeysTracker:  uniqueKeysTracker,
		listenerEnabled:    listenerEnabled,
	}
}

func (s *NoneImpl) apply(impression *dtos.Impression, now int64) bool {
	s.impressionsCounter.Inc(impression.FeatureName, now, 1)
	s.uniqueKeysTracker.Track(impression.FeatureName, impression.KeyName)

	return false
}

// Apply track the total amount of evaluations and the unique keys.
func (s *NoneImpl) Apply(impressions []dtos.Impression) ([]dtos.Impression, []dtos.Impression) {
	now := time.Now().UTC().UnixNano()
	forListener := make([]dtos.Impression, 0, len(impressions))

	for index := range impressions {
		s.apply(&impressions[index], now)
	}

	if s.listenerEnabled {
		forListener = impressions
	}

	return make([]dtos.Impression, 0, 0), forListener
}

// ApplySingle description
func (s *NoneImpl) ApplySingle(impression *dtos.Impression) bool {
	now := time.Now().UTC().UnixNano()

	return s.apply(impression, now)
}
