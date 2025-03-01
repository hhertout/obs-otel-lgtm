FROM golang:1.24.0-alpine

RUN apk add --no-cache git \
    build-base \
    bash \
    curl \
    jq

# Installer xk6
RUN go install go.k6.io/xk6/cmd/xk6@v0.9.2

# Construire k6 avec des versions spécifiques pour tous les plugins
RUN xk6 build \
    --output /usr/local/bin/k6 \
    --with github.com/grafana/xk6-browser \
    --with github.com/grafana/xk6-output-opentelemetry \
    --with github.com/wosp-io/xk6-playwright\
    --with github.com/szkiba/xk6-ansible-vault \
    --with github.com/grafana/xk6-dashboard \
    --with github.com/grafana/xk6-faker

WORKDIR /tests

COPY ./.k6 .

ENV K6_BROWSER_ENABLED=true
ENV K6_OTEL_METRIC_PREFIX=k6_
ENV K6_OTEL_SERVICE_NAME=k6
ENV OTEL_SERVICE_NAME=k6
ENV K6_BROWSER_HEADLESS=false

CMD ["tail", "-f", "/dev/null"]