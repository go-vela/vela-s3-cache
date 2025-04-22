// SPDX-License-Identifier: Apache-2.0

package main

import (
	"compress/flate"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"

	"github.com/go-vela/vela-s3-cache/pkg/archiver"
)

const rebuildAction = "rebuild"

// Rebuild represents the plugin configuration for rebuild information.
type Rebuild struct {
	// sets the name of the bucket
	Bucket string
	// set the compression level for the archive
	CompressionLevel int
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

	// use OS's tmp dir for archive creation
	dir := os.TempDir()

	// make sure the target directory exists
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("unable to create target directory %q for archive: %w", dir, err)
		}
	}

	p := filepath.Join(dir, r.Filename)

	logrus.Debugf("determined temporary file path as %s", p)

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("failed to create temporary file %s for cache archive: %w", p, err)
	}
	defer os.Remove(f.Name())

	logrus.Debugf("created temporary file %s", f.Name())

	// forcing format until we support more formats
	a, err := archiver.NewArchiver("tar.gz",
		archiver.WithCompressionLevel(r.CompressionLevel),
		archiver.WithPreservePath(r.PreservePath),
	)
	if err != nil {
		return err
	}

	// archive the objects in the mount paths provided
	err = a.Archive(context.Background(), r.Mount, f)
	if err != nil {
		return err
	}

	logrus.Debugf("archiving artifact in path %s complete", f.Name())

	stat, err := os.Stat(f.Name())
	if err != nil {
		return err
	}

	//nolint:gosec // G115: integer overflow conversion should be handled via max()
	logrus.Infof("archive %s created with size %s", f.Name(), humanize.Bytes(uint64(max(0, stat.Size()))))

	// set a timeout on the request to the cache provider
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	logrus.Debugf("putting archive %s in bucket %s in path: %s", f.Name(), r.Bucket, r.Namespace)

	// create an options object for the upload
	mObj := minio.PutObjectOptions{
		ContentType: "application/gzip", // gzip is the closest for tar.gz https://www.iana.org/assignments/media-types/media-types.xhtml
	}

	n, err := mc.FPutObject(ctx, r.Bucket, r.Namespace, f.Name(), mObj)
	if err != nil {
		return fmt.Errorf("failed to upload cache archive to bucket %s at path %s: %w", r.Bucket, r.Namespace, err)
	}

	//nolint:gosec // G115: integer overflow conversion should be handled via max()
	logrus.Infof("cache rebuild action completed. %s of data rebuilt and stored", humanize.Bytes(uint64(max(0, n.Size))))

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
			return fmt.Errorf("mount not found: %s, make sure file or directory exists", mount)
		}
	}

	// verify valid compression level is provided
	// using compress/flate levels as a guide for compression level range
	if r.CompressionLevel < flate.DefaultCompression || r.CompressionLevel > flate.BestCompression {
		return fmt.Errorf("compression level must be between -1 and 9")
	}

	return nil
}
