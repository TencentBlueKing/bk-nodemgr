FROM alpine:3.21

# install dependencies.
RUN apk add jq vim busybox-extras redis

# install mongodb with old package cause it doesn’t exist for Alpine after version 3.9.
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/community' >> /etc/apk/repositories
RUN apk add mongodb && rm -f /usr/bin/mongod /usr/bin/mongos

RUN mkdir -p \
    /bk-nodeman/bin \
    /bk-nodeman/etc \
    /bk-nodeman/cert \
    /bk-nodeman/log \
    /bk-nodeman/static \
    /bk-nodeman/file/tools \
    /bk-nodeman/file/cache \
    /bk-nodeman/file/local \

COPY bk-nodeman-backend /bk-nodeman/bin/
COPY bk-nodeman-application /bk-nodeman/bin/
COPY bk-nodeman-file /bk-nodeman/bin/
COPY tools/ /bk-nodeman/file/tools
COPY dist/ /bk-nodeman/static/

WORKDIR /bk-nodeman/bin