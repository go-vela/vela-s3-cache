// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestS3Cache_Repo_Validate(t *testing.T) {
	// setup types
	r := &Repo{
		Owner:       "foo",
		Name:        "bar",
		Branch:      "main",
		BuildBranch: "dev",
	}

	err := r.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Repo_Validate_NoOwner(t *testing.T) {
	// setup types
	r := &Repo{
		Owner:       "",
		Name:        "bar",
		Branch:      "main",
		BuildBranch: "dev",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Repo_Validate_NoName(t *testing.T) {
	// setup types
	r := &Repo{
		Owner:       "foo",
		Name:        "",
		Branch:      "main",
		BuildBranch: "dev",
	}

	err := r.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}
