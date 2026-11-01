# barness 动态模型目录：背景与可行性研究

历史研究日期：2026-10-29（Asia/Shanghai）。导航更新：2026-11-01。状态：历史 Pi 背景研究保留；下文的候选建议不作为当前规范。

当前入口：

- [models.dev 指定端点研究](models-dev-research.md)：实际 v1 schema、字段意义、有损投影、barness 映射与 118 条初始迁移覆盖。
- [当前设计](design.md)：本期 JSON 加载、维护 CLI 与候选受控验证的有效范围和决策。
- [讨论记录](discussion.md)及 [ADR-0058](../../docs/adr/0058-barness-catalog-maintenance-and-isolated-loading.md)：当前架构边界；[ADR-0057](../../docs/adr/0057-barness-independent-model-catalog.md) 保留旧版历史。

本报告保留 2026-10-29 当时的事实和建议，包括 Pi 的线上更新、最低版本和发布机制。当前主要维护源已改为 models.dev；本期不实施线上发布或在线更新，不采用最低 barness-ai 版本、`schemaVersion` 或旧软件矩阵。下文与当前设计冲突的建议仅作历史比较，不应据此实施；受控兼容验证也不代表厂商实测资格。

本地基线：`5be6828d13561cf0fee2b7e56f1cd0f9678da7af`。Pi 当前源码基线：`e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a`；历史对照：Pi `v1.1.0`，`abe508e1b89912adde45528136c3221eb69acdd7`。外部核心结论以固定提交源码为准，避免 `main` 页面缓存与移动分支混用。

本次只研究和记录事实。不修改产品代码，不发起模型推理或付费 live 请求。按[测试流程 SOP](../../docs/agents/verification.md)，文档验证只检查来源、引用及事实对应，不以文档研究宣称产品验证或发布通过。

## 1. 结论先行

1. **可行。** 可以把“目录数据更新”从“barness-ai 代码发布”中分离；现有 `Config.Catalog` 已提供装配入口，调用方可以交付自有目录。首版若在启动时加载，主要不确定性在维护和兼容政策，公共调用接口有较好的保留空间。[Catalog](../../ai/catalog.go)、[Client 构造](../../ai/client.go)
2. **用户所指的 pi.dev 属实。** 当前 Pi coding-agent 运行时会读取 pi.dev 的 Provider 目录；它不是仅在构建时从 models.dev 生成源码。Pi 同时保留内置基线、独立目录发布和运行时缓存三部分。[远程目录实现](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts)、[生成器说明](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/README.md)
3. **“协议兼容”不足以保证“旧程序能使用新目录”。** 新硬约束、Provider 代码分支和计费解释也需要旧运行时理解。Pi 为此保留最低客户端版本与历史目录；其维护者还记录过目录更换 API、旧客户端继续使用错误实现的事故。[最低版本](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/model-catalog-minimum-version.mjs)、[兼容检查的事故说明](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/check-model-catalog-compat.mjs)
4. **barness 需要一次边界整理，不能只增加下载器。** xAI、Kimi、部分图片/分类能力目前明确绑定固定型号；这些限制包含已接受的首期支持边界。若要同协议新型号以后只发数据，须重新确认并迁移这些边界，不能直接删除校验。[xAI 校验](../../ai/catalog_xai.go)、[Kimi 校验](../../ai/kimi_catalog.go)、[图片校验](../../ai/image_pricing.go)、[分类能力](../../ai/classifier_capabilities.go)、[ADR-0054](../../docs/adr/0054-barness-ai-xai-responses.md)、[ADR-0056](../../docs/adr/0056-barness-ai-xai-images.md)、[Kimi R01](../barness-ai-kimi/spec.md)
5. **建议借鉴 Pi 的分层，不直接照搬其数据和所有运行时策略。** barness 自有目录应保有发布权；外部数据用于发现与整理。目录中出现、客户端能执行、租户获授权、取得自身支持证据，是不同事实。具体纳入承诺和生效时机留待讨论。[目录纳入准则](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)、[操作与绑定授权](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md)

## 2. 需要区分的四种版本与证明

以下是研究中的分析维度，不表示已新增领域术语。

| 维度 | 回答的问题 | barness 当前基础 | 缺口或待决定事项 |
| --- | --- | --- | --- |
| 目录数据格式版本 | 能否正确解析字段及其含义？ | `Catalog` 与三种模型类型已有 JSON 形状 | 还没有独立的公共发布包 schema 与统一加载契约 |
| 目录内容版本与哈希 | 本次到底采用哪一份能力和价格？ | `Catalog.Version`、`CatalogHash` 已进入调用元数据 | 发布包原始字节摘要应与既有 typed Catalog 哈希区分 |
| 运行时实现兼容边界 | 此程序是否实现目录要求的协议、参数、硬约束和价格解释？ | 固定 adapter、Provider profiles、`ModelCompat` 等代码 | 需决定用最低运行时版本还是更细的能力声明表达；首版不宜同时引入复杂的能力协商 |
| 支持证据 | 这一操作 × Provider × 协议 × 型号已有何种自身验证？ | 原生路线、支持矩阵、目录准则和留存证据 | 需决定新型号“声明兼容”与“正式支持”的关系，不自动扩大历史证据 |

依据：[Catalog](../../ai/catalog.go)、[哈希实现](../../ai/catalog_snapshot.go)、[调用元数据](../../ai/call_runtime.go)、[adapter 注册](../../ai/adapter.go)、[ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md)、[ADR-0055](../../docs/adr/0055-route-scoped-evidence-applicability.md)。

用户提出的“每家 Provider 保证新型号向下兼容协议”可作为设计前提接受。它解决的是服务端继续接受旧协议；不替代后三个独立问题：目录字段能否被旧程序理解、旧实现是否满足新增硬约束、barness 对型号承担什么支持承诺。

## 3. Pi 实际如何更新

### 3.1 从上游到用户进程的完整链条

```text
models.dev 等上游 + Pi 自有 Provider/型号族规则
    → Pi 目录生成、校验与客户端兼容检查
    → R2 上的不可变内容 revision + 目录索引
    → pi.dev 按客户端版本选择 Provider 目录
    → coding-agent 恢复缓存、联网刷新、叠加内置目录
    → 已实现的 Provider/API 代码执行请求
```

这个链条分别位于 `packages/ai/catalog/`、仓库发布脚本/CI 和 `packages/coding-agent/`。不能把“pi-ai 的内置目录生成”与“Pi 应用的在线目录更新”说成同一机制。[生成器说明](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/README.md)、[发布工作流](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/.github/workflows/publish-model-catalog.yml)、[宿主装配](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/model-runtime.ts)

### 3.2 目录来源和归属

**事实：** 当前 Pi 生成器主要读取 models.dev v2 的 experimental API，同时有 OpenRouter、Vercel AI Gateway 等独立来源；Provider 定义补充 API、目标、默认兼容项、纠错和上游没有的模型；型号族规则再应用共同事实。输出既能用于随程序发布的内置数据，也能生成 pi.dev 所需的 JSON。[生成规则和输出](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/README.md)

**事实：** models.dev 自称社区共同维护的开放模型数据库，数据通过版本库中的文件维护并生成 API；它不是所有厂商共同担保的协议支持证书。[models.dev 项目说明](https://github.com/anomalyco/models.dev)

**事实：** Pi 对上游做显式归一化。当前转换只把按上下文阈值变化的部分价格映射为其阶梯价格；地区、服务等级和时间条件价格没有直接对应项。转换到 Pi 成本结构时，缺失基础费率会落到零。[上游转换实现](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/sources/models-dev.ts)

**推断：** barness 若使用 models.dev，应在维护端吸收它，而不是把其响应视作已经满足 barness 的 `Catalog` 与价格语义。尤其不能把外部缺失费率直接当成免费；应沿用本库对估算依据和缺失事实的区分。[估算成本和价格快照定义](../../GLOSSARY.md)

### 3.3 发布机制

| 已核实行为 | 源码依据 | 可借鉴之处 |
| --- | --- | --- |
| CI 成功、定时任务或手动入口可以触发生成；生成后校验并执行客户端兼容检查，之后发布 R2 | [发布工作流](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/.github/workflows/publish-model-catalog.yml) | 目录可独立于程序版本发布；“自动发现”不等于跳过发布校验 |
| revision 是完整 `models.all.json` 字节的 SHA-256；目录按 revision 保留分片，先上传内容再切换索引 | [发布脚本](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/publish-model-catalog.mjs) | 内容可追溯，上传中断不应提前宣布新目录可用 |
| 索引保留各个最低 Pi 版本对应的目录，记录 source commit 与发布时间；同一最低版本更新其当前 revision | [发布脚本](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/publish-model-catalog.mjs) | 历史客户端可以停留在适合自己的目录 |
| 存储协议有单独 schemaVersion；请求可以区分旧聊天表示与全类型表示 | [目录协议](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/model-catalog-protocol.ts) | 格式演进与内容更新分开处理 |

这里能证明内容按 revision 保留以及“最后发布索引”的顺序；**不能据此宣称 Pi 已提供面向用户的完整回滚命令、签名体系或撤销机制**。这些不在本次取得的公开源码证据中。barness 初版可以先讨论“保留不可变发布包、把发布指针改回旧包”的简单回滚语义，不需要先建完整供应链平台。

### 3.4 运行时刷新、缓存和覆盖

**事实：** 主要远程包装器位于 coding-agent，默认请求 `https://pi.dev/api/models/providers/{provider}`，携带支持类型。普通目录以 4 小时为新鲜窗口、4 秒为单次尝试超时；支持强制刷新、ETag/304，比较远端 `Last-Modified` 和本地内置数据生成时间；失败保留已有状态，未知模型类型被过滤。[远程包装器](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts)

**事实：** 默认缓存为 agent 目录中的 `models-store.json`，按 Provider 保存模型、检查时间、修改时间和 ETag；持久化读写借用文件锁实现，内存交接做独立复制。[持久缓存](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/models-store.ts)、[存储接口](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/models-store.ts)

**事实：** 内置 Provider 在应用装配时加远程目录包装；Radius 走自己的动态能力，OpenRouter 有按用户 key 个性化的特殊路径。`ModelRuntime.create` 默认不在创建时联网；网络权限、离线模式和后续刷新入口共同决定是否请求，不能说每个 SDK 构造都会自动联网。[装配代码](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/model-runtime.ts)

**事实：** 官方用户文档提供 `pi update --models` 强制刷新；自定义 `models.json` 可以添加/替换型号，`modelOverrides` 可修改既有型号的元数据。[用户文档](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/docs/models.md)

**限制：** 普通远程路径是内置目录加 overlay，同 Provider 内以类型和型号 ID 覆盖。远端不再列出某条内置型号，并不会把内置条目删除。因此它天然适合补充/纠错，不直接表达完整目录替换或型号撤回。barness 不应在没有讨论删除语义时照搬 overlay。[合并实现](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts)

### 3.5 新目录如何避免伤害旧程序

**事实：** Pi 公开共享协议根据 User-Agent 中的版本重定向到显式 `pi-version` URL，再选择不高于该客户端版本的最高 `minimumPiVersion`。这属于服务端目录选择，不只是告警；`types` 仅区分响应表示，不是对所有协议能力逐项协商。[版本选择协议](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/model-catalog-protocol.ts)

**事实：** 当前最低版本是 `1.2.1`。源码要求在目录需要旧客户端没有的行为时提高门槛，并让旧客户端留在之前门槛的目录；注释明确以 kimi-coding 从 Messages 迁移到 Chat Completions 为例。[最低版本定义](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/model-catalog-minimum-version.mjs)

**事实：** 发布前兼容检查针对 **npm 最新 release**：取该 release 的已发布内置目录，叠加候选远程数据，把数据放回旧 release 源码并运行 TypeScript 编译。它能发现部分 Provider 与 API 类型不一致；并不是对所有仍受服务的历史版本逐一执行，也不是请求行为的 E2E 证明。[检查实现](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/check-model-catalog-compat.mjs)

**事实：** 远程解析没有完整运行时 schema 验证，只检查基本容器、`id` 存在和已知类型，再将 Provider 固定为本请求的 Provider。未知 compat 字段即使被对象保留，也不意味着旧实现会执行它；发布索引的 schema 校验不能代替模型语义校验。[远程解析](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts)

**事实：** Provider 只有已有代码实现：单 API Provider 使用固定实现；多 API、图像与分类路径按 `model.api` 选择已注册实现。无对应实现会产生错误。JSON 不能为旧程序安装新 adapter。[Provider 构造与派发](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/models.ts)

**推断：** Pi 的总体方法值得借鉴，但它不是“目录天然向下兼容”的证明。barness 应将“最低可用实现”作为自有目录发布约束；新硬约束无法用旧字段表达时，旧程序应继续使用兼容目录或明确拒绝该发布包，不能靠 JSON 解码成功判断安全可用。

### 3.6 历史对照与线上观测

Pi `v1.1.0` 已经包含 pi.dev 运行时 overlay、4 小时窗口和 ETag 逻辑，不能把动态更新误认为当前开发分支才出现的能力。该 tag 的路径仍是 `packages/ai/scripts/generate-models.ts`；当前开发分支已转为 `packages/ai/catalog/`。[v1.1.0 远程实现](https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/core/remote-catalog-provider.ts)、[历史生成器](https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/ai/scripts/generate-models.ts)、[当前目录布局](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/catalog/README.md)

2026-10-29 对公开目录进行无认证 GET，实际观测到：

- 以 `User-Agent: pi/1.2.1` 请求全类型 OpenAI 目录，最终 URL 带 `pi-version=1.2.1`。
- HTTP 200，JSON 数组，返回 ETag、Last-Modified 和 60 秒公共缓存头。
- 未传 `types` 的旧表示为型号对象映射。此观测说明公开服务与上述协议的对应行为存在，不证明整个私有服务实现均与公开注释一致。

可重复只读观测命令：

```sh
curl -sS -L -D - -A 'pi/1.2.1' \
  'https://pi.dev/api/models/providers/openai?types=chat,image,classifier' \
  -o /tmp/pi-openai-catalog.json
```

线上输入：[全类型公开端点](https://pi.dev/api/models/providers/openai?types=chat,image,classifier&pi-version=1.2.1)、[旧表示公开端点](https://pi.dev/api/models/providers/openai)。两者是移动数据，不应拿本次返回的型号数和 ETag 作为未来不变的断言。

公开协议源码提到 `earendil-works/pi.dev` 中有同份文件，但该仓库本次公开访问为 404。因此服务端规则的依据是 Pi 仓库公开共享协议及上述在线观测，未声称审阅了 pi.dev 的全部服务端源码。[共享协议说明](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/scripts/model-catalog-protocol.ts)

## 4. barness 现有能力和真实障碍

### 4.1 现有 Catalog 可保留的部分

| 事实 | 位置 | 对本需求的含义 |
| --- | --- | --- |
| `Catalog` 已包括 `Version`、聊天、图像和分类模型集合；`Config.Catalog` 可注入 | [Catalog](../../ai/catalog.go)、[Config/NewClient](../../ai/client.go) | 不必要求其他模块改成消费 Pi 或 models.dev 的数据类型 |
| 构造时复制、验证、计算目录哈希并建索引；查询返回独立副本 | [Client](../../ai/client.go)、[查询](../../ai/catalog_query.go)、[快照](../../ai/catalog_snapshot.go) | 可以在边界把发布包转成现有 Catalog，然后让调用保持现有只读方式 |
| 一次调用记录目录版本/哈希；绑定允许型号另行授权 | [运行时](../../ai/call_runtime.go)、[Binding](../../ai/binding.go) | 目录更新不能顺带扩大租户权限，价格追溯仍需固定调用快照 |
| 型号身份包含操作、Provider、API、ID | [ADR-0020](../../docs/adr/0020-barness-ai-model-operations-and-binding-authorization.md) | 不应照抄仅按 Provider/type/ID 的去重方式 |
| ADR-0018 已区分硬约束与可选优化 | [ADR-0018](../../docs/adr/0018-barness-ai-catalog-inclusion-criterion.md) | 已有表达能力覆盖全部硬约束时可走数据更新；新增硬约束仍走代码适配 |

### 4.2 必须处理、但不能在研究阶段自行放宽的边界

1. **固定型号校验。** xAI chat 只接受 `grok-4.7`；Kimi 只接受 `kimi-k3`；xAI 图片只接受 `grok-imagine-image-2.0`；Jev 的图片/Box 能力也绑定特定 `jev-1.14.0`。这些包括明确的首期支持决策，不是新增 JSON 条目就能通过的通用目录规则。[xAI](../../ai/catalog_xai.go)、[Kimi](../../ai/kimi_catalog.go)、[图像](../../ai/image_pricing.go)、[分类](../../ai/classifier_capabilities.go)、[ADR-0054](../../docs/adr/0054-barness-ai-xai-responses.md)、[ADR-0056](../../docs/adr/0056-barness-ai-xai-images.md)
2. **Provider/型号行为仍有代码来源。** Provider 协议能力私有于实现；部分 Gemini 预算与回放形态、OpenAI 服务等级费用倍率依型号作代码判断。新型号即使能通过目录验证，也可能仍需要行为评估。[Responses 能力](../../ai/responses_capabilities.go)、[Chat 能力](../../ai/chat_compat.go)、[Messages 能力](../../ai/anthropic_capabilities.go)、[Gemini 选项](../../ai/gemini_options.go)、[Gemini 请求](../../ai/gemini_request.go)、[Responses 用量](../../ai/responses_usage.go)
3. **嵌入数据装载不等于公共发布协议。** MiniMax、地区目录和 Kimi 内置数据有固定版本、数量、型号或来源哈希检查。这些用于当前维护的已审核数据，不能机械改成读 URL。[MiniMax](../../ai/catalog_minimax.go)、[地区目录](../../ai/catalog_regional.go)、[Kimi 数据](../../ai/catalog_kimi_data.go)
4. **JSON 可解码不等于可公开分发。** 当前没有统一远程发布包加载器；普通聊天校验没有全面强制版本、ID、Provider 与正数 limits；部分嵌入 JSON 的严格解码不形成全局契约。远端边界应独立验证输入，再转换为现有类型。[目录验证](../../ai/catalog_validate.go)、[目录快照](../../ai/catalog_snapshot.go)
5. **现有 Catalog 哈希不是下载文件哈希。** 它针对 Go 类型 JSON 表示计算，数组顺序参与结果。未来下载包的字节完整性摘要和既有运行时目录身份不能混成一个值。[哈希实现](../../ai/catalog_snapshot.go)
6. **长驻进程热生效需要单独设计。** `NewClient` 固定目录，同时持有 admission limiter 与 observer 等状态。不断替换重叠存活的 Client 可能叠加局部容量；不能把“反复 NewClient”宣称为无影响的热更新实现。[Client 生命周期](../../ai/client.go)、[快照一致性决策](../../docs/adr/0003-barness-ai-snapshot-consistency.md)
7. **下游已有有意固定目录的消费者。** 中文评估配置会比较目录版本/哈希与传入目录。自动换成最新完整目录可能使旧配置被拒绝；“尽量不影响其他模块”应允许继续选择历史快照，或显式更新其配置，不能静默重写 pin。[评估计划加载](../../ai/examples/chineseeval/plan.go)、[路线证据适用性判断](../../ai/internal/testkit/release/evaluation.go)

## 5. 可选路径比较

### 5.1 谁拥有目录发布权

| 路径 | 新型号到达速度 | 与现有语义契合度 | 判断 |
| --- | --- | --- | --- |
| 继续随 Go 程序内置并发版 | 取决于代码发版 | 高 | 可保留作离线基线，单独使用无法解决当前问题 |
| 运行时直接采用 models.dev | 快，但同步外部格式和修订 | 低：不含 barness 固定边界与自身证据 | 不建议作为正式目录权威源 |
| 运行时直接采用 pi.dev | 快，并已有缓存/版本服务 | 中低：Pi 的兼容门槛约束 Pi，实现和证据不约束 barness | 可参考、可作发现源，不直接作为 barness 可执行目录 |
| barness 维护、审核和独立发布自己的目录 | 维护完成即可发数据，无须总是发代码 | 高 | 建议讨论采用；可从外部来源生成变更草案 |
| Provider 官方模型列表直接变可执行目录 | 因 Provider 而异 | 必须另外补齐能力、价格与本库约束 | 可选发现输入，不能预设所有列表已提供完整目录语义 |

后三行是基于第 3、4 节事实的架构判断，不是已接受选型。

### 5.2 更新在哪里生效

| 路径 | 可保留边界 | 额外工作 | 研究判断 |
| --- | --- | --- | --- |
| 外部更新器/Host 下载，启动前注入 `Config.Catalog` | 现有 Client 不可变快照和公共查询 | 发布包、加载验证、缓存和选择流程 | 首版最克制；长驻服务要重装配或重启才采用新目录 |
| Host 长驻管理并把目录快照交给新调用 | 可以保持公共模型结构；需重新定义目录读取点 | 明确调用 pin、并发发布、失败与资源状态共享 | 若首版必须不停机生效，需要作为显式设计目标 |
| barness-ai 自己联网、持缓存并周期刷新 | 库内承担更多基础设施 | 网络策略、文件位置、生命周期、云端共享方式 | 不作为默认建议；会扩张基础协议模块职责 |

“更新机制不全放在 barness-ai”与用户原始目标一致，也与 Pi 把远程包装放在 coding-agent 层的做法相容；但具体归属仍由讨论决定。

## 6. 可供讨论的最小维护闭环

下面是一条候选工作流，不是已批准实施计划。

| 步骤 | 主要责任 | 最小可审阅产物 |
| --- | --- | --- |
| 发现变化 | 维护工具或维护者 | 新型号/价格/限制的来源链接、抓取时间与原始来源标识；不直接写租户授权 |
| 建立 barness 定义 | 目录维护者 | 现有 Catalog 可表达的元数据、硬约束、来源与差异；保持四维身份 |
| 判断是否需要代码 | barness-ai 维护者 | 已有实现可消费，或必须先新增字段/Provider 行为/adapter 的明确结论 |
| 审核纳入与验证 | 按用户选择的支持承诺 | 目录差异、受影响路线、验证/不适用依据；不借新数据追认旧证据 |
| 生成独立发布包 | 发布工具 | 格式版本、内容版本、最低兼容实现、内容摘要、来源提交与 Catalog 负载 |
| 发布和保留 | 发布流程 | 不可变包及其来源；最后切换可变发布指针；保留可退回的旧包 |
| 检查并采用 | Host/更新器 | 检查版本与输入，完整校验后才采用；失败继续使用上次有效目录或内置基线，并给出诊断 |
| 回滚和追溯 | Host/发布维护者 | 回到明确的旧包；已发生调用继续对应它原来使用的目录版本/哈希 |

建议首版只选择一种简单的实现兼容表示，例如最低 barness-ai 版本；是否需要精细能力集合，要等出现最低版本无法合理表达的真实用例再定。签名、多级委派、跨地域发布平台等未被当前需求证明必要，不列为默认前置条件。

需要由后续决策确定的细节还包括：完整目录替换或 overlay；如何处理移除/停服；本地自定义覆盖的优先级；发布包版本与既有 `Catalog.Version` 的对应；获取到新目录和开始采用新目录的区别。这些应从根决策继续展开，不能在首轮预设答案。

## 7. 讨论入口与约束

优先确认三个独立的根决策：

1. **发布权：** 是否由 barness 自有审核目录作为唯一正式更新源，Pi/models.dev/官方文档只是维护输入？研究建议是。
2. **纳入承诺：** 同协议新型号只有兼容声明、尚无该型号自身 live 时，是否可先进入可调用目录？当前规则包含路线/型号自身证据要求；可以保持，也可以明确修订，但不能静默变更。研究不替用户选这一项。
3. **首版生效目标：** 第一阶段是否允许在启动或明确重装配时采用新目录，还是必须让长驻 Client 不停机采用？前者更容易保留现有契约；后者需要专门设计快照与资源边界。

后续问题依赖以上答案，再讨论维护仓库/包结构、兼容分发、合并删除、覆盖与缓存等。已接受决策才应进入 ADR，已确认概念才进入根 `GLOSSARY.md`。当前不新增 accepted ADR，也不把本报告建议写成规范。

## 8. 研究边界与未验证事项

- 已核对当前和 v1.1.0 的 Pi 一手源码、公开官方文档、公开目录响应，以及本地 Catalog/adapter/授权边界。
- 未读取 Pi 私有网站仓库，未检验其生产 R2 库存、全部历史目录或运营回滚过程。
- Pi 发布兼容检查是源码能力边界证据，不是所有型号实际请求正确性的保证；本次没有运行 Pi 自身测试，也没有厂商 live 调用。
- 本次没有实施 barness 发布包或更新器，因此不宣称远程数据已能被当前 barness 安全采用。
- 若后续修改校验、调用快照或运行时激活，要先按本库测试 SOP 设计失败场景，再写实现；研究阶段不补写实现之后的单元测试。

## 9. 补充核查：Pi 是否只更新已配置的 Provider

补充日期：2026-10-29。仍固定 Pi 提交 `e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a`；下列行号按原始源码计数。

**结论：** 对普通内置 Provider，Pi 确实按 Provider 请求目录，而且只有能解析出凭据的 Provider 才进入联网刷新。但“只配置 OpenAI”须明确指 Anthropic 也没有环境 key、已保存凭据或其他可解析认证来源；仅把当前模型选成 OpenAI，并不会阻止已配置的 Anthropic 更新。缓存恢复的范围又比联网范围更大。

### 9.1 三层范围不能混为一谈

| 层次 | 精确行为与来源 |
| --- | --- |
| 注册与枚举 | `ModelRuntime.create()` 注册全部内置 Provider，给普通 Provider 加远程包装。`ModelsImpl.refresh()` 枚举已注册且有 `refreshModels` 的 Provider；传入 `options.providers` 时再按 ID 缩小。没有该参数时，并非先只选当前模型的 Provider。[装配 L229–240](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/model-runtime.ts#L229-L240)、[枚举 L565–574](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/models.ts#L565-L574) |
| 缓存恢复与网络过滤 | 每个候选先执行 `runProviderRefreshPhase(..., false, ...)`，恢复本地缓存；允许联网时，再调用 `resolveRefreshCredential()`。`if (!credential) return` 后才进入 `allowNetwork=true` 的阶段。所以“无凭据不联网”并不等于“无凭据连缓存也不读”。[调用链 L581–596](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/models.ts#L581-L596) |
| 远程请求分片 | 普通包装器向 `/api/models/providers/{providerId}` 发送该 Provider 的目录请求；并非下载全体 Provider 后仅在界面隐藏未配置项。[requestCatalog L75–95](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts#L75-L95) |

凭据判定的确切含义：OAuth 必须是该 Provider 支持的认证方式，过期时先刷新；非 OAuth 路径调用 `provider.auth.apiKey.resolve()`，返回空则跳过。这里的条件是“能够解析认证”，不是预先向厂商证明 key 真实有效。[resolveRefreshCredential L627–655](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/models.ts#L627-L655)

标准 API-key resolver 先读已保存 key，否则读 Provider 定义的环境变量；配置组合器还支持 `models.json` 中的 key 来源。因此，注册 Provider、具有 `models.json` 条目、当前选中 Provider、能够解析凭据，是不同条件。[envApiKeyAuth L9–29](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/auth/helpers.ts#L9-L29)、[配置认证解析 L462–478](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/provider-composer.ts#L462-L478)

### 9.2 何时触发

| 入口 | 已核实行为 |
| --- | --- |
| 创建 ModelRuntime | 默认恢复缓存；只有显式 `allowModelNetwork=true` 且不离线时，创建阶段才允许网络。[create L250–264](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/model-runtime.ts#L250-L264) |
| 交互模式启动 | TUI 初始化后发起一次后台刷新，15 秒总时限。[InteractiveMode.run L1171–1180](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/modes/interactive/interactive-mode.ts#L1171-L1180) |
| RPC 启动 | 非离线时启动后台刷新，15 秒总时限。[main L951–958](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/main.ts#L951-L958) |
| 打开模型选择器等交互入口 | 先展示已有快照，再调用共用刷新协调器；并非整个进程只检查一次。[ModelSelector 构造及刷新 L158–196](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/modes/interactive/components/model-selector.ts#L158-L196) |
| `pi update --models` | 先以不联网方式创建 runtime，然后调用 `refresh({allowNetwork:true, force:true})`；仍受上述 Provider 凭据过滤。[refreshModelCatalogs L607–621](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/package-manager-cli.ts#L607-L621) |

这些调用入口通常不传 `providers`，所以尝试范围是全部可刷新 Provider；实际普通 HTTP 还会被凭据判定、离线策略和四小时新鲜窗口缩小。四小时是刷新函数内的检查窗口，不应解释为固定每四小时启动一次全量下载的后台定时器。[新鲜窗口判断 L130–145](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts#L130-L145)

### 9.3 OpenRouter 是用户凭据个性化的例外

OpenRouter 采用 `withPersonalizedCatalog`：先取得其 pi.dev 分片，再按用户凭据请求 OpenRouter `/models/user`，用返回集合缩小聊天型号并调整对应限制和价格；图像、分类模型原样保留。个性化结果被保存并用于离线读取。该路径跳过普通包装器的四小时/ETag 重验证政策，但不绕过 `ModelsImpl.refresh()` 的凭据过滤。[个性化包装 L187–234](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/coding-agent/src/core/remote-catalog-provider.ts#L187-L234)、[personalizeOpenRouterCatalog L16–61](https://github.com/earendil-works/pi/blob/e65b3b88c08412bb8ffbd21fb5dc5756064a6e5a/packages/ai/src/providers/openrouter.ts#L16-L61)

### 9.4 与 barness 已确认方向的关系

本轮讨论已经选择：barness 自有审核和独立发布；按所配置 Provider 更新；安装包含默认目录；首版启动检查一次，不做进程内刷新；只维护目录内容版本，不引入此前研究建议的兼容代次。原第 6 节的最低实现版本建议属于研究阶段备选，已不作为本轮设计要求。新协议或新增必需行为仍在原需求的代码适配范围内。

Pi 的事实支持“按 Provider 分片、未配置者不联网”的方向；它的默认选择依据是认证可解析，且存在进程内后续刷新。barness 是否采用 Host 显式 Provider 列表、凭据发现或两者交集，以及每 Provider 内容版本和聚合 `Catalog.Version` 的算法，仍需讨论，不能仅因 Pi 如此就视为已决定。

本次只读外部源码并追加本节，没有请求模型推理或更新本地运行目录。

## 10. 补充核查：Provider 官方 Models API 能证明什么

核查日期：2026-10-29。仅阅读 OpenAI、Anthropic、Gemini 官方文档，未使用凭据调用模型列表或推理 API。以下字段对应当日文档，不假设旧 API 版本也返回相同字段。

| Provider | 官方列表 API 的确证信息 | 账户、权限与缺失的解释边界 |
| --- | --- | --- |
| OpenAI | `GET /v1/models` 返回 `id`、`created`、`object`、`owned_by`，当前还有可选的 `shutdown_date`；没有价格、上下文和请求参数兼容定义。[List models](https://developers.openai.com/api/reference/resources/models/methods/list) | 官方 RBAC 把此能力定义为列出组织有权访问的模型；调用还要求 key 和项目角色具备 `api.model.read`。这不是无身份的全球型号清单。[RBAC](https://developers.openai.com/api/docs/guides/rbac) |
| Anthropic | 当前文档展示 API 版本 `2026-10-21`。除 ID、名称、发布时间，还返回能力集合、`max_input_tokens`、`max_tokens`、`line` 及生命周期字段；有分页和 lifecycle 过滤。不能概括为“只返回 ID”。[List Models](https://platform.claude.com/docs/en/api/models/list) | 请求可指定 Workspace；`deprecated` 仍允许已有访问资格的组织调用。默认列表只含 active/deprecated，retired 要显式请求；因此缺失可能来自分页或过滤，不能单凭缺失认定全球退役。[同一接口的筛选和生命周期说明](https://platform.claude.com/docs/en/api/models/list) |
| Gemini | `GET /v1beta/models` 分页返回 Model：名称、基础型号 ID、版本、说明、输入/输出 token 上限、支持的方法、thinking，以及 temperature/topP/topK 等采样元数据。`topK` 为空时，该参数不允许发送，是可直接采纳的一项约束；没有价格字段。[Models API](https://ai.google.dev/api/models) | 官方指南明确部分 2.5 型号仅向过去使用过它们的用户开放，同时明确尚未弃用、仍继续服务。这证明账户可访问性不等于全球生命周期；但列表接口文档未承诺返回集合一定精确等于每个账号此刻可推理的集合，不把这种过滤行为写成事实。[型号指南](https://ai.google.dev/gemini-api/docs/models) |

另一个重要区分：Anthropic 的 `retires_at` 是可修订的计划日期；即使日期已过，只要 `lifecycle` 仍为 deprecated，也不代表退役已发生，以 lifecycle 为准。OpenAI 的 `shutdown_date` 则是明确提供的停服日期信息，不能因其为 null 就反推永不下线。[Anthropic 字段定义](https://platform.claude.com/docs/en/api/models/list)、[OpenAI 字段定义](https://developers.openai.com/api/reference/resources/models/methods/retrieve)

**用于目录维护的推断：** 官方 Models API 是可用的一手事实来源，能提供的能力和生命周期信息因 Provider、API 版本而异；应保留来源时间、服务目标、版本、分页/筛选和账户作用域，按其明确字段解释。列表成功、单项缺失、认证/权限失败是不同观察，不能将后三者合成为“厂商全球停服”。显式生命周期字段可作为进一步维护依据，不应退化为只比较 ID 的增删。

**不能从 ID 补造的内容：** 三家上述返回结构均未提供完整价格定义。ID 本身也不编码完整硬约束；Anthropic 甚至明确要求不能从 ID 推断 `line`。已返回的 limits/capabilities/参数限制可以作为依据，其余费率、必需请求行为和本库已实现范围仍需官方定价/协议资料及 barness 的维护判断。此结论不否认官方 API 已提供的真实能力字段。[Anthropic 字段定义](https://platform.claude.com/docs/en/api/models/list)、[Gemini Model 定义](https://ai.google.dev/api/models)、[OpenAI Model 定义](https://developers.openai.com/api/reference/resources/models/methods/list)
