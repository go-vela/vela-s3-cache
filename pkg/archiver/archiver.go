// SPDX-License-Identifier: Apache-2.0

package archiver

import (
	"context"
	"fmt"
	"io"
)

// Archiver is the interface for archiving and unarchiving files. It should be implemented by all archivers.
type Archiver interface {
	Archive(ctx context.Context, src []string, dest io.Writer) error
	Unarchive(ctx context.Context, src io.Reader, dest string) error
}

// Option is a function that can be used to configure an Archiver.
type Option func(*Options)

// Options are the options for an Archiver.
type Options struct {
	CompressionLevel int64
	PreservePath     bool
}

// WithCompressionLevel sets the compression level for the archiver.
func WithCompressionLevel(level int64) Option {
	return func(o *Options) {
		o.CompressionLevel = level
	}
}

// WithPreservePath sets whether to preserve the path of the source files in the archive.
// If the source is a file, this settings controls whether the path will be preserved.
// For directories, the path is always preserved.
func WithPreservePath(preservePath bool) Option {
	return func(o *Options) {
		o.PreservePath = preservePath
	}
}

// NewArchiver creates a new Archiver based on the given format and options.
func NewArchiver(format string, opts ...Option) (Archiver, error) {
	// defaults, although we always send in what is
	// configured in the plugin (or its defaults)
	options := &Options{
		CompressionLevel: -1,
		PreservePath:     false,
	}

	// apply options
	for _, opt := range opts {
		opt(options)
	}

	// create archiver based on format
	switch format {
	case "tar.gz":
		return &TarGzipArchiver{
			CompressionLevel: options.CompressionLevel,
			PreservePath:     options.PreservePath,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported archive format: %s (supported formats: 'tar.gz')", format)
	}
}
