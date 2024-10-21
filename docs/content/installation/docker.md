---
title: "Docker"
# tags: ["v0.18.0"]
date: 2022-04-23T23:58:56+01:00
weight: 30
---



## Using resticprofile from a docker image ##

You can run resticprofile inside a docker container. It is probably the easiest way to install resticprofile (and restic at the same time) and keep it updated.

**But** be aware that you will need to mount your backup source (and destination if it's local) as a docker volume.
Depending on your operating system, the backup might be **slower**. Volumes mounted on a mac OS host are well known for being quite slow.

By default, the resticprofile container starts at `/resticprofile`. So you can feed a configuration this way:

```shell
docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile
```

You can list your profiles:
```shell
docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile profiles
```

### Container host name

Each time a container is started, it gets assigned a new random name.

You might want to force a hostname when starting your container via docker run (flags `-h` or `--hostname`):

```shell
docker run -it --rm -v $PWD:/resticprofile -h my-hostname creativeprojects/resticprofile -n profile backup
```

### Platforms

Starting from version `0.18.0`, the resticprofile docker image is available in these 2 platforms:
- linux/amd64
- linux/arm64/v8 (compatible with raspberry pi 64bits)

### rclone

Starting from version `0.18.0`, the resticprofile docker image also includes [rclone][1].

### Release cycle on docker hub

The docker image is uploaded automatically on docker hub when a new release is published on github.
Also, the `latest` tag is updated to the latest release.

Every week, the `latest` image is rebuilt to include the latest updates from restic, rclone and the base image (alpine).

Another image is also updated after every commit on the main branch. It's tagged with `nightly`.
Please be aware that this image might be unstable and should not be used in production.

## Scheduling with docker compose

There's an example in the contribution section how to schedule backups in a long running container.
The configuration needs to specify `crond` as a scheduler.

See [contrib][2]

[1]: https://rclone.org/
[2]: https://github.com/creativeprojects/resticprofile/tree/master/contrib/schedule-in-docker
