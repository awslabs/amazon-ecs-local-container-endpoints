FROM golang:1.17

WORKDIR /go/src/github.com/awslabs/amazon-ecs-local-container-endpoints

COPY go.mod go.sum ./
ARG GOPROXY=direct
RUN go mod download  # The first build will take 2~3 minutes but will be cached for future builds.

COPY . .

CMD ["echo", "run some tests please"]
