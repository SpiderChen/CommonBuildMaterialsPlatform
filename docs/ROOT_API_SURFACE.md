# 根目录接口总览（按项目）

更新时间：2026-06-20

该文档用于对齐三个独立项目的最外层接口入口，避免前端、网关和脚本把 ERP 与内部运营平台混在一起。

## 1. ERP 后端（`ERP/backend/`）

- 根路由：`/api`
- 产品身份：`common-build-materials-erp`
- 健康检查：`GET /api/health`
- 实时事件：`GET /api/events`
- 认证：
  - `POST /api/auth/login`
  - `GET|POST /api/auth/sso/...`
- 身份同步：`/api/scim/v2/...`
- 回调：
  - `POST /api/finance/tax/callback`
  - `POST /api/finance/collections/callback`
- 设备与现场上报：
  - `POST /api/iot/vehicle/location/report`
  - `POST /api/iot/protocols/gps/ingest`
  - `POST /api/weighbridge/protocols/scale/ingest`
- ERP 业务域：
  - `/api/bootstrap`
  - `/api/dashboard`
  - `/api/master/*`
  - `/api/contracts`
  - `/api/orders`
  - `/api/production-plans`
  - `/api/laboratory`
  - `/api/quality`
  - `/api/dispatch-center`
  - `/api/dispatch-orders`
  - `/api/weighbridge`
  - `/api/delivery`
  - `/api/statements`
  - `/api/procurement`
  - `/api/finance`
  - `/api/portal`
  - `/api/vehicle`
  - `/api/iot`
  - `/api/rules`
  - `/api/integrations`
  - `/api/approvals`
  - `/api/reports`
  - `/api/system`
  - `/api/public`
- 明确不属于 ERP：
  - `/api/product-ops/*`
  - 客户实例运营、授权续费运营、客户现场探针、内部告警治理、第三方监控接入、更新包灰度和端内更新任务编排属于 `OperationsPlatform/`。
  - ERP 默认构建不编译 `ERP/backend/internal/appliance/product_*.go` 的历史运营台处理器；这些文件仅保留在 `legacy_product_ops` build tag 下作为迁移期源码。

## 2. Operations 平台（`OperationsPlatform/backend/`）

- 根路由：`/api`
- 健康检查：`GET /api/health`
- 概览：`GET /api/summary`
- 客户：
  - `GET /api/customers`
  - `POST /api/customers`
  - `POST /api/customers/{id}/renewals`
  - `GET /api/renewals`
  - `GET /api/renewals/{id}/license-package`
- 告警：
  - `GET /api/alerts`
  - `POST /api/alerts/{id}/ack`
  - `POST /api/alerts/{id}/resolve`
- 更新包：
  - `GET /api/update-packages`
  - `POST /api/update-packages`
  - `POST /api/update-packages/{id}/publish`
  - `POST /api/update-packages/{id}/assign`
- 任务与审计：
  - `GET /api/assignments`
  - `GET /api/audit-logs`

## 3. 工控网关（`industrial-control-gateway/`）

- HTTP：
  - `GET /healthz`
  - `GET /metrics`
  - `POST /ingest`
- 非 HTTP：
  - TCP 行协议监听（`ICG_TCP_ADDR`）
  - 文件轮询（`ICG_FILE`）

## 4. 前端调用点对账（`ERP/frontend/src/services/api.ts`）

- 生成口径：`ERP/frontend/src/services/api.ts` 中每个 `this.request(...)` 调用。
- 说明：未显式指定 `method` 的接口按 `GET` 计算。
- 维护要求：当前端新增、删除或改签名前，应同步确认对应后端路由。
