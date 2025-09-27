---
title: "Docker"
weight: 30
---



## Using resticprofile from a docker image ##

You can run resticprofile in a Docker container, which is the easiest way to install and update resticprofile and restic simultaneously.

However, you must mount your backup source (and local destination, if applicable) as a Docker volume. On macOS, backups may be slower due to the known performance issues with mounted volumes.

### Registries

The official image is available on docker hub
```shell
docker pull creativeprojects/resticprofile:latest
```
as well as on Github Container Registry
```shell
docker pull ghcr.io/creativeprojects/resticprofile:latest
```

### Configuration

By default, the resticprofile container starts at `/resticprofile`. So you can feed a configuration this way:

```shell
docker run -it --rm -v $PWD/examples:/resticprofile ghcr.io/creativeprojects/resticprofile
```

You can list your profiles:
```shell
docker run -it --rm -v $PWD/examples:/resticprofile ghcr.io/creativeprojects/resticprofile profiles
```

### Container host name

Each time a container starts, it is assigned a random name.

To set a specific hostname, use the `-h` or `--hostname` flag with `docker run`:

```shell
docker run -it --rm -v $PWD:/resticprofile -h my-hostname ghcr.io/creativeprojects/resticprofile -n profile backup
```

### Platforms

The resticprofile docker image is available in these 2 platforms:
- linux/amd64
- linux/arm64/v8 (compatible with raspberry pi 64bits)

### rclone

The resticprofile docker image also includes the latest version of [rclone][1].

### Container imager release cycle

The Docker image is automatically uploaded to both registries when a new release is published on GitHub. The `latest` tag is updated to match the release.

Each week, the `latest` image is rebuilt to include updates from Restic, Rclone, and the Alpine base image.

After every commit to the main branch, another image is updated and tagged as `nightly`. This image may be unstable and is not recommended for production use.

## Scheduling with docker compose

There's an example in the contribution section how to schedule backups in a long running container.
The configuration needs to specify `crond` as a scheduler.

See [contrib][2]

[1]: https://rclone.org/
[2]: https://github.com/creativeprojects/resticprofile/tree/master/contrib/schedule-in-docker
