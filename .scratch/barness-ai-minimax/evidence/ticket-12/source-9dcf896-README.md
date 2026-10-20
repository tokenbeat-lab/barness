# Ticket12 — offline corrective-1 checkpoint

**PENDING HUMAN APPROVAL.** Implementation at `4ff6dca`, production followup
`5abc1a5`, bounded synthetic fault peer `6ff05b8`. No user `.env` or real Key
read, vendor call, APPROVED receipt or canonical corrective intent was created.
Original GLM66/66 and its Responses FAIL remain immutable. The prospective
additional72/96 means this renewal's GLM ceiling138/162 and legacy440/464,
other302 unchanged; prior GLM-task history is separate, not erased.

[Contract](../../issues/12-glm-corrective-1.md),
[commands and approval boundary](../../tools/GLM-CORRECTIVE-1.md),
[ADR0045](../../../../docs/adr/0045-glm-corrective-capture-round.md),
[summary](checkpoint-summary.json), [failure inventory](failure-inventory.md).
The original actual failure remains in
[the immutable8900034 capsule](../ticket-10/international-formal-glm/README.md).

| Measured checkpoint | Result | Evidence |
| --- | --- | --- |
| 4ff complete GLM executable suite | ordinary61 PASS; race61 PASS; both audits0; exit0 | `glm-corrective-4ff-full.json` |
| Failure-first followup | event-overflow report lost knowninput20; ordinary futureplan incorrectly374/398; then green | `glm-corrective-followup.json` (exact red base+patch/source map and command receipts) |
| 5abc corrected production | ordinary4 PASS/audit0, tool CLI35 PASS, both vets0; three existing M3 formal public verifiers PASS | `glm-corrective-followup.json` |
| 5abc intermediate race | four test assertions PASS; implementer interrupted slow final audit, exit-15, no final audit; **not completed Go PASS and not a confirmed timeout** | [correction](race-status-correction.md), original provisional annotation retained |
| 6ff single-fault fixture | ordinary4 PASS; race4 PASS; audits0; both vets0; all1,033 frozen source files byte-identical during checks | `glm-corrective-6ff-focused.json` |
| Current M3 preservation | each263 source files unchanged; original report hashes unchanged; no M3 spending | `checkpoint-summary.json`, three public verifier receipts in followup archive |

These are offline assertions, including deliberately FAILing child reports.
The legal synthetic response retains2,047,892 complete raw bytes; escaped
exchange2,114,598 exceeds the old replay2MiB bound; event evidence is about2.2MiB
with all4094 immutable deltas. Length/no-text, raw cap overflow and event cap
overflow remain FAIL. Known Attempt usage survives evidence refusal. Current
and historical replay preserve exact final wire, Result, Observer and observed
cancellation; replay never publishes actual support.

The two independent review axes are retained byte-exact: `4ff6dca-*-review.*`
(Standards0, Spec2 open then), `5abc1a5-*-review.*` (both0 open after fixes), and
`6ff05b8-*-review.*` (both0 open for fixture-only delta). Reports are never
relabeled. [Final input context](final-input-plan.md) is the parent's read-only
plan, not authorization. The actual R1 raw/views and final legacy/effect release
inputs do not exist yet; no full-release PASS is claimed.

Every listed archive was independently streamed back and compared against its
complete file inventory, with physical manifest Artifact hashes checked. Input
Fixture attribution hashes are kept distinct. Early uncommitted TDD diagnostics
are historical observations (some fixture setup/decode failures), not a paid
source freeze; their recorded claims are unchanged. The exact followup red
patch and final source maps permit repeatable scoped verification.

All four complete archives survive worktree cleanup in
`$COMMON/.evidence/minimax-retained/`:

- `glm-corrective-4ff-full.tar.gz` —15,823,059 bytes;15,658 regular files; retained
  in common ignored storage. Its tracked JSON and compressed full inventory pin
  every byte. It exceeds the8MiB curated-blob threshold.
- `glm-corrective-followup.tar.gz` —6,024,969 bytes;4,041 files; also tracked.
- `glm-corrective-6ff-focused.tar.gz` —4,003,250 bytes;3,677 files; also tracked.
- `glm-corrective-initial-stages.tar.gz` —6,641,296 bytes;6,433 files; also tracked.

Matching `*.json` indexes and `*-files.json.gz` inventories record hashes,
expanded sizes, manifest outcomes and audit presence. No original raw bytes,
assertions, source pins or failed history were repaired. `curation-tool.py` is
the reused bounded archiver: finite entry/depth/size limits, no symlink following,
full tar streaming byte verification. Checksums cover all curated tracked files.

Reproduce on the pinned source with an empty credential environment and a fresh
output directory:

```sh
env -i PATH="$PATH" HOME="$HOME" BARNESS_AI_EVIDENCE_DIR=/tmp/glm-corrective-repeat go test -tags live -count=1 -timeout=5m -run '^TestGLMCorrective' ./ai/live
env -i PATH="$PATH" HOME="$HOME" BARNESS_AI_EVIDENCE_DIR=/tmp/glm-corrective-repeat-race go test -race -tags live -count=1 -timeout=5m -run '^TestGLMCorrective' ./ai/live
```

Run the tools from their directory with `python3 -B -m unittest test_paid
test_renewal test_glm_corrective.Corrective -v`; both `go vet ./...` and
`go vet -tags live ./...` are retained. The unchanged Python at6ff has the35-test
5abc receipt; there is no fabricated final-HEAD repetition. Parent will prepare
and perform the full no-key live check on the final integrated HEAD before
asking for the distinct paid budget approval.

## Parent integration and portable current archive plan

Integration merged byte-exact at `a080b37`; [merge proof](integration-merge-proof.json) checks all1,033 generation files against reviewed6ff and each original M3 263-pin map. [Archive portability](archive-portability.md) confirms no Go change is required. If corrective-1 is explicitly approved and all three actual captures PASS, use a fresh `.evidence/minimax-ticket12-current-glm` for raw archive publication and same-directory support-view derivation. Force-track all three original generated raw ZIPs with raw selectors/checkpoints before deriving views, then all three generated view ZIPs with the view selector before L freeze. Do not relocate files or rewrite raw-index checkpoints. Original captures and their sources remain immutable.

The original implementer checksum receipt and overview are retained byte-exact as `source-a080b37-SHA256SUMS` and `source-a080b37-README.md`; their paths describe Git source a080b37. The current `SHA256SUMS` also covers parent integration/context additions. Raw captures, original indexes and reviews were not changed.
