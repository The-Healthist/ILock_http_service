# iLock 项目 CI/CD 自动化部署指南

本指南介绍如何使用GitHub Actions为iLock项目设置自动化构建和部署流程。

## 工作原理

这个CI/CD流程会在每次推送代码到GitHub仓库时自动触发以下操作：

1. 在GitHub提供的Ubuntu环境中构建应用
2. 打包Docker镜像并推送到Docker Hub
3. 登录到您的服务器，自动部署最新版本
4. 进行健康检查，如果部署失败则自动回滚

## 前提条件

1. 将您的iLock项目代码托管在GitHub上
2. 拥有Docker Hub账号
3. 拥有可访问的Linux服务器，已安装Docker和Docker Compose

## 设置步骤

### 1. 设置GitHub Secrets

在GitHub仓库中添加以下secrets（项目设置 > Secrets and variables > Actions）：

- `DOCKER_USERNAME`: 您的Docker Hub用户名
- `DOCKER_PASSWORD`: 您的Docker Hub密码
- `SSH_HOST`: 您的服务器IP地址，例如 39.108.49.167
- `SSH_PORT`: SSH端口，默认为 22
- `SSH_USERNAME`: SSH用户名，例如 root
- `SSH_PASSWORD`:  

### 2. 准备服务器

确保您的服务器满足以下条件：

1. 已安装Docker和Docker Compose
2. 创建了部署目录：`/root/ilock`
3. 配置了正确的`.env`文件
4. 确保docker-compose.yml文件中app服务使用如下配置：
   ```yaml
   app:
     image: ${DOCKER_USERNAME}/ilock-service:latest
     # 其他配置...
   ```

### 3. 修改项目的docker-compose.yml

如果您的项目中`docker-compose.yml`使用的是本地构建（build: .），需要修改为使用Docker Hub镜像：

```yaml
app:
  image: ${DOCKER_USERNAME}/ilock-service:latest
  # 而不是 build: .
```

并确保服务器上设置了环境变量`DOCKER_USERNAME`或在docker-compose.yml文件中直接填写您的Docker Hub用户名。

### 4. 推送代码触发部署

将工作流配置文件`.github/workflows/build.yml`提交到您的仓库。每次推送代码时，GitHub Actions会自动执行构建和部署。

## 工作流程详解

1. **构建阶段**:
   - 拉取代码
   - 安装Go环境
   - 下载依赖
   - 构建Linux可执行文件
   - 构建Docker镜像并推送

2. **部署阶段**:
   - 备份现有代码和配置
   - 拉取新镜像
   - 重启服务
   - 健康检查（多次尝试）
   - 如果失败则回滚

## 自定义配置

根据项目需求，您可能需要调整以下内容：

1. **端口配置**: 
   - 默认使用端口20033进行健康检查
   - 如需修改，请更新`.github/workflows/build.yml`中的健康检查URL

2. **健康检查路径**:
   - 默认检查`/api/ping`端点
   - 如API路径不同，请相应更新

3. **备份逻辑**:
   - 当前备份controller、models、services目录
   - 可根据项目结构增加其他重要目录

## 故障排除

如果部署失败，请检查：

1. GitHub Actions日志中的错误信息
2. 确保所有secrets配置正确
3. 检查服务器上的Docker日志：`docker logs ilock_app`
4. 确保服务器上的docker-compose.yml文件正确引用Docker Hub镜像

## 手动部署备选方案

如果需要手动部署，您仍然可以使用之前的`deploy_ilock.bat`脚本进行部署。 