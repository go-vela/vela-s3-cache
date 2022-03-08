// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	// ErrInvalidAction defines the error type when the
	// Action provided to the Plugin is unsupported.
	ErrInvalidAction = errors.New("invalid action provided")
)

// Plugin represents the required information for structs.
type Plugin struct {
	// config arguments loaded for the plugin
	Config *Config
	// flush arguments loaded for the plugin
	Flush *Flush
	// rebuild arguments loaded for the plugin
	Rebuild *Rebuild
	// restore arguments loaded for the plugin
	Restore *Restore
	// repo settings loaded for the plugin
	Repo *Repo
}

// Exec runs the plugin with the settings passed from user.
func (p *Plugin) Exec() (err error) {
	logrus.Info("s3 cache plugin starting...")

	// create a minio client
	logrus.Info("creating an s3 client")

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
			"%w: %s (Valid actions: %s, %s, %s)",
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

	// validate repo configuration
	err = p.Repo.Validate()
	if err != nil {
		return err
	}

	// validate action specific configuration
	switch p.Config.Action {
	case flushAction:
		err := p.Flush.Configure(p.Repo)
		if err != nil {
			return err
		}

		// validate flush action
		return p.Flush.Validate()
	case rebuildAction:
		err := p.Rebuild.Configure(p.Repo)
		if err != nil {
			return err
		}

		// validate rebuild action
		return p.Rebuild.Validate()
	case restoreAction:
		err := p.Restore.Configure(p.Repo)
		if err != nil {
			return err
		}

		// validate restore action
		return p.Restore.Validate()
	default:
		return fmt.Errorf(
			"%w: %s (Valid actions: %s, %s, %s)",
			ErrInvalidAction,
			p.Config.Action,
			flushAction,
			rebuildAction,
			restoreAction,
		)
	}
}

// buildNamespace is a helper function to create a namespace
// given a Repo object and path fragment inputs.
func buildNamespace(r *Repo, prefix, path, filename string) string {
	// set the default path for where to store the object
	p := filepath.Join(prefix, r.Owner, r.Name, filename)

	// Path was supplied and will override default
	if len(path) > 0 {
		p = filepath.Join(path, filename)
	}

	return filepath.Clean(p)
}
