// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	// ErrInvalidAction defines the error type when the
	// Action provided to the Plugin is unsupported.
	ErrInvalidAction = errors.New("invalid action provided")
)

type (

	// Plugin represents the required information for structs
	Plugin struct {
		// config arguments loaded for the plugin
		Config *Config
		// flush arguments loaded for the plugin
		Flush *Flush
		// rebuild arguments loaded for the plugin
		Rebuild *Rebuild
		// restore arguments loaded for the plugin
		Restore *Restore
		// repo settings loaded for the plugin
		Repo Repo
	}

	// Repo represents the available settings for repository
	Repo struct {
		Owner        string
		Name         string
		Branch       string
		CommitBranch string
	}
)

// Exec runs the plugin with the settings passed from user
func (p *Plugin) Exec() (err error) {
	logrus.Info("s3 cache plugin starting...")

	// create a minio client
	logrus.Info("Creating an s3 client")

	mc, err := p.Config.New()
	if err != nil {
		return err
	}

	logrus.Info("s3 client created")

	// execute action specific configuration
	switch p.Config.Action {
	case flushAction:
		// execute flush action
		return p.Flush.Exec(mc)
	case rebuildAction:
		// execute rebuild action
		return p.Rebuild.Exec(mc)
	case restoreAction:
		// execute restore action
		return p.Restore.Exec(mc)
	default:
		return fmt.Errorf(
			"%s: %s (Valid actions: %s, %s, %s)",
			ErrInvalidAction,
			p.Config.Action,
			flushAction,
			rebuildAction,
			restoreAction,
		)
	}
}

// Validate verifies the Config is properly configured.
func (p *Plugin) Validate() error {
	logrus.Debug("validating plugin configuration")

	// validate config configuration
	err := p.Config.Validate()
	if err != nil {
		return err
	}

	// validate action specific configuration
	switch p.Config.Action {
	case flushAction:
		err := p.Flush.Configure(p.Repo)
		if err != nil {
			return nil
		}

		// validate flush action
		return p.Flush.Validate()
	case rebuildAction:
		err := p.Rebuild.Configure(p.Repo)
		if err != nil {
			return nil
		}

		// validate rebuild action
		return p.Rebuild.Validate()
	case restoreAction:
		err := p.Restore.Configure(p.Repo)
		if err != nil {
			return nil
		}

		// validate restore action
		return p.Restore.Validate()
	default:
		return fmt.Errorf(
			"%s: %s (Valid actions: %s, %s, %s)",
			ErrInvalidAction,
			p.Config.Action,
			flushAction,
			rebuildAction,
			restoreAction,
		)
	}
}
