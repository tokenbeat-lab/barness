# Pacing修复复核

冻结目标`f11b7976373859299392b1baaf083ec84d9f3ad6`，整体比较点`1889322b136f9daa8e34ba94c1aa8a9832b964f7`。真实红绿与有限425项库存见[pacing implementation](pacing-implementation-01.md)。

## Standards — PASS

复核目标：`f11b7976373859299392b1baaf083ec84d9f3ad6`；修复差异`fafd72e...f11b797`，整体比较点仍为`1889322b136f9daa8e34ba94c1aa8a9832b964f7`。首轮报告保留不变。

**已解决：** 首轮P2授权输入歧义。`regional_grant.go:84`起先逐个读取原始JSON键，再进行typed投影：`strings.EqualFold`识别大小写及Unicode简单折叠别名，要求精确`callIntervalMs`拼写；`seen`拒绝重复键（包括转义后同名）；null在映射成零值前拒绝。后续int解码拒绝非整数及溢出，`Check`保留0..60000和非Kimi非零限制。其他既有grant字段的解码语义保持，不扩大本轮修复。符合AGENTS.md原则3、4。

**时序复核：** `regional_pacing.go`现要求完整四类Observer生命周期、全部非零时间及非递减顺序；不能将`CallFinished`移到`AttemptFinished`之前来伪造下一轮间隔。初次等待继续以原始授权消费时间为锚点，后续以已核验完成时间为锚点；report/receipt/grant配置相等、进程结束上界、预算前等待及timer/signal释放路径保持。符合原则5、9。

**证据：** 只读核对公开原生进程红例021727（5项FAIL）、时序伪造红例021955（4项FAIL），以及022059（26项PASS、退出0）原件。未把定向结果当作完整冻结普通验收；后续独立回归仍由根任务收尾。新文件/改动文件均不足500行，无新增抽象、兼容shim或公共SDK行为。

**未解决：0；新硬违规：0；需报告的smell判断：0。** 本次未运行测试、读取凭据、访问厂商或修改仓库；仅写出本份离仓审查报告。


## Spec

Spec targeted re-review: PASS (0 remaining findings)

Immutable target: `f11b7976373859299392b1baaf083ec84d9f3ad6`. Reviewed replacement `git diff fafd72e...f11b797` and its interaction with the original `1889322...fafd72e` amendment. Original report remains unchanged at `spec.md`. This was a read-only review; no tests, credentials or vendor calls were used.

1. **Resolved: ambiguous interval aliases and null.** `RegionalGrant.UnmarshalJSON` now scans top-level keys before typed decoding, rejects any case-folded spelling other than exact `callIntervalMs`, rejects repeated exact keys, and rejects explicit null. `strings.EqualFold` also covers Unicode simple-fold aliases accepted by the standard JSON struct decoder. The final integer decode plus `Check` retains fractional/overflow and 0..60000/non-Kimi enforcement; omission remains zero. This satisfies R12's integer/default/range contract without altering unrelated legacy field decoding. Added public process cases cover canonical-plus-alias, exact duplicate, alias null, uppercase and Unicode alias inputs.

2. **Resolved: impossible observation chronology.** The paced consumer now requires exactly four observations with the four expected lifecycle kinds, nonzero timestamps, and nondecreasing order from CallStarted through AttemptStarted and AttemptFinished to CallFinished. Only then does it compare the first start with durable authorization consumption and subsequent starts with the previous terminal, using the granted interval. An earlier CallFinished cannot now precede its recorded AttemptFinished to manufacture a gap. The resealed `timeline-order` case directly exercises that contradiction, satisfying §10.4's timing-tamper refusal requirement.

No new scope deviation or spec mismatch found in the replacement. The bounded wait location, process/grant cancellation, zero-reservation interruption behavior, canonical scope binding, default-zero legacy path, no-retry behavior and source/binary rules remain as reviewed. This result resolves the two source findings; it does not substitute for the separately running scoped regression, frozen ordinary acceptance or authorized international candidate execution.


复核：Standards0项、Spec0项；两轴PASS。完整新冻结普通与国际站候选执行由后续原件独立判定。
