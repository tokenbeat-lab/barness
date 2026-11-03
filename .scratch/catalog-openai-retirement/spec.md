# OpenAI chat 目录下架

2026-11-03 用户要求：只保留 GPT-5.6、GPT-6 及以上型号，其余使用 Catalog CLI SOP 下架。

范围仅为 `chat / openai`。保留 `gpt-5.6-luna`、`gpt-5.6-sol`、
`gpt-5.6-terra`、`gpt-6-astra`、`gpt-6-luna`、`gpt-6-sol`、
`gpt-6.1-sol`、`gpt-6.1-terra` 当前已有的接口身份。
Daybreak 无明确版本号，不在本次保留集合。其他 Provider、image/classifier 和
保留模型的定义、价格及顺序不变。下架记录及历史证据保留。

使用 `withdraw --dry-run/--confirm`、`check`、`apply --dry-run/--confirm`，
逐身份处理，导出完整 review 后同步 Provider 源、原字节内置快照和完整基线。
交付版本化 JSON、changelog、代码/分发判定和全部文件摘要索引。
不执行软件发布、部署或厂商调用。

回归发现当前 OpenAI 正式验证工具仍选择已下架型号。同步其当前目标为
`gpt-5.6-luna`、`gpt-5.6-sol`、`gpt-6-sol`，低推理选项使用已声明的 `off`。
删除会重新添加已下架 pro 型号并请求厂商的旧探针；旧测量、报告和 pin 原字节保留。
此调整只影响验证工具/测试，不改变 SDK 核心，不授予新目标的厂商实测资格。

验收：37 条接口身份下架；OpenAI chat 剩 8 个型号、15 条接口身份。
目录总体由 125 条变为 88 条（82 chat、3 image、3 classifier）。
逐候选公共兼容检查通过，最终目录与保留集合精确相等，其他定义相等。
