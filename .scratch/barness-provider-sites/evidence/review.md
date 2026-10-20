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
