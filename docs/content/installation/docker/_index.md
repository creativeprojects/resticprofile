---
title: "Docker"
date: 2022-04-23T23:58:56+01:00
weight: 30
---



## Using resticprofile from a docker image ##

You can run resticprofile inside a docker container. It is probably the easiest way to install resticprofile (and restic at the same time) and keep it updated.

**But** be aware that you will need to mount your backup source (and destination if it's local) as a docker volume.
Depending on your operating system, the backup might be **slower**. Volumes mounted on a mac OS host are well known for being quite slow.

By default, the resticprofile container starts at `/resticprofile`. So you can feed a configuration this way:

```
$ docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile
```

You can list your profiles:
```
$ docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile profiles
```

### Container host name

Each time a container is started, it gets assigned a new random name.

You can force a hostname
- in your container:
```
$ docker run -it --rm -v $PWD:/resticprofile -h my-machine creativeprojects/resticprofile -n profile backup
```
- in your configuration:

```toml
[profile]
host = "my-machine"
```
