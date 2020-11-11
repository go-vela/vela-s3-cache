// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"testing"
	"time"
)

func TestS3Cache_Rebuild_Validate(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Rebuild{
		Timeout:  timeout,
		Bucket:   "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
		Mount:    []string{"testdata/hello.txt"},
	}

	err := r.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Rebuild_Validate_NoBucket(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Rebuild{
		Timeout:  timeout,
		Prefix:   "foo/bar",
		Filename: "archive.tar",
		Mount:    []string{"testdata/hello.txt"},
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Rebuild_Validate_NoFilename(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Rebuild{
		Timeout: timeout,
		Bucket:  "bucket",
		Prefix:  "foo/bar",
		Mount:   []string{"testdata/hello.txt"},
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Rebuild_Validate_NoTimeout(t *testing.T) {
	// setup types
	r := &Rebuild{
		Bucket:   "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
		Mount:    []string{"testdata/hello.txt"},
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Rebuild_Validate_NoMount(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Rebuild{
		Timeout:  timeout,
		Bucket:   "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Rebuild_Validate_MissingMount(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Rebuild{
		Timeout:  timeout,
		Bucket:   "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
		Mount:    []string{"testdata/bye.txt"},
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}
