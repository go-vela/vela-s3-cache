// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"fmt"
	"strings"

	"github.com/minio/minio-go/v6"
	"github.com/minio/minio-go/v6/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// Config represents the plugin configuration for s3 config information.
type Config struct {
	// action to perform against the s3 instance
	Action              string
	CaCert              string
	CaCertPath          string
	Server              string
	AcceleratedEndpoint string
	AccessKey           string
	SecretKey           string
	SessionToken        string
	Token               string
	Region              string
}

// New creates an Minio client for managing artifacts.
func (c *Config) New() (*minio.Client, error) {
	logrus.Trace("creating new Minio client from plugin configuration")

	// default to amazon aws s3 storeage
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
		creds = credentials.NewStaticV4(c.AccessKey, c.SecretKey, c.Token)
	} else {
		creds = credentials.NewIAM("")

		// See if the IAM role can be retrieved
		_, err := creds.Get()
		if err != nil {
			return nil, err
		}
	}

	mc, err := minio.NewWithCredentials(endpoint, creds, useSSL, "")
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

	// verify action is provided
	if len(c.Action) == 0 {
		return fmt.Errorf("no config action provided")
	}

	// verify access key is provided
	if len(c.AccessKey) == 0 {
		return fmt.Errorf("no access key providedooooooo") //[here] Revert before making PR.
	}

	// verify secret key is provided
	if len(c.SecretKey) == 0 {
		return fmt.Errorf("no secret key provided")
	}

	return nil
}
