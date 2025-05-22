@echo off
echo Generating Swagger documentation...

set GOROOT=C:\Program Files\Go
set PATH=%PATH%;%GOROOT%\bin;%USERPROFILE%\go\bin
set GOBIN=%USERPROFILE%\go\bin

echo Installing swag CLI...
go install github.com/swaggo/swag/cmd/swag@latest

echo Running swag init...
"%USERPROFILE%\go\bin\swag.exe" init -g main.go -o docs

echo Done! 