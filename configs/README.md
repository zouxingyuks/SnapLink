rules.json 语法

```json
{
  // ID表示规则的唯一ID（可选）。
  "id": "string",
  // Resource表示资源名称。
  "resource": "string",
  "tokenCalculateStrategy": "TokenCalculateStrategy",
  "controlBehavior": "ControlBehavior",
  // Threshold表示StatIntervalInMs期间的阈值
  // 如果StatIntervalInMs是1000（1秒），Threshold表示QPS
  "threshold": "float64",
  "relationStrategy": "RelationStrategy",
  "refResource": "string",
  // MaxQueueingTimeMs仅在ControlBehavior为Throttling时生效。
  // 当MaxQueueingTimeMs为0时，意味着Throttling只控制请求间隔，
  // 超过阈值的请求将直接被拒绝。
  "maxQueueingTimeMs": "uint32",
  "warmUpPeriodSec": "uint32",
  "warmUpColdFactor": "uint32",
  // StatIntervalInMs指示统计间隔，它是流规则的可选设置。
  // 如果用户没有设置StatIntervalInMs，意味着使用资源的默认度量统计。
  // 如果用户指定的StatIntervalInMs无法重用资源的全局统计，
  // 则sentinel将为此规则生成独立的统计结构。
  "statIntervalInMs": "uint32",

  // 自适应流量控制算法相关参数
  // 限制：LowMemUsageThreshold > HighMemUsageThreshold && MemHighWaterMarkBytes > MemLowWaterMarkBytes
  // 如果当前内存使用量小于或等于MemLowWaterMarkBytes，则阈值为LowMemUsageThreshold
  // 如果当前内存使用量大于或等于MemHighWaterMarkBytes，则阈值为HighMemUsageThreshold
  // 如果当前内存使用量在(MemLowWaterMarkBytes, MemHighWaterMarkBytes)之间，则阈值在(HighMemUsageThreshold, LowMemUsageThreshold)之间
  "lowMemUsageThreshold": "int64",
  "highMemUsageThreshold": "int64",
  "memLowWaterMarkBytes": "int64",
  "memHighWaterMarkBytes": "int64"
}

```

