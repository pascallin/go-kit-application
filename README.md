# go-micro-services

## run

1. download dependent packages

```
go mod download
```

2. consul

```
docker-compose up -d
```

3. run commands

```shell script
go run ./cmd/addsvc/main.go
go run ./cmd/stringsvc/main.go
go run ./cmd/gateway/main.go
```