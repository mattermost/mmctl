package matchers

import (
	"fmt"
	"strings"
)

// ContainsStringMatcher matches strings contain one of the substrings in the split
type ContainsStringMatcher struct {
	Matcher
	substrings []string
}

// Match returns true if the key contains one of the substrings in the split
func (m *ContainsStringMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	matchingKey, err := m.matchingKey(key, attributes)
	if err != nil {
		m.logger.Warning(fmt.Sprintf("ContainsAllOfSetMatcher: %s", err.Error()))
		return false
	}

	asString, ok := matchingKey.(string)
	if !ok {
		m.logger.Error("ContainsAllOfSetMatcher: Failed to type-assert string")
		return false
	}

	for _, substring := range m.substrings {
		if strings.Contains(asString, substring) {
			return true
		}
	}

	return false
}

// NewContainsStringMatcher returns a new instance of ContainsStringMatcher
func NewContainsStringMatcher(negate bool, substrings []string, attributeName *string) *ContainsStringMatcher {
	return &ContainsStringMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		substrings: substrings,
	}
}
