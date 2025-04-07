// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// ArchiverTestSuite defines a set of tests that any Archiver implementation should pass.
type ArchiverTestSuite struct {
	// archiver instance
	TestArchiver Archiver
}

// RunTests runs all the tests in the suite against the archiver.
func (s *ArchiverTestSuite) RunTests(t *testing.T) {
	t.Run("Suite/BasicArchiveUnarchive", s.testBasicArchiveUnarchive)
	t.Run("Suite/ArchiveMultipleFiles", s.testArchiveMultipleFiles)
	t.Run("Suite/ArchiveWithSymlinks", s.testArchiveWithSymlinks)
	t.Run("Suite/PathTraversalPrevention", s.testPathTraversalPrevention)
	t.Run("Suite/ContextCancellation", s.testContextCancellation)
	t.Run("Suite/UnarchiveDirectories", s.testUnarchiveDirectories)
	t.Run("Suite/FilePermissions", s.testFilePermissions)
	t.Run("Suite/LargeFile", s.testLargeFile)
	t.Run("Suite/ErrorHandling", s.testErrorHandling)
	t.Run("Suite/ModificationTimePreservation", s.testModificationTimePreservation)
	t.Run("Suite/EmptyDirectories", s.testEmptyDirectories)
	t.Run("Suite/SymlinkChainAttack", s.testSymlinkChainAttack)
	t.Run("Suite/CircularSymlink", s.testCircularSymlink)
}

// testBasicArchiveUnarchive tests basic archive and unarchive functionality.
func (s *ArchiverTestSuite) testBasicArchiveUnarchive(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "test content"

	createTestFile(t, testFile, testContent, 0600)

	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{testFile}, destDir)

	// verify the result
	verifyFileContent(t, filepath.Join(destDir, "test.txt"), testContent)
}

// testArchiveMultipleFiles tests archiving and unarchiving multiple files.
func (s *ArchiverTestSuite) testArchiveMultipleFiles(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create files and directories
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testContent1 := "test content 1"
	createTestFile(t, testFile1, testContent1, 0600)

	testFile2 := filepath.Join(tmpDir, "test2.txt")
	testContent2 := "test content 2"
	createTestFile(t, testFile2, testContent2, 0600)

	subDir := filepath.Join(tmpDir, "subdir")
	createTestDir(t, subDir, 0755)

	testFile3 := filepath.Join(subDir, "test3.txt")
	testContent3 := "test content 3"
	createTestFile(t, testFile3, testContent3, 0600)

	// create destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{testFile1, testFile2, subDir}, destDir)

	// verify results
	verifyFileContent(t, filepath.Join(destDir, "test1.txt"), testContent1)
	verifyFileContent(t, filepath.Join(destDir, "test2.txt"), testContent2)
	verifyFileContent(t, filepath.Join(destDir, "subdir", "test3.txt"), testContent3)

	// verify subdirectory was created
	subDirInfo := verifyFileExists(t, filepath.Join(destDir, "subdir"))
	if !subDirInfo.IsDir() {
		t.Errorf("expected subdirectory, got file")
	}
}

// testArchiveWithSymlinks tests archiving and unarchiving files with symlinks.
func (s *ArchiverTestSuite) testArchiveWithSymlinks(t *testing.T) {
	// skip on Windows as symlinks require special permissions
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink test on Windows")
	}

	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a real file
	realFile := filepath.Join(tmpDir, "realfile.txt")
	realContent := "real file content"
	createTestFile(t, realFile, realContent, 0600)

	// create a symlink to the real file
	symlinkFile := filepath.Join(tmpDir, "symlink.txt")
	createSymlink(t, "realfile.txt", symlinkFile)

	// create a real directory
	realDir := filepath.Join(tmpDir, "realdir")
	createTestDir(t, realDir, 0755)

	// create a file in the real directory
	realDirFile := filepath.Join(realDir, "dirfile.txt")
	dirFileContent := "directory file content"
	createTestFile(t, realDirFile, dirFileContent, 0600)

	// create a symlink to the real directory
	symlinkDir := filepath.Join(tmpDir, "symlinkdir")
	createSymlink(t, "realdir", symlinkDir)

	// create destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{realFile, symlinkFile, realDir, symlinkDir}, destDir)

	// verify results
	verifyFileContent(t, filepath.Join(destDir, "realfile.txt"), realContent)
	verifyIsSymlink(t, filepath.Join(destDir, "symlink.txt"), "realfile.txt")
	verifyFileContent(t, filepath.Join(destDir, "realdir", "dirfile.txt"), dirFileContent)
	verifyIsSymlink(t, filepath.Join(destDir, "symlinkdir"), "realdir")
}

// testPathTraversalPrevention tests that the archiver prevents path traversal attacks.
func (s *ArchiverTestSuite) testPathTraversalPrevention(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a simple file to include in our archive
	testFile := filepath.Join(tmpDir, "safe.txt")
	createTestFile(t, testFile, "safe content", 0600)

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// create a directory outside the destination to verify files don't get extracted there
	outsideDir := filepath.Join(tmpDir, "outside")
	createTestDir(t, outsideDir, 0755)

	// create a marker file in the outside directory
	markerFile := filepath.Join(outsideDir, "marker.txt")
	createTestFile(t, markerFile, "marker content", 0600)

	// archive the safe file
	var buf bytes.Buffer
	if err := s.TestArchiver.Archive(ctx, []string{testFile}, &buf); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	// unarchive the file
	if err := s.TestArchiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir); err != nil {
		t.Fatalf("Unarchive() error = %v", err)
	}

	// verify the safe file was extracted
	verifyFileContent(t, filepath.Join(destDir, "safe.txt"), "safe content")
}

// testContextCancellation tests that the archiver respects context cancellation.
func (s *ArchiverTestSuite) testContextCancellation(t *testing.T) {
	// create a temporary directory with a test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	createTestFile(t, testFile, "test content", 0600)

	// create a canceled context
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	// try to archive with canceled context
	var buf bytes.Buffer
	err := s.TestArchiver.Archive(ctx, []string{testFile}, &buf)
	if err == nil {
		t.Errorf("Archive() with canceled context should return an error")
	}

	if errors.Is(err, context.Canceled) {
		t.Logf("expected context.Canceled error, got: %v", err)
	}

	// create a valid archive for unarchive test
	validCtx := t.Context()
	validBuf := archiveFiles(t, validCtx, s.TestArchiver, []string{testFile})

	// create destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// try to unarchive with canceled context
	err = s.TestArchiver.Unarchive(ctx, bytes.NewReader(validBuf.Bytes()), destDir)
	if err == nil {
		t.Errorf("Unarchive() with canceled context should return an error")
	}

	if errors.Is(err, context.Canceled) {
		t.Logf("expected context.Canceled error, got: %v", err)
	}
}

// testUnarchiveDirectories tests that the archiver correctly handles directories.
func (s *ArchiverTestSuite) testUnarchiveDirectories(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a directory structure to archive
	dirPath := filepath.Join(tmpDir, "testdir")
	createTestDir(t, dirPath, 0755)

	// create a file in the directory
	filePath := filepath.Join(dirPath, "testfile.txt")
	fileContent := "test content inside directory"
	createTestFile(t, filePath, fileContent, 0600)

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{dirPath}, destDir)

	// verify results
	extractedDirPath := filepath.Join(destDir, "testdir")
	dirInfo := verifyFileExists(t, extractedDirPath)
	if !dirInfo.IsDir() {
		t.Errorf("expected directory, got file")
	}

	verifyFileContent(t, filepath.Join(extractedDirPath, "testfile.txt"), fileContent)
}

// testFilePermissions tests that the archiver preserves file permissions.
func (s *ArchiverTestSuite) testFilePermissions(t *testing.T) {
	// skip on Windows as permissions work differently
	if runtime.GOOS == "windows" {
		t.Skip("skipping permissions test on Windows")
	}

	ctx := t.Context()
	tmpDir := t.TempDir()

	// create files with different permissions
	executableFile := filepath.Join(tmpDir, "executable.sh")
	executableContent := "#!/bin/sh\necho 'Hello, World!'"
	createTestFile(t, executableFile, executableContent, 0755)

	readOnlyFile := filepath.Join(tmpDir, "readonly.txt")
	readOnlyContent := "read-only content"
	createTestFile(t, readOnlyFile, readOnlyContent, 0444)

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{executableFile, readOnlyFile}, destDir)

	// verify permissions
	extractedExecutable := filepath.Join(destDir, "executable.sh")
	execInfo := verifyFileExists(t, extractedExecutable)
	if execInfo.Mode().Perm()&0111 == 0 {
		t.Errorf("executable file lost execute permission: %v", execInfo.Mode())
	}

	extractedReadOnly := filepath.Join(destDir, "readonly.txt")
	readOnlyInfo := verifyFileExists(t, extractedReadOnly)
	if readOnlyInfo.Mode().Perm()&0222 != 0 {
		t.Errorf("read-only file gained write permission: %v", readOnlyInfo.Mode())
	}
}

// testLargeFile tests that the archiver can handle larger files.
func (s *ArchiverTestSuite) testLargeFile(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		t.Skip("skipping large file test in short mode")
	}

	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a large file (10MB)
	largeFile := filepath.Join(tmpDir, "large.dat")
	largeSize := 10 * 1024 * 1024 // 10MB
	createLargeFile(t, largeFile, largeSize)

	// calculate a checksum of the original file
	originalChecksum, err := calculateChecksum(t, largeFile)
	if err != nil {
		t.Fatalf("failed to calculate original checksum: %v", err)
	}

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{largeFile}, destDir)

	// calculate checksum of the extracted file
	extractedFile := filepath.Join(destDir, "large.dat")
	extractedChecksum, err := calculateChecksum(t, extractedFile)
	if err != nil {
		t.Fatalf("failed to calculate extracted checksum: %v", err)
	}

	// compare checksums
	if originalChecksum != extractedChecksum {
		t.Errorf("checksums don't match: original=%s extracted=%s", originalChecksum, extractedChecksum)
	}
}

// testErrorHandling tests the archiver's behavior with invalid inputs.
func (s *ArchiverTestSuite) testErrorHandling(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// test archiving a non-existent file
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")
	var buf bytes.Buffer
	err := s.TestArchiver.Archive(ctx, []string{nonExistentFile}, &buf)
	if err == nil {
		t.Errorf("Archive() with non-existent file should return an error")
	}

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// test unarchiving invalid data
	invalidData := []byte("this is not a valid archive")
	err = s.TestArchiver.Unarchive(ctx, bytes.NewReader(invalidData), destDir)
	if err == nil {
		t.Errorf("Unarchive() with invalid data should return an error")
	}

	// create a valid file for testing
	validFile := filepath.Join(tmpDir, "valid.txt")
	createTestFile(t, validFile, "valid content", 0600)

	// create a valid archive
	buf.Reset()
	validBuf := archiveFiles(t, ctx, s.TestArchiver, []string{validFile})

	// test unarchiving to a non-existent directory
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")
	// make sure the directory doesn't exist
	os.RemoveAll(nonExistentDir)

	// this should actually succeed as the archiver should create the directory
	err = s.TestArchiver.Unarchive(ctx, bytes.NewReader(validBuf.Bytes()), nonExistentDir)
	if err != nil {
		t.Errorf("Unarchive() to non-existent directory should create the directory: %v", err)
	}

	// check if the directory was created
	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Errorf("destination directory was not created")
	}
}

// testModificationTimePreservation tests that the archiver preserves file modification times.
func (s *ArchiverTestSuite) testModificationTimePreservation(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "test content"
	createTestFile(t, testFile, testContent, 0600)

	// set a specific modification time (1 hour in the past)
	specificTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	setFileModTime(t, testFile, specificTime)

	// verify the modification time was set
	fileInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat test file: %v", err)
	}
	originalModTime := fileInfo.ModTime()

	// ensure the time was set correctly
	if !originalModTime.Equal(specificTime) {
		t.Fatalf("failed to set modification time correctly: got %v, want %v",
			originalModTime, specificTime)
	}

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{testFile}, destDir)

	// check the modification time of the extracted file
	extractedFile := filepath.Join(destDir, "test.txt")
	extractedInfo := verifyFileExists(t, extractedFile)
	extractedModTime := extractedInfo.ModTime()

	// compare the modification times
	// allow a small tolerance (1 second) for filesystem precision differences
	timeDiff := extractedModTime.Sub(originalModTime)
	if timeDiff < -1*time.Second || timeDiff > 1*time.Second {
		t.Errorf("modification time not preserved: original=%v extracted=%v diff=%v",
			originalModTime, extractedModTime, timeDiff)
	}
}

// testEmptyDirectories tests that the archiver correctly handles empty directories.
func (s *ArchiverTestSuite) testEmptyDirectories(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// create an empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	createTestDir(t, emptyDir, 0755)

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive and unarchive
	archiveAndUnarchive(t, ctx, s.TestArchiver, []string{emptyDir}, destDir)

	// check if the empty directory was created
	emptyDirPath := filepath.Join(destDir, "empty")
	emptyDirInfo := verifyFileExists(t, emptyDirPath)

	if !emptyDirInfo.IsDir() {
		t.Errorf("expected directory, got file")
	}
}

// testNestedSymlinkAttack tests that the archiver prevents nested symlink attacks.
func (s *ArchiverTestSuite) testSymlinkChainAttack(t *testing.T) {
	// skip on Windows as symlinks require special permissions
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink chain attack test on Windows")
	}

	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a directory structure that allows for a relative path to escape
	// the destination directory when extracted

	// create a directory for our test files
	testDir := filepath.Join(tmpDir, "test")
	createTestDir(t, testDir, 0755)

	// create a subdirectory
	subDir := filepath.Join(testDir, "subdir")
	createTestDir(t, subDir, 0755)

	// create a symlink in the subdirectory that points outside using a deep relative path
	// this path will be valid when extracted but will point outside the destination
	symlink2 := filepath.Join(subDir, "link2")
	createSymlink(t, "../../../../etc/passwd", symlink2) // deep enough to escape any destination

	// create a symlink that points to the subdirectory
	symlink1 := filepath.Join(testDir, "link1")
	createSymlink(t, "subdir", symlink1)

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive the test directory using the archiver interface
	var buf bytes.Buffer
	if err := s.TestArchiver.Archive(ctx, []string{testDir}, &buf); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	// try to unarchive - this should fail with a security error
	err := s.TestArchiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir)

	// we expect an error related to symlink security
	if err == nil {
		t.Errorf("Expected security error for symlink chain attack, but got none")
	} else {
		t.Logf("Got expected security error: %v", err)
	}
}

// testCircularSymlink tests that the archiver prevents circular symlink references.
func (s *ArchiverTestSuite) testCircularSymlink(t *testing.T) {
	t.Log("Running circular symlink test")
	// skip on Windows as symlinks require special permissions
	if runtime.GOOS == "windows" {
		t.Skip("skipping circular symlink test on Windows")
	}

	ctx := t.Context()
	tmpDir := t.TempDir()

	// create a directory for our test files
	testDir := filepath.Join(tmpDir, "test")
	createTestDir(t, testDir, 0755)

	// create circular symlinks
	symlink1 := filepath.Join(testDir, "link1")
	createSymlink(t, "link2", symlink1)

	symlink2 := filepath.Join(testDir, "link2")
	createSymlink(t, "link1", symlink2)

	// create a destination directory
	destDir := filepath.Join(tmpDir, "dest")
	createTestDir(t, destDir, 0755)

	// archive the test directory using the archiver interface
	var buf bytes.Buffer
	if err := s.TestArchiver.Archive(ctx, []string{testDir}, &buf); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	// try to unarchive - this should fail with a circular reference error
	err := s.TestArchiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir)

	// we expect an error related to circular symlinks
	if err == nil {
		t.Errorf("expected error for circular symlink reference, but got none")
	} else {
		t.Logf("got expected error: %v", err)
	}
}

// createTestFile creates a file with the given content and permissions.
func createTestFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

// createTestDir creates a directory with the given permissions
//
//nolint:unparam // perm always gets 0755 in our tests currently
func createTestDir(t *testing.T, path string, perm os.FileMode) {
	t.Helper()

	if err := os.MkdirAll(path, perm); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
}

// createSymlink creates a symbolic link.
func createSymlink(t *testing.T, oldname, newname string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink creation on Windows")
	}
	if err := os.Symlink(oldname, newname); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}
}

// archiveFiles archives the given files using the provided archiver
//
//nolint:revive // context is not first, it's fine in the test
func archiveFiles(t *testing.T, ctx context.Context, archiver Archiver, files []string) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	if err := archiver.Archive(ctx, files, &buf); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	return &buf
}

// unarchiveBuffer unarchives the given buffer to the destination directory
//
//nolint:revive // context is not first, it's fine in the test
func unarchiveBuffer(t *testing.T, ctx context.Context, archiver Archiver, buf *bytes.Buffer, destDir string) {
	t.Helper()

	if err := archiver.Unarchive(ctx, bytes.NewReader(buf.Bytes()), destDir); err != nil {
		t.Fatalf("Unarchive() error = %v", err)
	}
}

// archiveAndUnarchive performs both archive and unarchive operations
//
//nolint:revive // context is not first, it's fine in the test
func archiveAndUnarchive(t *testing.T, ctx context.Context, archiver Archiver, files []string, destDir string) {
	t.Helper()

	buf := archiveFiles(t, ctx, archiver, files)
	unarchiveBuffer(t, ctx, archiver, buf, destDir)
}

// verifyFileContent checks if a file exists and has the expected content.
func verifyFileContent(t *testing.T, path, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	if string(content) != expectedContent {
		t.Errorf("file content = %q, want %q", string(content), expectedContent)
	}
}

// verifyFileExists checks if a file exists.
func verifyFileExists(t *testing.T, path string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", path, err)
	}

	return info
}

// verifyIsSymlink checks if a file is a symlink and points to the expected target.
func verifyIsSymlink(t *testing.T, path, expectedTarget string) {
	t.Helper()

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("failed to lstat %s: %v", path, err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink", path)
	}

	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("failed to read symlink %s: %v", path, err)
	}

	if target != expectedTarget {
		t.Errorf("symlink %s points to %q, want %q", path, target, expectedTarget)
	}
}

// createLargeFile creates a file of the specified size with deterministic content.
func createLargeFile(t *testing.T, path string, size int) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}
	defer f.Close()

	// generate data
	data := make([]byte, 1024) // 1KB buffer
	for i := range size / 1024 {
		// fill with a repeating pattern
		for j := range data {
			data[j] = byte((i + j) % 256)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatalf("failed to write to large file: %v", err)
		}
	}
}

// setFileModTime sets the modification time of a file.
func setFileModTime(t *testing.T, path string, modTime time.Time) {
	t.Helper()

	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("failed to set modification time: %v", err)
	}
}

// calculateChecksum calculates the SHA256 checksum of a file.
func calculateChecksum(t *testing.T, filePath string) (string, error) {
	t.Helper()

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
