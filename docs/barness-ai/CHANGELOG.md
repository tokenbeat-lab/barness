# barness-ai Changelog

## 未发布

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
适配器、目录加载器或依赖。已具备外部目录加载能力的 Host 可分发该 JSON，重新
加载并创建新 Client，无需为 Terra 升级 SDK。使用 `ai.BuiltinCatalog()` 的程序
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
