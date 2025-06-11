#!/bin/bash
# iLock服务 - 部署脚本

# 服务器配置
TARGET_HOST=${TARGET_HOST:-"117.72.193.54"}
TARGET_PORT=${TARGET_PORT:-"22"}
TARGET_USER=${TARGET_USER:-"root"}
TARGET_PASS=${TARGET_PASS:-"1090119your@"}
BACKUP_DIR="./backups"
VERSION=${VERSION:-"1.4.0"}  # 默认版本号

# 颜色输出
function print_info() { echo -e "\033[0;34m[INFO] $1\033[0m"; }
function print_success() { echo -e "\033[0;32m[SUCCESS] $1\033[0m"; }
function print_error() { echo -e "\033[0;31m[ERROR] $1\033[0m"; }
function print_warning() { echo -e "\033[0;33m[WARNING] $1\033[0m"; }

# 检查sshpass
if ! command -v sshpass &> /dev/null; then
    print_error "未安装sshpass，请先安装："
    echo "  macOS: brew install hudochenkov/sshpass/sshpass"
    echo "  Linux: sudo apt-get install sshpass"
    exit 1
fi

# 检查备份目录
if [ ! -d "$BACKUP_DIR" ]; then
    print_error "备份目录不存在，请先运行备份脚本：./backup.sh"
    exit 1
fi

# 检查备份信息文件
if [ ! -f "$BACKUP_DIR/backup_info.txt" ]; then
    print_error "备份信息文件不存在，请先运行备份脚本：./backup.sh"
    exit 1
fi

# 定义SSH和SCP命令
function ssh_cmd() {
    export SSHPASS="$TARGET_PASS"
    sshpass -e ssh -o StrictHostKeyChecking=no -p "$TARGET_PORT" "$TARGET_USER@$TARGET_HOST" "$@"
}

function scp_to() {
    export SSHPASS="$TARGET_PASS"
    sshpass -e scp -o StrictHostKeyChecking=no -P "$TARGET_PORT" "$1" "$TARGET_USER@$TARGET_HOST:$2"
}

# 读取备份信息
source "$BACKUP_DIR/backup_info.txt"

# 准备目标服务器环境
print_info "准备目标服务器环境..."
ssh_cmd "mkdir -p /root/ilock/backups /root/ilock/logs /root/ilock/configs"

# 检查Docker安装
print_info "检查Docker安装..."
if ! ssh_cmd "command -v docker > /dev/null 2>&1"; then
    print_info "安装Docker..."
    
    # 尝试使用官方脚本安装Docker
    ssh_cmd "curl -fsSL https://get.docker.com -o get-docker.sh"
    ssh_cmd "sh get-docker.sh || true"
    
    # 检查Docker是否成功安装
    if ! ssh_cmd "command -v docker > /dev/null 2>&1"; then
        print_warning "自动安装Docker失败，尝试手动安装..."
        
        # 尝试使用包管理器安装
        if ssh_cmd "command -v apt-get > /dev/null 2>&1"; then
            print_info "检测到Debian/Ubuntu系统，使用apt安装..."
            ssh_cmd "apt-get update && apt-get install -y docker.io"
        elif ssh_cmd "command -v yum > /dev/null 2>&1"; then
            print_info "检测到CentOS/RHEL系统，使用yum安装..."
            ssh_cmd "yum install -y docker"
        else
            print_error "无法自动安装Docker，请手动安装后再运行此脚本"
            exit 1
        fi
    fi
    
    # 启动Docker服务
    ssh_cmd "systemctl enable docker || true"
    ssh_cmd "systemctl start docker || true"
    
    # 再次检查Docker是否可用
    if ! ssh_cmd "docker --version"; then
        print_error "Docker安装失败，请手动安装Docker后再运行此脚本"
        exit 1
    fi
    
    print_success "Docker安装成功"
fi

# 检查Docker Compose安装
print_info "检查Docker Compose安装..."
if ! ssh_cmd "command -v docker-compose > /dev/null 2>&1"; then
    print_info "安装Docker Compose..."
    
    # 尝试多个镜像源
    COMPOSE_URLS=(
        "https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-\$(uname -s)-\$(uname -m)"
        "https://get.daocloud.io/docker/compose/releases/download/v2.21.0/docker-compose-\$(uname -s)-\$(uname -m)"
        "https://ghproxy.com/https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-\$(uname -s)-\$(uname -m)"
    )
    
    COMPOSE_INSTALLED=false
    for URL in "${COMPOSE_URLS[@]}"; do
        print_info "尝试从 $URL 下载Docker Compose..."
        if ssh_cmd "curl -L \"$URL\" -o /usr/local/bin/docker-compose"; then
            ssh_cmd "chmod +x /usr/local/bin/docker-compose"
            COMPOSE_INSTALLED=true
            print_success "Docker Compose安装成功"
            break
        fi
    done
    
    # 如果所有镜像源都失败，尝试使用包管理器安装
    if [ "$COMPOSE_INSTALLED" = false ]; then
        print_warning "从镜像源下载失败，尝试使用包管理器安装..."
        if ssh_cmd "command -v apt-get > /dev/null 2>&1"; then
            ssh_cmd "apt-get update && apt-get install -y docker-compose"
        elif ssh_cmd "command -v yum > /dev/null 2>&1"; then
            ssh_cmd "yum install -y docker-compose"
        fi
    fi
    
    # 确保权限正确
    ssh_cmd "chmod +x /usr/local/bin/docker-compose || true"
    
    # 检查是否成功安装
    if ! ssh_cmd "docker-compose --version"; then
        print_error "Docker Compose安装失败，请手动安装后再运行此脚本"
        exit 1
    fi
else
    # 即使已经安装，也确保权限正确
    ssh_cmd "chmod +x \$(which docker-compose) || true"
fi

# 配置Docker镜像加速
print_info "配置Docker镜像加速..."
ssh_cmd "mkdir -p /etc/docker"
ssh_cmd "cat > /etc/docker/daemon.json << 'EOF'
{
  \"registry-mirrors\": [
    \"https://docker.1ms.run\",
    \"https://registry.cn-hangzhou.aliyuncs.com\",
    \"https://mirror.baidubce.com\",
    \"https://hub-mirror.c.163.com\"
  ],
  \"max-concurrent-downloads\": 3,
  \"max-concurrent-uploads\": 3,
  \"log-driver\": \"json-file\",
  \"log-opts\": {
    \"max-size\": \"10m\",
    \"max-file\": \"3\"
  }
}
EOF"
ssh_cmd "systemctl daemon-reload && systemctl restart docker"

# 上传配置文件
print_info "上传配置文件..."
scp_to "$BACKUP_DIR/.env" "/root/ilock/"

# 修改docker-compose.yml中的版本号
print_info "准备docker-compose.yml文件..."
sed "s|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:$VERSION|g" "$BACKUP_DIR/docker-compose.yml" > "$BACKUP_DIR/docker-compose.modified.yml"

# 检测架构
ARCH=$(uname -m)
if [[ "$ARCH" == "arm64" ]]; then
    print_info "检测到ARM架构，添加平台参数..."
    PLATFORM="linux/amd64"
    sed -i '' -e "s|image: mysql:8.0|image: mysql:8.0\\n    platform: $PLATFORM|g" \
              -e "s|image: redis:7.0-alpine|image: redis:7.0-alpine\\n    platform: $PLATFORM|g" \
              -e "s|image: eclipse-mosquitto:2.0|image: eclipse-mosquitto:2.0\\n    platform: $PLATFORM|g" \
              -e "s|image: stonesea/ilock-http-service:$VERSION|image: stonesea/ilock-http-service:$VERSION\\n    platform: $PLATFORM|g" \
              "$BACKUP_DIR/docker-compose.modified.yml"
fi

scp_to "$BACKUP_DIR/docker-compose.modified.yml" "/root/ilock/docker-compose.yml"

# 上传configs目录
if [ -d "$BACKUP_DIR/configs" ]; then
    print_info "上传configs目录..."
    ssh_cmd "mkdir -p /root/ilock/configs"
    
    # 使用tar打包并通过ssh传输
    tar -czf "$BACKUP_DIR/configs.tar.gz" -C "$BACKUP_DIR" configs
    scp_to "$BACKUP_DIR/configs.tar.gz" "/root/ilock/"
    ssh_cmd "cd /root/ilock && tar -xzf configs.tar.gz && rm configs.tar.gz"
else
    print_warning "configs目录不存在，将创建基本配置..."
    # 创建基本配置目录
    ssh_cmd "mkdir -p /root/ilock/configs/mysql /root/ilock/configs/redis /root/ilock/configs/mqtt"
    
    # 创建Redis配置
    ssh_cmd "cat > /root/ilock/configs/redis/redis.conf << 'EOF'
# Redis配置
port 6379
bind 0.0.0.0
protected-mode yes
maxmemory 256mb
maxmemory-policy allkeys-lru
EOF"

    # 创建MQTT配置
    ssh_cmd "cat > /root/ilock/configs/mqtt/mosquitto.conf << 'EOF'
# 监听端口
listener 1883
listener 8883
listener 9001
protocol websockets

# 持久化设置
persistence true
persistence_location /mosquitto/data/
persistence_file mosquitto.db

# 日志设置
log_dest file /mosquitto/log/mosquitto.log
log_type all

# 默认允许匿名访问
allow_anonymous true
EOF"
fi

# 创建MQTT数据目录
if [ -d "$BACKUP_DIR/mqtt" ]; then
    print_info "上传MQTT数据目录..."
    ssh_cmd "mkdir -p /root/ilock/mqtt"
    
    # 使用tar打包并通过ssh传输
    tar -czf "$BACKUP_DIR/mqtt.tar.gz" -C "$BACKUP_DIR" mqtt
    scp_to "$BACKUP_DIR/mqtt.tar.gz" "/root/ilock/"
    ssh_cmd "cd /root/ilock && tar -xzf mqtt.tar.gz && rm mqtt.tar.gz"
else
    print_info "创建MQTT目录..."
    ssh_cmd "mkdir -p /root/ilock/mqtt/data /root/ilock/mqtt/log"
fi

# 配置防火墙
print_info "配置防火墙..."
if ssh_cmd "command -v ufw > /dev/null 2>&1"; then
    print_info "使用UFW配置防火墙..."
    ssh_cmd "ufw allow 20033/tcp && ufw allow 1883/tcp && ufw allow 8883/tcp && ufw allow 9001/tcp"
elif ssh_cmd "command -v firewall-cmd > /dev/null 2>&1"; then
    print_info "使用Firewalld配置防火墙..."
    ssh_cmd "firewall-cmd --permanent --add-port=20033/tcp && firewall-cmd --permanent --add-port=1883/tcp && firewall-cmd --permanent --add-port=8883/tcp && firewall-cmd --permanent --add-port=9001/tcp && firewall-cmd --reload"
else
    print_warning "未检测到支持的防火墙，请手动配置防火墙规则"
fi

# 启动服务
print_info "启动服务..."
# 设置更长的超时时间，并添加重试机制
MAX_RETRIES=3
for ((i=1; i<=MAX_RETRIES; i++)); do
    print_info "尝试拉取镜像并启动服务 (尝试 $i/$MAX_RETRIES)..."
    
    # 尝试单独拉取每个镜像
    print_info "拉取MySQL镜像..."
    ssh_cmd "docker pull mysql:8.0 || true"
    
    print_info "拉取Redis镜像..."
    ssh_cmd "docker pull redis:7.0-alpine || true"
    
    print_info "拉取MQTT镜像..."
    ssh_cmd "docker pull eclipse-mosquitto:2.0 || true"
    
    print_info "拉取应用镜像..."
    ssh_cmd "docker pull stonesea/ilock-http-service:$VERSION || true"
    
    # 尝试启动服务
    if ssh_cmd "cd /root/ilock && DOCKER_CLIENT_TIMEOUT=600 COMPOSE_HTTP_TIMEOUT=600 docker compose up -d"; then
        print_success "服务启动命令执行成功"
        break
    else
        if [ $i -eq $MAX_RETRIES ]; then
            print_error "多次尝试后服务启动失败"
        else
            print_warning "服务启动失败，将重试..."
            sleep 10
        fi
    fi
done

# 检查服务状态
print_info "等待服务启动..."
for i in {1..60}; do  # 增加等待时间到2分钟
    if ssh_cmd "docker ps | grep 'ilock' | wc -l" | grep -q '[1-9]'; then
        print_success "服务已启动，检查服务状态..."
        ssh_cmd "docker ps | grep 'ilock'"
        break
    fi
    
    if [ $i -eq 60 ]; then
        print_error "服务启动超时，请检查日志："
        ssh_cmd "docker logs \$(docker ps -q -f name=ilock) 2>&1 || echo '没有找到相关容器'"
        exit 1
    fi
    
    echo -n "."
    sleep 2
done

# 等待应用响应
print_info "等待应用响应..."
for i in {1..30}; do
    if ssh_cmd "curl -s http://localhost:20033/api/ping" > /dev/null 2>&1; then
        print_success "应用已成功响应！"
        break
    fi
    
    if [ $i -eq 30 ]; then
        print_warning "应用响应超时，可能需要更长时间启动"
        print_info "查看应用日志："
        ssh_cmd "docker logs \$(docker ps -q -f name=ilock) 2>&1 || echo '没有找到相关容器'"
    fi
    
    echo -n "."
    sleep 2
done

print_success "部署完成！应用已部署到 $TARGET_HOST:20033" 