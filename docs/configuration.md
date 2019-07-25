## Configuration

### Credentials

The ECS Local Endpoints container uses the AWS SDK for Go, and thus it supports all of its [methods of configuration](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html). We recommend providing credentials via an AWS CLI Profile. To do this, mount `$HOME/.aws/` ([`%UserProfile%\.aws` on Windows](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)) into the container. As shown in the example Compose file, the container path of the volume should be `/home/.aws/` because the environment variable `HOME` is set to `/home` in the image. This way, inside the container, the SDK will be able to find credentials at `$HOME/.aws/`. To use a non-default profile, set the `AWS_PROFILE` environment variable on the Local Endpoints container.

The Local Endpoints container will retrieve temporary session credentials from STS.  To provide a custom CA bundle for the STS client, mount your certificates file into the Local Endpoints container at any of the following locations:
* `/etc/ssl/certs/ca-certificates.crt`
* `/etc/pki/tls/certs/ca-bundle.crt`
* `/etc/ssl/ca-bundle.pem`
* `/etc/pki/tls/cacert.pem`
* `/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem`

For example, on an Ubuntu machine, you can mount your machine's certificates file at `/etc/ssl/certs/ca-certificates.crt` into the Local Endpoint container at `/etc/ssl/certs/ca-certificates.crt`.

### Docker

Local Endpoints responds to Metadata requests with real data about the containers running on your machine. In order to do this, you must mount the [Docker socket](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-socket-option) into the container. Make sure the Local Endpoints container is given a volume with source path `/var/run` and container path `/var/run`.

### Environment Variables

General Configuration:
* `ECS_LOCAL_METADATA_PORT` - Set the port that the container listens at. The default is `80`.

Task Metadata Configuration: while Local Endpoints returns real runtime information obtained from Docker in metadata requests, some values have no relevance locally and are mocked:
* `CLUSTER_ARN` - Set the 'cluster' name which is returned in Task Metadata responses. Default: `ecs-local-cluster`.
* `TASK_ARN` - Set ARN of the mock local 'task' which your containers will appear to be part of in Task Metadata responses. Default: `arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152`.
* `TASK_DEFINITION_FAMILY` - Set family name for the mock task definition which your containers will appear to be part of in Task Metadata responses. Default: `esc-local-task-definition`.
* `TASK_DEFINITION_REVISION` - Set the Task Definition revision. Default: `1`.
