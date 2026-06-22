# CLAUDE.md

本文件为 Claude Code（claude.ai/code）在此仓库中工作时提供指导。

## 改动边界与编译失败处理（强约束）

**严格限定改动范围**：一次任务只允许修改与该任务直接相关的服务/文件。例如任务只针对 `zec` 时，**严禁**修改 `bmc`、`vm`、`sdn`、`zga`、`zlb`、`zdns`、`traffic`、`zrm`、`keypair`、`pvtdns` 等其他服务的代码，也不要修改 `provider.go`、`connectivity/client.go`、`common/` 等公共文件，除非任务本身就要求改它们。

**禁止跨项目"顺手"改动**：不要为了让 `go build ./...` 通过，而去修改其他服务里已经存在的、与本任务无关的报错。这些报错可能是仓库已有的、由 SDK 版本/分支差异造成的预存在状态。

**不要运行 `go fmt ./...` 或 `make fmtcheck` 后批量"顺手"修复格式**：这会改动大量无关文件。只对本次任务涉及的文件单独跑 `gofmt -l <file>` 验证格式。

**编译不通过时的处理**：
1. 先确认报错是不是本次改动**直接**引入的
2. 如果报错来自其他文件/服务，先告知用户，**不要**擅自去改
3. 让用户决定是修 SDK 版本、加 replace 指令、还是用别的方式处理
4. 用户没明确授权前，不要为了"绕过编译"去修改无关代码

## 参数顺序规范（强约束）

**必填参数在前，非必填参数在后**：无论是 `.go`、`.md`、`.html` 还是其他任何文件，定义参数/字段/属性时，必填项必须排在所有可选项之前。

## Schema 字段校验与文档同步规范（强约束）

**`ValidateFunc` 与 `Description` 必须同步**：凡是 API 文档规定了取值范围或枚举值的字段，schema 定义中必须同时满足：
1. `ValidateFunc`：添加对应的范围校验（如 `validation.IntBetween`、`validation.StringInSlice` 等）
2. `Description`：在描述末尾明确写出范围或枚举，例如 `Range: 0-65535.` 或 `Valid values: \`TCP\`, \`UDP\`.`

`Description` 是自动生成 HTML 文档（`make gendoc`）的唯一来源，缺少范围说明会导致生成的文档中用户看不到有效值约束。两者缺一不可。

## 构建、测试与开发命令

```bash
# 构建 provider 二进制文件
make build

# 构建并安装到 Terraform 插件目录（~/.terraform.d/plugins/）
# 注意：Makefile 硬编码了 OS_ARCH=darwin_amd64 和 NAMESPACE=test；其他平台需覆盖：
#   make install OS_ARCH=linux_arm64
make install

# 运行单元测试
make test

# 运行验收测试（需要 TF_ACC=1，超时 120 分钟）
make testacc

# 针对特定资源运行验收测试
TF_ACC=1 go test ./zenlayercloud -v -run TestAccZenlayerCloudBmcEip -timeout 120m

# 检查 Go 格式
make fmtcheck

# 手动格式化代码
go fmt ./...

# 从源码生成文档
make gendoc

# 为所有平台构建发布二进制
make release
```

## Provider 架构

这是一个使用 Terraform Plugin SDK v2（`v2.19.0`）的 Zenlayer Cloud Terraform Provider。SDK 通过本地兄弟目录引入：`replace github.com/zenlayer/zenlayercloud-sdk-go => ../zenlayercloud-sdk-go`。

### 关键组件

**入口（`main.go`）** — 通过 `plugin.Serve()` 启动 provider，支持 delve 调试。

**Provider 配置（`zenlayercloud/provider.go`）**
- Schema 字段：`access_key_id`、`access_key_password`、`domain`、`scheme`、`client_timeout`
- 在 `resourcesMap()` 和 `dataSourcesMap()` 中注册所有资源和数据源
- 环境变量：`ZENLAYERCLOUD_ACCESS_KEY_ID`、`ZENLAYERCLOUD_ACCESS_KEY_PASSWORD`、`ZENLAYERCLOUD_CLIENT_TIMEOUT`、`ZENLAYERCLOUD_SCHEME`、`ZENLAYERCLOUD_DOMAIN`

**客户端管理（`zenlayercloud/connectivity/client.go`）**
- `ZenlayerCloudClient` 通过 `WithXxxClient()` 方法懒加载所有服务连接
- 额外客户端：`WithCcsClient()`（ccs20250901）、`WithUsrClient()`（user20240529）、`WithZrmClient()`（zrm20251014）

### 服务目录结构

`zenlayercloud/services/` 下的服务：`keypair/`、`pvtdns/`、`traffic/`、`zdns/`、`zec/`、`zlb/`、`zrm/`

旧服务（BMC、VM、SDN、ZGA）直接放在 `zenlayercloud/` 根目录下。

### ZEC 双 SDK 版本模式

ZEC 同时使用两个 SDK 版本：
- `zec20240401`（`WithZecClient()`）— 旧版本，用于部分删除/变更操作
- `zec20250901`（`WithZec2Client()`）— 当前版本，用于读取和大多数创建/更新操作

新增 ZEC 资源时默认使用 `WithZec2Client()`。查看 `service_zenlayer_zec.go` 中的现有方法来确认特定 API 操作属于哪个客户端。

### 资源实现模式

```go
func ResourceZenlayerCloudZecFoo() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceZenlayerCloudZecFooCreate,
        ReadContext:   resourceZenlayerCloudZecFooRead,
        UpdateContext: resourceZenlayerCloudZecFooUpdate,
        DeleteContext: resourceZenlayerCloudZecFooDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(30 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(10 * time.Minute),
        },
        Schema: map[string]*schema.Schema{...},
    }
}
```

**关键模式：**
- 服务操作通过封装了 `*connectivity.ZenlayerCloudClient` 的 `ZecService`（或类似结构体）进行
- 异步操作使用 `resource.StateChangeConf` + `WaitForStateContext` 轮询直到终态
- 可重试的 API 调用使用 `resource.RetryContext` + `common.RetryError(ctx, err, additionalCodes...)`
- 标准额外重试码：`common.InternalServerError` 和 `sdkcommon.NetworkError`
- 创建后不可变更的字段使用 `ForceNew: true`
- ZEC/ZLB 资源的标签更新通过 ZRM 服务：`zrm.NewZrmService(client).ModifyResourceTags(ctx, d, resourceId)`
- 删除时遇到 `NOT_FOUND` 错误应返回 `nil`（幂等删除）

**源文件文档：** 每个资源文件顶部有块注释，包含描述、HCL 示例用法和导入说明。

### 公共工具（`zenlayercloud/common/`）

- `RetryError(ctx, err, ...codes)` — 将 SDK 错误分类为可重试或不可重试；`ServiceNotAvailable` 始终可重试
- `QueryAllPaginatedResource[T]` — 泛型助手，并发分页查询（最多 50 个 goroutine）并汇总结果
- `GoRoutineLimit` — 有界 goroutine 池，用于并发 API 调用
- `LogApiRequest` — 所有 API 调用的结构化日志（错误时 warn，成功时 debug）
- `ParseResourceId(id, length)` — 按 `:` 分隔符拆分复合 ID
- `GetTags` / `TagsToMap` — `TypeMap` schema 字段的标签读写助手
- 超时环境变量覆盖：`ZENLAYERCLOUD_WRITE_RETRY_TIMEOUT`、`ZENLAYERCLOUD_READ_RETRY_TIMEOUT`、`ZENLAYERCLOUD_BMC_CREATE_TIMEOUT`、`ZENLAYERCLOUD_VM_CREATE_TIMEOUT` 等

### 扩展文件（`extension_*.go` 和 `extension_zec.go`）

服务专属常量和状态值。ZEC 的所有状态常量（镜像、磁盘、实例、EIP、NAT、VNIC、CIDR、快照）都在 `zenlayercloud/services/zec/extension_zec.go` 中。

### 注册新资源

实现资源文件后，在 `provider.go` 中注册：
1. 在 `resourcesMap()` 中对应服务注释块下添加
2. 在 `dataSourcesMap()` 中添加数据源
3. 在 `provider.go` 顶部的 provider 文档注释中添加资源名

参考仓库根目录的 `CLAUDE_API_WORKFLOW.md`，了解新增 API 的完整流程——服务目录决策、SDK 接入、资源/数据源脚手架和文档生成。

### 测试

- 单元测试：同包下的 `_test.go`；运行 `make test`
- 验收测试：需要 `TF_ACC=1` 和真实凭证；运行 `make testacc`
- `basic_test.go` 定义共享测试常量（默认可用区、密码、实例规格）

### SDK 集成

使用 `github.com/zenlayer/zenlayercloud-sdk-go`（本地 replace 指令指向 `../zenlayercloud-sdk-go`）：
- `bmc20221120`、`vm20230313`、`sdn20230830`、`zga20230706` — 旧服务
- `zec20240401`、`zec20250901` — 弹性计算（双版本，见上文）
- `zlb20250401`、`zdns20251101`、`traffic20240326`、`zrm20251014` — 较新服务
- `ccs20250901`、`user20240529` — 客户端中已有但尚未完全开放的额外服务
