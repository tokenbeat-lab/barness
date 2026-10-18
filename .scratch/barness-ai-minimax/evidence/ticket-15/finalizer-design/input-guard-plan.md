# Fixed input guards for the new six-process finalizer

Prepared read-only from committed integration `4d83ca2b00b76f313c32d62f020a331a4cf3b88c`; this does not inspect or approve worker ticket15 WIP. No tests, verifier, publisher, Key or network call was executed. Write only this plan under `/tmp`.

## 1. Three identities must stay separate

- Final keyless gate source **F** is a new committed final tree, chosen after new actuals, permitted GLM data publication and the four-row default-matrix update are complete. Freeze F and current source bytes before building the new gate.
- Each new native report keeps its actual measured GLM/effects commit, original command, capture and canonical receipt. Each reused generic/text native report keeps `8fbd17006677774368d4d77d924a8672fd205c51`; M3 keeps `89000346b48480d94c3d6f6523629870f4e7b1b0`. Never replace those with F.
- Old L's full freeze remains stale. Never call `c.guard(oldL)` as a prerequisite for reuse, change its snapshot, rebind `c.ROOT`, or bypass its refusal. Current applicability comes from the existing native readers plus an explicit unchanged relevant-source proof. The prior plan `/tmp/barness-repair-six-publication-impact-review.json` limits publication to four GLM data files and four new matrix rows.

## 2. Authenticate the old roots, without a locator or extraction

Use an explicit trusted common root. Open every path component with descriptor-relative no-follow semantics, regular-file checks and bounded reads; JSON must reject duplicate fields. Reject missing/extra/linked/resized/changed files. Do not trust an archive manifest's absolute `original` path or search another worktree. Authenticate the metadata hashes below from committed bytes, then the archive/inventory hashes, and use only the fixed prefixes. A reused on-disk subtree must have exactly the inventory's member set and per-file size/SHA. Recheck before and after each native step; 465 old L files plus 525 M3 files are sufficient for the selected trees. Do not copy, repin or rewrite old source/config/capture files.

**Old L:** `.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/index.json` SHA `d0a17474721deef08bd10e004762b117f9b669288b9dbfac7d8041b725622d62`.
Its compressed inventory SHA is `7639de8797a2a6034446eda072b4cad528e75c7c6a9546d0be5137a2adfcd025`; decoded inventory is 357010B, SHA `bb1a7c604a884199a926879d0196542e04533c1ebd0b39f1df2b2a360cd1f7c3`, 2228 files.
The common-root archive `.evidence/minimax-retained/ticket-13-actual-legacy/actual-legacy-raw.tar.gz` is 5833549B, SHA `b8ebed319b4f6f3a7ea40d18fe9bf967f92959ca472cc41d1caf0b24b4a5ca87`; tar prefix `routed-legacy-actual`.
Use the established bounded reader anchor `b280863f49f4a56cb507ee7e6ba753c73e38c9162236592e9f886ed9a64c3c2f` (implementation/verify.py:32MiB compressed/192MiB decoded/32768 entries), not an unbounded tar extraction. The native subtree roots below are relative to `$COMMON/.evidence/minimax-routed-legacy-1/`; archive keys are `actual/` plus these paths.

| Native input | Fixed relative root | Files | Native report SHA256 |
|---|---|---:|---|
| openai-responses | `live/openai-responses/20261018T082351.964640000Z` | 53 | `d60839f2f9561d837fa822dbd69577e8e84948083b3ccd84e913d2ee31f56c43` |
| openai-chat | `live/openai-chat/20261018T082415.346674000Z` | 55 | `9639a1f1186c5261d605b0e6b7d7852414a2600f1e0c7b7befa990f758295c5b` |
| anthropic-messages | `live/anthropic-messages/20261018T082432.962715000Z` | 74 | `3de558312237206ebf07264ca60f3c9ce058385dd96b74bc5da306a7fce8b214` |
| google-gemini | `live/google-gemini/20261018T082500.989942000Z` | 33 | `630ef3ef5774e6197c6ae8abcedeb9de7926d95b2972d5df205662b7c4c68ea1` |
| deepseek-responses | `live/deepseek-responses/20261018T082519.737452000Z` | 41 | `227f61ff7cbd23d34cddf10a4f63f6919fe850a550dd2fbf128ef5ad1b6524a8` |
| deepseek-chat | `live/deepseek-chat/20261018T082530.544023000Z` | 53 | `9f1dd332cde1895c20b1aa59a0447c78b6c1938f502aa5380be6adb00be9399e` |
| openai-images | `live/openai-images/20261018T082544.322311000Z` | 21 | `1dcb77d06f1025ff6ca397599f56c093a142da272ba89ccc06e18b29fa1257ed` |
| google-interactions-image | `live/google-interactions-image/20261018T082613.527826000Z` | 15 | `31e81c996a2bbd3be666f369b8d542807fbd730daac0da2d978f3b985b830dba` |
| openai-decisions | `live/openai-decisions/20261018T082643.731045000Z` | 85 | `e132afe1b54b94481e3ac57421a9a3af1d1da6e5a7cda8cdfb6c27f077aacae7` |
| typesafe-classifier | `live/typesafe-classifier/20261018T082648.885174000Z` | 15 | `673cad94278a0cbdc73bc0155d5ecf4fec9a95d6bda5aa4511dae538eb9e1e31` |
| text-jev | `text-jev/20261018T082658.776631000Z` | 10 | `46e2d6d34b4e8e759723ce21e922fc372840ba3a30d6af055a337133276975c6` |
| text-luna | `text-luna/20261018T082706.819196000Z` | 10 | `64dc5b7ae0ae4cfcbe823cb6f9af3284c10f6649467cc428bf6e749d7a691a38` |

Only these ten generic plus two text trees are reusable. The same archive contains **old** protocol114, visual2 and derived support; do not select them for the new finalizer. For the twelve reused rows, archive `canonical/{name}.json` corresponds to common-root `$COMMON/.evidence/minimax-launches/routed-validation-1-{name}.json`, not the older unprefixed intent. Reopen and hash the exact original canonical receipt, original command and authorization/source/binary/plan/check/review metadata against the retained `frozen/` inventory; this authenticates history without asserting that old whole-tree freeze is current. Preserve all39 old receipts and545/545 separately from the new finite six.

**M3:** `.scratch/barness-ai-minimax/evidence/ticket-10/international-formal-glm/international-formal-glm-8900034.json` is 992099B, SHA `121075a2a34a35c9966c6bf2c81eff29f20b587df103c3eb4f071ebf703183b4`.
Its sibling archive `international-formal-glm-8900034.tar.gz` is 5511291B, SHA `8b649a95fbc5d0e668173400a92ce2395ef86dc491762f6b9fff9a8ecf5aa07e`. Tar member names match its `files` keys directly. Select exactly `family=minimax` and these three jobs (175 files each), not any GLM/forensic family stored beside them. Native roots are `$COMMON/.evidence/minimax-global-formal-glm/{job}/{session}`; archive roots are `minimax/{job}/{session}`.

| M3 job | Session | Report SHA256 | Seal SHA256 |
|---|---|---|---|
| mm-global-formal-m3-chat | `20261018T030234.929339000Z` | `e65cd00d598befd62b9af32e2daa42e0f06d83878187afe83d2f90c164dd03ff` | `0da5d4da4cab42b59006dd2fc419e4e5c44729784caa2221e4352c0c0fd7dd1e` |
| mm-global-formal-m3-responses | `20261018T030308.589357000Z` | `33fb060912110b35680c0baadd37566f3380d165998c1ec38334d47f129dc425` | `51159322f22edfa6ca1ebe13bbdbbda4f4628aa5e80ecc8546131f53d0f57a86` |
| mm-global-formal-m3-messages | `20261018T030335.626044000Z` | `4932f71f744a2532269ed64005447e6ab98f859cd288ccefa1350ed747e907fd` | `4260cedaaa01324a1fd53ba775116ddef00bfea09c6e4931bff28a189e75c4d1` |

Also pin each `live-report.json`, `minimax-ledger.json`, native manifest/audit and captured canonical `launches/{job}.json` from this same index. The current `ai/live/testdata/minimax-formal/origins.json` hash is `f50b4e5edb3ff9a3e0725f63077c98b73c53fd7f1838342ce532c995ff4b4aaa`; it is a trusted repository fallback only when the canonical receipt is absent. A present invalid canonical must fail, not fall back. Existing `minimaxlive.Verify` checks the recorded ledger receipts as an exact prefix of the common formal ledger: do not demand whole-ledger equality if the canonical append-only chain legitimately grew.

## 3. Applicability is a second, independent guard

At F, keep current catalog version `2026-10-18.1-minimax` and hash `sha256:3718603f6ad6cee8a5b66251bede572b62e58fd290b3d7fc7a49504a1d274e4f`. The checked-in catalog snapshot byte SHA is `fd6d59db596e8e4a81f970e50d6fe0567f86306957c1e934c1fae5e1c36dce63`. No runtime/model/price/catalog embed/SDK/shared producer drift is allowed for old generic/text/M3 reuse.

- M3: public current `minimaxverify -formal` must retain its exact263 execution maps, SDK, native origin/seal/ledger and current full-catalog checks. The independent MiniMax README path outside the broader1048 launcher selection still counts; never drop it.
- Text2: keep original full860 snapshots and CodeHash byte-exact; existing `release.currentEvaluationSource` compares the261 relevant paths, while `chineseeval.Load` checks catalog and task/config. Use the current public Chinese verifier, not an old frozen binary as proof of current applicability.
- Generic10: compare original report/manifest/catalog/SDK/binding/route/scenario artifacts and existing public consumer checks; keep original measured commit and matrix times. The existing affected-combination rule is mandatory even though these native reports have no whole-tree hash map.
- New GLM3, protocol114 and visual2 must be actual new own reports under the reviewed six contract. Pin their canonical command/report/source/binary/plan/checks/review/authorization hashes and settle actual counters from native bytes. New round identity/human108/132 authority must come only from the eventual committed reviewed contract; do not invent names or consume authority from old72/96 or routed continuation receipts. Strict GLM selector publication requires all3 same-source complete PASS and canonical capturePins; no old Chat/Messages splice.
- Reopen new Jev's own protocol and visual trees and verify via current public `jevformal -require-live`. Existing protocol all-Go and visual non-test-Go source maps must match F; data-only GLM selectors are outside these native selections. All inputs get a bounded full subtree membership/hash inventory in the final plan.

## 4. Matrix boundary: four writes, thirteen preserved rows

Old matrix is authenticated by old L inventory member `derivation/matrix-after.json`:55889B, SHA `7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f`, equal to current committed default matrix. Before any merge, freeze its bytes.

Preserve these13 row objects and their order, capabilities and all timestamps exactly: `openai-responses, anthropic-messages, google-gemini, openai-chat, deepseek-responses, deepseek-chat, typesafe-classifier, openai-images, google-interactions-image, openai-decisions, minimax-chat, minimax-responses, minimax-messages`. A convenience fingerprint is `e4f98ab323f257e88277c0f642870efb4e1d918fd19b78579f3b732a1a404663`: SHA256 of UTF-8 `(json.dumps(selected_rows, sort_keys=True, separators=(',', ':'), ensure_ascii=False)+'
')`, selected in the old matrix order. Prefer parsed deep equality against the authenticated old matrix in the guard; this fingerprint does not replace native verification.

Apply only the new own `zhipu-chat/zhipu-responses/zhipu-messages/typesafe-classifier-jev-1.14.0` reports through the existing public supportmatrix CLI, first on a temporary copy then default. Do not remerge old13 reports (their old FinishedAt must not be refreshed). Verify exact17 identities and old13 equality, save both before/after full matrix hashes and real merge receipts. A temporary reader-validation matrix must remain separate from the default/history; the full public releasegate already validates each report against its matching current row without rewriting history. Commit permitted data before choosing final F. The new matrix hash becomes a fixed final input; no later mutation is allowed during the gate.

## 5. Fresh final plan and direct compiled gate

A small fixed finalizer is enough: prepare a fresh exclusive output directory, compute the exact fixed17 native roots and four effects, current F source/selector/catalog/matrix pins and operational hashes; build only fixed keyless tools; save a plan; obtain its independent review; run once. No user-selectable policies/commands, global ROOT rebinding or generic fallback layer.

The17 live inputs are oldgeneric10 + oldM3three + newGLMthree + new-derived-Jevone. Effects are oldtext-jev/oldtext-luna + newvisual-jev/newvisual-decisions; newprotocol114 is transitive inside new Jev. Require distinct owned roots, complete native identities and each native report's original measured commit. Hash every source/capture/manifest/audit/seal/config/fixture artifact selected by the trusted inventory, not just top-level reports. Refuse unexpected symlink/FIFO/extra/missing files or unknown consumption/identity; preserve honest finite model failures rather than substituting older passing evidence. Actual verifier refusals are retained as failures, never coerced to0. Any false full verdict remains false.

Compile the **current** public `ai/release/cmd/releasegate` at F with a real build receipt and new binary SHA; the old A2 `c54a7d...` binary is historical and cannot be reused as a fresh certificate. Run that compiled binary directly. Never use `go run` to launch it: the retained RED/GREEN probe demonstrated GOROOT's PATH prepend bypassing the observer.

Use the unchanged numeric observer SHA `aec940df7db842009d9c977142dc35f604022b380ce383af7ce2dd85383879d7` and absolute real Go `/usr/local/go/bin/go` SHA `005640c7ff93028cb704283b0f737f2db3faf8b51b2561170c769b83905da646`. Preserve its exact-root matching (`BARNESS_GATE_RECORD_REPO_ROOT=F checkout cwd`), exclusive receipt creation and only these four argv:

```text
vet ./...
vet -tags live ./...
test -count=1 -timeout=20m ./...
test -race -count=1 -timeout=75m ./...
```

The final public gate argv contains all17 `-live` roots and all four effect flags, with **no** `-bundle`/`-race` adoption overrides. Keep a credential-free allowlist environment, exact helper/observer/realGo/build/compiledCLI/input plan hashes and the existing finite6600s outer gate cap. An interrupted receipt stays STARTING/unknown; do not replace it with a Boolean or automatic rerun. Recheck F/committed source, all original/new inputs, matrix/catalog/selectors, helper and binary pins before/after every keyless validation and the gate; hold HEAD through actual ordinary/race completion. Record true integer vet/vet-live/ordinary/race exits, outer exit and actual report presence/pass, including1/false.

## Primary committed references

- Existing authority and scoped reuse: ADR0045/0046; base spec Testing Decisions§6; ticket15/design v2.
- Old legacy pin/membership protocol: ticket13/actual-legacy/{index.json,inventory.json.gz,verify.py}; M3 formal original index and `minimaxlive/{origin.go,verify.go,source.go}`.
- Native applicability: `release/{live.go,live_manifest.go,evaluation.go,evaluation_source.go}`, `supportmatrix/{glm_source.go,glm_derivation.go,glm_derivation_source.go,matrix.go}`, `jevformal/verify.go`, `visualeval/bundle.go`.
- Publication write-set: `renewal_glm.py:96-141`; matrix-only save: `live/cmd/supportmatrix/main.go:36-59`.
- Direct compiled observer proof and original failure provenance: ticket13/keyless-launch-history and final-gate. Old helper `a1ecf4a0...` is a historical example only: its `c.guard(L)` and source equality predicates must not be carried into this new plan.
