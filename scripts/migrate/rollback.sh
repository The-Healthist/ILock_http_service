#!/bin/bash
# iLock 系统回滚脚本

# 版本设置
CURRENT_VERSION="2.3.0"  # 当前版本
ROLLBACK_VERSION="2.2.0" # 要回滚到的版本

# 目标服务器设置
TARGET_HOST="目标服务器IP"
TARGET_PORT="22"
TARGET_USERNAME="root"
TARGET_PASSWORD="密码"

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
  export SSHPASS="$TARGET_PASSWORD"
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$TARGET_PORT" "$TARGET_USERNAME@$TARGET_HOST" "$@"
}

# 确认回滚
print_warning "即将从版本 $CURRENT_VERSION 回滚到版本 $ROLLBACK_VERSION"
read -p "是否继续？(y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  print_info "回滚已取消"
  exit 1
fi

# 停止服务
print_info "停止服务..."
ssh_cmd "cd /root/ilock && docker-compose down"

# 修改docker-compose.yml中的版本
print_info "更新docker-compose.yml中的版本..."
ssh_cmd "cd /root/ilock && sed -i 's/image: stonesea\/ilock-http-service:$CURRENT_VERSION/image: stonesea\/ilock-http-service:$ROLLBACK_VERSION/' docker-compose.yml"

# 拉取旧版本镜像
print_info "拉取旧版本镜像..."
ssh_cmd "cd /root/ilock && docker-compose pull"

# 启动服务
print_info "启动服务..."
ssh_cmd "cd /root/ilock && docker-compose up -d"

# 等待服务就绪
print_info "等待服务就绪..."
ssh_cmd "cd /root/ilock && for i in {1..30}; do if docker-compose ps | grep -q 'Up (healthy)'; then echo '所有服务已就绪！'; break; fi; if [ \$i -eq 30 ]; then echo '服务启动超时'; docker-compose logs; exit 1; fi; echo '等待服务就绪... (尝试 '\$i'/30)'; sleep 2; done"

# 检查服务状态
print_info "检查服务状态..."
ssh_cmd "cd /root/ilock && docker-compose ps"

print_success "回滚完成！"
print_info "服务状态："
ssh_cmd "cd /root/ilock && docker-compose ps" 