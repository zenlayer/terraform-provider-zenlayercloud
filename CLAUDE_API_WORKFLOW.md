# Terraform Provider Zenlayer - 添加新 API 接口工作流程分析

## 概述

本文档详细说明在 Terraform Provider Zenlayer 中添加新 API 接口（新资源或数据源）所需的完整工作流程。

---

## 一、添加新 API 接口的完整步骤

### 步骤 1：确定服务类型和架构

**决策点：是添加到现有服务还是创建新服务？**

- **添加到现有服务**：如果 API 属于已有服务（如 BMC、ZEC、ZLB）
- **创建新服务**：如果是全新的 Zenlayer 服务

**推荐架构**：新服务使用 `services/` 目录结构
```
zenlayercloud/services/newservice/
├── resource_zc_newservice_*.go      # 资源文件
├── data_source_zc_newservice_*.go   # 数据源文件
├── service_zenlayer_newservice.go   # 服务封装
└── extension_newservice.go           # 常量和枚举
```

---

### 步骤 2：SDK 客户端集成（新服务需要）

**文件：`zenlayercloud/connectivity/client.go`**

1. **导入 SDK 包**
```go
import (
    newService "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/newService20240101"
)
```

2. **添加客户端字段**
```go
type ZenlayerCloudClient struct {
    // ... 现有字段
    NewServiceConn *newService.Client
}
```

3. **实现 WithXxxClient() 方法**
```go
func (client *ZenlayerCloudClient) WithNewServiceClient() *newService.Client {
    if client.NewServiceConn != nil {
        return client.NewServiceConn
    }
    config := client.NewConfig()
    client.NewServiceConn, _ = newService.NewClient(config, client.SecretKeyId, client.SecretKeyPassword)
    client.NewServiceConn.WithRequestClient(ReqClient)
    return client.NewServiceConn
}
```

---

### 步骤 3：创建资源文件

**文件：`zenlayercloud/services/newservice/resource_zc_newservice_resource.go`**

```go
package newservice

import (
    "context"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
    newService "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/newService20240101"
)

// 资源函数（注意大写开头，用于导出）
func ResourceZenlayerCloudNewServiceResource() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceZenlayerCloudNewServiceResourceCreate,
        ReadContext:   resourceZenlayerCloudNewServiceResourceRead,
        UpdateContext: resourceZenlayerCloudNewServiceResourceUpdate,
        DeleteContext: resourceZenlayerCloudNewServiceResourceDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: map[string]*schema.Schema{
            // 参数定义
        },
    }
}
```

---

### 步骤 4：实现 CRUD 操作

#### 4.1 Create 操作
```go
func resourceZenlayerCloudNewServiceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*connectivity.ZenlayerCloudClient)

    // 创建 SDK 请求
    request := newService.NewCreateResourceRequest()

    // 设置请求参数
    if v, ok := d.GetOk("param_name"); ok {
        request.ParamName = v.(string)
    }

    // 调用 SDK
    response, err := client.WithNewServiceClient().CreateResource(request)
    if err != nil {
        return diag.FromErr(err)
    }

    // 设置资源 ID
    d.SetId(*response.ResourceId)

    // 异步状态等待（如需要）
    stateConf := &resource.StateChangeConf{
        Pending: []string{"PENDING", "CREATING"},
        Target:  []string{"RUNNING"},
        Refresh: resource.StateRefreshFunc(func() (interface{}, string, error) {
            // 查询状态
            resp, err := client.WithNewServiceClient().DescribeResource(newService.NewDescribeResourceRequest())
            if err != nil {
                return nil, "", err
            }
            return resp, *resp.Status, nil
        }),
        Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
        Delay:      10 * time.Second,
        MinTimeout: 5 * time.Second,
    }

    _, err = stateConf.WaitForStateContext(ctx)
    if err != nil {
        return diag.Errorf("error waiting for resource to become active: %s", err)
    }

    return resourceZenlayerCloudNewServiceResourceRead(ctx, d, meta)
}
```

#### 4.2 Read 操作
```go
func resourceZenlayerCloudNewServiceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*connectivity.ZenlayerCloudClient)

    request := newService.NewDescribeResourceRequest()
    request.ResourceId = &[]string{d.Id()}[0]

    response, err := client.WithNewServiceClient().DescribeResource(request)
    if err != nil {
        // 资源不存在，返回 nil 表示已删除
        if isNotFoundError(err) {
            d.SetId("")
            return nil
        }
        return diag.FromErr(err)
    }

    // 更新 state
    d.Set("param_name", response.ParamName)
    d.Set("computed_attr", response.ComputedAttr)

    // 处理列表类型
    if response.ListAttr != nil {
        list := make([]map[string]interface{}, len(response.ListAttr))
        for i, item := range response.ListAttr {
            list[i] = map[string]interface{}{
                "field1": item.Field1,
                "field2": item.Field2,
            }
        }
        if err := d.Set("list_attr", list); err != nil {
            return diag.FromErr(err)
        }
    }

    return nil
}
```

#### 4.3 Update 操作
```go
func resourceZenlayerCloudNewServiceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*connectivity.ZenlayerCloudClient)

    if d.HasChange("param_name") {
        request := newService.NewModifyResourceRequest()
        request.ResourceId = &[]string{d.Id()}[0]
        request.ParamName = &[]string{d.Get("param_name").(string)}[0]

        _, err := client.WithNewServiceClient().ModifyResource(request)
        if err != nil {
            return diag.FromErr(err)
        }
    }

    return resourceZenlayerCloudNewServiceResourceRead(ctx, d, meta)
}
```

#### 4.4 Delete 操作
```go
func resourceZenlayerCloudNewServiceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*connectivity.ZenlayerCloudClient)

    request := newService.NewDeleteResourceRequest()
    request.ResourceId = &[]string{d.Id()}[0]

    _, err := client.WithNewServiceClient().DeleteResource(request)
    if err != nil {
        // 忽略"资源不存在"错误
        if !isNotFoundError(err) {
            return diag.FromErr(err)
        }
    }

    return nil
}
```

---

### 步骤 5：创建 Extension 文件

**文件：`zenlayercloud/services/newservice/extension_newservice.go`**

```go
package newservice

const (
    ResourceStatusPending  = "PENDING"
    ResourceStatusRunning  = "RUNNING"
    ResourceStatusStopped  = "STOPPED"
    ResourceStatusError    = "ERROR"
)

var (
    ResourceOperatingStatus = []string{
        ResourceStatusPending,
        ResourceStatusRunning,
        ResourceStatusStopped,
    }
)
```

---

### 步骤 6：注册到 Provider

**文件：`zenlayercloud/provider.go`**

1. **导入服务包**
```go
import (
    // ... 现有导入
    "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/newservice"
)
```

2. **在 resourcesMap() 中注册**
```go
func resourcesMap() map[string]*schema.Resource {
    return map[string]*schema.Resource{
        // ... 现有资源
        "zenlayercloud_newservice_resource": newservice.ResourceZenlayerCloudNewServiceResource(),
    }
}
```

3. **在 dataSourcesMap() 中注册（如果是数据源）**
```go
func dataSourcesMap() map[string]*schema.Resource {
    return map[string]*schema.Resource{
        // ... 现有数据源
        "zenlayercloud_newservice_resources": newservice.DataSourceZenlayerCloudNewServiceResources(),
    }
}
```

4. **更新文件顶部的文档注释**
```go
/*
...
New Service

  Data Source
    zenlayercloud_newservice_resources

  Resource
    zenlayercloud_newservice_resource
*/
```

---

### 步骤 7：更新 go.mod（如需要）

如果需要新的 SDK 版本：
```bash
go get github.com/zenlayer/zenlayercloud-sdk-go@v0.x.x
go mod tidy
```

---

### 步骤 8：编写测试

**文件：`zenlayercloud/services/newservice/resource_zc_newservice_resource_test.go`**

```go
package newservice

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccZenlayerCloudNewServiceResource_basic(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: testAccProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccNewServiceResourceBasic(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("zenlayercloud_newservice_resource.test", "param_name", "value"),
                ),
            },
        },
    })
}

func testAccNewServiceResourceBasic() string {
    return `
resource "zenlayercloud_newservice_resource" "test" {
    param_name = "value"
}
`
}
```

---

### 步骤 9：代码检查和构建

```bash
# 格式化代码
go fmt ./...

# 检查格式
make fmtcheck

# 构建
make build

# 运行测试
make test

# 运行验收测试（需要配置环境变量）
TF_ACC=1 go test ./zenlayercloud/services/newservice -v -run TestAcc -timeout 120m
```

---

### 步骤 10：生成文档

```bash
make gendoc
```

---

## 二、关键文件路径总结

| 操作 | 文件路径 | 说明 |
|------|---------|------|
| SDK 客户端 | `zenlayercloud/connectivity/client.go` | 添加新服务的客户端连接 |
| Provider 注册 | `zenlayercloud/provider.go` | 注册资源和数据源 |
| 资源实现 | `zenlayercloud/services/<service>/resource_zc_<service>_<resource>.go` | 资源 CRUD 操作 |
| 数据源实现 | `zenlayercloud/services/<service>/data_source_zc_<service>_<datasource>.go` | 数据源查询操作 |
| Extension | `zenlayercloud/services/<service>/extension_<service>.go` | 常量和枚举定义 |
| 测试文件 | `zenlayercloud/services/<service>/resource_zc_<service>_<resource>_test.go` | 单元测试和验收测试 |

---

## 三、重要约定和最佳实践

### 3.1 命名约定

- **资源类型名**：`zenlayercloud_<service>_<resource>`（全小写，下划线分隔）
- **资源函数名**：`ResourceZenlayerCloud<Service><Resource>`（大驼峰，导出）
- **CRUD 函数名**：`resourceZenlayerCloud<Service><Resource><Operation>`（小驼峰，私有）
- **数据源类型名**：`zenlayercloud_<service>_<datasource>s`（通常复数形式）

### 3.2 Schema 字段类型

```go
// 必需参数
"name": {
    Type:        schema.TypeString,
    Required:    true,
    Description: "描述",
}

// 可选参数
"optional_param": {
    Type:         schema.TypeString,
    Optional:     true,
    Computed:     true,  // 可同时设置，表示可选但由服务生成默认值
    ForceNew:     true,  // 改变此参数需要重建资源
    Description:  "描述",
}

// 计算属性（只读）
"computed_attr": {
    Type:        schema.TypeString,
    Computed:    true,
    Description: "描述",
}

// 列表类型
"tags": {
    Type:        schema.TypeMap,
    Optional:    true,
    Computed:    true,
    Description: "标签",
}

// 对象列表
"network_interfaces": {
    Type:        schema.TypeList,
    Computed:    true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "ip_address": {Type: schema.TypeString, Computed: true},
        },
    },
}
```

### 3.3 错误处理

```go
// 标准错误处理
if err != nil {
    return diag.Errorf("failed to create resource: %s", err)
}

// 资源不存在处理（Read 和 Delete 中）
if isNotFoundError(err) {
    d.SetId("")
    return nil
}
```

### 3.4 标签管理（Tags）

```go
import common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"

// 从 Terraform state 获取标签
if tags := common2.GetTags(d, "tags"); len(tags) > 0 {
    request.Tags = &newService.TagAssociation{}
    for k, v := range tags {
        request.Tags.Tags = append(request.Tags.Tags, &newService.Tag{
            Key:   &k,
            Value: &v,
        })
    }
}
```

### 3.5 超时配置

```go
Timeouts: &schema.ResourceTimeout{
    Create: schema.DefaultTimeout(10 * time.Minute),
    Update: schema.DefaultTimeout(10 * time.Minute),
    Delete: schema.DefaultTimeout(10 * time.Minute),
}

// 使用超时
timeout := d.Timeout(schema.TimeoutCreate)
```

---

## 四、常见场景处理

### 4.1 异步操作状态等待

```go
stateConf := &resource.StateChangeConf{
    Pending: []string{"PENDING", "CREATING"},
    Target:  []string{"RUNNING"},
    Refresh: resource.StateRefreshFunc(func() (interface{}, string, error) {
        resp, err := client.WithNewServiceClient().DescribeResource(request)
        if err != nil {
            return nil, "", err
        }
        return resp, *resp.Status, nil
    }),
    Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
    Delay:      10 * time.Second,
    MinTimeout: 5 * time.Second,
}

_, err = stateConf.WaitForStateContext(ctx)
```

### 4.2 分页查询（数据源）

```go
func dataSourceZenlayerCloudNewServiceResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*connectivity.ZenlayerCloudClient)

    var allResources []*newService.Resource
    var marker *string

    for {
        request := newService.NewDescribeResourcesRequest()
        request.Marker = marker
        request.PageSize = &[]int64{100}[0]

        response, err := client.WithNewServiceClient().DescribeResources(request)
        if err != nil {
            return diag.FromErr(err)
        }

        allResources = append(allResources, response.ResourceSet...)

        if response.Marker == nil {
            break
        }
        marker = response.Marker
    }

    // 转换并设置到 state
    resources := make([]map[string]interface{}, len(allResources))
    for i, r := range allResources {
        resources[i] = map[string]interface{}{
            "id":   *r.ResourceId,
            "name": *r.Name,
        }
    }

    d.SetId(dataSourceHash(resources))
    if err := d.Set("resources", resources); err != nil {
        return diag.FromErr(err)
    }

    return nil
}
```

---

## 五、验证清单

完成添加新 API 接口后，需要验证：

- [ ] `make fmtcheck` 通过
- [ ] `make build` 成功
- [ ] `make test` 通过
- [ ] `TF_ACC=1 go test ...` 验收测试通过（如果配置了测试环境）
- [ ] `make gendoc` 生成文档
- [ ] 资源可以创建
- [ ] 资源可以读取
- [ ] 资源可以更新（如果有可更新字段）
- [ ] 资源可以删除
- [ ] 资源可以导入（`terraform import`）
- [ ] 数据源可以查询
- [ ] 文档注释完整且准确
- [ ] 示例代码可运行

---

## 六、参考示例

### 完整资源示例：BMC EIP
- 文件：`zenlayercloud/resource_zc_bmc_eip.go`
- 包含完整的 CRUD 实现、状态等待、标签处理

### 完整服务示例：ZEC
- 目录：`zenlayercloud/services/zec/`
- 包含多个资源和数据源的最佳实践

---

## 总结

添加新 API 接口到 Terraform Provider Zenlayer 的核心工作：

1. **SDK 集成** - 添加客户端连接
2. **资源实现** - 编写 CRUD 操作
3. **Provider 注册** - 在 provider.go 中注册
4. **测试编写** - 确保功能正确
5. **文档更新** - 更新注释和示例

整个过程遵循 Terraform Plugin SDK v2 的标准模式，结合 Zenlayer Cloud SDK 的特性实现完整的资源管理功能。
