#!/bin/bash
# iLock Simplified Deployment Script

# 版本设置
VERSION="1.3.0"

# 部署配置
BACKUP_ENABLED=true  # 启用备份功能
AUTO_MIGRATE=true    # 启用自动迁移功能

# Server settings
SSH_HOST="39.108.49.167"
SSH_PORT="22"
SSH_USERNAME="root"
SSH_PASSWORD="1090119your@"

# Docker Hub settings
DOCKER_USERNAME="stonesea"
DOCKER_PASSWORD="1090119your"

# 颜色输出函数
function print_info() { echo -e "\033[0;34m[INFO] $1\033[0m"; }
function print_success() { echo -e "\033[0;32m[SUCCESS] $1\033[0m"; }
function print_error() { echo -e "\033[0;31m[ERROR] $1\033[0m"; }
function print_warning() { echo -e "\033[0;33m[WARNING] $1\033[0m"; }

# 检查必要工具
command -v swag >/dev/null 2>&1 || { print_error "需要安装swag工具！请运行: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1; }
command -v docker >/dev/null 2>&1 || { print_error "需要安装Docker！"; exit 1; }
command -v sshpass &> /dev/null || { brew install sshpass || { print_error "sshpass安装失败！"; exit 1; }; }

# 定义SSH和SCP命令的函数
function ssh_cmd() {
  export SSHPASS="$SSH_PASSWORD"
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USERNAME@$SSH_HOST" "$@"
}

function scp_cmd() {
  export SSHPASS="$SSH_PASSWORD"
  sshpass -e scp -o StrictHostKeyChecking=no -P "$SSH_PORT" "$@" "$SSH_USERNAME@$SSH_HOST:/root/ilock/"
}

# 准备备份脚本
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
  # 使用docker exec执行MySQL备份
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

# 备份Redis数据
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

# 重新生成Swagger文档
print_info "重新生成Swagger文档..."
swag init -g main.go

# Login to Docker Hub并构建推送镜像
print_info "登录Docker Hub并构建推送镜像..."
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker buildx create --use --name multi-platform-builder || true
docker buildx build --platform linux/amd64 -t "$DOCKER_USERNAME/ilock-http-service:$VERSION" -t "$DOCKER_USERNAME/ilock-http-service:latest" --push .

# 更新docker-compose.yml版本
print_info "更新docker-compose.yml版本为 $VERSION..."
sed -i '' "s|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:$VERSION|" docker-compose.yml

# 创建修改后的docker-compose.yml文件
cat > docker-compose.modified.yml << EOF
services: 
  app: 
    image: stonesea/ilock-http-service:$VERSION
    container_name: ilock_http_service 
    restart: always 
    ports: 
      - '20033:20033'
    volumes: 
      - ./logs:/app/logs 
      - ./.env:/app/.env 
    environment: 
      - ENV_TYPE=SERVER 
      - ALIYUN_ACCESS_KEY=\${ALIYUN_ACCESS_KEY} 
      - ALIYUN_RTC_APP_ID=\${ALIYUN_RTC_APP_ID} 
      - ALIYUN_RTC_REGION=\${ALIYUN_RTC_REGION} 
      - DEFAULT_ADMIN_PASSWORD=\${DEFAULT_ADMIN_PASSWORD} 
    depends_on: 
      db: 
        condition: service_healthy 
      redis: 
        condition: service_healthy 
    networks: 
      - ilock_network 
    healthcheck: 
      test: ['CMD', 'curl', '-f', 'http://localhost:20033/api/ping']
      interval: 10s 
      timeout: 5s 
      retries: 3 
      start_period: 10s 
 
  db: 
    image: mysql:8.0 
    container_name: ilock_mysql 
    restart: always 
    ports: 
      - '3310:3306'
    volumes: 
      - mysql_data:/var/lib/mysql 
    environment: 
      - MYSQL_ROOT_PASSWORD=\${MYSQL_ROOT_PASSWORD} 
      - MYSQL_DATABASE=\${MYSQL_DATABASE} 
    command: --default-authentication-plugin=mysql_native_password 
    networks: 
      - ilock_network 
    healthcheck: 
      test: ['CMD', 'mysqladmin', 'ping', '-h', 'localhost']
      interval: 10s 
      timeout: 5s 
      retries: 3 
 
  redis: 
    image: redis:7.0-alpine 
    container_name: ilock_redis 
    restart: always 
    ports: 
      - '6380:6379'
    volumes: 
      - redis_data:/data 
    networks: 
      - ilock_network 
    healthcheck: 
      test: ['CMD', 'redis-cli', 'ping']
      interval: 10s 
      timeout: 5s 
      retries: 3 
       
  mqtt: 
    image: eclipse-mosquitto:2.0 
    container_name: ilock_mqtt 
    restart: always 
    ports: 
      - '1883:1883'
      - '8883:8883'
      - '9001:9001'
    volumes: 
      - ./mqtt/config:/mosquitto/config 
      - ./mqtt/data:/mosquitto/data 
      - ./mqtt/log:/mosquitto/log 
    networks: 
      - ilock_network 
    healthcheck: 
      test: ['CMD', 'mosquitto_sub', '-t', '$$SYS/#', '-C', '1', '-i', 'healthcheck', '-W', '3']
      interval: 10s 
      timeout: 5s 
      retries: 3 
 
networks: 
  ilock_network: 
    driver: bridge 
 
volumes: 
  mysql_data: 
  redis_data:  
EOF

# 准备MQTT配置
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

# 拷贝文件到服务器
print_info "复制部署文件到服务器..."
scp_cmd docker-compose.modified.yml .env mqtt/config/mosquitto.conf backup_script.sh

# 服务器上执行部署
print_info "执行部署前准备工作..."
ssh_cmd "cd /root/ilock && mkdir -p mqtt/config mqtt/data mqtt/log backups && chmod +x backup_script.sh"

# 创建数据库备份（如果启用）
if [ "$BACKUP_ENABLED" = true ]; then
  print_info "创建数据库备份..."
  ssh_cmd "cd /root/ilock && ./backup_script.sh"
  
  if [ $? -ne 0 ]; then
    print_warning "备份过程可能存在问题，但将继续部署..."
  else
    print_success "数据库备份完成！"
  fi
else
  print_warning "数据库备份功能已禁用，跳过备份步骤"
fi

# 部署新版本
ssh_cmd "cd /root/ilock && mv docker-compose.modified.yml docker-compose.yml"
ssh_cmd "cd /root/ilock && chmod 644 mqtt/config/mosquitto.conf"
ssh_cmd "cd /root/ilock && docker-compose down && docker-compose pull && docker-compose up -d"

# 检查服务状态
print_info "检查服务状态..."
ssh_cmd "cd /root/ilock && for i in {1..30}; do if docker-compose ps | grep -q 'Up'; then echo '服务已启动！'; docker-compose ps; break; fi; echo '等待服务启动... (尝试 '\$i'/30)'; sleep 2; done"

# 执行数据库迁移
if [ "$AUTO_MIGRATE" = true ]; then
  print_info "检查数据库迁移脚本..."
  if ssh_cmd "cd /root/ilock && docker exec ilock_http_service ls -la /app/run_migrations.sh 2>/dev/null"; then
    print_info "执行数据库迁移..."
    ssh_cmd "cd /root/ilock && docker exec ilock_http_service /app/run_migrations.sh"
    
    if [ $? -ne 0 ]; then
      print_error "数据库迁移失败！请检查日志获取详细信息。"
    else
      print_success "数据库迁移成功完成。"
    fi
  else
    print_warning "未找到迁移脚本，跳过数据库迁移步骤。"
  fi
else
  print_warning "自动数据库迁移功能已禁用。"
fi

# 检查应用是否正常
ssh_cmd "cd /root/ilock && curl -s http://localhost:20033/api/ping > /dev/null 2>&1 && echo '应用运行正常！' || echo '应用启动失败！'"

if [ $? -eq 0 ]; then
  print_success "部署成功！当前版本: $VERSION"
  print_info "已创建备份，可在服务器的 /root/ilock/backups 目录查看"
else
  print_error "部署可能存在问题。请检查日志: ssh $SSH_USERNAME@$SSH_HOST 'cd /root/ilock && docker-compose logs'"
fi

# 创建简易回滚脚本
cat > rollback.sh << EOF
#!/bin/bash
# iLock回滚脚本

SSH_HOST="$SSH_HOST"
SSH_PORT="$SSH_PORT"
SSH_USERNAME="$SSH_USERNAME"
SSH_PASSWORD="$SSH_PASSWORD"

if [ \$# -ne 1 ]; then
  echo "用法: \$0 <回滚到的版本>"
  echo "例如: \$0 1.2.0"
  exit 1
fi

TARGET_VERSION=\$1
echo "正在回滚到版本 \$TARGET_VERSION..."

export SSHPASS="$SSH_PASSWORD"
sshpass -e ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USERNAME@$SSH_HOST" "cd /root/ilock && sed -i 's|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:\$TARGET_VERSION|' docker-compose.yml && docker-compose pull && docker-compose up -d"

echo "回滚完成！"
EOF

chmod +x rollback.sh
print_info "已创建回滚脚本 rollback.sh，可在需要时使用"

# 清理临时文件
rm -f docker-compose.modified.yml backup_script.sh 