FROM alpine:latest

LABEL maintainer Fred <fred@gcreativeprojects.tech>

ARG ARCH=amd64

COPY build/restic-${ARCH} /usr/bin/restic
COPY resticprofile /usr/bin/resticprofile

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
