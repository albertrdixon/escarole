FROM escarole_base
MAINTAINER Albert Dixon <albert@dixon.rocks>

RUN apk add --update \
      ca-certificates \
      py-html5lib \
      py-lxml \
      py-mako \
      py-openssl \
      py-pillow \
      python \
      unrar \
    && mkdir /data

ENV SB_DATA=/data
COPY escarole.yml /