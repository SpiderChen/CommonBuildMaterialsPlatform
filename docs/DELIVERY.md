# 交付说明

## 交付状态

当前仓库按“客户侧 ERP + 厂商内部 OperationsPlatform + 独立工控接入服务”交付。

`ERP/` 是最终客户使用的私有化建材 ERP。它不再默认进入“产品运营台”，也不暴露产品运营页面；登录后进入 ERP 工作台，覆盖订单、生产、调度、地磅、签收、对账、采购、财务、审批、实验室管理和公共签收等客户侧能力。

`OperationsPlatform/` 是厂商内部使用的独立运营平台，负责客户部署台账、授权续费、系统异常告警、客户端/服务端更新包、更新分配和 GPS 转发器。

`industrial-control-gateway/` 是独立现场工控接入服务，可按客户现场拓扑单独部署。

## ERP 交付主线

- 基础资料：公司、站点、客户、项目、产品、物料、司机、车辆、承运商、库存、堆场和筒仓。
- 销售到收款：合同、订单、签收、对账、应收、开票、收款和催收。
- 生产到发货：生产计划、任务、批次、调度、地磅、过磅记录、车辆轨迹和工地签收。
- 实验室与质量：基础配比建档、生产线配比、试配、审批生效/停用、生产试样、原料检验、样品试验复核、仪器校准和质量异常闭环。
- 采购到付款：采购申请、采购订单、入库、库存流水、供应商对账、应付和付款。
- 系统底座：授权校验、插件、审批流、数据字典、备份恢复、API Gateway、SSE/Redis/RabbitMQ 事件。

## 项目边界

- `ERP/backend/`：客户侧 Go ERP 服务端，可独立运行、测试和构建。
- `ERP/frontend/`：客户侧 Wails + React + HeroUI 客户端，可独立运行和构建。
- `OperationsPlatform/`：厂商内部运营平台，独立运行、独立数据文件、独立部署脚本。
- `industrial-control-gateway/`：通用工控接入服务，可独立接入 PLC/OPC/Modbus/边缘机数据并转发。

项目之间通过 HTTP/SSE/TCP/文件协议通信，源码、依赖、构建脚本、Go module / npm package 和容器镜像保持独立。

## 默认端口

- ERP Backend API: `127.0.0.1:8088`
- ERP Frontend dev: 由 Wails/Vite 自动分配
- OperationsPlatform: `127.0.0.1:8095`
- Industrial control gateway HTTP: `127.0.0.1:19101`

## 演示账号

| 系统 | 角色 | 账号 | 密码 |
| --- | --- | --- | --- |
| ERP | 平台管理员 | `admin` | `admin123` |
| ERP | 调度员 | `dispatcher` | `dispatch123` |
| ERP | 司机 | `driver` | `driver123` |
| ERP | 客户用户 | `customer` | `customer123` |
| ERP | 实验室质检员 | `quality` | `quality123` |

OperationsPlatform 的演示账号见 `OperationsPlatform/README.md`。

## 核心演示流程

1. 启动 ERP 后端：`cd ERP && ./backend/scripts/dev.sh`。
2. 启动 ERP 前端：`cd ERP && ./frontend/scripts/dev.sh`。
3. 用 `admin/admin123` 登录，进入客户侧 ERP 工作台。
4. 在“ERP 工作台”查看订单、生产、调度、地磅、签收、对账、采购、财务和审批状态。
5. 在“实验室管理”查看配比、试配、样品、试验复核、仪器校准和质量异常闭环。
6. 需要内部运营能力时，单独启动 `OperationsPlatform`，不要把产品运营台挂回 ERP。

## 数据与授权

ERP 服务端数据默认保存到 `ERP/backend/data/app.vault`，采用 AES-GCM 加密。生产部署时请设置：

```bash
export CBMP_DATA_KEY="customer-strong-random-key"
```

设置 `CBMP_POSTGRES_DSN` 后，ERP 服务端可把加密快照写入 PostgreSQL，并通过领域行恢复数据。设置 `CBMP_BACKUP_DIR` 与 `CBMP_BACKUP_KEY` 后，可生成加密备份和恢复演练报告。
