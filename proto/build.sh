#!/bin/bash

protoc --go_out=./ --vison_out=./ --go-grpc_out=./ ./source/*.proto --vison_opt=--serverName=demo--handlePath=../handler


