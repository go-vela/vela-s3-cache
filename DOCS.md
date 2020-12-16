## Description

This plugin enables you to cache build resources in an [s3](https://aws.amazon.com/s3/) compatible store for a Vela pipeline.

Source Code: https://github.com/go-vela/vela-s3-cache

Registry: https://hub.docker.com/r/target/vela-s3-cache

## Usage

> **NOTE:**
>
> Users should refrain from using latest as the tag for the Docker image.
>
> It is recommended to use a semantically versioned tag instead.

Sample of restoring a cache:

```yaml
steps:
  - name: restore_cache
    image: target/vela-s3-cache:latest
    pull: always
    parameters:
      action: restore
      bucket: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
```

Sample of rebuilding a cache:

```yaml
steps:
  - name: rebuild_cache
    image: target/vela-s3-cache:latest
    pull: always
    parameters:
      action: rebuild
      bucket: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
      mount:
        - .gradle
```

Sample of flushing a cache:

```yaml
steps:
  - name: flushing_cache
    image: target/vela-s3-cache:latest
    pull: always
    parameters:
      action: flush
      bucket: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
```

## Secrets

> **NOTE:** Users should refrain from configuring sensitive information in your pipeline in plain text.

### Internal

Users can use [Vela internal secrets](https://go-vela.github.io/docs/concepts/pipeline/secrets/) to substitute these sensitive values at runtime:

```diff
steps:
  - name: restore_cache
    image: target/vela-s3-cache:latest
    pull: always
+   secrets: [ s3_cache_access_key, s3_cache_secret_key ]
    parameters:
      action: restore
      bucket: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
-     access_key: AKIAIOSFODNN7EXAMPLE
-     secret_key: 123456789QWERTYEXAMPLE
```

> This example will add the secrets to the `restore_cache` step as environment variables:
>
> * `S3_CACHE_ACCESS_KEY=<value>`
> * `S3_CACHE_SECRET_KEY=<value>`

### External

The plugin accepts the following files for authentication:

| Parameter       | Volume Configuration                                                              |
| --------------- | --------------------------------------------------------------------------------- |
| `access_key`    | `/vela/parameters/s3-cache/access_key`, `/vela/secrets/s3-cache/access_key`       |
| `secret_key`    | `/vela/parameters/s3-cache/secret_key`, `/vela/secrets/s3-cache/secret_key`       |
| `session_token` | `/vela/parameters/s3-cache/session_token`, `/vela/secrets/s3-cache/session_token` |

Users can use [Vela external secrets](https://go-vela.github.io/docs/concepts/pipeline/secrets/origin/) to substitute these sensitive values at runtime:

```diff
steps:
  - name: restore_cache
    image: target/vela-s3-cache:latest
    pull: always
+   secrets: [ s3_cache_access_key, s3_cache_secret_key ]
    parameters:
      action: restore
      bucket: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
-     access_key: AKIAIOSFODNN7EXAMPLE
-     secret_key: 123456789QWERTYEXAMPLE
```

> This example will read the secret values in the volume stored at `/vela/secrets/`

## Parameters

> **NOTE:**
>
> The plugin supports reading all parameters via environment variables or files.
>
> Any values set from a file take precedence over values set from the environment.
>
> The s3 bucket set with the `bucket` parameter is expected to be created beforehand.

The following parameters can used to configure all image actions:

| Name                   | Description                                 | Required | Default         | Environment Variables                                                        |
| ---------------------- | ------------------------------------------- | -------- | --------------- | ---------------------------------------------------------------------------- |
| `accelerated_endpoint` | s3 accelerated instance to communicate with | `false`  | `N/A`           | `PARAMETER_ACCELERATED_ENDPOINT`<br>`S3_CACHE_ACCELERATED_ENDPOINT`          |
| `access_key`           | access key for communication with s3        | `true`   | `N/A`           | `PARAMETER_ACCESS_KEY`<br>`S3_CACHE_ACCESS_KEY`<br>`AWS_ACCESS_KEY_ID`       |
| `action`               | action to perform against s3                | `true`   | `N/A`           | `PARAMETER_ACTION`<br>`S3_CACHE_ACTION`                                      |
| `build_branch`         | branch name from build for the repository   | `false`  | **set by Vela** | `PARAMETER_BUILD_BRANCH`<br>`VELA_BUILD_BRANCH`                              |
| `bucket`               | name of the s3 bucket                       | `true`   | `N/A`           | `PARAMETER_BUCKET`<br>`S3_CACHE_BUCKET`                                      |
| `log_level`            | set the log level for the plugin            | `false`  | `info`          | `PARAMETER_LOG_LEVEL`<br>`S3_CACHE_LOG_LEVEL`                                |
| `org`                  | name of the org for the repository          | `true`   | **set by Vela** | `PARAMETER_ORG`<br>`VELA_REPO_ORG`                                           |
| `path`                 | custom path for the object(s)               | `false`  | `N/A`           | `PARAMETER_PATH`<br>`S3_CACHE_PATH`                                          |
| `prefix`               | path prefix for the object(s)               | `false`  | `N/A`           | `PARAMETER_PREFIX`<br>`S3_CACHE_PREFIX`                                      |
| `repo`                 | name of the repository                      | `true`   | **set by Vela** | `PARAMETER_REPO`<br>`VELA_REPO_NAME`                                         |
| `repo_branch`          | default branch for the Vela repository      | `false`  | **set by Vela** | `PARAMETER_REPO_BRANCH`<br>`VELA_REPO_BRANCH`                                |
| `secret_key`           | secret key for communication with s3        | `true`   | `N/A`           | `PARAMETER_SECRET_KEY`<br>`S3_CACHE_SECRET_KEY`<br>`AWS_SECRET_ACCESS_KEY`   |
| `server`               | s3 instance to communicate with             | `true`   | `N/A`           | `PARAMETER_SERVER`<br>`S3_CACHE_SERVER`                                      |
| `session_token`        | session token for communication with s3     | `true`   | `N/A`           | `PARAMETER_SESSION_TOKEN`<br>`S3_CACHE_SESSION_TOKEN`<br>`AWS_SESSION_TOKEN` |

### Restore

The following parameters are used to configure the `restore` action:

| Name       | Description                                                | Required | Default       | Environment Variables                       |
| ---------- | ---------------------------------------------------------- | -------- | ------------- | ------------------------------------------- |
| `filename` | the name of the cache object                               | `true`   | `archive.tgz` | `PARAMETER_FILENAME`<br>`S3_CACHE_FILENAME` |
| `mount`    | the file or directories locations to build your cache from | `true`   | `N/A`         | `PARAMETER_MOUNT`<br>`S3_CACHE_MOUNT`       |
| `timeout`  | the timeout for the call to s3                             | `false`  | `10m`         | `PARAMETER_TIMEOUT`<br>`S3_CACHE_TIMEOUT`   |

### Rebuild

The following parameters are used to configure the `rebuild` action:

| Name       | Description                    | Required | Default       | Environment Variables                       |
| ---------- | ------------------------------ | -------- | ------------- | ------------------------------------------- |
| `filename` | the name of the cache object   | `true`   | `archive.tgz` | `PARAMETER_FILENAME`<br>`S3_CACHE_FILENAME` |
| `timeout`  | the timeout for the call to s3 | `false`  | `10m`         | `PARAMETER_TIMEOUT`<br>`S3_CACHE_TIMEOUT`   |

### Flush

The following parameters are used to configure the `flush` action:

| Name  | Description                                             | Required | Default | Environment Variables             |
| ----- | ------------------------------------------------------- | -------- | ------- | --------------------------------- |
| `age` | delete the objects past a specific age (i.e. 60m, 8h)   | `false`  | `336h`  | `PARAMETER_AGE`<br>`S3_CACHE_AGE` |

## Template

COMING SOON!

## Troubleshooting

You can start troubleshooting this plugin by tuning the level of logs being displayed:

```diff
steps:
  - name: restore_cache
    image: target/vela-s3-cache:latest
    pull: always
    parameters:
      action: restore
      bucket: mybucket
+     log_level: trace
      server: mybucket.s3-us-west-2.amazonaws.com
```

Below are a list of common problems and how to solve them:
