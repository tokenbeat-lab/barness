# 工单 06：运行时失败目录（编码前）

公共 Seam 沿用 spec 已确认的 Client.Classify / HookedClient.Classify；混合操作使用
Client.Complete / GenerateImages，观察 Host resolver、loopback HTTP、Result/Attempt、
Observer 与既有资源 Probe。审查固定点为 `26d955b260c2ac3f08bdce3bba810772d5948d49`。

| Case | 先表达的失败 | 必须观察的行为 |
| --- | --- | --- |
| OD40 | 默认重试、429/503/504、头覆盖、Retry-After 顺序/日期/无效值、延迟上限、次数耗尽 | 零值一次；显式有限重试；精确等待、Attempt/厂商 ID；504 差异独立记录 |
| OD41 | 2xx 解码/答案/拒答/读取/回调故障重放 | 复跑 DecisionsUsageAfter2xx；唯一终态、合法 usage、一次响应回调、释放资源 |
| OD42 | Binding/Credential 后端在取消后成功返回；header/payload 取消；connect/header/read-idle/call 超时；body/backoff/admission 取消 | Code/Phase 与 Attempt 正确，可信归属，body/permit/waiter 归零，后续调用可用 |
| OD43 | unary 输出/错误 body 越界、深 JSON、全题/概率响应过大 | 输出/错误 exact/over，最多读 limit+1；深度拒绝；不套 SSE 限额、不重放 |
| OD44 | 同名 Binding/Model、同账户/同 key 误合并租户授权；并发 Jev/Luna/chat/image 串结果/费用/归属 | 同时在途 barrier；分离捕获、结果、价格、Attempt、Observer；race 无发现 |
| OD45 | 首次失败后修改 Binding/Credential/Catalog 与回调引用；凭据后端修改 Binding 入参 | 在明确复制之后修改；所有 retry 字节、头、key、目录/hash、图片权限固定；配置只读一次 |
| OD46 | Observer 拥塞/error/panic、请求/图片/答案/错误体泄漏 | metadata-only，有界计数；成功和失败结果保持，整包审计无发现 |
| OD58 | disabled Binding、移除型号白名单、撤销图片权限影响在途调用或触发 fallback | 新调用前置拒绝，已 pin 调用/重试和已报告费用不变，其他路线/租户继续可用 |

只为新 Decisions 路线补充公共运行时证据；现有 Jev/shared-runtime 回归继续执行。
不建立第二套生产 runtime；压力、Pi Decisions parity、live/发布由后续工单证明。
测试 fixture 的文字/图片/错误均是合成数据；测试 key 已注册脱敏。每次执行新建证据目录，
保留失败历史；最终包包含脚本、捕获、终态、usage、Attempt、Observer、资源、重放命令与 SHA-256。
