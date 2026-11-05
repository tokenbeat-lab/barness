# barness-ai Changelog

## 未发布

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
[ADR 修订](../adr/0014-barness-ai-deepseek-responses.md) ·
[唯一当前状态](../../.scratch/catalog-deepseek-pro-responses/status.md) · [本次交付](../../artifacts/catalog/catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8/20261104T030633Z-d3d7bbbb/delivery.md)。

完整普通实际 PASS/exit0；两段并集覆盖全部 936 个已发现顶层测试，固定必需场景 83 项 PASS，原生 audit 均为 0。最终同源生成、构建、vet 和 live 标签编译/vet 均 exit0。

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
[唯一当前状态](../../.scratch/catalog-haiku55/status.md) · [本次交付](../../artifacts/catalog/catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402/20261104T013101Z-18d7fe28/delivery.md)。

完整普通实际 PASS/exit0；两段并集覆盖全部 935 个已发现顶层测试，固定必需场景 73 项 PASS，原生 audit 均为 0。最终同源生成、构建、vet 和 live 标签编译/vet 均 exit0。

### 2026-11-04：Anthropic chat 仅保留 4.8 及以上型号

按维护者要求，`chat / anthropic / anthropic-messages` 保留 Fable 5、5.1，
Opus 4.8、5、5.5，Sonnet 5、5.5，共 7 条身份。Haiku 当前无满足门槛的型号。
逐条下架 `claude-haiku-4-5`、`claude-haiku-4-5-20251001`、`claude-opus-4-5`、
`claude-opus-4-5-20251101`、`claude-opus-4-6`、`claude-opus-4-7`、
`claude-sonnet-4-5`、`claude-sonnet-4-5-20250929`、`claude-sonnet-4-6`，共 9 条。
全部执行 CLI withdraw 预览/确认、check、apply 预览/确认；旧定义和追溯保留。
总目录 88 → 79 条（73 chat、3 image、3 classifier）。Anthropic 内容版本 1 → 10。
其他 Provider 版本及全部保留定义、价格、相对顺序不变。

| 标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5` |
| Catalog.Hash | `sha256:40d4c5fe7551fb06e79210b676e69e96a47b6b1b200dde5f6c07e243f970b2f0` |
| JSON 文件 SHA-256 | `44d7caa8f36ce1b5b7914d014215be1e3625b453677cabe54d891b59ac5a9bb8` |
| 维护 revision | `sha256:2047fad03a8020c866481cb8549e9e0a28f4a0a767646e65f4e13b41a381349f` |

**代码与分发影响：** 比较基准/当前 HEAD 为 `665862ff90074eecda0a573050114de9ce88c80d`，
实际受测为本次工作区。SDK 核心运行代码、运行依赖和 schema 均无变化。
维护源、完整审阅基线、生成 JSON、相关 E2E、验证工具型号和示例配置有变化。
正式验证默认 Haiku 4.5 调整为 Sonnet 5.5；历史报告及原测量不改写，不授予新真实资格。
ChineseEvaluation 新建 v7，仅变更配置版本与目录 pin；v1–v6 原字节保留。
外部加载使用者可向本次实际受测程序分发原 JSON，再显式加载并创建新 Client；
现有 Client 不热更新。任意其他旧程序版本兼容性未验证。
编译内置目录使用者需要重建并分发程序，才能取得本次默认目录。

**验证：** 9 次候选检查全部 PASS/exit0；公开 Client 下架 E2E 从预期 RED 转为 PASS。
完整受影响普通回归及最终同源生成/构建/vet 的结果由完成索引记录。
live、race、压力按纯下架影响不适用；软件发布、部署均未执行。
[唯一当前状态](../../.scratch/catalog-anthropic-retirement/status.md) ·
[本次交付](../../artifacts/catalog/catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5/20261103T222139Z-9274b0a7/delivery.md)。

完整普通回归实际 PASS/exit0；两段并集覆盖全部 934 个已发现顶层测试，固定必需场景 54 项 PASS，原生 audit 均为 0。最终同源生成、构建、vet 和 live 标签编译/vet 均 exit0。

### 2026-11-03：OpenAI chat 下架旧型号

按维护者要求，仅保留 GPT-5.6、GPT-6 及以上的现有 OpenAI chat 身份：
`gpt-5.6-luna`、`gpt-5.6-sol`、`gpt-5.6-terra`、`gpt-6-astra`、
`gpt-6-luna`、`gpt-6-sol`、`gpt-6.1-sol`、`gpt-6.1-terra`。
其余 20 个型号（含 Daybreak 别名）的 37 条四维身份逐条通过
`withdraw` 预览/冻结、`check`、`apply` 预览/确认下架；完整身份清单保存在交付记录。
剩余 OpenAI chat 为 Responses 8 条、Chat Completions 7 条。
总目录从 125 条缩减为 88 条（82 chat、3 image、3 classifier）。
下架定义、来源和追加历史保留，其他 Provider、image/classifier、保留模型定义、
价格及相对顺序不变。OpenAI 内容版本由 2 升至 39，其他 Provider 版本不变。

| 目录标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2` |
| Catalog.Hash | `sha256:9dff935e34e42b96adc8afeb34c2db689447c936f1c584928ea08b3ed0a7b65a` |
| JSON 文件 SHA-256 | `e706fdc1dd13aed8799478da342467a810005e769420abce1ed0493dd8c0d9e6` |

**代码与分发影响：** 比较基准为 `b041eff5ca74393a88f66916b13ce1aaeec12ee8`。
核心运行代码、模型字段和运行依赖未改变。变更包括目录数据与完整审阅基线、
相关 E2E 历史夹具装配、文档、OpenAI 验证工具默认目标，以及 ChineseEvaluation 默认配置 v6（只改配置版本
和目录 pin，v1–v5 原字节保留）。有外部加载能力且通过本次检查的受测程序可只
接收版本化 JSON，Host 加载后创建新 Client；现有 Client 不热更新。编译内置默认
使用者需要重建并分发程序。其他旧程序版本的兼容性未验证。

OpenAI 验证工具默认目标改为 `gpt-5.6-luna`、`gpt-5.6-sol` 和 `gpt-6-sol`，
quiet 使用已声明支持的 `off`；删除会重新加入已下架 pro 型号并请求厂商的旧探针。
这涉及 `ai/internal/testkit/supportmatrix/` 与 `ai/live/` 的验证合同和目标，不改变
SDK 核心。旧报告、测量和 pin 不重写，新目标尚未取得本次新的厂商实测资格。

**验证：** 37 次候选代码兼容检查均实际 PASS/exit0。内置下架 E2E 先 RED，
同步后目录/保留调用/37 条拒绝/完整基线及当前配置定向普通 PASS/exit0。
完整受影响普通范围及生成、构建、vet 的实际结果、源码输入和回执见
[本次唯一状态](../../.scratch/catalog-openai-retirement/status.md)；未执行不计通过。
live 不适用：无新厂商事实；race/压力不适用：并发、资源及设计负载未改变。
本次没有软件发布、部署或新型号厂商资格声明。
版本化 JSON、完整审阅包、changelog 快照和完成索引见
[本次交付](../../artifacts/catalog/catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2/20261103T114943Z-b12de1f9/delivery.md)。

最终完整普通实际 PASS/exit0，墙钟 1109.84 秒；固定必需场景 42 项全部通过，全部 audit 为 0。最终同源生成、构建、vet 均 exit0。实际受测工作区源码/输入与结果以本次状态及不可变完成索引为准。

### 2026-11-02：目录定义型号，运行库校验能力

移除 Kimi chat、xAI Responses chat、xAI image 和 TypeSafe SystemOne 图片/Box
classifier 按具体型号 ID 准入的限制。目录仍须提供非空、唯一的四维身份和当前
实现支持的协议、模态、容量、推理、兼容标志及价格；精确查找、响应身份、原生
历史约束和 Host `Binding.AllowedModels` 授权保持。新目录身份不自动获得调用权限
或厂商实测资格。Gemini 请求和历史编码、OpenAI 服务等级计价中的型号分支属于
独立行为，本次保留；决定及适用范围见 [ADR-0059](../adr/0059-catalog-defined-model-identities.md)。

本次改变 SDK 核心代码；要使用新准入策略，使用方需先升级并重建一次。既有程序
使用相同的 125 条目录不因此需要升级。此后在这四类已实现能力范围内，
经审核的外部 JSON 可用于新增同类身份，Host 需显式加载并创建新 Client；仍须按
目标程序验证兼容。更新编译内置默认目录仍需重建分发。当前 125 条内置定义、
维护源、完整基线及目录版本和哈希不变，合成测试型号不会加入真实目录。

**验证：** 完整普通在 `8d97cb6002722aceeb7fb3ccfaf5eb27e0991f11` 实际
PASS、exit 0；329 个固定必需场景全部通过，主包 17,232 PASS、226 NOT_RUN，
四个 audit 均为 0。生成、构建和 vet 各自 exit 0。未运行场景不算通过；旧失败和
取消原件保留。53 个动态场景对应 48 个独立 packet，历史 archive 的 2,189 项文件
及九条合成 CLI 检查已核验；历史来源通过仅表示固定历史上下文符合原合同，不授予
当前程序厂商资格。文档收尾后 2,171 个源码和 84 个选定输入与受测原件相等。
live 不适用：无需新厂商事实；race 不适用：共享状态与取消路径未变；压力不适用：
设计负载未变。未签发软件版本或新厂商资格；见[本次状态](../../.scratch/catalog-model-id-policy/status.md)。

### 2026-11-02：Kimi K3.1 Pro 内置目录

目录新增 `chat / moonshot` 与 `chat / moonshot-cn` 的 `kimi-k3.1-pro`，
每站分别提供 `openai-completions`、`openai-responses`、`anthropic-messages`
三条身份。目录共 125 条，原 119 条定义、价格和相对顺序保持。
两站 Provider 内容版本分别由 1 增至 4，其余 Provider 版本保持。
本库当前实施范围为文本输入、最多输出 131,072 token；该上限不代表厂商最大值。
Host 的认证与型号授权不会因目录更新自动扩大。

| 目录标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c` |
| Catalog.Hash | `sha256:93a1692f982836d9ad478252313b73e75eeee4904e6d79e3bcc915cd54a44005` |
| JSON 文件 SHA-256 | `6fb2842be21797a6aa448a92244d12b4c528c212d74eff8e16220a883e89818a` |

国际站标准 USD/百万 token 为输入 2、输出 10、缓存读 0.2、5 分钟缓存写 2；
中国站按 2026-11-02 的 1 USD = 6.7146 CNY 固定换算，分别为 1.9360795878、
10.1271855360、0.1936079588、1.9360795878。1 小时缓存写为已舍入输入费率的
两倍；中国站美元值是固定换算估计，不是厂商美元报价。

**代码与分发影响：** 相对 `ceabec8b58c42377f4f643ccb87c1bc1d4beb456`，
`ai/kimi_catalog.go` 核心型号白名单新增该 ID，运行依赖不变；文档收尾后 2,166 个
受测源码和 167 个选定输入与 `97e3fef86d6948705eb70abc97976c64c725b442` 的
实测原件逐项相等。旧基准程序对首个国际站 Chat 候选的 120 条快照实际返回
`code_gap`、exit 1；该探针不是对最终 125 条快照的旧程序实测。旧程序需升级
SDK 并重建后，再经兼容检查和 Host 显式加载才能使用新 JSON。其他程序版本未验证。
已将整组维护源、原快照及完整审阅基线同步入库；使用编译内置默认目录的程序
需要重建分发。ChineseEvaluation 示例默认配置更新至 v5，仅修改配置版本和目录
身份；v1 至 v4 原字节保留，任务、型号、数据集和预算不变。

**验证：** 六条候选与应用原件已留存，内置定向 GREEN 与完整普通均 PASS、exit 0。
完整普通实际受测提交 `97e3fef86d6948705eb70abc97976c64c725b442`，墙钟
21 分 44.86 秒；主包 17,124 个场景 PASS、226 个 NOT_RUN，203 个必需场景全部
PASS，audit 0。生成、构建和 vet 各自 exit 0。旧失败、取消和因 HEAD 变化标为
`INVALID_EVIDENCE` 的回执均保留，不计为通过。live 不适用：没有新厂商事实或
资格声明；race 不适用：没有共享状态或取消路径变化；压力不适用：没有设计负载
变化。未执行项不算通过；不声明新厂商资格或软件发布。实际证据和当前状态见
[本次状态](../../.scratch/catalog-kimi-k31-pro/status.md)。

### 2026-11-02：Catalog CLI 交付规范

目录维护 SOP 现在要求每次更新 changelog，使用真实 Catalog.Version 命名不可变
JSON 快照，并把本次交付物归集到 `artifacts/catalog/<Catalog.Version>/<批次>/`。
每次同步内置目录必须明确核心运行代码/依赖是否变化，以及外部 JSON 与编译内置
两类使用者的升级或重建要求。完整索引绑定文件摘要和原验证引用。

当前 Terra 交付已按该规范归集，见[统一交付入口](../../artifacts/catalog/README.md)。
本次只更新流程文档及归集原产物，没有修改 SDK、CLI 实现或 JSON 格式。

### 2026-11-02：GPT-6.1 Terra 同步内置目录

新增 `chat / openai / openai-responses / gpt-6.1-terra`，支持文本和图片输入。
默认目录由 118 条增至 119 条（113 chat、3 image、3 classifier），原 118 条
定义、价格和相对顺序保持。OpenAI 内容版本由 1 升至 2，其他 Provider 保持 1。
维护源、内置快照和完整 20 文件审阅基线同时采用已批准 revision。

| 目录标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6` |
| Catalog.Hash | `sha256:ba76c9699428740561b35f8e66412cb697570f9e34b90a5680b66d0c455e0717` |
| JSON 文件 SHA-256 | `04107af45ba03e3a4e7660aa5255997303b8c1442b74fc6d905b37fba0ba80ca` |

**代码与分发影响：** 本轮相对 `37a283ee` 没有修改 barness-ai SDK 核心、Provider
适配器、目录加载器或依赖。同一 JSON 已在 `37a283ee` 受控程序上通过代码兼容检查；
仅对该受测程序且 Host 显式外部加载/注入的情况，给出可只分发 JSON、重新加载并
创建新 Client、无需为 Terra 升级 SDK 的结论。其他程序版本仍需各自兼容检查。
使用 `ai.BuiltinCatalog()` 的程序
将目录编入二进制，需要重新构建和分发程序才能取得新的默认目录。
尚不具备外部目录加载能力的旧程序需要先升级；本次未验证任意旧发布版。

ChineseEvaluation 示例 CLI 的默认配置由 v3 改为 v4。三份新配置仅改变配置版本、
目录版本和目录哈希，任务、型号、数据集和有限预算保持。v1/v2/v3 及原目录快照
保留，严格拒绝不匹配的目录 pin；历史报告不会因此获得新的资格。

**验证：** GPT-6.1 Terra 子 agent 完成 28 个必需原生场景（含 14 个 Terra 场景），
全部 PASS；完整受影响普通回归、生成、构建与 vet 均 exit0，4 个 bundle audit
发现为 0。实际受测提交为 `d4dbafaa1537ada3f6a56ade817dc63ad20d9f9d`。
live、race 和压力按本次影响判定为 NOT_APPLICABLE；没有新型号厂商实测资格、
线上发布或部署结论。“未发布”不表示已签发软件版本号。

来源与完整原件见[目录维护 SOP](catalog-maintenance.md)和
[本次唯一当前状态](../../.scratch/catalog-gpt61-terra/status.md)。
