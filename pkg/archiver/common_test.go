// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestIsPathWithinBoundary(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		dir      string
		expected bool
	}{
		{
			name:     "exact same path",
			path:     "/path/to/dir",
			dir:      "/path/to/dir",
			expected: true,
		},
		{
			name:     "child path",
			path:     "/path/to/dir/child",
			dir:      "/path/to/dir",
			expected: true,
		},
		{
			name:     "parent path",
			path:     "/path/to",
			dir:      "/path/to/dir",
			expected: false,
		},
		{
			name:     "sibling path",
			path:     "/path/to/other",
			dir:      "/path/to/dir",
			expected: false,
		},
		{
			name:     "with dot segments",
			path:     "/path/to/dir/./child",
			dir:      "/path/to/dir",
			expected: true,
		},
		{
			name:     "with double dot segments",
			path:     "/path/to/dir/../dir/child",
			dir:      "/path/to/dir",
			expected: true,
		},
		{
			name:     "with double dot segments leading to parent",
			path:     "/path/to/../dir/child",
			dir:      "/path/to/dir",
			expected: false,
		},
		{
			name:     "different roots",
			path:     "/other/path",
			dir:      "/path/to/dir",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			dir:      "/path/to/dir",
			expected: false,
		},
		{
			name:     "empty directory",
			path:     "/path/to/dir",
			dir:      "",
			expected: false,
		},
		{
			name:     "relative paths",
			path:     "dir/child",
			dir:      "dir",
			expected: true,
		},
		{
			name:     "mixed relative and absolute paths",
			path:     "dir/child",
			dir:      "/dir",
			expected: false,
		},
		{
			name:     "with trailing separators",
			path:     "/path/to/dir/child/",
			dir:      "/path/to/dir/",
			expected: true,
		},
		{
			name:     "similar prefix but not within",
			path:     "/path/to/directory",
			dir:      "/path/to/dir",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPathWithinBoundary(tt.path, tt.dir)
			if result != tt.expected {
				t.Errorf("isPathWithinBoundary(%q, %q) = %v, want %v",
					tt.path, tt.dir, result, tt.expected)
			}
		})
	}
}

func TestFilterRedundantPaths(t *testing.T) {
	// create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// create nested directories and files
	nestedDir := filepath.Join(tmpDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("file1"), 0600); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	file2 := filepath.Join(nestedDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("file2"), 0600); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	tests := []struct {
		name     string
		paths    []string
		expected []string
	}{
		{
			name:     "no redundant paths",
			paths:    []string{file1, file2},
			expected: []string{file1, file2},
		},
		{
			name:     "directory and file inside",
			paths:    []string{tmpDir, file1},
			expected: []string{tmpDir},
		},
		{
			name:     "nested directories",
			paths:    []string{tmpDir, nestedDir},
			expected: []string{tmpDir},
		},
		{
			name:     "mixed order",
			paths:    []string{file1, tmpDir, file2},
			expected: []string{tmpDir},
		},
		{
			name:     "single path",
			paths:    []string{tmpDir},
			expected: []string{tmpDir},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := filterRedundantPaths(tt.paths)
			if err != nil {
				t.Fatalf("filterRedundantPaths() error = %v", err)
			}

			// sort both slices for comparison
			sort.Strings(filtered)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(filtered, tt.expected) {
				t.Errorf("filterRedundantPaths() = %v, want %v", filtered, tt.expected)
			}
		})
	}
}
