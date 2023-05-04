package mocks

// MockFilter is a mocked implementation of BloomFilter
type MockFilter struct {
	AddCall      func(data string)
	ContainsCall func(data string) bool
	ClearCall    func()
}

// Add mock
func (m MockFilter) Add(data string) {
	m.AddCall(data)
}

// Contains mock
func (m MockFilter) Contains(data string) bool {
	return m.ContainsCall(data)
}

// Clear mock
func (m MockFilter) Clear() {
	m.ClearCall()
}
