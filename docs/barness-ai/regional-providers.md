# GLM 与 MiniMax 地区服务

实现与验收的唯一当前状态见[地区 Provider 规格](../../.scratch/barness-provider-sites/spec.md#当前状态)。新增身份先作为候选接入；普通 E2E、共享 codec 和已有另一市场的支持不能替代自身真实资格。

| 服务 | ProviderID | 首批型号 |
| --- | --- | --- |
| 智谱中国站 | `zhipu` / `ProviderZhipu` | `glm-5.3`、`glm-5.2` |
| Z.AI 国际站 | `zai` / `ProviderZAI` | `glm-5.3`、`glm-5.2` |
| MiniMax 国际站 | `minimax` / `ProviderMiniMax` | `MiniMax-M3` |
| MiniMax 中国站 | `minimax-cn` / `ProviderMiniMaxCN` | `MiniMax-M3` |

每个服务分别使用 Chat Completions、Responses、Messages。型号身份仍为 `(Operation, Provider, API, ModelID)`，wire 型号名不带地区后缀。目录、能力、价格、账户授权和原生来源分别核对；新增服务没有扩大图像或分类操作授权。

## Host 装配

[GLM 示例](../../ai/examples/glmhost/README.md)和[MiniMax 示例](../../ai/examples/minimaxhost/README.zh-CN.md)的 `Authorization.ProviderID` 必须显式提供；直接迁移内部调用者，没有旧签名 shim。原有 `zhipu`、`minimax` 的语义、Binding 与历史身份保持原义。

同租户的不同服务创建不同 BindingID 和 CredentialRef。凭据元数据记录 Provider、租户、账户、用途、协议放行与版本，Resolver 校对后才调用秘密读取函数。同服务、同账户、同用途的已授权协议可以共用凭据记录；标准 API 不授予 Coding Plan / M Plan 权益。新中国站 M Plan 不从国际站候选继承。

官方预设按服务和协议核对 URL，拒绝串站、错路径和携带凭据的地址。网关须使用显式可信配置，受控 HTTP 回环还需 `AllowLoopbackHTTP`。核心继续信任 Host 的 endpoint/Transport，不通过域名、语言或地域猜测 Provider，也不因认证、限流、余额或网络失败换站、换 Key、换型号。

[独立候选数据及原始来源](../../ai/internal/testkit/regional/README.md)保留新增九个完整身份、价格与能力缺项。人民币费率使用固定日期和来源的汇率换算为 USD/M，不在调用中联网换汇。候选不自动写入 `BuiltinCatalog`；Host 自定义目录仍须遵守 Provider 协议能力和快照校验。缺少缓存写入价格不表示免费，正写入量会保留 usage 并报告缺价失败。估算不代表套餐余额或最终扣费。

## 缓存与账户准入迁移

所有既有缓存派生消费者统一使用 `barness-ai/prompt-cache-key/v2`：版本域后依次写入 Tenant、Provider、AccountScope、Session 的 UTF-8 字节长度（uint64 大端）和原始字节，再 SHA-256，输出 64 位小写十六进制。不做大小写或 Unicode 归一化，CredentialRef/Version 不参与。发送字段、retention=none 和 SessionID 省略条件沿用协议能力。

升级和整体回滚各可能造成一次缓存命中率变化；不双发 v1/v2，也不保留 v1 回退。持久化封套没有新增 Site 字段。跨服务可继续可见对话与工具配对，签名等原生状态按现有政策降级；同服务同账户轮换 Key 不导致原生来源失效。

`AdmissionRequest` 新增真实 `ProviderID`，由固定 Binding 填写。迁移位置式 struct literal、固定字段断言和账户聚合键；`Admit` 签名保持原义。例如账户键可使用 `struct { Provider ai.ProviderID; Account string }`。同一真实服务账户是否跨租户共享额度由 Host 明确决定，库内租户与进程限制保持原义。

## 撤销与回滚

1. 禁用新增服务 Binding，拒绝该站新调用。已固定调用及其有界初始重试仍按原快照结束。
2. 从旧进程将加载的 Catalog 移除全部新增身份，再回滚代码及目录版本。旧 Host 不加载新身份自定义目录，不依赖未知 Provider 默认分支拒绝。
3. 保留新服务的历史、失败、原始捕获和消费账本。其他站保持独立授权；禁止映射到兄弟 Provider。

## 可重复验证

在仓库根目录以新的绝对输出目录执行普通公共 Client / Host 场景：

```sh
BARNESS_AI_EVIDENCE_DIR=/absolute/path/to/new-regional-evidence \
go test ./ai/e2e -run '^TestRegionalProvider' -count=1
```

原生 bundle 保留请求、脚本、事件、Result、Attempt、usage、降级、资源归零、断言和 Hash，使用 manifest 的 replay 命令复现。实际执行结果与原件入口见[规格当前状态](../../.scratch/barness-provider-sites/spec.md#当前状态)。

发布按 [ADR-0051](../adr/0051-scoped-verification-receipts.md) 核验功能五项回执；受影响服务及业务效果仍需自身证明，源码或目录变化后由原生消费者判断旧证据适用性。pi 仅作架构参考。付费资格、范围与预算，以及实际 race 范围依[测试流程 SOP](../agents/verification.md)单独授权。
