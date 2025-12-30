---
title: "Docker"
weight: 15
---

Docker images are available at `creativeprojects/resticprofile:latest` and `ghcr.io/creativeprojects/resticprofile:latest`

To run an instance of resticprofile, run: 
```shell
docker run -it --rm -v $PWD/examples:/resticprofile ghcr.io/creativeprojects/resticprofile
```

{{% notice style="warning" %}}
- You must mount your backup source (and local destination, if applicable) as a Docker volume.
- On macOS, backups may be slower due to the known performance issues with mounted volumes.
{{% /notice %}}

Resticprofile's docker images also include restic and rclone.

{{< toc >}}

## Configuration

By default, the resticprofile container starts at `/resticprofile`. So you can feed a configuration this way:

```shell
docker run -it --rm -v $PWD/examples:/resticprofile ghcr.io/creativeprojects/resticprofile
```

## List profiles

```shell
docker run -it --rm -v $PWD/examples:/resticprofile ghcr.io/creativeprojects/resticprofile profiles
```

## Container host name

To set a specific hostname, use the `-h` or `--hostname` flag with `docker run`:

```shell
docker run -it --rm -v $PWD:/resticprofile -h my-hostname ghcr.io/creativeprojects/resticprofile -n profile backup
```
