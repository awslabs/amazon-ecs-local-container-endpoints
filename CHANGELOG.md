# Changelog


## 1.4.2
* Security - Update security patches (#228)

## 1.4.1
* Security - Update security patches (#194)

## 1.4.0
* Feature - Add support for ARM64 binaries and docker images (#59)

## 1.3.0
* Feature - Add support for V4 and generic metadata endpoints (#38)

## 1.2.0
* Feature - Add support for assuming roles in other accounts with path `/role-arn/{role arn}` (#36)

## 1.1.0
* Bug - Set expiration timestamp on temporary credentials (#26)
* Feature - Change base image to amazonlinux to support sourcing credentials from an external process (#30)
* Feature - Add support for custom endpoints for STS and IAM (#16)
* Enhancement - Print verbose error messages for credential chain problems (#25)

## 1.0.1
* Enhancement - Add custom user agent header for calls made to STS and IAM (#9)

## 1.0.0
* Feature - Support vending temporary credentials to containers from a base set of credentials
* Feature - Support vending temporary credentials to containers from an IAM Role
* Feature - [Support Task Metadata V2 Paths](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html):
	- Task Metadata - `/v2/metadata`
	- Container Metadata - `/v2/metadata/<container-id>`
	- Task Stats - `/v2/stats`
	- Container Stats - `/v2/stats/<container-id>`
* Feature - [Support Task Metadata V3 Paths](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html):
	- Container Metadata - `/v3` OR `/v3/containers/<container>`
	- Task Metadata - `/v3/task` OR `/v3/containers/<container>/task`
	- Container Stats - `/v3/stats` OR `/v3/containers/<container>/stats`
	- Task Stats - `/v3/task/stats` OR `/v3/containers/<container>/task/stats`
