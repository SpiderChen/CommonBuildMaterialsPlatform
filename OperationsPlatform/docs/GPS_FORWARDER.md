# GPS 转发器

GPS 转发器已经融合进 `CommonBuildMaterialsOperationsPlatform`，作为运营平台项目内的独立后端命令交付。

源码位置：

```text
backend/cmd/gps-forwarder
backend/internal/gpsforwarder
```

它用来接收车载 GPS / 北斗终端、第三方 GPS 平台或串口落盘定位，完成解析、校验、去重和转发。

## 能力

- HTTP 接入：`POST /ingest`
- TCP 行协议接入：设置 `GPSF_TCP_ADDR`
- 文件轮询接入：设置 `GPSF_FILE`
- GPS JSON / CSV 解析
- 经纬度、速度、设备号校验
- 设备号 + 时间 + 经纬度去重
- HTTP 转发、JSONL 文件转发、stdout 兜底
- 默认 `protocol-frame` 转发模式，可直接对接主平台 GPS 协议帧接口
- HMAC 签名头：`X-CBMP-Timestamp`、`X-CBMP-Signature`
- 健康检查：`GET /healthz`
- 指标：`GET /metrics`

## 本地运行

```bash
./scripts/gps-forwarder-dev.sh
```

默认监听 `0.0.0.0:19102`。

## 常用环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `GPSF_ADDR` | `0.0.0.0:19102` | HTTP 服务地址 |
| `GPSF_TCP_ADDR` | 空 | TCP 行协议监听地址，例如 `0.0.0.0:19112` |
| `GPSF_FILE` | 空 | 轮询读取的 GPS 落盘文件 |
| `GPSF_OUT_FILE` | 空 | 标准化 JSONL 输出文件 |
| `GPSF_HTTP_TARGETS` | 空 | 逗号分隔 HTTP 转发地址 |
| `GPSF_HTTP_BEARER_TOKEN` | 空 | HTTP 转发 Bearer token |
| `GPSF_DEVICE_KEY` | 空 | 转发到私有化 ERP 设备接口时使用的 `X-Device-Key` |
| `GPSF_SHARED_SECRET` | 空 | HTTP 转发 HMAC 密钥 |
| `GPSF_PROTOCOL` | `gps-json` | 默认协议名 |
| `GPSF_SOURCE` | `gps-forwarder` | 默认来源 |
| `GPSF_FORWARD_MODE` | `protocol-frame` | `protocol-frame` 或 `location` |
| `GPSF_DEDUPE_WINDOW` | `10m` | 重复点过滤窗口 |
| `GPSF_INCLUDE_RAW` | `0` | 标准化输出是否包含原始帧 |

对接主平台 GPS 协议帧时可设置：

```bash
export GPSF_HTTP_TARGETS="http://127.0.0.1:8088/api/iot/protocols/gps/ingest"
export GPSF_FORWARD_MODE="protocol-frame"
export GPSF_PROTOCOL="gps-json"
export GPSF_DEVICE_KEY="device-demo-key-2"
```

## 示例

JSON:

```json
{
  "deviceNo": "GPS1000002",
  "plateNo": "粤B22336",
  "longitude": 113.9412,
  "latitude": 22.5428,
  "speed": 38.5,
  "direction": 96,
  "mileage": 2001.5,
  "accStatus": 1,
  "locationTime": "2026-06-18 12:05:00"
}
```

CSV:

```text
GPS,GPS1000002,粤B22336,113.9412,22.5428,38.5,96,2001.5,1,2026-06-18 12:05:00
```

## 验证与构建

```bash
./scripts/test.sh
./scripts/build.sh
```
