# Scheduling inside a container with a non-root user

You can schedule your backups with resticprofile by running `crond` inside a container.
This version shows how to use [supercronic](https://github.com/aptible/supercronic) to run the scheduled backups as a non-root user.

You can create a container with this modified version from the official image:

```Dockerfile
FROM alpine:latest

LABEL org.opencontainers.image.documentation="https://creativeprojects.github.io/resticprofile/"
LABEL org.opencontainers.image.source="https://github.com/creativeprojects/resticprofile"


ARG ARCH=amd64
ENV TZ=Etc/UTC

COPY build/restic-${ARCH} /usr/bin/restic
COPY build/rclone-${ARCH} /usr/bin/rclone
COPY resticprofile /usr/bin/resticprofile

RUN apk add --no-cache openssh-client-default curl tzdata ca-certificates supercronic && \
    chmod +x /usr/bin/restic /usr/bin/rclone /usr/bin/resticprofile && \
    adduser -D -h /resticprofile resticprofile && \
    mkdir -p /resticprofile && \
    touch /resticprofile/crontab && \
    chown -R resticprofile:resticprofile /resticprofile

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
```

Here's a `docker-compose` example loading configuration from a `.env` file:

```yaml
version: '2'

services:
  scheduled-backup:
    image: creativeprojects/resticprofile:${RP_VERSION:-latest}
    container_name: backup_container
    hostname: backup_container
    user: resticprofile:resticprofile
    entrypoint: '/bin/sh'
    command:
      - '-c'
      - 'resticprofile schedule --all && supercronic /resticprofile/crontab'
    volumes:
      - '${RP_CONFIG}:/resticprofile/profiles.yaml:ro'
      - '${RP_KEYFILE}:/resticprofile/key:ro'
      - '${BACKUP_SOURCE}:/source:ro'
      - '${RP_REPOSITORY}:/restic_repo'
    environment:
      - TZ=${TIMEZONE:-Etc/UTC}

```

with the corresponding resticprofile configuration running a backup every 15 minutes:

```yaml

global:
  scheduler: crontab:-:/resticprofile/crontab

default:
  password-file: key
  repository: "local:/restic_repo"
  initialize: true
  backup:
    source: /source
    exclude-caches: true
    one-file-system: true
    schedule: "*:00,05,10,15,20,25,30,35,40,45,50,55"
    schedule-permission: user
    check-before: true

```

## More information

[Discussion on Supersonic](https://github.com/creativeprojects/resticprofile/issues/288)

[Discussion on non-root container](https://github.com/creativeprojects/resticprofile/issues/321)

