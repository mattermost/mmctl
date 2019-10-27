package commands

import (
	"testing"

	"github.com/mattermost/mmctl/mocks"
	"github.com/mattermost/mmctl/printer"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type MmctlUnitTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	client   *mocks.MockClient
}

func (s *MmctlUnitTestSuite) SetupSuite() {
	printer.SetFormat(printer.FORMAT_JSON)
}

func (s *MmctlUnitTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.client = mocks.NewMockClient(s.mockCtrl)
}

func (s *MmctlUnitTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	printer.Clean()
}

func TestMmctlSuite(t *testing.T) {
	suite.Run(t, new(MmctlUnitTestSuite))
}
