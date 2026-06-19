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

## 示例

JSON:

```json
{
  "deviceNo": "PLC-HZS180-01",
  "assetNo": "NS-HZS180",
  "eventType": "telemetry",
  "timestamp": "2026-06-18 10:20:00",
  "readings": {
    "cement": {"value": 8.32, "unit": "t"},
    "water": 1.2
  }
}
```

CSV:

```text
PLC,PLC-HZS180-01,NS-HZS180,telemetry,cement=8.32:t:good|water=1.2:t,2026-06-18 10:20:00
```

## 验证与构建

```bash
./scripts/test.sh
./scripts/build.sh
```
