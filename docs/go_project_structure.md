# Go 项目结构优化建议

## 当前项目结构分析

目前项目结构已经有了一定的组织，但可以根据Go的标准项目布局进一步优化。当前结构的主要特点：

- 使用了 `internal/error` 目录存放错误处理相关代码
- 有清晰的 `controllers`、`models`、`services` 等目录
- 包含配置、路由、中间件等常见组件

## Go标准项目布局建议

参考Go社区广泛采用的项目布局标准，建议对项目结构进行以下优化：

### 1. 核心目录结构

#### `/cmd`

存放主要的应用程序入口。

```
cmd/
  server/
    main.go    # HTTP服务入口
  migration/
    main.go    # 数据库迁移工具
```

#### `/internal`

存放不希望被外部导入的代码。建议将现有的业务逻辑代码移至此目录：

```
internal/
  app/                  # 应用层
    controllers/        # 控制器
    middleware/         # 中间件
    routes/             # 路由
  domain/               # 领域层
    models/             # 数据模型
    services/           # 业务服务
  infrastructure/       # 基础设施层
    config/             # 配置
    database/           # 数据库
    mqtt/               # MQTT客户端
  error/                # 错误处理（已有）
    code/               # 错误码
    response/           # 响应格式
```

#### `/pkg`

存放可以被外部项目安全导入的代码包。

```
pkg/
  utils/                # 通用工具函数
  logger/               # 日志工具
  validator/            # 验证工具
```

### 2. 辅助目录

#### `/configs`

配置文件模板或默认配置：

```
configs/
  config.yaml.example   # 配置文件示例
  production.yaml       # 生产环境配置
  development.yaml      # 开发环境配置
```

#### `/docs`

项目文档：

```
docs/
  api/                  # API文档
  error_handling.md     # 错误处理文档
  deployment.md         # 部署文档
  go_project_structure.md  # 项目结构文档
```

#### `/scripts`

各种构建、安装、分析等操作的脚本：

```
scripts/
  build.sh              # 构建脚本
  deploy.sh             # 部署脚本
  migration.sh          # 迁移脚本
```

#### `/test`

额外的外部测试应用程序和测试数据：

```
test/
  api_test/             # API测试
  data/                 # 测试数据
```

## 具体迁移建议

1. **创建 `cmd` 目录**：
   - 将 `main.go` 移动到 `cmd/server/main.go`
   - 如有其他可执行程序，也放在对应子目录中

2. **重组 `internal` 目录**：
   - 将 `controllers` 移动到 `internal/app/controllers`
   - 将 `models` 移动到 `internal/domain/models`
   - 将 `services` 移动到 `internal/domain/services`
   - 将 `middleware` 移动到 `internal/app/middleware`
   - 将 `routes` 移动到 `internal/app/routes`
   - 将 `config` 移动到 `internal/infrastructure/config`

3. **创建 `pkg` 目录**：
   - 将通用工具函数移动到 `pkg/utils`
   - 将日志相关代码移动到 `pkg/logger`

4. **保持 `internal/error` 目录**：
   - 当前的错误处理结构良好，可以保留

## 优势与收益

1. **更好的代码组织**：遵循标准布局使代码结构更加清晰
2. **提高可维护性**：明确的分层有助于理解代码职责
3. **促进代码复用**：区分内部代码和可复用代码
4. **便于团队协作**：标准结构使新团队成员更容易上手
5. **符合Go社区最佳实践**：采用广泛认可的项目组织方式

## 实施步骤

1. 创建新的目录结构
2. 逐步迁移现有代码
3. 更新导入路径
4. 测试确保功能正常
5. 更新文档和构建脚本

## 参考资料

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go官方文档：组织Go模块](https://go.dev/doc/modules/layout)
- [Effective Go](https://golang.org/doc/effective_go) 