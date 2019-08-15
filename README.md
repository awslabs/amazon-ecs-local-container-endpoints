Amazon ECS Local Container Endpoints
====================================

A container that provides local versions of the [ECS Task IAM Roles endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html) and the [ECS Task Metadata Endpoints](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint.html). This project will help you test applications locally before you deploy to ECS/Fargate.

This repository contains the source code for the project. To use it, pull the [amazon/amazon-ecs-local-container-endpoints:latest image from Docker Hub](https://hub.docker.com/r/amazon/amazon-ecs-local-container-endpoints/).

#### Table of Contents
* [Tutorial](https://aws.amazon.com/blogs/compute/a-guide-to-locally-testing-containers-with-amazon-ecs-local-endpoints-and-docker-compose/)
* [Setup Networking](docs/setup-networking.md)
  * [Option 1: Use a User Defined Docker Bridge Network](docs/setup-networking.md#option-1-use-a-user-defined-docker-bridge-network-recommended)
  * [Option 2: Set up iptables rules](docs/setup-networking.md#option-2-set-up-iptables-rules)
* [Configuration](docs/configuration.md)
  * [Credentials](docs/configuration.md#credentials)
  * [Docker](docs/configuration.md#docker)
  * [Environment Variables](docs/configuration.md#environment-variables)
* [Features](docs/features.md)
  * [Vend Credentials to Containers](docs/features.md#vend-credentials-to-containers)
  * [Metadata](docs/features.md#metadata)
    * [Task Metadata V2](docs/features.md#task-metadata-v2)
    * [Task Metadata V3](docs/features.md#task-metadata-v3)

#### Security disclosures

If you think you’ve found a potential security issue, please do not post it in the Issues.  Instead, please follow the instructions [here](https://aws.amazon.com/security/vulnerability-reporting/) or email AWS security directly at [aws-security@amazon.com](mailto:aws-security@amazon.com).

#### License

This library is licensed under the Apache 2.0 License.
