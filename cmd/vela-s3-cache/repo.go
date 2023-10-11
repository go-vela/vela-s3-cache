// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Repo represents the available settings for repository.
type Repo struct {
	Owner       string
	Name        string
	Branch      string
	BuildBranch string
}

// Validate verifies the repo configuration.
func (r *Repo) Validate() error {
	logrus.Trace("validating repo configuration")

	if len(r.Owner) == 0 {
		return fmt.Errorf("no repo owner provided")
	}

	if len(r.Name) == 0 {
		return fmt.Errorf("no repo name provided")
	}

	return nil
}
