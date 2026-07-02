# 通用工控对接项目

`industrial-control-gateway/` 是独立 Go 项目，用来把 PLC、OPC 边缘机、Modbus 网关、拌合楼采集程序或串口落盘文件统一标准化，再转发给主平台或客户自己的 MES/SCADA 中台。

## 能力

- HTTP 接入：`POST /ingest`
- TCP 行协议接入：设置 `ICG_TCP_ADDR`
- 文件轮询接入：设置 `ICG_FILE`
- JSON / CSV 工控帧标准化
- HTTP 转发、JSONL 文件转发、stdout 兜底
- `protocol-frame` 转发模式，可直接对接主平台协议帧接口
- HMAC 签名头：`X-CBMP-Timestamp`、`X-CBMP-Signature`
- 健康检查：`GET /healthz`
- 指标：`GET /metrics`

## 本地运行

```bash
./scripts/dev.sh
```

默认监听 `0.0.0.0:19101`。

## 常用环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `ICG_ADDR` | `0.0.0.0:19101` | HTTP 服务地址 |
| `ICG_TCP_ADDR` | 空 | TCP 行协议监听地址，例如 `0.0.0.0:19111` |
| `ICG_FILE` | 空 | 轮询读取的工控落盘文件 |
| `ICG_OUT_FILE` | 空 | 标准化 JSONL 输出文件 |
| `ICG_HTTP_TARGETS` | 空 | 逗号分隔 HTTP 转发地址 |
| `ICG_HTTP_BEARER_TOKEN` | 空 | HTTP 转发 Bearer token |
| `ICG_SHARED_SECRET` | 空 | HTTP 转发 HMAC 密钥 |
| `ICG_PROTOCOL` | `industrial-json` | 默认协议名 |
| `ICG_SOURCE` | `industrial-control-gateway` | 默认来源 |
| `ICG_FORWARD_MODE` | `normalized` | `normalized` 或 `protocol-frame` |
| `ICG_INCLUDE_RAW` | `0` | 标准化输出是否包含原始帧 |

对接主平台生产协议帧时可设置：

```bash
export ICG_HTTP_TARGETS="http://127.0.0.1:8088/api/production-plans/protocols/plant/ingest"
export ICG_FORWARD_MODE="protocol-frame"
export ICG_PROTOCOL="plant-json"
export ICG_HTTP_BEARER_TOKEN="设备或用户令牌"
```

ERP 生产线不需要手工填写接口地址；生产线档案的 `code` 需要和网关上报的 `plantCode` 一致。`protocol-frame` 模式会保留 `raw`，并在识别到 plant 批次帧时补充 ERP 可消费的 `payload`。

对接生产线暂存仓位、骨料仓料位时，把目标地址切到：

```bash
export ICG_HTTP_TARGETS="http://127.0.0.1:8088/api/production-plans/protocols/buffer/ingest"
export ICG_FORWARD_MODE="protocol-frame"
export ICG_PROTOCOL="buffer-json"
export ICG_HTTP_BEARER_TOKEN="设备或用户令牌"
```

仓位上报的 `bufferCode` 需要和 ERP 生产线管理里的暂存仓位编码一致。

对接站点堆场、堆位料位时，把目标地址切到：

```bash
export ICG_HTTP_TARGETS="http://127.0.0.1:8088/api/production-plans/protocols/yard/ingest"
export ICG_FORWARD_MODE="protocol-frame"
export ICG_PROTOCOL="yard-json"
export ICG_HTTP_BEARER_TOKEN="设备或用户令牌"
```

堆场上报的 `yardCode` 和 `pileCode` 需要分别和 ERP 采购库存里的堆场编码、堆位编码一致。`yard-json`、`yard_level`、`stockpile_level`、`pile_level` 都会被转换为 ERP 堆位料位 payload。

## 示例

生产批次 JSON:

```json
{
  "deviceNo": "PLANT-NS-AMP240",
  "plantCode": "NS-AMP240",
  "taskId": 20,
  "batchNo": "PLC-BATCH-001",
  "quantity": 6,
  "operator": "PLC-A",
  "qualityStatus": "pending",
  "status": "released",
  "completedAt": "2026-06-20 09:20:00",
  "eventType": "batch"
}
```

生产批次 CSV:

```text
PLANT,PLANT-NS-AMP240,NS-AMP240,batch,taskId=20|quantity=6,2026-06-20 09:20:00
```

骨料仓料位 JSON:

```json
{
  "deviceNo": "PLANT-NS-AMP240",
  "plantCode": "NS-AMP240",
  "bufferCode": "NS-AMP240-AGG-01",
  "materialId": 3,
  "quantity": 32.5,
  "moistureRate": 4.2,
  "qualityStatus": "passed",
  "status": "active",
  "eventType": "buffer_level",
  "reportedAt": "2026-06-21 22:50:00"
}
```

堆场堆位料位 JSON:

```json
{
  "deviceNo": "YARD-NS-AGG",
  "yardCode": "NS-YARD-AGG",
  "pileCode": "NS-YARD-SAND-01",
  "materialId": 3,
  "quantity": 538.5,
  "moistureRate": 4.7,
  "qualityStatus": "passed",
  "status": "active",
  "eventType": "yard_level",
  "reportedAt": "2026-06-21 23:10:00"
}
```

## 验证与构建

```bash
./scripts/test.sh
./scripts/build.sh
```
