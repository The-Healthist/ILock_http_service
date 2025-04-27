#!/bin/bash

# 确保脚本在出错时停止执行
set -e

echo "开始部署 iLock 服务..."

# 确保目录存在
mkdir -p logs

# 登录到 Docker Hub
echo "登录 Docker Hub..."
docker login

# 停止并删除现有容器
echo "停止现有服务..."
docker-compose down --volumes --remove-orphans || true

# 删除旧镜像
echo "删除旧镜像..."
docker rmi stonesea/ilock-service:latest || true

# 直接拉取指定镜像
echo "拉取最新镜像..."
docker pull stonesea/ilock-service:latest

# 启动服务
echo "启动服务..."
docker-compose up -d

# 等待服务启动
echo "等待服务启动..."
sleep 15

# 检查容器状态
echo "检查容器状态..."
docker-compose ps

# 健康检查
echo "进行健康检查..."
for i in {1..30}; do
    if curl -s http://localhost:20033/api/ping > /dev/null; then
        echo "服务启动成功！"
        echo "容器状态："
        docker-compose ps
        exit 0
    fi
    echo "尝试 $i/30: 服务尚未就绪..."
    
    # 显示容器日志
    echo "容器日志："
    docker-compose logs --tail=50 app
    
    sleep 2
done

echo "服务启动失败"
echo "完整容器日志："
docker-compose logs
docker-compose ps
exit 1 