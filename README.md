# immria-api
Dockerized Golang API Server for Immria

## Requirements
* Install [Taskfile](https://taskfile.dev/installation/)

## Setup
* `cp config-example.yml config.yml`
* Fill out params
* `docker compose build`
* `docker compose up -d`

## Commands
* `go run internal/main.go` -> Starts API 
* `go build -o bin/go-rest-api internal/main.go` -> Generates an executable binary with our HTTP server
* `task swagger.validate`
* `task swagger.doc`
* `task swagger.gen`