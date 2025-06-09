#!/bin/bash
# 数据库备份脚本

# 备份目录
BACKUP_DIR="/root/ilock/backups/migration"
mkdir -p $BACKUP_DIR

# 当前时间戳
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
MYSQL_BACKUP_FILE="$BACKUP_DIR/ilock_db_$TIMESTAMP.sql"
REDIS_BACKUP_FILE="$BACKUP_DIR/redis_dump_$TIMESTAMP.rdb"

# 检查是否存在docker-compose.yml文件
if [ ! -f "/root/ilock/docker-compose.yml" ]; then
  echo "错误: 未找到docker-compose.yml文件"
  exit 1
fi

# 备份MySQL数据
echo "开始备份MySQL数据..."
if docker ps | grep -q ilock_mysql; then
  docker exec ilock_mysql sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --all-databases' > "$MYSQL_BACKUP_FILE"
  
  if [ $? -eq 0 ]; then
    echo "MySQL备份成功: $MYSQL_BACKUP_FILE"
    gzip "$MYSQL_BACKUP_FILE"
    echo "MySQL备份已压缩: $MYSQL_BACKUP_FILE.gz"
  else
    echo "MySQL备份失败!"
    exit 1
  fi
else
  echo "MySQL容器未运行，无法备份"
  exit 1
fi

# 备份Redis数据
echo "开始备份Redis数据..."
if docker ps | grep -q ilock_redis; then
  docker exec ilock_redis sh -c 'redis-cli save && cat /data/dump.rdb' > "$REDIS_BACKUP_FILE"
  
  if [ $? -eq 0 ]; then
    echo "Redis备份成功: $REDIS_BACKUP_FILE"
    gzip "$REDIS_BACKUP_FILE"
    echo "Redis备份已压缩: $REDIS_BACKUP_FILE.gz"
  else
    echo "Redis备份失败!"
    exit 1
  fi
else
  echo "Redis容器未运行，无法备份"
  exit 1
fi

# 备份MQTT配置和数据
echo "开始备份MQTT配置和数据..."
MQTT_BACKUP_DIR="$BACKUP_DIR/mqtt_backup_$TIMESTAMP"
mkdir -p "$MQTT_BACKUP_DIR"

if [ -d "/root/ilock/mqtt" ]; then
  cp -r /root/ilock/mqtt/* "$MQTT_BACKUP_DIR/"
  if [ $? -eq 0 ]; then
    echo "MQTT配置和数据备份成功: $MQTT_BACKUP_DIR"
    tar -czf "$MQTT_BACKUP_DIR.tar.gz" -C "$BACKUP_DIR" "mqtt_backup_$TIMESTAMP"
    echo "MQTT备份已压缩: $MQTT_BACKUP_DIR.tar.gz"
    rm -rf "$MQTT_BACKUP_DIR"
  else
    echo "MQTT备份失败!"
  fi
else
  echo "MQTT目录不存在，跳过备份"
fi

# 备份.env文件
if [ -f "/root/ilock/.env" ]; then
  cp /root/ilock/.env "$BACKUP_DIR/env_backup_$TIMESTAMP"
  echo ".env文件备份成功: $BACKUP_DIR/env_backup_$TIMESTAMP"
else
  echo ".env文件不存在，无法备份"
fi

# 备份docker-compose.yml文件
if [ -f "/root/ilock/docker-compose.yml" ]; then
  cp /root/ilock/docker-compose.yml "$BACKUP_DIR/docker-compose_$TIMESTAMP.yml"
  echo "docker-compose.yml文件备份成功: $BACKUP_DIR/docker-compose_$TIMESTAMP.yml"
else
  echo "docker-compose.yml文件不存在，无法备份"
fi

echo "全部备份完成！备份文件存放在: $BACKUP_DIR"
