// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-vela/vela-s3-cache/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	_ "github.com/joho/godotenv/autoload"
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

	// create new CLI application
	app := cli.NewApp()

	// Plugin Information

	app.Name = "vela-s3-cache"
	app.HelpName = "vela-s3-cache"
	app.Usage = "Vela S3 cache plugin for managing a build cache in S3"
	app.Copyright = "Copyright 2020 Target Brands, Inc. All rights reserved."
	app.Authors = []*cli.Author{
		{
			Name:  "Vela Admins",
			Email: "vela@target.com",
		},
	}

	// Plugin Metadata

	app.Action = run
	app.Compiled = time.Now()
	app.Version = v.Semantic()

	// Plugin Flags
	app.Flags = []cli.Flag{

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LOG_LEVEL", "S3_CACHE_LOG_LEVEL"},
			FilePath: "/vela/parameters/s3-cache/log_level,/vela/secrets/s3-cache/log_level",
			Name:     "log.level",
			Usage:    "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Value:    "info",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ACTION", "S3_CACHE_ACTION"},
			FilePath: "/vela/parameters/s3-cache/action,/vela/secrets/s3-cache/action",
			Name:     "config.action",
			Usage:    "action to perform against the s3 cache instance",
		},

		// Cache Flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_BUCKET", "S3_CACHE_BUCKET"},
			FilePath: "/vela/parameters/s3-cache/bucket,/vela/secrets/s3-cache/bucket",
			Name:     "bucket",
			Usage:    "name of the s3 bucket",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_PREFIX", "S3_CACHE_PREFIX"},
			FilePath: "/vela/parameters/s3-cache/prefix,/vela/secrets/s3-cache/prefix",
			Name:     "prefix",
			Usage:    "path prefix for all cache default paths",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_FILENAME", "S3_CACHE_FILENAME"},
			FilePath: "/vela/parameters/s3-cache/filename,/vela/secrets/s3-cache/filename",
			Name:     "filename",
			Usage:    "Filename for the item place in the cache",
			Value:    "archive.tgz",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_PATH", "S3_CACHE_PATH"},
			FilePath: "/vela/parameters/s3-cache/path,/vela/secrets/s3-cache/path",
			Name:     "path",
			Usage:    "path to store the cache file",
		},
		&cli.DurationFlag{
			EnvVars:  []string{"PARAMETER_TIMEOUT", "S3_CACHE_TIMEOUT"},
			FilePath: "/vela/parameters/s3-cache/timeout,/vela/secrets/s3-cache/timeout",
			Name:     "timeout",
			Usage:    "Default timeout for cache requests",
			Value:    10 * time.Minute,
		},

		// Flush Flags

		&cli.DurationFlag{
			EnvVars:  []string{"PARAMETER_AGE", "PARAMETER_FLUSH_AGE", "S3_CACHE_AGE"},
			FilePath: "/vela/parameters/s3-cache/age,/vela/secrets/s3-cache/age",
			Name:     "flush.age",
			Usage:    "flush cache files older than # days",
			Value:    14 * 24 * time.Hour,
		},

		// Rebuild Flags

		&cli.StringSliceFlag{
			EnvVars:  []string{"PARAMETER_MOUNT", "S3_CACHE_MOUNT"},
			FilePath: "/vela/parameters/s3-cache/mount,/vela/secrets/s3-cache/mount",
			Name:     "rebuild.mount",
			Usage:    "list of files/directories to cache",
		},

		&cli.BoolFlag{
			EnvVars:  []string{"PARAMETER_PRESERVE_PATH", "S3_PRESERVE_PATH"},
			FilePath: "/vela/parameters/s3-cache/preserve_path,/vela/secrets/s3-cache/preserve_path",
			Name:     "rebuild.preserve_path",
			Value:    false,
			Usage:    "whether to preserve the relative directory structure during the tar process",
		},

		// S3 Flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SERVER", "CACHE_S3_SERVER", "S3_CACHE_SERVER"},
			FilePath: "/vela/parameters/s3-cache/server,/vela/secrets/s3-cache/server",
			Name:     "config.server",
			Usage:    "s3 server to store the cache",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ACCELERATED_ENDPOINT", "CACHE_S3_ACCELERATED_ENDPOINT", "S3_CACHE_ACCELERATED_ENDPOINT"},
			FilePath: "/vela/parameters/s3-cache/accelerated_endpoint,/vela/secrets/s3-cache/accelerated_endpoint",
			Name:     "config.accelerated_endpoint",
			Usage:    "s3 accelerated endpoint",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ACCESS_KEY", "S3_CACHE_ACCESS_KEY", "CACHE_S3_ACCESS_KEY", "AWS_ACCESS_KEY_ID"},
			FilePath: "/vela/parameters/s3-cache/access_key,/vela/secrets/s3-cache/access_key",
			Name:     "config.access_key",
			Usage:    "s3 access key for authentication to server",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SECRET_KEY", "S3_CACHE_SECRET_KEY", "CACHE_S3_SECRET_KEY", "AWS_SECRET_ACCESS_KEY"},
			FilePath: "/vela/parameters/s3-cache/secret_key,/vela/secrets/s3-cache/secret_key",
			Name:     "config.secret_key",
			Usage:    "s3 secret key for authentication to server",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SESSION_TOKEN", "S3_CACHE_SESSION_TOKEN", "CACHE_S3_SESSION_TOKEN", "AWS_SESSION_TOKEN"},
			FilePath: "/vela/parameters/s3-cache/session_token,/vela/secrets/s3-cache/session_token",
			Name:     "config.session_token",
			Usage:    "s3 session token",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_REGION", "CACHE_S3_REGION", "S3_CACHE_REGION"},
			FilePath: "/vela/parameters/s3-cache/region,/vela/secrets/s3-cache/region",
			Name:     "config.region",
			Usage:    "s3 region for the region of the bucket",
		},

		// Build information (for setting defaults)
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ORG", "VELA_REPO_ORG"},
			FilePath: "/vela/parameters/s3-cache/org,/vela/secrets/s3-cache/org",
			Name:     "repo.org",
			Usage:    "repository org",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_REPO", "VELA_REPO_NAME"},
			FilePath: "/vela/parameters/s3-cache/repo,/vela/secrets/s3-cache/repo",
			Name:     "repo.name",
			Usage:    "repository name",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_REPO_BRANCH", "VELA_REPO_BRANCH"},
			FilePath: "/vela/parameters/s3-cache/repo_branch,/vela/secrets/s3-cache/repo_branch",
			Name:     "repo.branch",
			Usage:    "default branch for the repository",
			Value:    "main",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_BUILD_BRANCH", "VELA_BUILD_BRANCH"},
			FilePath: "/vela/parameters/s3-cache/build_branch,/vela/secrets/s3-cache/repo/build_branch",
			Name:     "repo.build.branch",
			Usage:    "git build branch",
			Value:    "main",
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

// run executes the plugin based off the configuration provided.
func run(c *cli.Context) error {
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
			Bucket:       c.String("bucket"),
			Filename:     c.String("filename"),
			Timeout:      c.Duration("timeout"),
			Mount:        c.StringSlice("rebuild.mount"),
			Path:         c.String("path"),
			Prefix:       c.String("prefix"),
			PreservePath: c.Bool("rebuild.preserve_path"),
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
	return p.Exec()
}
