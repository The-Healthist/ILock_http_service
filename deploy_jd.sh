

#!/bin/bash
# iLock Manual Deployment Script for macOS - 安全部署版本

# 版本设置
VERSION="1.3.0"

# 部署配置
BACKUP_ENABLED=true
AUTO_MIGRATE=true
FORCE_RECREATE=false  # 设置为true将重新创建容器，但保留数据卷
DOCKER_PULL_TIMEOUT=600  # Docker拉取超时时间(秒)，增加到10分钟
MAX_RETRIES=3  # 最大重试次数
USE_OFFLINE_IMAGES=false  # 不使用离线镜像，直接从Docker Hub拉取

# Server settings - 请修改为您的京东云服务器信息
SSH_HOST="117.72.193.54"
SSH_PORT="22"
SSH_USERNAME="root"
SSH_PASSWORD="1090119your@"

# Docker Hub settings
DOCKER_USERNAME="stonesea"
DOCKER_PASSWORD="1090119your"

# 颜色输出函数
function print_info() {
  echo -e "\033[0;34m[INFO] $1\033[0m"
}

function print_success() {
  echo -e "\033[0;32m[SUCCESS] $1\033[0m"
}

function print_warning() {
  echo -e "\033[0;33m[WARNING] $1\033[0m"
}

function print_error() {
  echo -e "\033[0;31m[ERROR] $1\033[0m"
}

# 检查命令是否存在
command -v swag >/dev/null 2>&1 || { print_error "需要安装swag工具！请运行: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1; }
command -v docker >/dev/null 2>&1 || { print_error "需要安装Docker！"; exit 1; }

# 检查sshpass是否安装
if ! command -v sshpass &> /dev/null; then
  print_warning "sshpass未安装，将尝试安装..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS系统
    brew install sshpass || { 
      print_error "sshpass安装失败！请手动安装: brew install sshpass"; 
      print_info "如果brew无法直接安装，请尝试: brew install hudochenkov/sshpass/sshpass";
      exit 1; 
    }
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux系统
    sudo apt-get update && sudo apt-get install -y sshpass || { print_error "sshpass安装失败！请手动安装: sudo apt-get install sshpass"; exit 1; }
  else
    print_error "无法识别的操作系统，请手动安装sshpass后重试"; 
    exit 1;
  fi
  print_success "sshpass安装成功"
fi

# 定义SSH和SCP命令的函数，自动使用密码
function ssh_cmd() {
  export SSHPASS="$SSH_PASSWORD"
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USERNAME@$SSH_HOST" "$@"
}

function scp_cmd() {
  export SSHPASS="$SSH_PASSWORD"
  sshpass -e scp -o StrictHostKeyChecking=no -P "$SSH_PORT" "$@" "$SSH_USERNAME@$SSH_HOST:/root/ilock/"
}

# 重新生成Swagger文档
print_info "Regenerating Swagger documentation..."
swag init -g main.go

# Login to Docker Hub
print_info "Logging in to Docker Hub..."
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

# Build Docker image with version - 添加平台参数
print_info "Building Docker image version $VERSION..."
docker build --platform linux/amd64 -t "$DOCKER_USERNAME/ilock-http-service:$VERSION" .

# Tag as latest as well
print_info "Tagging as latest..."
docker tag "$DOCKER_USERNAME/ilock-http-service:$VERSION" "$DOCKER_USERNAME/ilock-http-service:latest"

# Push Docker image to Docker Hub
print_info "Pushing versioned Docker image to Docker Hub..."
docker push "$DOCKER_USERNAME/ilock-http-service:$VERSION"

print_info "Pushing latest Docker image to Docker Hub..."
docker push "$DOCKER_USERNAME/ilock-http-service:latest"

# 更新docker-compose.yml中的版本号
print_info "Updating docker-compose.yml with version $VERSION..."
# macOS版本使用sed进行替换
sed -i '' "s|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:$VERSION|" docker-compose.yml

# 准备备份脚本 - 在服务器上创建
cat > backup_script.sh << 'EOF'
#!/bin/bash
# 数据库备份脚本

# 备份目录
BACKUP_DIR="/root/ilock/backups"
BACKUP_COUNT=7  # 保留最近7次备份

# 创建备份目录
mkdir -p $BACKUP_DIR

# 当前时间戳
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/ilock_db_$TIMESTAMP.sql"

# 检查MySQL容器是否运行
if docker ps | grep -q ilock_mysql; then
  echo "Creating MySQL backup..."
  # 使用docker exec执行MySQL备份，避免需要MySQL密码
  docker exec ilock_mysql sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --all-databases' > "$BACKUP_FILE"
  
  if [ $? -eq 0 ]; then
    echo "MySQL backup created successfully: $BACKUP_FILE"
    # 压缩备份文件
    gzip "$BACKUP_FILE"
    echo "Backup compressed: $BACKUP_FILE.gz"
    
    # 清理旧备份，只保留最近的BACKUP_COUNT个
    ls -t "$BACKUP_DIR"/ilock_db_*.sql.gz | tail -n +$((BACKUP_COUNT+1)) | xargs -r rm
    echo "Old backups cleaned up, keeping most recent $BACKUP_COUNT backups"
  else
    echo "MySQL backup failed!"
    exit 1
  fi
else
  echo "MySQL container is not running, skipping backup"
fi

# 备份Redis数据（如果需要）
if docker ps | grep -q ilock_redis; then
  echo "Backing up Redis data..."
  REDIS_BACKUP_FILE="$BACKUP_DIR/ilock_redis_$TIMESTAMP.rdb"
  docker exec ilock_redis sh -c 'redis-cli save && cat /data/dump.rdb' > "$REDIS_BACKUP_FILE"
  
  if [ $? -eq 0 ]; then
    echo "Redis backup created successfully: $REDIS_BACKUP_FILE"
    # 压缩备份文件
    gzip "$REDIS_BACKUP_FILE"
    echo "Redis backup compressed: $REDIS_BACKUP_FILE.gz"
    
    # 清理旧备份，只保留最近的BACKUP_COUNT个
    ls -t "$BACKUP_DIR"/ilock_redis_*.rdb.gz | tail -n +$((BACKUP_COUNT+1)) | xargs -r rm
  else
    echo "Redis backup failed!"
  fi
else
  echo "Redis container is not running, skipping backup"
fi

echo "Backup process completed!"
EOF

# 准备MQTT配置文件
mkdir -p mqtt/config
cat > mqtt/config/mosquitto.conf << 'EOF'
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
EOF

# 准备Docker镜像加速配置
cat > setup_docker_mirror.sh << 'EOF'
#!/bin/bash

# 创建或更新Docker配置目录
mkdir -p /etc/docker

# 创建daemon.json配置文件
cat > /etc/docker/daemon.json << 'INNEREOF'
{
  "registry-mirrors": [
    "https://docker.1ms.run",
    "https://docker.mybacc.com",
    "https://dytt.online",
    "https://lispy.org",
    "https://docker.xiaogenban1993.com",
    "https://docker.yomansunter.com",
    "https://aicarbon.xyz",
    "https://666860.xyz",
    "https://docker.zhai.cm",
    "https://a.ussh.net",
    "https://hub.littlediary.cn",
    "https://hub.rat.dev",
    "https://docker.m.daocloud.io",
    "https://registry.cn-hangzhou.aliyuncs.com"
  ]
}
INNEREOF

# 重启Docker服务
systemctl daemon-reload
systemctl restart docker

echo "Docker镜像加速配置完成"
EOF

# Copy files to server using scp with password
print_info "Copying deployment files to server..."
scp_cmd docker-compose.yml .env backup_script.sh setup_docker_mirror.sh

# 删除本地备份脚本
rm backup_script.sh

# 执行部署前的准备工作
print_info "Executing pre-deployment tasks on server..."
ssh_cmd "cd /root/ilock && chmod +x backup_script.sh && mkdir -p /root/ilock/backups"

# 创建MQTT所需目录
print_info "Creating required directories for MQTT..."
ssh_cmd "cd /root/ilock && mkdir -p mqtt/config mqtt/data mqtt/log"

# 上传MQTT配置文件
print_info "Uploading MQTT configuration file..."
# 直接在服务器上创建配置文件，避免上传问题
ssh_cmd "cat > /root/ilock/mqtt/config/mosquitto.conf << 'EOFMQTT'
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
EOFMQTT"

# 确保配置文件有正确的权限
ssh_cmd "chmod 644 /root/ilock/mqtt/config/mosquitto.conf"

# 配置Docker镜像加速
print_info "Setting up Docker mirror for faster image pulling..."
ssh_cmd "cd /root/ilock && chmod +x setup_docker_mirror.sh && ./setup_docker_mirror.sh"

# 创建数据库备份（如果启用）
if [ "$BACKUP_ENABLED" = true ]; then
  print_info "Creating database backup before deployment..."
  ssh_cmd "cd /root/ilock && ./backup_script.sh"
else
  print_warning "Database backup is disabled, skipping backup step"
fi

# 更新Docker镜像但不重建容器
print_info "Updating Docker containers while preserving data..."
if [ "$FORCE_RECREATE" = true ]; then
  RECREATE_FLAG="--force-recreate"
  print_warning "Force recreate flag is enabled. Containers will be recreated but volumes preserved."
else
  RECREATE_FLAG=""
fi

# 拉取新镜像并更新服务
print_info "Setting Docker client timeout to $DOCKER_PULL_TIMEOUT seconds..."
# 先停止并移除旧容器，避免ContainerConfig错误
ssh_cmd "cd /root/ilock && docker-compose down"

# 修改服务器上的docker-compose.yml，添加平台参数
print_info "Adding platform parameters to docker-compose.yml on server..."
ssh_cmd "cd /root/ilock && sed -i 's|image: mysql:8.0|image: mysql:8.0\\n    platform: linux/amd64|' docker-compose.yml"
ssh_cmd "cd /root/ilock && sed -i 's|image: redis:7.0-alpine|image: redis:7.0-alpine\\n    platform: linux/amd64|' docker-compose.yml"
ssh_cmd "cd /root/ilock && sed -i 's|image: eclipse-mosquitto:2.0|image: eclipse-mosquitto:2.0\\n    platform: linux/amd64|' docker-compose.yml"
ssh_cmd "cd /root/ilock && sed -i 's|image: stonesea/ilock-http-service:$VERSION|image: stonesea/ilock-http-service:$VERSION\\n    platform: linux/amd64|' docker-compose.yml"

# 检查MQTT配置文件是否存在
print_info "Verifying MQTT configuration..."
ssh_cmd "cd /root/ilock && ls -la mqtt/config/ && cat mqtt/config/mosquitto.conf"

# 然后拉取镜像并启动服务
ssh_cmd "cd /root/ilock && echo 'Pulling new images...' && COMPOSE_HTTP_TIMEOUT=$DOCKER_PULL_TIMEOUT docker-compose pull || (echo '首次拉取失败，正在重试...' && COMPOSE_HTTP_TIMEOUT=$DOCKER_PULL_TIMEOUT docker-compose pull) && echo 'Updating services...' && docker-compose up -d $RECREATE_FLAG"

# 确保.env文件权限正确
print_info "Ensuring .env file has proper permissions..."
ssh_cmd "cd /root/ilock && chmod 644 .env"

# 等待MySQL就绪
print_info "Waiting for MySQL to be ready..."
ssh_cmd "cd /root/ilock && for i in {1..30}; do if docker-compose ps db | grep -q 'Up'; then if docker exec ilock_mysql mysqladmin ping -h localhost --silent; then echo 'MySQL is ready!'; break; fi; fi; if [ \$i -eq 30 ]; then echo 'MySQL startup timeout'; docker-compose logs db; exit 1; fi; echo 'MySQL starting... (attempt '\$i'/30)'; sleep 2; done"

# 等待Redis就绪
print_info "Waiting for Redis to be ready..."
ssh_cmd "cd /root/ilock && for i in {1..30}; do if docker-compose ps redis | grep -q 'Up'; then if docker exec ilock_redis redis-cli ping | grep -q 'PONG'; then echo 'Redis is ready!'; break; fi; fi; if [ \$i -eq 30 ]; then echo 'Redis startup timeout'; docker-compose logs redis; exit 1; fi; echo 'Redis starting... (attempt '\$i'/30)'; sleep 2; done"

# 等待MQTT就绪
print_info "Waiting for MQTT to be ready..."
ssh_cmd "cd /root/ilock && for i in {1..30}; do if docker-compose ps mqtt | grep -q 'Up'; then echo 'MQTT is ready!'; break; fi; if [ \$i -eq 30 ]; then echo 'MQTT startup timeout'; docker-compose logs mqtt; exit 1; fi; echo 'MQTT starting... (attempt '\$i'/30)'; sleep 2; done"

# 如果MQTT启动失败，显示日志
print_info "Checking MQTT logs if service is not ready..."
ssh_cmd "cd /root/ilock && if ! docker-compose ps mqtt | grep -q 'Up'; then echo 'MQTT failed to start, checking logs:'; docker-compose logs mqtt; fi"

# 检查是否需要运行数据库迁移
if [ "$AUTO_MIGRATE" = true ]; then
  print_info "Checking for migration script..."
  if ssh_cmd "cd /root/ilock && docker exec ilock_http_service ls -la /app/run_migrations.sh 2>/dev/null"; then
    print_info "Running database migrations..."
    ssh_cmd "cd /root/ilock && docker exec ilock_http_service /app/run_migrations.sh"
    
    if [ $? -ne 0 ]; then
      print_error "Database migration failed! Check the logs for details."
      exit 1
    else
      print_success "Database migration completed successfully."
    fi
  else
    print_warning "Migration script not found in container. Skipping database migration."
  fi
else
  print_warning "Automatic database migration is disabled."
fi

# 检查应用容器的状态
print_info "Checking application service status..."
ssh_cmd "cd /root/ilock && docker-compose ps && docker exec ilock_http_service ls -la /app"

# 等待应用服务就绪
print_info "Waiting for application service to be ready..."
ssh_cmd "cd /root/ilock && for i in {1..60}; do if docker-compose ps app | grep -q 'Up'; then if curl -s http://localhost:20033/api/ping > /dev/null 2>&1; then echo 'Application service started successfully!'; docker-compose ps; exit 0; fi; fi; if [ \$i -eq 60 ]; then echo 'Application service timeout'; docker-compose logs app; exit 1; fi; if [ \$((\$i % 5)) -eq 0 ]; then echo 'Application service starting... (attempt '\$i'/60)'; docker-compose logs --tail=10 app; fi; sleep 2; done"

# 检查SSH返回值来判断部署是否成功
if [ $? -ne 0 ]; then
  print_error "Deployment failed. Please check the logs."
else
  print_success "Deployment successful! Deployed version $VERSION"
  print_info "To rollback this deployment if needed, use: ./rollback.sh $VERSION"
  
  # 提供一些有用的后续命令
  echo ""
  print_info "Useful commands:"
  echo "  - View logs: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST 'cd /root/ilock && docker-compose logs -f'"
  echo "  - Check status: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST 'cd /root/ilock && docker-compose ps'"
  echo "  - List backups: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST 'ls -la /root/ilock/backups/'"
fi

# 为了方便回滚，我们创建一个回滚脚本
cat > rollback.sh << EOF
#!/bin/bash
# Rollback script for iLock deployment

# 服务器设置
SSH_HOST="$SSH_HOST"
SSH_PORT="$SSH_PORT"
SSH_USERNAME="$SSH_USERNAME"
SSH_PASSWORD="$SSH_PASSWORD"

if [ \$# -ne 1 ]; then
  echo "Usage: \$0 <version_to_rollback_from>"
  exit 1
fi

VERSION_TO_ROLLBACK=\$1
PREVIOUS_VERSION="1.1.0"  # 设置为上一个稳定版本

echo "Rolling back from version \$VERSION_TO_ROLLBACK to \$PREVIOUS_VERSION"
echo "This will update docker-compose.yml and restart the service"
read -p "Continue? (y/n) " -n 1 -r
echo 
if [[ ! \$REPLY =~ ^[Yy]\$ ]]; then
  echo "Rollback aborted"
  exit 1
fi

# Update docker-compose.yml
sed -i '' "s|image: stonesea/ilock-http-service:\$VERSION_TO_ROLLBACK|image: stonesea/ilock-http-service:\$PREVIOUS_VERSION|" docker-compose.yml

# Copy to server and restart using sshpass
export SSHPASS="\$SSH_PASSWORD"
sshpass -e scp -o StrictHostKeyChecking=no -P "\$SSH_PORT" docker-compose.yml "\$SSH_USERNAME@\$SSH_HOST:/root/ilock/"
sshpass -e ssh -o StrictHostKeyChecking=no -p "\$SSH_PORT" "\$SSH_USERNAME@\$SSH_HOST" "cd /root/ilock && docker-compose pull && docker-compose up -d"

echo "Rollback completed!"
EOF

chmod +x rollback.sh

# 删除本地备份脚本和临时文件
rm -f setup_docker_mirror.sh
