# DeepSeek V4 Pro Responses 目录交付

修正 ADR-0014 的 Pro Chat-only 错误结论，优先补齐 chat / deepseek / openai-responses / deepseek-v4-pro。
Pro 推荐 Responses，由 Host 显式选择 Binding；原 Chat/Flash 及其他定义保持，不新增自动协议路由或隐式 fallback。
目录 81 条（75 chat、3 image、3 classifier），DeepSeek 4 条、内容版本 2；原 80 条定义与价格、顺序不变。
文本、1M context、384k output，当前官方 none/low/high/max effort 映射；服务默认 high 是 metadata，推荐示例显式 high。
使用已有文本/函数工具/推理/完整历史子集；不扩大到 web_search/custom apply_patch。
峰值 input/output/cacheRead/cacheWrite 为 1.32/3.96/0.044/0 USD/百万 token，作为保守固定估算；低谷价为一半，当前不按时段自动计费。

版本：`catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8`。Catalog.Hash：`sha256:e6dae4b736a90c2dc7a6a4a0891c66a535af314be77eb780a7bb2090ef81473d`。
完整 JSON SHA-256：`810061d335f93a2835516234dd9398d65243258e2c733026cca8c1381e9d90b8`。维护 revision：`sha256:b61ef265a715462f70ad6609588e44c4fbbed9b636a95ed92114cb0abe536257`。
[运行 JSON](catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8.json) · [变更记录](CHANGELOG.md) · [普通验证](verification/ordinary-summary.json) · [候选检查](verification/candidate-checks.json) · [状态](verification/status.md) · [ADR 当前修订](verification/ADR-0014-current.md)。

HEAD/实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`，起点含此前未提交的 Anthropic 下架和 Haiku 接入；实际基准由 before-source/before-builtin/原文快照固定。
本次 SDK 核心、运行依赖和 schema 变化均 NO；仅目录/完整基线、相关 E2E、示例 v9 与文档变化，正式 live 目标未改。
官方原捕获与完整人工定义/决定经 add/check/apply/export，三处同步后归集。候选/定向/完整普通/生成/构建/vet/live 标签编译/vet 均通过。
原预期 RED 保留；旧 Flash 厂商测量原件不改写，不变为 Pro 厂商实测资格。

外部加载：仅本次受测程序取得兼容依据，可只分发原 JSON，由 Host 显式加载并创建新 Client；现有 Client 不热更新。其他旧版本未验证。
编译内置：更新默认目录须重建并分发程序。live/race/压力按影响不适用；未提交、未发布、未部署，不产生新厂商实测资格。

review/ 与采用 baseline 相等；官方 Responses/thinking 补充原件在 verification/supplemental-sources。测试胶囊、CLI 与工具依赖外部原路径，见 external-originals.json/候选索引，不声明全部证据自包含。运行 JSON 可独立使用。
恢复前置目录需 withdraw 新身份并重新 check/apply/export、同步三处和另建配置 pin；旧历史不重写。本次未回滚。
