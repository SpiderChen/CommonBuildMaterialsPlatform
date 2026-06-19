# 部署说明

## 单机部署

```bash
./scripts/build.sh
CBM_OPS_ADDR=:8095 \
CBM_OPS_DATA=/var/lib/cbm-ops/ops.json \
CBM_OPS_FRONTEND_DIR=./frontend \
./backend/dist/cbm-ops
```

GPS 转发器：

```bash
./scripts/build.sh
GPSF_ADDR=0.0.0.0:19102 \
GPSF_HTTP_TARGETS="https://customer-erp.example.com/api/iot/protocols/gps/ingest" \
GPSF_DEVICE_KEY="customer-device-key" \
GPSF_FORWARD_MODE=protocol-frame \
./backend/dist/gps-forwarder
```

## Docker Compose

```bash
docker compose -f deploy/docker-compose.yml up --build
```

默认端口：

```text
8095  运营平台
19102 GPS 转发器
```

## 数据目录

生产环境建议挂载：

```text
/var/lib/cbm-ops
```

其中 `ops.json` 保存客户部署、续费、告警、更新包和审计日志。

## 运行边界

- 不连接客户侧 ERP 数据库。
- 不读取客户服务器文件。
- 不执行远程系统命令。
- 仅通过后续约定的心跳、告警上报和更新检查 API 与客户侧 ERP 通信。
- GPS 转发器仅转发客户授权配置允许的定位数据入口，不作为 ERP 业务服务运行。
