// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// processItem processes an file system item based on its type.
func (t *TarGzipArchiver) processItem(ctx context.Context, header *tar.Header, targetPath string, tarReader *tar.Reader, destAbs string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return t.processDirectory(targetPath, header)
	case tar.TypeReg, tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
		return t.processFile(ctx, targetPath, header, tarReader)
	case tar.TypeSymlink:
		return t.processSymlink(header, targetPath, destAbs)
	case tar.TypeLink:
		return t.processHardLink(header, targetPath, destAbs)
	default:
		return fmt.Errorf("unsupported tar header type: %s (%d)", header.Name, header.Typeflag)
	}
}

// processDirectory creates a directory.
func (t *TarGzipArchiver) processDirectory(targetPath string, header *tar.Header) error {
	if err := os.MkdirAll(targetPath, header.FileInfo().Mode()); err != nil {
		return err
	}

	return os.Chtimes(targetPath, time.Now(), header.ModTime)
}

// processFile extracts a file from a tar archive.
func (t *TarGzipArchiver) processFile(ctx context.Context, path string, header *tar.Header, reader *tar.Reader) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// check if file already exists
	// this can happen when you created an archive with PreservePath turned off
	// and archived two files with the same name in different locations.
	// it does not get deduped in the archive.
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file conflict detected: %s already exists", path)
	} else if !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode())
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing file: %w", closeErr)
		}
	}()

	limitedReader := io.LimitReader(reader, header.Size)

	buffer := make([]byte, 32*1024) // 32KB buffer
	if _, err = io.CopyBuffer(file, limitedReader, buffer); err != nil {
		return err
	}

	return os.Chtimes(path, time.Now(), header.ModTime)
}

// processSymlink creates a symbolic link.
func (t *TarGzipArchiver) processSymlink(header *tar.Header, targetPath string, destAbs string) error {
	linkTarget := header.Linkname

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// check for absolute symlinks
	if filepath.IsAbs(linkTarget) {
		return fmt.Errorf("absolute symlinks are not supported: %s -> %s",
			header.Name, header.Linkname)
	}

	linkDir := filepath.Dir(targetPath)
	//nolint:gosec // G305: File traversal handled in isPathWithinBoundary
	resolvedTarget := filepath.Join(linkDir, linkTarget)
	resolvedTarget = filepath.Clean(resolvedTarget)

	// verify that the symlink target doesn't escape the destination directory
	if !isPathWithinBoundary(resolvedTarget, destAbs) {
		return fmt.Errorf("symlink target path traversal attempt detected: %s -> %s (resolves to %s)",
			header.Name, header.Linkname, resolvedTarget)
	}

	// check for direct circular references
	if resolvedTarget == targetPath {
		return fmt.Errorf("circular symlink reference detected: %s -> %s", header.Name, header.Linkname)
	}

	// check if the target is already a symlink that points back to this symlink
	if existingTarget, isSymlink := t.extractedSymlinks[resolvedTarget]; isSymlink {
		backTarget := filepath.Join(filepath.Dir(resolvedTarget), existingTarget)
		backTarget = filepath.Clean(backTarget)

		if backTarget == targetPath {
			return fmt.Errorf("circular symlink reference detected: %s -> %s -> %s",
				header.Name, header.Linkname, header.Name)
		}
	}

	// check for symlink chains that could lead outside the destination
	if err := t.checkSymlinkChain(targetPath, resolvedTarget, destAbs, 0); err != nil {
		return err
	}

	// remove existing file/directory before creating symlink
	if err := os.RemoveAll(targetPath); err != nil {
		return err
	}

	if err := os.Symlink(header.Linkname, targetPath); err != nil {
		return err
	}

	// track this symlink for future chain validation
	t.extractedSymlinks[targetPath] = linkTarget

	return nil
}

// processHardLink creates a hard link.
func (t *TarGzipArchiver) processHardLink(header *tar.Header, targetPath string, destAbs string) error {
	//nolint:gosec // G305: File traversal handled in isPathWithinBoundary
	linkTarget := filepath.Join(destAbs, header.Linkname)

	if !isPathWithinBoundary(linkTarget, destAbs) {
		return fmt.Errorf("hard link target path traversal attempt detected: %s -> %s",
			header.Name, header.Linkname)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// remove existing file/directory before creating hard link
	if err := os.RemoveAll(targetPath); err != nil {
		return err
	}

	return os.Link(linkTarget, targetPath)
}

// getTargetPath calculates the target path for a file and checks for path traversal.
func (t *TarGzipArchiver) getTargetPath(name string, destAbs string) (string, error) {
	cleanedName := filepath.Clean(name)

	// check for absolute paths (even if not for current system)
	if strings.HasPrefix(cleanedName, string(filepath.Separator)) ||
		(filepath.Separator != '/' && strings.HasPrefix(cleanedName, "/")) ||
		(filepath.Separator != '\\' && strings.HasPrefix(cleanedName, "\\")) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", name)
	}

	// check for windows-style absolute paths (e.g., C:\Windows). not really needed
	// for our use case, but here just in case.
	if len(cleanedName) > 1 && cleanedName[1] == ':' &&
		((cleanedName[0] >= 'A' && cleanedName[0] <= 'Z') || (cleanedName[0] >= 'a' && cleanedName[0] <= 'z')) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", name)
	}

	// join the destination with the cleaned name
	targetPath := filepath.Join(destAbs, cleanedName)

	// check if the target path is within the destination directory
	if !isPathWithinBoundary(targetPath, destAbs) {
		return "", fmt.Errorf("path traversal detected: %s", name)
	}

	return targetPath, nil
}

// checkSymlinkChain recursively follows symlink chains to detect path traversal attempts
// and circular references.
func (t *TarGzipArchiver) checkSymlinkChain(originalLink, targetPath, destAbs string, depth int) error {
	maxDepth := 10

	// prevent infinite recursion
	if depth >= maxDepth {
		return fmt.Errorf("symlink chain too deep (max %d): %s", maxDepth, originalLink)
	}

	// check if the target is a previously extracted symlink
	if linkTarget, isSymlink := t.extractedSymlinks[targetPath]; isSymlink {
		// this target is itself a symlink, so we need to follow it
		linkDir := filepath.Dir(targetPath)

		// resolve the next target in the chain
		nextTarget := filepath.Join(linkDir, linkTarget)
		nextTarget = filepath.Clean(nextTarget)

		// check if this link in the chain escapes the destination
		if !isPathWithinBoundary(nextTarget, destAbs) {
			return fmt.Errorf("symlink chain traversal detected: %s -> ... -> %s (resolves outside destination)",
				originalLink, nextTarget)
		}

		// check for circular references - if the next target is the original link
		if nextTarget == originalLink {
			return fmt.Errorf("circular symlink reference detected: %s -> ... -> %s", originalLink, nextTarget)
		}

		// recursively check the next link in the chain
		return t.checkSymlinkChain(originalLink, nextTarget, destAbs, depth+1)
	}

	return nil
}
