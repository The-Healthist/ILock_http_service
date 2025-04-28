# iLock 智能门禁系统

## 项目概述

iLock是一个基于Go语言开发的智能门禁管理系统，提供了强大的门禁控制、视频通话和紧急情况处理功能。系统采用Docker容器化部署，便于在各种环境中快速安装和更新。

## 系统架构

- **后端**: Go + Gin框架 + GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.4.1
- **部署**: Docker + Docker Compose
- **通讯**: RESTful API + RTC视频通话

## 主要功能

- 用户管理（管理员、物业人员、居民）
- 设备管理（智能门锁监控和控制）
- 视频通话（访客与居民之间的实时沟通）
- 紧急情况处理（火灾、入侵、医疗等紧急事件）
- 完整的认证和权限管理

## 部署指南

### 前置要求

1. **服务器环境**:
   - Linux服务器（推荐Ubuntu 20.04或CentOS 8）
   - Docker和Docker Compose已安装
   - 开放端口：8080(HTTP), 3308(MySQL), 6380(Redis)

2. **本地环境**（用于部署）:
   - Windows操作系统
   - 已安装PuTTY工具集（包含pscp.exe和plink.exe）

### 快速部署（Windows）

我们提供了一个一键部署脚本`deploy_ilock.bat`，可以自动完成打包、上传和部署过程：

1. **下载部署脚本**并保存到本地项目根目录。

2. **执行部署脚本**：
   - 双击运行`deploy_ilock.bat`
   - 按提示确认或修改服务器配置信息
   - 脚本会自动检查、打包、上传和部署项目

3. **验证部署**：
   - 脚本会自动验证服务是否成功启动
   - 访问`http://服务器IP:8080`检查服务运行状态
   - 访问`http://服务器IP:20033/swagger/index.html`查看API文档

### 手动部署

如果你需要手动部署，可以按照以下步骤操作：

1. **克隆代码到本地**：
   ```bash
   git clone <repository-url>
   cd ilock-http-service
   ```

2. **创建环境配置文件**：
   ```bash
   cp .env.example .env
   # 编辑.env文件，设置数据库和JWT等配置
   ```

3. **上传到服务器**：
   ```bash
   scp -r ./* user@server:/path/to/ilock-service/
   ```

4. **在服务器上启动服务**：
   ```bash
   cd /path/to/ilock-service/
   docker-compose up -d
   ```

### 服务器配置参考

默认配置如下，可以根据需要在部署时修改：

- **服务器**:
  - IP: 39.108.49.167
  - 部署目录: /root/ilock
  - SSH端口: 22

- **数据库**:
  - 主机: localhost
  - 端口: 3309
  - 用户: root
  - 密码: 1090119your
  - 数据库名: ilock_db

- **Docker镜像**:
  - 镜像仓库: https://goproxy.cn,direct
  - 服务端口: 20033

## 更新与维护

### 使用部署脚本更新

当系统需要更新时，你可以使用同样的部署脚本：

1. 将需要更新的文件准备好（例如`controller/base/building_controller.go`等）
2. 运行部署脚本，它会自动创建备份并更新文件
3. 系统会自动验证更新是否成功，如果失败会自动回滚

### 常用维护命令

- **查看服务日志**:
  ```bash
  docker-compose logs -f backend
  ```

- **重启服务**:
  ```bash
  docker-compose restart backend
  ```

- **回滚到之前版本**:
  ```bash
  cd /root/ilock && \
  cp -r backup/controller/base/* controller/base/ && \
  cp -r backup/models/base/* models/base/ && \
  cp -r backup/services/base/* services/base/ && \
  docker-compose down && \
  docker-compose build backend && \
  docker-compose up -d
  ```

## API文档

系统集成了Swagger文档，部署后可以通过以下地址访问：

http://服务器IP:20033/swagger/index.html

主要API端点包括：

- **认证**: `/api/auth/login`
- **管理员**: `/api/admins/*`
- **物业人员**: `/api/staffs/*`
- **居民**: `/api/residents/*`
- **设备**: `/api/devices/*`
- **通话记录**: `/api/calls/*`
- **紧急情况**: `/api/emergency/*`


## 系统特性

- **自动备份与回滚**: 在更新前自动创建备份，更新失败时自动回滚
- **基于角色的访问控制**: 系统管理员、物业人员和居民具有不同的权限
- **安全通信**: 基于JWT的API认证
- **视频通话集成**: 集成阿里云RTC提供实时视频通话
- **容器化部署**: 使用Docker和Docker Compose简化部署和维护
- **紧急响应系统**: 快速处理火灾、入侵等紧急情况
- **健康检查**: 服务健康状态监控，确保系统稳定运行

## 故障排除

1. **服务无法启动**:
   - 检查Docker和Docker Compose是否正确安装
   - 检查端口是否被占用: `netstat -tunlp`
   - 查看容器日志: `docker-compose logs backend`

2. **数据库连接失败**:
   - 检查数据库配置是否正确
   - 确认数据库服务是否运行: `docker-compose ps mysql`
   - 尝试手动连接数据库验证凭据

3. **API响应错误**:
   - 检查JWT密钥配置
   - 确认请求格式是否正确
   - 查看服务日志了解详细错误信息

4. **部署脚本执行失败**:
   - 确保PuTTY工具集已正确安装并在PATH中
   - 检查服务器连接信息是否正确
   - 验证本地文件存在且权限正确

## 许可证

版权所有 © 2024 iLock开发团队