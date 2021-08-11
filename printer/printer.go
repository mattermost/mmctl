// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const (
	FormatPlain = "plain"
	FormatJSON  = "json"
)

type Printer struct { //nolint
	writer  io.Writer
	eWriter io.Writer

	Format     string
	Single     bool
	Lines      []interface{}
	ErrorLines []interface{}

	cmd        *cobra.Command
	serverAddr string
}

type printOpts struct {
	format    string
	pagerPath string
	single    bool
	usePager  bool
	shortStat bool
}

var printer Printer

func init() {
	printer.writer = os.Stdout
	printer.eWriter = os.Stderr
}

// SetFormat sets the format for the final output of the printer
func SetFormat(t string) {
	printer.Format = t
}

func SetCommand(cmd *cobra.Command) {
	printer.cmd = cmd
}

func SetServerAddres(addr string) {
	printer.serverAddr = addr
}

// SetSingle sets the single flag on the printer. If this flag is set, the
// printer will check the size of stored elements before printing, and
// if there is only one, it will be printed on its own instead of
// inside a list
func SetSingle(single bool) {
	printer.Single = single
}

// PrintT prints an element. Depending on the format, the element can be
// formatted and printed as a structure or used to populate the
// template
func PrintT(templateString string, v interface{}) {
	switch printer.Format {
	case FormatPlain:
		t := template.Must(template.New("").Parse(templateString))
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, v); err != nil {
			PrintError("Can't print the message using the provided template: " + templateString)
		}
		tplString := tpl.String()
		printer.Lines = append(printer.Lines, tplString)
	case FormatJSON:
		printer.Lines = append(printer.Lines, v)
	}
}

// Print an element. If the format requires a template, the element
// will be printed as a structure with field names using the print
// verb %+v
func Print(v interface{}) {
	PrintT("{{printf \"%+v\" .}}", v)
}

// Flush writes the elements accumulated in the printer
func Flush() error {
	opts := printOpts{
		format: printer.Format,
		single: printer.Single,
	}

	cmd := printer.cmd
	if cmd != nil {
		shortStat, err := printer.cmd.Flags().GetBool("short-stat")
		if err == nil && printer.cmd.Name() == "list" && printer.cmd.Parent().Name() != "auth" {
			opts.shortStat = shortStat
		}
	}

	b, err := printer.linesToBytes(opts)
	if err != nil {
		return err
	}
	lines := lineCount(b)

	isTTY := checkInteractiveTerminal() == nil
	enablePager := isTTY && (termHeight(os.Stdout) < lines) // calculate if we should enable paging

	pager := os.Getenv("PAGER")
	if enablePager {
		enablePager = pager != ""
	}

	opts.usePager = enablePager
	opts.pagerPath = pager

	err = printer.printBytes(b, opts)
	if err != nil {
		return err
	}

	// after all, print errors
	printer.printErrors()

	defer func() {
		printer.Lines = []interface{}{}
		printer.ErrorLines = []interface{}{}
	}()

	if cmd == nil || cmd.Name() != "list" || printer.cmd.Parent().Name() == "auth" {
		return nil
	}

	// the command is a list command, we may want to
	// take care of the stat flags
	noStat, err := cmd.Flags().GetBool("no-stat")
	if err != nil {
		return err
	}

	// print stats
	switch {
	case noStat:
		// do nothing
	case !opts.shortStat:
		// should not go to pager
		if isTTY && !enablePager {
			fmt.Fprintf(printer.eWriter, "\n") // add a one line space before statistical data
		}
		fallthrough
	case len(printer.Lines) > 0:
		entity := cmd.Parent().Name()
		container := strings.TrimSuffix(printer.serverAddr, "api/v4")
		if container != "" {
			container = fmt.Sprintf(" on %s", container)
		}
		fmt.Fprintf(printer.eWriter, "There are %d %ss%s\n", len(printer.Lines), entity, container)
	}

	return nil
}

// Clean resets the printer's accumulated lines
func Clean() {
	printer.Lines = []interface{}{}
	printer.ErrorLines = []interface{}{}
}

// GetLines returns the printer's accumulated lines
func GetLines() []interface{} {
	return printer.Lines
}

// GetErrorLines returns the printer's accumulated error lines
func GetErrorLines() []interface{} {
	return printer.ErrorLines
}

// PrintError prints to the stderr.
func PrintError(msg string) {
	printer.ErrorLines = append(printer.ErrorLines, msg)
}

func (p Printer) linesToBytes(opts printOpts) (b []byte, err error) {
	if opts.shortStat {
		return
	}

	switch opts.format {
	case FormatPlain:
		var buf bytes.Buffer
		for i := range p.Lines {
			buf.WriteString(fmt.Sprintf("%s\n", p.Lines[i]))
		}
		b = buf.Bytes()
	case FormatJSON:
		switch {
		case opts.single && len(p.Lines) == 0:
			return
		case opts.single && len(p.Lines) == 1:
			b, err = json.MarshalIndent(p.Lines[0], "", "  ")
		default:
			b, err = json.MarshalIndent(p.Lines, "", "  ")
		}
		b = append(b, '\n')
	}
	return
}

func (p Printer) printBytes(b []byte, opts printOpts) error {
	if !opts.usePager {
		fmt.Fprintf(p.writer, "%s", b)
		return nil
	}

	c := exec.Command(opts.pagerPath) // nolint:gosec

	in, err := c.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not create the stdin pipe: %w", err)
	}

	c.Stdout = p.writer
	c.Stderr = p.eWriter

	go func() {
		defer in.Close()
		_, _ = io.Copy(in, bytes.NewReader(b))
	}()

	if err := c.Start(); err != nil {
		return fmt.Errorf("could not start the pager: %w", err)
	}

	return c.Wait()
}

func (p Printer) printErrors() {
	for i := range printer.ErrorLines {
		fmt.Fprintln(printer.eWriter, printer.ErrorLines[i])
	}
}
