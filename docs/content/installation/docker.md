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

For further reference, consult the [docker reference page](/reference/docker).
