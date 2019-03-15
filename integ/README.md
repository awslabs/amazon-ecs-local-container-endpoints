### Setup for running integration tests

The Local Credentials Service Integration tests use your local `default` AWS Profile for base credentials.

Create an IAM role named `ecs-local-endpoints-integ-role`.
Attach the `AmazonS3ReadOnlyAccess` policy to it. (This includes the required `"s3:List*"` permissions).

The trust policy should be the following:

{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": <ARN of local IAM user signified by default AWS Profile>
      },
      "Action": "sts:AssumeRole"
    }
  ]
}

You can get the ARN of your local default AWS Profile with:
```
aws sts get-caller-identity
```
