# ILock 智能门禁系统部署文档

## 目录

1. [系统概述](#系统概述)
2. [部署准备](#部署准备)
3. [部署流程](#部署流程)
4. [本地开发部署](#本地开发部署)
5. [多云部署](#多云部署)
6. [数据保护措施](#数据保护措施)
7. [版本管理](#版本管理)
8. [故障恢复](#故障恢复)
9. [常见问题](#常见问题)
10. [部署脚本参数说明](#部署脚本参数说明)

## 系统概述

ILock智能门禁系统是一套基于Docker容器的微服务架构应用，由以下几个主要组件组成：

- **应用服务 (app)**: 主要业务逻辑服务
- **MySQL数据库 (db)**: 存储持久化数据
- **Redis缓存 (redis)**: 提供高速缓存
- **MQTT服务 (mqtt)**: 提供实时通信服务

系统使用Docker Compose进行容器编排，便于快速部署和管理。

## 部署准备

### 环境要求

- macOS系统 (对于Windows系统请使用deploy.bat)
- Docker (20.10.0+)
- Docker Compose (2.0.0+)
- SSH客户端
- 权限: 对目标服务器的SSH访问权限

### 工具安装

1. **Swagger工具**: 用于生成API文档
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. **Docker**: [安装说明](https://docs.docker.com/get-docker/)

## 部署流程

### 1. 克隆代码库

```bash
git clone https://github.com/yourusername/ilock-http-service.git
cd ilock-http-service
```

### 2. 配置环境变量

编辑`.env`文件，设置必要的环境变量:

```
MYSQL_ROOT_PASSWORD=your_secure_password
MYSQL_DATABASE=ilock
ALIYUN_ACCESS_KEY=your_aliyun_key
ALIYUN_RTC_APP_ID=your_rtc_app_id
ALIYUN_RTC_REGION=cn-hangzhou
DEFAULT_ADMIN_PASSWORD=your_admin_password
```

### 3. 配置部署脚本

编辑`deploy.sh`脚本，更新以下配置项:

```bash
# 版本设置
VERSION="1.3.0"  # 当前要部署的版本

# 部署配置
BACKUP_ENABLED=true  # 是否在部署前备份数据
AUTO_MIGRATE=true    # 是否自动运行数据库迁移
FORCE_RECREATE=false # 是否强制重建容器(保留数据卷)

# 服务器设置
SSH_HOST="your.server.ip"
SSH_PORT="22"
SSH_USERNAME="root"
SSH_PASSWORD="your_password"

# Docker Hub设置
DOCKER_USERNAME="yourusername"
DOCKER_PASSWORD="your_password"
```

### 4. 运行部署脚本

```bash
chmod +x deploy.sh  # 确保脚本有执行权限
./deploy.sh
```

部署过程会执行以下步骤:

1. 生成Swagger文档
2. 登录Docker Hub
3. 构建Docker镜像并打标签
4. 推送Docker镜像到Docker Hub
5. 更新docker-compose.yml中的版本号
6. 在服务器上创建备份(如启用)
7. 更新服务器上的容器
8. 等待服务就绪
9. 创建回滚脚本

## 本地开发部署

如果您没有配置到服务器的SSH连接，或仅在本地开发环境工作，可以使用`local_build.sh`脚本进行本地构建和测试。

### 1. 本地构建脚本

```bash
chmod +x docs/docs_deploy/local_build.sh  # 确保脚本有执行权限
./docs/docs_deploy/local_build.sh
```

该脚本执行以下步骤:

1. 生成Swagger文档
2. 询问是否需要推送到Docker Hub(可选)
3. 构建Docker镜像并打标签
4. 更新docker-compose.yml中的版本号
5. 询问是否在本地启动容器进行测试(可选)

### 2. 本地测试

如果选择在本地启动容器进行测试，脚本会运行`docker-compose up -d`命令。您可以通过以下命令查看容器运行状态:

```bash
docker-compose ps
```

查看日志:

```bash
docker-compose logs -f
```

停止容器:

```bash
docker-compose down
```

### 3. 配置远程部署

当您准备好部署到远程服务器时，需要完成以下步骤:

1. 配置SSH访问权限
   ```bash
   # 生成SSH密钥对(如果还没有)
   ssh-keygen -t rsa -b 4096
   
   # 将公钥复制到服务器
   ssh-copy-id -i ~/.ssh/id_rsa.pub root@your.server.ip
   ```

2. 更新`deploy.sh`中的服务器信息
   ```bash
   # 服务器设置
   SSH_HOST="your.server.ip"
   SSH_PORT="22"
   SSH_USERNAME="root"
   SSH_PASSWORD="your_password"  # 如果使用密钥认证，可以省略此项
   ```

3. 运行完整的部署脚本
   ```bash
   ./deploy.sh
   ```

## 多云部署

ILock系统支持在多个云平台上部署，目前已适配阿里云和京东云。

### 阿里云部署

阿里云部署使用`docs/docs_deploy/aliyun/deploy.sh`脚本，该脚本假设服务器已经完成基本配置。

```bash
chmod +x docs/docs_deploy/aliyun/deploy.sh
./docs/docs_deploy/aliyun/deploy.sh
```

### 京东云部署

京东云部署使用`docs/docs_deploy/jdcloud/deploy.sh`脚本，该脚本包含了服务器初始化功能，适合在全新服务器上部署。

1. **配置脚本**

   首先编辑`docs/docs_deploy/jdcloud/deploy.sh`，更新以下配置:
   
   ```bash
   # 服务器设置
   SSH_HOST="your.jdcloud.ip"  # 京东云服务器IP
   SSH_PORT="22"
   SSH_USERNAME="root"
   SSH_PASSWORD="your_password"
   
   # 部署配置
   INIT_SERVER=true  # 首次部署时设置为true，将安装必要软件
   ```

2. **执行部署**

   ```bash
   chmod +x docs/docs_deploy/jdcloud/deploy.sh
   ./docs/docs_deploy/jdcloud/deploy.sh
   ```

3. **服务器初始化**

   京东云部署脚本包含以下初始化步骤:
   
   - 更新系统包
   - 安装必要依赖
   - 安装Docker和Docker Compose(如果尚未安装)
   - 创建应用目录结构
   - 配置防火墙规则
   - 增加系统文件描述符限制
   - 设置时区为Asia/Shanghai

4. **后续部署**

   首次部署完成后，后续部署时可以将`INIT_SERVER`设置为`false`，跳过初始化步骤:
   
   ```bash
   # 部署配置
   INIT_SERVER=false  # 后续部署时设置为false
   ```

### 多云环境管理

当在多个云平台上部署ILock系统时，建议采用以下管理策略:

1. **版本一致性**: 确保所有环境使用相同版本的应用
2. **配置差异化**: 使用不同的`.env`文件管理各环境的配置差异
3. **负载均衡**: 考虑使用DNS负载均衡在多云之间分配流量
4. **数据同步**: 如需要在多云间同步数据，考虑实施数据库复制策略

## 数据保护措施

系统采用多层次的数据保护机制，确保在升级过程中数据安全:

### 自动备份

部署前自动备份MySQL和Redis数据:

- MySQL: 使用`mysqldump`备份所有数据库
- Redis: 使用`SAVE`命令和导出RDB文件备份
- 备份文件存储在服务器的`/root/ilock/backups/`目录
- 自动保留最近7次备份，删除过旧备份

### 数据卷持久化

使用Docker卷进行数据持久化:

- MySQL: `mysql_data`卷存储数据库文件
- Redis: `redis_data`卷存储Redis数据
- MQTT: 使用目录映射存储配置和数据

### 安全更新策略

部署过程使用安全更新策略:

- 默认模式: 只更新镜像，不重建容器
- 可选模式: 使用`FORCE_RECREATE=true`重建容器但保留数据卷
- 不使用`--volumes`参数，确保数据卷不被删除

## 版本管理

### 版本命名

使用语义化版本命名:

- 主版本.次版本.修订版本 (例如: 1.3.0)
- 主版本: 不兼容的API更改
- 次版本: 向后兼容的功能添加
- 修订版本: 向后兼容的问题修复

### 镜像标签

每个版本创建两个Docker镜像标签:

- 版本标签: `stonesea/ilock-http-service:1.3.0`
- 最新标签: `stonesea/ilock-http-service:latest`

## 故障恢复

### 回滚流程

使用`rollback.sh`脚本回滚到上一个稳定版本:

```bash
./rollback.sh 1.3.0  # 从1.3.0回滚到1.1.0
```

回滚过程执行以下步骤:

1. 更新docker-compose.yml中的镜像版本
2. 将更新后的文件复制到服务器
3. 拉取回滚版本镜像并重启服务

### 手动恢复

如需从备份恢复数据:

1. 列出可用备份:
   ```bash
   ssh root@your.server.ip 'ls -la /root/ilock/backups/'
   ```

2. 选择备份文件并恢复:
   ```bash
   ssh root@your.server.ip 'cd /root/ilock && \
   gunzip -c /root/ilock/backups/ilock_db_20230516_120000.sql.gz | \
   docker exec -i ilock_mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD"'
   ```

## 常见问题

### 部署失败

如果部署失败，请检查:

1. Docker Hub登录凭据是否正确
2. 服务器是否有足够的磁盘空间
3. 检查日志:
   ```bash
   ssh root@your.server.ip 'cd /root/ilock && docker-compose logs -f'
   ```

### 数据库连接错误

可能原因:

1. MySQL容器未正常启动
2. 密码配置错误
3. 数据库初始化失败

解决方法:

1. 检查MySQL日志:
   ```bash
   ssh root@your.server.ip 'cd /root/ilock && docker-compose logs db'
   ```
2. 确认`.env`文件包含正确的数据库配置
3. 尝试重启MySQL容器:
   ```bash
   ssh root@your.server.ip 'cd /root/ilock && docker-compose restart db'
   ```

### MQTT服务问题

可能原因:

1. 配置文件丢失
2. 端口冲突
3. 权限问题

解决方法:

1. 检查MQTT日志:
   ```bash
   ssh root@your.server.ip 'cd /root/ilock && docker-compose logs mqtt'
   ```
2. 确认mqtt/config目录包含必要的配置文件
3. 检查端口是否被占用:
   ```bash
   ssh root@your.server.ip 'netstat -tulpn | grep 1883'
   ```

## 部署脚本参数说明

### 配置参数

| 参数名 | 说明 | 默认值 |
|--------|------|--------|
| VERSION | 部署版本号 | 1.3.0 |
| BACKUP_ENABLED | 是否在部署前备份数据 | true |
| AUTO_MIGRATE | 是否自动运行数据库迁移 | true |
| FORCE_RECREATE | 是否强制重建容器(保留数据卷) | false |

### 服务器参数

| 参数名 | 说明 | 示例 |
|--------|------|------|
| SSH_HOST | 服务器IP地址 | 39.108.49.167 |
| SSH_PORT | SSH端口 | 22 |
| SSH_USERNAME | SSH用户名 | root |
| SSH_PASSWORD | SSH密码 | (敏感信息) |

### Docker Hub参数

| 参数名 | 说明 | 示例 |
|--------|------|------|
| DOCKER_USERNAME | Docker Hub用户名 | stonesea |
| DOCKER_PASSWORD | Docker Hub密码 | (敏感信息) |

---

本文档最后更新: 2024年6月1日
