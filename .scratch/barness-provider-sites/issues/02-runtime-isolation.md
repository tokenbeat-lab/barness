# 02 缓存 v2 与准入隔离

Status: resolved
Type: task
Blocked by: none

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R11 R13 R14 R17。

## Ownership

ai/cache_key.go、cacheScope 生产与消费者的参数迁移；ai/admission.go、ai/call_runtime.go 及公共 runtime E2E。旧 Admission 断言迁移。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。


## Answer

2026-10-20：实现完成，开发阶段普通 E2E 通过。整体实施与发布状态仍以 [spec 当前状态](../spec.md#当前状态) 为唯一入口。

- `cacheScope` 从固定 Binding 接收实际 Provider；全部现有 Chat / Responses 消费者统一切换精确 v2 算法。使用 uint64 大端 UTF-8 字节长度前缀，不归一化、不加入 CredentialRef / Version，不保留 v1 或双发。既有 wire 的缓存及亲和键值发生一次变化，长度仍为 64 位小写十六进制；命中率影响与回滚规则沿用设计。
- `AdmissionRequest.ProviderID` 由每次 Attempt 的固定 Binding 填写。Host 账户策略可按 `(ProviderID, AccountScopeID)` 区分服务；既有租户 / 进程许可机制不变。迁移所有现有精确 DTO 断言。
- RP12 以独立 Python hashlib / struct 固定向量，经公共 Client 捕获验证字节算法、分隔符碰撞对、组合 / 分解 Unicode、大小写、同账户换 key、各协议省略规则及 Hook 防篡改。RP14 / RP15 通过屏障使四服务 × 两租户的 8 个请求同时在途；修改 Host 与调用者数据后，重试仍复用原 bytes、endpoint、key 和版本，撤销后的新调用零凭据读取；Result、事件信封、Hooks 和 Observer 保留真实归属。
- RP15 的受控 Host 将同一实际服务账户的单个许可明确跨租户共享，验证另一租户在该站被拒绝、同名账户的其他三站继续工作、取消后许可可复用。RP16 覆盖 4 服务 × 3 协议 × 7 种队列 / 解析 / 截断 / 取消 / 退避终止路径，核对单次结算、无流重放、兄弟调用可用、body / permit / waiter / event / call 归零。

执行来源：实现提交 `b2a1d91`、`1530c1a`，融合 integration `77e8796` 后使用正式常量；最终受测代码提交 `aef195c70b51c23cc2fc4b2222c9fad8c6275b5a`。本文仅补充关闭记录，不改变已测代码或输入。

实际命令、数字退出码、起止时间、执行来源 Hash、原始脚本、请求、事件、Result、Attempt、usage、Observer、资源与审计指针集中于 [工件索引](../../../.evidence/provider-sites/runtime/index.json)：

- `integrated-runtime/receipt.json`：`go test ./ai/e2e -run '^TestRegionalProvider' -count=1 -v`，退出码 0；294 个原生工件案例均 PASS（含已融合工单 01 的协议案例），包耗时 2.012s。
- `integrated-old-runtime/receipt.json`：受影响旧 options/cache/admission/isolation/GLM/MiniMax 场景，退出码 0；378 个原生工件案例均 PASS，包耗时 7.924s。精确 argv 见回执。
- 实际行为 red 保留于 `red-cache-execution` 与 `red-admission`；初始 fixture 编译 / 场景修正及未融合 profile 的失败回执原样保留，索引逐项解释。最终两个 bundle 的原生脱敏审计通过。

未读取真实秘密，未调用厂商服务；`paid: NOT_RUN`，`race: NOT_RUN`。这些开发阶段事实不授予 live、race 或发布通过；完整门禁由工单 05 维护。
