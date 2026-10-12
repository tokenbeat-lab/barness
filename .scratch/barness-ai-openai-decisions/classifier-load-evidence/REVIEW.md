# 工单 08 两轴 code-review

固定点 c4cc495ddba6c7fbe625ccaede74989d4341fe88；完整 diff
`git diff c4cc495ddba6c7fbe625ccaede74989d4341fe88...HEAD`。
独立并行 Standards/Spec 审查提交 fd1b05a，修复后复审 fd1b05a + 9c3e0ff，最终复审再包含 0856b2b。

## Standards

首次发现两项 P2：

- `map[string]int64` 把 JSON null gauge 当零，违反 AGENTS.md 边界验证与 ADR-0017 实际资源读数要求。
- 结果仅核对任一已知租户和一个 Attempt，云端八份复制单个租户结果仍可能导入；缺完整租户与调用归属证明。

判断性 Duplicated Code：原报告与新增报告重复占用率及零资源验证。

修复：资源读数采用非空指针，拒绝缺项/null；验证独立 RequestID、每租户四个结果、
Binding/目录版本/hash、完整 Attempt 和 Observer 终态、q0..q7 完整题目键。
共用 validLoadRatio/loadResourcesReleased 消除重复规则。

最终独立只读复审：明确规范问题 0；判断性代码异味 0。

## Spec

首次发现 P1：内存与 TotalAlloc 全改为 1，仍可能通过，虽然 Provider 捕获证明已经复制数 MiB。
首次发现 P2：重复结果、空 Attempt、错答案键未被拒绝，与工单“缺失或矛盾失败”及 R19 不符。

修复：TotalAlloc 与 EOF 可达堆不能小于实际完整请求副本 + Client 完整 unary buffer 的已知字节下界；
保留结果的 heap 至少覆盖 Provider 捕获副本。结果完整归属与题目集合按上述规则验证。
新增 actual artifact 变异都先保留 RED，再修复为 GREEN；变异包重算 manifest 哈希，证明语义核验。

最终独立只读复审：剩余 Spec 问题 0，无 scope creep。原四路线和新 bounded batch 设计符合批准工单；
新三批共用一个 Client；没有要求旧四路线批次必须共用该 Client。十组合最终门禁由工单 13 接线。

最终摘要：Standards 0、Spec 0；全部审查缺口及判断性异味已关闭。

追加自查（0856b2b）：同一 MaxQuestionBytes 同时约束 public/native，后者多 17 字节
name/type。公共捕获断言先 RED，随后报告和导入改为实际 native 4050 字节并核对 RawMessage。
两轴再次独立只读复审均 0；负载、限额与预算未改变。
