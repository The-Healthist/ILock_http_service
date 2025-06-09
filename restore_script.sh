#!/bin/bash
# 数据恢复脚本

# 获取最新备份文件
MYSQL_BACKUP=$(ls -t /root/ilock/backups/ilock_db_*.sql.gz | head -n 1)
REDIS_BACKUP=$(ls -t /root/ilock/backups/redis_dump_*.rdb.gz | head -n 1)
MQTT_BACKUP=$(ls -t /root/ilock/backups/mqtt_backup_*.tar.gz | head -n 1)

if [ -z "$MYSQL_BACKUP" ] || [ -z "$REDIS_BACKUP" ]; then
  echo "错误: 未找到备份文件!"
  exit 1
fi

# 解压MySQL备份
echo "解压MySQL备份文件: $MYSQL_BACKUP"
gunzip -c "$MYSQL_BACKUP" > /root/ilock/mysql_backup.sql

# 解压Redis备份
echo "解压Redis备份文件: $REDIS_BACKUP"
gunzip -c "$REDIS_BACKUP" > /root/ilock/redis_dump.rdb

# 解压MQTT备份
if [ -n "$MQTT_BACKUP" ]; then
  echo "解压MQTT备份文件: $MQTT_BACKUP"
  mkdir -p /root/ilock/mqtt_temp
  tar -xzf "$MQTT_BACKUP" -C /root/ilock/mqtt_temp
  cp -r /root/ilock/mqtt_temp/* /root/ilock/mqtt/
  rm -rf /root/ilock/mqtt_temp
fi

# 启动数据库但不启动应用
echo "启动MySQL和Redis容器..."
cd /root/ilock
docker-compose up -d db redis mqtt

# 等待MySQL就绪 - 增加重试和超时
echo "等待MySQL就绪..."
max_attempts=60
for i in $(seq 1 $max_attempts); do
  # 先检查容器是否存在
  if docker ps | grep -q ilock_mysql; then
    # 然后检查MySQL是否可以响应
    if docker exec ilock_mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
      echo "MySQL已就绪!"
      break
    fi
  fi
  
  if [ $i -eq $max_attempts ]; then
    echo "MySQL启动超时! 查看日志："
    docker-compose logs db
    exit 1
  fi
  
  echo "等待MySQL启动... (第$i次尝试/共$max_attempts次)"
  sleep 5
done

# 恢复MySQL数据
echo "恢复MySQL数据..."
cat /root/ilock/mysql_backup.sql | docker exec -i ilock_mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" || {
  echo "MySQL数据恢复失败! 可能是因为版本不兼容，尝试替代方案..."
  
  # 尝试修复常见的版本兼容性问题
  sed 's/ROW_FORMAT=DYNAMIC/ROW_FORMAT=COMPACT/g' /root/ilock/mysql_backup.sql > /root/ilock/mysql_backup_fixed.sql
  
  # 重试导入
  cat /root/ilock/mysql_backup_fixed.sql | docker exec -i ilock_mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD"
  
  if [ $? -eq 0 ]; then
    echo "MySQL数据修复并恢复成功!"
  else
    echo "MySQL数据恢复失败!"
    exit 1
  fi
}

# 等待Redis就绪 - 增加重试和超时
echo "等待Redis就绪..."
max_attempts=60
for i in $(seq 1 $max_attempts); do
  # 先检查容器是否存在
  if docker ps | grep -q ilock_redis; then
    # 然后检查Redis是否可以响应
    if docker exec ilock_redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
      echo "Redis已就绪!"
      break
    fi
  fi
  
  if [ $i -eq $max_attempts ]; then
    echo "Redis启动超时! 查看日志："
    docker-compose logs redis
    exit 1
  fi
  
  echo "等待Redis启动... (第$i次尝试/共$max_attempts次)"
  sleep 5
done

# 恢复Redis数据
echo "恢复Redis数据..."
docker stop ilock_redis
if [ -f "/root/ilock/redis_dump.rdb" ]; then
  # 获取Redis数据目录
  REDIS_DATA_DIR=$(docker volume inspect ilock_redis_data -f '{{.Mountpoint}}' 2>/dev/null || echo "/root/ilock/redis_data")
  
  # 确保目录存在
  mkdir -p "$REDIS_DATA_DIR"
  
  # 复制RDB文件
  cp /root/ilock/redis_dump.rdb "$REDIS_DATA_DIR/dump.rdb"
  chmod 644 "$REDIS_DATA_DIR/dump.rdb"
  
  docker start ilock_redis
  sleep 3
  
  # 验证Redis数据是否成功加载
  if docker exec ilock_redis redis-cli info keyspace | grep -q "keys="; then
    echo "Redis数据恢复成功!"
  else
    echo "Redis数据可能未完全恢复，继续尝试启动服务..."
  fi
else
  echo "Redis备份文件不存在!"
  docker start ilock_redis
fi

# 启动剩余服务
echo "启动应用服务..."
docker-compose up -d

# 清理临时文件
rm -f /root/ilock/mysql_backup.sql /root/ilock/mysql_backup_fixed.sql /root/ilock/redis_dump.rdb

echo "数据恢复完成!"
