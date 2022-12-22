package mocks

import "github.com/splitio/go-split-commons/v4/dtos"

type MockImpressionsCountStorage struct {
	RecordImpressionsCountCall func(impressions dtos.ImpressionsCountDTO) error
	GetImpressionsCountCall    func() (dtos.ImpressionsCountDTO, error)
}

func (m MockImpressionsCountStorage) RecordImpressionsCount(impressions dtos.ImpressionsCountDTO) error {
	return m.RecordImpressionsCountCall(impressions)
}

func (m MockImpressionsCountStorage) GetImpressionsCount() (dtos.ImpressionsCountDTO, error) {
	return m.GetImpressionsCountCall()
}
