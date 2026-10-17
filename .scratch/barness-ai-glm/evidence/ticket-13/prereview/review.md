# GLM prepaid source review

## Standards

Reviewed fixed baseline `f5cf5362610a8bdb8f3e170f408f3858c149e540` through `808d6d551408bf357ab897c93b3f9134bbb7aa2d` (tickets 01–12 at `d8b00f7`, then ticket 13 delta). Read-only; no paid/network calls. Known global hook recovery regression is corrected in 13 and excluded. Historical formal evidence is not treated as current support.

Documented-standard violations:

- **[P2] Bound GLM error-envelope parsing before repeated allocation** — `ai/glm_failures.go:19–23`. The hunk `return f.glmCode(envelope.Error)` reparses and copies the complete remaining JSON subtree at every nested `error`, producing quadratic work/allocation and ignoring cancellation during parsing. An exact-function standalone probe with a valid 100,005-byte, 9,999-level envelope allocated **2,103,884,384 bytes** and took **1.96 seconds**; HTTP status handling parses it twice. Such a body can fit the configured error-body allowance while overwhelming concurrent shared calls. This violates **AGENTS.md §5** (fault/concurrency/resource design) and **ADR-0002 Consequences** (limits apply before reads/allocations). Parse once or bound accepted envelope nesting to documented shapes, returning unknown code for excess nesting. Probe: `/tmp/barness-glm-error-review.go`.

- **[P2] Bound derived-support inventory traversal before hashing files** — `ai/internal/testkit/supportmatrix/glm_derivation.go:118–139,166–170`. The hunk accepts every `"original/"` descendant and calls `glmRead(p)` before comparing the actual inventory. Unlike the direct import's 512-file guard, derived views have no traversal/file-count/aggregate-byte limit; even an invalid tiny declared inventory can cause arbitrarily many 16 MiB reads and unbounded path-map growth before rejection. Nested empty directories also escape file counts. This violates **AGENTS.md §§3–5** (validate untrusted boundaries and bound failure/resource costs) and **ADR-0036's finite owned-artifact contract**. Enforce an early finite path/file/total-byte bound, or verify paths against the trusted raw archive before reading.

Heuristic baseline: no additional actionable smell findings. Six independent catalog literals are explicitly required and correctly documented as deliberate repetition.

Totals: **2 documented-standard findings; 0 heuristic findings**. Ticket 13 operational scripts, Host example, complete rollback and additive trace changes introduce no additional Standards finding in this snapshot.

## Spec

Spec review: baseline `f5cf5362610a8bdb8f3e170f408f3858c149e540` → ticket12 `d8b00f7abe3648616815b5e0b48c1dfb1b0fa1bc`, plus ticket13 delta through `808d6d5`.

1. **[P2] Preserve authoritative empty reasoning** — `ai/responses.go:346–348`. R09 requires: “正文、推理与工具参数的 delta 和 authoritative done 都要处理”. After a reasoning delta `"stale draft"`, a legal `output_item.done` reasoning content object with `text:""` leaves `Thinking.Thinking == "stale draft"`: `cmp.Or` treats the authoritative empty string as absent. Native signature correctly contains empty text, so visible Result and native replay also disagree. Select GLM authoritative content by presence, including empty strings, rather than truthiness; preserve existing provider semantics. Public Client/TLS reproduction fails `TestPrereviewGLMAuthoritativeEmptyReasoning`.

2. **[P2] Reject legacy forced-tool fields at the final Chat boundary** — `ai/chat_glm_capabilities.go:15–19` and `ai/chat_glm.go:131–138`. R11 requires: “首期仅接入 tool_choice=auto；其他明确指定形式请求前拒绝，包括 full 的 required/具名强制和 simple 的 none，不能静默改写或靠上游忽略满足请求。” An `OnPayload` callback adding `functions:[…]` and `function_call:{name:"lookup"}` passes both authorization and the GLM final guard, sends one provider request, and returns success against a valid text fixture. The deny list and tool validator check only modern `tools`/`tool_choice`; legacy fields bypass this implemented capability subset. Reject both legacy fields for GLM after callbacks (also `function_call` from SamplingParams), with zero requests/Attempts. Public reproduction fails `TestPrereviewGLMLegacyForcedTool`.

Both findings are evidence-backed and unchanged in `808d6d5`. No additional ticket13 source finding. Pending current-source GLM renewals, legacy paid collection and final release evidence were intentionally excluded, as instructed; this is not final-release approval.

Read-only reproduction: `/tmp/barness-glm-prereview-extra_test.go`, overlay `/tmp/barness-glm-prereview-overlay.json`; `BARNESS_AI_EVIDENCE_DIR=/tmp/barness-glm-prereview-evidence GOPROXY=off GOTOOLCHAIN=local go test -overlay /tmp/barness-glm-prereview-overlay.json ./ai/e2e -run '^TestPrereviewGLM' -count=1 -timeout 2m`. Both fail; evidence `/tmp/barness-glm-prereview-evidence/20261017T151010.746405000Z`. No repository edits or external calls.

Standards: 2 findings, worst P2. Spec: 2 findings, worst P2. Paid collection remains unstarted pending fixes and corrected-source review.
