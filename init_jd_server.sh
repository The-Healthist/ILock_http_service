#!/bin/bash
# 京东云服务器初始化脚本 - 用于重装Ubuntu 22.04后的环境配置
# 作者: Claude
# 日期: 2024年7月4日

# 颜色输出函数
function print_info() { echo -e "\033[0;34m[INFO] $1\033[0m"; }
function print_success() { echo -e "\033[0;32m[SUCCESS] $1\033[0m"; }
function print_error() { echo -e "\033[0;31m[ERROR] $1\033[0m"; }
function print_warning() { echo -e "\033[0;33m[WARNING] $1\033[0m"; }

# 服务器信息
JD_SSH_HOST="117.72.193.54"
JD_SSH_PORT="22"
JD_SSH_USERNAME="root"
JD_SSH_PASSWORD="1090119your@"

# 是否允许交互式输入IP
INTERACTIVE_IP=true

# 配置
MAX_RETRIES=3
SSH_TIMEOUT=30
SSH_CONNECT_TIMEOUT=60

# 如果开启交互式输入，询问用户是否需要修改IP
if [ "$INTERACTIVE_IP" = true ]; then
  echo "当前京东云服务器IP: $JD_SSH_HOST"
  read -p "是否需要更新IP地址? (y/n): " update_ip
  if [[ "$update_ip" =~ ^[Yy]$ ]]; then
    read -p "请输入新的IP地址: " new_ip
    if [[ -n "$new_ip" ]]; then
      JD_SSH_HOST="$new_ip"
      print_info "已更新IP地址为: $JD_SSH_HOST"
    fi
  fi
fi

# 检查sshpass是否安装
if ! command -v sshpass &> /dev/null; then
  print_warning "sshpass未安装，将尝试安装..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS系统
    brew install sshpass || { 
      print_error "sshpass安装失败！请手动安装: brew install sshpass"; 
      exit 1; 
    }
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux系统
    sudo apt-get update && sudo apt-get install -y sshpass || { print_error "sshpass安装失败！"; exit 1; }
  else
    print_error "无法识别的操作系统，请手动安装sshpass后重试"; 
    exit 1;
  fi
  print_success "sshpass安装成功"
fi

# 定义SSH命令函数，带重试机制
function ssh_cmd() {
  local retries=0
  local result=1
  
  while [ $retries -lt $MAX_RETRIES ] && [ $result -ne 0 ]; do
    if [ $retries -gt 0 ]; then
      print_warning "SSH连接失败，第 $retries 次重试..."
      sleep 5
    fi
    
    export SSHPASS="$JD_SSH_PASSWORD"
    sshpass -e ssh -o StrictHostKeyChecking=accept-new -o ConnectTimeout=$SSH_CONNECT_TIMEOUT -p "$JD_SSH_PORT" "$JD_SSH_USERNAME@$JD_SSH_HOST" "$@"
    result=$?
    retries=$((retries + 1))
  done
  
  if [ $result -ne 0 ]; then
    print_error "执行命令失败，已重试 $MAX_RETRIES 次"
    return 1
  fi
  
  return 0
}

# 定义SCP命令函数，带重试机制
function scp_cmd() {
  local retries=0
  local result=1
  
  while [ $retries -lt $MAX_RETRIES ] && [ $result -ne 0 ]; do
    if [ $retries -gt 0 ]; then
      print_warning "SCP传输失败，第 $retries 次重试..."
      sleep 5
    fi
    
    export SSHPASS="$JD_SSH_PASSWORD"
    sshpass -e scp -o StrictHostKeyChecking=accept-new -o ConnectTimeout=$SSH_CONNECT_TIMEOUT -P "$JD_SSH_PORT" "$1" "$JD_SSH_USERNAME@$JD_SSH_HOST:$2"
    result=$?
    retries=$((retries + 1))
  done
  
  if [ $result -ne 0 ]; then
    print_error "文件传输失败，已重试 $MAX_RETRIES 次"
    return 1
  fi
  
  return 0
}

# 检查IP是否可ping通
print_info "检查IP $JD_SSH_HOST 是否可达..."
print_info "执行: ping -c 1 -W 3 $JD_SSH_HOST"

if ! ping -c 1 -W 3 $JD_SSH_HOST > /dev/null 2>&1; then
  print_warning "无法ping通服务器IP: $JD_SSH_HOST"
  print_warning "服务器可能阻止了ICMP ping，将直接尝试SSH连接..."
  # 不退出，继续尝试SSH连接
else
  print_success "IP $JD_SSH_HOST 可达"
fi

# 检查端口连接
print_info "检查SSH端口连接..."
if nc -z -w 5 $JD_SSH_HOST $JD_SSH_PORT 2>/dev/null; then
  print_success "端口 $JD_SSH_PORT 连接成功"
else
  print_error "无法连接到端口 $JD_SSH_PORT"
  print_warning "请检查服务器防火墙设置"
  exit 1
fi

# 检查连接
print_info "检查与京东云服务器的SSH连接..."
if ! ssh_cmd "echo 连接成功"; then
  print_error "无法连接到服务器，SSH连接失败"
  print_warning "请检查以下可能的问题:"
  print_warning "1. 用户名或密码是否正确"
  print_warning "2. 服务器是否允许密码登录"
  print_warning "3. 是否有防火墙或安全组限制"
  exit 1
fi
print_success "SSH连接成功"

# 步骤1: 更新系统包
print_info "步骤1: 更新系统包..."
if ! ssh_cmd "apt-get update && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y"; then
  print_error "系统包更新失败"
  print_warning "尝试修复可能的锁文件问题..."
  ssh_cmd "rm -f /var/lib/dpkg/lock* /var/lib/apt/lists/lock* /var/cache/apt/archives/lock*"
  print_warning "重试系统更新..."
  if ! ssh_cmd "apt-get update && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y"; then
    print_error "系统包更新再次失败，继续下一步"
  fi
else
  print_success "系统包更新完成"
fi

# 步骤2: 安装基本工具
print_info "步骤2: 安装基本工具..."
if ! ssh_cmd "DEBIAN_FRONTEND=noninteractive apt-get install -y curl wget git vim net-tools htop iotop"; then
  print_error "基本工具安装失败"
  print_warning "尝试修复可能的锁文件问题..."
  ssh_cmd "rm -f /var/lib/dpkg/lock* /var/lib/apt/lists/lock* /var/cache/apt/archives/lock*"
  print_warning "重试安装基本工具..."
  if ! ssh_cmd "DEBIAN_FRONTEND=noninteractive apt-get install -y curl wget git vim net-tools htop iotop"; then
    print_error "基本工具安装再次失败，继续下一步"
  fi
else
  print_success "基本工具安装完成"
fi

# 步骤3: 安装Docker
print_info "步骤3: 安装Docker..."
if ! ssh_cmd "curl -fsSL https://get.docker.com | sh"; then
  print_error "Docker安装失败"
  print_warning "尝试替代方法安装Docker..."
  ssh_cmd "DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io"
fi

# 确认Docker安装并启用
if ! ssh_cmd "systemctl enable docker && systemctl start docker && docker --version"; then
  print_error "Docker启动失败，请手动检查"
else
  print_success "Docker安装完成"
fi

# 步骤4: 安装Docker Compose
print_info "步骤4: 安装Docker Compose..."
if ! ssh_cmd 'curl -L "https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose'; then
  print_error "Docker Compose下载失败"
  print_warning "尝试替代方法安装Docker Compose..."
  ssh_cmd "DEBIAN_FRONTEND=noninteractive apt-get install -y docker-compose"
fi

# 检查Docker Compose安装
if ! ssh_cmd "docker-compose --version || docker compose version"; then
  print_error "Docker Compose安装失败，请手动检查"
else
  print_success "Docker Compose安装完成"
fi

# 步骤5: 创建必要的目录结构
print_info "步骤5: 创建目录结构..."
if ! ssh_cmd "mkdir -p /root/ilock/backups /root/ilock/logs /root/ilock/mqtt/config /root/ilock/mqtt/data /root/ilock/mqtt/log"; then
  print_error "目录结构创建失败"
else
  print_success "目录结构创建完成"
fi

# 步骤6: 配置防火墙
print_info "步骤6: 配置防火墙..."
ssh_cmd "ufw --force enable" || true
ssh_cmd "ufw allow 22/tcp && ufw allow 20033/tcp && ufw allow 1883/tcp && ufw allow 8883/tcp && ufw allow 9001/tcp" || true
print_success "防火墙配置完成"

# 步骤7: 配置Docker镜像加速
print_info "步骤7: 配置Docker镜像加速..."
if ! ssh_cmd 'mkdir -p /etc/docker && cat > /etc/docker/daemon.json << EOF
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://registry.cn-hangzhou.aliyuncs.com"
  ]
}
EOF
systemctl restart docker'; then
  print_error "Docker镜像加速配置失败"
else
  print_success "Docker镜像加速配置完成"
fi

# 步骤8: 配置系统参数
print_info "步骤8: 优化系统参数..."
if ! ssh_cmd 'cat >> /etc/sysctl.conf << EOF
# 优化网络参数
net.ipv4.tcp_max_syn_backlog = 8192
net.core.somaxconn = 8192
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_tw_buckets = 5000
net.ipv4.tcp_tw_reuse = 1
EOF
sysctl -p'; then
  print_error "系统参数优化失败"
else
  print_success "系统参数优化完成"
fi

# 步骤9: 配置MQTT默认配置
print_info "步骤9: 配置MQTT默认配置..."
if ! ssh_cmd 'cat > /root/ilock/mqtt/config/mosquitto.conf << EOF
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
EOF'; then
  print_error "MQTT配置失败"
else
  print_success "MQTT配置完成"
fi

# 步骤10: 验证安装
print_info "步骤10: 验证安装..."
if ! ssh_cmd "docker --version && docker-compose --version || docker compose version"; then
  print_error "Docker验证失败"
fi

if ! ssh_cmd "ufw status"; then
  print_error "防火墙验证失败"
fi

if ! ssh_cmd "ls -la /root/ilock/"; then
  print_error "目录验证失败"
fi

print_success "验证完成"

# 完成初始化
print_success "京东云服务器初始化完成!"
print_info "服务器地址: $JD_SSH_HOST"
print_info "现在您可以运行迁移脚本将数据从阿里云迁移到京东云"
print_info "建议的下一步: ./migrate_to_jd_amd64.sh" 