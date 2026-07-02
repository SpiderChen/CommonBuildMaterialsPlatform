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

首次创建数据文件时默认写入空数据集；只有显式设置 `CBM_OPS_SEED_DEMO=1` 才会注入演示客户、告警和更新包，生产环境不要开启该变量。

授权续费会生成客户侧 ERP 可导入的签名授权包。生产环境必须配置 `CBM_OPS_LICENSE_ISSUER_PRIVATE_KEY` 为固定 Ed25519 私钥，并把对应公钥配置到客户侧 ERP 的 `CBMP_LICENSE_TRUSTED_PUBLIC_KEYS`；未配置私钥时续费授权包签发会失败。

更新包创建必须提交真实制品文件内容，运营平台会计算 `sha256` 并使用 `CBM_OPS_UPDATE_SIGNING_SECRET` 生成 HMAC 签名。未配置该密钥时，创建更新包会失败；updater 下载时返回的是原始制品内容，不再生成占位包。

客户侧 updater、告警上报和更新包下载必须使用客户部署台账中的 `updaterToken`。运营平台不会接受 `licenseId` 作为 updater token，也会忽略客户侧上报 body 中伪造的 `customerId`。

## 运行边界

- 不连接客户侧 ERP 数据库。
- 不读取客户服务器文件。
- 不执行远程系统命令。
- 仅通过后续约定的心跳、告警上报和更新检查 API 与客户侧 ERP 通信。
- GPS 转发器仅转发客户授权配置允许的定位数据入口，不作为 ERP 业务服务运行。
