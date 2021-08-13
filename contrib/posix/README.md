# Default configuration for POSIX systems

**Layout for `/etc/resticprofile`**:

* `profiles.conf` - host centric default configuration
* `profiles.d/*` - host centric backup profiles (`*.toml` & `*.yaml`)
* `conf.d/*` - overrides & extra configuration

The layout is used in `deb`, `rpm` and `apk` packages of `resticprofile`

**Generated files**:
* `conf.d/default-repository.secret` - during installation, only if missing

**Referenced files and paths**:
* `conf.d/default-repository-self-signed-pub.pem` - TLS public cert (self-signed only)
* `conf.d/default-repository-client.pem` - TLS client cert
* `/var/lib/prometheus/node-exporter/resticprofile-*.prom` - Prometheus files
* `$TMPDIR/resticprofile-*` - Status and lock files

# Quick Start

## Installation

* RPM: `rpm -i "resticprofile-VERSION-ARCH.rpm"`
* DEB: `dpkg -i "resticprofile-VERSION-ARCH.deb"`

## Configuration
Setup repository and validate system backup profile:
```shell
cd /etc/resticprofile/
vim conf.d/repository.conf
vim profiles.d/system.toml
```

## Test config and backup
```shell
resticprofile -n root show
resticprofile -n root --dry-run backup
resticprofile -n root backup
```
