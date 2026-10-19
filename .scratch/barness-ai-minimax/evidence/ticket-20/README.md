# Ticket20 — offline visual cancellation evidence

The reviewed source fix is `d02b5adb862e98d73672b47bfd1e3142a2d6082d`. This frontdoor preserves its keyless TDD history; it grants no current live or release qualification. The old full gate and all original vendor reports, selectors, source maps and failures remain unchanged. Current formal applicability is eight of eleven groups; protocol114 and both visual groups require a separately authorized renewal.

`source-proof.json` compares the actual precommit HEAD plus WIP compiler inputs with committed d02 bytes: all 1,040 Go/JSON/module files match. These are measured WIP commands, **not postcommit test runs**. `source-impact.json` explicitly identifies its independent translation of the unchanged native selectors: complete selected sets remain equal for M3×3 (268 each), GLM×3 (220 each) and text×2 (265 each). Protocol114 and visual×2 are affected. This translation is not public CLI introspection or a replacement for native source verification.

## Actual history

The [index](index.json) pins 45 numeric command receipts and independently checked main-manifest distributions. Raw setup failures and source versions are retained alongside semantic failures.

| Stage | Actual result |
| --- | --- |
| Original extracted producer, ordinary / race | exit1; 8 FAIL each; audit8 |
| First attempted GREEN | exit1; 8 FAIL; audit0; nullable absence mismatch |
| Scope-only GREEN | 14 PASS, ordinary / race; audit0 |
| Scope-only existing visual race | exit1; 49 PASS / 3 FAIL; audit3 |
| Real post-pin header cancellation RED | exit1; 2 FAIL; audit2 |
| Missing-access RED | exit1; main manifest 2 FAIL, including containing case |
| Jev pre-pin Capability / Request | baseline public verify0; both private derived fixtures RED1, then GREEN0 |
| Existing visual regressions at preceding finite-lifecycle source | ordinary116 PASS; race52 PASS; audit0 |
| Later deadline race | exit1; 3 PASS / 1 FAIL; audit0 |
| Real response-read cancellation RED | exit1; 2 FAIL; audit0; one actual HTTP Attempt per call |
| Sent cancellation fix | ordinary / race20 PASS; affected deadline4 PASS and accounting2 PASS; audit0 |
| Null access RED | exit1; main manifest 2 PASS / 2 FAIL; its three negative leaves 2 PASS / 1 FAIL |
| Final required-integer fix | ordinary / race23 PASS; audit0; both vet commands0 |

Compiler/setup failures (`setup.log`, first `after-pin-red`, mistyped Jev baseline path) are labelled separately from behavior RED. The first GREEN failures and later regressions remain failures in their original artifacts. Jev pre-pin fixtures are explicitly **derived consumer fixtures, not observed native executions of those exact phases**. Public no-send replay validates the cancellation category; the original phase/access combination is checked against the source-derived finite seven-row table. No timing collaborator, fake context error, audit exemption or production test flag was added.

## Immutable raw custody

The common-root archive directory is `.evidence/minimax-retained/ticket-20-visual-unsent/`. Git stores only bounded inventories, source proofs, receipts and the reader. Each archive uses the existing unchanged full-reader envelope: 256 MiB compressed, 2 GiB decoded tar, 160,000 regular files, 200,000 entries, 64 MiB per member, 40 MiB decoded / 8 MiB compressed inventory, and 16 MiB per observed metadata file / 64 MiB total. Links and extraction are refused. The split keeps each capsule inside these fixed limits.

| Capsule | Files | Raw bytes | Archive bytes | SHA256 |
| --- | ---: | ---: | ---: | --- |
| core | 90,632 | 444,540,791 | 133,170,950 | `e3b20e4eda1a2fdcc33a43558b15cca0a733af189dcbedff6db149a3ee55a9fa` |
| ordinary | 63,821 | 311,290,348 | 87,323,271 | `bbaea5407aa0d593936ec117e165a1d2267bcec74e7c5227ed6c01766aa0ba6a` |
| race | 48,882 | 239,737,476 | 66,772,084 | `c5c6687265e011b4e181dffde93b5b87599710d0d75a121fb2a5ea54cc44645c` |

Creation used exclusive fresh outputs and bounded regular-file reads. Every private raw input was rehashed after sealing; all remained equal. All three capsules include the source proof, source-impact proof, actual recorder/fixture producer, retention producer bytes and the retained diagnosis/finite-table/Spec-review records. Original source review findings and their narrow closure are distinct. Other parent-owned review reports remain in the global review chain.

Run from this checkout, replacing `COMMON_ROOT` with the checkout holding the three archives:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 .scratch/barness-ai-minimax/evidence/ticket-20/verify.py --common-root COMMON_ROOT
```

The reader pins and reuses the existing bounded readers, streams every regular member, checks whole inventories and hashes, recomputes manifest/audit counts and verifies exact integer numeric exits plus the WIP/commit proof. It executes no captured command and extracts nothing. [Creation](creation-command.json) and [readback](readback-command.json) both exited0. Paid calls0; real credentials read0. All owned command windows are closed.
