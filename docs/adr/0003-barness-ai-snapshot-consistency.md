---
status: accepted
date: 2026-10-01
---

# barness-ai 首期对无法取得一致配置快照的调用直接失败

绑定与凭据可能由不同宿主后端解析，读取之间发生更新或撤销时不能拼装身份不一致的请求。原研究允许发 Provider 请求前重新解析或失败；barness-ai 首期选择直接失败，不在同一逻辑调用内自动重新解析，以保持一次调用的快照、截止时间和观测清晰，并避免与 Provider 网络重试叠加。

## Considered Options

- 有界重新解析：可能提高轮换期间可用性，但需另行规定次数、退避、截止时间、撤销优先级及配置读取的观测，且不能沿用 Provider 的重试计数。
- 直接失败：语义简单、没有隐藏的配置重试放大；接受短暂前置失败的代价，首期采用。

## Consequences

- 仅在无法证明租户、账户、凭据引用和版本一致时拒绝；后端若直接取得一致的新快照，可以继续调用，不能将所有并发更新都视作冲突。
- 拒绝发生在任何 Provider 推理请求前，通过错误阶段和脱敏上下文说明一致性校验失败，不输出 key、秘密引用内部细节或伪造 HTTP Attempt。
- 轮换/更新竞争可能导致短暂失败。宿主可在自己的截止时间和预算内以新 RequestID 再次发起逻辑调用，不能沿用原 RequestID 假装是已经固定快照的网络重试。
- 已取得一致快照并进入请求尝试的调用及其网络重试继续使用原快照；新调用重新解析。此决策不承诺撤回已发出的请求。
- 若运行证据表明前置失败显著影响可用性，可通过后续 ADR 增加有界重解析；需先定义上述策略并用配置竞争场景验收。内部重构不保留双路径兼容层。

行为与验收见 [spec 的快照契约](../../.scratch/barness-ai/spec.md#3-租户授权服务绑定与配置快照) 和 E07。此决策是本规范对原研究开放选项的收窄，不表示原研究已验证其可用性代价。

## 工单 07：解析边界上的独立快照（2026-10-08）

BindingResolver 交回绑定后，立即复制其型号/托管工具切片和 Retry 的可变延迟值；
CredentialResolver 接收另一份独立副本。不能把复制延迟到凭据解析完成，否则宿主轮换源配置，
或凭据后端修改入参，就能改变已经解析的绑定预算。交接期间宿主仍不得并发修改同一源对象。
此修复共同适用于聊天和 unary，不增加重解析或配置兼容路径。

解析开始前和返回后都检查 context；即使后端成功返回，也不能把已结束的调用标成已解析，
不能继续读取下一项配置。绑定/凭据阶段保留原分类和未解析归属，不发 Attempt。
先红后绿证据：TestUnaryResolutionInterruption、TestUnaryBindingPolicySnapshot，见
[工单 07](../../.scratch/barness-ai-pi-1.0/unary-failures-evidence/README.md)。

## 工单 15：三操作共用 Client（2026-10-08）

E07-mixed-snapshot 对 chat、OpenAI image、Google image、TypeSafe 分别验证操作/Provider/
端点/账户/凭据引用的前置更新竞争：冲突只读一次绑定和凭据，无 Provider 请求；已发送
调用的重试固定原绑定、key、目录及重试策略，新调用才采用新快照，撤销只拒绝新调用。
各操作同名 Binding/同型号 ID 依然按租户与四维型号身份授权；同一 Provider/账户下
不同操作 Binding 合法共享凭据引用。构造输入的目录/子策略、调用输入/选项的可变
结构独立复制。不可变本地 OpenOperations 明确要求重新装配以更换配置/凭据；云端
仍由宿主 resolver 更新。停用 TypeSafe/OpenAI image/Google image 不承诺撤回已发送请求，
保留原尝试消耗，其他操作可继续使用。未引入跨路线 fallback 或双路径兼容实现。

[混合操作证据](../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/README.md)。

## Decisions 运行时与撤销补证据（2026-10-12）

工单 06 用共享 Client 的 Jev/Luna/chat/image、跨 Tenant 同名 Binding/Model 及
共享 Account Scope/key 验证同一快照契约。实际收到请求的 barrier 后才修改源
Catalog、Binding、Credential、图片/问题和回调引用；重试固定全部发送字节、
header、key 与价格快照。CredentialResolver 修改其 Binding 入参不能改变调用副本。

禁用 Luna Binding、移除型号白名单和撤销图片权限分别拒绝新调用；已 pin 请求及
重试仍完成，已报告消耗保留，其他路线和 Tenant 仍可调用。图片撤销不撤销文本；
恢复 Host 授权由新调用读取，库内没有 fallback。此补证据不更改 ADR-0003 的撤销边界。
见[公共契约](../barness-ai/decisions-runtime.md)及[可重放证据](../../.scratch/barness-ai-openai-decisions/runtime-evidence/README.md)。
