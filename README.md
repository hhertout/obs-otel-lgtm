# Observability docker compose

This repository contains a Docker Compose configuration to quickly set up a complete observability stack using Grafana. It includes all necessary components to monitor, visualize, and analyze your application's performance and logs. By following the instructions below, you can have a fully functional observability environment up and running in no time.

## Run it

Simply use

```bash
docker compose up -d
```

Then go to `http://localhost:3000` to connect on Grafana.

## Enable Logs

### Install the loki plugin for docker :

```bash
docker plugin install grafana/loki-docker-driver:3.3.2-arm64 --alias loki --grant-all-permissions
```

Then, add the following lines in your docker compose file:

```yaml
x-logging: &default-logging
  driver: loki
  options:
    loki-url: http://host.docker.internal:3100/loki/api/v1/push
    mode: non-blocking
    max-size: "10m"
    max-file: "3"
    loki-retries: "0"
    loki-timeout: "1s"
    keep-file: "true"
```

and add the loggin capability in your service:

```yaml
services:
  service1:
    # other configurations...
    logging: *default-logging

  service2:
    # other configurations...
    logging: *default-logging
```

## K6
