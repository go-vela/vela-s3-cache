// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

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
