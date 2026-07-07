FROM alpine:latest

LABEL org.opencontainers.image.documentation="https://creativeprojects.github.io/resticprofile/"
LABEL org.opencontainers.image.source="https://github.com/creativeprojects/resticprofile"
LABEL org.opencontainers.image.title="resticprofile"

ARG TARGETARCH
ENV TZ=Etc/UTC

COPY build/restic-${TARGETARCH} /usr/bin/restic
COPY build/rclone-${TARGETARCH} /usr/bin/rclone
COPY resticprofile /usr/bin/resticprofile

RUN apk add --no-cache ca-certificates curl logrotate openssh-client-default tzdata

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
