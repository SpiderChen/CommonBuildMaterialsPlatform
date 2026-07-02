# Frontend: Common Build Materials ERP Client

`frontend/` 是客户侧建材 ERP 的 Wails + React + HeroUI 客户端项目。

ERP 客户端面向最终客户的日常业务操作，不承载厂商内部“产品运营台”。授权续费、客户实例台账、客户现场告警、客户端/服务端更新包和灰度发布等内部运营能力归属 `../../OperationsPlatform/`。

## 核心界面

- ERP 工作台：订单、生产、调度、地磅、签收、对账、采购、财务、审批和经营指标总览。
- 基础资料：集团组织、客户/项目归入销售，产品/站点/司机/车辆归入履约，物料/库存归入采购库存；各台账页使用通用分页表格，支持查询、弹窗新增、弹窗编辑和删除。
- 实验室管理：基础配比、生产线配比、版本审批、试配记录、生产试样、原料检验、样品台账、试验复核、仪器校准、质量异常闭环和质量 KPI。
- 公共签收：支持工地签收公开链接和签收附件。
- 系统能力：通过共享 API 客户端访问 ERP 的基础资料、生产、调度、地磅、签收、结算、采购、库存、财务、报表、插件和系统配置接口。

## API 连接

Web 私有化部署优先使用当前域名下的 `/api`；Wails 壳回落到 `http://127.0.0.1:8088/api`。

登录页右上角和已登录后的右上角设置都提供“连接设置”：选择“本地模式”时使用上述默认本机 API，选择“服务端模式”时可填写局域网、服务器或云端 ERP 地址。

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

`cbmp-client-updater` 随客户侧客户端交付，只负责本端客户端更新执行；更新任务的运营编排应由 `OperationsPlatform` 承担。

## 渐进式模块开发

ERP 工作台位于 `frontend/src/views/ERPWorkbenchView.tsx`，实验室页面位于 `frontend/src/views/LaboratoryView.tsx`，公共签收页面位于 `frontend/src/views/PublicSignView.tsx`，共享组件放在 `frontend/src/components/`，API 调用集中在 `frontend/src/services/`。前端不导入后端源码，只消费接口契约。
