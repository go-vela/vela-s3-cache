// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

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
		Root:     "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
	}

	err := r.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Restore_Validate_NoRoot(t *testing.T) {
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
		Root:    "bucket",
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
		Root:     "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Restore_Configure(t *testing.T) {
	testCases := []struct {
		desc    string
		repo    *Repo
		restore *Restore
		want    string
	}{
		{
			desc:    "basic",
			repo:    &Repo{"foo", "bar", "", ""},
			restore: &Restore{},
			want:    "foo/bar",
		},
		{
			desc:    "prefix",
			repo:    &Repo{"foo", "bar", "", ""},
			restore: &Restore{Prefix: "prefix"},
			want:    "prefix/foo/bar",
		},
		{
			desc:    "path",
			repo:    &Repo{"foo", "bar", "", ""},
			restore: &Restore{Path: "path"},
			want:    "foo/bar/path",
		},
		{
			desc:    "prefix and path",
			repo:    &Repo{"foo", "bar", "", ""},
			restore: &Restore{Path: "path", Prefix: "prefix"},
			want:    "prefix/foo/bar/path",
		},
		{
			desc:    "all fail",
			repo:    &Repo{},
			restore: &Restore{},
			want:    "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.restore.Configure(tC.repo)
			if tC.want == "" && err == nil {
				t.Error("should error on empty Namespace")
			}

			if tC.restore.Namespace != tC.want {
				t.Errorf("test name: %s\nwant: %s, got: %s", tC.desc, tC.want, tC.restore.Namespace)
			}
		})
	}
}
