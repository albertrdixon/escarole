FROM alpine:3.3
MAINTAINER Albert Dixon <albert@dixon.rocks>

ENTRYPOINT ["tini", "--", "/bin/escarole"]
RUN echo "http://dl-4.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories \
    && echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
    && apk add --update git tini
COPY bin/escarole-linux /bin/escarole