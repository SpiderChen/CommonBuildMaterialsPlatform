# 功能完成度审计

审计时间：2026-06-19

结论：当前实现已按“单独产品运营交付平台 + ERP 实验室管理”重塑主线。它不是 SaaS 系统；当前可交付主流程包含产品运营链路“客户实例登记 -> 客户现场探针上报 -> 日志/APM/链路事件上报 -> 第三方监控事件接入 -> 可配置规则自动异常告警/恢复 -> 告警聚合/抑制/升级/通知审计 -> 授权续费任务跟踪 -> 续费报价 -> 续费审批 -> 续费电子签 -> 续费合同 -> 续费回款 -> 续费开票 -> 续费外部系统同步/失败重试 -> 授权签发/续期 -> 客户端/服务端更新包发布 -> 更新灰度批次预检/平台执行/端内 updater 执行/回滚 -> 执行步骤审计 -> 客户实例版本同步 -> 加密备份与私有化运行”，以及实验室链路“配比建档 -> 试配记录 -> 配比审批生效/停用 -> 生产任务默认取当前生效配比 -> 批次保留配比版本 -> 生产试块/原料检验 -> 样品试验复核 -> 仪器校准约束 -> 不合格自动异常 -> 原因和纠正措施闭环”。

## 已验证可用

| 范围 | 证据 |
| --- | --- |
| 独立项目边界 | `backend/`、`frontend/` 与 `industrial-control-gateway/` 各自有依赖、脚本、README 和构建产物；服务端 updater 随 `backend/` 交付，客户端 updater 随 `frontend/` 交付；根目录只做交付编排 |
| Go 服务端 | `backend/cmd/appliance` 与 `backend/cmd/cbmp-server-updater` 可构建，`./backend/scripts/test.sh` 通过 |
| Wails + React + HeroUI 客户端 | Web 构建通过，Wails 桌面构建脚本保留在 `frontend/scripts/build.sh`，客户端 updater 位于 `frontend/cmd/cbmp-client-updater` |
| 产品运营台 | 前端默认进入 `ProductOpsView`，围绕客户实例、交付初始化基础资料、授权续费、系统异常报警、客户端/服务端更新包组织 |
| 产品运营 API 接入 | `/api/` 返回 `product-ops-appliance` 产品身份；CORS 已允许 `X-CBMP-Probe-Token`、`X-CBMP-Monitoring-Token`、`X-CBMP-Updater-Token` 和签名头，便于客户现场探针、监控系统和端内 updater 在私有化网络中接入 |
| 产品运营总览 | `/api/product-ops/overview` 返回 KPI、客户实例、续费任务、报警、更新包、授权门户、运行态和最近事件 |
| 客户实例管理 | `/api/product-ops/instances` 可新增或更新客户实例，字段覆盖客户名称、授权 ID、水印、部署入口、客户端/服务端版本、授权到期日、续费负责人和续费阶段 |
| 客户现场心跳 | `/api/product-ops/instances/{id}/heartbeat` 可更新客户实例在线状态和最近心跳时间 |
| 基础资料增量维护 | 产品运营台已有“交付初始化基础资料”页面区块，`/api/master/{resource}` 已覆盖客户、项目、产品、物料、司机、车辆、承运商、站点和库存新增闭环；新增记录会写入加密数据仓、审计日志和 SSE 事件 |
| 实验室管理 | 新增 `/api/laboratory/overview` 和前端“实验室管理”导航，覆盖实验室 KPI、配比台账、试配记录、生产试块、原料检验、样品台账、试验复核、仪器校准和质量异常闭环 |
| 配比管理 | `/api/laboratory/mix-designs` 支持新增、修订、审批生效、停用和试配记录；同一产品同一站点只保留一个当前生效配比，生产任务默认选用当前生效版本，历史生产批次保留原 `mixDesignId` |
| 实验室试验与异常 | `/api/laboratory/samples`、`/api/laboratory/samples/{id}/tests`、`/api/laboratory/tests/{id}/review` 支持样品登记、试验和复核；不合格复核自动创建质量异常，`/api/laboratory/exceptions/{id}/handle` 记录原因和纠正措施并关闭 |
| 仪器校准 | `/api/laboratory/equipment` 与 `/api/laboratory/equipment/{id}/calibrations` 支持仪器登记和校准；过期或禁用仪器会阻止新试验，临期/过期进入实验室 KPI |
| 客户现场探针 | `/api/product-ops/probes/report` 支持 token 上报客户现场版本、CPU、内存、磁盘、队列积压、错误数和状态；会更新实例健康状态、保存报告并自动维护探针告警 |
| 日志/APM/链路事件 | `/api/product-ops/telemetry/report` 支持 token 上报日志、APM、trace 和 metric 事件；会保存事件、识别慢请求/错误/4xx/5xx 并自动生成 `source=telemetry` 告警 |
| 第三方监控接入 | `/api/product-ops/monitoring/integrations` 可维护 Prometheus/Grafana/Sentry/Zabbix/custom 接入，`/api/product-ops/monitoring/rules` 可维护指标规则，`/api/product-ops/monitoring/report` 使用 token 上报事件并按规则自动生成 `source=monitoring` 告警 |
| 系统异常报警 | `/api/product-ops/alerts` 可创建客户现场异常，`/api/product-ops/alerts/policies` 可维护聚合/抑制/升级/通知策略，`/api/product-ops/alerts/channels` 可维护 SSE/local/webhook/enterprise_wechat/sms/itsm 通知通道；Webhook、企业微信、短信和 ITSM 会生成类型化出站 JSON，支持 Bearer token、HMAC 签名、通道超时、失败重试和通知审计；`/api/product-ops/alerts/notifications/{id}/retry` 可重试失败通知，`/api/product-ops/alerts/{id}/escalate` 可人工升级，`/api/product-ops/alerts/{id}/handle` 可处理闭环；通知投递审计已进入产品运营总览 |
| 授权续费任务 | `/api/product-ops/renewals` 可创建或更新续费任务，`/api/product-ops/renewals/{id}/close` 可关闭任务并记录完成时间 |
| 续费报价/合同/回款/开票 | `/api/product-ops/renewals/{id}/quote` 可生成报价，`/api/product-ops/renewals/{id}/approval` 可提交/通过/驳回审批，`/api/product-ops/renewals/{id}/contract` 可确认合同，`/api/product-ops/renewals/{id}/esign` 可发送/完成电子签，`/api/product-ops/renewals/{id}/payment` 可登记部分或全额回款，`/api/product-ops/renewals/{id}/invoice` 可生成续费发票并同步任务阶段；`/api/product-ops/renewals/integrations` 可维护电子签、财务和税控集成，`/api/product-ops/renewals/sync-records/{id}/retry` 可重试失败同步流水并回写电子签/发票状态，`/api/product-ops/renewals/sync-callback` 可接收第三方异步回调并做 HMAC 验签 |
| 授权签发与验签 | 授权中心可签发 Ed25519 离线 license 包、下载导出、导入验签、校验到期日和站点/车辆额度 |
| 授权续期与吊销 | 已签发授权包可在线续期生成新包，吊销列表会让已吊销 license 被验签链路拒绝 |
| 总部授权门户 | `/api/system/license/portal` 可按客户聚合授权包、到期风险、模块覆盖率、下载轨迹和最近授权操作 |
| 客户端/服务端更新包 | 更新包模型包含 `component`、`packageType`、artifact 元信息、base 版本和目标 SHA 字段，运营台按客户端和服务端分别展示，可发布真实 full 包体或 `cbmp-copy-v1` 差分 patch、下载真实 artifact/patch、应用和跟踪状态 |
| 更新中心下载与发布 | 系统更新中心可发布更新包、列出脱敏元信息、下载导出验签后的真实 artifact 包体、应用版本和回滚版本状态，并记录下载次数、发布人、应用人和审计 |
| 更新灰度批次 | `/api/product-ops/rollouts` 可基于更新包创建客户实例批次，`/api/product-ops/rollouts/{id}/advance` 可推进、失败标记或回滚，并同步客户实例版本 |
| 更新执行器 | `/api/product-ops/rollouts/{id}/execute` 可对批次客户实例做 dry-run 预检、更新包验签、现场快照、分发、停止组件、安装、健康检查、发布结果和回滚记录，并写入执行步骤审计 |
| 端内系统更新 | `/api/product-ops/rollouts/{id}/system-update-tasks` 可下发端内更新任务，`/api/product-ops/system-updates/tasks` 可用实例 token 轮询任务，updater token 可受控下载对应更新包 envelope；`frontend/cmd/cbmp-client-updater` 只执行 `client` 组件更新，`backend/cmd/cbmp-server-updater` 只执行 `server` 组件更新，均可断点续传下载、校验 artifact/patch SHA256、把 full 包体还原到蓝绿发布目录，或按 `baseVersion` 和 `cbmp-copy-v1` copy/data 指令从上一 release 重建目标 artifact 并校验 `targetArtifactSha256`，维护本地状态、向安装脚本提供 `CBMP_ARTIFACT_PATH`、执行可选命令、失败自动回滚，并通过 `/api/product-ops/system-updates/tasks/{taskNo}/report` 回传进度/结果；CLI 可输出 Linux systemd、launchd 和 Windows 服务模板 |
| 加密本地数据仓 | `backend/internal/appliance/store.go` 使用 AES-GCM 保存 `backend/data/app.vault` |
| PostgreSQL 加密快照 | 设置 `CBMP_POSTGRES_DSN` 后可保存加密快照并恢复领域行数据 |
| 灾备恢复 | 系统备份创建、列表、恢复 API 和恢复演练报告已接入，运营台可创建加密备份、查看备份文件、执行恢复和运行恢复演练；备份文件使用 AES-GCM 加密 |
| API Gateway 配置中心 | 系统接口和运营台均可维护路由、灰度 upstream/比例和连接排空窗口，可生成 Nginx 配置片段、重载计划、排空探针检查和网关事件审计 |
| SSE 实时事件 | `/api/events` 可推送客户实例、异常报警、更新和授权相关事件 |
| 私有化部署骨架 | Nginx Gateway、PostgreSQL、Redis、RabbitMQ、MinIO、ClickHouse compose 文件已提供 |

## 当前产品边界

| 项目 | 状态 | 说明 |
| --- | --- | --- |
| 单体运营平台形态 | 已实现 | 默认数据、权限裁剪、SSO/SCIM 自动开通和前端页面都围绕客户产品实例；历史租户字段只做旧数据读取兼容和响应脱敏，不参与业务运行 |
| 客户实例 | 已实现首版 | 作为运营台主对象，承载授权、版本、入口、心跳和续费风险 |
| 实验室管理 | 已实现首版 | 已覆盖配比版本链、试配、审批生效/停用、生产试块、原料检验、样品试验复核、仪器校准约束、质量异常闭环、站点权限裁剪、审计日志、SSE 事件和前端工作台 |
| 授权续费 | 已实现首版 | 可签发、下载、续期、导入验签、吊销，并在门户聚合风险；续费任务已可跟踪负责人、阶段、金额、到期日、下次跟进、报价、审批、电子签、合同、回款、开票、外部集成同步、失败重试和关闭状态 |
| 系统异常报警 | 已实现多通道首版 | 可创建、关联实例、按严重级别展示和处理闭环；探针、日志/APM/trace 事件和第三方监控事件可自动生成告警，产品告警规则可按来源、组件、指标、操作符和阈值配置；告警策略已支持聚合窗口、重复抑制、自动升级、人工升级、SSE/Webhook/企业微信/短信/ITSM 通道、Bearer token、HMAC 签名、通道超时、投递失败、人工重试和通知审计 |
| 客户端更新包 | 已实现首版 | 已与服务端包区分，支持发布真实 artifact、下载、SHA256 校验、Ed25519 非对称签名验签、HMAC 兼容签名、应用动作、下载次数记录和客户实例灰度批次 |
| 服务端更新包 | 已实现首版 | 已与客户端包区分，支持发布真实 artifact、下载、SHA256 校验、Ed25519 非对称签名验签、HMAC 兼容签名、应用、回滚状态、应用人记录和客户实例灰度批次 |
| 真实二进制热更新 | 已实现端内 updater 可交付首版 | 已有真实 full artifact 包体、`cbmp-copy-v1` 二进制差分 patch、服务端 SHA256 校验、目标包 SHA256 校验、`ed25519` 非对称包签名、`hmac-sha256` 兼容签名、端内 updater 离线验签、应用/回滚状态、灰度批次、dry-run 预检、平台受控执行、端内更新任务下发/轮询/受控下载/回执、`.part` 断点续传缓存、蓝绿发布目录、本地 `state.json/current.json`、`CBMP_ARTIFACT_PATH` 安装脚本变量、失败自动回滚、Linux/macOS/Windows 服务模板、执行步骤审计和客户实例版本同步；后续可继续深化 OS 原生安装器和本端服务编排适配 |
| 自动化续费工单 | 已实现首版 | 当前可维护续费任务、负责人、阶段、金额、到期日、下次跟进、报价、审批、电子签、合同、部分回款、全额回款自动关闭、开票、电子签/财务/税控外部集成配置、同步流水、失败重试、HMAC 回调验签和客户实例阶段同步；后续可继续扩展具体厂商协议、回调字段映射和银行/税控专有字段 |
| 外部告警接入 | 已实现首版 | 当前已有客户现场探针 token 上报、日志/APM/trace/metric 事件上报、第三方监控 token 接入、指标规则配置、慢请求/错误/5xx/自定义指标自动告警和恢复闭环；重复事件会进入告警聚合、抑制、升级和通知投递策略 |

## 下一步建议

1. 继续围绕产品运营台补深客户授权续费、异常报警、客户端/服务端更新包和端内 updater 执行链路。
2. 优先补深客户端 updater 和服务端 updater 的 OS 原生安装器、本端服务编排适配；继续扩展具体电子签/财务/税控厂商适配器、字段映射、电话值班联动和各厂商专有字段。
3. 保持 `backend/`、`frontend/` 和 `industrial-control-gateway/` 三个独立项目，端内更新能力分别归属服务端与客户端项目，便于后续按端演进和独立交付。
