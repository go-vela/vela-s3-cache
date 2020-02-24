## Description

This plugin enables you to cache build resources in an [s3](https://aws.amazon.com/free/storage/?sc_channel=PS&sc_campaign=acquisition_US&sc_publisher=google&sc_medium=ACQ-P%7CPS-GO%7CBrand%7CDesktop%7CSU%7CStorage%7CS3%7CUS%7CEN%7CText&sc_content=s3_e&sc_detail=aws%20s3&sc_category=Storage&sc_segment=293617570035&sc_matchtype=e&sc_country=US&s_kwcid=AL!4422!3!293617570035!e!!g!!aws%20s3&ef_id=EAIaIQobChMIzvj50J_q5wIVUPDACh12ugBiEAAYASAAEgLHwvD_BwE:G:s) compatible store for a Vela pipeline.

Source Code: https://github.com/go-vela/vela-s3-cache

Registry: https://hub.docker.com/r/target/vela-s3-cache

## Usage

Sample of restoring a cache:

```yaml
steps:
  - name: restore_cache
    image: target/vela-s3-cache:v0.1.0
    pull: true
    parameters:
      action: restore
      root: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
```

Sample of rebuilding a cache:

```yaml
steps:
  - name: rebuild_cache
    image: target/vela-s3-cache:v0.1.0
    pull: true
    parameters:
      action: rebuild
      root: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
      mount:
        - .gradle
```

Sample of flushing a cache:

```yaml
steps:
  - name: flushing_cache
    image: target/vela-s3-cache:v0.1.0
    pull: true
    parameters:
      action: flush
      root: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
```

## Secrets

**NOTE: Users should refrain from configuring sensitive information in your pipeline in plain text.**

You can use Vela secrets to substitute sensitive values at runtime:

```diff
steps:
  - name: restore_cache
    image: target/vela-s3-cache:v0.1.0
    pull: true
+   secrets: [ cache_access_key, cache_secret_key ]
    parameters:
      action: restore
      root: mybucket
      server: mybucket.s3-us-west-2.amazonaws.com
-     access_key: AKIAIOSFODNN7EXAMPLE
-     secret_key: 123456789QWERTYEXAMPLE
```

## Parameters

The following parameters are used to configure the image:

| Name        | Description                          | Required | Default |
| ----------- | ------------------------------------ | -------- | ------- |
| `action`    | action to perform against s3         | `true`   | `N/A`   |
| `log_level` | set the log level for the plugin     | `true`   | `info`  |
| `server`    | s3 instance to communicate with      | `true`   | `N/A`   |
| `access_key`| access key for communication with s3 | `true`   | `N/A`   |
| `secret_key`| secret key for communication with s3 | `true`   | `N/A`   |
| `root`      | name of the s3 bucket                | `true`   | `N/A`   |
| `prefix`    | path prefix for the object(s)        | `true`   | `N/A`   |

#### Restore

The following parameters are used to configure the `restore` action:

| Name       | Description                                                | Required | Default |
| ---------- | ---------------------------------------------------------- | -------- | ------  |
| `filename` | the name of the cache object                               | `false`  | `true`  |
| `mount`    | the file or directories locations to build your cache from | `true`   | `N/A`   |
| `timeout`  | the timeout for the call to s3                             | `false`  | `true`  |


#### Rebuild

The following parameters are used to configure the `rebuild` action:

| Name       | Description                    | Required | Default |
| ---------- | ------------------------------ | -------- | ------  |
| `filename` | the name of the cache object   | `false`  | `true`  |
| `timeout`  | the timeout for the call to s3 | `false`  | `true`  |

#### Flush

The following parameters are used to configure the `flush` action:

| Name    | Description                          | Required | Default |
| ------- | ------------------------------------ | -------- | ------- |

#### Upload

## Template

COMING SOON!

## Troubleshooting

Below are a list of common problems and how to solve them:
