// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	outputJSON bool

	Version = "6.2.0"
	// SHA1 from git, output of $(git rev-parse HEAD)
	gitCommit = "dev mode"
	// State of git tree, either "clean" or "dirty"
	gitTreeState = "unknown"
	// Build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate = "unknown"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of mmctl.",
	RunE:  versionCmdF,
}

func init() {
	VersionCmd.Flags().BoolVar(&outputJSON, "json", false, "print JSON instead of text")

	RootCmd.AddCommand(VersionCmd)
}

func versionCmdF(cmd *cobra.Command, args []string) error {
	v := getVersionInfo()
	res := v.String()
	if outputJSON {
		j, err := v.JSONString()
		if err != nil {
			return errors.Wrap(err, "unable to generate JSON from version info")
		}
		res = j
	}
	fmt.Println(res)
	return nil
}

type Info struct {
	Version      string
	GitCommit    string
	GitTreeState string
	BuildDate    string
	GoVersion    string
	Compiler     string
	Platform     string
}

func getVersionInfo() Info {
	return Info{
		Version:      Version,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns the string representation of the version info
func (i *Info) String() string {
	b := strings.Builder{}
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "mmctl:\n")
	fmt.Fprintf(w, "Version:\t%s\n", i.Version)
	fmt.Fprintf(w, "GitCommit:\t%s\n", i.GitCommit)
	fmt.Fprintf(w, "GitTreeState:\t%s\n", i.GitTreeState)
	fmt.Fprintf(w, "BuildDate:\t%s\n", i.BuildDate)
	fmt.Fprintf(w, "GoVersion:\t%s\n", i.GoVersion)
	fmt.Fprintf(w, "Compiler:\t%s\n", i.Compiler)
	fmt.Fprintf(w, "Platform:\t%s\n", i.Platform)

	w.Flush()
	return b.String()
}

// JSONString returns the JSON representation of the version info
func (i *Info) JSONString() (string, error) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
