package commands

import (
	"errors"
	"io/ioutil"

	"github.com/mattermost/mattermost-server/model"

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
	RootCmd.AddCommand(LicenseCmd)
}

func uploadLicenseCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Enter one license file to upload")
	}

	fileBytes, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	if _, response := c.UploadLicenseFile(fileBytes); response.Error != nil {
		return response.Error
	}

	CommandPrettyPrintln("Uploaded license file")

	return nil
}

func removeLicenseCmdF(c *model.Client4, cmd *cobra.Command, args []string) error {
	if _, response := c.RemoveLicenseFile(); response.Error != nil {
		return response.Error
	}

	CommandPrettyPrintln("Removed license")

	return nil
}
