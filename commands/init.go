// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"
)

var (
	insecureSignatureAlgorithms = map[x509.SignatureAlgorithm]bool{
		x509.SHA1WithRSA:   true,
		x509.DSAWithSHA1:   true,
		x509.ECDSAWithSHA1: true,
	}
	expectedSocketMode os.FileMode = os.ModeSocket | 0600
)

func CheckVersionMatch(version, serverVersion string) bool {
	maj, min, _ := model.SplitVersion(version)
	srvMaj, srvMin, _ := model.SplitVersion(serverVersion)

	return maj == srvMaj && min == srvMin
}

func withClient(fn func(c client.Client, cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("local") {
			c, err := InitUnixClient(viper.GetString("local-socket-path"))
			if err != nil {
				return err
			}
			return fn(c, cmd, args)
		}

		c, serverVersion, err := InitClient(viper.GetBool("insecure-sha1-intermediate"))
		if err != nil {
			return err
		}
		valid := CheckVersionMatch(Version, serverVersion)
		if !valid {
			if viper.GetBool("strict") {
				printer.PrintError("ERROR: server version " + serverVersion + " doesn't match with mmctl version " + Version + ". Strict flag is set, so the command will not be run")
				os.Exit(1)
			}
			printer.PrintError("WARNING: server version " + serverVersion + " doesn't match mmctl version " + Version)
		}

		return fn(c, cmd, args)
	}
}

func localOnlyPrecheck(cmd *cobra.Command, args []string) {
	local := viper.GetBool("local")
	if !local {
		fmt.Fprintln(os.Stderr, "This command can only be run in local mode")
		os.Exit(1)
	}
}

func disableLocalPrecheck(cmd *cobra.Command, args []string) {
	local := viper.GetBool("local")
	if local {
		fmt.Fprintln(os.Stderr, "This command cannot be run in local mode")
		os.Exit(1)
	}
}

func isValidChain(chain []*x509.Certificate) bool {
	// check all certs but the root one
	certs := chain[:len(chain)-1]

	for _, cert := range certs {
		if _, ok := insecureSignatureAlgorithms[cert.SignatureAlgorithm]; ok {
			return false
		}
	}
	return true
}

func VerifyCertificates(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	// loop over certificate chains
	for _, chain := range verifiedChains {
		if isValidChain(chain) {
			return nil
		}
	}
	return fmt.Errorf("insecure algorithm found in the certificate chain. Use --insecure-sha1-intermediate flag to ignore. Aborting")
}

func NewAPIv4Client(instanceURL string, allowInsecure bool) *model.Client4 {
	client := model.NewAPIv4Client(instanceURL)
	userAgent := fmt.Sprintf("mmctl/%s (%s)", Version, runtime.GOOS)
	client.HttpHeader = map[string]string{"User-Agent": userAgent}

	if !allowInsecure {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				VerifyPeerCertificate: VerifyCertificates,
			},
		}

		client.HttpClient = &http.Client{Transport: transport}
	}

	return client
}

func InitClientWithUsernameAndPassword(username, password, instanceURL string, allowInsecure bool) (*model.Client4, string, error) {
	client := NewAPIv4Client(instanceURL, allowInsecure)

	_, response := client.Login(username, password)
	if response.Error != nil {
		return nil, "", response.Error
	}
	return client, response.ServerVersion, nil
}

func InitClientWithMFA(username, password, mfaToken, instanceURL string, allowInsecure bool) (*model.Client4, string, error) {
	client := NewAPIv4Client(instanceURL, allowInsecure)
	_, response := client.LoginWithMFA(username, password, mfaToken)
	if response.Error != nil {
		return nil, "", response.Error
	}
	return client, response.ServerVersion, nil
}

func InitClientWithCredentials(credentials *Credentials, allowInsecure bool) (*model.Client4, string, error) {
	client := NewAPIv4Client(credentials.InstanceURL, allowInsecure)

	client.AuthType = model.HEADER_BEARER
	client.AuthToken = credentials.AuthToken

	_, response := client.GetMe("")
	if response.Error != nil {
		return nil, "", response.Error
	}

	return client, response.ServerVersion, nil
}

func InitClient(allowInsecure bool) (*model.Client4, string, error) {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return nil, "", err
	}
	return InitClientWithCredentials(credentials, allowInsecure)
}

func InitWebSocketClient() (*model.WebSocketClient, error) {
	credentials, err := GetCurrentCredentials()
	if err != nil {
		return nil, err
	}
	client, appErr := model.NewWebSocketClient4(strings.Replace(credentials.InstanceURL, "http", "ws", 1), credentials.AuthToken)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "unable to create the websockets connection")
	}
	return client, nil
}

func InitUnixClient(socketPath string) (*model.Client4, error) {
	if err := checkValidSocket(socketPath); err != nil {
		return nil, err
	}

	tr := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	client := model.NewAPIv4Client("http://_")
	client.HttpClient = &http.Client{Transport: tr}

	return client, nil
}
