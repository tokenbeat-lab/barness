# barness-ai 未发布变更

### 2026-11-04：DeepSeek V4 Pro 优先接入 Responses

修正 ADR-0014 中“Pro 只在 Chat”的错误排除结论。官方已明确支持 Pro Responses；本次优先新增
`chat / deepseek / openai-responses / deepseek-v4-pro`，把 Responses 作为 Pro 的推荐调用入口。
Host 显式选择 Binding，不新增自动协议选择或隐式 fallback。Pro Chat 与 Flash 双接口独立保留。
目录 80 → 81 条（75 chat、3 image、3 classifier），DeepSeek 3 → 4 条、内容版本 1 → 2。
原 80 条定义、价格及相对顺序不变，其他 Provider 不变。

新增身份为文本输入、1,000,000 context、384,000 output；Responses effort 映射为
none/low/high/max（minimal→low、medium/xhigh→high），服务默认 high 作为 metadata。
使用已有无状态 Responses 子集：文本、函数工具、推理和完整历史；不扩大到 web_search/custom apply_patch。
价格冻结峰值 USD/百万 token：input 1.32、output 3.96、cacheRead 0.044、cacheWrite 0。
官方低谷价为一半；此处是保守峰值估算，不按日历自动计费，也不声称等同任意时段真实账单。
[官方 Responses 指南](https://api-docs.deepseek.com/guides/responses_api/)、
[模型与价格](https://api-docs.deepseek.com/quick_start/pricing/)、[推理映射](https://api-docs.deepseek.com/guides/thinking_mode/)。

| 标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8` |
| Catalog.Hash | `sha256:e6dae4b736a90c2dc7a6a4a0891c66a535af314be77eb780a7bb2090ef81473d` |
| JSON 文件 SHA-256 | `810061d335f93a2835516234dd9398d65243258e2c733026cca8c1381e9d90b8` |
| 维护 revision | `sha256:b61ef265a715462f70ad6609588e44c4fbbed9b636a95ed92114cb0abe536257` |

按官方原捕获、完整人工定义/逐字段决定执行 CLI add 预览/确认 → check → apply 预览/确认，
导出完整 review 并同步维护源、原字节内置快照和完整 baseline。
比较 HEAD 为 `665862ff90074eecda0a573050114de9ce88c80d`，起点含此前未提交的 Anthropic 下架/Haiku 接入；
实际基准另由本批 before-source、before-builtin 和 before-code 保存。SDK 核心、依赖、schema 无变化。
Go 变化仅相关 E2E 与示例默认目录 pin v9；v1–v8 原字节保留。旧 ADR 原文和测量留存，当前勘误不冒充 Pro 厂商实测。

公开 Client 验证先因缺型号 RED，采用后专项 PASS；覆盖字段、映射、工具/历史、价格、授权、
无状态字段隔离、ServiceTier 拒绝及无隐式 Chat fallback。候选完整检查 PASS/exit0。
完整普通、生成、构建、vet 的真实结果见完成索引。live/race/压力按影响不适用；未发布软件或部署。
已通过兼容检查的原 JSON 可仅分发给本次受测程序，由 Host 显式加载并创建新 Client；现有 Client 不热更新。
编译内置默认目录须重建分发；其他旧版本兼容性未验证。
[ADR 修订](verification/ADR-0014-current.md) ·
[唯一当前状态](verification/status.md) · [本次交付](delivery.md)。

完整普通实际 PASS/exit0；两段并集覆盖全部 936 个已发现顶层测试，固定必需场景 83 项 PASS，原生 audit 均为 0。最终同源生成、构建、vet 和 live 标签编译/vet 均 exit0。
