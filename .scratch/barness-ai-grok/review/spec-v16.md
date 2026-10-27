# Spec review — issues 21/22

Fixed base: `e1f6986677c905b48d80a6b01fc0c0f125427c82`. Reviewed pin: `c82a24753cc970265f37377e5ee3e2a975b7fa6a`. Exact command: `git diff e1f6986677c905b48d80a6b01fc0c0f125427c82...c82a24753cc970265f37377e5ee3e2a975b7fa6a`.

**1 finding (P2).**

**[P2] Reject an escaped private claim-store directory before claiming.** Issue22 line18 requires “拒绝 worktree、外置目录及符号链接逃逸；仅写自己的私有 bundle/store”. In `ai/e2e/testdata/process-grant-fixture/main.go:81`, the helper calls `livegrant.Claim` after validating only cwd, `.git` itself and output ancestry. An otherwise valid `barness-process-e2e-*` repository whose `.git/barness-live-process-grants` is an existing symlink to an external directory passes those guards. `ai/internal/testkit/livegrant/claim.go:52–55` follows that symlink through `os.MkdirAll` and `os.OpenRoot`; subsequent lock and ID/scope claim writes therefore escape the owned repository. This breaks the helper’s explicit private-store boundary, including its promise not to write a real common store. Reject a symlink or externally resolved store path before Claim and add a public helper countercase using an external synthetic target, asserting zero target changes. No production consumer relaxation is needed.

No additional missing/partial requirement, unauthorized scope expansion or incorrect implementation found. Issue21 now uses bounded process/workflow deadlines, parent-deadline allowance and the existing collector’s cooperative group cleanup; its slow-start/cancellation checks retain the separate32-case native inventory. Issue22 adds the complete seven-scene synthetic Host/provenance/downgrade contract and typed claim evidence without changing a production consumer.

The retained local GREEN and tool results are implementation evidence; complete frozen ordinary/native applicability verification remains pending. Recordv16 still says E2E FAIL and live/race NOT_RUN. Changes to the global ordinary source do not authorize rebinding v8. This review performed only source/evidence reads: no tests, builds, consumers, claims, credential/store inspection or vendor calls; no repository edits.
