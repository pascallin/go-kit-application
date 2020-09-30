# go-git-application

A micro-services demo base on go-kit examples

## run

1. download dependent packages

``` 
go mod download
```

2. consul & zipkin

there are some `docker-compose` files in my other github repository([go to reference](https://github.com/pascallin/devops))

``` 
git clone https://github.com/pascallin/devops.git

cd ./zipkin
docker-compose up -d

cd ./consul
docker-compose up -d
```

3. run commands

```shell script
go run ./cmd/addsvc/main.go
go run ./cmd/stringsvc/main.go
go run ./cmd/gateway/main.go
```
