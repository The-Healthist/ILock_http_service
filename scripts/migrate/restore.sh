#!/bin/bash
# iLock服务 - 数据恢复脚本

# 服务器配置
TARGET_HOST=${TARGET_HOST:-"117.72.193.54"}
TARGET_PORT=${TARGET_PORT:-"22"}
TARGET_USER=${TARGET_USER:-"root"}
TARGET_PASS=${TARGET_PASS:-"1090119your@"}
BACKUP_DIR="./backups"
VERSION=${VERSION:-"1.4.0"}  # 指定要恢复的版本号

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

# 检查目标服务器上的应用状态
print_info "检查目标服务器上的应用状态..."
if ssh_cmd "cd /root/ilock && [ -f docker-compose.yml ] && docker-compose ps | grep -q 'Up'"; then
    print_warning "目标服务器上的应用正在运行，将先停止..."
    ssh_cmd "cd /root/ilock && docker-compose down"
fi

# 上传MySQL备份
print_info "上传MySQL备份..."
ssh_cmd "mkdir -p /root/ilock/backups"
scp_to "$BACKUP_DIR/$MYSQL_BACKUP" "/root/ilock/backups/"

# 上传Redis备份
print_info "上传Redis备份..."
scp_to "$BACKUP_DIR/$REDIS_BACKUP" "/root/ilock/backups/"

# 确保容器目录存在
print_info "准备容器数据目录..."
ssh_cmd "mkdir -p /root/ilock/configs /root/ilock/mysql_data /root/ilock/redis_data"

# 上传configs目录
if [ -d "$BACKUP_DIR/configs" ]; then
    print_info "上传configs目录..."
    
    # 使用tar打包并通过ssh传输
    tar -czf "$BACKUP_DIR/configs.tar.gz" -C "$BACKUP_DIR" configs
    scp_to "$BACKUP_DIR/configs.tar.gz" "/root/ilock/"
    ssh_cmd "cd /root/ilock && tar -xzf configs.tar.gz && rm configs.tar.gz"
else
    print_warning "configs目录不存在，将使用默认配置"
fi

# 创建恢复脚本
print_info "创建恢复脚本..."
cat > restore_remote.sh << 'EOF'
#!/bin/bash
# 远程恢复脚本

# 备份信息文件
BACKUP_INFO_FILE="/root/ilock/backups/backup_info.txt"
if [ ! -f "$BACKUP_INFO_FILE" ]; then
    echo "备份信息文件不存在！"
    exit 1
fi

# 读取备份信息
source "$BACKUP_INFO_FILE"

# 确保MySQL和Redis容器不在运行
echo "确保容器已停止..."
cd /root/ilock
docker-compose down > /dev/null 2>&1

# 恢复MySQL数据
echo "恢复MySQL数据..."
# 解压备份文件
gunzip -c /root/ilock/backups/$MYSQL_BACKUP > /root/ilock/backups/mysql_restore.sql

# 启动临时MySQL容器
echo "启动临时MySQL容器..."
docker run --name temp_mysql -e MYSQL_ROOT_PASSWORD=root -v /root/ilock/mysql_data:/var/lib/mysql -d mysql:8.0

# 等待MySQL启动
echo "等待MySQL启动..."
sleep 15

# 导入备份
echo "导入数据库备份..."
cat /root/ilock/backups/mysql_restore.sql | docker exec -i temp_mysql mysql -uroot -proot

# 停止并移除临时容器
echo "清理临时MySQL容器..."
docker stop temp_mysql
docker rm temp_mysql

# 恢复Redis数据
echo "恢复Redis数据..."
mkdir -p /root/ilock/redis_data
gunzip -c /root/ilock/backups/$REDIS_BACKUP > /root/ilock/redis_data/dump.rdb
chmod 644 /root/ilock/redis_data/dump.rdb

echo "数据恢复完成！"
EOF

# 上传备份信息
print_info "上传备份信息..."
scp_to "$BACKUP_DIR/backup_info.txt" "/root/ilock/backups/"

# 上传恢复脚本到服务器
print_info "上传恢复脚本..."
scp_to "restore_remote.sh" "/root/ilock/restore.sh"
rm restore_remote.sh

# 执行恢复
print_info "执行数据恢复..."
ssh_cmd "chmod +x /root/ilock/restore.sh && cd /root/ilock && ./restore.sh"

# 修改docker-compose.yml中的版本号
print_info "更新docker-compose.yml文件中的版本号为 $VERSION..."
ssh_cmd "cd /root/ilock && sed -i 's|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:$VERSION|g' docker-compose.yml"

# 重启服务
print_info "重启服务..."
ssh_cmd "cd /root/ilock && docker-compose pull && docker-compose up -d"

# 检查服务状态
print_info "等待服务启动..."
for i in {1..30}; do
    if ssh_cmd "cd /root/ilock && docker-compose ps | grep 'Up' | wc -l" | grep -q '[1-9]'; then
        print_success "服务已启动，检查服务状态..."
        ssh_cmd "cd /root/ilock && docker-compose ps"
        break
    fi
    
    if [ $i -eq 30 ]; then
        print_error "服务启动超时，请检查日志："
        ssh_cmd "cd /root/ilock && docker-compose logs"
        exit 1
    fi
    
    echo -n "."
    sleep 2
done

# 验证应用是否响应
print_info "验证应用响应..."
for i in {1..30}; do
    if ssh_cmd "curl -s http://localhost:20033/api/ping" > /dev/null 2>&1; then
        print_success "应用响应成功！"
        break
    fi
    
    if [ $i -eq 30 ]; then
        print_warning "应用响应超时，可能需要更长时间启动"
        print_info "查看应用日志："
        ssh_cmd "cd /root/ilock && docker-compose logs app"
    fi
    
    echo -n "."
    sleep 2
done

print_success "恢复完成！应用已恢复到版本 $VERSION" 