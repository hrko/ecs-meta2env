This tool has been replaced by a re-implemented version in Rust.

https://github.com/hrko/ecs-meta2env-rs

# ecs-meta2env

`ecs-meta2env` is a tool designed to export values from ECS container metadata endpoints to environment variables. This is particularly useful for passing ECS metadata to applications like Fluent Bit to add metadata to logs.

## Download

You can download the latest release from the [releases page](https://github.com/hrko/ecs-meta2env/releases/latest).

## Usage

`ecs-meta2env` is intended to be used as an entrypoint for a container. It fetches metadata from the ECS container metadata endpoint and exports the values to environment variables.

### Example Dockerfile

Below is an example of how to use `ecs-meta2env` in a Dockerfile:

```Dockerfile
FROM debian:bookworm-slim AS ecs-meta2env-downloader
RUN apt-get update && apt-get install -y curl
RUN if [ "$(uname -m)" = "x86_64" ]; then ARCH="amd64"; else ARCH="arm64"; fi && \
    curl -L -o /usr/local/bin/ecs-meta2env https://github.com/hrko/ecs-meta2env/releases/download/v1.0.0/ecs-meta2env-linux-$ARCH && \
    chmod +x /usr/local/bin/ecs-meta2env

FROM <original-image>
COPY --from=ecs-meta2env-downloader /usr/local/bin/ecs-meta2env /usr/local/bin/ecs-meta2env
ENTRYPOINT ["/usr/local/bin/ecs-meta2env", "<original-entrypoint...>"]
```

## Environment Variables

`ecs-meta2env` will export the following environment variables:

* `X_ECS_CLUSTER`
* `X_ECS_TASK_ARN`
* `X_ECS_FAMILY`
* `X_ECS_REVISION`
* `X_ECS_SERVICE_NAME`
* `X_ECS_CONTAINER_NAME`
* `X_ECS_CONTAINER_DOCKER_NAME`
* `X_ECS_CONTAINER_ARN`

## Development

### Prerequisites

* Go 1.23.1 or later
* Task (taskfile.dev)
* Dev Container (optional)

### Building

To build the project, run the following command:

```sh
task build
```

This will create the binaries in the `./bin` directory.

### Testing

To run the tests, run the following command:

```sh
go test
```

## References

* [Original Idea](https://github.com/aws/aws-for-fluent-bit/issues/62#issuecomment-925702432): The idea of inserting a shell script at the entry point is suggested here, but `ecs-meta2env` was created to achieve the same for containers with no shell or limited built-in commands, such as the *distroless* container.
* [Amazon ECS task metadata endpoint version 4](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4.html)