@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

:: 设置颜色代码
set "GREEN=[92m"
set "YELLOW=[93m"
set "RED=[91m"
set "NC=[0m"

echo %GREEN%=====================================================%NC%
echo %GREEN%      iLock HTTP Service 一键部署脚本                %NC%
echo %GREEN%=====================================================%NC%

:: 检查PuTTY工具集是否在PATH中或在脚本所在目录
set "PSCP_PATH="
set "PLINK_PATH="

:: 首先检查当前目录
if exist "%~dp0pscp.exe" (
    set "PSCP_PATH=%~dp0pscp.exe"
    echo %YELLOW%在当前目录找到pscp.exe%NC%
) else (
    where pscp.exe >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        set "PSCP_PATH=pscp.exe"
        echo %YELLOW%在系统PATH中找到pscp.exe%NC%
    )
)

if exist "%~dp0plink.exe" (
    set "PLINK_PATH=%~dp0plink.exe"
    echo %YELLOW%在当前目录找到plink.exe%NC%
) else (
    where plink.exe >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        set "PLINK_PATH=plink.exe"
        echo %YELLOW%在系统PATH中找到plink.exe%NC%
    )
)

:: 如果未找到工具，提示下载
if "%PSCP_PATH%"=="" (
    echo %RED%错误: 找不到pscp.exe工具%NC%
    echo %YELLOW%请访问 https://www.chiark.greenend.org.uk/~sgtatham/putty/latest.html 下载PuTTY工具集%NC%
    echo %YELLOW%并将pscp.exe和plink.exe放在此脚本所在目录或添加到系统PATH中%NC%
    echo.
    echo %YELLOW%按任意键退出...%NC%
    pause > nul
    exit /b 1
)

if "%PLINK_PATH%"=="" (
    echo %RED%错误: 找不到plink.exe工具%NC%
    echo %YELLOW%请访问 https://www.chiark.greenend.org.uk/~sgtatham/putty/latest.html 下载PuTTY工具集%NC%
    echo %YELLOW%并将pscp.exe和plink.exe放在此脚本所在目录或添加到系统PATH中%NC%
    echo.
    echo %YELLOW%按任意键退出...%NC%
    pause > nul
    exit /b 1
)

:: 配置服务器信息
set "REMOTE_USER=root"
set "REMOTE_HOST=39.108.49.167"
set "REMOTE_PORT=22"
set "REMOTE_DIR=/root/ilock"
set "REMOTE_PASS=1090119your@"

:: 配置数据库信息
set "DB_HOST=localhost"
set "DB_PORT=3309"
set "DB_USER=root"
set "DB_PASS=1090119your"
set "DB_NAME=ilock_db"

:: 配置Docker镜像信息
set "DOCKER_REGISTRY=https://goproxy.cn,direct"
set "DOCKER_IMAGE=ilock-service"
set "DOCKER_TAG=latest"
set "SERVICE_PORT=20033"

:: 是否需要更改配置？
echo %YELLOW%是否需要修改服务器配置 (y/n) [n]: %NC%
set /p CHANGE_CONFIG=""
if /i "%CHANGE_CONFIG%"=="y" (
    echo %YELLOW%服务器IP地址 [%REMOTE_HOST%]: %NC%
    set /p NEW_HOST=""
    if not "%NEW_HOST%"=="" set "REMOTE_HOST=%NEW_HOST%"
    
    echo %YELLOW%SSH端口 [%REMOTE_PORT%]: %NC%
    set /p NEW_PORT=""
    if not "%NEW_PORT%"=="" set "REMOTE_PORT=%NEW_PORT%"
    
    echo %YELLOW%用户名 [%REMOTE_USER%]: %NC%
    set /p NEW_USER=""
    if not "%NEW_USER%"=="" set "REMOTE_USER=%NEW_USER%"
    
    echo %YELLOW%部署目录 [%REMOTE_DIR%]: %NC%
    set /p NEW_DIR=""
    if not "%NEW_DIR%"=="" set "REMOTE_DIR=%NEW_DIR%"
    
    echo %YELLOW%服务器密码 [隐藏]: %NC%
    set /p REMOTE_PASS=""
    
    echo %YELLOW%服务端口 [%SERVICE_PORT%]: %NC%
    set /p NEW_SERVICE_PORT=""
    if not "%NEW_SERVICE_PORT%"=="" set "SERVICE_PORT=%NEW_SERVICE_PORT%"
)

echo %GREEN%必要工具检查通过!%NC%

:: 创建临时目录
if exist "temp" rd /s /q "temp"
mkdir temp

:: 创建部署配置文件
echo # 数据库配置 > temp\.env
echo DB_HOST=mysql >> temp\.env
echo DB_PORT=3306 >> temp\.env
echo DB_USER=root >> temp\.env
echo DB_PASSWORD=%DB_PASS% >> temp\.env
echo DB_NAME=%DB_NAME% >> temp\.env
echo DB_TIMEZONE=Asia/Shanghai >> temp\.env
echo DB_MIGRATION_MODE=alter >> temp\.env
echo. >> temp\.env
echo # 服务器配置 >> temp\.env
echo SERVER_PORT=%SERVICE_PORT% >> temp\.env
echo. >> temp\.env
echo # JWT配置 >> temp\.env
echo JWT_SECRET_KEY=ilock-secret-key-%RANDOM%%RANDOM% >> temp\.env
echo. >> temp\.env
echo # 阿里云RTC >> temp\.env
echo ALIYUN_ACCESS_KEY=67613a6a74064cad9859c8f794980cae >> temp\.env
echo ALIYUN_RTC_APP_ID=md3fh5x4 >> temp\.env
echo ALIYUN_RTC_REGION=cn-hangzhou >> temp\.env
echo. >> temp\.env
echo # Redis配置 >> temp\.env
echo REDIS_HOST=redis >> temp\.env
echo REDIS_PORT=6380 >> temp\.env
echo REDIS_DB=0 >> temp\.env

:: 创建远程部署脚本
echo #!/bin/bash > temp\deploy.sh
echo set -e >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 颜色输出 >> temp\deploy.sh
echo GREEN='\033[0;32m' >> temp\deploy.sh
echo YELLOW='\033[1;33m' >> temp\deploy.sh
echo RED='\033[0;31m' >> temp\deploy.sh
echo NC='\033[0m' # No Color >> temp\deploy.sh
echo. >> temp\deploy.sh
echo CURRENT_DIR="$(pwd)" >> temp\deploy.sh
echo. >> temp\deploy.sh
echo echo -e "${YELLOW}开始执行远程部署...${NC}" >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 检查Docker和Docker Compose是否安装 >> temp\deploy.sh
echo echo -e "${YELLOW}检查Docker和Docker Compose...${NC}" >> temp\deploy.sh
echo if ! command -v docker ^&^> /dev/null; then >> temp\deploy.sh
echo   echo -e "${RED}Docker未安装，开始安装Docker...${NC}" >> temp\deploy.sh
echo   curl -fsSL https://get.docker.com ^| sh >> temp\deploy.sh
echo   systemctl enable docker >> temp\deploy.sh
echo   systemctl start docker >> temp\deploy.sh
echo else >> temp\deploy.sh
echo   echo -e "${GREEN}Docker已安装${NC}" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 检查docker-compose >> temp\deploy.sh
echo if ! command -v docker-compose ^&^> /dev/null; then >> temp\deploy.sh
echo   echo -e "${RED}Docker Compose未安装，开始安装Docker Compose...${NC}" >> temp\deploy.sh
echo   curl -L "https://github.com/docker/compose/releases/download/v2.24.6/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose >> temp\deploy.sh
echo   chmod +x /usr/local/bin/docker-compose >> temp\deploy.sh
echo   ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose >> temp\deploy.sh
echo else >> temp\deploy.sh
echo   echo -e "${GREEN}Docker Compose已安装${NC}" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 检查服务状态 >> temp\deploy.sh
echo echo -e "${YELLOW}检查当前服务状态...${NC}" >> temp\deploy.sh
echo docker-compose ps 2^>/dev/null ^|^| echo "Docker Compose服务未运行" >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 解压部署包 >> temp\deploy.sh
echo echo -e "${YELLOW}解压部署包...${NC}" >> temp\deploy.sh
echo if [ -f "ilock_service.zip" ]; then >> temp\deploy.sh
echo   unzip -o ilock_service.zip -d temp/ >> temp\deploy.sh
echo   echo "部署包解压完成" >> temp\deploy.sh
echo else >> temp\deploy.sh
echo   echo -e "${RED}错误: 找不到部署包 ilock_service.zip${NC}" >> temp\deploy.sh
echo   exit 1 >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 创建备份目录 >> temp\deploy.sh
echo echo -e "${YELLOW}创建备份目录...${NC}" >> temp\deploy.sh
echo TIMESTAMP=$(date +%%Y%%m%%d_%%H%%M%%S) >> temp\deploy.sh
echo BACKUP_DIR="$CURRENT_DIR/backups/$TIMESTAMP" >> temp\deploy.sh
echo mkdir -p "$BACKUP_DIR" >> temp\deploy.sh
echo mkdir -p "$BACKUP_DIR/controller/base" >> temp\deploy.sh
echo mkdir -p "$BACKUP_DIR/models/base" >> temp\deploy.sh
echo mkdir -p "$BACKUP_DIR/services/base" >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 备份现有文件 >> temp\deploy.sh
echo echo -e "${YELLOW}备份现有文件...${NC}" >> temp\deploy.sh
echo if [ -f "docker-compose.yml" ]; then >> temp\deploy.sh
echo   cp docker-compose.yml "$BACKUP_DIR/" >> temp\deploy.sh
echo   echo "docker-compose.yml已备份" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo if [ -f ".env" ]; then >> temp\deploy.sh
echo   cp .env "$BACKUP_DIR/" >> temp\deploy.sh
echo   echo ".env已备份" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo if [ -f "main.go" ]; then >> temp\deploy.sh
echo   cp main.go "$BACKUP_DIR/" >> temp\deploy.sh
echo   echo "main.go已备份" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo # 备份关键目录文件 >> temp\deploy.sh
echo if [ -d "controller/base" ]; then >> temp\deploy.sh
echo   cp -r controller/base/* "$BACKUP_DIR/controller/base/" >> temp\deploy.sh
echo   echo "控制器目录已备份" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo if [ -d "models/base" ]; then >> temp\deploy.sh
echo   cp -r models/base/* "$BACKUP_DIR/models/base/" >> temp\deploy.sh
echo   echo "模型目录已备份" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo if [ -d "services/base" ]; then >> temp\deploy.sh
echo   cp -r services/base/* "$BACKUP_DIR/services/base/" >> temp\deploy.sh
echo   echo "服务目录已备份" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 备份数据库(如果可能) >> temp\deploy.sh
echo if docker-compose exec -T mysql mysqldump -u%DB_USER% -p%DB_PASS% %DB_NAME% ^> "$BACKUP_DIR/%DB_NAME%.sql" 2^>/dev/null; then >> temp\deploy.sh
echo   echo "数据库已备份到 $BACKUP_DIR/%DB_NAME%.sql" >> temp\deploy.sh
echo else >> temp\deploy.sh
echo   echo "数据库备份失败或服务未运行，继续部署" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 停止现有服务 >> temp\deploy.sh
echo echo -e "${YELLOW}停止现有服务...${NC}" >> temp\deploy.sh
echo docker-compose down 2^>/dev/null ^|^| echo "无需停止服务" >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 应用新文件 >> temp\deploy.sh
echo echo -e "${YELLOW}应用部署文件...${NC}" >> temp\deploy.sh
echo cp -rf temp/* . >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 使用新的.env配置 >> temp\deploy.sh
echo if [ -f ".env.example" ] ^&^& [ ! -f ".env" ]; then >> temp\deploy.sh
echo   cp .env.example .env >> temp\deploy.sh
echo   echo "使用示例.env文件创建配置" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 创建日志目录 >> temp\deploy.sh
echo echo -e "${YELLOW}创建日志目录...${NC}" >> temp\deploy.sh
echo mkdir -p "$CURRENT_DIR/logs" >> temp\deploy.sh
echo chmod -R 755 "$CURRENT_DIR/logs" >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 启动服务 >> temp\deploy.sh
echo echo -e "${YELLOW}开始构建和启动服务...${NC}" >> temp\deploy.sh
echo docker-compose build --no-cache >> temp\deploy.sh
echo docker-compose up -d >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 等待服务启动 >> temp\deploy.sh
echo echo -e "${YELLOW}等待服务启动...${NC}" >> temp\deploy.sh
echo sleep 10 >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 验证服务状态 >> temp\deploy.sh
echo echo -e "${YELLOW}验证服务状态...${NC}" >> temp\deploy.sh
echo docker-compose ps >> temp\deploy.sh
echo echo "检查服务健康状态..." >> temp\deploy.sh
echo if curl -s http://localhost:%SERVICE_PORT%/api/ping ^| grep -q "pong"; then >> temp\deploy.sh
echo   DEPLOY_SUCCESS=true >> temp\deploy.sh
echo   echo -e "${GREEN}服务健康检查通过!${NC}" >> temp\deploy.sh
echo else >> temp\deploy.sh
echo   # 尝试多次检查，可能服务启动较慢 >> temp\deploy.sh
echo   for i in {1..5}; do >> temp\deploy.sh
echo     echo "重试健康检查 $i/5..." >> temp\deploy.sh
echo     sleep 5 >> temp\deploy.sh
echo     if curl -s http://localhost:%SERVICE_PORT%/api/ping ^| grep -q "pong"; then >> temp\deploy.sh
echo       DEPLOY_SUCCESS=true >> temp\deploy.sh
echo       echo -e "${GREEN}服务健康检查通过!${NC}" >> temp\deploy.sh
echo       break >> temp\deploy.sh
echo     fi >> temp\deploy.sh
echo   done >> temp\deploy.sh
echo   if [ "$DEPLOY_SUCCESS" != "true" ]; then >> temp\deploy.sh
echo     DEPLOY_SUCCESS=false >> temp\deploy.sh
echo     echo -e "${RED}服务健康检查失败!${NC}" >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 判断是否部署成功 >> temp\deploy.sh
echo if [ "$DEPLOY_SUCCESS" != "true" ]; then >> temp\deploy.sh
echo   echo -e "${RED}错误: 服务未正常启动，开始回滚...${NC}" >> temp\deploy.sh
echo   echo -e "${YELLOW}停止服务...${NC}" >> temp\deploy.sh
echo   docker-compose down >> temp\deploy.sh
echo   echo -e "${YELLOW}恢复备份...${NC}" >> temp\deploy.sh
echo   if [ -f "$BACKUP_DIR/docker-compose.yml" ]; then >> temp\deploy.sh
echo     cp "$BACKUP_DIR/docker-compose.yml" . >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo   if [ -f "$BACKUP_DIR/.env" ]; then >> temp\deploy.sh
echo     cp "$BACKUP_DIR/.env" . >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo   if [ -f "$BACKUP_DIR/main.go" ]; then >> temp\deploy.sh
echo     cp "$BACKUP_DIR/main.go" . >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo   # 恢复关键目录文件 >> temp\deploy.sh
echo   if [ -d "$BACKUP_DIR/controller/base" ]; then >> temp\deploy.sh
echo     cp -r "$BACKUP_DIR/controller/base/"* controller/base/ >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo   if [ -d "$BACKUP_DIR/models/base" ]; then >> temp\deploy.sh
echo     cp -r "$BACKUP_DIR/models/base/"* models/base/ >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo   if [ -d "$BACKUP_DIR/services/base" ]; then >> temp\deploy.sh
echo     cp -r "$BACKUP_DIR/services/base/"* services/base/ >> temp\deploy.sh
echo   fi >> temp\deploy.sh
echo   echo -e "${YELLOW}重新启动服务...${NC}" >> temp\deploy.sh
echo   docker-compose up -d >> temp\deploy.sh
echo   echo -e "${RED}部署失败，已回滚到之前的版本${NC}" >> temp\deploy.sh
echo   exit 1 >> temp\deploy.sh
echo else >> temp\deploy.sh
echo   echo -e "${GREEN}部署成功!${NC}" >> temp\deploy.sh
echo   echo -e "${GREEN}=====================================================${NC}" >> temp\deploy.sh
echo   echo -e "${GREEN}      部署完成!                                      ${NC}" >> temp\deploy.sh
echo   echo -e "${GREEN}=====================================================${NC}" >> temp\deploy.sh
echo   echo -e "${YELLOW}服务已成功部署到: http://$(hostname -I ^| awk '{print $1}'):%SERVICE_PORT%${NC}" >> temp\deploy.sh
echo   echo -e "${YELLOW}Swagger API文档: http://$(hostname -I ^| awk '{print $1}'):%SERVICE_PORT%/swagger/index.html${NC}" >> temp\deploy.sh
echo   echo -e "${YELLOW}备份保存在: $BACKUP_DIR${NC}" >> temp\deploy.sh
echo fi >> temp\deploy.sh
echo. >> temp\deploy.sh
echo # 清理临时文件 >> temp\deploy.sh
echo echo -e "${YELLOW}清理临时文件...${NC}" >> temp\deploy.sh
echo rm -rf temp >> temp\deploy.sh
echo rm -f ilock_service.zip >> temp\deploy.sh
echo. >> temp\deploy.sh
echo echo -e "${GREEN}=====================================================${NC}" >> temp\deploy.sh
echo echo -e "${GREEN}      部署脚本执行完毕                               ${NC}" >> temp\deploy.sh
echo echo -e "${GREEN}=====================================================${NC}" >> temp\deploy.sh
echo echo -e "${YELLOW}查看日志: docker-compose logs -f${NC}" >> temp\deploy.sh
echo echo -e "${YELLOW}重启服务: docker-compose restart${NC}" >> temp\deploy.sh
echo echo -e "${YELLOW}停止服务: docker-compose down${NC}" >> temp\deploy.sh

echo %YELLOW%检查项目文件...%NC%

:: 检查必要文件
set MISSING_FILES=0

if not exist "main.go" (
    echo %YELLOW%警告: 未找到main.go文件%NC%
    set /a MISSING_FILES+=1
)

if not exist "go.mod" (
    echo %YELLOW%警告: 未找到go.mod文件%NC%
    set /a MISSING_FILES+=1
)

if not exist "Dockerfile" (
    echo %YELLOW%警告: 未找到Dockerfile文件%NC%
    set /a MISSING_FILES+=1
)

if not exist "docker-compose.yml" (
    echo %YELLOW%警告: 未找到docker-compose.yml文件%NC%
    set /a MISSING_FILES+=1
)

if %MISSING_FILES% gtr 0 (
    echo %YELLOW%发现%MISSING_FILES%个文件可能缺失，建议检查项目完整性%NC%
    set /p CONTINUE="是否继续打包部署？ (y/n) [n]: "
    if /i not "%CONTINUE%"=="y" (
        echo %RED%已取消部署%NC%
        exit /b 1
    )
)

:: 打包应用
echo %YELLOW%开始打包应用...%NC%
set "ZIP_FILE=ilock_service.zip"

:: 清理之前的打包文件
if exist "%ZIP_FILE%" del /q "%ZIP_FILE%"

:: 检查是否存在PowerShell
powershell -command "echo PowerShell可用" > nul 2>&1
if %ERRORLEVEL% equ 0 (
    echo %YELLOW%使用PowerShell打包...%NC%
    
    :: 创建排除文件列表
    echo .git/ > exclude.txt
    echo .vscode/ >> exclude.txt
    echo .idea/ >> exclude.txt
    echo logs/ >> exclude.txt
    echo backups/ >> exclude.txt
    echo *.log >> exclude.txt
    echo *.zip >> exclude.txt
    echo *.tar.gz >> exclude.txt
    echo temp/ >> exclude.txt
    echo node_modules/ >> exclude.txt
    echo deploy_ilock.bat >> exclude.txt
    echo exclude.txt >> exclude.txt
    echo services/weather/ >> exclude.txt
    echo controller/weather/ >> exclude.txt
    echo models/weather/ >> exclude.txt
    
    :: 使用PowerShell创建排除某些目录的ZIP
    powershell -command "& { Add-Type -A 'System.IO.Compression.FileSystem'; $exclude = Get-Content .\exclude.txt; $items = Get-ChildItem -Path . -Exclude $exclude; $zip = [IO.Compression.ZipFile]::Open('%ZIP_FILE%', [IO.Compression.ZipArchiveMode]::Create); foreach ($item in $items) { $relativePath = $item.FullName.Substring($PWD.Path.Length + 1); if ($item.PSIsContainer) { $null = [System.IO.Directory]::CreateDirectory([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), $relativePath)); foreach ($file in (Get-ChildItem -Path $item.FullName -Recurse -File)) { $relativeFilePath = $file.FullName.Substring($PWD.Path.Length + 1); if (-not ($exclude | Where-Object { $relativeFilePath -match $_ })) { $entry = $zip.CreateEntry($relativeFilePath); $writer = New-Object System.IO.BinaryWriter $entry.Open(); $writer.Write([System.IO.File]::ReadAllBytes($file.FullName)); $writer.Dispose(); } } } else { $entry = $zip.CreateEntry($relativePath); $writer = New-Object System.IO.BinaryWriter $entry.Open(); $writer.Write([System.IO.File]::ReadAllBytes($item.FullName)); $writer.Dispose(); } }; $zip.Dispose(); }"
    del exclude.txt
) else (
    echo %YELLOW%使用第三方工具打包...%NC%
    :: 尝试使用可能存在的其他打包工具
    where 7z.exe >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        7z a -tzip "%ZIP_FILE%" * -xr!.git -xr!.vscode -xr!logs -xr!*.log -xr!*.zip -xr!temp -xr!node_modules -xr!deploy_ilock.bat -xr!services/weather -xr!controller/weather -xr!models/weather
    ) else (
        where zip.exe >nul 2>&1
        if %ERRORLEVEL% equ 0 (
            zip -r "%ZIP_FILE%" * -x ".git/*" ".vscode/*" "logs/*" "*.log" "*.zip" "deploy_ilock.bat" "temp/*" "node_modules/*" "services/weather/*" "controller/weather/*" "models/weather/*"
        ) else (
            echo %RED%错误: 找不到可用的打包工具 (PowerShell, 7z 或 zip)%NC%
            echo %RED%请安装其中一种打包工具后再试%NC%
            exit /b 1
        )
    )
)

if not exist "%ZIP_FILE%" (
    echo %RED%错误: 打包失败，未生成ZIP文件%NC%
    exit /b 1
)

echo %GREEN%应用打包完成: %ZIP_FILE%，大小: !%~z%ZIP_FILE%! 字节%NC%

:: 验证服务器连接
echo %YELLOW%验证服务器连接...%NC%
echo y | "%PLINK_PATH%" -ssh -P %REMOTE_PORT% -pw %REMOTE_PASS% %REMOTE_USER%@%REMOTE_HOST% "echo 连接成功" 2>nul
if %ERRORLEVEL% neq 0 (
    echo %RED%错误: 无法连接到服务器 %REMOTE_HOST%，请检查服务器信息%NC%
    exit /b 1
)
echo %GREEN%服务器连接验证成功!%NC%

:: 使用PLINK创建远程目录
echo %YELLOW%创建远程目录...%NC%
echo y | "%PLINK_PATH%" -ssh -P %REMOTE_PORT% -pw %REMOTE_PASS% %REMOTE_USER%@%REMOTE_HOST% "mkdir -p %REMOTE_DIR% && mkdir -p %REMOTE_DIR%/logs"

:: 上传部署文件和环境配置
echo %YELLOW%上传部署文件和环境配置...%NC%
"%PSCP_PATH%" -P %REMOTE_PORT% -pw %REMOTE_PASS% temp\deploy.sh %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_DIR%/
"%PSCP_PATH%" -P %REMOTE_PORT% -pw %REMOTE_PASS% temp\.env %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_DIR%/.env.new

:: 上传应用程序ZIP包
echo %YELLOW%上传应用程序包 (这可能需要一些时间)...%NC%
for /f "tokens=1,2 delims= " %%a in ('wmic datafile where "name='%~dp0%ZIP_FILE:\=\\%'" get FileSize /value ^| find "="') do set SIZE=%%b
set /a SIZE_MB=%SIZE% / 1048576
echo 文件大小: %SIZE_MB% MB
"%PSCP_PATH%" -P %REMOTE_PORT% -pw %REMOTE_PASS% "%ZIP_FILE%" %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_DIR%/

:: 执行远程部署脚本
echo %YELLOW%执行远程部署脚本...%NC%
echo y | "%PLINK_PATH%" -ssh -P %REMOTE_PORT% -pw %REMOTE_PASS% %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && chmod +x deploy.sh && ./deploy.sh"

:: 检查部署结果
echo %YELLOW%验证部署结果...%NC%
echo y | "%PLINK_PATH%" -ssh -P %REMOTE_PORT% -pw %REMOTE_PASS% %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && docker-compose ps" > temp_result.txt
findstr /C:"Up" temp_result.txt > nul
if %ERRORLEVEL% equ 0 (
    echo %GREEN%部署成功! 服务已启动.%NC%
) else (
    echo %RED%警告: 服务可能未正常启动，请登录服务器检查日志%NC%
)
del temp_result.txt

:: 清理临时文件
echo %YELLOW%清理临时文件...%NC%
rd /s /q temp
if exist "%ZIP_FILE%" del /q "%ZIP_FILE%"

echo %GREEN%=====================================================%NC%
echo %GREEN%      部署完成!                                      %NC%
echo %GREEN%=====================================================%NC%
echo %YELLOW%服务器: %NC%%REMOTE_HOST%
echo %YELLOW%安装目录: %NC%%REMOTE_DIR%
echo %YELLOW%服务地址: %NC%http://%REMOTE_HOST%:%SERVICE_PORT%
echo %YELLOW%Swagger API: %NC%http://%REMOTE_HOST%:%SERVICE_PORT%/swagger/index.html
echo %YELLOW%查看服务日志可以使用: %NC%ssh %REMOTE_USER%@%REMOTE_HOST% -p %REMOTE_PORT%
echo %GREEN%=====================================================%NC%

echo.
echo %YELLOW%常用维护命令(使用PuTTY连接服务器后执行):%NC%
echo 1. 查看日志: cd %REMOTE_DIR% ^&^& docker-compose logs -f
echo 2. 重启服务: cd %REMOTE_DIR% ^&^& docker-compose restart
echo 3. 停止服务: cd %REMOTE_DIR% ^&^& docker-compose down
echo 4. 回滚版本: cd %REMOTE_DIR% ^&^& cp -r backup/controller/base/* controller/base/ ^&^& ^
                cp -r backup/models/base/* models/base/ ^&^& ^
                cp -r backup/services/base/* services/base/ ^&^& ^
                docker-compose down ^&^& docker-compose build backend ^&^& docker-compose up -d
echo.
echo %YELLOW%部署脚本执行完毕，按任意键退出...%NC%
pause > nul 