# 产品运营交付平台

当前交付目标是一个单独的内部运营平台，不是 SaaS 系统。平台用于管理已经购买或正在使用我们产品的客户实例，核心工作是授权续费、系统异常报警、客户端更新包管理和服务端更新包管理。

(ERP)是最终客户在使用的系统
(OperationsPlatform)是我们用来维护客户单体系统的平台
(industrial-control-gateway)用于部署在工控电脑上采集数据的中间服务

平台部署后直接维护客户产品实例；客户实例承载授权、续费、异常告警、客户端版本、服务端版本和端内系统更新任务。

## 核心场景

- 客户实例管理：登记客户名称、授权 ID、水印、部署入口、客户端版本、服务端版本、授权到期日、续费负责人、续费阶段和最近心跳。
- 授权续费：签发 Ed25519 离线授权包、下载授权包、在线续期、导入验签、吊销授权，并用续费任务、报价、审批、电子签、合同、回款和发票记录跟踪负责人、阶段、金额、到期日和下次跟进时间；电子签、财务回款和税控开票支持外部集成配置、出站同步流水、失败记录和人工重试。
- 实验室管理：覆盖配比版本、试配记录、配比审批生效/停用、生产试块、原料检验、样品试验、仪器校准、质量异常闭环和质量报表；生产任务默认选用同站点当前生效配比，历史批次保留当时配比版本。
- 系统异常报警：接收和维护客户现场异常，支持客户现场探针、日志、APM、链路追踪和第三方监控平台事件自动上报资源指标、版本、错误数、队列积压、接口状态码、耗时和自定义指标，按可配置规则、聚合窗口、抑制时长、升级策略、企业微信/短信/ITSM/Webhook 通知通道、Token/HMAC 签名、超时控制、投递失败重试、严重级别、来源、客户实例和处理状态形成闭环。
- 客户端/服务端更新包管理：区分 `client` 与 `server` 更新包，支持发布真实 full artifact 和 `cbmp-copy-v1` 二进制差分包，服务端 SHA256 校验、目标包 SHA256 校验、Ed25519 非对称包签名、HMAC 兼容签名、受控下载、端内 updater 解包落盘、离线验签、差分重建、应用、回滚状态、下载次数、灰度批次创建、执行前预检、平台受控执行、客户端/服务端 updater 任务下发/轮询/回执、回滚和客户实例版本同步。
- 私有化交付运维：通过加密 vault 或 PostgreSQL 加密快照保存数据，支持运营台备份/恢复演练、API Gateway 路由/灰度/排空/reload 配置和 SSE 实时事件。

## 当前架构

```text
Wails + React + HeroUI 客户端
        |
        v
API Gateway / Nginx
        |
        v
Go Appliance 服务端
        |
        +-- product-ops      客户实例、续费风险、报价合同回款、异常报警总览
        +-- laboratory       配比管理、试配、样品试验、仪器校准、质量异常闭环
        +-- license-daemon   授权签发、续期、验签、吊销
        +-- update-center    客户端/服务端更新包、发布、下载、应用、回滚、灰度批次、受控执行审计、端内更新任务
        +-- backup-center    加密备份、恢复、演练报告
        +-- event-bus        SSE / Redis / RabbitMQ
        |
        v
AES-GCM vault / PostgreSQL / Redis / RabbitMQ

客户端 updater（随 frontend 客户端交付）
        |
        +-- 只执行 client 组件更新、断点续传下载包、校验包、蓝绿发布、失败自动回滚、回执进度和结果

服务端 updater（随 backend 服务端交付）
        |
        +-- 只执行 server 组件更新、断点续传下载包、校验包、蓝绿发布、失败自动回滚、回执进度和结果
```

## 项目边界

- `backend/`：独立 Go Appliance 服务端项目，提供产品运营、授权、更新、备份、系统接口和 `cbmp-server-updater` 服务端更新器。
- `frontend/`：独立 Wails + React + HeroUI 客户端项目，默认进入“产品运营台”，并交付 `cbmp-client-updater` 客户端更新器。
- `industrial-control-gateway/`：独立工控对接项目，可按客户现场需要单独部署。
- 根目录 `scripts/` 只做本地联调和交付编排，不承载业务实现。
- 前后端只通过 HTTP API 与 SSE 通信；后端不导入前端源码，前端不导入后端源码。
- 后端默认只提供 API，只有显式配置 `CBMP_FRONTEND_DIR` 或 `--frontend` 时才托管静态前端。

## 已实现主线

- `/api/product-ops/overview`：聚合客户实例、授权续费风险、系统异常报警、客户端/服务端更新包、授权门户、运行态和最近事件。
- `/api/product-ops/instances`：新增或更新客户产品实例。
- `/api/product-ops/instances/{id}/heartbeat`：接收客户现场心跳并更新在线状态。
- 产品运营台“交付初始化基础资料”：可维护客户、项目、产品、物料、司机、车辆、承运商、站点和库存；对应 `/api/master/{resource}` 新增记录会写入加密数据仓、审计日志和 SSE 事件。
- `/api/laboratory/overview`：聚合实验室 KPI、配比版本、试配记录、生产质检、原料质检、样品台账、试验记录、仪器校准、质量异常、可操作批次和入库单。
- `/api/laboratory/mix-designs` 与 `/api/laboratory/mix-designs/{id}/revise|approve|retire|trial-runs`：维护配比版本链、试配、审批生效、停用和当前生效版本唯一性。
- `/api/laboratory/samples`、`/api/laboratory/samples/{id}/tests`、`/api/laboratory/tests/{id}/review`：登记样品、记录试验并复核；不合格复核会自动生成质量异常。
- `/api/laboratory/equipment`、`/api/laboratory/equipment/{id}/calibrations`：维护实验室仪器和校准记录；过期或禁用仪器禁止用于新试验。
- `/api/laboratory/exceptions`、`/api/laboratory/exceptions/{id}/handle`：创建、分派和关闭质量异常，形成原因和纠正措施闭环。
- `/api/product-ops/probes/report`：客户现场探针使用实例 token 上报健康状态，平台自动更新实例、记录探针历史并生成或恢复告警。
- `/api/product-ops/telemetry/report`：客户现场日志、APM、trace 或 metric 事件使用实例 token 上报，平台自动保存事件、识别慢请求/错误/5xx 并生成异常告警。
- `/api/product-ops/monitoring/integrations`：维护 Prometheus、Grafana、Sentry、Zabbix 或自定义监控接入 token。
- `/api/product-ops/monitoring/rules`：维护产品运营专用告警规则，支持来源、组件、指标、操作符、阈值、级别和通知通道。
- `/api/product-ops/monitoring/report`：第三方监控系统使用接入 token 上报事件，平台按规则匹配并自动生成 `source=monitoring` 告警。
- `/api/product-ops/alerts`：创建客户现场异常报警。
- `/api/product-ops/alerts/policies`：维护告警聚合、抑制、升级和通知策略。
- `/api/product-ops/alerts/channels`：维护 SSE、本地、Webhook、企业微信、短信和 ITSM 等告警通知通道，记录 endpoint、token、secret、重试次数、超时、最近成功和最近错误；出站投递会按通道类型生成对应 JSON，并支持 Bearer token 与 HMAC 签名头。
- `/api/product-ops/alerts/notifications/{id}/retry`：对失败的告警通知执行人工重试，并更新投递审计。
- `/api/product-ops/alerts/{id}/escalate`：人工升级告警并写入通知审计。
- `/api/product-ops/alerts/{id}/handle`：处理报警并形成闭环。
- `/api/product-ops/renewals`：创建或更新授权续费任务。
- `/api/product-ops/renewals/{id}/quote`：为续费任务生成报价并推进任务阶段。
- `/api/product-ops/renewals/{id}/approval`：提交、通过或驳回续费报价/合同审批，并同步报价状态与客户实例续费阶段。
- `/api/product-ops/renewals/{id}/contract`：确认续费合同并关联已审批报价。
- `/api/product-ops/renewals/{id}/esign`：发送或完成续费合同电子签，记录签署人、签署链接、签名和签署时间。
- `/api/product-ops/renewals/{id}/payment`：登记续费回款，支持部分回款和全额回款自动关闭任务。
- `/api/product-ops/renewals/{id}/invoice`：按合同和回款生成续费发票记录，保存税率、税额、税控状态和发票文件地址。
- `/api/product-ops/renewals/integrations`：维护续费外部集成，支持 `esign`、`payment/finance`、`tax` 和 `all` 场景，生产可替换为真实电子签、财务和税控 HTTP endpoint。
- `/api/product-ops/renewals/sync-records/{id}/retry`：对失败的续费外部同步流水执行人工重试，并将结果回写电子签或发票状态。
- `/api/product-ops/renewals/sync-callback`：接收电子签、财务或税控平台异步回调，可按 `syncNo`、外部流水号或资源编号匹配同步记录；配置了 `secret` 的集成必须使用 HMAC 签名头。
- `/api/product-ops/renewals/{id}/close`：关闭续费任务并记录完成时间。
- `/api/product-ops/rollouts`：基于已发布更新包创建客户实例灰度批次。
- `/api/product-ops/rollouts/{id}/advance`：推进、失败标记或回滚灰度批次中的客户实例，并同步客户实例版本。
- `/api/product-ops/rollouts/{id}/execute`：对灰度批次中的客户实例执行更新预检、验签、分发、安装、健康检查和回滚，并记录执行步骤。
- `/api/product-ops/rollouts/{id}/system-update-tasks`：运营台把灰度批次目标下发为客户端或服务端端内更新任务，记录任务号、更新包、目标版本、token 提示和执行单。
- `/api/product-ops/system-updates/tasks`：客户端/服务端 updater 使用实例 token 轮询待执行任务，平台返回更新包下载地址、checksum、signature 和目标版本。
- `/api/product-ops/system-updates/tasks/{taskNo}/report`：客户端/服务端 updater 回传执行进度、结果、当前版本和错误信息，平台同步更新执行单、灰度批次和客户实例版本。
- `/api/system/updates`：发布或更新客户端/服务端更新包。
- 授权中心：支持 license 包签发、持久化下载、续期签发、导入验签、额度校验、激活历史、吊销列表和总部授权门户；续费链路已支持报价、审批、电子签、合同、回款、开票、外部同步流水和失败重试状态联动。
- 更新中心：支持更新包列表、发布真实 full 包体和 `cbmp-copy-v1` 差分 patch、自动计算/校验 artifact SHA256、配置 `CBMP_UPDATE_SIGNING_PRIVATE_KEY` 后生成 `ed25519` 非对称包签名，未配置时兼容 `hmac-sha256` 包签名，下载导出真实 artifact 或 patch、验签、应用、回滚状态、灰度批次、执行前预检、受控执行步骤审计、端内 updater 任务下发/轮询/解包落盘/离线验签/差分重建/回执、发布人、应用人和下载次数记录，并已区分客户端与服务端组件。
- API Gateway 配置中心：支持运营台维护网关路由、稳定/灰度 upstream、灰度比例、连接排空、启停状态、reload 计划、Nginx 配置片段和网关事件。
- 备份中心：支持运营台创建加密备份、查看备份文件、执行恢复、运行恢复演练并查看演练报告。

## 本地启动

后端：

```bash
./backend/scripts/dev.sh
```

前端：

```bash
./frontend/scripts/dev.sh
```

通用工控对接：

```bash
./industrial-control-gateway/scripts/dev.sh
```

端内 updater：

```bash
./backend/dist/cbmp-server-updater -print-sample-config
./frontend/build/bin/cbmp-client-updater -print-sample-config
```

## 构建

后端：

```bash
./backend/scripts/build.sh
```

前端桌面端：

```bash
./frontend/scripts/build.sh
```

通用工控对接：

```bash
./industrial-control-gateway/scripts/build.sh
```

端内 updater：

```bash
./backend/scripts/build.sh
./frontend/scripts/updater-build.sh
```

根目录交付编排：

```bash
./scripts/build.sh
```

## 数据与授权

服务端数据默认保存到 `backend/data/app.vault`，采用 AES-GCM 加密。生产部署时请设置：

```bash
export CBMP_DATA_KEY="customer-strong-random-key"
```

设置 `CBMP_POSTGRES_DSN` 后，服务端可把加密快照写入 PostgreSQL，并通过领域行恢复数据。设置 `CBMP_BACKUP_DIR` 与 `CBMP_BACKUP_KEY` 后，可生成加密备份和恢复演练报告。

## 交付文档

- [交付说明](./docs/DELIVERY.md)
- [私有化部署](./docs/DEPLOYMENT.md)
- [功能完成度审计](./docs/FUNCTION_AUDIT.md)
- [根目录接口总览](./docs/ROOT_API_SURFACE.md)
