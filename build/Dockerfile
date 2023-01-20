FROM alpine:latest

LABEL maintainer Fred <fred@gcreativeprojects.tech>

ARG ARCH=amd64

COPY build/restic-${ARCH} /usr/bin/restic
COPY build/rclone-${ARCH} /usr/bin/rclone
COPY resticprofile /usr/bin/resticprofile

RUN apk add --no-cache openssh

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
