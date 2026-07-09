---
title: "Hooks - HTTP"
weight: 22
tags: [ "monitoring", "healthchecks.io" ]
---

{{< toc >}}

## Send HTTP messages before and after a job

As well as being able to run [shell commands]({{% relref "run_hooks" %}}), you can now send HTTP messages before, after (success or failure) running a restic command.

The sections that allow sending HTTP hooks are:
- backup
- copy
- check
- forget
- prune

{{% notice tip %}}
You might notice that's the same sections that can also be scheduled
{{% /notice %}}

Each of these commands can send 4 different types of hooks:

- send-before
- send-after
- send-after-fail
- send-finally

The configuration is the same for each of these 4 types of hooks:

| Name | Required | Default | Notes |
|:-----|:---------|:--------|:------|
| url | Yes | None | URL of your Webhook |
| method | No | GET | This is the HTTP method (GET, POST, HEAD, etc.) |
| skip-tls-verification | No | False | **This is not recommended**: Use only if you're using your own server with a self-signed certificate |
| headers | No | User-Agent set to resticprofile | This is a subsection with a list of `name` and `value` |
| body | No | Empty | Used to send data to the Webhook (POST, PUT, PATCH) |
| body-template | No | None | Template file to generate the body (in go template format) |


A few environment variables will be available to construct the url and the body:
- `PROFILE_NAME`
- `PROFILE_COMMAND`: backup, check, forget, etc.

Additionally, for the `send-after-fail` hooks, these environment variables will be available:
- `ERROR` containing the latest error message
- `ERROR_COMMANDLINE` containing the command line that failed
- `ERROR_EXIT_CODE` containing the exit code of the command line that failed
- `ERROR_STDERR` containing any message that the failed command sent to the standard error (stderr)

URL encoding is applayed for variables `ERROR`, `ERROR_COMMANDLINE` and `ERROR_STDERR` if they are used in URL.

The `send-finally` hooks are also getting the environment of `send-after-fail` when any previous operation has failed (except any `send` operation).

Failures in any `send-*` are logged but do not influence environment or return code.

## body-template

Templates pull from the standard go template to build the webhook body. The object passed as an argument to the template is:

- `ProfileName`    **string**
- `ProfileCommand` **string**
- `Error`          **ErrorContext**
- `Stdout`         **string**

The type **ErrorContext** is available after an error occurred (otherwise all fields are blank):
- `Message`     **string**
- `CommandLine` **string**
- `ExitCode`    **string**
- `Stderr`      **string**

## Self-signed certificates

If your monitoring system is using self-signed certificates, you can import them in resticprofile (and you don't need to rely on the `skip-tls-verification` flag)

The parameter is in the `global` section and is called `ca-certificates`: it contains a list of certificate files (PEM).

### timeout

The default timeout for all HTTP requests is 30 seconds.

You can change the default timeout in the `global` section with a parameter called `send-timeout`.

The format is like:
- 30s
- 2m
- 1m20s

### global configuration example


{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  send-timeout = "10s"
  ca-certificates = [ "ca-chain.cert.pem" ]
```


{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  send-timeout: 10s
  ca-certificates:
    - ca-chain.cert.pem

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl

global {
  send-timeout = "10s"
  ca-certificates = "ca-chain.cert.pem"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "global": {
    "send-timeout": "10s",
    "ca-certificates": [
      "ca-chain.cert.pem"
    ]
  }
}
```

{{% /tab %}}
{{< /tabs >}}

## order of `send-*`

Here's the flow of HTTP hooks:

```mermaid
graph TD
  LOCK(set resticprofile lock)
  UNLOCK(delete resticprofile lock)
  LOCK --> SB
  SB('send-before') --> RUN
  RUN(run restic command, or group of commands)
  RUN -->|Success| SA
  RUN -->|Error| SAF
  SA('send-after') --> SF
  SAF('send-after-fail') --> SF
  SF('send-finally')
  SF --> UNLOCK
```

{{% notice style="warning" title="resticprofile lock" %}}
The local resticprofile lock is surrounding the whole process. It means that the `run-after-fail` target is not called if the lock cannot be obtained. This is a limitation of the current implementation. 
{{% /notice %}}
