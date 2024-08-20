// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"

	"github.com/go-vela/archiver/v3"
)

const rebuildAction = "rebuild"

// Rebuild represents the plugin configuration for rebuild information.
type Rebuild struct {
	// sets the name of the bucket
	Bucket string
	// sets the path for where to store the object
	Path string
	// sets the prefix for where to store the object
	Prefix string
	// sets the name of the cache object
	Filename string
	// sets the timeout on the call to s3
	Timeout time.Duration
	// sets the file or directories locations to build your cache from
	Mount []string
	// will hold our final namespace for the path to the objects
	Namespace string
	// whether to preserve the relative directory structure during the tar process
	PreservePath bool
}

// Exec formats and runs the actions for rebuilding a cache in s3.
func (r *Rebuild) Exec(mc *minio.Client) error {
	logrus.Trace("running rebuild with provided configuration")

	t := archiver.NewTarGz()
	t.PreservePath = r.PreservePath

	logrus.Debug("determining temp directory for archive")

	f := filepath.Join(os.TempDir(), r.Filename)

	logrus.Debugf("archiving artifact in path %s", f)

	// archive the objects in the mount path provided
	err := t.Archive(r.Mount, f)
	if err != nil {
		return err
	}

	stat, err := os.Stat(f)
	if err != nil {
		return err
	}

	logrus.Infof("archive %s created, %s", f, humanize.Bytes(uint64(stat.Size())))

	logrus.Debugf("opening artifact %s for reading", f)

	obj, err := os.Open(f)
	if err != nil {
		return err
	}

	logrus.Debugf("archive %s opened for reading", f)

	// set a timeout on the request to the cache provider
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	logrus.Debugf("putting archive %s in bucket %s in path: %s", f, r.Bucket, r.Namespace)

	// create an options object for the upload
	mObj := minio.PutObjectOptions{
		ContentType: "application/tar",
	}

	// upload the object to the specified location in the bucket
	n, err := mc.PutObject(ctx, r.Bucket, r.Namespace, obj, -1, mObj)
	if err != nil {
		return err
	}

	u := uint64(n.Size)
	logrus.Infof("cache rebuild action completed. %s of data rebuilt and stored", humanize.Bytes(u))

	return nil
}

// Configure prepares the rebuild fields for the action to be taken.
func (r *Rebuild) Configure(repo *Repo) error {
	logrus.Trace("configuring rebuild action")

	// construct the object path
	path := buildNamespace(repo, r.Prefix, r.Path, r.Filename)

	logrus.Debugf("created bucket path %s", path)

	// store it in the namespace
	r.Namespace = path

	return nil
}

// Validate verifies the Rebuild is properly configured.
func (r *Rebuild) Validate() error {
	logrus.Trace("validating rebuild action configuration")

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

	// verify mount is provided
	if len(r.Mount) == 0 {
		return fmt.Errorf("no mount provided")
	}

	// validate that the source exists
	for _, mount := range r.Mount {
		_, err := os.Lstat(mount)
		if err != nil {
			return fmt.Errorf("mount: %s, make sure file or directory exists", mount)
		}
	}

	return nil
}
