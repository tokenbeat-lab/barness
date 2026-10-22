# Full ordinary input audit

Static consumer audit only; no test/build/consumer execution and no repository mutation. Primary inspected at `ed6d4aba84b340efe99a7b909974537e4779c954`; pending qualification source also inspected at `/private/tmp/barness-kimi/qualification`, HEAD `09a336ff56534d405ac189c1e65a0c22a880f996` (including its working tree).

Intended argv: `go test -json ./ai/e2e ./ai ./ai/internal/testkit/supportmatrix ./ai/internal/testkit/release ./ai/live/cmd/regionalverify ./ai/live/cmd/regionalpromote -count=1 -timeout=20m`.

`ordinary-input-candidates.json`: **145 -> 122 files**, 3 additions / 26 removals, 28,469,472 bytes. Every retained path is an existing regular nonsymlink file and its byte count was rechecked. Inventory SHA-256: `be6a2fade5c676e8f12b4ecacbe291a0698f6d2ef16843067c534d0d87d6e7c8`. Enumerate every row as a separate `--input <path>` when freezing. This JSON is an inventory, not a run receipt or proof that the final source is frozen.

| Added path | Actual consumer |
| --- | --- |
| `.scratch/barness-ai-glm/evidence/ticket-13/rollback-probe.txt` | `ai/e2e/glm_release_launcher_test.go:23` and `glm_corrective_launcher_test.go:23` copy it; GLM `rollback.py:104` reads it to compile the public restored-baseline probe; `commands.py:42` hashes it. |
| `.scratch/barness-ai-minimax/evidence/ticket-08/rollback-probe.txt` | MiniMax `rollback.py:106` reads it to compile its restored-baseline probe, invoked by `ai/e2e/minimax_full_rollback_test.go:40`. |
| `ai/internal/testkit/minimax/pricing-source.md` | `price.go:20` embeds these exact Markdown bytes, then `VerifyPriceSources` verifies them. Markdown is excluded by the outer default source snapshot, including embedded Markdown. |

Retained input groups and their entry points:

| Input group | Consumer / collection rule |
| --- | --- |
| Six `.evidence/minimax-provider-sites-final-round2-raw/{glm-*,support-*}.zip` | `glm_formal_usage_test.go:87`, `glm_formal_reports_test.go:45`, `glm_support_view_limits_test.go:28`, and `supportmatrix/glm_derivation_source.go:98`. Exact paths are selected by current `ai/live/testdata/glm-formal/{live-captures,support-views}.json`; the current indexes select only these six ZIPs. Pin outer ZIP files; extracted members are derived inputs. |
| GLM `commands.py`, `run.py`, `glm_launcher.py`, `corrective.py`, `rollback.py` plus probe | `glm_release_launcher_test.go`, `glm_corrective_launcher_test.go`, `glm_full_rollback_test.go`. Copy/import/hash/execute chain checked; launcher fixtures use owned temporary synthetic Git roots and fake Go. |
| MiniMax rollback script plus probe | `minimax_full_rollback_test.go` invokes the actual script; it reads the sibling probe. |
| Decisions `jev-baseline/{baseline.zip,manifest.json,verify.py}` and `text-choice-evidence/{text-choice.zip,manifest.json,verify.py}` | `jev_baseline_evidence_test.go`, `decisions_evidence_test.go`: copy outer archive/manifest to temporary bundles and invoke the actual verifier script. Verifiers read inside those ZIPs, without following provenance strings into current `.scratch`/`.evidence`. |
| Decisions `runtime-evidence/runtime.zip` | `release_rollback_import_test.go:20` directly opens the ZIP and extracts selected `runs/full` entries. |
| Entire Decisions `formal-support-evidence/live` loose tree | `release_review_import_test.go:102` uses `os.CopyFS`. All **85 files** are enumerated; no missing loose files. |
| Chinese evaluation `live` and `not-run` loose fixtures | `chinese_evaluation_test.go:294` copies all live files; `chinese_routes_process_test.go:201` executes verification; `release/artifacts_test.go:74` consumes both. All **6 live + 5 not-run** files enumerated. |
| Chinese mixed-operations `pressure.json` | `release/artifacts_test.go:35`, `release/classifier_load_test.go:14` read historical fixture bytes in ordinary mode; they do not start pressure generation. |
| `.scratch/jev-1.14-multimodal-box/verify-11-evidence.py` | `ai/internal/testkit/jevformal/verify.go:150` pins it and `:203` invokes it if source prerequisites permit. Keep this executable boundary even when current source drift causes an earlier refusal. |
| `ai/internal/testkit/{regional,kimi}/README.md` | Native `RegionalExecutionSource` reads all files under these two prefixes; GLM source also reads regional prefix. These hashes are observed by ordinary source/qualification consumers. Pending qualification expands the Kimi prefix. Both Markdown files stay explicit. |

Removed paths and reasons are enumerated exactly in sibling `final-input-audit-delta.json`:

- Three superseded `.evidence/barness-ai-glm/ticket-11/*.zip`: no longer selected by either current GLM index. Current six outer archives retained.
- Three `.evidence/minimax-research-captures/.../response.sse`: provenance references only. Ordinary mirror/Responses cases open independent checked-in `ai/e2e/testdata/minimax-observed` fixtures and do not follow `sourceRelativePath`.
- Five GLM historical maintenance scripts (`index.py`, both `preserve*.py`, `reconcile-rollback-history.py`, `seal.py`): no copy/import/exec/hash in the selected ordinary paths. The actual `commands.sources()` list is explicit.
- Fifteen Jev historical receipt/seal/other verifier files: current consumer selects only `verify-11-evidence.py`. `verify-11` imports standard-library modules and consumes the supplied bundle, without invoking the other historical seal/verification tools.

Do not include nested historical source snapshots or giant archived evidence trees merely because their names occur in provenance. Regional/Kimi promotion tests create their own exclusive `.evidence/...` temporary directories; these are outputs, not pre-existing fixture inputs. Pending qualification adds generated temporary grants/captures/proposals; its actual code, compressed official sources, candidate data and test fixtures are already within the default `ai/` snapshot, with the Kimi README exception above. `ai/internal/testkit/minimax/README.md` is neither embedded by the pricing module nor covered by the native source-prefix selection, so it is not an input.

Two rollback boundaries also read **immutable historical Git objects**, not corresponding current working-tree files: GLM baseline `f5cf5362610a8bdb8f3e170f408f3858c149e540` / tree `e0e006866c606fdab3bf02be0322bde2eae02300`; MiniMax baseline `ecfd4b4d6ec54db30453f4b5112989e9fdbd0782` / tree `b9798a4e96dfb08cbfcce45c182915589b42eed1`. Their scripts check these exact tree identities, archive all tracked blobs, preserve full restored-file SHA/mode inventories and execute the restored old-provider E2E. Do not substitute current `.scratch` files or enumerate mutable `.git` pack layout as selected inputs. Preserve those native rollback outputs with the final ordinary receipt.

Final-freeze follow-up for root: capture final source plus these explicit input hashes after all merges. Re-evaluate the inventory if a consumer/input changes; this audit does not renew old live evidence or authorize live/race/pressure.
