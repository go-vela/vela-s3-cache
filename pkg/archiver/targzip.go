// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TarGzipArchiver is an Archiver that compresses and adds files to a tar archive.
type TarGzipArchiver struct {
	CompressionLevel int
	PreservePath     bool

	extractedSymlinks map[string]string
}

// make sure TarGzipArchiver implements Archiver.
var _ Archiver = &TarGzipArchiver{}

// Archive compresses and adds files to a tar archive.
func (t *TarGzipArchiver) Archive(ctx context.Context, src []string, dest io.Writer) (err error) {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	filteredSrc, err := filterRedundantPaths(src)
	if err != nil {
		return fmt.Errorf("failed to filter redundant paths: %w", err)
	}

	gzipWriter, err := gzip.NewWriterLevel(dest, t.CompressionLevel)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := gzipWriter.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)

	defer func() {
		closeErr := tarWriter.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for _, source := range filteredSrc {
		if err := t.archiveSource(ctx, source, tarWriter); err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return nil
}

// Unarchive decompresses and extracts files from a tar archive.
func (t *TarGzipArchiver) Unarchive(ctx context.Context, src io.Reader, dest string) (err error) {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// initialize symlink tracking for this extraction
	t.extractedSymlinks = make(map[string]string)

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of destination: %w", err)
	}

	if err := os.MkdirAll(destAbs, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	gzipReader, err := gzip.NewReader(src)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := gzipReader.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	tarReader := tar.NewReader(gzipReader)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// get the target path and check for path traversal
		targetPath, err := t.getTargetPath(header.Name, destAbs)
		if err != nil {
			return err
		}

		// process the file based on its type
		if err := t.processItem(ctx, header, targetPath, tarReader, destAbs); err != nil {
			return err
		}
	}

	return nil
}
