#!/bin/bash
# iLock 系统迁移脚本

# 版本设置
VERSION="2.3.0"

# 目标服务器设置
TARGET_HOST="117.72.193.54"
TARGET_PORT="22"
TARGET_USERNAME="root"
TARGET_PASSWORD="1090119your@"

# 备份目录设置
BACKUP_DIR="$(dirname "$0")/backup"
LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/*.tar.gz | head -n1)

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

# 检查sshpass是否安装
if ! command -v sshpass &> /dev/null; then
  print_warning "sshpass未安装，将尝试安装..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install sshpass || { 
      print_error "sshpass安装失败！请手动安装: brew install sshpass"; 
      exit 1; 
    }
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sudo apt-get update && sudo apt-get install -y sshpass || { print_error "sshpass安装失败！请手动安装: sudo apt-get install sshpass"; exit 1; }
  else
    print_error "无法识别的操作系统，请手动安装sshpass后重试"; 
    exit 1;
  fi
  print_success "sshpass安装成功"
fi

# 定义SSH和SCP命令的函数，自动使用密码
function ssh_cmd() {
  export SSHPASS="$TARGET_PASSWORD"
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$TARGET_PORT" "$TARGET_USERNAME@$TARGET_HOST" "$@"
}

function scp_cmd() {
  export SSHPASS="$TARGET_PASSWORD"
  sshpass -e scp -o StrictHostKeyChecking=no -P "$TARGET_PORT" "$@" "$TARGET_USERNAME@$TARGET_HOST:/root/ilock/"
}

# 检查备份文件
if [ ! -f "$LATEST_BACKUP" ]; then
  print_error "未找到备份文件！请先运行 backup.sh"
  exit 1
fi

print_info "使用备份文件: $LATEST_BACKUP"

# 创建目标服务器目录
print_info "创建目标服务器目录..."
ssh_cmd "mkdir -p /root/ilock"

# 上传备份文件
print_info "上传备份文件..."
scp_cmd "$LATEST_BACKUP"

# 解压备份文件
print_info "解压备份文件..."
ssh_cmd "cd /root/ilock && tar xzf $(basename "$LATEST_BACKUP")"

# 修复.env文件
print_info "修复.env文件..."
# 只生成新的MQTT_CLIENT_ID并替换
NEW_MQTT_CLIENT_ID="mqttx_$(openssl rand -hex 8)"
echo "新的MQTT_CLIENT_ID: $NEW_MQTT_CLIENT_ID"
ssh_cmd "cd /root/ilock && sed -i 's/^MQTT_CLIENT_ID=.*/MQTT_CLIENT_ID=$NEW_MQTT_CLIENT_ID/' .env"

# 从备份中恢复其他环境变量
print_info "恢复环境变量..."
ssh_cmd "cd /root/ilock && if [ -f env.tar.gz ]; then tar xzf env.tar.gz && if [ -d env ]; then cat env/* >> .env && rm -rf env; fi; fi"

# 创建MQTT配置
print_info "创建MQTT配置..."
ssh_cmd "cd /root/ilock && mkdir -p mqtt/config mqtt/data mqtt/log"

# 创建MQTT配置文件
print_info "创建MQTT配置文件..."
cat > mosquitto.conf << 'EOF'
# MQTT Broker Configuration
listener 1883
protocol mqtt

# Authentication - 允许匿名访问，方便测试
allow_anonymous true

# Persistence
persistence true
persistence_location /mosquitto/data/
persistence_file mosquitto.db

# Logging
log_dest file /mosquitto/log/mosquitto.log
log_type all
connection_messages true
log_timestamp true

# Security
allow_zero_length_clientid false

# Performance
max_queued_messages 1000
max_inflight_messages 20
max_connections 1000

# 在Mosquitto 2.0中，不再使用topic指令，而是使用ACL
# 允许所有用户访问所有主题
acl_file /mosquitto/config/acl.conf
EOF

# 创建ACL文件
cat > acl.conf << 'EOF'
# 允许所有用户访问所有主题
topic readwrite #
EOF

# 上传配置文件
print_info "上传MQTT配置文件..."
scp_cmd mosquitto.conf
scp_cmd acl.conf
ssh_cmd "cd /root/ilock && mkdir -p mqtt/config && mv mosquitto.conf mqtt/config/ && mv acl.conf mqtt/config/"

# 验证配置文件
print_info "验证MQTT配置文件..."
ssh_cmd "cd /root/ilock && ls -la mqtt/config/ && cat mqtt/config/mosquitto.conf"

# 设置MQTT目录权限
print_info "设置MQTT目录权限..."
ssh_cmd "cd /root/ilock && chmod -R 777 mqtt"

# 创建docker-compose.yml文件
print_info "创建docker-compose.yml文件..."
cat > docker-compose.yml << EOF
version: '3.8'
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
      - LOCAL_DB_HOST=db
      - LOCAL_DB_USER=root
      - LOCAL_DB_PASSWORD=\${MYSQL_ROOT_PASSWORD}
      - LOCAL_DB_NAME=\${MYSQL_DATABASE}
      - LOCAL_DB_PORT=3306
      - SERVER_DB_HOST=db
      - SERVER_DB_USER=root
      - SERVER_DB_PASSWORD=\${MYSQL_ROOT_PASSWORD}
      - SERVER_DB_NAME=\${MYSQL_DATABASE}
      - SERVER_DB_PORT=3306
      - MQTT_BROKER_URL=tcp://mqtt:1883
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
      test: ["CMD", "mosquitto_sub", "-t", "\\$SYS/#", "-C", "1", "-i", "healthcheck", "-W", "3"]
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

# 上传docker-compose.yml
print_info "上传docker-compose.yml..."
scp_cmd docker-compose.yml

# 配置Docker镜像加速
print_info "配置Docker镜像加速..."
cat > setup_docker_mirror.sh << 'EOF'
#!/bin/bash

# 创建或更新Docker配置目录
mkdir -p /etc/docker

# 创建daemon.json配置文件
cat > /etc/docker/daemon.json << 'INNEREOF'
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://dockerproxy.net",
    "https://hub.rat.dev",
    "https://docker.1ms.run"
  ]
}
INNEREOF

# 重启Docker服务
systemctl daemon-reload
systemctl restart docker

echo "Docker镜像加速配置完成"
EOF

scp_cmd setup_docker_mirror.sh
ssh_cmd "cd /root/ilock && chmod +x setup_docker_mirror.sh && ./setup_docker_mirror.sh"

# 停止并清理现有容器
print_info "停止并清理现有容器..."
ssh_cmd "cd /root/ilock && docker-compose down -v || true"
ssh_cmd "docker system prune -f || true"

# 拉取镜像
print_info "拉取Docker镜像..."
ssh_cmd "cd /root/ilock && docker-compose pull"

# 启动服务
print_info "启动服务..."
ssh_cmd "cd /root/ilock && docker-compose up -d"

# 等待服务就绪
print_info "等待服务就绪..."
ssh_cmd "cd /root/ilock && for i in {1..60}; do if docker-compose ps | grep -q 'Up (healthy)'; then echo '所有服务已就绪！'; break; fi; if [ \$i -eq 60 ]; then echo '服务启动超时'; docker-compose logs; exit 1; fi; echo '等待服务就绪... (尝试 '\$i'/60)'; sleep 5; done"

# 检查服务状态
print_info "检查服务状态..."
ssh_cmd "cd /root/ilock && docker-compose ps"

# 如果有服务处于Restarting状态，尝试重新启动
print_info "检查是否有服务需要重新启动..."
if ssh_cmd "cd /root/ilock && docker-compose ps | grep -q 'Restarting'"; then
  print_warning "检测到服务重启中，尝试重新启动所有服务..."
  ssh_cmd "cd /root/ilock && docker-compose restart"
  print_info "等待服务重新启动..."
  ssh_cmd "cd /root/ilock && sleep 30 && docker-compose ps"
fi

# 清理临时文件
rm -f setup_docker_mirror.sh docker-compose.yml mosquitto.conf acl.conf

print_success "迁移完成！"
print_info "服务状态："
ssh_cmd "cd /root/ilock && docker-compose ps" 