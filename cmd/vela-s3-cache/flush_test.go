// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"testing"
)

func TestS3Cache_Flush_Validate(t *testing.T) {
	// setup types
	f := &Flush{
		Root: "bucket",
	}

	err := f.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Flush_Validate_NoRoot(t *testing.T) {
	// setup types
	f := &Flush{}

	err := f.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Flush_Configure(t *testing.T) {
	testCases := []struct {
		desc  string
		repo  *Repo
		flush *Flush
		want  string
	}{
		{
			desc:  "basic",
			repo:  &Repo{"foo", "bar", "", ""},
			flush: &Flush{},
			want:  "foo/bar",
		},
		{
			desc:  "prefix",
			repo:  &Repo{"foo", "bar", "", ""},
			flush: &Flush{Prefix: "prefix"},
			want:  "prefix/foo/bar",
		},
		{
			desc:  "path",
			repo:  &Repo{"foo", "bar", "", ""},
			flush: &Flush{Path: "path"},
			want:  "foo/bar/path",
		},
		{
			desc:  "prefix and path",
			repo:  &Repo{"foo", "bar", "", ""},
			flush: &Flush{Path: "path", Prefix: "prefix"},
			want:  "prefix/foo/bar/path",
		},
		{
			desc:  "all fail",
			repo:  &Repo{},
			flush: &Flush{},
			want:  "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.flush.Configure(tC.repo)
			if tC.want == "" && err == nil {
				t.Error("should error on empty Namespace")
			}

			if tC.flush.Namespace != tC.want {
				t.Errorf("test name: %s\nwant: %s, got: %s", tC.desc, tC.want, tC.flush.Namespace)
			}
		})
	}
}
