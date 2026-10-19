# 20 — 保留视觉调用在发送前取消的真实证据

Type: task
Status: resolved

## Scope

修复 visualeval 的逐样本记录及核验。真实调用进入公共 Classify 后，发送前取消必须保留一个逻辑调用、零 HTTP Attempt、阶段对应的实测 credential 读取与归属、两条生命周期事件及零资源。有限状态由独立七行控制流表约束：Scope/Binding 未解析 read0；Jev-only Capability 未解析 read0；Credential 未解析 read0或1；Request 为 Jev-only 未解析 read0 或两路线已解析 read1；Admission 已解析 read1。不得伪造请求或把调用降为从未执行。审计规则、预算、Provider/runtime、旧报告和完整 source maps 保持原字节。另修复已发送取消分支对 JSON null/unset 的缺席型号比较；其 Unreported、零 usage、空输出及捕获体/型号证明约束不放宽。

## Failure inventory

1. 现有 writer 将空 wire.Request 写成零字节 request.json，导致独立审计拒绝。
2. 调用前取消/到期在 jev/decisions × plain/hooked 中必须全部可复现；不得依赖调度或 sleep。
3. 历史 scope-only 阶段：零 credential 读取不能被一律判断为缺失 snapshot。当时仅接纳 scope 终止，后续真实 post-pin RED 证明此方案不完整；当前合同采用 Scope 节的有限七行状态表。
4. verifier 不能一律要求已解析身份、一个 Attempt 和四条事件；也不能因此放宽已发送分支。
5. 未发送分支的伪造 wire、credential、Attempt、已解析归属或缺失/额外 Observer 记录必须拒绝，即使私有副本已重新密封。
6. honest FAIL 的完整 12 样本分母、一个逻辑调用及其余未执行状态必须可回放；require-live 必须拒绝。
7. 现有已发送成功/失败及 deadline 回放不回归；资源和零重试上限保持。
8. scope 不是唯一合法未发送状态：公共 TransformHeaders 取消必须产生 resolved + credential1 + PhaseRequest + 0 Attempt，而起始事件仍未解析、终态事件完整已解析。先取得真实 RED，再扩展此形态。
9. 历史第二阶段：scope/binding 为 0、credential 可为 0或1，而 request/admission 当时一律要求已解析/read1。此规则遗漏 Jev pin 前 Request；已由第12项和 Scope 的有限七行状态表取代。未发送 wire 始终完全为空，访问次数必须实测。
10. 不得把 admission/request 的 aborted 终态改成 scope 的 error，也不得以一律相等的起始/终态 metadata 拒绝合法已解析调用。
11. 未发送访问记录必须明确提供唯一的整数 credentialReads；缺失字段不能借 int 零值冒充实测 0，伪造次数必须拒绝。
12. Jev 在 pin 前图片检查可能返回 Capability 或 Request 取消：两者都未解析/read0，Request 仍为 aborted。消费者应接受这些由源码证明的有限形态；Decisions 不允许这两个未解析分支。对应派生包仅是私有消费者 fixture，不是假称在公共执行中观察到该精确阶段。
13. 已发送 response-read 取消可保留完整捕获体但未进入解码：1 Attempt、Unreported、零 usage、无返回型号。序列化为 JSON null 后，Nullable.IsZero 不能识别缺席状态。先以真实公开核验 RED 和受控公共 Client 在读完响应时取消的 E2E 证明，再最小修正编码缺席比较；不得接受非零 usage/伪造型号。
14. 独立源码复核指出 JSON null 也可被 map[string]int 解为默认 0。必须先保留真实消费者 null 接受 RED，并保持 bool/小数拒绝，再使用非 nil 的必填整数解码；原候选 a203 和原正确整数原件不能覆盖。

## Diagnostic provenance

父层已封存原始失败和快速 public CLI RED；独立诊断 `/tmp/barness-visual-deadline-diagnosis-spec.json` SHA256 `633e888e65d83b9afa041a3331d3cfd0c173e62171eeaf58ccbfabf69abbe0bd`。
H1（发送前 deadline）由原始 v08 的 scope terminal、0 Attempt、2 Observation、空 wire 确认；H2（发送后丢 capture）及 H3（覆盖/密封不一致）与原件不符。此票不修改原件、不授予新的付费轮或正式资格。

后续独立领域复核 `/tmp/barness-ticket20-unsent-lifecycle-domain.json` SHA256 `26eadb9664d2935a94f6a59f444a277f79105dd903f0261bba46ca2334aa9cf5` 指出已解析但未发送也是合法状态。最初 scope-only GREEN 仅证明该阶段，不代表完整修复；原命令、源码和原始包另存，不能覆盖。

## Test plan

先写公共 Client → 原 production writer → bundle.finish → verify 的八组合回归。仅抽取现有逐样本 block，接受外层 precheck 之后的真实 context，先保留行为 RED，再修复。受控使用真实已取消/已过期 context；transport 无网络能力。package-specific evidence 与 offline E2E package 分开，不干扰 release.FindBundle。

## Answer

离线修复源码 `d02b5adb862e98d73672b47bfd1e3142a2d6082d` 已通过 Standards / Spec 精确复核。最终受控场景普通与 race 各23 PASS、audit0，双 vet exit0；测量时是 a203 HEAD 加 WIP，1,040 项编译/JSON/module 字节与 d02 提交完全一致，不能写成提交后测试。

[证据入口](../evidence/ticket-20/README.md) 保留原始行为 RED、setup 失败、失败 GREEN、finite lifecycle 扩展、已发送取消 null/unset RED、访问计数 null 接纳 RED及最终 GREEN。三个独占创建的有界归档已整包流式核验0，原始输入前后哈希一致；没有读取 Key 或新增付费调用。此票仍待原始留存的独立双轴复核后关闭；不表示协议114与两个 visual 的当前正式资格已建立。

## Answer

离线修复已精确合入d02：不再写入虚构的空请求JSON；按七种有限控制流校验未发送调用的实测归属、凭据读取、两条生命周期事件与零资源；正确重放已发送但未解码的null型号取消包。必填credentialReads使用非nil整数，null/bool/小数及伪造次数均拒绝。原始调用仍计入完整分母，失败不授予require-live。

最终23个确定性回归普通/race均23PASS、audit0，双vet0；实际WIP1040份编译/JSON/module字节与提交d02相等。原scope/post-pin/response-read/null接受RED、失败GREEN、setup和派生夹具保留。三个完整有界归档203335成员及45实际命令收据经两轴独立留存复核0；[证据及审查](../evidence/ticket-20/README.md)。本票关闭离线修复及留存，不改写旧全量FAIL，不授予当前三组live或全量发布认证；21承接新证明/发布门禁，新增36/36尚未获授权。
