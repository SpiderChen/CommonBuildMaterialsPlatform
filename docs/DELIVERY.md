# 交付说明

## 交付状态

当前交付目标是一个单独的产品运营交付平台，不是 SaaS 系统。平台面向内部运营、交付、客服和版本发布团队，用于管理使用我们产品的客户实例。

交付演示和私有化部署启动后直接维护客户产品实例，实例上承载授权、续费、异常告警、客户端版本、服务端版本和端内系统更新任务。

本次更新已完成产品运营主线：

- 新增产品运营总览接口 `/api/product-ops/overview`，聚合客户实例、授权续费任务、授权续费风险、系统异常报警、客户端/服务端更新包、授权门户、运行态和最近事件。
- 新增客户实例模型与接口，可登记客户名称、授权 ID、水印、部署入口、客户端版本、服务端版本、授权到期日、续费负责人、续费阶段和最近心跳。
- 基础资料增量维护已补齐产品运营台页面入口和 API，覆盖客户、项目、产品、物料、司机、车辆、承运商、站点和库存的新增闭环，所有新增记录都会进入加密数据仓、审计日志和 SSE 事件。
- 新增客户现场探针上报接口 `/api/product-ops/probes/report`，现场服务端或客户端可使用实例 token 上报版本、CPU、内存、磁盘、队列积压和错误数，平台会自动生成或恢复探针告警。
- 新增客户现场可观测事件上报接口 `/api/product-ops/telemetry/report`，日志、APM、trace 和 metric 事件可使用实例 token 上报，平台会自动识别慢请求、错误、5xx 并生成异常告警。
- 新增第三方监控接入和产品告警规则，Prometheus、Grafana、Sentry、Zabbix 或自定义系统可使用接入 token 上报指标事件，平台会按来源、组件、指标、操作符和阈值生成系统异常告警。
- 新增系统异常报警模型与接口，可手工创建异常，也可由探针、可观测事件和第三方监控自动创建，按客户实例关联、按严重级别展示，并支持聚合窗口、重复抑制、自动升级、人工升级、企业微信/短信/ITSM/Webhook 通知通道、Bearer token、HMAC 签名、通道超时、投递失败重试、通知审计和处理闭环。
- 续费任务已支持创建、更新和关闭，可跟踪负责人、阶段、金额、到期日、下次跟进时间和风险等级，并已补齐报价、审批、电子签、合同、回款、开票、外部集成同步、失败记录和人工重试商业闭环。
- 更新包已区分 `client` 与 `server` 组件，运营台可分别发布真实 full artifact 包体和 `cbmp-copy-v1` 二进制差分 patch、查看元信息、下载真实包体或 patch、服务端 SHA256 校验、目标包 SHA256 校验、Ed25519 非对称包签名、HMAC 兼容签名、应用，并可创建灰度批次按客户实例执行预检、平台受控更新、客户端/服务端 updater 任务下发/轮询/断点续传下载/解包落盘/离线验签/差分重建/失败自动回滚/回执、步骤审计、失败标记、回滚和同步版本。
- 授权中心保留 Ed25519 离线授权包签发、下载、续期、导入验签、额度校验、吊销和总部授权门户。
- 前端默认进入“产品运营台”，主界面围绕客户实例、授权续费、系统异常报警、客户端/服务端更新包组织。

## 项目边界

当前仓库拆成三个独立项目：

- `backend/`：Go Appliance 服务端，可独立运行、测试、构建，并随服务端交付 `cbmp-server-updater`。
- `frontend/`：Wails + React + HeroUI 客户端，可独立运行和构建，并随客户端交付 `cbmp-client-updater`。
- `industrial-control-gateway/`：通用工控对接服务，可独立接入 PLC/OPC/Modbus/边缘机数据并转发。

项目之间通过 HTTP/SSE/TCP/文件协议通信，源码、依赖、构建脚本、Go module / npm package 和容器镜像都保持独立。后端默认只提供 API，不挂载 `frontend/dist`；只有交付单体包需要时才显式设置 `CBMP_FRONTEND_DIR` 或 `--frontend`。

默认端口：

- Backend API: `127.0.0.1:8088`
- Frontend dev: 由 Wails/Vite 自动分配
- Industrial control gateway HTTP: `127.0.0.1:19101`
- Client updater: 随客户端部署，主动轮询平台，不固定监听端口
- Server updater: 随服务端部署，主动轮询平台，不固定监听端口

## 演示账号

| 角色 | 账号 | 密码 |
| --- | --- | --- |
| 产品运营管理员 | `admin` | `admin123` |

当前运营平台主线使用 `admin/admin123` 进行演示。客户现场探针、第三方监控、续费外部系统和端内 updater 使用各自配置的 token 接入。

## 核心演示流程

1. 启动后端：`./backend/scripts/dev.sh`。
2. 启动前端：`./frontend/scripts/dev.sh`。
3. 用 `admin/admin123` 登录，进入“产品运营台”。
4. 查看总览 KPI：客户实例、在线实例、异常实例、待续费授权、开放告警、严重告警、抑制告警、升级告警、续费任务、高危续费、客户端更新包、服务端更新包。
5. 在“客户实例”登记或编辑客户现场信息，包括授权 ID、水印、入口地址、客户端/服务端版本、授权到期日和续费阶段。
6. 在“交付初始化基础资料”维护客户、项目、产品、物料、站点、库存、司机、承运商和车辆；这些动作也可通过 `/api/master/customers`、`/api/master/projects`、`/api/master/products`、`/api/master/materials`、`/api/master/sites`、`/api/master/inventory`、`/api/master/drivers`、`/api/master/vehicles` 和 `/api/master/carriers` 批量或脚本化完成。
7. 在“客户现场探针上报”查看客户现场自动回传的版本、资源、队列和错误指标；也可用实例 token 测试调用探针上报接口。
8. 在“日志 / APM / 链路事件”查看客户现场日志、慢请求、5xx 和 trace 事件；也可用实例 token 测试上报，严重事件会自动进入系统异常报警。
9. 在“监控接入 / 告警规则”维护第三方监控 token 和产品告警规则，并用测试外部事件触发自动告警。
10. 在“告警聚合 / 抑制 / 升级”维护聚合窗口、抑制时长、升级对象和通知通道；在“通知通道”维护 SSE、本地、Webhook、企业微信、短信或 ITSM endpoint、token、secret 和超时秒数，并查看每次聚合、抑制、升级的投递状态、失败原因和重试结果。
11. 在“系统异常报警”创建客户现场异常，选择严重级别和来源，或查看探针/可观测/第三方监控事件自动生成的异常，可人工升级或处理闭环。
12. 在“授权续费任务”创建或更新续费跟进任务，填写负责人、阶段、金额、到期日和下次跟进时间。
13. 在“续费报价 / 合同 / 回款”生成续费报价、确认合同并登记回款；全额回款后任务会自动关闭，客户实例续费阶段会同步更新。
14. 在“续费审批 / 电子签 / 发票”提交续费审批、审批通过、发送并完成电子签，再按合同和回款生成续费发票。
15. 在“续费外部集成”维护电子签、财务回款和税控开票 endpoint。演示时可把 endpoint 改为 `mock://fail` 触发失败同步记录，再改回 `mock://success` 并在同步流水中点“重试”验证业务状态回写；真实第三方平台也可通过 `/api/product-ops/renewals/sync-callback` 回调同步签署、收款或开票结果。
16. 在“授权续费”签发或续期授权包，并在授权门户查看客户到期风险、模块覆盖和最近授权操作。
17. 在“客户端 / 服务端更新包”发布新包，可选择 `full` 或 `delta`；full 包填写 artifact 内容让平台自动生成 checksum，delta 包填写 `baseVersion`、`cbmp-copy-v1` patch、`baseArtifactSha256` 和 `targetArtifactSha256`，平台会校验 patch SHA256 并把目标 SHA 纳入签名；配置 `CBMP_UPDATE_SIGNING_PRIVATE_KEY` 后平台会生成 `ed25519` 非对称签名并展示公钥指纹，未配置时兼容生成 `hmac-sha256` 签名；分别查看客户端和服务端包的文件名、大小、包类型、base 版本和 SHA256，下载验签后的真实 artifact 或 patch，按版本应用更新。
18. 在“更新灰度批次”选择更新包和目标客户实例，创建批次后先做执行预检；内网可直连时用“执行一台”平台受控执行，客户端或服务端自身执行更新时用“端内更新”生成端内系统更新任务。
19. `cbmp-client-updater` 只处理 `component=client` 的任务，`cbmp-server-updater` 只处理 `component=server` 的任务；它们使用实例 token 调用 `/api/product-ops/system-updates/tasks` 轮询任务，按返回的下载地址、checksum 和 signature 获取更新包 envelope，并校验其中的 artifact 或 patch SHA256；如果包为 `ed25519` 签名，updater 会使用 envelope 中的公钥在本端离线验签。默认把下载写入 `rootDir/downloads/{taskNo}/package.json.part` 做断点续传，完成后把 envelope 保存为 `package.json`；full 包会把真实 artifact 还原到本端蓝绿发布目录，delta 包会按 `baseVersion` 查找当前/历史 release 的 base artifact，用 `cbmp-copy-v1` copy/data 指令重建目标 artifact，并用 `targetArtifactSha256` 校验后落盘。updater 会维护 `state.json/current.json`，执行安装或回滚。安装脚本可通过 `CBMP_ARTIFACT_PATH` 获取包体路径；安装或健康检查失败时默认自动回滚上一版本，再调用 `/api/product-ops/system-updates/tasks/{taskNo}/report` 回传进度和结果。
20. 在“端内系统更新任务”查看任务号、客户、组件版本、任务进度、更新包、token 提示、最近回执日志和最终结果；端内 updater 不在线时可由运营人员点“手动确认回执”作为人工兜底并留下回执审计。
21. 在“更新执行器”查看每次更新执行单的验签状态、执行人、客户、组件版本、dry-run 标记、端内更新进度、步骤明细和最终结果。
22. 在“API Gateway 配置中心”维护路由路径、稳定 upstream、灰度 upstream、灰度比例、读取超时和启停状态；可执行 25% 灰度、清零灰度、连接排空、停止排空，查看 reload 计划、Nginx 配置片段和网关事件。
23. 在“灾备与恢复演练”创建加密备份、查看备份文件和大小、运行恢复演练、查看演练检查项与对象计数；需要回滚平台数据时可选择备份执行恢复，操作会留下审计记录。

## 数据与授权

服务端数据默认保存到 `backend/data/app.vault`，采用 AES-GCM 加密。生产部署时请设置：

```bash
export CBMP_DATA_KEY="customer-strong-random-key"
```

设置 `CBMP_POSTGRES_DSN` 后，服务端可把加密快照写入 PostgreSQL，并通过领域行恢复数据。设置 `CBMP_BACKUP_DIR` 与 `CBMP_BACKUP_KEY` 后，可生成加密备份和恢复演练报告。

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
./backend/dist/cbmp-server-updater -print-service-files -binary /opt/cbmp/server-updater/cbmp-server-updater -config /etc/cbmp/server-updater.json
./frontend/build/bin/cbmp-client-updater -print-service-files -binary /opt/cbmp/client-updater/cbmp-client-updater -config /etc/cbmp/client-updater.json
```

根目录交付编排：

```bash
./scripts/build.sh
```

私有化部署见 [DEPLOYMENT.md](./DEPLOYMENT.md)，功能完成度见 [FUNCTION_AUDIT.md](./FUNCTION_AUDIT.md)。
