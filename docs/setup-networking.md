## Setting Up Networking

ECS Local Container Endpoints supports 3 endpoints:
* The [ECS Task IAM Roles endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)
* The [Task Metadata V2 Endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v2.html)
* The [Task Metadata V3 Endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html)

The Task Metadata V2 and Credentials endpoints require the Local Endpoints container to be able to receive requests made to the special IP Address, `169.254.170.2`.

There are two methods to achieve this.

#### Option 1: Use a User Defined Docker Bridge Network (Recommended)

If you launch containers into a custom [bridge network](https://docs.docker.com/network/bridge/), you can specify that the ECS Local Endpoints container will receive `169.254.170.2` as its IP address in the network. The endpoints will only be reachable inside this network, so all your containers must run inside of it. The [example Docker Compose file](../examples/docker-compose.yml) in this repository shows how to create this network using Compose.

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
