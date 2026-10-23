# Kimi optional cache reads and finite allocation costs

Authoritative steering: `/tmp/barness-kimi-cost-bounds/contract.md`; code baseline `2d55fcf`, implementation `178d8c3`, integrated frozen source **`3a722ba3e89808dcbf081758dd33dfeeed813e67`** after merging the root's updated contract at `a1bf7e1`.

Chat/Responses now treat absent/null cached reads as semantic zero for completeness, after Chat's existing nested/legacy selection. Invalid selected counters still fail. Missing writes remain unknown observations and UsagePartial. With known input P/output O/read R, costs enumerate feasible endpoint allocations W=0 and W=P-R under actual or both possible write TTLs. Min/max use feasible amounts, supporting reversed/equal Host rates; equal totals resolve. Public observed writes are unchanged and Cost equals Minimum. Known P/O also rejects inconsistent total_tokens even when writes are missing. Messages read requirements, old Providers and Kimi's rejection of tiered pricing remain intact.

The captured China response is copied as exact original responseBody UTF-8 bytes, never rewritten. Fixture `ai/e2e/testdata/kimi-cn-chat/optional-read.sse` is44727 bytes, SHA-256 `b2b11202e8e178f527dd0014558c6f130eb56426d1f6b4a73db63d7dc552b185`; origin JSON binds old FAIL manifest `4ca9d3e44750c2be46fcf5784fa4d1612e7d43c346d682535df8d505d3e0005e` and exchange artifact `91f2148450e52751e7ad502e19f3f33b9a71ac1993dab907781782d24fd37603`. Public replay now reports P94/O196/reasoning178/read0/write0 as complete/resolved. Old real failure, grant and ledger remain untouched; this replay grants no live qualification.

All commands use SOP `run.py --feature barness-ai-kimi-cost-bounds --stage e2e`. Exact argv/exit/source/bundles are in each original receipt; `/tmp/barness-kimi-cost-bounds/tdd-attempts.json` gives their human-readable chronology:

| Attempt | Actual outcome |
| --- | --- |
| `20261023T010306.858803Z-motwals1` | exit1: real SSE actual partial/partial versus expected complete/resolved, before any runtime edit. Matrix also had observer artifact setup failures from slash-bearing RequestID. |
| `20261023T010324.443959Z-cez0xs1c` | exit1: read-default slice made real SSE pass; matrix setup failures remained. Not claimed as all-green. |
| `20261023T010357.747797Z-i1g20d8u` | exit1: missing-write expectation changes preceded bounds implementation, but observer setup failure remained. |
| `20261023T010440.143908Z-2gry5xxw` | exit1: clean red matrix after fixing test-only RequestID labels and restoring original kimi_usage.go exactly; optional reads, real SSE, missing-write/allocation/rate/equality/overflow assertions fail at public Client. |
| `20261023T010519.126380Z-17aj43w3` | exit1: controlled native pipeline red before final runtime implementation; synthetic Chat/Responses scripts exercise default reads and missing writes. No vendor requests. |
| `20261023T010603.311739Z-5l_701sw` | exit0: initial cost/read scenarios, Messages invariance and existing tier-refusal guard green. |
| `20261023T010701.035304Z-3hdpuc45` | exit1: newly identified known-P/O total contradiction reproduced before tightening equality guard. |
| `20261023T010729.306969Z-dphmc9i0` | exit0: frozen primary scoped green;388 native E2E PASS cases, no skipped cases, audit findings empty. Changed public matrices, real replay, six-route controlled native execution/read-only consumption, old Chat/Anthropic/GLM/MiniMax usage. ai also ran26 PASS test records. |
| `20261023T010943.585780Z-pcqrosab` | exit0: same-source supplemental10 native PASS cases; exact Host overflow/containment names, old DeepSeek Responses usage and unchanged Messages missing-charge qualification. |

The supplemental selection corrects unmatched planned generic `KimiHostUsage`/`ResponsesUsage` names in the primary selector; neither is claimed to have run there. The two final manifests form an exact finite398-PASS-case set in `cache-bounds-implementation-01.json`, with precise manifest/audit hashes and actual Go test inventories. The root owns final full-package validation, new native builds/grants and publication state.

All9 attempt roots, including setup failures, are retained byte-for-byte in primary `.evidence/barness-ai-kimi-cost-bounds/e2e/`:6250 original files. Comparison proof `.evidence/barness-ai-kimi-cost-bounds/retention-01.json`, SHA-256 `859a310d7122fde7424dac4529989d954e9e2f12357e7f8d9b1390ecf40a28b8`. No old pin, capture, grant, source or original attempt was overwritten. Frozen source is1351 files with no extra selected inputs; unchanged after Markdown evidence closeout.

Native controlled qualification now proves finite cost bounds with partial missing-write reports and optional/default reads. The native assertion describes reliable finite bounds rather than claiming every counter is present. Requests, reservations, retries, grants, and lifecycle mechanisms are unchanged. No `.env`, credential, vendor call, paid retry, race, pressure or child full-suite run occurred. Those stages remain NOT_RUN; `release.pass=false`, and issue07's human gates remain pending. Revert implementation `178d8c3` to roll back while retaining every original and source-bound failure.
