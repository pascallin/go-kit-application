[![Go Report Card](https://goreportcard.com/badge/github.com/pascallin/go-kit-application)](https://goreportcard.com/report/github.com/pascallin/go-kit-application)
[![Go Reference](https://pkg.go.dev/badge/github.com/pascallin/go-kit-application.svg)](https://pkg.go.dev/github.com/pascallin/go-kit-application)

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
# for development
air

go build -o go-kit-application
go-kit-application
```
