// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"testing"
	"time"
)

func TestS3Cache_Plugin_Validate(t *testing.T) {
	// setup types
	timeout, _ := time.ParseDuration("10m")

	p := &Plugin{
		Config: &Config{
			Action:    "flush",
			AccessKey: "123456",
			SecretKey: "654321",
			Server:    "https://server",
		},
		Repo: &Repo{
			Owner:       "foo",
			Name:        "bar",
			Branch:      "main",
			BuildBranch: "main",
		},
		Flush: &Flush{
			Root: "bucket",
		},
		Rebuild: &Rebuild{
			Timeout:  timeout,
			Root:     "bucket",
			Filename: "archive.tar",
			Mount:    []string{"/path/to/cache"},
		},
		Restore: &Restore{
			Timeout:  timeout,
			Root:     "bucket",
			Filename: "archive.tar",
		},
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Plugin_buildNamespace(t *testing.T) {
	testCases := []struct {
		desc     string
		repo     *Repo
		prefix   string
		path     string
		filename string
		want     string
	}{
		{
			desc:     "basic",
			repo:     &Repo{"foo", "bar", "", ""},
			prefix:   "",
			path:     "",
			filename: "",
			want:     "foo/bar",
		},
		{
			desc:     "prefix",
			repo:     &Repo{"foo", "bar", "", ""},
			prefix:   "prefix",
			path:     "",
			filename: "",
			want:     "prefix/foo/bar",
		},
		{
			desc:     "path",
			repo:     &Repo{"foo", "bar", "", ""},
			prefix:   "",
			path:     "custom/path",
			filename: "",
			want:     "custom/path",
		},
		{
			desc:     "prefix and path - use path",
			repo:     &Repo{"foo", "bar", "", ""},
			prefix:   "prefix",
			path:     "custom/path",
			filename: "",
			want:     "custom/path",
		},
		{
			desc:     "path w/ filename",
			repo:     &Repo{"foo", "bar", "", ""},
			prefix:   "",
			path:     "custom/path",
			filename: "archive.tgz",
			want:     "custom/path/archive.tgz",
		},
		{
			desc:     "all fail",
			repo:     &Repo{},
			prefix:   "",
			path:     "",
			filename: "",
			want:     ".",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			path := buildNamespace(tC.repo, tC.prefix, tC.path, tC.filename)

			if path != tC.want {
				t.Errorf("test name: %s\nwant: %s, got: %s", tC.desc, tC.want, path)
			}
		})
	}
}
