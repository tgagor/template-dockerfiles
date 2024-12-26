FROM alpine:{{ .alpine }}

RUN apk add --no-cache openjdk{{ .java }}

RUN curl -fsLo tomcat.tar.gz https://dlcdn.apache.org/tomcat/tomcat-{{ .tomcat | splitList "." | first }}/v{{ .tomcat }}/bin/apache-tomcat-{{ .tomcat }}.tar.gz && \
    tar xvf tomcat.tar.gz -C /opt/tomcat
