// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/klauspost/compress/flate"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	_ "github.com/joho/godotenv/autoload"

	"github.com/go-vela/vela-s3-cache/version"
)

//nolint:funlen // ignore function length due to comments and flags
func main() {
	// capture application version information
	v := version.New()

	// serialize the version information as pretty JSON
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logrus.Fatal(err)
	}

	// output the version information to stdout
	fmt.Fprintf(os.Stdout, "%s\n", string(bytes))

	cmd := &cli.Command{
		Name:      "vela-s3-cache",
		Version:   v.Semantic(),
		Usage:     "Vela S3 cache plugin for managing a build cache in S3",
		Copyright: "Copyright 2020 Target Brands, Inc. All rights reserved.",
		Action:    run,
	}

	// Plugin Flags
	cmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "log.level",
			Usage: "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_LOG_LEVEL"),
				cli.EnvVar("S3_CACHE_LOG_LEVEL"),
				cli.File("/vela/parameters/s3-cache/log_level"),
				cli.File("/vela/secrets/s3-cache/log_level"),
			),
			Value: "info",
		},
		&cli.StringFlag{
			Name:  "config.action",
			Usage: "action to perform against the s3 cache instance",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_ACTION"),
				cli.EnvVar("S3_CACHE_ACTION"),
				cli.File("/vela/parameters/s3-cache/action"),
				cli.File("/vela/secrets/s3-cache/action"),
			),
		},

		// Cache Flags
		&cli.StringFlag{
			Name:  "bucket",
			Usage: "name of the s3 bucket",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_BUCKET"),
				cli.EnvVar("S3_CACHE_BUCKET"),
				cli.File("/vela/parameters/s3-cache/bucket"),
				cli.File("/vela/secrets/s3-cache/bucket"),
			),
		},
		&cli.StringFlag{
			Name:  "prefix",
			Usage: "path prefix for all cache default paths",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_PREFIX"),
				cli.EnvVar("S3_CACHE_PREFIX"),
				cli.File("/vela/parameters/s3-cache/prefix"),
				cli.File("/vela/secrets/s3-cache/prefix"),
			),
		},
		&cli.StringFlag{
			Name:  "filename",
			Usage: "filename for the item place in the cache",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_FILENAME"),
				cli.EnvVar("S3_CACHE_FILENAME"),
				cli.File("/vela/parameters/s3-cache/filename"),
				cli.File("/vela/secrets/s3-cache/filename"),
			),
			Value: "archive.tgz",
		},
		&cli.StringFlag{
			Name:  "path",
			Usage: "path to store the cache file",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_PATH"),
				cli.EnvVar("S3_CACHE_PATH"),
				cli.File("/vela/parameters/s3-cache/path"),
				cli.File("/vela/secrets/s3-cache/path"),
			),
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "default timeout for cache requests",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_TIMEOUT"),
				cli.EnvVar("S3_CACHE_TIMEOUT"),
				cli.File("/vela/parameters/s3-cache/timeout"),
				cli.File("/vela/secrets/s3-cache/timeout"),
			),
			Value: 10 * time.Minute,
		},

		// Flush Flags
		&cli.DurationFlag{
			Category: "Flush",
			Name:     "flush.age",
			Usage:    "flush cache files older than # days",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_AGE"),
				cli.EnvVar("PARAMETER_FLUSH_AGE"),
				cli.EnvVar("S3_CACHE_AGE"),
				cli.File("/vela/parameters/s3-cache/age"),
				cli.File("/vela/secrets/s3-cache/age"),
			),
			Value: 14 * 24 * time.Hour,
		},

		// Rebuild Flags
		&cli.IntFlag{
			Category: "Rebuild",
			Name:     "rebuild.compression_level",
			Usage:    "compression level for the cache file (-1 to 9)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_COMPRESSION_LEVEL"),
				cli.EnvVar("S3_CACHE_COMPRESSION_LEVEL"),
				cli.File("/vela/parameters/s3-cache/compression_level"),
				cli.File("/vela/secrets/s3-cache/compression_level"),
			),
			Value: flate.DefaultCompression, // -1 is the carryover default value from <v0.9.0 of this plugin
		},
		&cli.StringSliceFlag{
			Category: "Rebuild",
			Name:     "rebuild.mount",
			Usage:    "list of files/directories to cache",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_MOUNT"),
				cli.EnvVar("S3_CACHE_MOUNT"),
				cli.File("/vela/parameters/s3-cache/mount"),
				cli.File("/vela/secrets/s3-cache/mount"),
			),
		},
		&cli.BoolFlag{
			Category: "Rebuild",
			Name:     "rebuild.preserve_path",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_PRESERVE_PATH"),
				cli.EnvVar("S3_PRESERVE_PATH"),
				cli.File("/vela/parameters/s3-cache/preserve_path"),
				cli.File("/vela/secrets/s3-cache/preserve_path"),
			),
			Value: false,
			Usage: "whether to preserve the relative directory structure during the tar process",
		},

		// S3 Flags
		&cli.StringFlag{
			Name:  "config.server",
			Usage: "s3 server to store the cache",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SERVER"),
				cli.EnvVar("CACHE_S3_SERVER"),
				cli.EnvVar("S3_CACHE_SERVER"),
				cli.File("/vela/parameters/s3-cache/server"),
				cli.File("/vela/secrets/s3-cache/server"),
			),
		},
		&cli.StringFlag{
			Name:  "config.accelerated_endpoint",
			Usage: "s3 accelerated endpoint",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_ACCELERATED_ENDPOINT"),
				cli.EnvVar("CACHE_S3_ACCELERATED_ENDPOINT"),
				cli.EnvVar("S3_CACHE_ACCELERATED_ENDPOINT"),
				cli.File("/vela/parameters/s3-cache/accelerated_endpoint"),
				cli.File("/vela/secrets/s3-cache/accelerated_endpoint"),
			),
		},
		&cli.StringFlag{
			Name:  "config.access_key",
			Usage: "s3 access key for authentication to server",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_ACCESS_KEY"),
				cli.EnvVar("S3_CACHE_ACCESS_KEY"),
				cli.EnvVar("CACHE_S3_ACCESS_KEY"),
				cli.EnvVar("AWS_ACCESS_KEY_ID"),
				cli.File("/vela/parameters/s3-cache/access_key"),
				cli.File("/vela/secrets/s3-cache/access_key"),
			),
		},
		&cli.StringFlag{
			Name:  "config.secret_key",
			Usage: "s3 secret key for authentication to server",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SECRET_KEY"),
				cli.EnvVar("S3_CACHE_SECRET_KEY"),
				cli.EnvVar("CACHE_S3_SECRET_KEY"),
				cli.EnvVar("AWS_SECRET_ACCESS_KEY"),
				cli.File("/vela/parameters/s3-cache/secret_key"),
				cli.File("/vela/secrets/s3-cache/secret_key"),
			),
		},
		&cli.StringFlag{
			Name:  "config.session_token",
			Usage: "s3 session token",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SESSION_TOKEN"),
				cli.EnvVar("S3_CACHE_SESSION_TOKEN"),
				cli.EnvVar("CACHE_S3_SESSION_TOKEN"),
				cli.EnvVar("AWS_SESSION_TOKEN"),
				cli.File("/vela/parameters/s3-cache/session_token"),
				cli.File("/vela/secrets/s3-cache/session_token"),
			),
		},
		&cli.StringFlag{
			Name:  "config.region",
			Usage: "s3 region for the region of the bucket",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_REGION"),
				cli.EnvVar("CACHE_S3_REGION"),
				cli.EnvVar("S3_CACHE_REGION"),
				cli.File("/vela/parameters/s3-cache/region"),
				cli.File("/vela/secrets/s3-cache/region"),
			),
		},

		// Build information (for setting defaults)
		&cli.StringFlag{
			Name:  "repo.org",
			Usage: "repository org",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_ORG"),
				cli.EnvVar("VELA_REPO_ORG"),
				cli.File("/vela/parameters/s3-cache/org"),
				cli.File("/vela/secrets/s3-cache/org"),
			),
		},
		&cli.StringFlag{
			Name:  "repo.name",
			Usage: "repository name",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_REPO"),
				cli.EnvVar("VELA_REPO_NAME"),
				cli.File("/vela/parameters/s3-cache/repo"),
				cli.File("/vela/secrets/s3-cache/repo"),
			),
		},
		&cli.StringFlag{
			Name:  "repo.branch",
			Usage: "default branch for the repository",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_REPO_BRANCH"),
				cli.EnvVar("VELA_REPO_BRANCH"),
				cli.File("/vela/parameters/s3-cache/repo_branch"),
				cli.File("/vela/secrets/s3-cache/repo_branch"),
			),
			Value: "main",
		},
		&cli.StringFlag{
			Name:  "repo.build.branch",
			Usage: "git build branch",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_BUILD_BRANCH"),
				cli.EnvVar("VELA_BUILD_BRANCH"),
				cli.File("/vela/parameters/s3-cache/build_branch"),
				cli.File("/vela/secrets/s3-cache/build_branch"),
			),
			Value: "main",
		},
	}

	if err = cmd.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

// run executes the plugin based off the configuration provided.
func run(ctx context.Context, c *cli.Command) error {
	// set the log level for the plugin
	switch c.String("log.level") {
	case "t", "trace", "Trace", "TRACE":
		logrus.SetLevel(logrus.TraceLevel)
	case "d", "debug", "Debug", "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "w", "warn", "Warn", "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "e", "error", "Error", "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "f", "fatal", "Fatal", "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "p", "panic", "Panic", "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	case "i", "info", "Info", "INFO":
		fallthrough
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.WithFields(logrus.Fields{
		"code":     "https://github.com/go-vela/vela-s3-cache",
		"docs":     "https://go-vela.github.io/docs/plugins/registry/pipeline/s3_cache",
		"registry": "https://hub.docker.com/r/target/vela-s3-cache",
	}).Info("Vela S3 Cache Plugin")

	// create the plugin
	p := &Plugin{
		// config configuration
		Config: &Config{
			Action:              c.String("config.action"),
			Server:              c.String("config.server"),
			AcceleratedEndpoint: c.String("config.accelerated_endpoint"),
			AccessKey:           c.String("config.access_key"),
			SecretKey:           c.String("config.secret_key"),
			SessionToken:        c.String("config.session_token"),
			Region:              c.String("config.region"),
		},
		// flush configuration
		Flush: &Flush{
			Bucket: c.String("bucket"),
			Age:    c.Duration("flush.age"),
			Path:   c.String("path"),
			Prefix: c.String("prefix"),
		},
		// rebuild configuration
		Rebuild: &Rebuild{
			Bucket:           c.String("bucket"),
			CompressionLevel: c.Int("rebuild.compression_level"),
			Filename:         c.String("filename"),
			Timeout:          c.Duration("timeout"),
			Mount:            c.StringSlice("rebuild.mount"),
			Path:             c.String("path"),
			Prefix:           c.String("prefix"),
			PreservePath:     c.Bool("rebuild.preserve_path"),
		},
		// restore configuration
		Restore: &Restore{
			Bucket:   c.String("bucket"),
			Filename: c.String("filename"),
			Timeout:  c.Duration("timeout"),
			Path:     c.String("path"),
			Prefix:   c.String("prefix"),
		},
		// repository configuration from environment
		Repo: &Repo{
			Owner:       c.String("repo.org"),
			Name:        c.String("repo.name"),
			Branch:      c.String("repo.branch"),
			BuildBranch: c.String("repo.build.branch"),
		},
	}

	// validate the plugin
	err := p.Validate()
	if err != nil {
		return err
	}

	// execute the plugin
	return p.Exec(ctx)
}
