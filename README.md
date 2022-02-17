[![Go Report Card](https://goreportcard.com/badge/github.com/pascallin/go-kit-application)](https://goreportcard.com/report/github.com/pascallin/go-kit-application)
[![Go Reference](https://pkg.go.dev/badge/github.com/pascallin/go-kit-application.svg)](https://pkg.go.dev/github.com/pascallin/go-kit-application)

# go-git-application

A micro-services demo base on go-kit examples

## Project structure

### Code structure

```shell
gateway
|
go-kit transport(http & grpc)
|
go-kit endpoint
|
go-kit service
```

### Using go-kit sd package

- `consul` for discovery
- `zipkin` tracing

### Other standard

- grpc health check endpoint

## Run

### Grpc prepare

```shell
versions:
protoc-3.19.3
protoc-gen-go@v1.26
protoc-gen-go-grpc@v1.1
```

reference: https://grpc.io/docs/languages/go/quickstart/

### Dependent packages

```shell
go mod download
```

### Infrastructure services

needed consul & zipkin

there are some `docker-compose` files in my other github repository([go to reference](https://github.com/pascallin/devops))

```shell
git clone https://github.com/pascallin/devops.git

cd ./zipkin
docker-compose up -d

cd ./consul
docker-compose up -d
```

### Development

All commands stay in `[service]/cmd` folder, we using [air](https://github.com/cosmtrek/air) as live reload tools.

```shell
cd addsvc && air
```

## TODO list

- Prometheus
- Dockerfile
- CI/CD
