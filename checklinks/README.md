# checklinks

`checklinks` is a small Go program that validates every link found in the
resticprofile source tree that starts with the documentation base URL.

## What it does

1. **Discovery** – walks all files under the current directory (skipping `.git`,
   `.github`, `.vscode`, `build`, `dist`, `public`, and `docs`) and extracts
   every URL that begins with the configured source base URL using a regular
   expression.
2. **Deduplication** – each unique URL is checked only once, and a small
   built-in exclusion list (e.g. JSON-schema endpoints) is skipped.
3. **Checking** – each URL is fetched with an HTTP GET request.  A response
   with a 2xx or 3xx status code is considered successful.
4. **Fragment validation** – when a URL contains a fragment (`#anchor`), the
   response body is scanned for `id="anchor"` so that broken anchors are
   caught as well.

## Modes of operation

### Local mode (default)

When no `-target` flag is provided, the tool starts an in-process TLS server
that serves the compiled HTML documentation from a local directory.  All
discovered links are rewritten to point at that local server so no network
access is required.

```
checklinks -dir ./public
```

### Remote mode

When a `-target` URL is supplied, all requests are redirected from the source
host to the target host at the TCP dial level.  This is useful for
smoke-testing a staging deployment without modifying source files.

```
checklinks -source https://creativeprojects.github.io/resticprofile \
           -target  https://staging.example.com
```

## Usage

```
checklinks [-v] [-source <url>] [-target <url>] [-dir <path>]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-v` | `false` | Print each link as it is discovered (verbose). |
| `-source` | `https://creativeprojects.github.io/resticprofile` | Base URL whose occurrences are searched for in files. |
| `-target` | *(empty)* | Alternative base URL to send requests to. When omitted, a local server is started and `-dir` is required. |
| `-dir` | *(empty)* | Directory containing the compiled HTML documentation. Required when `-target` is not given. |

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All links are valid. |
| `1` | One or more links could not be verified, or the tool encountered a fatal configuration error. |

## Building

```sh
go build ./checklinks
```

Or run directly without building:

```sh
go run ./checklinks -dir ./public
```
