FROM --platform=$BUILDPLATFORM alpine:{{ .alpine }}

ARG TARGETOS
ARG TARGETARCH

# Install tini in specific architecture
# https://github.com/krallin/tini
ADD https://github.com/krallin/tini/releases/latest/download/tini-$TARGETARCH /usr/local/sbin/tini
RUN chmod +x /usr/local/sbin/tini

ENTRYPOINT ["/usr/local/sbin/tini", "--"]
