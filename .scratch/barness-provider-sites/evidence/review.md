# 地区 Provider 双轴审查

审查来源：`68dec5a...683301827c66f8f866b7142840c5f405f9684944`。两个独立只读 reviewer；最终修复与复审另附于本文末尾。原始报告保留如下。

## Standards

Standards review: frozen `68dec5a...6833018`; repository unchanged.

Hard violations:

- **[P1] Consume authorization by grant identity, not filename.** `/Users/cyber/RestoX/barness/ai/live/regional_config_test.go:155–158` uses `os.OpenFile(p.grantPath+".used", ...O_EXCL...)`. Copying the identical grant to another pathname permits another process to spend it. Actual controlled reproduction: the first failed run made seven loopback attempts; the copied grant, with identical ID and SHA, made seven more, with no reservation rejection. Both `.used` markers remained. This contradicts `ai/live/REGIONAL.md:37` (failed/canceled/successful grants cannot be reused), the function’s explicit new-human-grant requirement, and `AGENTS.md` principle 5 (idempotency); `docs/agents/verification.md` requires independently scoped paid authorization. Atomically reserve immutable grant identity in one durable consumption store, and cover copy/rename and concurrent reuse. Evidence: `/tmp/barness-regional/grant-path-probe/result.json`; repeatable script: `/tmp/barness-regional/grant-path-probe.py`.

- **[P1] Reject a binding target that differs from the Host authorization before reading its secret.** `/Users/cyber/RestoX/barness/ai/examples/glmhost/credentials.go:42–45` and `/Users/cyber/RestoX/barness/ai/examples/minimaxhost/credentials.go:42–45` reconstruct authorized routes but omit `b.Endpoint == expected.Endpoint`. A legitimate Zhipu record with a resolver-returned ZAI endpoint reads the China secret and attaches it to the international request; international MiniMax credentials similarly reach the CN target. Public Client reproduction used a recording RoundTripper, with one secret read and Attempt per case, zero network calls. This violates `AGENTS.md` principle 4 and ADR-0049’s fixed Host endpoint/account/credential boundary. Compare the supplied target with the independently authorized target, including explicit gateway configuration. Reproduction and results: `/tmp/barness-regional/host-endpoint-probe.go`, `/tmp/barness-regional/host-endpoint-probe-result.json`.

Baseline-smell judgment: **Repeated Switches** at `ai/hooks_run.go:56,104,154`: the same four-provider membership condition controls callback panic recovery three times. A shared narrowly named recovery policy would prevent drift; this is a maintenance judgment, not a release blocker. Separate protocol profiles and deliberately independent credential/catalog data are consistent with the documented architecture.

Pending own live/formal qualification and manual race are permitted unfinished verification, not implementation violations. No live, paid, or race execution occurred.


## Spec

Spec review of `68dec5a...6833018` (read-only). Two requirements implemented incorrectly:

- **[P1] Reject unsuccessful native terminals before qualifying support.** `/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/regional_stream.go:73–89` accepts any nonempty Chat `finish_reason`; `error`, `length`, or `content_filter` therefore count as successful qualification terminals while `regional_capture.go:94` checks only the separately supplied Result. Reproduction: clone a fresh passing controlled `zai-chat` bundle, change a full-stream raw SSE terminal from `stop` to `error`, update that exchange’s manifest inventory hash, and reseal. Keep Result/Observer/report untouched. `regionalverify -controlled` still has no check rejecting this disagreement; the same verifier serves qualification/formal consumption. Validate legal raw terminal semantics and agreement with the consumed Result. Spec R06, line 91: “必须覆盖…失败终态”; R18, line 169: “报告的实际…必须相符…不能认证当前支持”。

- **[P1] Derive reported usage from native responses rather than trusting Result summaries.** `/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/regional_totals.go:59–65` sums Result usage, while `regional_capture.go:97` trusts Attempt.UsageReporting and `regional_stream.go:72–84` never reads usage. Reproduction: in a passing non-cancel full-stream case, remove `usage` from every raw SSE JSON frame, preserving model, content and terminal; update only exchange inventory and reseal. Producer Result/Attempt/Observer still claim reported usage, so the consumer accepts a capture that never reported it. Similarly, changing only raw token counts escapes comparison. Independently normalize native counters/presence and compare them with Result, Attempt, Observer and budget totals. Preserve legal cancellation prefixes with absent usage/terminal, plus partial usage and explicitly reported zero as distinct states. Spec R15, line 155: “保持缺少 usage、部分 usage、明确零值…的区别”; R18, line 169 forbids stale/foreign/replay proof replacing current support.

No material scope creep identified. Pending own live, promotion, race and final integration bookkeeping are expressly allowed open and are not findings. These mutation paths are static review reproductions; retained historical controlled captures currently refuse source drift, and were neither modified nor refreshed. No live, paid or race run was invoked.


Standards：2 项 P1、1 项非阻断维护判断；Spec：2 项 P1。每轴最高等级均为 P1。

## 第一次复审

修复来源 `6833018...2d89c45`；原始四项 P1 与 recovery-policy 均已有修复及定向绿例，复审尚未通过。Spec 独立复审原文如下；新取消前缀问题继续交同一 implementer。

Independent Spec re-review of `6833018...2d89c454f5bb2af01b153a41cdbdd0b67f887430`.

- **RESOLVED — original P1 raw-terminal finding.** `/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/regional_replay.go:110–121` replays completed responses through the public Client and requires successful Stop/ToolUse plus exact consumed-message agreement. Independent offline mutation `stop → error` now exits 1 with `upstream_error/stream`.

- **RESOLVED — original P1 raw-usage/output finding.** `regional_replay.go:119–122` compares actual decoded usage/reporting/content/native fields; `regional_totals.go:59–69` uses decoded usage for totals. Independent removal of raw usage and changing raw prompt tokens to 999 both exit 1. All six retained controlled native bundles independently verified successfully, including legitimate partial/explicit-zero and cancellation cases. The final parent manifest retains 66 passing raw-mutation/cancellation assertions.

- **REMAINING [P1] — malformed complete frames can be skipped before cancellation content.** `/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/regional_replay.go:130–136` uses `continue` for a decoded event whose JSON is invalid. This permits later bytes to justify content the real parser could never reach. Reproduction: clone the final `zai-chat` controlled bundle; prepend to `glm-5.3/cancel` a valid empty assistant-role frame, then `data: {broken}\n\n`, followed by the original `known-content` frame. Update only exchange inventory and seal. `regionalverify -controlled` exits **0**, although real consumption fails at the malformed event before emitting known-content. Stop searching beyond a malformed complete event; continue accepting valid pre-cancellation prefixes, read-ahead and genuinely incomplete trailing frames. Spec R06 line 91 requires “失败终态和取消”; R18 line 169 requires actual report/evidence agreement. Probe and unchanged-metadata clone: `/tmp/barness-regional/spec-rereview-probes/cancel-malformed-before-content.json`.

Original captures/source pins were preserved. Only offline consumers and synthetic temporary copies were used; no live/vendor/paid/race execution. Own live qualification, promotion and race remain legitimately open.


Standards 复审指出：同一 `ExpiresAt` instant 的不同 UTC offset 经 time.Time JSON 编码可能得到不同 scope hash；应规范 UTC 后再生成 semantic scope。该回合未完成最终报告（工具内容过滤），未计为复审通过；由公共 E2E 红绿与后续独立复审继续验证。
