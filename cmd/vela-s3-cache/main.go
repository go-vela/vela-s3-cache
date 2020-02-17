// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := cli.NewApp()
	app.Name = "s3 cache plugin"
	app.Usage = "s3 cache plugin"
	app.Action = run
	app.Flags = []cli.Flag{

		cli.StringFlag{
			EnvVar: "PARAMETER_LOG_LEVEL,VELA_LOG_LEVEL,ARTIFACTORY_LOG_LEVEL",
			Name:   "log.level",
			Usage:  "set log level - options: (trace|debug|info|warn|error|fatal|panic)",
			Value:  "info",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_ACTION,CONFIG_ACTION,ARTIFACTORY_ACTION",
			Name:   "config.action",
			Usage:  "action to perform against the s3 cache instance",
		},

		// Cache information
		cli.StringFlag{
			EnvVar: "PARAMETER_ROOT",
			Name:   "root",
			Usage:  "path prefix for all cache default paths",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_FILENAME",
			Name:   "filename",
			Usage:  "Filename for the item place in the cache",
			Value:  "archive.tgz",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_PATH",
			Name:   "path",
			Usage:  "path to store the cache file",
		},
		cli.StringSliceFlag{
			EnvVar: "PARAMETER_MOUNT",
			Name:   "mount",
			Usage:  "list of files/directories to cache",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_FLUSH_AGE",
			Name:   "age",
			Usage:  "flush cache files older than # days",
		},
		cli.DurationFlag{
			EnvVar: "PARAMETER_TIMEOUT",
			Name:   "timeout",
			Usage:  "Default timeout for cache requests",
			Value:  10 * time.Minute,
		},

		// S3 information
		cli.StringFlag{
			EnvVar: "PARAMETER_SERVER,PARAMETER_ENDPOINT,CACHE_S3_ENDPOINT,CACHE_S3_SERVER,S3_ENDPOINT",
			Name:   "config.server",
			Usage:  "s3 server to store the cache",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_ACCELERATED_ENDPOINT,CACHE_S3_ACCELERATED_ENDPOINT",
			Name:   "config.accelerated-endpoint",
			Usage:  "s3 accelerated endpoint",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_ACCESS_KEY,CACHE_S3_ACCESS_KEY,AWS_ACCESS_KEY_ID",
			Name:   "config.access-key",
			Usage:  "s3 access key for authentication to server",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_SECRET_KEY,CACHE_S3_SECRET_KEY,AWS_SECRET_ACCESS_KEY",
			Name:   "config.secret-key",
			Usage:  "s3 secret key for authentication to server",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_SESSION_TOKEN,CACHE_S3_SESSION_TOKEN,AWS_SESSION_TOKEN",
			Name:   "config.session-token",
			Usage:  "s3 session token",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_REGION,CACHE_S3_REGION",
			Name:   "config.region",
			Usage:  "s3 region for the region of the bucket",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_CA_CERT,CACHE_S3_CA_CERT",
			Name:   "config.ca_cert",
			Usage:  "ca cert to connect to s3 server",
		},
		cli.StringFlag{
			EnvVar: "PARAMETER_CA_CERT_PATH,CACHE_S3_CA_CERT_PATH",
			Name:   "config.ca_cert_path",
			Usage:  "location of the ca cert to connect to s3 server",
			Value:  "/etc/ssl/certs/ca-certificates.crt",
		},

		// Build information (for setting defaults)
		cli.StringFlag{
			EnvVar: "REPOSITORY_ORG",
			Name:   "repo.owner",
			Usage:  "repository owner",
		},
		cli.StringFlag{
			EnvVar: "REPOSITORY_NAME",
			Name:   "repo.name",
			Usage:  "repository name",
		},
		cli.StringFlag{
			EnvVar: "REPOSITORY_BRANCH",
			Name:   "repo.branch",
			Usage:  "repository default branch",
			Value:  "master",
		},
		cli.StringFlag{
			EnvVar: "REPOSITORY_BRANCH",
			Name:   "repo.commit.branch",
			Usage:  "git commit branch",
			Value:  "master",
		},
	}

	if err := app.Run(os.Args); err != nil {
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
			CaCert:              c.String("config.ca_cert"),
			CaCertPath:          c.String("config.ca_cert_path"),
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

	// validate the plugin configuration
	err = p.Validate()
	if err != nil {
		return err
	}

	// start the plugin execution
	err = p.Exec()
	if err != nil {
		return err
	}

	return nil
}
