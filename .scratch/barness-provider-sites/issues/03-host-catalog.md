# 03 Host 装配与独立候选目录

Status: resolved
Type: task
Blocked by: none

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R02 R03 R04 R07 R08 R09 R15 R16 R19。

## Ownership

ai/examples/glmhost、ai/examples/minimaxhost、各调用者迁移；独立候选与官方来源/价格记录；公共 Host/Catalog E2E。未获自身 live 证明不得晋升 BuiltinCatalog。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。


## Answer

Host / 候选实现完成；功能当前状态仍唯一见 [spec](../spec.md#当前状态)。

- `glmhost.Bindings` 改为显式 Authorization，`minimaxhost.Authorization` 必填 Provider；旧调用者直接迁移。四服务三协议各自固定 BindingID、官方 base、账户、用途与 credential ref。官方 URL 校验拒绝串站、协议路径、userinfo、query、fragment；显式可信网关独立配置，不能借网关标志绕过已知官方服务检查（含大小写/末尾点变体）。
- 两个 Host 的 `Credentials` 示例在调用 ReadSecret 前核对独立 CredentialRecord 的 Provider、Tenant、AccountScope、Purpose、AllowedAPIs、Active 及版本；归属/用途错配零秘密读取，版本冲突不重新取 key。同服务同账户同用途的三协议可显式共享记录，跨服务记录拒绝。
- [独立候选与来源说明](../../../ai/internal/testkit/regional/README.md)：Z.AI 六身份与 MiniMax CN 三身份为各自完整字面数据，不晋升 BuiltinCatalog。保存十二份官方原件的 lossless gzip、原始/压缩 SHA256、URL、读取时间和固定 snapshot pin。Z.AI USD/M1.4/4.4/0.26；CN MiniMax 自身 CNY 标准双档按独立抓取的 CFETS2026-10-16 USD/CNY6.7220 换算。缺失的 cache-write 价格明确为 null；与票01合并后，真实报告正写入保留 usage 并 protocol/stream 失败，不认证免费。CN2048 为来源冲突下的保守目录支持上限，独立记录理由、协议缺项及自身限额证明后的修订条件；测试调用预算另为64，不冒充厂商实测上限。
- 公共 Client / Host / Catalog E2E 覆盖四服务全部十八完整身份、四入口、Hooks、实际路由/归属、独立目录快照、凭据元数据拒绝、缺价写入、撤销与删除新增身份后的回滚。所有流为明确标注的合成受控脚本。

实际命令与原件保留于主 checkout 的 `.evidence/provider-sites/host/`，未搬进测试 worktree：

- 原始行为 RED：[显式服务](../../../.evidence/provider-sites/host/red-explicit/receipt.json)、[自身价格数据消费的未知写入](../../../.evidence/provider-sites/host/red-unknown-writes/receipt.json)、[官方网关绕过](../../../.evidence/provider-sites/host/red-trusted-official/receipt.json)、[主机变体](../../../.evidence/provider-sites/host/red-official-host-variants/receipt.json)、[空 fragment](../../../.evidence/provider-sites/host/red-empty-fragment/receipt.json) 均实际退出1；接口尚不存在的编译 RED 也原样保留，不冒称运行了 case。
- 整合票01后，`go test ./ai/e2e -count=1 -json -run '^Test(GLMHost|MiniMaxHost|RegionalProvider)'`：[回执](../../../.evidence/provider-sites/host/final-host-regression/receipt.json) 退出0，484实际测试节点，467原生 case 全PASS；[原生 manifest](../../../.evidence/provider-sites/host/final-host-regression/bundle/20261020T014712.494926000Z/manifest.json)、audit findings为空。回执旁的 source.json 保存实际完整 Go/模块/候选与嵌入原件指纹；实际 argv、起止 UTC 时间、执行来源和 case 清单均在回执。
- `go vet ./ai/examples/glmhost ./ai/examples/minimaxhost ./ai/examples/internal/hostconfig ./ai/internal/testkit/regional`：[回执](../../../.evidence/provider-sites/host/final-vet-host/receipt.json) 退出0。文档19处本地引用存在，新增文本/代码文件均不超过500行。

`race: NOT_RUN`，自身新增 live / 付费验证 `NOT_RUN`，`release.pass: false`。九身份的 INCOMPLETE / NOT_RUN 与来源冲突仍明确保留，候选可供另行授权的真实资格验证；普通通过不授予晋升或完整发布资格。未读取真实 key 文件、未调用模型服务。


票02 runtime 整合后复验：已合并 integration `11475c697948942caf6c15ba85f5e776198c7072`，执行来源 `03ea9ca574528fccf4eaafa5fcedef6c00375361` 的同一组合命令实际退出0，626测试节点 /603原生 case 全PASS，audit findings为空。[最终 runtime 整合回执](../../../.evidence/provider-sites/host/final-runtime-integrated-host/receipt.json)、[原生 manifest](../../../.evidence/provider-sites/host/final-runtime-integrated-host/bundle/20261020T014808.432636000Z/manifest.json) 固定更新后的完整源码集合；前述旧来源的失败/通过原件全部保留，未改时间或标签。
