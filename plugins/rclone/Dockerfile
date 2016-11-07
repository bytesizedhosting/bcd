FROM bytesized/debian-base
MAINTAINER DedSec 

ENV version=1.2.17

RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y wget
RUN apt-get install -y unzip
RUN apt-get install -y man

RUN curl -O http://downloads.rclone.org/rclone-current-linux-amd64.zip
RUN unzip rclone-current-linux-amd64.zip
RUN cd rclone-*-linux-amd64 && \
      cp rclone /usr/sbin/ && \
      chown root:root /usr/sbin/rclone && \
          chmod 755 /usr/sbin/rclone

VOLUME /config /media /data
COPY static /

EXPOSE 64738


