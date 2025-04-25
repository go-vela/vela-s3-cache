// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/klauspost/compress/gzip"
)

func TestTarGzipArchiver(t *testing.T) {
	suite := &ArchiverTestSuite{
		TestArchiver: &TarGzipArchiver{
			CompressionLevel: gzip.DefaultCompression,
		},
	}

	// run generic test suite
	suite.RunTests(t)

	// run tests specific to this archiver
	t.Run("CompressionLevel", testTarGzipArchiverCompressionLevel)
	t.Run("CompressionLevelEffectiveness", testCompressionLevelEffectiveness)
	t.Run("UnsupportedHeaderType", testUnsupportedHeaderType)
	t.Run("PathTraversalPrevention", testPathTraversalPrevention)
	t.Run("PreservePath", testPreservePath)
	t.Run("HardLinks", testHardLinks)
}

// testTarGzipArchiverCompressionLevel tests the compression level functionality
// which is specific to TarGzipArchiver.
func testTarGzipArchiverCompressionLevel(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		want        Archiver
		compression int
		wantErr     bool
	}{
		{
			name:        "tar.gz",
			format:      "tar.gz",
			compression: -1,
			want:        &TarGzipArchiver{CompressionLevel: -1},
		},
		{
			name:        "compression level 3",
			format:      "tar.gz",
			compression: 3,
			want:        &TarGzipArchiver{CompressionLevel: 3},
		},
		{
			name:    "unsupported format",
			format:  "foo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArchiver(tt.format, WithCompressionLevel(tt.compression))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewArchiver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(TarGzipArchiver{})); diff != "" {
				t.Errorf("NewArchiver() mismatch (-want +got): %s", diff)
			}
		})
	}
}

// testCompressionLevelEffectiveness tests the effectiveness of different compression levels.
func testCompressionLevelEffectiveness(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a file with highly compressible content (repeated text)
	testFile := filepath.Join(tmpDir, "compressible.txt")

	// create 100KB of highly compressible content (repeated pattern)
	var contentBuilder strings.Builder
	repeatedText := "This is a test string that will be repeated many times to create a file that compresses well. "
	for range 1000 {
		contentBuilder.WriteString(repeatedText)
	}
	compressibleContent := contentBuilder.String()

	if err := os.WriteFile(testFile, []byte(compressibleContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// test with different compression levels
	compressionLevels := []struct {
		level       int
		description string
	}{
		{gzip.NoCompression, "no compression"},
		{gzip.BestSpeed, "best speed"},
		{gzip.BestCompression, "best compression"},
	}

	var archiveSizes []int

	for _, cl := range compressionLevels {
		// create an archiver with the specific compression level
		archiver := &TarGzipArchiver{
			CompressionLevel: cl.level,
		}

		// archive the file
		var buf bytes.Buffer
		if err := archiver.Archive(ctx, []string{testFile}, &buf); err != nil {
			t.Fatalf("Archive() with %s error = %v", cl.description, err)
		}

		// record the archive size
		size := buf.Len()
		archiveSizes = append(archiveSizes, size)

		t.Logf("Archive size with %s (level %d): %d bytes", cl.description, cl.level, size)

		// verify the archive can be extracted correctly
		destDir := filepath.Join(tmpDir, fmt.Sprintf("dest_%d", cl.level))
		if err := os.MkdirAll(destDir, 0755); err != nil {
			t.Fatalf("failed to create destination directory: %v", err)
		}

		if err := archiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir); err != nil {
			t.Fatalf("Unarchive() with %s error = %v", cl.description, err)
		}

		// verify the extracted file has the correct content
		extractedFile := filepath.Join(destDir, "compressible.txt")
		extractedContent, err := os.ReadFile(extractedFile)
		if err != nil {
			t.Fatalf("failed to read extracted file: %v", err)
		}

		if string(extractedContent) != compressibleContent {
			t.Errorf("extracted content with %s does not match original", cl.description)
		}
	}

	// verify that higher compression levels produce smaller archives
	if len(archiveSizes) >= 3 {
		// noCompression should be larger than BestSpeed
		if archiveSizes[0] <= archiveSizes[1] {
			t.Errorf("expected no compression (%d bytes) to be larger than best speed (%d bytes)",
				archiveSizes[0], archiveSizes[1])
		}

		// bestSpeed should be larger than BestCompression
		if archiveSizes[1] <= archiveSizes[2] {
			t.Errorf("expected best speed (%d bytes) to be larger than best compression (%d bytes)",
				archiveSizes[1], archiveSizes[2])
		}
	}
}

func testUnsupportedHeaderType(t *testing.T) {
	ctx := t.Context()
	archiver := &TarGzipArchiver{}
	tmpDir := t.TempDir()

	// create a tar.gz archive with an unsupported header type
	var unsupportedBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&unsupportedBuf)
	tarWriter := tar.NewWriter(gzipWriter)

	// add a normal file first to ensure the archive is valid
	normalHeader := &tar.Header{
		Name:     "normal.txt",
		Mode:     0644,
		Size:     int64(len("normal content")),
		Typeflag: tar.TypeReg,
	}
	if err := tarWriter.WriteHeader(normalHeader); err != nil {
		t.Fatalf("failed to write normal header: %v", err)
	}
	if _, err := tarWriter.Write([]byte("normal content")); err != nil {
		t.Fatalf("failed to write normal content: %v", err)
	}

	// add a header with an unsupported type
	// use a custom type value that's not one of the standard types
	unsupportedHeader := &tar.Header{
		Name:     "unsupported",
		Mode:     0644,
		Typeflag: 'Z', // custom type that's not supported
	}
	if err := tarWriter.WriteHeader(unsupportedHeader); err != nil {
		t.Fatalf("failed to write unsupported header: %v", err)
	}

	tarWriter.Close()
	gzipWriter.Close()

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	// try to unarchive the archive with unsupported header type
	err := archiver.Unarchive(ctx, bytes.NewReader(unsupportedBuf.Bytes()), destDir)

	// the current implementation returns an error for unsupported header types
	if err == nil {
		t.Errorf("Unarchive() with unsupported header type should fail")
	}

	// check that the error message mentions the unsupported header type
	if !strings.Contains(err.Error(), "unsupported tar header type") {
		t.Errorf("Error message should mention unsupported header type, got: %v", err)
	}

	// even though there's an error, the normal file should still be extracted
	// since the error occurs after processing the normal file
	normalPath := filepath.Join(destDir, "normal.txt")
	if _, err := os.Stat(normalPath); os.IsNotExist(err) {
		t.Logf("Note: Normal file was not extracted due to the error")
	} else {
		content, err := os.ReadFile(normalPath)
		if err != nil {
			t.Fatalf("failed to read extracted normal file: %v", err)
		}
		if string(content) != "normal content" {
			t.Errorf("normal file content = %q, want %q", string(content), "normal content")
		}
	}
}

func testPathTraversalPrevention(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a base destination directory so each test can run in isolation
	// and avoid interference between tests.
	baseDestDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(baseDestDir, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	// create a directory outside the destination to verify files don't get extracted there
	outsideDir := filepath.Join(tmpDir, "outside")
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("failed to create outside directory: %v", err)
	}

	// create a file outside the destination for hard link testing
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside content"), 0600); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	// define test cases
	tests := []struct {
		name           string
		headerName     string
		headerType     byte // tar.TypeReg, tar.TypeSymlink, etc.
		linkname       string
		content        string
		expectError    bool
		errorSubstring string // expected substring in error message
	}{
		{
			name:        "normal file",
			headerName:  "normal.txt",
			headerType:  tar.TypeReg,
			content:     "normal content",
			expectError: false,
		},
		{
			name:           "path traversal with relative path",
			headerName:     "../outside/evil.txt",
			headerType:     tar.TypeReg,
			content:        "evil content",
			expectError:    true,
			errorSubstring: "path traversal",
		},
		{
			name:           "path traversal with absolute path",
			headerName:     "/etc/passwd",
			headerType:     tar.TypeReg,
			content:        "fake passwd",
			expectError:    true,
			errorSubstring: "absolute paths",
		},
		{
			name:           "path traversal with symlink",
			headerName:     "symlink",
			headerType:     tar.TypeSymlink,
			linkname:       "../outside",
			expectError:    true,
			errorSubstring: "symlink target",
		},
		{
			// this tests the isPathWithinBoundary check in processSymlink
			name:           "path traversal with deceptive symlink",
			headerName:     "deceptive_symlink",
			headerType:     tar.TypeSymlink,
			linkname:       "subdir/../../outside", // appears to be within dest but resolves outside
			expectError:    true,
			errorSubstring: "symlink target path traversal",
		},
		{
			// this tests the isPathWithinBoundary check in processHardLink
			name:           "path traversal with hard link",
			headerName:     "hardlink",
			headerType:     tar.TypeLink,
			linkname:       "../outside/outside.txt", // points to a file outside dest
			expectError:    true,
			errorSubstring: "hard link target path traversal",
		},
		{
			// another test for the isPathWithinBoundary check in processHardLink
			name:           "path traversal with deceptive hard link",
			headerName:     "deceptive_hardlink",
			headerType:     tar.TypeLink,
			linkname:       "subdir/../../outside/outside.txt", // appears to be within dest but resolves outside
			expectError:    true,
			errorSubstring: "hard link target path traversal",
		},
		{
			// this tests the check for absolute symlinks in processSymlink
			name:           "absolute symlink path",
			headerName:     "absolute_symlink",
			headerType:     tar.TypeSymlink,
			linkname:       "/etc/passwd", // absolute path
			expectError:    true,
			errorSubstring: "absolute symlinks are not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create a unique destination directory for each test case
			destDir := filepath.Join(baseDestDir, strings.ReplaceAll(tt.name, " ", "_"))
			if err := os.MkdirAll(destDir, 0755); err != nil {
				t.Fatalf("failed to create destination directory: %v", err)
			}

			archiver := &TarGzipArchiver{}

			// create a tar.gz archive for this test case
			var buf bytes.Buffer
			gzipWriter := gzip.NewWriter(&buf)
			tarWriter := tar.NewWriter(gzipWriter)

			// for hard link tests, we need to create a valid file first
			if tt.headerType == tar.TypeLink {
				// add a normal file that could be a valid hard link target
				normalHeader := &tar.Header{
					Name:     "normal.txt",
					Mode:     0644,
					Size:     int64(len("normal content")),
					Typeflag: tar.TypeReg,
				}
				if err := tarWriter.WriteHeader(normalHeader); err != nil {
					t.Fatalf("failed to write normal header: %v", err)
				}
				if _, err := tarWriter.Write([]byte("normal content")); err != nil {
					t.Fatalf("failed to write normal content: %v", err)
				}
			}

			// add the test case header
			testHeader := &tar.Header{
				Name:     tt.headerName,
				Mode:     0644,
				Typeflag: tt.headerType,
			}

			switch tt.headerType {
			case tar.TypeReg:
				testHeader.Size = int64(len(tt.content))
			case tar.TypeSymlink, tar.TypeLink:
				testHeader.Linkname = tt.linkname
			}

			if err := tarWriter.WriteHeader(testHeader); err != nil {
				t.Fatalf("failed to write test header: %v", err)
			}

			if tt.headerType == tar.TypeReg {
				if _, err := tarWriter.Write([]byte(tt.content)); err != nil {
					t.Fatalf("failed to write test content: %v", err)
				}
			}

			tarWriter.Close()
			gzipWriter.Close()

			// try to unarchive the archive
			err := archiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir)

			// check if we got an error as expected
			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error for %s, but got none", tt.name)
					return
				}

				// if no specific error substring is expected, any error is fine
				if tt.errorSubstring == "" {
					t.Logf("Got expected error: %v", err)
					return
				}

				// for hard links, accept either path traversal or file conflict errors
				if tt.headerType == tar.TypeLink &&
					(strings.Contains(err.Error(), tt.errorSubstring) ||
						strings.Contains(err.Error(), "file conflict detected")) {
					t.Logf("Got acceptable error for hard link: %v", err)
					return
				}

				// for other cases, verify the specific error substring
				if !strings.Contains(err.Error(), tt.errorSubstring) {
					t.Errorf("expected error to contain %q, but got: %v", tt.errorSubstring, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}

				return
			}

			// no error was expected
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// for normal files, verify they were extracted correctly
			if tt.headerType == tar.TypeReg {
				extractedPath := filepath.Join(destDir, tt.headerName)
				content, err := os.ReadFile(extractedPath)
				if err != nil {
					t.Errorf("failed to read extracted file: %v", err)
					return
				}

				if string(content) != tt.content {
					t.Errorf("extracted content = %q, want %q", string(content), tt.content)
				}
			}

			// for path traversal attempts, verify the file doesn't exist outside the destination
			if strings.Contains(tt.name, "path traversal") {
				// check that no files were created in the outside directory
				entries, err := os.ReadDir(outsideDir)
				if err != nil {
					t.Logf("Error reading outside directory: %v", err)
				} else {
					// we should only have the outside.txt file we created
					if len(entries) > 1 {
						t.Errorf("Found unexpected files in outside directory: %v", entries)
					}
				}
			}
		})
	}
}

func testPreservePath(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a nested directory structure
	nestedDir := filepath.Join(tmpDir, "level1", "level2")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	// create a file in the nested directory
	testFile := filepath.Join(nestedDir, "test.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// define test cases
	tests := []struct {
		name         string
		preservePath bool
		source       string
		isDir        bool
		checkFunc    func(string, string) bool // function to check if path is as expected (takes destDir as second param)
	}{
		{
			name:         "directory with preserve path true",
			preservePath: true,
			source:       filepath.Join(tmpDir, "level1"),
			isDir:        true,
			checkFunc: func(path, destDir string) bool {
				// with preserve path true for directory, we expect the full path structure
				relPath, err := filepath.Rel(destDir, path)
				if err != nil {
					return false
				}
				// should include level1/level2/test.txt
				return relPath == filepath.Join("level1", "level2", "test.txt")
			},
		},
		{
			name:         "directory with preserve path false",
			preservePath: false,
			source:       filepath.Join(tmpDir, "level1"),
			isDir:        true,
			checkFunc: func(path, destDir string) bool {
				// with preserve path false for directory, we still expect the directory name to be preserved
				relPath, err := filepath.Rel(destDir, path)
				if err != nil {
					return false
				}
				// should still include level1/level2/test.txt because we preserve directory names
				return relPath == filepath.Join("level1", "level2", "test.txt")
			},
		},
		{
			name:         "individual file with preserve path true",
			preservePath: true,
			source:       testFile,
			isDir:        false,
			checkFunc: func(path, _ string) bool {
				// with preserve path true for file, we expect the full path
				return strings.Contains(path, "level1") &&
					strings.Contains(path, "level2") &&
					strings.HasSuffix(path, "test.txt")
			},
		},
		{
			name:         "individual file with preserve path false",
			preservePath: false,
			source:       testFile,
			isDir:        false,
			checkFunc: func(path, destDir string) bool {
				// with preserve path false for file, we expect just the filename
				relPath, err := filepath.Rel(destDir, path)
				if err != nil {
					return false
				}
				// should be just test.txt without any directories
				return relPath == "test.txt"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create an archiver with the specific preserve path setting
			archiver := &TarGzipArchiver{
				PreservePath:     tt.preservePath,
				CompressionLevel: gzip.DefaultCompression,
			}

			// create destination directory
			destDir := filepath.Join(tmpDir, "dest_"+strings.ReplaceAll(tt.name, " ", "_"))
			if err := os.MkdirAll(destDir, 0755); err != nil {
				t.Fatalf("failed to create destination directory: %v", err)
			}

			// archive and unarchive
			var buf bytes.Buffer
			if err := archiver.Archive(ctx, []string{tt.source}, &buf); err != nil {
				t.Fatalf("Archive() error = %v", err)
			}

			if err := archiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir); err != nil {
				t.Fatalf("Unarchive() error = %v", err)
			}

			// find all files in the destination directory
			var foundFiles []string
			err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					foundFiles = append(foundFiles, path)
					t.Logf("Found file: %s", path)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("Error walking directory: %v", err)
			}

			// find the test.txt file in the output
			var testFilePath string
			for _, file := range foundFiles {
				if strings.HasSuffix(file, "test.txt") {
					testFilePath = file
					break
				}
			}

			if testFilePath == "" {
				t.Fatalf("could not find test.txt in the output files")
			}

			// verify the content
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			if string(content) != testContent {
				t.Errorf("file content = %q, want %q", string(content), testContent)
			}

			// for debugging, print the relative path from destDir
			relPath, err := filepath.Rel(destDir, testFilePath)
			if err != nil {
				t.Logf("Could not get relative path: %v", err)
			} else {
				t.Logf("Relative path: %s", relPath)
			}

			// verify the path structure using the custom check function
			if !tt.checkFunc(testFilePath, destDir) {
				t.Errorf("path %q does not match expected structure for %s",
					testFilePath, tt.name)

				// additional debug info
				relPath, _ := filepath.Rel(destDir, testFilePath)
				t.Logf("Relative path from destDir: %s", relPath)
			}
		})
	}
}

func testHardLinks(t *testing.T) {
	// skip on Windows as hard links work differently
	if runtime.GOOS == "windows" {
		t.Skip("skipping hard links test on Windows")
	}

	ctx := t.Context()
	archiver := &TarGzipArchiver{}
	tmpDir := t.TempDir()

	// create a source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	sourceContent := "source content"
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0600); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// create a hard link to the source file
	hardLinkFile := filepath.Join(tmpDir, "hardlink.txt")
	if err := os.Link(sourceFile, hardLinkFile); err != nil {
		t.Fatalf("failed to create hard link: %v", err)
	}

	// verify the hard link was created correctly
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		t.Fatalf("failed to stat source file: %v", err)
	}

	hardLinkInfo, err := os.Stat(hardLinkFile)
	if err != nil {
		t.Fatalf("failed to stat hard link file: %v", err)
	}

	// on Unix-like systems, files with the same inode are hard links to each other
	if os.SameFile(sourceInfo, hardLinkInfo) == false {
		t.Fatalf("expected files to be hard links to each other")
	}

	// create a tar.gz archive manually to ensure we have hard links
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	// add the source file
	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("failed to read source file: %v", err)
	}

	sourceHeader, err := tar.FileInfoHeader(sourceInfo, "")
	if err != nil {
		t.Fatalf("failed to create source header: %v", err)
	}
	sourceHeader.Name = "source.txt"

	if err := tarWriter.WriteHeader(sourceHeader); err != nil {
		t.Fatalf("failed to write source header: %v", err)
	}

	if _, err := tarWriter.Write(sourceData); err != nil {
		t.Fatalf("failed to write source data: %v", err)
	}

	// add the hard link
	hardLinkHeader, err := tar.FileInfoHeader(hardLinkInfo, "")
	if err != nil {
		t.Fatalf("failed to create hard link header: %v", err)
	}
	hardLinkHeader.Name = "hardlink.txt"
	hardLinkHeader.Typeflag = tar.TypeLink
	hardLinkHeader.Linkname = "source.txt"

	if err := tarWriter.WriteHeader(hardLinkHeader); err != nil {
		t.Fatalf("failed to write hard link header: %v", err)
	}

	// close the writers
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	// unarchive using our archiver
	if err := archiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir); err != nil {
		t.Fatalf("Unarchive() error = %v", err)
	}

	// get the paths for verification
	extractedSourcePath := filepath.Join(destDir, "source.txt")
	extractedHardLinkPath := filepath.Join(destDir, "hardlink.txt")

	// verify the extracted files are hard links to each other
	extractedSourceInfo, err := os.Stat(extractedSourcePath)
	if err != nil {
		t.Fatalf("failed to stat extracted source file: %v", err)
	}

	extractedHardLinkInfo, err := os.Stat(extractedHardLinkPath)
	if err != nil {
		t.Fatalf("failed to stat extracted hard link file: %v", err)
	}

	if os.SameFile(extractedSourceInfo, extractedHardLinkInfo) == false {
		t.Errorf("extracted files are not hard links to each other")
	}

	// verify the content of the hard link
	extractedHardLinkContent, err := os.ReadFile(extractedHardLinkPath)
	if err != nil {
		t.Fatalf("failed to read extracted hard link: %v", err)
	}

	if string(extractedHardLinkContent) != sourceContent {
		t.Errorf("hard link content = %q, want %q", string(extractedHardLinkContent), sourceContent)
	}

	// modify the source file and verify the hard link is also modified
	newContent := "modified content"
	if err := os.WriteFile(extractedSourcePath, []byte(newContent), 0600); err != nil {
		t.Fatalf("failed to modify extracted source file: %v", err)
	}

	// read the hard link again
	modifiedHardLinkContent, err := os.ReadFile(extractedHardLinkPath)
	if err != nil {
		t.Fatalf("failed to read modified hard link: %v", err)
	}

	if string(modifiedHardLinkContent) != newContent {
		t.Errorf("modified hard link content = %q, want %q", string(modifiedHardLinkContent), newContent)
	}
}
