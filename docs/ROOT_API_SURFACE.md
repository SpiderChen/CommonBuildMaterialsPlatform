# 根目录接口总览（按项目）

更新时间：2026-06-19

该文档用于对齐三个独立项目的最外层接口入口，避免前端、网关和脚本对接口边界各自“单独记忆”。

## 1) ERP 后端（`ERP/backend/`）

- 根路由：`/api`
- 健康检查：`GET /api/health`
- 实时事件：`GET /api/events`
- 认证：
  - `POST /api/auth/login`
  - `GET|POST /api/auth/sso/...`（`/providers`, `/sso/<provider>/start`）
- `sso` 回调：
  - `POST|GET /api/auth/sso/<provider>/callback`
- 身份同步：`/api/scim/v2/...`
- 回调：
  - `POST /api/product-ops/renewals/sync-callback`
  - `POST /api/finance/tax/callback`
  - `POST /api/finance/collections/callback`
- 客户现场上报：
  - `POST /api/product-ops/probes/report`
  - `POST /api/iot/vehicle/location/report`（地磅或设备事件上报）
  - `POST /api/iot/protocols/gps/ingest`（工控网关/GPS设备接入）
  - `POST /api/weighbridge/protocols/scale/ingest`（地磅上报）
  - `POST /api/product-ops/telemetry/report`
  - `POST /api/product-ops/monitoring/report`
  - `POST /api/product-ops/system-updates/tasks/{taskNo}/report`
- 产品运营：`/api/product-ops/{subpath}`（例如：`overview`, `instances`, `alerts`, `renewals`, `rollouts`, `system-updates`, `telemetry`, `monitoring`）
- 系统：`/api/system/{subpath}`（例如：`updates`, `license`, `gateway`, `backups`, `runtime`, `approval-flows`, `dictionaries`）
- 通用业务域：`/api/auth`, `/api/bootstrap`, `/api/dashboard`, `/api/master/*`, `/api/orders`, `/api/contracts`, `/api/finance`, `/api/dispatch*`, `/api/delivery`, `/api/weighbridge`, `/api/vehicle`, `/api/iot`, `/api/rules`, `/api/integrations`, `/api/approvals`, `/api/portal`, `/api/public`, `/api/production-plans`, `/api/laboratory`, `/api/quality`, `/api/reports`
- 下载与任务特例：
  - `GET /api/system/updates/{id}/download`（需 updater token）
  - `GET /api/product-ops/system-updates/tasks`（端内 updater 轮询任务）

## 2) Operations 平台（`OperationsPlatform/backend/`）

- 根路由：`/api`
- 健康检查：`GET /api/health`
- 概览：`GET /api/summary`
- 客户：
  - `GET /api/customers`
  - `POST /api/customers/{id}/renewals`
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

## 3) 工控网关（`industrial-control-gateway/`）

- HTTP：
  - `GET /healthz`
  - `GET /metrics`
  - `POST /ingest`
- 非 HTTP：
  - TCP 行协议监听（`ICG_TCP_ADDR`）
  - 文件轮询（`ICG_FILE`）

## 4) 前端调用点对账（`ERP/frontend/src/services/api.ts`）

- 生成口径：`ERP/frontend/src/services/api.ts` 中每个 `this.request(...)` 调用。
- 说明：未显式指定 `method` 的接口按 `GET` 计算。
- 维护要求：当前端新增、删除或改签名前，应先更新本表并完成对应后端路由对齐。

| 方法 | 前端调用路径 |
| --- | --- |
| DELETE | /api/vehicle/fences/${id} |
| GET | /api/approvals |
| GET | /api/auth/sso/providers |
| GET | /api/bootstrap |
| GET | /api/dashboard |
| GET | /api/delivery/sign |
| GET | /api/delivery/sign-attachments |
| GET | /api/delivery/sign-links |
| GET | /api/dispatch-center/overview |
| GET | /api/dispatch-orders |
| GET | /api/dispatch-orders/carrier-settlements |
| GET | /api/dispatch-orders/schedules |
| GET | /api/finance/invoices/${id}/download |
| GET | /api/finance/overview |
| GET | /api/integrations/overview |
| GET | /api/laboratory/overview |
| GET | /api/master/export?resource=${encodeURIComponent(resource)} |
| GET | /api/orders |
| GET | /api/portal/complaints |
| GET | /api/portal/overview |
| GET | /api/procurement/overview |
| GET | /api/product-ops/overview |
| GET | /api/production-plans/overview |
| GET | /api/public/delivery-sign/${token} |
| GET | /api/quality/overview |
| GET | /api/reports |
| GET | /api/rules |
| GET | /api/statements |
| GET | /api/system/approval-flows |
| GET | /api/system/backups |
| GET | /api/system/backups/drills |
| GET | /api/system/dictionaries |
| GET | /api/system/gateway |
| GET | /api/system/license/issues |
| GET | /api/system/license/packages |
| GET | /api/system/license/packages/${id}/download |
| GET | /api/system/license/portal |
| GET | /api/system/license/revocations |
| GET | /api/system/license/verify |
| GET | /api/system/map-config |
| GET | /api/system/plugins |
| GET | /api/system/plugins/runs |
| GET | /api/system/runtime |
| GET | /api/system/security |
| GET | /api/system/updates |
| GET | /api/system/updates/${id}/download |
| GET | /api/vehicle/alarms |
| GET | /api/vehicle/fence-events${suffix} |
| GET | /api/vehicle/fences |
| GET | /api/vehicle/location/latest |
| GET | /api/vehicle/track/replay?vehicleId=${vehicleId} |
| GET | /api/weighbridge/device-events |
| GET | /api/weighbridge/ticket-prints |
| GET | /api/weighbridge/ticket-voids |
| GET | /api/weighbridge/tickets |
| GET | /api/weighbridge/weight-records |
| POST | /api/approvals/${id}/act |
| POST | /api/auth/login |
| POST | /api/auth/sso/${providerCode}/start |
| POST | /api/contracts/${contractId}/attachments |
| POST | /api/contracts/${id}/revise |
| POST | /api/contracts/${id}/submit |
| POST | /api/laboratory/equipment |
| POST | /api/laboratory/equipment/${id}/calibrations |
| POST | /api/laboratory/exceptions |
| POST | /api/laboratory/exceptions/${id}/handle |
| POST | /api/laboratory/mix-designs |
| POST | /api/laboratory/mix-designs/${id}/approve |
| POST | /api/laboratory/mix-designs/${id}/retire |
| POST | /api/laboratory/mix-designs/${id}/revise |
| POST | /api/laboratory/mix-designs/${id}/trial-runs |
| POST | /api/laboratory/samples |
| POST | /api/laboratory/samples/${sampleId}/tests |
| POST | /api/laboratory/tests/${id}/review |
| POST | /api/delivery/sign |
| POST | /api/delivery/sign-links |
| POST | /api/delivery/sign/${signId}/attachments |
| POST | /api/dispatch-orders |
| POST | /api/dispatch-orders/${id}/status |
| POST | /api/dispatch-orders/carrier-settlements/generate |
| POST | /api/dispatch-orders/schedules |
| POST | /api/finance/collection-templates |
| POST | /api/finance/collections/${id}/handle |
| POST | /api/finance/collections/${id}/send |
| POST | /api/finance/collections/generate |
| POST | /api/finance/payment-plans |
| POST | /api/finance/payment-plans/${id}/settle |
| POST | /api/finance/receipts |
| POST | /api/finance/red-letters |
| POST | /api/finance/red-letters/${id}/approve |
| POST | /api/finance/supplier-statements |
| POST | /api/finance/supplier-statements/${id}/approve |
| POST | /api/iot/vehicle/location/batch |
| POST | /api/master/carriers |
| POST | /api/master/customer-blacklists |
| POST | /api/master/customer-blacklists/${id}/release |
| POST | /api/master/customer-complaints |
| POST | /api/master/customer-complaints/${id}/close |
| POST | /api/master/customer-contacts |
| POST | /api/master/customer-contacts/${id}/default |
| POST | /api/master/customer-profiles |
| POST | /api/master/customer-profiles/evaluate |
| POST | /api/master/customers |
| POST | /api/master/drivers |
| POST | /api/master/import |
| POST | /api/master/inventory |
| POST | /api/master/materials |
| POST | /api/master/price-policies |
| POST | /api/master/pricing/evaluate |
| POST | /api/master/products |
| POST | /api/master/projects |
| POST | /api/master/sites |
| POST | /api/master/tax-rates |
| POST | /api/master/vehicles |
| POST | /api/orders |
| POST | /api/orders/${id}/approve |
| POST | /api/portal/complaints |
| POST | /api/portal/dispatches/${id}/exception |
| POST | /api/procurement/receipts |
| POST | /api/procurement/stocktakes |
| POST | /api/procurement/stocktakes/${id}/review |
| POST | /api/procurement/transfers |
| POST | /api/procurement/transfers/${id}/complete |
| POST | /api/product-ops/alerts |
| POST | /api/product-ops/alerts/${id}/escalate |
| POST | /api/product-ops/alerts/${id}/handle |
| POST | /api/product-ops/alerts/channels |
| POST | /api/product-ops/alerts/notifications/${id}/retry |
| POST | /api/product-ops/alerts/policies |
| POST | /api/product-ops/instances |
| POST | /api/product-ops/monitoring/integrations |
| POST | /api/product-ops/monitoring/report |
| POST | /api/product-ops/monitoring/rules |
| POST | /api/product-ops/renewals |
| POST | /api/product-ops/renewals/${id}/approval |
| POST | /api/product-ops/renewals/${id}/close |
| POST | /api/product-ops/renewals/${id}/contract |
| POST | /api/product-ops/renewals/${id}/esign |
| POST | /api/product-ops/renewals/${id}/invoice |
| POST | /api/product-ops/renewals/${id}/payment |
| POST | /api/product-ops/renewals/${id}/quote |
| POST | /api/product-ops/renewals/integrations |
| POST | /api/product-ops/renewals/sync-records/${id}/retry |
| POST | /api/product-ops/rollouts |
| POST | /api/product-ops/rollouts/${id}/advance |
| POST | /api/product-ops/rollouts/${id}/system-update-tasks |
| POST | /api/product-ops/rollouts/${id}/execute |
| POST | /api/product-ops/telemetry/report |
| POST | /api/product-ops/system-updates/tasks |
| POST | /api/product-ops/system-updates/tasks/${encodeURIComponent(taskNo)}/report |
| POST | /api/production-plans/${planId}/tasks |
| POST | /api/production-plans/reports/generate |
| POST | /api/production-plans/tasks/${taskId}/batches |
| POST | /api/public/delivery-sign/${token} |
| POST | /api/quality/inspections |
| POST | /api/quality/raw-inspections |
| POST | /api/quality/raw-inspections/${id}/review |
| POST | /api/quality/samples/${id}/test |
| POST | /api/rules/alarms/${id}/handle |
| POST | /api/statements/${id}/confirm |
| POST | /api/system/approval-flows |
| POST | /api/system/approval-flows/${id}/status |
| POST | /api/system/backups |
| POST | /api/system/backups/${encodeURIComponent(name)}/restore |
| POST | /api/system/backups/drills |
| POST | /api/system/dictionaries |
| POST | /api/system/dictionaries/${id}/status |
| POST | /api/system/field-policies |
| POST | /api/system/field-policies/${id}/toggle |
| POST | /api/system/gateway |
| POST | /api/system/gateway/reload |
| POST | /api/system/gateway/routes/${id}/canary |
| POST | /api/system/gateway/routes/${id}/drain |
| POST | /api/system/gateway/routes/${id}/status |
| POST | /api/system/license/import |
| POST | /api/system/license/issues |
| POST | /api/system/license/packages/${id}/renew |
| POST | /api/system/license/revoke |
| POST | /api/system/mfa/users/${userId}/disable |
| POST | /api/system/mfa/users/${userId}/enable |
| POST | /api/system/mfa/users/${userId}/enroll |
| POST | /api/system/org/${resource}/${id}/status |
| POST | /api/system/org/companies |
| POST | /api/system/org/departments |
| POST | /api/system/plugins/${pluginId}/run |
| POST | /api/system/plugins/install |
| POST | /api/system/scim/providers |
| POST | /api/system/scim/providers/${id}/status |
| POST | /api/system/sso/providers |
| POST | /api/system/sso/providers/${id}/status |
| POST | /api/system/updates |
| POST | /api/system/updates/${id}/apply |
| POST | /api/system/updates/${id}/rollback |
| POST | /api/vehicle/fences |
| POST | /api/weighbridge/tickets |
| POST | /api/weighbridge/tickets/${id}/reprint |
| POST | /api/weighbridge/tickets/${id}/void/approve |
| POST | /api/weighbridge/tickets/${id}/void/request |
| POST | /api/weighbridge/tickets/return |
| POST | /api/weighbridge/tickets/transfer |
| POST | /api/weighbridge/tickets/waste |
| PUT | /api/vehicle/fences/${id} |

## 5) 执行规则（本清单更新触发条件）

1. 新增或变更任何对外 API，必须先更新本清单，随后再更新对应领域 `README`。
2. 前端新增调用点：
   - 若新增了 `ERP/frontend/src/services/api.ts` 接口，必须补齐 `ERP/backend/internal/appliance` 对应路由处理分支并回写本清单。
3. 发布前一次性对齐：
   - `docs/ROOT_API_SURFACE.md`
   - 运行时代码（`Routes()`/`apiHandler` 入口）
   - `README` 与部署文档中的接口入口说明
