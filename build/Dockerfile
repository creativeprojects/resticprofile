FROM alpine:latest

LABEL maintainer Fred <fred@gcreativeprojects.tech>

COPY build/restic /usr/bin/
COPY resticprofile /usr/bin/

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
