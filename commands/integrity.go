// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"fmt"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/cobra"
)

var IntegrityCmd = &cobra.Command{
	Use:    "integrity",
	Short:  "Check database records integrity.",
	Long:   "Perform a relational integrity check which returns information about any orphaned record found.",
	Args:   cobra.NoArgs,
	PreRun: localOnlyPrecheck,
	RunE:   withClient(integrityCmdF),
}

func init() {
	IntegrityCmd.Flags().Bool("confirm", false, "Confirm you really want to run a complete integrity check that may temporarily harm system performance")
	IntegrityCmd.Flags().BoolP("verbose", "v", false, "Show detailed information on integrity check results")
	RootCmd.AddCommand(IntegrityCmd)
}

func printRelationalIntegrityCheckResult(data model.RelationalIntegrityCheckData, verbose bool) {
	printer.PrintT("Found {{len .Records}} in relation {{ .ChildName }} orphans of relation {{ .ParentName }}", data)
	if !verbose {
		return
	}
	const null = "NULL"
	const empty = "empty"
	for _, record := range data.Records {
		var parentID string

		switch {
		case record.ParentId == nil:
			parentID = null
		case *record.ParentId == "":
			parentID = empty
		default:
			parentID = *record.ParentId
		}

		if record.ChildId != nil {
			if parentID == null || parentID == empty {
				fmt.Printf("  Child %s (%s.%s) has %s ParentIdAttr (%s.%s)\n", *record.ChildId, data.ChildName, data.ChildIdAttr, parentID, data.ChildName, data.ParentIdAttr)
			} else {
				fmt.Printf("  Child %s (%s.%s) is missing Parent %s (%s.%s)\n", *record.ChildId, data.ChildName, data.ChildIdAttr, parentID, data.ChildName, data.ParentIdAttr)
			}
		} else {
			if parentID == null || parentID == empty {
				fmt.Printf("  Child has %s ParentIdAttr (%s.%s)\n", parentID, data.ChildName, data.ParentIdAttr)
			} else {
				fmt.Printf("  Child is missing Parent %s (%s.%s)\n", parentID, data.ChildName, data.ParentIdAttr)
			}
		}
	}
}

func printIntegrityCheckResult(result model.IntegrityCheckResult, verbose bool) {
	switch data := result.Data.(type) {
	case model.RelationalIntegrityCheckData:
		printRelationalIntegrityCheckResult(data, verbose)
	default:
		printer.PrintError("invalid data type")
	}
}

func integrityCmdF(c client.Client, command *cobra.Command, args []string) error {
	confirmFlag, _ := command.Flags().GetBool("confirm")
	if !confirmFlag {
		var confirm string
		fmt.Println("This check may harm performance on live systems. Are you sure you want to proceed? (YES/NO): ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			printer.PrintError("Aborted")
			return nil
		}
	}

	verboseFlag, _ := command.Flags().GetBool("verbose")

	results, resp := c.CheckIntegrity()
	if resp.Error != nil {
		return fmt.Errorf("unable to perform integrity check. Error: %w", resp.Error)
	}
	for _, result := range results {
		if result.Err != nil {
			printer.PrintError(result.Err.Error())
			continue
		}
		printIntegrityCheckResult(result, verboseFlag)
	}

	return nil
}
