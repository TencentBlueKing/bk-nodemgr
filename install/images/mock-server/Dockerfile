FROM alpine:3.21

# set timezone.
ENV TZ Asia/Shanghai

RUN apk add --no-cache tzdata

RUN mkdir -p \
    /mock-server/bin \
    /mock-server/etc \
    /mock-server/log \
    /mock-server/data

COPY mock-server /mock-server/bin/

WORKDIR /mock-server/bin
