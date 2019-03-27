# Changelog

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
