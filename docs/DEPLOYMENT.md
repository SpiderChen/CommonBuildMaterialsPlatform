# 私有化部署

当前私有化部署对象是客户侧建材 ERP。厂商内部“产品运营台”不属于 ERP 部署入口，应部署和访问 `OperationsPlatform/`。

## ERP 交付组件

- `ERP/backend/Dockerfile`：客户侧 Go ERP 服务端镜像。
- `ERP/backend/dist/cbmp-server-updater`：随服务端项目构建的本端服务端 updater。
- `ERP/frontend/Dockerfile`：Wails/React Web 静态前端并封装到 Nginx。
- `ERP/frontend/build/bin/cbmp-client-updater`：随客户端项目构建的本端客户端 updater。
- `industrial-control-gateway/Dockerfile`：通用工控对接服务镜像，可按需部署。
- `deploy/nginx.conf`：ERP 前端静态入口和 API Gateway，代理 HTTP API 与 SSE 实时事件。
- `deploy/docker-compose.enterprise.yml`：ERP 私有化依赖编排，包含 PostgreSQL、Redis、RabbitMQ、MinIO、ClickHouse，并提供可选 `iot` profile 启动工控接入项目。

## ERP 运行时接入状态

- ERP 根接口 `/api/` 返回 `common-build-materials-erp` 产品身份。
- ERP 客户端默认进入客户业务工作台，不显示“产品运营台”。
- ERP 服务端对 `/api/product-ops/*` 返回不可用；客户实例、授权续费运营、客户现场探针、内部告警治理和更新包灰度编排归属 `OperationsPlatform/`。
- 实验室主接口为 `/api/laboratory/overview`，返回实验室 KPI、基础配比、生产线配比、试配记录、生产质检、原料质检、样品台账、试验记录、仪器校准、质量异常、可操作批次和入库单。
- 客户侧基础资料、合同、订单、生产、调度、地磅、签收、结算、采购、库存、财务、报表、插件、审批、字典和系统配置仍通过 ERP `/api/*` 接口提供。
- `industrial-control-gateway/` 是独立服务，可在需要设备接入或现场网关转发时单独部署；它不承载产品运营台。

## 本地私有化启动

```bash
cd deploy
CBMP_DATA_KEY="replace-with-customer-strong-key" docker compose -f docker-compose.enterprise.yml up --build
```

访问入口：

- ERP 入口：http://127.0.0.1:8080
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
4. Wails 桌面客户端仍由 `ERP/frontend/` 独立构建；容器入口使用前端镜像提供 Web/Gateway，后端镜像只提供 API。
5. 工控对接是独立服务，生产部署时可按项目现场网络拓扑单独放在设备网段、DMZ 或主 ERP 同网段。
6. 客户侧 updater 随 ERP 客户端/服务端部署；跨客户版本发布、灰度和更新任务运营由 `OperationsPlatform/` 管理。
