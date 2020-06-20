## Features

### Vend Credentials to Containers

The AWS CLI, and all of the AWS SDKs, will look for the environment variable `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` as part of their [default credential provider chain](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default).

If the variable exists, then the SDKs will try to obtain credentials by making requests to `http://169.254.170.2$AWS_CONTAINER_CREDENTIALS_RELATIVE_URI`. The ECS Agent injects this environment variable into containers running on ECS, and responds to requests at the endpoint. This is how [IAM Roles for Tasks](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html) is implemented under the hood.

You can set AWS_CONTAINER_CREDENTIALS_RELATIVE_URI to one of three different values on your application container:
* `"/creds"` - With this value, Local Endpoints returns temporary credentials obtained by calling [sts:GetSessionToken](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html#stsapi_comparison). These credentials will have the same permissions as the base credentials given to the Local Endpoints container, with a few exceptions. **The returned credentials will not be able to access the IAM APIs or the STS APIs**, except for sts:AssumeRole and sts:GetCallerIdentity.
* `"/role/{role name}"` - With this value, your application container receives credentials obtained via assuming the given role name. This could be a Task IAM Role, or it could be any other IAM Role. The role must exist in the same AWS account as for your default credentials.
* `"/role-arn/{role arn}"` - With this value, your application container receives credentials obtained via assuming the given role arn. This could be a Task IAM Role, or it could be any other IAM Role. Use this format when the role exists in a different AWS account to your default credentials.

**Note:** *We do not recommend using production credentials or production roles when testing locally. Modifying the trust policy of a production role changes its security boundary. More importantly, using credentials with access to production when testing locally could lead to accidental changes in your production account. We recommend using a separate account for testing.*

If you use the second or third options, make sure your IAM Role contains the following trust policy:
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

However, in a few cases, this will not work. This is because the Local Endpoints container needs to be able to determine which container a request for V3 metadata came from. Local Endpoints attempts to use the IP address in the request to determine this. If you use the [example Docker Compose file](../examples/docker-compose.yml) with a bridge network, then this IP lookup will work. However, if you use different network settings, then the Local Endpoints will not be able to determine which container a request came from. In this case, set `ECS_CONTAINER_METADATA_URI` to `http://169.254.170.2/v3/containers/{container name}`. The value for `container name` can be any unique substring of your container's name. By setting a custom request URL, the Local Endpoints container can determine which container a request came from.

#### Task Metadata V4

V4 Metadata uses `ECS_CONTAINER_METADATA_URI_V4` environment variable. In most cases, you can set `ECS_CONTAINER_METADATA_URI_V4` to `http://169.254.170.2/v3`. Similarly, in a few cases when Local Endpoints container doesn't know which container a request came from, you have to set `ECS_CONTAINER_METADATA_URI_V4` to `http://169.254.170.2/v3/containers/{container name}`. Please see the description above on V3 to understand why you'll need to specify the container name in the path.

However, compared to V3, V4 includes additional network metadata when querying the task metadata endpoint (see [here](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4.html)). Please refer to this [example](../examples/v4) if you want to include those additional V4 metadata. You can use the generic metadata injection feature (described below) to add the additional metadata fields included in V4.

#### Generic Metadata Injection

As mentioned above in the previous section, to inject generic metadata, you'll need to have those additional metadata in JSON files. Then specify paths for the JSON files by using `CONTAINER_METADATA_PATH` and `TASK_METADATA_PATH` environment variables. More specifically, `CONTAINER_METADATA_PATH` is the extra metadata for each container, which will override their counterparts in the normal response. Also, `TASK_METADATA_PATH` is for task level metadata, which is used for top level fields describing the task that are not in the containers list.
