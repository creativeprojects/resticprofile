# Contributing to resticprofile

Thanks for your interest in contributing! This document explains how to set up
your environment, build the project, and run the tests.

## Prerequisites

- **Go** — version `1.26` or later (see [`.github/workflows/env`](.github/workflows/env)
  for the exact version used by CI).
- **make** — all common tasks are wired through the [`Makefile`](Makefile).
- **git** — the repository uses a submodule for the documentation theme. Clone with
  submodules, or initialise them after cloning:

  ```shell
  git clone --recurse-submodules https://github.com/creativeprojects/resticprofile.git
  # or, if already cloned:
  git submodule update --init --recursive
  ```

Optional tooling (downloaded automatically into `$GOBIN` by the relevant make
targets, so you usually don't install these by hand):

- [`gotestsum`](https://github.com/gotestyourself/gotestsum) — test runner used by the `test*` targets
- [`mockery`](https://github.com/vektra/mockery) — regenerates mocks before tests
- [`golangci-lint`](https://github.com/golangci/golangci-lint) — linter (pinned version, see `.github/workflows/env`)
- [`restic`](https://restic.net/) — only needed if you want to run resticprofile against a real repository

Run `make help` at any time to list all available targets.

## Building

Build the binary for your current platform:

```shell
make build
```

This produces a `resticprofile` binary in the repository root. The target first
runs `prepare_build` (verifies your Go installation and downloads module
dependencies), then compiles with version metadata embedded via `-ldflags`.

Other build targets:

| Target                  | Description                                                |
| ----------------------- | ---------------------------------------------------------- |
| `make install`          | Build and install the binary into `$GOBIN`                 |
| `make build-no-selfupdate` | Build without the self-update feature (`no_self_update` tag) |
| `make build-mac`        | Cross-compile for macOS (amd64 + arm64)                    |
| `make build-linux`      | Cross-compile for Linux (amd64 + arm64)                    |
| `make build-windows`    | Cross-compile for Windows (amd64 + arm64)                  |
| `make build-all`        | Cross-compile for all of the above                         |

To remove build artifacts (binaries, coverage files, generated mocks, etc.):

```shell
make clean
```

## Running the tests

Run the full unit test suite:

```shell
make test
```

The `test` target automatically:

1. Installs `gotestsum` (into `$GOBIN`) if needed.
2. Runs `prepare_test`, which regenerates mocks with `mockery`.
3. Builds the test helper binaries under `testhelpers/` (`test-args`,
   `test-echo`, `test-crontab`, `test-shell`) and exposes their location via the
   `TEST_HELPERS` environment variable.
4. Runs the tests with `gotestsum`.

Useful variations:

| Target            | Description                                                   |
| ----------------- | ------------------------------------------------------------- |
| `make test-short` | Run tests in short mode (`-short`)                            |
| `make test-race`  | Run tests with the race detector (short mode)                 |
| `make test-ci`    | Run tests as CI does: race detector, `-short`, `fuse` build tag, coverage profile and JUnit report |
| `make coverage`   | Generate a coverage profile and open the HTML report          |

### Running a subset of tests

The test targets honour the `TESTS` variable (default `./...`). For example, to
run the tests of a single package:

```shell
make test TESTS=./config/...
```

You can also run the standard Go tooling directly, but remember to point
`TEST_HELPERS` at the directory containing the helper binaries built by
`make test-helpers`:

```shell
make test-helpers
TEST_HELPERS=$(pwd)/build/ go test ./config/...
```

### FUSE tests

Some tests are guarded behind the `fuse` build tag. They require FUSE
support on your machine.

### SSH client tests

The SSH client integration tests need a containerised SSH server and are run
separately:

```shell
make start-ssh-server   # spins up the SSH server via docker compose
make ssh-test           # runs the SSH client tests (ssh build tag)
make stop-ssh-server    # tears the server down and cleans up
```

These require Docker (with `docker compose`) and `ssh-keygen`.

## Linting

CI runs `golangci-lint`. To run it locally:

```shell
make lint        # lint for darwin, linux and windows build targets
make fix         # run go mod tidy, go fix, and golangci-lint --fix
```

The linter version is pinned in [`.github/workflows/env`](.github/workflows/env)
and the configuration lives in [`.golangci.yml`](.golangci.yml).

## Generated files

Some files are generated and should be regenerated when you change their
sources:

- **Mocks** are regenerated automatically by `make prepare_test` (used by the
  test targets), based on [`.mockery.yml`](.mockery.yml).
- **`go generate`** is run as part of `make test-ci`; you can also run
  `go generate ./...` directly.
- **JSON schema** and the **configuration reference** are generated from the
  built binary with `make generate-jsonschema` and
  `make generate-config-reference`. The documentation site is built with
  `make documentation` (requires Hugo and the docs submodule).

## Continuous integration

Pull requests are validated on Linux, macOS, Windows and several BSDs. The
Linux/macOS/Windows jobs run the shared workflow
[`.github/workflows/run-tests-os.yml`](.github/workflows/run-tests-os.yml),
which essentially performs:

```shell
make build
make test-ci
```

Before opening a pull request, it's a good idea to run at least:

```shell
make build
make lint
make test
```

## Submitting changes

1. Fork the repository and create a topic branch from `master`.
2. Make your changes, keeping the existing code style and adding tests where it
   makes sense.
3. Run `make build`, `make lint` and `make test` and make sure they pass.
4. Open a pull request describing your change.

Thanks for contributing! 🎉
