# CommonBuildMaterialsOperationsPlatform

建材产品运营平台是内部使用的独立项目，用来运营已经私有化部署到客户服务器上的 `CommonBuildMaterialsPlatform` ERP 产品。

这个项目不承载 ERP 生产业务，也不引入 SaaS 租户模型。它管理的是客户部署实例、授权续费、系统异常告警、客户端更新包、服务端更新包和 GPS 转发器。

## 项目边界

- `CommonBuildMaterialsPlatform/`：客户侧私有化 ERP 产品，部署在客户服务器。
- `CommonBuildMaterialsOperationsPlatform/`：厂商内部运营平台，部署在我方运维环境。
- 两个项目代码、运行时、数据文件和部署脚本互相独立。

## 已包含能力

- 客户部署台账：从空台账创建客户部署，维护客户名称、授权 ID、部署地址、联系人、客户端/服务端版本、健康状态。
- 授权续费：登记新到期日、版本、站点/车辆额度，生成带 Ed25519 签名的 ERP 可导入授权包，并写入续费、下载和审计记录。
- 系统异常告警：支持运营端手工创建、客户侧/探针 token 上报、同类未关闭告警去重续报，并按客户、来源、等级追踪打开、确认和关闭状态。
- 更新包管理：区分客户端和服务端更新包，支持创建、发布、分配到客户，并提供客户端/服务端 updater 检查任务、下载包和回传执行结果的接口。
- 客户侧通信：updater、告警上报和更新包下载都使用客户部署台账中的 `updaterToken` 认证，不接受 `licenseId` 代替 token。
- GPS 转发器：接收客户侧 GPS / 北斗终端、第三方 GPS 平台或串口落盘定位，转发到客户私有化 ERP。
- 本地数据存储：后端使用 JSON 文件保存运营数据，默认路径为 `backend/data/ops.json`；新数据文件默认为空，只有显式设置 `CBM_OPS_SEED_DEMO=1` 才注入演示客户数据。
- 静态前端控制台：后端可直接托管 `frontend/`，无前端构建依赖。

## 本地运行

运营平台：

```bash
./scripts/dev.sh
```

默认访问地址：

```text
http://127.0.0.1:8095
```

GPS 转发器：

```bash
./scripts/gps-forwarder-dev.sh
```

默认监听：

```text
http://127.0.0.1:19102
```

## 验证

```bash
./scripts/test.sh
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `CBM_OPS_ADDR` | `:8095` | 后端监听地址 |
| `CBM_OPS_DATA` | `backend/data/ops.json` | 本地运营数据文件 |
| `CBM_OPS_FRONTEND_DIR` | `frontend` | 静态前端目录 |
| `CBM_OPS_SEED_DEMO` | 空 | 设置为 `1` 时，首次创建数据文件会写入演示客户、告警和更新包 |
| `CBM_OPS_LICENSE_ISSUER_PRIVATE_KEY` | 空 | 授权续费必需的 Ed25519 私钥；客户侧 ERP 必须配置对应公钥到 `CBMP_LICENSE_TRUSTED_PUBLIC_KEYS` |
| `CBM_OPS_UPDATE_SIGNING_SECRET` | 空 | 更新包 HMAC 签名密钥；创建更新包时必需，客户 updater 下载前由运营平台校验后才返回制品 |

GPS 转发器环境变量见 [GPS_FORWARDER.md](docs/GPS_FORWARDER.md)。

## API 摘要

- 根目录总览入口：[根目录接口总览](../docs/ROOT_API_SURFACE.md)

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/api/summary` | 运营概览 |
| `GET` | `/api/customers` | 客户部署台账 |
| `POST` | `/api/customers` | 新增客户部署台账 |
| `POST` | `/api/customers/{id}/renewals` | 登记授权续费 |
| `GET` | `/api/renewals` | 授权续费记录 |
| `GET` | `/api/renewals/{id}/license-package` | 下载 ERP 可导入的签名授权包 |
| `GET` | `/api/alerts` | 系统告警 |
| `POST` | `/api/alerts` | 运营端新增告警 |
| `POST` | `/api/alerts/{id}/ack` | 确认告警 |
| `POST` | `/api/alerts/{id}/resolve` | 关闭告警 |
| `POST` | `/api/product-ops/alerts/report` | 客户侧或探针上报告警/恢复状态 |
| `GET` | `/api/update-packages` | 更新包列表 |
| `POST` | `/api/update-packages` | 创建更新包 |
| `POST` | `/api/update-packages/{id}/publish` | 发布更新包 |
| `POST` | `/api/update-packages/{id}/assign` | 分配更新包 |
| `POST` | `/api/product-ops/system-updates/tasks` | 客户端/服务端 updater 检查待执行更新任务 |
| `GET` | `/api/system/updates/{id}/download` | updater 下载已分配更新包 |
| `POST` | `/api/product-ops/system-updates/tasks/{taskNo}/report` | updater 回传下载、应用、失败或回滚结果 |

## GPS 转发器

GPS 转发器已从 ERP 项目并入当前运营平台项目，源码位于：

```text
backend/cmd/gps-forwarder
backend/internal/gpsforwarder
```

它不读取客户 ERP 数据库，也不修改客户业务数据，只负责接收、校验、去重和转发定位帧。
