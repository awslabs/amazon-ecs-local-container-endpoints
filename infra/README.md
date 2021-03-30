# Continuous delivery pipelines

This package uses the [AWS Cloud Development Kit (CDK)](https://github.com/awslabs/aws-cdk) to model AWS CodePipeline
pipelines and to provision them with AWS CloudFormation.

* pipeline.ts: Builds and publishes the base Docker image for amazon/amazon-ecs-local-container-endpoints.

This creates a CodePipeline pipeline which consists of a source stage that uses a GitHub webhook, and build stages that
use AWS CodeBuild to build, publish and verify Docker images for both amd64 and arm64 architectures to DockerHub.

             +------------+       +----------------------------+       +-----------------------------+
             |   SOURCE   |       |           BUILD            |       |           VERIFY            |
             +------------+       +-------------+--------------+       +--------------+--------------+
             |            |       |   AMD64     |   ARM64      |       |   AMD64      |   ARM64      |
             |            |       +-------------+--------------+       +--------------+--------------+
             | GitHub     |       |docker build | docker build |       | docker pull  | docker pull  |
             | webhook    |+----->|             |              |+----->|              |              |
             |            |       | docker push | docker push  |       |verify endpts | verify endpts|
             |            |       |             |              |       |              |              |
             +------------+       +-------------+--------------+       +--------------+--------------+

## GitHub Access Token
To release using this pipeline, we use the release account for this repo
(ecs-local-container-endpoints+release@amazon.com).

The source stage requires a GitHub [personal access token](https://github.com/settings/tokens). For the official
release, this token is owned by the [ecs-cicd-bot account]( https://github.com/ecs-cicd-bot) and is stored in secrets
manager.

If you want to release a fork of this repo, create a GitHub access token with access to your fork of the repo, including
"admin:repo_hook" and "repo" permissions.  Then store the token in Secrets Manager:

```
aws secretsmanager create-secret --region us-west-2 --name EcsDevXGitHubToken --secret-string <my-github-personal-access-token>
```

## Deploy
Our current release treats the pipeline as a one-time build and push. Extending this pipeline to support full CI is
planned (TBD).  Deploying the pipeline stack will build and release the current version of this repo to DockerHub.

To deploy this pipeline, make sure you have the AWS CDK CLI installed: `npm i -g aws-cdk`

From the `infra` folder, install and build everything: `npm install && npm run build`

Using temporary credentials from the release account (ecs-local-container-endpoints+release@amazon.com), deploy the
pipeline stack:

*NOTE*: You may need to set the `CDK_DEFAULT_ACCOUNT` environment variable to the release account ID locally.

```
cdk deploy --app 'node pipeline.js'
```

See the pipeline in the CodePipeline console.

Once the pipeline run succeeds, destroy the stack:

```
cdk destroy --app 'node pipeline.js'
```

**NOTE**: When developing and testing, remember that any changes to `pipeline.ts` will require the stack to be re-build
with `npm run build` and redeployed with `cdk deploy --app 'node pipeline.js'`

