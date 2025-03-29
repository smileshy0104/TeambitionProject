#!/bin/bash

# 定义函数
build_macos_amd64() {
    echo "编译macOS版本64位"
    export CGO_ENABLED=0
    export GOOS=darwin
    export GOARCH=amd64
    go build -o project-user/target/project-user project-user/main.go
    go build -o project-api/target/project-api project-api/main.go
}

build_linux_amd64() {
    echo "编译Linux版本64位"
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    go build -o project-user/target/project-user project-user/main.go
    go build -o project-api/target/project-api project-api/main.go
}

# 主程序
clear
echo "请选择要编译的系统环境："
echo "1. macOS_amd64"
echo "2. linux_amd64"

read -p "请选择: " action

if [ "$action" -eq 1 ]; then
    build_macos_amd64
elif [ "$action" -eq 2 ]; then
    build_linux_amd64
else
    echo "无效的选择"
    exit 1
fi
