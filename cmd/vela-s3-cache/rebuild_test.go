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
		Root:     "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
		Mount:    []string{"testdata/hello.txt"},
	}

	err := r.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Rebuild_Validate_NoRoot(t *testing.T) {
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
		Root:    "bucket",
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
		Root:     "bucket",
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
		Root:     "bucket",
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
		Root:     "bucket",
		Prefix:   "foo/bar",
		Filename: "archive.tar",
		Mount:    []string{"testdata/bye.txt"},
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Rebuild_Configure(t *testing.T) {
	testCases := []struct {
		desc    string
		repo    *Repo
		rebuild *Rebuild
		want    string
	}{
		{
			desc:    "basic",
			repo:    &Repo{"foo", "bar", "", ""},
			rebuild: &Rebuild{},
			want:    "foo/bar",
		},
		{
			desc:    "prefix",
			repo:    &Repo{"foo", "bar", "", ""},
			rebuild: &Rebuild{Prefix: "prefix"},
			want:    "prefix/foo/bar",
		},
		{
			desc:    "path",
			repo:    &Repo{"foo", "bar", "", ""},
			rebuild: &Rebuild{Path: "path"},
			want:    "foo/bar/path",
		},
		{
			desc:    "prefix and path",
			repo:    &Repo{"foo", "bar", "", ""},
			rebuild: &Rebuild{Path: "path", Prefix: "prefix"},
			want:    "prefix/foo/bar/path",
		},
		{
			desc:    "all fail",
			repo:    &Repo{},
			rebuild: &Rebuild{},
			want:    "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.rebuild.Configure(tC.repo)
			if tC.want == "" && err == nil {
				t.Error("should error on empty Namespace")
			}

			if tC.rebuild.Namespace != tC.want {
				t.Errorf("test name: %s\nwant: %s, got: %s", tC.desc, tC.want, tC.rebuild.Namespace)
			}
		})
	}
}
