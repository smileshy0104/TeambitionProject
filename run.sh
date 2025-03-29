#!/bin/bash

# 设置字符编码为 UTF-8
export LANG=en_US.UTF-8

# 进入 project-user 目录
cd project-user || { echo "目录 project-user 不存在"; exit 1; }

# 构建 Docker 镜像
docker build -t project-user:latest .

# 返回上一级目录
cd ..

# 启动 Docker Compose 服务
docker-compose up -d
