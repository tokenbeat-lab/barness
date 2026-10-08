# Code review：工单 16

固定点为任务开始提交 `0d240c1a4ece0dd692a607e883c74be72d89eff8`。
初次审查实现提交 `9899f9b`，两轴由独立并行 agent 按 code-review 技能执行。

## Standards

- [P2] 导入尝试只重算汇总，未验证 token/cost 范围、尝试 ID 或 HTTP 成功语义。
  实际副本将首尝试 input=532 改为 −532 并同步改汇总后仍 COMPLETE。
  AGENTS.md 原则 3/9：边界校验与可验证性。
- [P2] 删除 observations-evaluation.json 或将 manifest.json 改成 {} 仍可通过。
  AGENTS.md 原则 9、ADR-0024：脱敏审计不能证明缺失文件与追溯完整。
- [P2] 命令/启动器泛化所有失败信息，丢失安全诊断上下文。AGENTS.md 原则 9。
- 判断性建议：样本状态字符串在三个不同职责中消费，可考虑小状态类型。
  当前固定任务保留明确分支及状态 E2E，未为单一任务扩展新的状态框架。

## Spec

- [P2] 首次调用前取消时，calls=0/attempts=0，却报告 FAIL、accuracy=0。
  工单要求“未运行报告 NOT_RUN”；取消之前没有执行不能虚构零分。
- 其他工单要求均已覆盖；fixture 与真实效果明确区分。

## 修正与验证

在既有报告及 CLI E2E 边界补红断言，验证上述实际失败后修正：
检查尝试身份、完整性、非负计数、token 总和、固定目录输入费率与成本、成功 HTTP/vendor ID；
核对 manifest 的版本/日期/提交和每一次 CallStarted/CallFinished、AttemptStarted/AttemptFinished；
命令输出固定 stage 诊断，启动器保存脱敏构建诊断；首次调用前取消保留原因并 NOT_RUN/null。
实际原包通过新核验，没有追加或舍弃真实调用。修正 E2E 37 个案例全部 PASS。
