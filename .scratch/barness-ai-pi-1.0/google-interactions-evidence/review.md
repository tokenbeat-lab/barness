# 工单 12 双轴代码审查

基点 `d5d69a48abbdabed69d9154f74489e5a58fe3d4d`，实现提交
`7bfd5d7`。两个独立审查者按 code-review 技能分别核对 AGENTS.md/领域 ADR
与工单 12/规格/实施设计；各发现 1 项 P2，先补公共 E2E 红灯再修复，并独立复核。

## Standards

**[P2] JSON 失败保留已解析用量**：原响应解码在尾部损坏/读取中断时提前返回，
丢失前缀合法用量和诊断 ID/型号，违反 ADR-0022 与 AGENTS.md 第 9 条。
已把登记移到终止错误映射之前，保持错误优先级及 Content 整体为空；新增 malformed
steps/trailing JSON fixture 与 read-abort 的结果、attempt、Observer 断言。
审查者复核：已修复，无遗留项；没有其他工程规范违规或需要处理的代码气味。

## Spec

**[P2] 缺思考总量误拒 partial**：模态明细包含思考时，省略 total_thought_tokens
会导致错误的合计一致性拒绝，违反工单“缺项为 partial”。已将包含关系核验限定为
输出/思考总量均报告，缺项保留已知明细与计价；新增 missing-thought fixture。
审查者复核：已修复，无遗留项；未发现范围扩张。

Standards：1 项 P2 已修复、0 项遗留；Spec：1 项 P2 已修复、0 项遗留。
