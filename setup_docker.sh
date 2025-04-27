#!/bin/bash

# 创建 docker 配置目录
sudo mkdir -p /etc/docker

# 配置 Docker 镜像加速器和代理
cat << EOF | sudo tee /etc/docker/daemon.json
{
    "registry-mirrors": [
        "https://mirror.ccs.tencentyun.com",
        "https://hub-mirror.c.163.com",
        "https://registry.docker-cn.com",
        "https://docker.mirrors.ustc.edu.cn"
    ],
    "max-concurrent-downloads": 10,
    "max-concurrent-uploads": 5,
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m",
        "max-file": "3"
    }
}
EOF

# 创建 systemd docker 服务配置目录
sudo mkdir -p /etc/systemd/system/docker.service.d

# 配置 HTTP 代理（如果需要的话，取消注释下面的内容）
# cat << EOF | sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf
# [Service]
# Environment="HTTP_PROXY=http://proxy:port"
# Environment="HTTPS_PROXY=http://proxy:port"
# Environment="NO_PROXY=localhost,127.0.0.1"
# EOF

# 重启 Docker 服务
sudo systemctl daemon-reload
sudo systemctl restart docker

# 验证配置
echo "Docker 配置信息："
docker info

echo "测试镜像拉取："
docker pull hello-world 