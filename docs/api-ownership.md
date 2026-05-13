# API 目录与服务归属规范

## 1. 目的

本文用于统一接口目录、网关路由、权限映射和服务注册的口径。

当前版本线不再使用业务域。API 归属只表达两个问题：

- 这个接口由哪个服务提供：通过 `sys_service_registry.http_prefix -> service_code` 推导
- 这个接口需要哪个资源组权限：`resources_group`

## 2. 核心规则

一个网关可见 API 由 `path + method` 唯一定位。

`sys_api_resources` 中同一个 `path + method` 只能有一条记录。它映射为：

```text
path + method -> resources_group
```

`resources_group` 用于角色授权。`service_code` 不存放在 API 目录表中，由网关根据服务注册表的 HTTP 前缀解析得到，用于鉴权命名空间和排查定位。

## 3. 字段语义

### 3.1 `path`

网关可见的 HTTP 路径，例如：

- `/admin-api/v1/apis`
- `/common-api/v1/site-messages/my`

路径中的变量使用 proto/OpenAPI 风格：

```text
/admin-api/v1/apis/{id}
```

### 3.2 `method`

HTTP 方法，例如：

- `GET`
- `POST`
- `PUT`
- `DELETE`

当前 Casbin 策略仍按 HTTP method 匹配。

### 3.3 `service_code`

接口技术服务归属，不属于 `sys_api_resources` 字段。

当前服务编码：

- `auth`
- `user`
- `admin`
- `common`

网关优先按 `sys_service_registry.http_prefix` 识别请求所属服务；未命中时回退到内置前缀。服务注册表自身仍保留 `service_code`。

### 3.4 `resources_group`

接口授权资源组。角色拥有某资源组上的 method 权限后，可以访问映射到该资源组的 API。

示例：

- `default`
- `user`
- `role`
- `menu`
- `api`
- `data`
- `admin`
- `site_message_manage`

## 4. 与服务注册的关系

`sys_service_registry` 是服务元数据和前缀解析表，主要字段：

- `service_code`
- `service_name`
- `http_prefix`
- `grpc_service`
- `status`
- `projection_enabled`

网关使用 `http_prefix -> service_code` 推导鉴权命名空间。

## 5. 与权限投影的关系

`sys_api_resources` 是 API 目录主数据。

`auth.casbin_rules` 是投影，不是主数据。修改 proto 或 HTTP route 不会自动改变权限，必须同步 API 目录并触发 `admin -> auth` 投影。

## 6. 新增 API 流程

新增或调整 API 时：

1. 确认 `path + method` 是否已存在。
2. 选择或新增合适的 `resources_group`。
3. 如果这是新的服务前缀，先在 `sys_service_registry` 登记 `http_prefix` 和 `service_code`。
4. 在角色/资源关系中授权给对应角色。
5. 触发权限投影同步。
6. 如果前端调用该接口，同步 OpenAPI 和生成客户端。

## 7. 已移除内容

以下内容不再属于当前设计：

- 业务域 owner
- API 资源表里的 `business_key`
- API 资源表里的 `service_key`
- API 资源表里的 `service_code`
- `domain_code`
- `sys_business_domain`
- `/admin-api/v1/platform/domains`
- `/system/platform/domain`
