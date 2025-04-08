// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// archiveSource archives a single source path to the tar writer.
func (t *TarGzipArchiver) archiveSource(ctx context.Context, source string, tarWriter *tar.Writer) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		header, err := t.createHeader(path, info)
		if err != nil {
			return err
		}

		// set the header name based on the path
		if err := t.setHeaderName(header, source, path); err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			return t.copyFileContent(path, tarWriter)
		}

		return nil
	})
}

// createHeader creates a tar header for the given file info.
func (t *TarGzipArchiver) createHeader(path string, info os.FileInfo) (*tar.Header, error) {
	// handle symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}

		header, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return nil, err
		}

		header.Typeflag = tar.TypeSymlink
		header.Linkname = linkTarget

		return header, nil
	}

	// regular file or directory
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil, err
	}

	return header, nil
}

// setHeaderName sets the name in the tar header based on the path.
func (t *TarGzipArchiver) setHeaderName(header *tar.Header, source, path string) error {
	// get file info to determine if source is a directory
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		// source is a directory - always preserve the directory name regardless of PreservePath setting
		// calculate relative path from source directory to the current file
		relPath, err := filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}

		// this will include the source directory name in the path
		header.Name = relPath
	} else {
		// source is a file - honor the PreservePath setting
		if t.PreservePath {
			// when preserving paths for a file, use the source path as is
			header.Name = source
		} else {
			// when not preserving paths for a file, just use the base name
			header.Name = filepath.Base(path)
		}
	}

	// ensure directories end with a slash
	if header.Typeflag == tar.TypeDir && !strings.HasSuffix(header.Name, "/") {
		header.Name += "/"
	}

	// ensure the header name doesn't start with a slash
	header.Name = strings.TrimPrefix(header.Name, "/")

	return nil
}

// copyFileContent copies the content of a file to the tar writer.
func (t *TarGzipArchiver) copyFileContent(path string, tarWriter *tar.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// limit the reader to the file size to prevent reading more data than expected
	limitedReader := io.LimitReader(file, fileInfo.Size())

	// use a buffer for better performance
	// 32KB for balance of memory usage and performance
	buffer := make([]byte, 32*1024)
	_, err = io.CopyBuffer(tarWriter, limitedReader, buffer)

	return err
}
