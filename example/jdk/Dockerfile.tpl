FROM {{ if .registry  }}{{ .registry }}/{{ end }}{{ if .prefix }}{{ .prefix }}/{{ end }}base:alpine{{ .alpine }}

RUN apk add --no-cache \
        amazon-corretto-{{ .java }} && \
    java -version
