// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
)

func TestS3Cache_Config_New(t *testing.T) {
	//TODO: write this test
}

func TestS3Cache_Config_Validate(t *testing.T) {
	// setup types
	c := &Config{
		Action:    "flush",
		AccessKey: "123456",
		SecretKey: "654321",
		Server:    "https://server",
	}

	err := c.Validate()
	if err != nil {
		t.Errorf("Validate returned err: %v", err)
	}
}

func TestS3Cache_Config_Validate_NoServer(t *testing.T) {
	// setup types
	c := &Config{
		Action:    "flush",
		AccessKey: "123456",
		SecretKey: "654321",
	}

	err := c.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Config_Validate_NoAction(t *testing.T) {
	// setup types
	c := &Config{
		AccessKey: "123456",
		SecretKey: "654321",
		Server:    "https://server",
	}

	err := c.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Config_Validate_NoAccessKey(t *testing.T) {
	// setup types
	c := &Config{
		Action:    "flush",
		SecretKey: "654321",
		Server:    "https://server",
	}

	err := c.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}

func TestS3Cache_Config_Validate_NoSecretKey(t *testing.T) {
	// setup types
	c := &Config{
		Action:    "flush",
		AccessKey: "123456",
		Server:    "https://server",
	}

	err := c.Validate()
	if err == nil {
		t.Errorf("Validate should have returned err")
	}
}
