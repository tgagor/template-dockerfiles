FROM {{ if .registry  }}{{ .registry }}/{{ end }}{{ if .prefix }}{{ .prefix }}/{{ end }}jdk:{{ .java }}-alpine{{ .alpine }} AS builder

RUN apk add --no-cache \
        bash \
        binutils

COPY jre-build.sh /usr/local/sbin/
RUN jre-build.sh --output /jre

FROM {{ if .registry  }}{{ .registry }}/{{ end }}{{ if .prefix }}{{ .prefix }}/{{ end }}base:alpine{{ .alpine }}

RUN apk add --no-cache \
        java-common

ENV JAVA_HOME=/usr/lib/jvm/java-{{ .java }}-amazon-corretto-jre
COPY --from=builder /jre $JAVA_HOME

RUN ln -sf ${JAVA_HOME} /usr/lib/jvm/default-jvm && \
    java -version
