#!/bin/bash
# iLock 数据备份脚本

# 源服务器设置
SOURCE_HOST="39.108.49.167"
SOURCE_PORT="22"
SOURCE_USERNAME="root"
SOURCE_PASSWORD="1090119your@"

# 备份目录设置
BACKUP_DIR="$(dirname "$0")/backup"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="ilock_backup_${TIMESTAMP}"

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

# 定义SSH命令的函数，自动使用密码
function ssh_cmd() {
  export SSHPASS="$SOURCE_PASSWORD"
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$SOURCE_PORT" "$SOURCE_USERNAME@$SOURCE_HOST" "$@"
}

# 创建备份目录
mkdir -p "$BACKUP_DIR"
print_info "创建备份目录: $BACKUP_DIR"

# 停止服务
print_info "停止源服务器上的服务..."
ssh_cmd "cd /root/ilock && docker-compose down"

# 备份MySQL数据
print_info "备份MySQL数据..."
ssh_cmd "cd /root/ilock && docker run --rm -v ilock_mysql_data:/source -v /tmp:/backup alpine tar czf /backup/mysql_data.tar.gz -C /source ."
ssh_cmd "cd /tmp && tar czf mysql_data.tar.gz mysql_data.tar.gz"

# 备份Redis数据
print_info "备份Redis数据..."
ssh_cmd "cd /root/ilock && docker run --rm -v ilock_redis_data:/source -v /tmp:/backup alpine tar czf /backup/redis_data.tar.gz -C /source ."
ssh_cmd "cd /tmp && tar czf redis_data.tar.gz redis_data.tar.gz"

# 备份环境配置文件
print_info "备份环境配置文件..."
ssh_cmd "cd /root/ilock && tar czf /tmp/env.tar.gz .env"

# 备份MQTT数据
print_info "备份MQTT数据..."
ssh_cmd "cd /root/ilock && tar czf /tmp/mqtt.tar.gz mqtt/"

# 创建备份包
print_info "创建备份包..."
ssh_cmd "cd /tmp && tar czf ${BACKUP_NAME}.tar.gz mysql_data.tar.gz redis_data.tar.gz env.tar.gz mqtt.tar.gz"

# 下载备份包
print_info "下载备份包..."
export SSHPASS="$SOURCE_PASSWORD"
sshpass -e scp -o StrictHostKeyChecking=no -P "$SOURCE_PORT" "$SOURCE_USERNAME@$SOURCE_HOST:/tmp/${BACKUP_NAME}.tar.gz" "$BACKUP_DIR/"

# 清理临时文件
print_info "清理临时文件..."
ssh_cmd "rm -f /tmp/mysql_data.tar.gz /tmp/redis_data.tar.gz /tmp/env.tar.gz /tmp/mqtt.tar.gz /tmp/${BACKUP_NAME}.tar.gz"

# 重启服务
print_info "重启源服务器上的服务..."
ssh_cmd "cd /root/ilock && docker-compose up -d"

# 验证备份
if [ -f "$BACKUP_DIR/${BACKUP_NAME}.tar.gz" ]; then
  print_success "备份完成！"
  print_info "备份文件位置: $BACKUP_DIR/${BACKUP_NAME}.tar.gz"
else
  print_error "备份失败！"
  exit 1
fi

# 显示备份文件大小
BACKUP_SIZE=$(du -h "$BACKUP_DIR/${BACKUP_NAME}.tar.gz" | cut -f1)
print_info "备份文件大小: $BACKUP_SIZE" 