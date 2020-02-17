// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mholt/archiver"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
)

const restoreAction = "restore"

// Restore represents the plugin configuration for Restore information.
type Restore struct {
	// sets the name of the bucket
	Root string
	// sets the path for where to store the object
	Prefix string
	// sets the name of the cache object
	Filename string
	// sets the timeout on the call to s3
	Timeout time.Duration
}

// Exec formats and runs the actions for restoring a cache in s3.
func (r *Restore) Exec(mc *minio.Client) error {
	logrus.Trace("running restore with provided configuration")

	logrus.Debugf("getting object info on bucket %s from path: %s", r.Root, r.Prefix)
	ok, err := mc.StatObject(r.Root, r.Prefix, minio.StatObjectOptions{})
	if ok.Key == "" {
		logrus.Error(err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	logrus.Debugf("getting object in bucket %s from path: %s", r.Root, r.Prefix)
	reader, err := mc.GetObjectWithContext(ctx, r.Root, r.Prefix, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	logrus.Debugf("creating file %s on system", r.Filename)
	localFile, err := os.Create(r.Filename)
	if err != nil {
		return err
	}
	defer localFile.Close()

	logrus.Debugf("get object of file %s", r.Filename)
	stat, err := reader.Stat()
	if err != nil {
		return err
	}

	logrus.Debugf("copy object data to local file %s", r.Filename)
	if _, err := io.CopyN(localFile, reader, stat.Size); err != nil {
		return err
	}

	logrus.Debug("get current working directory")
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	logrus.Debugf("unarchiving file %s into directory %s", r.Filename, pwd)
	err = archiver.Unarchive(r.Filename, pwd)
	if err != nil {
		return err
	}

	return nil
}

// Configure prepares the restore fields for the action to be taken
func (r *Restore) Configure(repo Repo) error {

	path := fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(r.Prefix, "/"), repo.Owner, repo.Name, r.Filename)

	logrus.Debugf("created bucket path %s", path)

	r.Prefix = strings.TrimLeft(path, "/")

	return nil
}

// Validate verifies the Restore is properly configured.
func (r *Restore) Validate() error {
	logrus.Trace("validating restore action configuration")

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

	return nil
}
