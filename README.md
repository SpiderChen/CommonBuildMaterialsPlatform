# CommonBuildMaterialsPlatform

本仓库包含三个边界清晰的项目：

- `ERP/`：客户侧私有化建材 ERP，部署在客户服务器，承载销售、生产、实验室、调度、地磅、签收、结算、采购、库存、财务和报表等日常业务。
- `OperationsPlatform/`：厂商内部产品运营平台，部署在我方运维环境，用来管理客户部署实例、授权续费、系统异常告警、客户端/服务端更新包和 GPS 转发器。
- `industrial-control-gateway/`：现场工控数据接入服务，可按客户现场需要独立部署，用于采集或转发 PLC、OPC、Modbus、边缘机等设备数据。

## 项目边界

ERP 是最终客户使用的业务系统，不应该出现“产品运营台”入口。客户实例台账、授权续费运营、客户现场探针、第三方监控接入、内部告警治理、灰度发布和端内更新任务编排都属于 `OperationsPlatform/`。

`OperationsPlatform/` 不承载 ERP 生产业务，也不直接修改客户业务数据。它只从厂商视角管理交付对象、授权、异常、版本和更新包。

`industrial-control-gateway/` 不依赖 ERP 前端，也不承载运营台；它只负责现场设备数据的接入、归一化和转发。

## ERP 核心能力

以下是 ERP 的目标业务范围；当前完成度和待补深清单见 [`docs/FUNCTION_AUDIT.md`](docs/FUNCTION_AUDIT.md)。

- 经营驾驶舱：订单、发货、车辆、质量、库存和财务指标总览。
- 基础资料：集团、公司/分公司、站点、客户、项目、产品、物料、司机、车辆、承运商、仓库和库存，按集团总部、公司和站点数据范围维护。
- 销售到收款：合同、订单、签收、对账、应收、开票、收款和催收。
- 生产到发货：生产计划、任务、批次、调度、地磅、过磅记录、车辆轨迹和工地签收。
- 实验室与质量：基础配比修订链、生产线配比、试配、原料检验、生产试样、样品试验、仪器校准和质量异常闭环。
- 采购到付款：采购申请、采购订单、入库、库存流水、供应商对账、应付和付款。
- 系统底座：授权校验、插件、审批流、数据字典、备份恢复、API Gateway、SSE/Redis/RabbitMQ 事件。

## OperationsPlatform 核心能力

- 客户部署台账：客户名称、授权 ID、部署地址、联系人、客户端/服务端版本和健康状态。
- 授权续费：登记新到期日、版本、站点/车辆额度，并写入续费和审计记录。
- 系统异常告警：按客户、来源、等级追踪打开、确认和关闭状态。
- 更新包管理：区分客户端和服务端更新包，支持创建、发布和分配到客户。
- GPS 转发器：接收客户侧 GPS/北斗终端、第三方 GPS 平台或串口落盘定位，并转发到客户私有化 ERP。

## 本地启动

ERP 后端：

```bash
cd ERP
./backend/scripts/dev.sh
```

ERP 前端：

```bash
cd ERP
./frontend/scripts/dev.sh
```

运营平台：

```bash
cd OperationsPlatform
./scripts/dev.sh
```

通用工控对接：

```bash
cd industrial-control-gateway
./scripts/dev.sh
```

## 构建

ERP 后端：

```bash
cd ERP
./backend/scripts/build.sh
```

ERP 前端：

```bash
cd ERP
./frontend/scripts/web-build.sh
```

运营平台：

```bash
cd OperationsPlatform
./scripts/build.sh
```

通用工控对接：

```bash
cd industrial-control-gateway
./scripts/build.sh
```
