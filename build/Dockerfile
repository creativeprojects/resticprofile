FROM alpine:latest

LABEL maintainer Fred <fred@gcreativeprojects.tech>

COPY restic* /usr/bin/

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
