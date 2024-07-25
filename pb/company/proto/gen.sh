#!/bin/bash
serverName="company"
protoc --go_out=../ --go_opt=paths=source_relative --go-grpc_out=../ --go-grpc_opt=paths=source_relative *.proto
echo "${serverName}：生成文件结束"
