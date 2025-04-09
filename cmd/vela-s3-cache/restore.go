// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"

	"github.com/go-vela/vela-s3-cache/pkg/archiver"
)

const restoreAction = "restore"

// Restore represents the plugin configuration for Restore information.
type Restore struct {
	// sets the name of the bucket
	Bucket string
	// sets the path for where to retrieve the object from
	Path string
	// sets the path for where to retrieve the object from
	Prefix string
	// sets the name of the cache object
	Filename string
	// sets the timeout on the call to s3
	Timeout time.Duration
	// will hold our final namespace for the path to the objects
	Namespace string
}

// Exec formats and runs the actions for restoring a cache in s3.
func (r *Restore) Exec(mc *minio.Client) error {
	logrus.Trace("running restore with provided configuration")

	logrus.Debugf("getting object info on bucket %s from path: %s", r.Bucket, r.Namespace)

	// set a timeout on the request to the cache provider
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	a, err := archiver.NewArchiver("tar.gz")
	if err != nil {
		return err
	}

	// collect metadata on the object
	objInfo, err := mc.StatObject(ctx, r.Bucket, r.Namespace, minio.StatObjectOptions{})
	if objInfo.Key == "" {
		logrus.Error(err)
		return nil
	}

	logrus.Debugf("getting object in bucket %s from path: %s", r.Bucket, r.Namespace)

	//nolint:gosec // G115: integer overflow conversion should be handled via max()
	logrus.Infof("%s to download", humanize.Bytes(uint64(max(0, objInfo.Size))))

	// retrieve the object in specified path of the bucket
	err = mc.FGetObject(ctx, r.Bucket, r.Namespace, r.Filename, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to retrieve object from bucket %s at path %s: %w", r.Bucket, r.Namespace, err)
	}

	defer func() {
		// delete the temporary archive file
		err = os.Remove(r.Filename)
		if err != nil {
			logrus.Debugf("delete of local archive file %s unsuccessful", r.Filename)
		}

		logrus.Debugf("local cache archive %s successfully deleted", r.Filename)
	}()

	stat, err := os.Stat(r.Filename)
	if err != nil {
		return err
	}

	//nolint:gosec // G115: integer overflow conversion should be handled via max()
	logrus.Infof("downloaded %s to %s on local filesystem", humanize.Bytes(uint64(max(0, stat.Size()))), r.Filename)

	logrus.Debug("getting current working directory")

	// grab the current working directory for unpacking the object
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	logrus.Debugf("unarchiving file %s into directory %s", r.Filename, pwd)

	data, err := os.Open(r.Filename)
	if err != nil {
		return err
	}
	defer data.Close()

	// expand the object back onto the filesystem
	err = a.Unarchive(ctx, data, pwd)
	if err != nil {
		return err
	}

	logrus.Infof("successfully restored cache archive %s", r.Filename)

	return nil
}

// Configure prepares the restore fields for the action to be taken.
func (r *Restore) Configure(repo *Repo) error {
	logrus.Trace("configuring restore action")

	// construct the object path
	path := buildNamespace(repo, r.Prefix, r.Path, r.Filename)

	logrus.Debugf("created bucket path %s", path)

	// store it in the namespace
	r.Namespace = path

	return nil
}

// Validate verifies the Restore is properly configured.
func (r *Restore) Validate() error {
	logrus.Trace("validating restore action configuration")

	// verify bucket is provided
	if len(r.Bucket) == 0 {
		return fmt.Errorf("no bucket provided")
	}

	// verify filename is provided
	if len(r.Filename) == 0 {
		return fmt.Errorf("no filename provided")
	}

	// verify timeout is provided
	if r.Timeout == 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	return nil
}
