#!/bin/bash
# iLock 服务迁移主脚本
# 整合备份、恢复和部署步骤

# 导入公共函数
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/common.sh"

# 显示帮助信息
function show_help() {
  echo "iLock 服务迁移工具 v$VERSION"
  echo ""
  echo "用法: $0 [选项]"
  echo ""
  echo "选项:"
  echo "  -h, --help       显示帮助信息"
  echo "  -b, --backup     仅执行备份步骤"
  echo "  -r, --restore    仅执行恢复步骤"
  echo "  -d, --deploy     仅执行部署步骤"
  echo "  -a, --all        执行完整迁移流程（默认）"
  echo ""
  echo "环境变量:"
  echo "  SOURCE_SSH_HOST         源服务器主机名/IP (默认: 39.108.49.167)"
  echo "  SOURCE_SSH_PORT         源服务器SSH端口 (默认: 22)"
  echo "  SOURCE_SSH_USERNAME     源服务器用户名 (默认: root)"
  echo "  SOURCE_SSH_PASSWORD     源服务器密码"
  echo ""
  echo "  TARGET_SSH_HOST         目标服务器主机名/IP (默认: 117.72.193.54)"
  echo "  TARGET_SSH_PORT         目标服务器SSH端口 (默认: 22)"
  echo "  TARGET_SSH_USERNAME     目标服务器用户名 (默认: root)"
  echo "  TARGET_SSH_PASSWORD     目标服务器密码"
  echo ""
  echo "示例:"
  echo "  $0 --all                执行完整迁移流程"
  echo "  $0 --backup             仅执行备份步骤"
  echo "  SOURCE_SSH_HOST=192.168.1.100 $0 --backup    使用自定义源服务器执行备份"
}

# 执行备份
function do_backup() {
  print_info "开始执行备份步骤..."
  bash "$SCRIPT_DIR/backup/backup.sh"
}

# 执行恢复
function do_restore() {
  print_info "开始执行恢复步骤..."
  bash "$SCRIPT_DIR/restore/restore.sh"
}

# 执行部署
function do_deploy() {
  print_info "开始执行部署步骤..."
  bash "$SCRIPT_DIR/deploy/deploy.sh"
}

# 执行完整迁移
function do_full_migration() {
  print_info "开始执行完整迁移流程..."
  
  # 执行备份
  do_backup
  if [ $? -ne 0 ]; then
    print_error "备份步骤失败，中止迁移！"
    exit 1
  fi
  
  # 执行恢复
  do_restore
  if [ $? -ne 0 ]; then
    print_error "恢复步骤失败，中止迁移！"
    exit 1
  fi
  
  # 执行部署
  do_deploy
  if [ $? -ne 0 ]; then
    print_error "部署步骤失败！"
    exit 1
  fi
  
  print_success "完整迁移流程执行成功！"
}

# 主函数
function main() {
  # 检查命令行参数
  if [ $# -eq 0 ]; then
    # 默认执行完整迁移
    do_full_migration
  else
    case "$1" in
      -h|--help)
        show_help
        ;;
      -b|--backup)
        do_backup
        ;;
      -r|--restore)
        do_restore
        ;;
      -d|--deploy)
        do_deploy
        ;;
      -a|--all)
        do_full_migration
        ;;
      *)
        echo "未知选项: $1"
        show_help
        exit 1
        ;;
    esac
  fi
}

# 执行主函数
main "$@" 