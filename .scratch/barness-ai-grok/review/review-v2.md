# Grok 修复复核 v2

原始两轴审查：[v1](review-v1.md)。修复源码 pin：`e3ad302d5bffc234a7ad3563235bffdd95d4d819`；代码提交：`51c95922`；集成合并：`5561d76dcd1a9ccee578835a10151ea1145e8d91`。修复定向普通回执：[原件](../../../.evidence/barness-ai-grok/e2e/20261026T030721.568691Z-pzit2yuv/receipt.json)；605测试事件、564 cases（含既有ID回归）实际PASS，exit0，不替代最终冻结来源完整普通验收。

## Standards

Reviewed the fix delta `a3c3a269d15c35563d2ec3c364595a7f179e0849...e3ad302d5bffc234a7ad3563235bffdd95d4d819` read-only.

- **Prior P1 — resolved.** `ai/responses_xai_input.go` now maps absent/null shorthand discriminators and explicit `message` to the same native-ID authorization key. Foreign assistant IDs are checked against the trusted baseline; plain shorthand and an already-authorized ID remain permitted. The added public Hook scenarios cover both omitted and null types, denial before HTTP, and legitimate cases. This restores `ai/README.md` invariant 9 and ADR-0005's native-reference boundary.

- **Prior P2 — resolved.** `ai/live/xai_finalize_test.go` preserves the original non-PASS report, writes/syncs staged content, and audits it before promotion. Later directory-sync or final-audit failures invalidate canonical and pending reports (`FAIL`, `Complete=false`, `AuditPassed=false`); failed rewrites attempt to unpublish the file and the process remains failed. Public process scenarios exercise an actually closed directory descriptor and an actual audit-write EISDIR, retain consumed grant/budget evidence and the original report, and check formal-consumer refusal. The documented residual limitation is unavoidable filesystem failure preventing even corrective writes/removal; it does not produce a successful process exit.

**New material standards concerns: none found.** The additional xAI-only header guard follows canonicalization without changing other Providers; duplicate-key checking is confined to protocol control objects; cross-Provider opaque-ID handling reuses the existing collision allocator. The new code remains within the file-size limit and adds no dependency, compatibility shim, generic registry or cross-Provider codec coupling. No additional actionable baseline smell identified.

Source review and retained evidence pointers only; no tests, live requests, race or pressure executed. Final frozen full ordinary verification remains pending; targeted receipts do not establish current full acceptance or release qualification.


## Spec

复核范围：`a3c3a269...e3ad302d5bffc234a7ad3563235bffdd95d4d819`，按指定 pin 读取；工作树较新的提交未混入判断。

- **已解决：备用缓存路由头。** `ai/hooks_run.go:70–74` 在 header 名归一化后，仅对 xAI 拒绝任何实际携带值的 `X-Grok-Conv-Id`，包括空字符串、多值与大小写变体；nil/空值集合不产生传输值。新增公共 E2E 同时覆盖零 HTTP 拒绝及旧 Provider 自定义头语义。符合 R16/R19 和设计第316行。

- **已解决：重复协议控制键。** `ai/responses_xai_json.go:13–84` 在 map 折叠之前检查请求、reasoning、工具、输入 item/part 与 text.format，识别转义后的同名键。`ai/responses_xai_request.go:46–64` 保留授权检查优先级，并按来源给出 invalid_request 或 callback_failed。业务 schema、参数/结果字符串和未知 native 扩展内部未被递归误判，也不重建原生数字或密文。新增公共场景覆盖四种可达来源、授权扩大和业务 JSON。符合 R07/R08/R19。

- **已解决：向旧 Responses 降级的工具关联碰撞。** `ai/opaque_history_ids.go:146–147` 将 xAI Responses 纳入既有完整来源 ID 的碰撞分配；调用与结果沿原重命名映射同步投影。新增 E2E 覆盖并行、跨轮和占用旧 ID，检查关联唯一、可见内容保留、私有状态移除与输入不变。符合 R12/R13/E18。

新增 shorthand 原生授权及候选报告最终落盘修复未发现新的规格问题。前者统一受担保的 message ID 识别；后者先保存原非PASS报告，再暂存/审计/晋升，晚期失败将 canonical/pending 作废并返回失败，新增子进程场景保留真实 I/O 故障和正式消费者拒绝证据。

本次为只读静态复核，未执行测试。剩余规格发现：0；范围扩张：0。完整冻结普通回执、真实 live、正式晋升和人工 race/压力仍分别待完成，不能据此宣称发布通过。


Standards：原2项已解决，剩余0；Spec：原3项已解决，剩余0。未宣称真实资格、人工阶段或正式发布通过。
