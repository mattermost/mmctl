// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckVersionMatch(t *testing.T) {
	testCases := []struct {
		Name          string
		Version       string
		ServerVersion string
		Expected      bool
	}{
		{
			Name:          "Both versions are equal",
			Version:       "1.2.3",
			ServerVersion: "1.2.3",
			Expected:      true,
		},
		{
			Name:          "Only patch version is different",
			Version:       "1.2.3",
			ServerVersion: "1.2.7",
			Expected:      true,
		},
		{
			Name:          "Major version is greater",
			Version:       "1.2.3",
			ServerVersion: "2.2.3",
			Expected:      false,
		},
		{
			Name:          "Major version is less",
			Version:       "1.2.3",
			ServerVersion: "0.2.3",
			Expected:      false,
		},
		{
			Name:          "Minor version is greater",
			Version:       "1.2.3",
			ServerVersion: "1.3.3",
			Expected:      false,
		},
		{
			Name:          "Minor version is less",
			Version:       "1.2.3",
			ServerVersion: "1.1.3",
			Expected:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			res := CheckVersionMatch(tc.Version, tc.ServerVersion)

			require.Equal(t, tc.Expected, res)
		})
	}
}

func TestVerifyCertificates(t *testing.T) {
	testCases := []struct {
		Name          string
		Chains        [][]*x509.Certificate
		ExpectedError bool
	}{
		{
			Name: "One chain with a root SHA1 cert",
			Chains: [][]*x509.Certificate{
				{
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
				},
			},
			ExpectedError: false,
		},
		{
			Name: "One chain with an intermediate SHA1 cert",
			Chains: [][]*x509.Certificate{
				{
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
				},
			},
			ExpectedError: true,
		},
		{
			Name: "One valid chain and other invalid",
			Chains: [][]*x509.Certificate{
				{
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
				},
				{
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
				},
			},
			ExpectedError: false,
		},
		{
			Name: "Two invalid chains",
			Chains: [][]*x509.Certificate{
				{
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
					{SignatureAlgorithm: x509.SHA1WithRSA},
				},
				{
					{SignatureAlgorithm: x509.SHA256WithRSA},
					{SignatureAlgorithm: x509.DSAWithSHA1},
					{SignatureAlgorithm: x509.DSAWithSHA1},
				},
			},
			ExpectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := VerifyCertificates([][]byte{}, tc.Chains)
			if tc.ExpectedError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
