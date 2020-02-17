// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
)

const flushAction = "flush"

// Flush represents the plugin configuration for flush information.
type Flush struct {
	// sets the name of the bucket
	Root string
	// sets the path prefix for which object(s) should be flushed
	Prefix string
	// sets path where the objects should be flushed
	Path string
	// sets the age of the objects to flush
	Age int
}

// Exec formats and runs the actions for flushing a cache in s3.
func (f *Flush) Exec(mc *minio.Client) error {
	logrus.Trace("running flush with provided configuration")

	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	logrus.Debugf("listing objects in bucket %s under path %s", f.Root, f.Path)
	objectCh := mc.ListObjectsV2(f.Root, f.Path, true, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("Failed to retrieve object %s: %s", object.Key, object.Err)
		}

		if object.LastModified.Before(time.Now().AddDate(0, 0, f.Age*-1)) {
			logrus.Debugf("removing object from bucket %s in path: %s", f.Root, f.Path)
			err := mc.RemoveObject(f.Root, fmt.Sprintf("%s/%s", f.Path, object.Key))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Configure prepares the flush fields for the action to be taken
func (f *Flush) Configure(repo Repo) error {
	logrus.Trace("configuring flush action")

	// configire path based on action
	path := fmt.Sprintf("%s/%s/%s", strings.TrimRight(f.Path, "/"), repo.Owner, repo.Name)

	if len(f.Path) > 0 {
		path = fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(f.Prefix, "/"), repo.Owner, repo.Name, f.Path)
	}
	logrus.Debugf("created bucket path %s", path)

	f.Path = strings.TrimLeft(path, "/")

	return nil
}

// Validate verifies the Flush is properly configured.
func (f *Flush) Validate() error {
	logrus.Trace("validating flush action configuration")

	// verify root is provided
	if len(f.Root) == 0 {
		return fmt.Errorf("no root provided")
	}

	return nil
}
