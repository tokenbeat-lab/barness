# barness-ai Changelog

## 未发布

### 2026-11-02：Kimi K3.1 Pro 外部目录（验证进行中）

外部目录新增 `chat / moonshot` 与 `chat / moonshot-cn` 的 `kimi-k3.1-pro`，
每站分别提供 `openai-completions`、`openai-responses`、`anthropic-messages`
三条身份。目标目录共 125 条，原 119 条定义、价格和相对顺序待最终核对。
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
`ai/kimi_catalog.go` 核心型号白名单新增该 ID，运行依赖不变；完整受测源码及
输入仍待最终核对。旧基准程序的受控检查实际返回 `code_gap`、exit 1；它需升级
SDK 并重建后，再经兼容检查和 Host 显式加载才能使用新 JSON。其他程序版本未验证。
本次目前仅维护外部 JSON，尚未同步内置维护源、快照和基线；若之后同步，使用
编译内置默认目录的程序需要重建分发。

**验证：** 六条候选与应用原件已留存，最终普通回归及交付核对尚未完成。一次
旧 Kimi 定向运行实际进程 exit 0、2,256 case PASS，但运行期间 HEAD 变化使
collector 判为 `INVALID_EVIDENCE`；该结果不作为本次通过回执。live、race、压力
按最终影响另行判定；不声明新厂商资格或软件发布。实际证据和当前状态见
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
