# Frontend: Wails + React + HeroUI Product Ops Client

`frontend/` 是独立客户端项目，负责产品运营交付平台的桌面端和 Web 端界面。

当前默认进入“产品运营台”，并提供“实验室管理”导航；产品运营台围绕客户产品实例、授权续费、系统异常报警、客户端更新包、服务端更新包和端内系统更新任务组织页面，实验室管理围绕配比、试配、质检、仪器和异常闭环组织页面。

## 核心界面

- 产品运营总览：客户实例、待续费授权、开放告警、严重告警、续费任务、更新包、通知审计、端内系统更新任务等 KPI。
- 实验室管理：配比台账、版本审批、试配记录、生产试块、原料检验、样品台账、试验复核、仪器校准、质量异常闭环和质量 KPI。
- 交付初始化基础资料：维护客户、项目、产品、物料、站点、库存、司机、承运商和车辆。
- 客户实例：维护授权 ID、水印、部署入口、客户端/服务端版本、授权到期日、续费负责人和续费阶段。
- 系统异常报警：客户现场探针、日志/APM/trace、第三方监控事件、聚合抑制、升级、通知通道和投递重试。
- 授权续费：续费任务、报价、审批、合同、电子签、回款、开票、外部集成同步和失败重试。
- 更新包管理：客户端/服务端 full artifact 和 `cbmp-copy-v1` 差分 patch 发布、patch/target SHA256、Ed25519 公钥指纹、HMAC 兼容签名元信息、真实包体或 patch 下载、灰度批次、平台受控执行、端内 updater 下发、回执和手动确认兜底。
- Gateway 运维：维护 API Gateway 路由、灰度 upstream、连接排空、reload 计划、Nginx 片段和网关事件。
- 灾备运维：在运营台创建加密备份、查看备份文件、执行恢复和运行恢复演练。

## API 连接

Web 私有化部署优先使用当前域名下的 `/api`；Wails 壳回落到 `http://127.0.0.1:8088/api`。

可通过环境变量覆盖：

```bash
VITE_API_BASE_URL=http://127.0.0.1:8088/api ./frontend/scripts/dev.sh
```

## 运行

先启动后端：

```bash
./backend/scripts/dev.sh
```

再启动客户端：

```bash
./frontend/scripts/dev.sh
```

## Web 构建

```bash
./frontend/scripts/web-build.sh
```

## 桌面构建

```bash
./frontend/scripts/build.sh
```

## 客户端 Updater

```bash
./frontend/scripts/updater-build.sh
./frontend/build/bin/cbmp-client-updater -print-sample-config
```

`cbmp-client-updater` 随客户端交付，只处理 `component=client` 的端内更新任务。

## 渐进式模块开发

产品运营页面位于 `frontend/src/views/ProductOpsView.tsx`，实验室页面位于 `frontend/src/views/LaboratoryView.tsx`，共享组件放在 `frontend/src/components/`，API 调用集中在 `frontend/src/services/`。前端不导入后端源码，只消费接口契约。
