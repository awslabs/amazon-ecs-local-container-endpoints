## Amazon ECS Local Container Endpoints

A container that provides local versions of the [ECS Task IAM Roles endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html) and the [ECS Task Metadata Endpoints](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint.html). This project will help you test applications locally before you deploy to ECS/Fargate.

This repository contains the source code for the project. To use it, pull the [amazon/amazon-ecs-local-container-endpoints:latest image from Docker Hub](https://hub.docker.com/r/amazon/amazon-ecs-local-container-endpoints/).

## Setting Up Networking

ECS Local Container Endpoints supports 3 endpoints:
* The [ECS Task IAM Roles endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)
* The [Task Metadata V2 Endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html)
* The [Task Metadata V3 Endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html)

The Task Metadata V2 and Credentials endpoints require the Local Endpoints container to be able to receive requests made to the special IP Address, `169.254.170.2`.

There are two methods to achieve this.

#### Option 1: Use a User Defined Docker Bridge Network (Recommended)

If you launch containers into a custom [bridge network](https://docs.docker.com/network/bridge/), you can specify that the ECS Local Endpoints container will receive `169.254.170.2` as its IP address in the network. The endpoints will only be reachable inside this network, so all your containers must run inside of it. The [example Docker Compose file](examples/docker-compose.yml) in this repository shows how to create this network using Compose.

This method is the recommended way of using ECS Local Container Endpoints.

#### Option 2: Set up iptables rules

If you use Linux, then you can set up routing rules to forward requests for `169.254.170.2`. This is the option used in production ECS, as noted in the [documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html). The following commands must be run to set up routing rules:

```
sudo sysctl -w net.ipv4.conf.all.route_localnet=1
sudo iptables -t nat -A PREROUTING -p tcp -d 169.254.170.2 --dport 80 -j DNAT --to-destination 127.0.0.1:51679
sudo iptables -t nat -A OUTPUT -d 169.254.170.2 -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 51679
sudo iptables-save
```

These commands enable local routing, and create a rule to forward packets sent to `169.254.170.2:80` to `127.0.0.1:51679`.

Once you set up these rules, you can run the Local Endpoints container as follows:

```
docker run -d -p 51679:51679 \
-v /var/run:/var/run \
-v $HOME/.aws/:/home/.aws/ \
-e "ECS_LOCAL_METADATA_PORT=51679" \
--name ecs-local-endpoints \
amazon/amazon-ecs-local-container-endpoints:latest
```

## Configuration

### Credentials

The ECS Local Endpoints container uses the AWS SDK for Go, and thus it supports all of its [methods of configuration](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html). We recommend providing credentials via an AWS CLI Profile. To do this, mount `$HOME/.aws/` ([`%UserProfile%\.aws` on Windows](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)) into the container. As shown in the example Compose file, the container path of the volume should be `/home/.aws/` because the environment variable `HOME` is set to `/home` in the image. This way, inside the container, the SDK will be able to find credentials at `$HOME/.aws/`. To use a non-default profile, set the `AWS_PROFILE` environment variable on the Local Endpoints container.

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

## Features

### Vend Credentials to Containers

The AWS CLI, and all of the AWS SDKs, will look for the environment variable `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` as part of their [default credential provider chain](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default).

If the variable exists, then the SDKs will try to obtain credentials by making requests to `http://169.254.170.2$AWS_CONTAINER_CREDENTIALS_RELATIVE_URI`. The ECS Agent injects this environment variable into containers running on ECS, and responds to requests at the endpoint. This is how [IAM Roles for Tasks](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html) is implemented under the hood.

You can set AWS_CONTAINER_CREDENTIALS_RELATIVE_URI to two different values on your application container:
* `"/creds"` - With this value, Local Endpoints returns temporary credentials obtained by calling [sts:GetSessionToken](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html#stsapi_comparison). These credentials will have the same permissions as the base credentials given to the Local Endpoints container.
* `"/role/{role name}"` - With this value, your application container receives credentials obtained via assuming the given role name. This could be a Task IAM Role, or it could be any other IAM Role.

**Note:** *We do not recommend using production credentials or production roles when testing locally. Modifying the trust policy of a production role changes its security boundary. More importantly, using credentials with access to production when testing locally could lead to accidental changes in your production account. We recommend using a separate account for testing.*

If you use the second option, make sure your IAM Role contains the following trust policy:
```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "ARN of your IAM User/identity"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

You can obtain the ARN of your local IAM identity by running:
```
aws --profile default sts get-caller-identity
```

### Metadata

For both V2 and V3, Local Endpoints defines a local 'task' as all containers running in a single Docker Compose project. If your container is running outside of Compose, then all currently running containers on your machine will be considered to be part of one local 'task'.

#### Task Metadata V2

No additional configuration is needed beyond that which is mentioned in the [Configuration](#configuration) section.

#### Task Metadata V3

V3 Metadata uses the `ECS_CONTAINER_METADATA_URI` environment variable. Unlike V2 metadata and Credentials, the IP address does not have to be `169.254.170.2`. If you only use V3 metadata, then the Local Endpoints container could listen at any IP address. If you choose this option, replace the IP address in the following examples.

In most cases, you can set `ECS_CONTAINER_METADATA_URI` to `http://169.254.170.2/v3`.

However, in a few cases, this will not work. This is because the Local Endpoints container needs to be able to determine which container a request for V3 metadata came from. Local Endpoints attempts to use the IP address in the request to determine this. If you use the [example Docker Compose file](examples/docker-compose.yml) with a bridge network, then this IP lookup will work. However, if you use different network settings, then the Local Endpoints will not be able to determine which container a request came from. In this case, set `ECS_CONTAINER_METADATA_URI` to `http://169.254.170.2/v3/containers/{container name}`. The value for `container name` can be any unique substring of your container's name. By setting a custom request URL, the Local Endpoints container can determine which container a request came from.

## License

This library is licensed under the Apache 2.0 License.
