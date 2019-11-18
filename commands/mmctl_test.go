package commands

import (
	"fmt"
	"os"

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

type MmctlE2ETestSuite struct {
	suite.Suite
	th *TestHelper
}

func (s *MmctlE2ETestSuite) SetupSuite() {
	printer.SetFormat(printer.FORMAT_JSON)

	var err error
	if s.th, err = setupTestHelper(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initialising E2E test helper. %s\n", err)
		fmt.Fprintln(os.Stderr, "Aborting E2E test execution")
		os.Exit(1)
	}
}

func (s *MmctlE2ETestSuite) TearDownTest() {
	printer.Clean()
}
