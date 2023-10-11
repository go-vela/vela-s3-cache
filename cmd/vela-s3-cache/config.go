// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// Config represents the plugin configuration for s3 config information.
type Config struct {
	// action to perform against the s3 instance
	Action              string
	Server              string
	AcceleratedEndpoint string
	AccessKey           string
	SecretKey           string
	SessionToken        string
	Region              string
}

// New creates an Minio client for managing artifacts.
func (c *Config) New() (*minio.Client, error) {
	logrus.Trace("creating new Minio client from plugin configuration")

	// default to amazon aws s3 storage
	endpoint := "s3.amazonaws.com"
	useSSL := true

	if len(c.Server) > 0 {
		useSSL = strings.HasPrefix(c.Server, "https://")

		if !useSSL {
			if !strings.HasPrefix(c.Server, "http://") {
				return nil, fmt.Errorf("invalid server %s: must to be a HTTP URI", c.Server)
			}

			endpoint = c.Server[7:]
		} else {
			endpoint = c.Server[8:]
		}
	}

	var creds *credentials.Credentials
	if len(c.AccessKey) > 0 && len(c.SecretKey) > 0 {
		creds = credentials.NewStaticV4(c.AccessKey, c.SecretKey, c.SessionToken)
	} else {
		creds = credentials.NewIAM("")

		// See if the IAM role can be retrieved
		_, err := creds.Get()
		if err != nil {
			return nil, err
		}
	}

	opts := &minio.Options{
		Creds:  creds,
		Secure: useSSL,
	}

	mc, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}

	if c.AcceleratedEndpoint != "" {
		mc.SetS3TransferAccelerate(c.AcceleratedEndpoint)
	}

	return mc, nil
}

// Validate verifies the Config is properly configured.
func (c *Config) Validate() error {
	logrus.Trace("validating config plugin configuration")

	// verify server is provided
	if len(c.Server) == 0 {
		return fmt.Errorf("no cache server provided")
	}

	// verify access key is provided
	if len(c.AccessKey) == 0 {
		return fmt.Errorf("no access key provided")
	}

	// verify secret key is provided
	if len(c.SecretKey) == 0 {
		return fmt.Errorf("no secret key provided")
	}

	// verify action is provided
	if len(c.Action) == 0 {
		return fmt.Errorf("no config action provided")
	}

	return nil
}
