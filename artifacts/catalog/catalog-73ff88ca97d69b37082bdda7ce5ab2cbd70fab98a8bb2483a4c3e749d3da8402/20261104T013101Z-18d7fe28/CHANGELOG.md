# barness-ai 未发布变更

### 2026-11-04：新增 Claude Haiku 5.5

仅新增 `chat / anthropic / anthropic-messages / claude-haiku-5-5`。
当前目录 79 → 80 条（74 chat、3 image、3 classifier），Anthropic 7 → 8 条、内容版本 10 → 11。
之前 79 条定义、价格及相对顺序保持；9 条旧 Anthropic 身份继续下架。
固定 ID 无日期别名；支持文本/图片、1,000,000 context、128,000 output。
已实现的托管 adaptive effort 配置提供 low/medium/high/xhigh/max 和前缀失配 drop_block，禁发 temperature。
服务默认 medium 作为 metadata 标注；现有 ADR-0019 SDK 的省略 effort 策略仍为 high，调用者可显式选 medium。
本次不声明关闭推理、PDF、Batch、原生浏览器/电脑工具；不更新正式验证目标或其他模型价格。

官方 USD/百万 token 基本费率为 input/output/cacheRead/cacheWrite：0.1/0.5/0.01/0.125。
提示词总量（普通输入+缓存读+缓存写）严格超过 100,000 时，整请求换为 0.5/2.5/0.05/0.625；
等于阈值仍用基础费率，1h 缓存写按选定输入费率的 2 倍计费。
已保留 models.dev 原响应；其 PDF 模态被 CLI 拒绝自动映射。按官方补充入口，
用未改写的官方原捕获、完整人工定义与逐字段决定完成 add/check/apply 和完整 review。
[官方规格](https://platform.claude.com/docs/en/models/haiku-5-5/overview)、
[迁移约束](https://platform.claude.com/docs/en/models/haiku-5-5/migration-guide)、
[官方价格](https://platform.claude.com/docs/en/about-claude/pricing#long-context-pricing)。

| 标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402` |
| Catalog.Hash | `sha256:d7ca061de668054669d4721038fa9cb29c94103815a8d2a318e9ee0025df9703` |
| JSON 文件 SHA-256 | `f29ff70c8b0a70469ee3b924c156e3c33b0bda079a8dbd1de91b25029d59b468` |
| 维护 revision | `sha256:8e46300c8052478e57982ed633b0682c6e2be4bd0ac9a5247653f9efae6e3c9e` |

**代码与分发影响：** HEAD 为 `665862ff90074eecda0a573050114de9ce88c80d`；起点含上一轮未提交的 Anthropic 下架变更。
本次 SDK 核心、运行依赖、schema 无变化，仅新增目录定义、完整基线、公共 E2E、示例配置 v8 与文档。
旧 v1–v7 配置原字节保留，v8 仅更新配置版本及目录 pin。历史下架 E2E 按原冻结身份投影，允许后续独立接入。
原 JSON 已在本次受测程序通过兼容检查；外部加载使用者可只分发 JSON，显式加载并创建新 Client；
现有 Client 不热更新。编译内置默认目录使用者须重建并分发程序。任意其他旧版本未验证。

**验证：** 公共发现 E2E 在采用前预期 RED，采用后已 PASS；已覆盖全部 effort、图片、工具历史、授权、
100k 上下边界、缓存读写触发换价及混合 TTL 计价。候选检查 PASS/exit0；完整普通与最终构建结果见完成索引。
首次 E2E 编译失败及维护 launcher 启动前失败原件保留，修正后独立重跑，不改写失败测量。
live、race、压力按影响不适用；软件发布、部署未执行，不取得厂商新实测资格。
[唯一当前状态](verification/status.md) · [本次交付](delivery.md)。

完整普通实际 PASS/exit0；两段并集覆盖全部 935 个已发现顶层测试，固定必需场景 73 项 PASS，原生 audit 均为 0。最终同源生成、构建、vet 和 live 标签编译/vet 均 exit0。
