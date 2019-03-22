FROM golang:1.12

WORKDIR /go/src/github.com/awslabs/amazon-ecs-local-container-endpoints
COPY . .

CMD GO111MODULE=on go test -mod=vendor -timeout=120s -v -cover ./integ/... || { echo 'Integration Test Failure' ; exit 1; }
