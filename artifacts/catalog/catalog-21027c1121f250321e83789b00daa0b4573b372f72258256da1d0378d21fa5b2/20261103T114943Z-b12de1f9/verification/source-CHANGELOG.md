# barness-ai Changelog

## 未发布

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
