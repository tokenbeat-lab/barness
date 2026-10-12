# 工单 04 两轴 code-review

固定基线 `17f2f070fa63678c310b75191dbb2ae012091468`。
审查命令：`git diff 17f2f070fa63678c310b75191dbb2ae012091468...HEAD`。
实施提交 `49f119c`，审查修复提交 `1e01c9a`。两个独立审查 agent 并行、只读。
范围为本票源码、公共 E2E、用量注释、契约、ADR 和失败清单；不包含起点已有的其他工作区改动。

## Standards

首次：0 documented breaches、1 possible smell、0 bugs。Mysterious Name：
`counts[0/1/2/5].reported` 依赖前面数组的隐含次序；属于低级别可维护性判断，不是硬性标准违反。
已将固定六项数组与读取位置一起改为具名索引，没有引入共享协议 parser。

复审：Mysterious Name 已解决。新增生命周期 E2E 通过公共调用验证重试、用量、Attempt、
Observer、可信 callback cause 及资源释放，没有新的实质 Standards 问题。
最终：0 documented breaches、0 unresolved smells、0 actionable bugs。

## Spec

首次：1 finding，OD41 在 Decisions 自身缺 read-failure / response-hook-error 的明确测试。
已补 malformed JSON、坏答案、有效 refusal、读取中断和响应回调失败，Binding 显式允许两次重试；
核对一 Attempt、用量完整性、终态、callback 次数/cause、Observer 和 body/permit 释放。

复审：OD41 缺口已解决；无新的实质 Spec finding，最终 0。
审查者纠正首次报告中“原用量测试默认零重试”的说法：newDecisionsWorld 已显式允许一次重试。
原用量转换、presence、checked sum、一致性、输入费率、阈值、坏答案用量保留和价格快照均符合本票。
正式目录、区域 premium 和真实账单继续按后续票交付。

最终计数：Standards 0，Spec 0。证据包和全套验证收据见 README 与 manifest。
