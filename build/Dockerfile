FROM alpine:latest

LABEL org.opencontainers.image.documentation="https://creativeprojects.github.io/resticprofile/"
LABEL org.opencontainers.image.source="https://github.com/creativeprojects/resticprofile"


ARG ARCH=amd64
ENV TZ=Etc/UTC

COPY build/restic-${ARCH} /usr/bin/restic
COPY build/rclone-${ARCH} /usr/bin/rclone
COPY resticprofile /usr/bin/resticprofile

RUN apk add --no-cache ca-certificates curl openssh-client-default tzdata

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
