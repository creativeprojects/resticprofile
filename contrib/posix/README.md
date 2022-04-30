# Default configuration for POSIX systems

**Layout for `/etc/resticprofile`**:

* `conf.d/*` - default configuration and config overrides
* `profiles.conf` - main configuration file 
* `profiles.d/*` - host centric backup profiles (`*.toml` & `*.yaml`)
* `repositories.d/*` - restic repository configuration
* `templates/*` - reusable config blocks and system templates

The layout is used in `deb`, `rpm` and `apk` packages of `resticprofile`

**Generated files**:
* `repositories.d/default-repository.secret` - during installation, only if missing

**Referenced files and paths**:
* `repositories.d/default-repository-self-signed-pub.pem` - TLS public cert (self-signed only)
* `repositories.d/default-repository-client.pem` - TLS client cert
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
vim repositories.d/default.conf
vim profiles.d/system.toml
```

## Verify configuration, backup & restore
```shell
resticprofile root.show
resticprofile --dry-run root.backup
resticprofile root.backup
resticprofile root.snapshots
resticprofile root.mount /mnt/restore &
```

## Maintenance (check & prune)
```shell
resticprofile maintenance.check
resticprofile maintenance.prune
resticprofile maintenance.schedule
resticprofile maintenance.unschedule
```
