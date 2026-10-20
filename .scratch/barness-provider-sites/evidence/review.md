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


## 最终独立复审

两轴独立复审固定 `83a7940290176b2889482fd7f58d67be07ccc0d1`；最终普通受测 `d4b63d9` 与该来源的差别仅为下述三个旧 golden 与 GLM 断言迁移，另经两轴独立补审。原四项 P1、追加两项 P1 及 maintenance 判断全部修复；各轴无剩余可操作发现。没有将未运行的真实资格或 race 计为通过。

### Standards 最终报告

Standards follow-up review: `83a7940290176b2889482fd7f58d67be07ccc0d1`, read-only. **No remaining concrete correctness finding in the reviewed fixes.**

- **Resolved — prior P1 grant reuse, including equivalent expiry offsets.** `ai/internal/testkit/supportmatrix/regional_reservation.go:52–59` clears ID and converts `ExpiresAt` to UTC before marshaling the semantic scope. Equal instants therefore have one timestamp representation. Reservation and consumption verification use that same key derivation. The trusted Git common-directory store reserves scope and ID under an exclusive nonblocking lock; exclusive creation, file/parent-directory synchronization, retained partial records, and exact authoritative receipt bytes preserve the local fail-closed behavior (`:65–158`; `regional_files.go:183`). Retained final ordinary evidence shows copied, relabeled, concurrent and offset variants each allow one reservation and reject the other before sending; partial records reject both with zero requests. The offset case retained its earlier RED before the fix.

- **Resolved — prior P1 credential target mismatch.** `ai/examples/glmhost/credentials.go:43` and `ai/examples/minimaxhost/credentials.go:43` compare the supplied endpoint against the independently authorized route before `ReadSecret`. Explicit gateways remain valid through `Authorization.Endpoints` and `TrustedGateway`. Parent evidence records both wrong-market cases passing with zero secret reads and successful legitimate Host routes.

- **Resolved — maintenance judgment.** `ai/hooks_run.go:182–184` centralizes the exact existing four-provider recovery policy. All three callback boundaries use it; other providers retain their prior behavior.

Independently compared traceability objects: all 394 legacy mappings are unchanged; the 40 regional items retain every old required case and change only additive `requiredCases`.

Evidence inspected: `.evidence/provider-sites/review-fixes/index.json` and `followup-1/index.json`, native case summaries and source at the reviewed commit. No new process experiment, network, credential, live, paid or race execution was performed for this follow-up.

The trusted Git common directory is one approval domain. Independent repositories/hosts require the Host’s shared authority; this review does not claim distributed approval enforcement or complete release qualification.


### Spec 最终报告

Independent final Spec re-review of `83a7940290176b2889482fd7f58d67be07ccc0d1` (implementation `579b461`).

- **RESOLVED — P1 malformed cancellation prefix.** `/Users/cyber/RestoX/barness/ai/internal/testkit/supportmatrix/regional_replay.go:130–139` now stops at the first complete invalid JSON event. Independently rebuilt the offline consumer and repeated the prior exact bypass on a fresh controlled-copy bundle: valid role frame → `{broken}` event → saved content now exits **1**, “canceled Result lacks observed raw prefix”. A valid upstream-error-before-content variant also refuses. No source/capture refresh was used.

- **RESOLVED — original P1 raw terminal.** `regional_replay.go:110–121` still requires a successful decoded terminal and exact Result agreement. Independent `stop → error` mutation exits **1** with `upstream_error/stream`.

- **RESOLVED — original P1 raw usage/output/native consistency.** `regional_replay.go:119–122,147–154` and `regional_totals.go:59–69` retain real Client decoding, exact message/reporting comparison and raw-derived totals. Independent missing-usage and contradictory-token mutations both exit **1**. Retained final evidence includes 72 passing raw-content/tool/native/usage/terminal and cancellation assertions.

- **Preserved legal cases.** All six fresh controlled native bundles independently verify with exit **0**. Independent cancellation read-ahead and incomplete trailing-frame mutations also exit **0**. Existing partial-usage and explicit-zero captures remain accepted. The delta retains every prior required case and all 394 legacy mappings byte-equivalently; seven required cases are additive.

**Remaining actionable Spec findings: none in the reviewed implementation/delta.** Evidence: `/tmp/barness-regional/spec-rereview-final-probes/`, including the reproduced negative and legal-positive CLI receipts. Original captures remain unchanged. Only offline consumers and clearly synthetic temporary copies were used; no live/vendor/paid/race execution. Formal own live, promotion, race and full ordinary release gates retain their actual pending status.


### 最终测试迁移补审 — Standards

Supplemental Standards review of `83a7940...228cc11` (`9f3a437`, `228cc11`): **no actionable finding**.

Only three fixed cache-key literals and the existing public GLM import E2E changed. Independent Python SHA-256 computation over the exact v2 domain and length-prefixed tuples confirms OpenAI `73fbbf…dd9ab` and DeepSeek `feafe4…dc6be`; both OpenAI scenarios share the documented tuple.

Retained command receipts show failing old expectations before each migration and successful targeted regressions afterward. The green manifests contain 31 and 56 PASS cases. These are corrections to established E2E expectations, with no new unit tests after production code.

The derived GLM view now requires the exact importer-source refusal, independently records differing execution/importer snapshots, and retains the separate exact raw execution-source refusal. It does not broaden acceptance to arbitrary integrity errors or refresh old proof/pins.

The diff changes no production, native execution-map input, archive or proof file. Independently compared all 47 historical vector-artifact Git blobs: zero changes. The `f384861…` execution-map remains applicable to these test-only changes.

Inspected both supplied evidence indexes and retained artifacts. No test/process rerun, network, credential access, paid call or race execution occurred. Full ordinary audit remains the parent’s separate gate.


### 最终测试迁移补审 — Spec

Spec supplement: read-only review of `83a7940...228cc11a491dea3d33b1abedac1f08bab070bff1` (including `9f3a437`).

**No actionable findings.** The diff changes exactly three fixed golden cache-key literals plus GLM historical-report assertions; no production implementation, proof, pin or historical vector changes.

The OpenAI and DeepSeek literals follow R11’s exact v2 domain and big-endian length-prefixed Tenant/actual Provider/Account/Session algorithm. The independent Python vectors/census retain fixed expected outputs; surrounding wire payload assertions remain unchanged. No additional wire exception is introduced.

The GLM derived-view expectation now matches the actual earlier guard in `glm_derivation.go:139–141`: exact `GLM consumer/importer source drift`, still requiring `ErrInvalidReport`. Tests independently compare current/captured importer and execution inventories, preserve the immutable matrix and reject refreshing the view. Direct raw-report/known-usage checks retain exact execution-source refusal; release audit retains exact integrity failure. This distinguishes failure boundaries without accepting stale evidence or weakening R18.

Reviewed retained red/green receipts and source-inventory records: runtime map remains `f384861…`; controlled facts and historical originals remain intact. No new tests, live/vendor/paid/race execution or secret access occurred during this supplement.


补审之后完整普通通过的实际结果见 [最终验证索引](verification.md)。报告引用的临时探针原件同时保留在仓库根 [审查目录](../../../.evidence/provider-sites/review/)；可持久化使用的修复/复审原生证据见 [首轮](../../../.evidence/provider-sites/review-fixes/index.json) 与 [边界轮](../../../.evidence/provider-sites/review-fixes/followup-1/index.json)。
