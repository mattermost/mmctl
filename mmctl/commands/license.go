package commands

import (
	"errors"
	"io/ioutil"

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
	RunE:    uploadLicenseCmdF,
}

func init() {
	LicenseCmd.AddCommand(UploadLicenseCmd)
	RootCmd.AddCommand(LicenseCmd)
}

func uploadLicenseCmdF(command *cobra.Command, args []string) error {
	c, err := InitClient()
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return errors.New("Enter one license file to upload")
	}

	var fileBytes []byte
	if fileBytes, err = ioutil.ReadFile(args[0]); err != nil {
		return err
	}

	if _, response := c.UploadLicenseFile(fileBytes); response.Error != nil {
		return response.Error
	}

	CommandPrettyPrintln("Uploaded license file")

	return nil
}
