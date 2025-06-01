#!/bin/bash
# iLock Manual Deployment Script for JD Cloud - 离线部署版本

# 版本设置
VERSION="1.3.0"

# 部署配置
BACKUP_ENABLED=true
AUTO_MIGRATE=true
FORCE_RECREATE=false  # 设置为true将重新创建容器，但保留数据卷
INIT_SERVER=false     # 服务器已初始化，设置为false
DOCKER_PULL_TIMEOUT=600  # Docker拉取超时时间(秒)，增加到10分钟
MAX_RETRIES=3  # 最大重试次数

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

# Build Docker image with version
print_info "Building Docker image version $VERSION..."
docker build --platform linux/amd64 -t "$DOCKER_USERNAME/ilock-http-service:$VERSION" .

# Tag as latest as well
print_info "Tagging as latest..."
docker tag "$DOCKER_USERNAME/ilock-http-service:$VERSION" "$DOCKER_USERNAME/ilock-http-service:latest"

# 更新docker-compose.yml中的版本号
print_info "Updating docker-compose.yml with version $VERSION..."
# macOS版本使用sed进行替换
sed -i '' "s|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:$VERSION|" docker-compose.yml

# 准备离线部署，保存Docker镜像到本地
print_info "准备离线部署，保存Docker镜像到本地..."

# 创建临时目录存放镜像文件
mkdir -p temp_images

# 保存主服务镜像
print_info "保存ilock-http-service镜像..."
docker save -o temp_images/app.tar stonesea/ilock-http-service:$VERSION

# 从docker-compose.yml中提取其他依赖镜像
print_info "提取依赖镜像信息..."
MYSQL_IMAGE=$(grep -o 'mysql:[0-9.]*' docker-compose.yml | head -1 || echo "mysql:8.0")
REDIS_IMAGE=$(grep -o 'redis:[0-9.]*' docker-compose.yml | head -1 || echo "redis:6.2")
MQTT_IMAGE=$(grep -o 'eclipse-mosquitto:[0-9.]*' docker-compose.yml | head -1 || echo "eclipse-mosquitto:2.0")

# 确保镜像名称格式正确
if [[ "$REDIS_IMAGE" == "redis:" ]]; then
  REDIS_IMAGE="redis:6.2"
  print_warning "Redis镜像版本未指定，使用默认版本 $REDIS_IMAGE"
fi

print_info "使用以下镜像版本:"
echo "MySQL: $MYSQL_IMAGE"
echo "Redis: $REDIS_IMAGE"
echo "MQTT: $MQTT_IMAGE"

# 是否使用离线部署
USE_OFFLINE_DEPLOY=false

if [ "$USE_OFFLINE_DEPLOY" = true ]; then
  # 拉取并保存依赖镜像
  print_info "拉取并保存MySQL镜像 ($MYSQL_IMAGE)..."
  docker pull --platform linux/amd64 $MYSQL_IMAGE
  docker save -o temp_images/mysql.tar $MYSQL_IMAGE

  print_info "拉取并保存Redis镜像 ($REDIS_IMAGE)..."
  docker pull --platform linux/amd64 $REDIS_IMAGE
  docker save -o temp_images/redis.tar $REDIS_IMAGE

  print_info "拉取并保存MQTT镜像 ($MQTT_IMAGE)..."
  docker pull --platform linux/amd64 $MQTT_IMAGE
  docker save -o temp_images/mqtt.tar $MQTT_IMAGE

  # 打包所有镜像文件
  print_info "打包所有镜像文件..."

  # 创建导入镜像的脚本
  cat > temp_images/import_images.sh << 'EOF'
#!/bin/bash
# 导入Docker镜像脚本

echo "开始导入Docker镜像..."

# 导入应用镜像
echo "导入应用镜像..."
docker load -i app.tar
if [ $? -ne 0 ]; then
  echo "导入应用镜像失败!"
  exit 1
fi

# 导入MySQL镜像
echo "导入MySQL镜像..."
docker load -i mysql.tar
if [ $? -ne 0 ]; then
  echo "导入MySQL镜像失败!"
  exit 1
fi

# 导入Redis镜像
echo "导入Redis镜像..."
docker load -i redis.tar
if [ $? -ne 0 ]; then
  echo "导入Redis镜像失败!"
  exit 1
fi

# 导入MQTT镜像
echo "导入MQTT镜像..."
docker load -i mqtt.tar
if [ $? -ne 0 ]; then
  echo "导入MQTT镜像失败!"
  exit 1
fi

echo "所有镜像导入成功!"
echo "可用的Docker镜像列表:"
docker images
EOF

  # 添加执行权限
  chmod +x temp_images/import_images.sh

  tar -czf docker_images.tar.gz -C temp_images .
else
  # 不使用离线部署，只上传主应用镜像
  print_info "不使用离线部署，将直接从Docker Hub拉取镜像..."
  # 推送镜像到Docker Hub
  print_info "推送镜像到Docker Hub..."
  docker push "$DOCKER_USERNAME/ilock-http-service:$VERSION"
  docker push "$DOCKER_USERNAME/ilock-http-service:latest"
fi

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

# 创建合并镜像文件的脚本
cat > merge_chunks.sh << 'EOF'
#!/bin/bash
# 合并分块上传的文件

echo "正在合并镜像文件块..."
cat docker_images.tar.gz.part* > docker_images.tar.gz
rm -f docker_images.tar.gz.part*
echo "镜像文件合并完成!"

# 解压并导入镜像
echo "正在解压镜像文件..."
mkdir -p docker_images
tar -xzf docker_images.tar.gz -C docker_images
cd docker_images
chmod +x import_images.sh
./import_images.sh
cd ..
echo "镜像导入完成!"
EOF

# 创建Docker镜像加速配置脚本
cat > setup_docker_mirror.sh << 'EOF'
#!/bin/bash
# 配置Docker镜像加速

echo '{
  "registry-mirrors": [
    "https://docker.1ms.run",
    "https://docker.mybacc.com",
    "https://dytt.online",
    "https://lispy.org",
    "https://docker.xiaogenban1993.com",
    "https://docker.yomansunter.com",
    "https://aicarbon.xyz",
    "https://666860.xyz",
    "https://hub.littlediary.cn",
    "https://hub.rat.dev",
    "https://docker.m.daocloud.io",
    "https://dockerproxy.net",
    "https://registry.docker-cn.com",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com",
    "https://docker.mirrors.ustc.edu.cn"
  ]
}' > /etc/docker/daemon.json

# 重启Docker服务
systemctl restart docker
echo "Docker镜像加速配置完成"
EOF

# Copy files to server using scp with password
print_info "Copying deployment files to server..."
scp_cmd docker-compose.yml .env backup_script.sh merge_chunks.sh setup_docker_mirror.sh

# 执行部署前的准备工作
print_info "Executing pre-deployment tasks on server..."
ssh_cmd "cd /root/ilock && chmod +x backup_script.sh merge_chunks.sh setup_docker_mirror.sh && mkdir -p /root/ilock/backups"

# 创建MQTT所需目录
print_info "Creating required directories for MQTT..."
ssh_cmd "cd /root/ilock && mkdir -p mqtt/config mqtt/data mqtt/log"

# 创建MQTT配置文件
print_info "Creating MQTT configuration file..."
cat > mosquitto.conf << 'EOF'
# Mosquitto配置文件
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
log_dest file /mosquitto/log/mosquitto.log
EOF

# 上传MQTT配置文件
scp_cmd mosquitto.conf
ssh_cmd "cd /root/ilock && mkdir -p mqtt/config && mv mosquitto.conf mqtt/config/"

# 配置Docker镜像加速
print_info "Setting up Docker mirror for faster image pulling..."
ssh_cmd "cd /root/ilock && ./setup_docker_mirror.sh"

# 分块上传镜像包到服务器
print_info "分块上传镜像包到服务器 \"这可能需要一些时间\"..."
# 获取文件大小（以字节为单位）
if [ "$USE_OFFLINE_DEPLOY" = true ]; then
  FILE_SIZE=$(stat -f%z docker_images.tar.gz)
  # 每个块的大小（10MB）
  CHUNK_SIZE=$((10 * 1024 * 1024))
  # 计算需要多少个块
  CHUNKS=$(( (FILE_SIZE + CHUNK_SIZE - 1) / CHUNK_SIZE ))

  print_info "文件大小: $(( FILE_SIZE / 1024 / 1024 ))MB, 将分成 $CHUNKS 个块上传"

  # 分割文件
  split -b $CHUNK_SIZE docker_images.tar.gz docker_images.tar.gz.part

  # 逐个上传分块
  for part in docker_images.tar.gz.part*; do
    print_info "上传分块 $part..."
    scp_cmd "$part"
    if [ $? -ne 0 ]; then
      print_error "上传分块 $part 失败，请检查网络连接"
      exit 1
    fi
  done
else
  print_info "跳过镜像上传，将直接从Docker Hub拉取镜像..."
fi

# 删除本地临时脚本和目录
rm -f backup_script.sh merge_chunks.sh setup_docker_mirror.sh mosquitto.conf
if [ "$USE_OFFLINE_DEPLOY" = true ]; then
  rm -rf temp_images
  rm -f docker_images.tar.gz.part*
fi

# 在服务器上合并分块并导入镜像
if [ "$USE_OFFLINE_DEPLOY" = true ]; then
  print_info "在服务器上合并分块并导入镜像..."
  ssh_cmd "cd /root/ilock && ./merge_chunks.sh"
fi

# 创建数据库备份（如果启用且不是首次部署）
if [ "$BACKUP_ENABLED" = true ] && [ "$INIT_SERVER" = false ]; then
  print_info "Creating database backup before deployment..."
  ssh_cmd "cd /root/ilock && ./backup_script.sh"
else
  print_warning "Database backup is skipped (either disabled or first deployment)"
fi

# 更新Docker镜像但不重建容器
print_info "Updating Docker containers while preserving data..."
if [ "$FORCE_RECREATE" = true ]; then
  RECREATE_FLAG="--force-recreate"
  print_warning "Force recreate flag is enabled. Containers will be recreated but volumes preserved."
else
  RECREATE_FLAG=""
fi

# 启动服务
print_info "启动服务..."
# 先停止并移除旧容器，避免ContainerConfig错误
ssh_cmd "cd /root/ilock && docker-compose down"
# 然后重新启动服务
if [ "$USE_OFFLINE_DEPLOY" = false ]; then
  # 直接从Docker Hub拉取镜像
  print_info "从Docker Hub拉取镜像..."
  ssh_cmd "cd /root/ilock && COMPOSE_HTTP_TIMEOUT=$DOCKER_PULL_TIMEOUT docker-compose pull"
fi

# 启动服务
for i in {1..3}; do
  ssh_cmd "cd /root/ilock && docker-compose up -d $RECREATE_FLAG"
  if [ $? -eq 0 ]; then
    print_success "服务启动成功！(尝试 $i/3)"
    break
  else
    if [ $i -eq 3 ]; then
      print_error "服务启动失败，已达到最大重试次数"
    else
      print_warning "服务启动失败，正在重试... (尝试 $i/3)"
      sleep 5
    fi
  fi
done

# 确保.env文件权限正确
print_info "Ensuring .env file has proper permissions..."
ssh_cmd "cd /root/ilock && chmod 644 .env"

# 等待MySQL就绪
print_info "Waiting for MySQL to be ready..."
ssh_cmd "cd /root/ilock && for i in {1..60}; do if docker-compose ps db | grep -q 'Up'; then if docker exec ilock_mysql mysqladmin ping -h localhost --silent; then echo 'MySQL is ready!'; break; fi; fi; if [ \$i -eq 60 ]; then echo 'MySQL startup timeout'; docker-compose logs db; exit 1; fi; if [ \$((\$i % 10)) -eq 0 ]; then echo 'MySQL starting... (attempt '\$i'/60)'; docker-compose logs --tail=5 db; fi; sleep 5; done"

# 等待Redis就绪
print_info "Waiting for Redis to be ready..."
ssh_cmd "cd /root/ilock && for i in {1..60}; do if docker-compose ps redis | grep -q 'Up'; then if docker exec ilock_redis redis-cli ping | grep -q 'PONG'; then echo 'Redis is ready!'; break; fi; fi; if [ \$i -eq 60 ]; then echo 'Redis startup timeout'; docker-compose logs redis; exit 1; fi; if [ \$((\$i % 10)) -eq 0 ]; then echo 'Redis starting... (attempt '\$i'/60)'; docker-compose logs --tail=5 redis; fi; sleep 5; done"

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
ssh_cmd "cd /root/ilock && for i in {1..60}; do if docker-compose ps app | grep -q 'Up'; then if curl -s http://localhost:20033/api/ping > /dev/null 2>&1; then echo 'Application service started successfully!'; docker-compose ps; exit 0; fi; fi; if [ \$i -eq 60 ]; then echo 'Application service timeout'; docker-compose logs app | tail -n 50; exit 1; fi; if [ \$((\$i % 5)) -eq 0 ]; then echo 'Application service starting... (attempt '\$i'/60)'; docker-compose logs --tail=10 app; fi; sleep 5; done"

# 检查SSH返回值来判断部署是否成功
if [ $? -ne 0 ]; then
  print_error "Deployment failed. Please check the logs."
  
  # 提供故障恢复建议
  print_info "故障排查建议:"
  echo "1. 检查服务器上的Docker镜像是否正确导入: ssh到服务器执行 'docker images'"
  echo "2. 检查容器状态: ssh到服务器执行 'cd /root/ilock && docker-compose ps'"
  echo "3. 查看应用日志: ssh到服务器执行 'cd /root/ilock && docker-compose logs app'"
  echo "4. 查看MySQL日志: ssh到服务器执行 'cd /root/ilock && docker-compose logs db'"
  echo "5. 检查网络连接: ssh到服务器执行 'docker network inspect ilock_ilock_network'"
  echo "6. 手动重启服务: ssh到服务器执行 'cd /root/ilock && docker-compose down && docker-compose up -d'"
  echo ""
  echo "快速登录服务器命令: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST"
else
  print_success "Deployment successful! Deployed version $VERSION"
  
  # 提供一些有用的后续命令
  echo ""
  print_info "Useful commands:"
  echo "  - View logs: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST 'cd /root/ilock && docker-compose logs -f'"
  echo "  - Check status: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST 'cd /root/ilock && docker-compose ps'"
  echo "  - List backups: export SSHPASS='$SSH_PASSWORD' && sshpass -e ssh -o StrictHostKeyChecking=no $SSH_USERNAME@$SSH_HOST 'ls -la /root/ilock/backups/'"
fi

print_info "清理本地临时文件..."
rm -f docker_images.tar.gz

print_info "离线部署脚本执行完成" 