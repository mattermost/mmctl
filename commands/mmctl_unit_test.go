// +build unit

package commands

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMmctlUnitSuite(t *testing.T) {
	suite.Run(t, new(MmctlUnitTestSuite))
}
