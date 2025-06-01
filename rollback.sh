#!/bin/bash
# Rollback script for iLock deployment

# 服务器设置
SSH_HOST="39.108.49.167"
SSH_PORT="22"
SSH_USERNAME="root"
SSH_PASSWORD="1090119your@"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <version_to_rollback_from>"
  exit 1
fi

VERSION_TO_ROLLBACK=$1
PREVIOUS_VERSION="1.1.0"  # 设置为上一个稳定版本

echo "Rolling back from version $VERSION_TO_ROLLBACK to $PREVIOUS_VERSION"
echo "This will update docker-compose.yml and restart the service"
read -p "Continue? (y/n) " -n 1 -r
echo 
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Rollback aborted"
  exit 1
fi

# Update docker-compose.yml
sed -i '' "s|image: stonesea/ilock-http-service:$VERSION_TO_ROLLBACK|image: stonesea/ilock-http-service:$PREVIOUS_VERSION|" docker-compose.yml

# Copy to server and restart using sshpass
export SSHPASS="$SSH_PASSWORD"
sshpass -e scp -o StrictHostKeyChecking=no -P "$SSH_PORT" docker-compose.yml "$SSH_USERNAME@$SSH_HOST:/root/ilock/"
sshpass -e ssh -o StrictHostKeyChecking=no -p "$SSH_PORT" "$SSH_USERNAME@$SSH_HOST" "cd /root/ilock && docker-compose pull && docker-compose up -d"

echo "Rollback completed!"
