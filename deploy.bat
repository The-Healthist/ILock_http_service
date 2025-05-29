@echo off
REM iLock Manual Deployment Script

REM 版本设置
set VERSION=1.1.0

REM Server settings
set SSH_HOST=39.108.49.167
set SSH_PORT=22
set SSH_USERNAME=root
set SSH_PASSWORD=1090119your@

REM Docker Hub settings
set DOCKER_USERNAME=stonesea
set DOCKER_PASSWORD=1090119your

REM 重新生成Swagger文档
echo Regenerating Swagger documentation...
swag init -g main.go

REM Login to Docker Hub
echo Logging in to Docker Hub...
docker login -u %DOCKER_USERNAME% -p %DOCKER_PASSWORD%

REM Build Docker image with version
echo Building Docker image version %VERSION%...
docker build -t %DOCKER_USERNAME%/ilock-http-service:%VERSION% .

REM Tag as latest as well
echo Tagging as latest...
docker tag %DOCKER_USERNAME%/ilock-http-service:%VERSION% %DOCKER_USERNAME%/ilock-http-service:latest

REM Push Docker image to Docker Hub
echo Pushing versioned Docker image to Docker Hub...
docker push %DOCKER_USERNAME%/ilock-http-service:%VERSION%

echo Pushing latest Docker image to Docker Hub...
docker push %DOCKER_USERNAME%/ilock-http-service:latest

REM 更新docker-compose.yml中的版本号
echo Updating docker-compose.yml with version %VERSION%...
type docker-compose.yml | findstr /v "image: stonesea/ilock-http-service" > docker-compose.tmp
for /f "tokens=1* delims=:" %%a in ('findstr /n "^" docker-compose.yml') do (
    if %%a equ 3 (
        echo     image: stonesea/ilock-http-service:%VERSION% >> docker-compose.new
    ) else (
        echo.%%b >> docker-compose.new
    )
)
move /y docker-compose.new docker-compose.yml
del docker-compose.tmp

REM Copy files to server using scp
echo Copying deployment files to server...
scp -P %SSH_PORT% docker-compose.yml .env %SSH_USERNAME%@%SSH_HOST%:/root/ilock/

REM Check if .env file was copied successfully
echo Verifying .env file on server...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && ls -la && cat .env | head -5"

REM Execute deployment commands on server using ssh
echo Executing deployment commands on server...

REM 使用直接的SSH命令在服务器上执行部署步骤，而不是通过脚本文件
echo Running deployment commands directly via SSH...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && mkdir -p /root/ilock/logs && echo '{\"registry-mirrors\": [\"https://docker.1ms.run\", \"https://docker.mybacc.com\", \"https://dytt.online\", \"https://lispy.org\", \"https://docker.xiaogenban1993.com\", \"https://docker.yomansunter.com\", \"https://aicarbon.xyz\", \"https://666860.xyz\", \"https://docker.zhai.cm\", \"https://a.ussh.net\", \"https://hub.littlediary.cn\", \"https://hub.rat.dev\", \"https://docker.m.daocloud.io\", \"https://registry.cn-hangzhou.aliyuncs.com\"]}' | sudo tee /etc/docker/daemon.json && sudo systemctl daemon-reload && sudo systemctl restart docker && echo 'Stopping existing services...' && docker-compose down --volumes --remove-orphans || true && echo 'Cleaning old images...' && docker rmi %DOCKER_USERNAME%/ilock-http-service:%VERSION% %DOCKER_USERNAME%/ilock-http-service:latest || true"

REM 确保.env文件权限正确
echo Ensuring .env file has proper permissions...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && chmod 644 .env"

REM 拉取镜像并等待MySQL和Redis就绪
echo Pulling images and starting services...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && echo 'Starting to pull images...' && docker-compose pull && echo 'Starting services...' && docker-compose up -d"

REM 等待MySQL就绪
echo Waiting for MySQL to be ready...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && for i in {1..30}; do if docker-compose ps db | grep -q 'Up'; then if docker exec ilock_mysql mysqladmin ping -h localhost --silent; then echo 'MySQL is ready!'; break; fi; fi; if [ $i -eq 30 ]; then echo 'MySQL startup timeout'; docker-compose logs db; exit 1; fi; echo 'MySQL starting... (attempt '$i'/30)'; sleep 2; done"

REM 等待Redis就绪
echo Waiting for Redis to be ready...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && for i in {1..30}; do if docker-compose ps redis | grep -q 'Up'; then if docker exec ilock_redis redis-cli ping | grep -q 'PONG'; then echo 'Redis is ready!'; break; fi; fi; if [ $i -eq 30 ]; then echo 'Redis startup timeout'; docker-compose logs redis; exit 1; fi; echo 'Redis starting... (attempt '$i'/30)'; sleep 2; done"

REM 检查应用容器的状态
echo Checking application service status...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && docker-compose ps && docker exec ilock_http_service ls -la /app"

REM 等待应用服务就绪 (修复语法错误)
echo Waiting for application service to be ready...
ssh -p %SSH_PORT% %SSH_USERNAME%@%SSH_HOST% "cd /root/ilock && for i in {1..60}; do if docker-compose ps app | grep -q 'Up'; then if curl -s http://localhost:20033/api/ping > /dev/null 2>&1; then echo 'Application service started successfully!'; docker-compose ps; exit 0; fi; fi; if [ $i -eq 60 ]; then echo 'Application service timeout'; docker-compose logs app; exit 1; fi; if [ $(($i %% 5)) -eq 0 ]; then echo 'Application service starting... (attempt '$i'/60)'; docker-compose logs --tail=10 app; fi; sleep 2; done"

REM 检查SSH返回值来判断部署是否成功
if %errorlevel% neq 0 (
  echo Deployment failed. Please check the logs.
) else (
  echo Deployment successful! Deployed version %VERSION%
)

pause