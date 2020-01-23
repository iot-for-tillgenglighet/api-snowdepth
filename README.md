# Introduction

This service is responsible for storing snowdepth telemetry and provide it to consumers via an API.

# Building and tagging with Docker

`docker build -f deployments/Dockerfile -t iot-for-tillgenglighet/api-snowdepth:latest .`

# Build for local testing with Docker Compose

`docker-compose -f ./deployments/docker-compose.yml build`

# Running locally with Docker Compose

`docker-compose -f ./deployments/docker-compose.yml up`

The ingress service will exit fatally and restart a couple of times until the RabbitMQ container is properly initialized and ready to accept connections. This is to be expected.

# Clean up the environment

`docker-compose -f ./deployments/docker-compose.yml down -v`

To clean up the environment properly after testing.
