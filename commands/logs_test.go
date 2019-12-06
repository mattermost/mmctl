package commands

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/spf13/cobra"
)

const (
	testLogInfo       = `{"level":"info","ts":1573516747,"caller":"app/server.go:490","msg":"Server is listening on [::]:8065"}`
	testLogInfoStdout = "2019-11-11T23:59:07Z info app/server.go:490 Server is listening on [::]:8065"
	testLogrusStdout  = "time=\"2019-11-11T23:59:07Z\" level=info msg=\"Server is listening on [::]:8065\" caller=\"app/server.go:490\""
)

func (s *MmctlUnitTestSuite) TestLogsCmd() {
	s.Run("Display single log line", func() {
		mockSingleLogLine := []string{testLogInfo}
		cmd := &cobra.Command{}
		cmd.Flags().Int("number", 1, "")

		s.client.
			EXPECT().
			GetLogs(0, 1).
			Return(mockSingleLogLine, &model.Response{Error: nil}).
			Times(1)

		data, err := testLogsCmdF(s.client, cmd, []string{})

		s.Require().Nil(err)
		s.Require().Len(data, 1)
		s.EqualValues(testLogInfoStdout, data[0])
	})

	s.Run("Display logs", func() {
		mockSingleLogLine := []string{testLogInfo}
		cmd := &cobra.Command{}

		s.client.
			EXPECT().
			GetLogs(0, 0).
			Return(mockSingleLogLine, &model.Response{Error: nil}).
			Times(1)

		data, err := testLogsCmdF(s.client, cmd, []string{})

		s.Require().Nil(err)
		s.Require().Len(data, 1)
		s.EqualValues(testLogInfoStdout, data[0])
	})

	s.Run("Display logs logrus format", func() {
		mockSingleLogLine := []string{testLogInfo}
		cmd := &cobra.Command{}
		cmd.Flags().Bool("logrus", true, "")
		cmd.Flags().Int("number", 1, "")

		s.client.
			EXPECT().
			GetLogs(0, 1).
			Return(mockSingleLogLine, &model.Response{Error: nil}).
			Times(1)

		data, err := testLogsCmdF(s.client, cmd, []string{})

		s.Require().Nil(err)
		s.Require().Len(data, 1)
		s.EqualValues(testLogrusStdout, data[0])
	})
}

// testLogsCmdF is a wrapper around the logsCmdF function to capture
// stdout for testing
func testLogsCmdF(client client.Client, cmd *cobra.Command, args []string) ([]string, error) {
	// Redirect stdout
	currStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call logsCmdF
	err := logsCmdF(client, cmd, args)

	// Stop capturing, set stdout back
	w.Close()
	os.Stdout = currStdout

	// Copy to buffer
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Split for individual lines, removing last as it is an empty string
	data := strings.Split(buf.String(), "\n")
	data = data[:len(data)-1]

	return data, err
}
