# 工单 13 实现复审

固定基线：`3b8d05be49c248529880e15e90e02195e56c5356`。
最终代码复审：`git diff 3b8d05be49c248529880e15e90e02195e56c5356...807ccba`。
两名独立审查 agent 按 `code-review` skill 分别审查 Standards / Spec。

初审的三项 P2 已修复：撤销证据空 JSON 可绕过、live inventory 删 Hash 项可绕过、
撤销诊断缺少定位上下文。先运行 `TestReleaseRevocationSemanticIntegrity` 与
`TestReleaseLiveInventoryCannotOmitProof` 获得 RED，再修改导入器并取得 GREEN；
完整原包的正例也通过。未以重写测试预期代替修复。

## Standards

Hard violations: **0 outstanding**. The original diagnostic finding is fixed:
revocation failures identify the internally generated case, artifact filename,
and validation rule without echoing artifact contents or decoder errors. The
new E2E mutations explicitly verify those diagnostics after passing inventory checks.

Judgment-call smell findings: **0**. Sharing `verifyArtifactInventory` between
offline and live imports follows AGENTS.md principle 7’s preference for reusing
established semantics; the consumers retain their independent validation responsibilities.

## Spec

**0 项剩余发现。**

- 撤销导入解析并交叉核对 Binding、请求捕获、结果、Attempt 和 Observer；空 JSON、
  错误身份、费用缺失及撤销配置不一致会失败。
- live/offline 共用完整 artifact inventory 校验，删除 manifest Hash 条目不能绕过验证。
- 新增 E2E 覆盖上述反例。迁移测试改为验证保留的迁移快照，符合保留历史且允许当前矩阵更新的要求。

复审时完整 gate 与打包尚未完成，未把这些待办计作代码缺陷；实际交付结论独立见
`release-report.json` / `verification.json`，不能由审查通过推断。

Standards 0；Spec 0；两轴均无剩余最高严重性问题。

## 交付阻断复核

历史关联核查发现 thinking-sampling 工单 07 仍为 `ready-for-agent`，原票明确规定后续
normal/race PASS 不等于根因确认或修复。Spec 审查进一步确认：保留实测发布报告原样，
另存 `delivery-verdict.json` 的最终 FAIL、原始失败证据链接，并保持工单 13 claimed /
ADR-0026 proposed，符合本票“保留拒绝 verdict、本票未完成”的规定；不能宣布验收完成。
