// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
)

const flushAction = "flush"

// Flush represents the plugin configuration for flush information.
type Flush struct {
	// sets the name of the bucket
	Root string
	// sets path to the objects to be flushed
	Path string
	// sets the path prefix for the object(s) to be flushed
	Prefix string
	// sets the age of the objects to flush
	Age time.Duration
	// will hold our final namespace for the path to the objects
	Namespace string
}

// Exec formats and runs the actions for flushing a cache in s3.
func (f *Flush) Exec(mc *minio.Client) error {
	logrus.Trace("running flush with provided configuration")

	// temp var for messaging to user
	objectsExist := false

	// create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	logrus.Infof("processing cached objects in path %s", f.Namespace)

	// lists all objects matching the path
	// in the specified bucket
	objectCh := mc.ListObjectsV2(f.Root, f.Namespace, true, doneCh)
	for object := range objectCh {
		// we got at least one object
		objectsExist = true

		if object.Err != nil {
			return fmt.Errorf("unable to retrieve object %s: %s", object.Key, object.Err)
		}

		logrus.Infof("  - %s; last modified: %s", object.Key, object.LastModified.String())

		// determine time in the past for flush cut off
		timeInPast := time.Now().Add(-f.Age)

		// check if the object meets the flush age
		if object.LastModified.Before(timeInPast) {
			logrus.Infof("    ├ '%s' flush age criteria met. removing object.", f.Age)

			// remove the object from the bucket
			err := mc.RemoveObject(f.Root, object.Key)
			if err != nil {
				return err
			}

			// verify that the object is gone, .RemoveObject fails silently
			// if the supplied path leads to an object that doesn't exist
			_, err = mc.StatObject(f.Root, object.Key, minio.StatObjectOptions{})
			if err != nil {
				logrus.Info("    ├ object successfully removed.")
			} else {
				return fmt.Errorf("object %s was not removed: %v", object.Key, err)
			}
		} else {
			logrus.Infof("    ├ '%s' flush age criteria not met. keeping object.", f.Age)
		}
	}

	if !objectsExist {
		logrus.Infof("no cache objects found at %s", f.Path)
	}

	logrus.Infof("cache flush action completed")

	return nil
}

// Configure prepares the flush fields for the action to be taken.
func (f *Flush) Configure(repo *Repo) error {
	logrus.Trace("configuring flush action")

	// set the path for where to look for objects
	path := filepath.Join(f.Prefix, repo.Owner, repo.Name, f.Path)

	if len(path) == 0 {
		return fmt.Errorf("constructed path is empty")
	}

	logrus.Debugf("created bucket path %s", path)

	// clean the path and store in Namespace
	f.Namespace = filepath.Clean(path)

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
