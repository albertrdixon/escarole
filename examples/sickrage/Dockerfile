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
    && addgroup -g 7000 sickrage \
    && adduser -s /sbin/nologin -D -H -G sickrage -u 7000 sickrage \
    && mkdir /data \
    && chown -Rc sickrage:sickrage /data

ENV SB_DATA=/data
COPY escarole.yml /