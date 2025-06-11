# iLock 系统部署指南

本文档提供 iLock 系统的部署和维护指南，适合技术和非技术人员使用。

## 目录

1. [部署前准备](#部署前准备)
2. [简易部署步骤](#简易部署步骤)
3. [版本更新方法](#版本更新方法)
4. [数据库管理](#数据库管理)
5. [回滚操作](#回滚操作)
6. [常见问题](#常见问题)

## 部署前准备

### 环境要求

- macOS 系统（用于执行部署脚本）
- 安装了 Docker 和 Docker Buildx
- 安装了 Go 和 Swag 工具
- 安装了 sshpass 工具

### 检查工具安装

1. 检查 Docker：`docker --version`
2. 检查 Swag：`swag --version`
3. 安装 sshpass：`brew install sshpass`

### 配置信息

部署脚本中已包含以下配置信息：

- 服务器 IP: 39.108.49.167
- SSH 端口: 22
- 用户名: root
- Docker Hub 账号: stonesea

## 简易部署步骤

### 第一次部署

1. **获取部署脚本**：确保您有 `deploy_simple.sh` 脚本文件。

2. **添加执行权限**：
   ```bash
   chmod +x deploy_simple.sh
   ```

3. **修改版本号**（如需要）：
   打开 `deploy_simple.sh`，修改 `VERSION="1.3.0"` 为您需要部署的版本。

4. **执行部署**：
   ```bash
   ./deploy_simple.sh
   ```

5. **验证部署**：
   部署完成后，脚本会自动检查服务状态。您也可以通过浏览器访问 `http://39.108.49.167:20033/api/ping` 验证。

### 重要说明

- 首次部署会在服务器上创建必要的目录结构
- 部署前会自动备份现有数据
- 部署过程包括：生成 Swagger 文档、构建推送镜像、更新配置文件、启动服务、执行数据库迁移

## 版本更新方法

当需要更新系统版本时，按照以下步骤操作：

1. **修改版本号**：
   打开 `deploy_simple.sh`，更新 `VERSION="1.3.0"` 为新版本号。

2. **执行部署脚本**：
   ```bash
   ./deploy_simple.sh
   ```

3. **确认更新成功**：
   脚本执行完成后，检查输出信息确认部署成功。

## 数据库管理

### 自动备份

部署脚本默认在每次部署前自动备份数据库。备份文件保存在服务器的 `/root/ilock/backups` 目录下，命名格式为：
- MySQL: `ilock_db_YYYYMMDD_HHMMSS.sql.gz`
- Redis: `ilock_redis_YYYYMMDD_HHMMSS.rdb.gz`

系统自动保留最近 7 次备份，旧备份会被自动删除。

### 手动备份

如需手动备份数据库，可执行以下命令：

```bash
export SSHPASS='1090119your@'
sshpass -e ssh -p 22 root@39.108.49.167 'cd /root/ilock && ./backup_script.sh'
```

### 数据库迁移

当系统版本更新包含数据库结构变更时，部署脚本会自动执行迁移：

1. 系统首先备份现有数据
2. 部署新版本应用
3. 执行数据库迁移脚本

如果您需要禁用自动迁移，可以在 `deploy_simple.sh` 中将 `AUTO_MIGRATE=true` 改为 `AUTO_MIGRATE=false`。

## 回滚操作

如果新版本部署后出现问题，可以使用回滚脚本将系统回退到之前的版本：

```bash
./rollback.sh 1.2.0  # 将系统回滚到 1.2.0 版本
```

回滚操作不会影响数据库数据，只会切换应用程序版本。

## 常见问题

### 部署失败

**问题**：执行部署脚本后显示失败。

**解决方法**：
1. 检查服务器连接是否正常
2. 检查 Docker Hub 登录是否成功
3. 查看服务器日志：
   ```bash
   export SSHPASS='1090119your@'
   sshpass -e ssh -p 22 root@39.108.49.167 'cd /root/ilock && docker-compose logs'
   ```

### 数据库迁移失败

**问题**：部署成功但数据库迁移失败。

**解决方法**：
1. 检查迁移日志
   ```bash
   export SSHPASS='1090119your@'
   sshpass -e ssh -p 22 root@39.108.49.167 'cd /root/ilock && docker exec ilock_http_service cat /app/logs/migration.log'
   ```
2. 如果是数据结构问题，可能需要手动执行 SQL 修复，请咨询开发人员

### 服务无法访问

**问题**：部署后无法通过 API 访问服务。

**解决方法**：
1. 检查服务状态
   ```bash
   export SSHPASS='1090119your@'
   sshpass -e ssh -p 22 root@39.108.49.167 'cd /root/ilock && docker-compose ps'
   ```
2. 查看应用日志
   ```bash
   export SSHPASS='1090119your@'
   sshpass -e ssh -p 22 root@39.108.49.167 'cd /root/ilock && docker-compose logs app'
   ```
3. 检查服务器防火墙是否开放了 20033 端口

### 备份管理

**问题**：如何查看或恢复备份文件？

**解决方法**：
1. 查看可用备份
   ```bash
   export SSHPASS='1090119your@'
   sshpass -e ssh -p 22 root@39.108.49.167 'ls -la /root/ilock/backups'
   ```
2. 恢复数据库备份需要执行特定的恢复命令，请联系开发人员获取帮助 