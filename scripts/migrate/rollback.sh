#!/bin/bash
# iLock服务 - 部署回滚脚本

# 服务器配置
TARGET_HOST=${TARGET_HOST:-"117.72.193.54"}
TARGET_PORT=${TARGET_PORT:-"22"}
TARGET_USER=${TARGET_USER:-"root"}
TARGET_PASS=${TARGET_PASS:-"1090119your@"}

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

# 解析参数
ROLLBACK_VERSION=""

# 如果没有提供版本号，显示帮助
if [ $# -eq 0 ]; then
    print_error "请提供要回滚到的版本号"
    echo "用法: $0 <回滚版本号>"
    echo "示例: $0 1.3.0"
    exit 1
fi

ROLLBACK_VERSION=$1

# 定义SSH命令
function ssh_cmd() {
    export SSHPASS="$TARGET_PASS"
    sshpass -e ssh -o StrictHostKeyChecking=no -p "$TARGET_PORT" "$TARGET_USER@$TARGET_HOST" "$@"
}

# 确认是否要继续
read -p "确定要回滚到版本 $ROLLBACK_VERSION 吗？[y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "回滚操作已取消"
    exit 0
fi

# 检查远程服务器上是否有docker-compose.yml文件
print_info "检查远程服务器状态..."
if ! ssh_cmd "[ -f /root/ilock/docker-compose.yml ]"; then
    print_error "目标服务器上未找到docker-compose.yml文件，无法回滚"
    exit 1
fi

# 备份当前配置
print_info "备份当前配置..."
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ssh_cmd "cd /root/ilock && cp docker-compose.yml docker-compose.yml.bak.$TIMESTAMP"

# 修改版本号
print_info "更新版本号为 $ROLLBACK_VERSION..."
ssh_cmd "cd /root/ilock && sed -i 's|image: stonesea/ilock-http-service:.*|image: stonesea/ilock-http-service:$ROLLBACK_VERSION|g' docker-compose.yml"

# 检查是否有备份目录，如果有，询问是否要恢复配置
BACKUP_EXISTS=$(ssh_cmd "[ -d /root/ilock/backups ] && echo 'yes' || echo 'no'")

if [ "$BACKUP_EXISTS" == "yes" ]; then
    # 检查是否有与版本匹配的配置备份
    CONFIG_BACKUP=$(ssh_cmd "find /root/ilock/backups -name \"configs_*_$ROLLBACK_VERSION.tar.gz\" | head -1")
    
    if [ ! -z "$CONFIG_BACKUP" ]; then
        read -p "找到版本 $ROLLBACK_VERSION 的配置备份，是否要恢复配置？[y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "恢复配置备份..."
            ssh_cmd "cd /root/ilock && tar -xzf $CONFIG_BACKUP -C /root/ilock"
        fi
    else
        print_warning "未找到版本 $ROLLBACK_VERSION 的配置备份，只会更新应用版本"
    fi
fi

# 停止当前服务
print_info "停止当前服务..."
ssh_cmd "cd /root/ilock && docker-compose down"

# 启动回滚版本
print_info "启动回滚版本 $ROLLBACK_VERSION..."
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

# 如果回滚失败，提供恢复选项
if [ $? -ne 0 ]; then
    print_error "回滚操作可能有问题，是否要恢复到回滚前的配置？"
    read -p "恢复到回滚前的配置？[y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "恢复到回滚前的配置..."
        ssh_cmd "cd /root/ilock && cp docker-compose.yml.bak.$TIMESTAMP docker-compose.yml && docker-compose down && docker-compose up -d"
    fi
else
    print_success "回滚操作完成！应用已回滚到版本 $ROLLBACK_VERSION"
fi

# 清理
print_info "清理临时文件..."
ssh_cmd "cd /root/ilock && rm -f docker-compose.yml.bak.$TIMESTAMP" 