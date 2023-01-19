FROM alpine:latest

LABEL maintainer Fred <fred@gcreativeprojects.tech>

ARG ARCH=amd64

RUN apk add --no-cache restic
RUN /usr/bin/restic self-update

RUN wget -O rclone.zip https://downloads.rclone.org/rclone-current-linux-${ARCH}.zip --no-check-certificate
RUN unzip rclone.zip
RUN cp rclone-*-linux-${ARCH}/rclone /usr/bin/rclone
RUN rm -r rclone-*-linux-${ARCH}

COPY resticprofile /usr/bin/resticprofile
#RUN wget -O install.sh https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
#RUN chmod +x install.sh
#RUN ./install.sh -b /usr/bin
#RUN rm ./install.sh

RUN apk add --no-cache openssh

VOLUME /resticprofile
WORKDIR /resticprofile

ENTRYPOINT ["resticprofile"]
CMD ["--help"]
