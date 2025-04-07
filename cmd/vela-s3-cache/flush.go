// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

const flushAction = "flush"

// Flush represents the plugin configuration for flush information.
type Flush struct {
	// sets the name of the bucket
	Bucket string
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

	bytesFreedCounter := uint64(0)

	// set a timeout on the request to the cache provider
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logrus.Infof("processing cached objects in path %s", f.Namespace)

	opts := minio.ListObjectsOptions{
		Prefix:    f.Namespace,
		Recursive: true,
	}
	// lists all objects matching the path
	// in the specified bucket
	objectCh := mc.ListObjects(ctx, f.Bucket, opts)
	for object := range objectCh {
		// we got at least one object
		objectsExist = true

		if object.Err != nil {
			return fmt.Errorf("unable to retrieve object %s: %w", object.Key, object.Err)
		}

		//nolint:gosec // G115: integer overflow conversion should be handled via max()
		objSize := uint64(max(0, object.Size))
		humanSize := humanize.Bytes(objSize)

		logrus.Infof("  - %s; last modified: %s; size: %s", object.Key, object.LastModified.String(), humanSize)

		// determine time in the past for flush cut off
		timeInPast := time.Now().Add(-f.Age)

		// check if the object meets the flush age
		if object.LastModified.Before(timeInPast) {
			logrus.Infof("    ├ '%s' flush age criteria met. removing object.", f.Age)

			// remove the object from the bucket
			err := mc.RemoveObject(ctx, f.Bucket, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				return err
			}

			// verify that the object is gone, .RemoveObject fails silently
			// if the supplied path leads to an object that doesn't exist
			_, err = mc.StatObject(ctx, f.Bucket, object.Key, minio.StatObjectOptions{})
			if err != nil {
				bytesFreedCounter += objSize

				logrus.Infof("    ├ object successfully removed, %s freed", humanSize)
			} else {
				return fmt.Errorf("object %s was not removed: %w", object.Key, err)
			}
		} else {
			logrus.Infof("    ├ '%s' flush age criteria not met. keeping object.", f.Age)
		}
	}

	if !objectsExist {
		logrus.Infof("no cache objects found at %s", f.Path)
	}

	logrus.Infof("cache flush action completed")

	if bytesFreedCounter > 0 {
		logrus.Infof("%s freed in total", humanize.Bytes(bytesFreedCounter))
	}

	return nil
}

// Configure prepares the flush fields for the action to be taken.
func (f *Flush) Configure(repo *Repo) error {
	logrus.Trace("configuring flush action")

	// construct the object path
	path := buildNamespace(repo, f.Prefix, f.Path, "")

	logrus.Debugf("created bucket path %s", path)

	// store it in the namespace
	f.Namespace = path

	return nil
}

// Validate verifies the Flush is properly configured.
func (f *Flush) Validate() error {
	logrus.Trace("validating flush action configuration")

	// verify bucket is provided
	if len(f.Bucket) == 0 {
		return fmt.Errorf("no bucket provided")
	}

	return nil
}
