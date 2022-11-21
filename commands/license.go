// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"errors"
	"io/ioutil"

	"github.com/mattermost/mmctl/v6/client"
	"github.com/mattermost/mmctl/v6/printer"

	"github.com/spf13/cobra"
)

var LicenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Licensing commands",
}

var UploadLicenseCmd = &cobra.Command{
	Use:     "upload [license]",
	Short:   "Upload a license.",
	Long:    "Upload a license. Replaces current license.",
	Example: "  license upload /path/to/license/mylicensefile.mattermost-license",
	RunE:    withClient(uploadLicenseCmdF),
}

var UploadLicenseStringCmd = &cobra.Command{
	Use:     "upload-string [license]",
	Short:   "Upload a license from a string.",
	Long:    "Upload a license from a string. Replaces current license.",
	Example: " license upload-string \"mylicensestring\"",
	RunE:    withClient(uploadLicenseStringCmdF),
}

var RemoveLicenseCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove the current license.",
	Long:    "Remove the current license and leave mattermost in Team Edition.",
	Example: "  license remove",
	RunE:    withClient(removeLicenseCmdF),
}

func init() {
	LicenseCmd.AddCommand(UploadLicenseCmd)
	LicenseCmd.AddCommand(RemoveLicenseCmd)
	LicenseCmd.AddCommand(UploadLicenseStringCmd)
	RootCmd.AddCommand(LicenseCmd)
}

func uploadLicenseStringCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("enter one license file to upload")
	}

	licenseBytes := []byte(args[0])

	if _, err := c.UploadLicenseFile(licenseBytes); err != nil {
		return err
	}

	printer.Print("Uploaded license file")

	return nil
}

func uploadLicenseCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("enter one license file to upload")
	}

	fileBytes, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	if _, err := c.UploadLicenseFile(fileBytes); err != nil {
		return err
	}

	printer.Print("Uploaded license file")

	return nil
}

func removeLicenseCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	if _, err := c.RemoveLicenseFile(); err != nil {
		return err
	}

	printer.Print("Removed license")

	return nil
}
