#!/bin/bash
# iLock服务 - 数据备份脚本

# 服务器配置
SSH_HOST=${SSH_HOST:-"39.108.49.167"}
SSH_PORT=${SSH_PORT:-"22"}
SSH_USER=${SSH_USER:-"root"}
SSH_PASS=${SSH_PASS:-"1090119your@"}
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 颜色输出
function print_info() { echo -e "\033[0;34m[INFO] $1\033[0m"; }
function print_success() { echo -e "\033[0;32m[SUCCESS] $1\033[0m"; }
function print_error() { echo -e "\033[0;31m[ERROR] $1\033[0m"; }

# 检查sshpass
if ! command -v sshpass &> /dev/null; then
    print_error "未安装sshpass，请先安装："
    echo "  macOS: brew install hudochenkov/sshpass/sshpass"
    echo "  Linux: sudo apt-get install sshpass"
    exit 1
fi

# 创建备份目录
mkdir -p $BACKUP_DIR

# 定义SSH和SCP命令
function ssh_cmd() {
    export SSHPASS="$SSH_PASS"
    sshpass -e ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" "$@"
}

function scp_from() {
    export SSHPASS="$SSH_PASS"
    sshpass -e scp -o StrictHostKeyChecking=no -P "$SSH_PORT" "$SSH_USER@$SSH_HOST:$1" "$2"
}

# 在服务器上创建备份脚本
print_info "创建服务器端备份脚本..."
cat > remote_backup.sh << 'EOF'
#!/bin/bash
# 服务器端备份脚本

# 备份目录
BACKUP_DIR="/root/ilock/backups"
mkdir -p $BACKUP_DIR

# 当前时间戳
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
MYSQL_BACKUP="$BACKUP_DIR/ilock_db_$TIMESTAMP.sql"
REDIS_BACKUP="$BACKUP_DIR/redis_dump_$TIMESTAMP.rdb"

echo "开始备份MySQL数据..."
if docker ps | grep -q ilock_mysql; then
    docker exec ilock_mysql sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --all-databases' > "$MYSQL_BACKUP"
    
    if [ $? -eq 0 ]; then
        echo "MySQL备份成功: $MYSQL_BACKUP"
        gzip "$MYSQL_BACKUP"
        echo "MySQL备份已压缩: $MYSQL_BACKUP.gz"
    else
        echo "MySQL备份失败!"
        exit 1
    fi
else
    echo "MySQL容器未运行，无法备份"
    exit 1
fi

echo "开始备份Redis数据..."
if docker ps | grep -q ilock_redis; then
    docker exec ilock_redis sh -c 'redis-cli save && cat /data/dump.rdb' > "$REDIS_BACKUP"
    
    if [ $? -eq 0 ]; then
        echo "Redis备份成功: $REDIS_BACKUP"
        gzip "$REDIS_BACKUP"
        echo "Redis备份已压缩: $REDIS_BACKUP.gz"
    else
        echo "Redis备份失败!"
        exit 1
    fi
else
    echo "Redis容器未运行，无法备份"
    exit 1
fi

# 备份配置文件
echo "备份配置文件..."
cp /root/ilock/.env "$BACKUP_DIR/env_$TIMESTAMP"

# 备份docker-compose文件
cp /root/ilock/docker-compose.yml "$BACKUP_DIR/docker-compose_$TIMESTAMP.yml"

# 备份configs目录
if [ -d "/root/ilock/configs" ]; then
    echo "备份configs目录..."
    tar -czf "$BACKUP_DIR/configs_$TIMESTAMP.tar.gz" -C /root/ilock configs
fi

# 备份MQTT配置
if [ -d "/root/ilock/mqtt" ]; then
    echo "备份MQTT数据目录..."
    tar -czf "$BACKUP_DIR/mqtt_$TIMESTAMP.tar.gz" -C /root/ilock mqtt
fi

echo "备份完成，文件保存在 $BACKUP_DIR"
echo "MYSQL_BACKUP=ilock_db_$TIMESTAMP.sql.gz" > "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
echo "REDIS_BACKUP=redis_dump_$TIMESTAMP.rdb.gz" >> "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
echo "ENV_BACKUP=env_$TIMESTAMP" >> "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
echo "COMPOSE_BACKUP=docker-compose_$TIMESTAMP.yml" >> "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
echo "CONFIGS_BACKUP=configs_$TIMESTAMP.tar.gz" >> "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
echo "MQTT_BACKUP=mqtt_$TIMESTAMP.tar.gz" >> "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
echo "TIMESTAMP=$TIMESTAMP" >> "$BACKUP_DIR/backup_info_$TIMESTAMP.txt"
EOF

# 上传备份脚本到服务器
print_info "上传备份脚本到服务器..."
ssh_cmd "mkdir -p /root/ilock/scripts"
export SSHPASS="$SSH_PASS"
sshpass -e scp -o StrictHostKeyChecking=no -P "$SSH_PORT" remote_backup.sh "$SSH_USER@$SSH_HOST:/root/ilock/scripts/backup.sh"
rm remote_backup.sh

# 执行备份
print_info "执行服务器备份..."
ssh_cmd "chmod +x /root/ilock/scripts/backup.sh && cd /root/ilock && ./scripts/backup.sh"
if [ $? -ne 0 ]; then
    print_error "备份失败！"
    exit 1
fi

# 下载备份信息文件
print_info "查找最新备份信息..."
LATEST_BACKUP_INFO=$(ssh_cmd "ls -t /root/ilock/backups/backup_info_*.txt | head -1")
if [ -z "$LATEST_BACKUP_INFO" ]; then
    print_error "未找到备份信息文件！"
    exit 1
fi

print_info "下载备份信息文件: $LATEST_BACKUP_INFO"
scp_from "$LATEST_BACKUP_INFO" "$BACKUP_DIR/backup_info.txt"

# 读取备份信息
source "$BACKUP_DIR/backup_info.txt"

# 下载备份文件
print_info "下载MySQL备份: $MYSQL_BACKUP"
scp_from "/root/ilock/backups/$MYSQL_BACKUP" "$BACKUP_DIR/"

print_info "下载Redis备份: $REDIS_BACKUP"
scp_from "/root/ilock/backups/$REDIS_BACKUP" "$BACKUP_DIR/"

print_info "下载环境配置: $ENV_BACKUP"
scp_from "/root/ilock/backups/$ENV_BACKUP" "$BACKUP_DIR/.env"

print_info "下载Docker配置: $COMPOSE_BACKUP"
scp_from "/root/ilock/backups/$COMPOSE_BACKUP" "$BACKUP_DIR/docker-compose.yml"

if [ ! -z "$CONFIGS_BACKUP" ]; then
    print_info "下载configs目录备份: $CONFIGS_BACKUP"
    scp_from "/root/ilock/backups/$CONFIGS_BACKUP" "$BACKUP_DIR/"
    mkdir -p "$BACKUP_DIR/configs"
    tar -xzf "$BACKUP_DIR/$CONFIGS_BACKUP" -C "$BACKUP_DIR"
fi

if [ ! -z "$MQTT_BACKUP" ]; then
    print_info "下载MQTT数据目录备份: $MQTT_BACKUP"
    scp_from "/root/ilock/backups/$MQTT_BACKUP" "$BACKUP_DIR/"
    mkdir -p "$BACKUP_DIR/mqtt"
    tar -xzf "$BACKUP_DIR/$MQTT_BACKUP" -C "$BACKUP_DIR"
fi

print_success "备份完成！所有文件已保存在 $BACKUP_DIR 目录"
echo "备份时间戳: $TIMESTAMP" 