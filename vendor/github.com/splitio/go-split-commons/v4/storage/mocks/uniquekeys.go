package mocks

type MockUniqueKeysStorage struct {
	CountCall   func() int64
	PopNRawCall func(n int64) ([]string, int64, error)
}

// Count mock
func (m MockUniqueKeysStorage) Count() int64 {
	return m.CountCall()
}

func (m MockUniqueKeysStorage) PopNRaw(n int64) ([]string, int64, error) {
	return m.PopNRawCall(n)
}
