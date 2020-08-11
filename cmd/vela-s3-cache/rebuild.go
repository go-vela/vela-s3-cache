// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-vela/archiver"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
)

const rebuildAction = "rebuild"

// Rebuild represents the plugin configuration for rebuild information.
type Rebuild struct {
	// sets the name of the bucket
	Root string
	// sets the path for where to store the object
	Prefix string
	// sets the name of the cache object
	Filename string
	// sets the timeout on the call to s3
	Timeout time.Duration
	// sets the file or directories locations to build your cache from
	Mount []string
}

// Exec formats and runs the actions for rebuilding a cache in s3.
func (r *Rebuild) Exec(mc *minio.Client) error {
	logrus.Trace("running rebuild with provided configuration")

	logrus.Debug("making path /tmp for artifact archive")

	// create a new tmp directory to build the upload object
	err := os.Mkdir("/tmp", 0400)
	if err != nil {
		return err
	}

	t := archiver.NewTarGz()

	f := fmt.Sprintf("/tmp/%s", r.Filename)
	logrus.Debugf("archiving artifact in path %s", f)

	// archive the objects in the mount path provided
	err = t.Archive(r.Mount, f)
	if err != nil {
		return err
	}

	logrus.Debugf("opening artifact in path %s", f)

	obj, err := os.Open(f)
	if err != nil {
		return err
	}

	// set a timeout on the request to the cache provider
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	logrus.Debugf("putting file %s in bucket %s in path: %s", f, r.Root, r.Prefix)

	// upload the object to the specified location in the bucket
	n, err := mc.PutObjectWithContext(ctx, r.Root, r.Prefix, obj, -1, minio.PutObjectOptions{ContentType: "application/tar"})
	if err != nil {
		return err
	}

	u := uint64(n)
	logrus.Infof("%s data rebuilt", humanize.Bytes(u))

	return nil
}

// Configure prepares the rebuild fields for the action to be taken
func (r *Rebuild) Configure(repo Repo) error {
	logrus.Trace("configuring rebuild action")

	// set the default prefix of where to save the object
	path := fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(r.Prefix, "/"), repo.Owner, repo.Name, r.Filename)

	logrus.Debugf("created bucket path %s", path)

	r.Prefix = strings.TrimLeft(path, "/")

	return nil
}

// Validate verifies the Rebuild is properly configured.
func (r *Rebuild) Validate() error {
	logrus.Trace("validating rebuild action configuration")

	// verify root is provided
	if len(r.Root) == 0 {
		return fmt.Errorf("no root provided")
	}

	// verify filename is provided
	if len(r.Filename) == 0 {
		return fmt.Errorf("no filename provided")
	}

	// verify timeout is provided
	if strings.EqualFold(r.Timeout.String(), "0s") {
		return fmt.Errorf("no timeout provided")
	}

	// verify mount is provided
	if len(r.Mount) == 0 {
		return fmt.Errorf("no mount provided")
	}

	return nil
}
