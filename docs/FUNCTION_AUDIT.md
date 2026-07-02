# 功能完成度审计

审计时间：2026-06-21

结论：当前实现已按“客户侧 ERP + 独立 OperationsPlatform”重新收口，并已把 ERP 客户侧生产业务主线补到可运行闭环：销售合同、订单、生产、调度、地磅、签收、客户对账、应收、开票/红字、催收、收款，以及采购、入库、供应商对账、应付和付款均有前后端操作链路。它可以表述为“生产业务主线闭环已完备”，但仍不能表述为所有企业级边角能力全部完成；系统配置、自动化端到端测试、权限矩阵验收和部署运维加固仍需继续补深。

## 已修正边界

| 范围 | 状态 | 证据 |
| --- | --- | --- |
| ERP 前端默认入口 | 已修正 | `ERP/frontend/src/App.tsx` 默认进入“ERP 工作台”，不再导入或渲染 `ProductOpsView` |
| ERP 前端产品运营页面 | 已移除 | `ERP/frontend/src/views/ProductOpsView.tsx` 已删除 |
| ERP 前端 API 客户端 | 已修正 | `ERP/frontend/src/services/api.ts` 不再保留 `/product-ops/*` 调用方法 |
| ERP 前端类型 | 已修正 | `ERP/frontend/src/services/types.ts` 已移除运营台专用类型 |
| ERP API 身份 | 已修正 | `/api/` 返回 `common-build-materials-erp` |
| ERP 产品运营 API 分发 | 已修正 | `/api/product-ops/*` 在 ERP 服务中返回不可用 |
| ERP 后端历史产品运营实现 | 已隔离 | `ERP/backend/internal/appliance/product_*.go` 已加 `legacy_product_ops` build tag；默认 `go list ./internal/appliance` 将其列入 `IgnoredGoFiles`，普通 ERP 构建不再编译旧运营台处理器 |
| ERP 默认数据初始化 | 已修正 | 新建 vault/Postgres snapshot 默认只写入系统底座、数据字典和内置管理员；客户、订单、车辆、合同、生产、财务等演示业务仅在 `CBMP_SEED_DEMO=1` 或 `CBMP_ERP_SEED_DEMO=1` 时注入 |
| ERP 外部集成模拟入口 | 已修正 | 工作流 Webhook、税控、催收和续费外部集成不再接受 `mock://`/本地 simulator；催收发送会真实 POST 供应商 endpoint，回执由签名 callback 更新最终状态 |
| ERP 车辆位置模拟 API | 已移除 | `/api/simulate/tick` 不再暴露，客户侧车辆轨迹只能通过设备/转发器上报链路写入 |
| ERP 文档 | 已修正 | `ERP/frontend/README.md` 与 `ERP/backend/README.md` 明确 ERP 不承载产品运营台 |
| 仓库边界文档 | 已修正 | 根 README、交付说明、部署说明均明确 `OperationsPlatform/` 承接内部运营能力 |

## ERP 当前主线

| 范围 | 状态 | 说明 |
| --- | --- | --- |
| ERP 工作台 | 生产主线闭环增强 | 前端默认展示订单、生产、调度、过磅记录、签收、对账、堆场、财务、审批和经营指标总览，并将基础资料分散到销售、履约、堆场管理等业务模块；补充销售订单弹窗提交、合同生命周期、原料入场、客户/供应商对账、开票/红字、催收、收款/付款计划、过磅记录作废复核和审批处理入口 |
| 实验室管理 | 已实现首版 | 覆盖基础配比修订链、生产线配比、试配、审批生效/停用、生产试样、原料检验、样品试验复核、仪器校准约束、质量异常闭环、站点权限裁剪、审计日志、SSE 事件和前端工作台 |
| 基础资料 | 首版 CRUD | 客户/项目并入销售业务，产品/站点/司机/车辆/承运商并入履约中心，物料/库存并入堆场管理；筒仓并入生产线管理，堆场作为物理堆放区；旧账位仅保留为库存批次字段，不再提供仓库/旧筒仓前台资料维护接口 |
| 销售到收款 | 主线闭环已打通 | 覆盖合同创建/提审/修订/附件、订单、签收、客户对账、应收、收款计划、开票提交/浏览器真实下载、红字信息表/红冲、催收任务生成/真实供应商外发/回执/关闭和经营报表钻取 |
| 生产到发货 | 主线闭环已打通 | 覆盖生产计划、任务、批次、调度、地磅、过磅记录重打、过磅记录作废申请/复核、车辆轨迹和工地签收 |
| 采购到付款 | 主线闭环已打通 | 覆盖采购申请、采购订单、入库、库存流水、供应商对账、应付生成、应付余额和供应商付款登记 |
| 系统底座 | 首版可运行 | 覆盖授权校验、授权包导入页、插件、审批流、数据字典、备份恢复、API Gateway、SSE/Redis/RabbitMQ 事件；系统配置和运维类页面仍需分模块补深 |

## OperationsPlatform 当前主线

| 范围 | 状态 | 说明 |
| --- | --- | --- |
| 客户部署台账 | 已接通基础闭环 | 支持从空数据文件创建客户部署台账，管理客户名称、授权 ID、部署地址、联系人、客户端/服务端版本和健康状态，并可继续登记续费、分配更新包和写入审计 |
| 授权续费 | 已接通基础闭环 | 登记新到期日、版本、站点/车辆额度，使用固定 Ed25519 签发私钥生成客户侧 ERP `/api/system/license/import` 可导入的授权包；ERP 仅信任 `CBMP_LICENSE_TRUSTED_PUBLIC_KEYS` 中的签发公钥，并写入续费、下载和审计记录 |
| 系统异常告警 | 已接通基础闭环 | 支持运营端手工创建、客户侧 updater token 上报、同类未关闭告警去重续报，并按客户、来源、等级追踪打开、确认和关闭状态；客户侧上报不接受 `licenseId` 代替 token，body 中伪造的 customerId 会被 token 绑定客户覆盖 |
| 更新包管理 | 已接通基础执行闭环 | 区分客户端和服务端更新包，支持上传真实制品、计算 sha256、HMAC 签名、发布、分配到客户，并向现有客户端/服务端 updater 提供检查任务、下载真实制品和结果回传接口，回传会更新分配状态、客户版本和审计 |
| GPS 转发器 | 已并入 | 接收客户侧 GPS/北斗终端、第三方 GPS 平台或串口落盘定位，并转发到客户私有化 ERP |

## 下一步建议

1. 继续补齐 ERP 客户侧深业务页面，而不是把产品运营能力挂回 ERP。
2. 为系统配置和其它边角主数据继续补齐独立维护页，并继续梳理每个模块的完整 CRUD 对账表。
3. 为合同、开票/红字、催收、供应商对账/付款、过磅记录作废复核、报表钻取和系统配置补自动化前端闭环测试与权限矩阵验收。
4. OperationsPlatform 后续补深签发/更新密钥托管与轮换、异常告警策略、灰度策略和端内 updater 安装健康检查策略。

## 本轮验证

- `cd ERP/backend && go test ./...` 通过。
- `cd ERP/backend && go test -tags legacy_product_ops ./internal/appliance -run '^$'` 通过，确认 legacy 文件隔离后仍可编译检查。
- `cd ERP/backend && go list -f '{{range .GoFiles}}{{println .}}{{end}}' ./internal/appliance` 确认默认构建不包含 `product_*.go` 历史运营台处理器。
- `cd ERP/frontend && npm run build` 通过。
- `cd OperationsPlatform/backend && go test ./...` 通过。
- `node --input-type=module --check < OperationsPlatform/frontend/src/app.js` 通过。
- 关键字审计确认 OperationsPlatform 默认代码不再生成占位更新包，不再接受 `licenseId` 代替 updater token。
