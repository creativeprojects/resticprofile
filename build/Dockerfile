FROM alpine:latest

LABEL maintainer Fred <fred@gcreativeprojects.tech>

ARG ARCH=amd64
ENV TZ=Etc/UTC

COPY build/restic-${ARCH} /usr/bin/restic
COPY build/rclone-${ARCH} /usr/bin/rclone
COPY resticprofile /usr/bin/resticprofile

RUN apk add --no-cache openssh-client-default curl tzdata ca-certificates busybox-suid && \
    chmod +x /usr/bin/restic /usr/bin/rclone /usr/bin/resticprofile && \
    mkdir -p /resticprofile && \
    adduser -D -h /resticprofile resticprofile && \
    chown -R resticprofile:resticprofile /resticprofile

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
