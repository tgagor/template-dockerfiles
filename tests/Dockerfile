ARG BASEIMAGE="alpine"
FROM alpine:$BASEIMAGE

# set default file encoding to UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

# do not store cache or temp files in image
VOLUME /tmp /var/cache/apk /var/tmp /root/.cache

# configure timezone
ARG TIMEZONE="UTC"
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/$TIMEZONE /etc/localtime && \
    echo "$TIMEZONE" > /etc/timezone && \
    apk del tzdata && \
    date

# install dependencies/tools
# https://github.com/krallin/tini
RUN apk add --no-cache \
    tini

ENTRYPOINT ["/sbin/tini", "--"]
