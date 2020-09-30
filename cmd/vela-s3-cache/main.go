// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := cli.NewApp()

	// Plugin Information

	app.Name = "vela-s3-cache"
	app.HelpName = "vela-s3-cache"
	app.Usage = "Vela S3 cache plugin for managing a build cache in S3"
	app.Copyright = "Copyright (c) 2020 Target Brands, Inc. All rights reserved."
	app.Authors = []*cli.Author{
		{
			Name:  "Vela Admins",
			Email: "vela@target.com",
		},
	}

	// Plugin Metadata

	app.Compiled = time.Now()
	app.Action = run

	// Plugin Flags

	app.Flags = []cli.Flag{

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_LOG_LEVEL", "VELA_LOG_LEVEL", "ARTIFACTORY_LOG_LEVEL"},
			FilePath: string("/vela/parameters/s3_cache/log_level,/vela/secrets/s3_cache/log_level"),
			Name:     "log.level",
			Usage:    "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Value:    "info",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ACTION", "CONFIG_ACTION", "ARTIFACTORY_ACTION"},
			FilePath: string("/vela/parameters/s3_cache/config/action,/vela/secrets/s3_cache/config/action"),
			Name:     "config.action",
			Usage:    "action to perform against the s3 cache instance",
		},

		// Cache Flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ROOT"},
			FilePath: string("/vela/parameters/s3_cache/root,/vela/secrets/s3_cache/root"),
			Name:     "root",
			Usage:    "name of the s3 bucket",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_PREFIX"},
			FilePath: string("/vela/parameters/s3_cache/prefix,/vela/secrets/s3_cache/prefix"),
			Name:     "prefix",
			Usage:    "path prefix for all cache default paths",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_FILENAME"},
			FilePath: string("/vela/parameters/s3_cache/filename,/vela/secrets/s3_cache/filename"),
			Name:     "filename",
			Usage:    "Filename for the item place in the cache",
			Value:    "archive.tgz",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_PATH"},
			FilePath: string("/vela/parameters/s3_cache/path,/vela/secrets/s3_cache/path"),
			Name:     "path",
			Usage:    "path to store the cache file",
		},
		&cli.StringSliceFlag{
			EnvVars:  []string{"PARAMETER_MOUNT"},
			FilePath: string("/vela/parameters/s3_cache/mount,/vela/secrets/s3_cache/mount"),
			Name:     "mount",
			Usage:    "list of files/directories to cache",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_FLUSH_AGE"},
			FilePath: string("/vela/parameters/s3_cache/age,/vela/secrets/s3_cache/age"),
			Name:     "age",
			Usage:    "flush cache files older than # days",
		},
		&cli.DurationFlag{
			EnvVars:  []string{"PARAMETER_TIMEOUT"},
			FilePath: string("/vela/parameters/s3_cache/timeout,/vela/secrets/s3_cache/timeout"),
			Name:     "timeout",
			Usage:    "Default timeout for cache requests",
			Value:    10 * time.Minute,
		},

		// S3 Flags

		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SERVER", "PARAMETER_ENDPOINT", "CACHE_S3_ENDPOINT", "CACHE_S3_SERVER", "S3_ENDPOINT"},
			FilePath: string("/vela/parameters/s3_cache/config/server,/vela/secrets/s3_cache/config/server"),
			Name:     "config.server",
			Usage:    "s3 server to store the cache",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ACCELERATED_ENDPOINT", "CACHE_S3_ACCELERATED_ENDPOINT"},
			FilePath: string("/vela/parameters/s3_cache/config/accelerated_endpoint,/vela/secrets/s3_cache/config/accelerated_endpoint"),
			Name:     "config.accelerated-endpoint",
			Usage:    "s3 accelerated endpoint",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_ACCESS_KEY", "CACHE_S3_ACCESS_KEY", "AWS_ACCESS_KEY_ID"},
			FilePath: string("/vela/parameters/s3_cache/config/access_key,/vela/secrets/s3_cache/config/access_key"),
			Name:     "config.access-key",
			Usage:    "s3 access key for authentication to server",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SECRET_KEY", "CACHE_S3_SECRET_KEY", "AWS_SECRET_ACCESS_KEY"},
			FilePath: string("/vela/parameters/s3_cache/config/secret_key,/vela/secrets/s3_cache/config/secret_key"),
			Name:     "config.secret-key",
			Usage:    "s3 secret key for authentication to server",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_SESSION_TOKEN", "CACHE_S3_SESSION_TOKEN", "AWS_SESSION_TOKEN"},
			FilePath: string("/vela/parameters/s3_cache/config/session_token,/vela/secrets/s3_cache/config/session_token"),
			Name:     "config.session-token",
			Usage:    "s3 session token",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_REGION", "CACHE_S3_REGION"},
			FilePath: string("/vela/parameters/s3_cache/config/region,/vela/secrets/s3_cache/config/region"),
			Name:     "config.region",
			Usage:    "s3 region for the region of the bucket",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_CA_CERT", "CACHE_S3_CA_CERT"},
			FilePath: string("/vela/parameters/s3_cache/config/ca_cert,/vela/secrets/s3_cache/config/ca_cert"),
			Name:     "config.ca-cert",
			Usage:    "ca cert to connect to s3 server",
		},
		&cli.StringFlag{
			EnvVars:  []string{"PARAMETER_CA_CERT_PATH", "CACHE_S3_CA_CERT_PATH"},
			FilePath: string("/vela/parameters/s3_cache/config/ca_cert_path,/vela/secrets/s3_cache/config/ca_cert_path"),
			Name:     "config.ca-cert-path",
			Usage:    "location of the ca cert to connect to s3 server",
			Value:    "/etc/ssl/certs/ca-certificates.crt",
		},

		// Build information (for setting defaults)
		&cli.StringFlag{
			EnvVars:  []string{"REPOSITORY_ORG"},
			FilePath: string("/vela/parameters/s3_cache/repo/owner,/vela/secrets/s3_cache/repo/owner"),
			Name:     "repo.owner",
			Usage:    "repository owner",
		},
		&cli.StringFlag{
			EnvVars:  []string{"REPOSITORY_NAME"},
			FilePath: string("/vela/parameters/s3_cache/repo/name,/vela/secrets/s3_cache/repo/name"),
			Name:     "repo.name",
			Usage:    "repository name",
		},
		&cli.StringFlag{
			EnvVars:  []string{"REPOSITORY_BRANCH"},
			FilePath: string("/vela/parameters/s3_cache/repo/branch,/vela/secrets/s3_cache/repo/branch"),
			Name:     "repo.branch",
			Usage:    "repository default branch",
			Value:    "master",
		},
		&cli.StringFlag{
			EnvVars:  []string{"REPOSITORY_BRANCH"},
			FilePath: string("/vela/parameters/s3_cache/repo/branch,/vela/secrets/s3_cache/repo/branch"),
			Name:     "repo.commit.branch",
			Usage:    "git commit branch",
			Value:    "master",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

// run executes the plugin based off the configuration provided.
func run(c *cli.Context) (err error) {
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
		"code": "https://github.com/go-vela/vela-s3-cache",
		"docs": "https://go-vela.github.io/docs/plugins/registry/s3-cache",
		"time": time.Now(),
	}).Info("Vela S3 Cache Plugin")

	// create the plugin
	p := Plugin{
		// config configuration
		Config: &Config{
			Action:              c.String("config.action"),
			Server:              c.String("config.server"),
			AcceleratedEndpoint: c.String("config.accelerated-endpoint"),
			AccessKey:           c.String("config.access-key"),
			SecretKey:           c.String("config.secret-key"),
			SessionToken:        c.String("session-token"),
			Region:              c.String("config.region"),
			CaCert:              c.String("config.ca-cert"),
			CaCertPath:          c.String("config.ca-cert-path"),
		},
		// flush configuration
		Flush: &Flush{
			Root:   c.String("root"),
			Prefix: c.String("prefix"),
			Path:   c.String("path"),
		},
		// rebuild configuration
		Rebuild: &Rebuild{
			Root:     c.String("root"),
			Filename: c.String("filename"),
			Timeout:  c.Duration("timeout"),
			Mount:    c.StringSlice("mount"),
			Prefix:   c.String("prefix"),
		},
		// restore configuration
		Restore: &Restore{
			Root:     c.String("root"),
			Filename: c.String("filename"),
			Timeout:  c.Duration("timeout"),
			Prefix:   c.String("prefix"),
		},
		// repository configuration from environment
		Repo: Repo{
			Owner:        c.String("repo.owner"),
			Name:         c.String("repo.name"),
			Branch:       c.String("repo.branch"),
			CommitBranch: c.String("repo.commit.branch"),
		},
	}

	logrus.Info("hellope world")

	// validate the plugin
	err = p.Validate()
	if err != nil {
		return err
	}

	// execute the plugin
	return p.Exec()
}
