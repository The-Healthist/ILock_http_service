# iLock HTTP 服务迁移工具

本工具用于将 iLock HTTP 服务从一个服务器迁移到另一个服务器。工具包含四个主要脚本：备份脚本、部署脚本、恢复脚本和回滚脚本。适用于新的项目结构。

## 前提条件

1. 安装 sshpass：
   - macOS: `brew install hudochenkov/sshpass/sshpass`
   - Linux: `sudo apt-get install sshpass`
2. 确保源服务器和目标服务器可以通过 SSH 访问

## 项目结构变更

新的项目结构遵循Go标准布局，主要变化：

1. 主程序入口移至 `/cmd/server/main.go`
2. 配置文件位于 `/configs` 目录
3. 业务逻辑位于 `/internal` 目录
4. 公共工具位于 `/pkg` 目录

Docker相关配置也相应做了调整，包括：
- Dockerfile中的构建路径调整
- docker-compose.yml中的挂载路径调整
- 配置文件的位置从应用根目录移至configs目录

## 脚本说明

1. **备份脚本 (backup.sh)**
   - 备份源服务器上的 MySQL 数据、Redis 数据、环境配置和 configs 配置目录
   - 备份 MQTT 数据目录
   - 下载备份到本地 `./backups` 目录

2. **部署脚本 (deploy.sh)**
   - 使用备份的配置文件在目标服务器上部署应用
   - 配置防火墙，安装 Docker 和 Docker Compose
   - 上传配置目录和数据目录
   - 拉取指定版本的 Docker 镜像并启动服务

3. **恢复脚本 (restore.sh)**
   - 恢复备份的数据到目标服务器
   - 修改应用版本号
   - 启动服务

4. **回滚脚本 (rollback.sh)**
   - 将应用回滚到指定版本
   - 可选择是否恢复对应版本的配置备份

## 使用步骤

### 步骤 1：备份源服务器数据

```bash
# 可选：设置源服务器配置
export SSH_HOST="源服务器IP"
export SSH_PORT="22"
export SSH_USER="root"
export SSH_PASS="你的密码"

# 执行备份
./backup.sh
```

### 步骤 2：部署到目标服务器

```bash
# 可选：设置目标服务器配置
export TARGET_HOST="目标服务器IP"
export TARGET_PORT="22"
export TARGET_USER="root"
export TARGET_PASS="你的密码"
export VERSION="1.4.0"  # 指定要部署的版本号

# 执行部署
./deploy.sh
```

### 步骤 3：仅恢复数据（可选）

如果你只想恢复数据而不进行完整部署：

```bash
# 可选：设置目标服务器配置
export TARGET_HOST="目标服务器IP"
export TARGET_PORT="22"
export TARGET_USER="root"
export TARGET_PASS="你的密码"
export VERSION="1.4.0"  # 指定要恢复的版本号

# 执行恢复
./restore.sh
```

### 步骤 4：回滚版本（如需）

如果部署出现问题，需要回滚到之前的版本：

```bash
# 可选：设置目标服务器配置
export TARGET_HOST="目标服务器IP"
export TARGET_PORT="22"
export TARGET_USER="root"
export TARGET_PASS="你的密码"

# 执行回滚，指定要回滚到的版本号
./rollback.sh 1.3.0
```

## 目录结构

服务器上的目录结构：

```
/root/ilock/
├── configs/               # 配置文件目录
│   ├── mysql/             # MySQL初始化脚本
│   ├── redis/             # Redis配置
│   └── mqtt/              # MQTT配置
├── backups/               # 备份文件目录
├── logs/                  # 日志目录
├── mqtt/                  # MQTT数据目录
│   ├── data/              # MQTT数据
│   └── log/               # MQTT日志
├── .env                   # 环境变量
└── docker-compose.yml     # Docker编排文件
```

## 常见问题

1. **备份失败**
   - 检查源服务器是否正常运行
   - 检查源服务器上的 Docker 容器状态

2. **部署失败**
   - 检查目标服务器网络连接
   - 检查目标服务器磁盘空间
   - 查看 Docker 日志：`docker-compose logs`

3. **恢复失败**
   - 确保已经完成备份步骤
   - 检查 MySQL 和 Redis 备份文件是否完整

4. **防火墙配置**
   - 确保目标服务器开放以下端口：
     - 20033 (HTTP 服务)
     - 1883 (MQTT)
     - 8883 (MQTT SSL)
     - 9001 (MQTT WebSocket) 