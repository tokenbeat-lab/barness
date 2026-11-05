# Anthropic chat 目录下架

2026-11-04 用户要求：Fable、Opus、Sonnet、Haiku 仅保留 4.8 及以上，使用 Catalog CLI SOP 下架其余型号。

范围仅为 chat / anthropic / anthropic-messages。
保留 claude-fable-5、claude-fable-5-1、claude-opus-4-8、claude-opus-5、
claude-opus-5-5、claude-sonnet-5、claude-sonnet-5-5，共 7 条。
下架 Haiku 4.5 两条，Opus 4.5 两条、4.6、4.7，以及 Sonnet 4.5 两条、4.6，共 9 条。
Haiku 无符合门槛的现有型号，不新增推测型号。其他 Provider 和保留定义、价格、顺序保持。

逐身份 withdraw dry-run/confirm → check → apply dry-run/confirm，保留历史。
完整 review、Provider 来源、原字节内置快照和 baseline 同步采用。
交付版本化 JSON、changelog、代码/分发判定和文件摘要索引。
验证工具选择当前在架型号，历史测量不改写，也不授予新型号真实资格。
示例根目录 pin 新建配置 revision，既有配置保持原字节。
验收：Anthropic 16 → 7；总目录 88 → 79（73 chat、3 image、3 classifier）。

测试说明：Anthropic 与 OpenAI 下架 E2E 各自绑定不同的用户规则、冻结快照和历史身份，因此独立维护；两者复用现有 Catalog/Client 受控 world，不把版本筛选规则抽象进 SDK。
