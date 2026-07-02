# Backend: Common Build Materials ERP

`backend/` 是客户侧私有化建材 ERP 的 Go 服务端项目，负责客户业务 API、加密数据仓、权限、审计、系统配置、备份和现场运行能力。

ERP 服务端面向最终客户的销售、生产、实验室、调度、地磅、签收、结算、采购、库存和财务管理。它不承载厂商内部“产品运营台”；客户实例、授权续费运营、客户现场探针、内部告警治理、更新包灰度和端内更新任务编排归属 `../../OperationsPlatform/`。

## 核心能力

- 经营驾驶舱：订单、发货、车辆、质量、库存和财务指标总览。
- 基础资料：公司、站点、客户、项目、产品、物料、司机、车辆、承运商、仓库和库存。
- 销售到收款：合同、订单、签收、对账、应收、开票、收款和催收。
- 生产到发货：生产计划、任务、批次、调度、地磅、票据、车辆轨迹和工地签收。
- 实验室与质量：配比版本链、试配、原料检验、生产试块、样品试验、仪器校准和质量异常闭环。
- 采购到付款：采购申请、采购订单、入库、库存流水、供应商对账、应付和付款。
- 系统底座：授权校验、插件、审批流、数据字典、备份恢复、API Gateway、SSE/Redis/RabbitMQ 事件。

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

产物位于 `backend/dist/cbmp-appliance` 和 `backend/dist/cbmp-server-updater`。服务端 updater 只负责客户现场服务端组件的本端更新执行；跨客户实例的运营编排应由 `OperationsPlatform` 承担。

## 接口总览

- 根目录对外接口入口参见：[根目录接口总览](../../docs/ROOT_API_SURFACE.md)

## 渐进式模块开发

新增 ERP 业务能力优先按领域拆到 `backend/internal/appliance/`：

- `dashboard.go`
- `master_data.go`
- `contracts.go`
- `orders.go`
- `production.go`
- `laboratory.go`
- `dispatch.go`
- `weighbridge.go`
- `delivery.go`
- `statements.go`
- `procurement.go`
- `finance.go`
- `reports.go`
- `system.go`

外部只通过 HTTP API、SSE 和设备接入协议通信；后端不反向依赖 `frontend/`。
