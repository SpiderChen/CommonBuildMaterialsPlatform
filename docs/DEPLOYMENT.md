# 私有化部署

当前交付部署的是产品运营交付平台，不是 SaaS 平台。部署对象是内部运营团队使用的一套独立系统，用来管理客户授权续费、系统异常报警、客户端更新包和服务端更新包。

上线后直接用运营账号维护客户产品实例、授权续费记录、告警接入和更新包。

## 交付组件

- `backend/Dockerfile`：Go Appliance 服务端镜像。
- `backend/dist/cbmp-server-updater`：随服务端项目构建的服务端 updater，用于执行 `component=server` 的端内系统更新任务。
- `frontend/Dockerfile`：Wails/React Web 静态前端并封装到 Nginx。
- `frontend/build/bin/cbmp-client-updater`：随客户端项目构建的客户端 updater，用于执行 `component=client` 的端内系统更新任务。
- `industrial-control-gateway/Dockerfile`：通用工控对接服务镜像，可按需部署。
- `deploy/nginx.conf`：前端静态入口和 API Gateway，代理 HTTP API 与 SSE 实时事件。
- `deploy/docker-compose.enterprise.yml`：私有化依赖编排，包含 PostgreSQL、Redis、RabbitMQ、MinIO、ClickHouse，并提供可选 `iot` profile 启动工控接入项目。

## 运行时接入状态

- 产品运营主接口为 `/api/product-ops/overview`，返回客户实例、授权续费任务、授权续费风险、系统异常报警、客户端/服务端更新包、授权门户、运行态和最近事件。
- 客户实例通过 `/api/product-ops/instances` 维护，通过 `/api/product-ops/instances/{id}/heartbeat` 接收现场心跳。
- 交付初始化和后续增量基础资料可在产品运营台“交付初始化基础资料”维护，也可通过 `/api/master/{resource}` 脚本化导入；当前覆盖客户、项目、产品、物料、司机、车辆、承运商、站点和库存；写入后会进入审计和实时事件。
- 客户现场探针通过 `/api/product-ops/probes/report` 上报，必须携带实例 `probeToken` 或请求头 `X-CBMP-Probe-Token`，平台会更新实例健康状态、保存探针报告，并按阈值自动生成或恢复系统告警。
- 客户现场日志、APM、trace 和 metric 事件通过 `/api/product-ops/telemetry/report` 上报，必须携带实例 `probeToken` 或请求头 `X-CBMP-Probe-Token`；平台会保存最近事件，并按慢请求、错误、4xx/5xx 自动生成 `source=telemetry` 的系统告警。
- 第三方监控接入通过 `/api/product-ops/monitoring/integrations` 和 `/api/product-ops/monitoring/rules` 维护，Prometheus、Grafana、Sentry、Zabbix 或自定义系统通过 `/api/product-ops/monitoring/report` 上报，必须携带接入 `token` 或请求头 `X-CBMP-Monitoring-Token`；平台会按指标规则自动生成 `source=monitoring` 的系统告警。
- 系统异常通过 `/api/product-ops/alerts` 创建，通过 `/api/product-ops/alerts/policies` 维护聚合窗口、抑制、升级和通知策略，通过 `/api/product-ops/alerts/channels` 维护 SSE、本地、Webhook、企业微信、短信或 ITSM 投递通道。Webhook、企业微信、短信和 ITSM 通道会按类型生成出站 JSON，支持 Bearer token、`X-CBMP-Timestamp`/`X-CBMP-Signature` HMAC 签名、通道超时和失败重试；通过 `/api/product-ops/alerts/notifications/{id}/retry` 重试失败通知，通过 `/api/product-ops/alerts/{id}/escalate` 人工升级，通过 `/api/product-ops/alerts/{id}/handle` 处理闭环；探针自动告警使用 `source=probe`，可观测事件自动告警使用 `source=telemetry`，第三方监控自动告警使用 `source=monitoring`。
- 续费任务通过 `/api/product-ops/renewals` 创建或更新，通过 `/api/product-ops/renewals/{id}/quote` 生成报价，通过 `/api/product-ops/renewals/{id}/approval` 提交/通过/驳回审批，通过 `/api/product-ops/renewals/{id}/contract` 确认合同，通过 `/api/product-ops/renewals/{id}/esign` 发送或完成电子签，通过 `/api/product-ops/renewals/{id}/payment` 登记回款，通过 `/api/product-ops/renewals/{id}/invoice` 生成续费发票；电子签、回款和发票动作会写入外部同步流水，外部集成通过 `/api/product-ops/renewals/integrations` 维护，失败流水可通过 `/api/product-ops/renewals/sync-records/{id}/retry` 人工重试并回写业务状态，第三方异步结果可通过 `/api/product-ops/renewals/sync-callback` 回调写入。全额回款会自动关闭任务，也可通过 `/api/product-ops/renewals/{id}/close` 手工关闭闭环。
- 授权包通过系统授权接口签发、续期、下载、导入、验签和吊销；运营台会按客户汇总到期风险。
- 更新包使用 `component=client` 或 `component=server` 区分客户端和服务端，通过 `/api/system/updates` 发布或更新；`packageType=full` 时真实包体以 `artifactContentBase64` 进入平台，服务端计算并校验 `sha256:<64位hex>`；`packageType=delta` 时 `artifactContentBase64` 是 `cbmp-copy-v1` patch，`checksum/artifactSha256` 校验 patch，`baseVersion/baseArtifactSha256/targetArtifactSha256` 校验差分来源和重建目标。生产建议设置 `CBMP_UPDATE_SIGNING_PRIVATE_KEY`，平台会生成 `ed25519:<base64url>` 非对称包签名、公钥和指纹，客户端/服务端 updater 可离线验签；未设置私钥时仍使用 `CBMP_UPDATE_SIGNING_SECRET` 生成 `hmac-sha256:<hex>` 兼容签名。普通列表只返回文件名、大小、公钥指纹、包类型、base 版本和摘要，下载接口才返回可还原的 artifact 包体或 patch，并记录下载次数、发布人和应用人。
- 更新灰度批次通过 `/api/product-ops/rollouts` 创建，通过 `/api/product-ops/rollouts/{id}/execute` 对客户实例执行预检、验签、受控更新、健康检查、回滚和步骤审计；通过 `/api/product-ops/rollouts/{id}/system-update-tasks` 可把批次目标下发给端内 updater。`cbmp-client-updater` 只处理 `component=client`，`cbmp-server-updater` 只处理 `component=server`；它们使用 `/api/product-ops/system-updates/tasks` 轮询任务，通过受控的 updater token 下载 `/api/system/updates/{id}/download`，默认使用 `rootDir/downloads/{taskNo}/package.json.part` 断点续传，完成后把 envelope 保存为 `package.json`；full 包会把真实 artifact 还原到蓝绿发布目录，delta 包会读取当前或历史 release 的 base artifact，按 `cbmp-copy-v1` copy/data 指令重建目标 artifact，并通过 `targetArtifactSha256` 校验后把 `CBMP_ARTIFACT_PATH` 交给安装脚本；`ed25519` 包会在本端使用公钥离线验签，安装或健康检查失败时自动回滚上一版本，并使用 `/api/product-ops/system-updates/tasks/{taskNo}/report` 回传执行进度、结果和当前版本；通过 `/api/product-ops/rollouts/{id}/advance` 保留人工推进、失败标记或回滚记录。执行成功会同步客户实例客户端或服务端版本，供运营台继续跟踪。
- 设置 `CBMP_POSTGRES_DSN` 后，服务端使用 PostgreSQL 保存 AES-GCM 加密快照 `cbmp_snapshot`，同时刷新领域行恢复表。
- 设置 `CBMP_POSTGRES_LOAD_FROM_DOMAIN=1` 后，服务端启动时优先从领域行恢复数据；若快照缺失，也会自动尝试恢复后重新写入快照。
- 设置 `CBMP_REDIS_ADDR` 后，实时事件会写入 Redis Stream `cbmp:events`。
- 设置 `CBMP_RABBITMQ_URL` 后，业务变更会发布到 topic exchange `cbmp.events`。
- 设置 `CBMP_BACKUP_DIR` 和 `CBMP_BACKUP_KEY` 后，系统备份 API 会在指定目录创建 AES-GCM 加密灾备快照。
- 未设置 Redis/RabbitMQ 时，实时事件仍通过本地 SSE 工作。
- `industrial-control-gateway/` 是独立服务，可在需要设备接入或现场网关转发时单独部署；它不影响产品运营台的授权、告警和更新包主流程。

## 本地私有化启动

```bash
cd deploy
CBMP_DATA_KEY="replace-with-customer-strong-key" docker compose -f docker-compose.enterprise.yml up --build
```

访问入口：

- 平台入口：http://127.0.0.1:8080
- RabbitMQ 管理台：http://127.0.0.1:15672
- MinIO 控制台：http://127.0.0.1:9001

如需同时启动通用工控接入项目：

```bash
cd deploy
CBMP_DATA_KEY="replace-with-customer-strong-key" docker compose --profile iot -f docker-compose.enterprise.yml up --build
```

可选接入口：

- 通用工控 HTTP：http://127.0.0.1:19101
- 通用工控 TCP：127.0.0.1:19111

## 生产部署要求

1. 必须设置强随机 `CBMP_DATA_KEY`。
2. 必须替换 PostgreSQL、RabbitMQ、MinIO 默认密码。
3. 建议设置 `CBMP_BACKUP_DIR`、`CBMP_BACKUP_KEY` 并定期执行恢复演练。
4. 生产环境建议设置 Ed25519 更新包私钥 `CBMP_UPDATE_SIGNING_PRIVATE_KEY`，格式为 `ed25519:<base64url private key or seed>`；仅在兼容旧包或低安全演示时使用强随机 `CBMP_UPDATE_SIGNING_SECRET` 生成 HMAC 签名。替换签名密钥前应先完成旧更新包下线或重签。
5. Wails 桌面客户端仍由 `frontend/` 独立构建；容器入口使用前端镜像提供 Web/Gateway，后端镜像只提供 API。
6. 工控对接是独立服务，生产部署时可按项目现场网络拓扑单独放在设备网段、DMZ 或主平台同网段。
7. 更新包、授权包和备份目录应纳入客户机房的备份策略，并限制为内部运营人员访问。
8. 告警通知通道生产部署时应把 `mock://success` 示例替换为真实 Webhook、企业微信、短信或 ITSM endpoint，并配置 token、secret、超时和重试次数；未配置 endpoint 的外部通道会被标记为失败并进入可重试审计。若外部系统验签，应按 `X-CBMP-Timestamp + "." + body` 使用 HMAC-SHA256 校验 `X-CBMP-Signature`。
9. 续费外部集成生产部署时应把 `local_esign`、`finance_bank`、`tax_gateway` 的 `mock://success` endpoint 替换为客户实际电子签、财务或税控平台的 HTTPS endpoint，并配置 token/secret、超时和重试次数；平台会保存每次请求/响应摘要、外部流水号、失败原因和重试结果。若第三方平台异步通知签署、收款或开票结果，应回调 `/api/product-ops/renewals/sync-callback`，配置了 `secret` 时必须带 `X-CBMP-Timestamp` 与 `X-CBMP-Signature`。
10. 客户端 updater 应随客户端部署，服务端 updater 应随服务端部署，二者使用对应客户实例的 token 轮询任务、下载已验签更新包、写入蓝绿发布目录、执行可配置停服/安装/启动/健康检查命令，并把进度/结果回传给平台；生产建议保持 `resumeDownloads=true`、`strictChecksum=true` 和 `autoRollbackOnFailure=true`，保留 `rootDir/downloads` 缓存用于断点续传和审计。安装脚本通过 `CBMP_RELEASE_DIR` 获取 release 目录，通过 `CBMP_ARTIFACT_PATH` 获取真实 artifact 文件路径。可用 `./backend/dist/cbmp-server-updater -print-service-files -binary /opt/cbmp/server-updater/cbmp-server-updater -config /etc/cbmp/server-updater.json` 输出服务端 updater 服务模板；可用 `./frontend/build/bin/cbmp-client-updater -print-service-files -binary /opt/cbmp/client-updater/cbmp-client-updater -config /etc/cbmp/client-updater.json` 输出客户端 updater 服务模板；平台侧只负责授权、任务、包、审计和版本同步。
