// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
	"time"
)

func TestS3Cache_Restore_Validate(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Restore{
		Timeout:  timeout,
		Bucket:   "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
	}

	err := r.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Restore_Validate_NoBucket(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Restore{
		Timeout:  timeout,
		Prefix:   "foo/bar",
		Filename: "archive.tar",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Restore_Validate_NoFilename(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	r := &Restore{
		Timeout: timeout,
		Bucket:  "bucket",
		Prefix:  "foo/bar",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Restore_Validate_NoTimeout(t *testing.T) {
	// setup types
	r := &Restore{
		Bucket:   "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}
