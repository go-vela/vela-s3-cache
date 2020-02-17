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
