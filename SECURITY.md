# Security Policy

## Supported versions

Security fixes are released for the latest version only. There are no backports
to older releases, so please check that the issue still exists in the
[latest release](https://github.com/creativeprojects/resticprofile/releases/latest)
before reporting it.

## Reporting a vulnerability

**Please do not report security vulnerabilities in public issues, discussions
or pull requests.**

Use GitHub's private vulnerability reporting instead:

1. Go to the [Security tab](https://github.com/creativeprojects/resticprofile/security)
   of the repository.
2. Click **Report a vulnerability**, or open the
   [new advisory form](https://github.com/creativeprojects/resticprofile/security/advisories/new)
   directly.

The report stays private between you and the maintainers until a fix is
published.

A useful report includes:

- the resticprofile version (`resticprofile version`), the operating system and
  how it was installed
- the relevant part of the configuration, with secrets removed
- the steps to reproduce the issue
- what an attacker could achieve, and what access they need to begin with

## What to expect

resticprofile is maintained by volunteers, so response times are best effort.
You should normally get a first answer within a week.

If the report is confirmed, a fix is prepared privately and published in a new
release together with a GitHub security advisory. You will be credited in the
advisory unless you prefer to stay anonymous.

## Scope

In scope:

- the resticprofile source code in this repository
- the released binaries, the Docker images and the `install.sh` script
- the self-update mechanism

Out of scope:

- **Vulnerabilities in restic itself.** Please report them to the
  [restic project](https://github.com/restic/restic/security).
- **Commands run from the configuration.** resticprofile runs the shell
  commands and hooks defined in its configuration file, with the privileges of
  the user running it. This is by design: anyone who can write to the
  configuration file can run commands as that user. Protect your configuration
  files accordingly, especially for profiles that are scheduled to run as root.
- Vulnerabilities in third-party dependencies that cannot be exploited through
  resticprofile. A regular issue or pull request to upgrade the dependency is
  welcome.

## Verifying a release

Each release publishes a `checksums.txt` file with the SHA-256 checksum of
every archive. The self-update command verifies the downloaded archive against
this file before replacing the binary.

To verify a manual download:

```shell
sha256sum --check --ignore-missing checksums.txt
```
