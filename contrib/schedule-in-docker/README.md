# Scheduling inside a docker container

[Discussion](https://github.com/creativeprojects/resticprofile/issues/74)

You can schedule your backups with resticprofile by running `crond` inside a container.

Here's a `docker-compose` example:

```yaml
version: '2'

services:
  scheduled-backup:
    image: creativeprojects/resticprofile:${RP_VERSION:-latest}
    container_name: backup_container
    hostname: backup_container
    entrypoint: '/bin/sh'
    command:
      - '-c'
      - 'resticprofile schedule --all && crond -f'
    volumes:
      - ~/.ssh:/root/.ssh
      - './profiles.yaml:/etc/resticprofile/profiles.yaml:ro'
      - './key:/etc/resticprofile/key:ro'
    environment:
      - TZ=Etc/UTC
```

with the corresponding resticprofile configuration running a backup every 15 minutes:

```yaml

global:
  scheduler: crond

default:
  password-file: key
  repository: sftp:storage_server:/tmp/backup
  initialize: true
  backup:
    source: /
    exclude-caches: true
    one-file-system: true
    schedule: "*:00,15,30,45"
    schedule-permission: system
    check-before: true

```
