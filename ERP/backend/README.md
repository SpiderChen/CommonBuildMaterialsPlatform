# Backend: Product Ops Appliance

`backend/` 是独立 Go 服务端项目，负责产品运营交付平台的 API、ERP 实验室管理、加密数据仓、授权、告警、续费、更新包、备份和私有化运行能力。

它不是 SaaS 服务端。平台对象是客户产品实例；客户实例直接承载授权、续费、现场探针、异常报警、客户端版本、服务端版本和端内系统更新任务。

## 核心能力

- 产品运营：客户实例、授权续费风险、系统异常报警、更新包和端内系统更新总览。
- 实验室管理：配比版本链、试配记录、配比审批生效/停用、生产试块、原料检验、样品试验、仪器校准、质量异常闭环和报表分析。
- 授权中心：Ed25519 离线 license 签发、下载、续期、导入验签、吊销和授权门户。
- 续费闭环：续费任务、报价、审批、合同、电子签、回款、开票、外部集成同步、失败重试和 HMAC 回调验签。
- 告警闭环：探针、日志/APM/trace、第三方监控事件、规则匹配、聚合抑制、升级、SSE/Webhook/企业微信/短信/ITSM 通道、Bearer token、HMAC 签名和通知重试。
- 更新中心：客户端/服务端更新包发布真实 full artifact 和 `cbmp-copy-v1` 差分 patch、patch/target SHA256 校验、Ed25519 非对称包签名、HMAC 兼容签名、受控下载、验签、应用、回滚、灰度批次、受控执行、端内系统更新任务下发和回执。
- 基础资料：客户、项目、产品、物料、司机、车辆、承运商、站点和库存可通过 `/api/master/{resource}` 增量维护，写入加密数据仓、审计日志和 SSE 事件。
- 私有化运行：AES-GCM 本地 vault、PostgreSQL 加密快照、备份恢复、API Gateway 配置、SSE/Redis/RabbitMQ 事件。

## 运行

```bash
./backend/scripts/dev.sh
```

默认监听 `127.0.0.1:8088`。

后端默认只提供 `/api/*` 和 `/api/events`，不托管前端静态文件。如需在单体交付包中由后端托管已构建 Web 前端，可显式指定：

```bash
CBMP_FRONTEND_DIR=../frontend/dist ./backend/scripts/dev.sh
# 或
./backend/dist/cbmp-appliance --frontend ../frontend/dist
```

## 验证

```bash
./backend/scripts/test.sh
```

## 构建

```bash
./backend/scripts/build.sh
```

产物位于 `backend/dist/cbmp-appliance` 和 `backend/dist/cbmp-server-updater`；其中 `cbmp-server-updater` 只处理 `component=server` 的端内更新任务。

## 接口总览

- 根目录对外接口入口参见：[根目录接口总览](../../docs/ROOT_API_SURFACE.md)

## 渐进式模块开发

新增产品运营能力优先按领域拆到 `backend/internal/appliance/`：

- `laboratory.go`
- `product_ops.go`
- `product_probe.go`
- `product_telemetry.go`
- `product_monitoring.go`
- `product_alert_governance.go`
- `product_renewal_enterprise.go`
- `product_renewal_integrations.go`
- `product_update_executor.go`
- `product_system_update.go`
- `master_create.go`

外部只通过 HTTP API、SSE 和端内 updater 协议通信；后端不反向依赖 `frontend/`。
