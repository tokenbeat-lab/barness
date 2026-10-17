# Ticket 10 prepay checkpoint

This is synthetic/offline implementation evidence. No actual credential file
or vendor endpoint has been accessed. M3 is still absent from BuiltinCatalog;
three genuine candidate reports, data-only promotion, three own formal reports,
M3.1 NOT_RUN reports, fourteen renewed legacy support identities, four renewed
business effects and the final actual release gate remain outstanding.

Source slices: ba8305a (sourced prices and initial review fixes), 9de9068
(actual-origin/ledger and complete tool integrity), 65b20b9 / 5decc7a (keyless
renewal and bounded GLM import fixes), ecbbeeb (final runtime and launcher
settlement). The tools-only keyless-plan correction follows these slices.

At ecbbeeb, all MiniMax offline public E2E passed with frozen pi differentials
and pressure enabled (44.150s), including full software rollback. Fifty-one
selected runtime cases also passed under race (5.033s). Both vets and the whole
no-key live suite passed (104.351s). Source-frozen prepare/check receipts have
credentialsRead=false. The final twelve controlled launcher checks pass,
including executing the emitted keyless verification sequence twice; nine
controlled renewal CLI checks pass. Synthetic CLI processes cannot certify
actual model support. The final exact-source preparation follows the tools fix.

```sh
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test ./ai/e2e -run TestMiniMax -count=1 -timeout=8m
go test -race ./ai/e2e -run 'TestMiniMax(ChatMixedReasoningFrames|ResponsesFunctionIdentityFrozen|FinalSamplingExactRanges|ResponsesSamplingMapExactRanges|ChatFinalToolIntegrity|NativeCallIDsAreDistinct|WritePriceUsesSelectedTier|OneHourWriteHasNoNativeAnthropicPrice|MessagesUsage|PromotionBoundary)$' -count=1 -timeout=4m
python3 .scratch/barness-ai-minimax/tools/test_paid.py
python3 .scratch/barness-ai-minimax/tools/test_renewal.py
python3 .scratch/barness-ai-minimax/tools/paid.py prepare .evidence/minimax-prepay-new
python3 .scratch/barness-ai-minimax/tools/paid.py check .evidence/minimax-prepay-new
(cd .scratch/barness-ai-minimax/evidence/ticket-10 && shasum -a 256 -c SHA256SUMS)
```

The three tarballs contain curated original case bytes and original manifests /
audits, plus capsule.json stating which cases were selected. Omitted cases are
explicit; these are not complete release-gate input bundles. All selected
artifact/fixture hashes were independently recomputed before archiving.
provenance.json pins original manifests, archives, selected inventory and logs.
The live-tools capsule preserves the original sealed failed research captures,
finite rerun refusals/replay, explicit zero usage and origin-forgery checks.
No raw capture or assertion outcome was rewritten. Text logs replace only
absolute checkout/home/temporary paths; their original hashes are retained.

Red logs precede the corresponding fixes. Two full MiniMax attempts failed only
because checkout bytes changed during rollback's before/after comparison; a
subsequent stable-root run passed. The first whole no-key live attempt failed
because nine ignored historical ZIP fixtures were absent from the worktree;
fixture-hydration.json records exact copied historical bytes and pins, and the
subsequent whole no-key run passed. These failures remain retained. Historical
ZIPs were neither renewed nor rehashed into current real support evidence.

The public fee DTOs remain unchanged. Actual positive unpriced automatic writes
or positive native one-hour writes fail with retained known usage. Pricing is
fixed to sourced M3 standard CNY/dated FX and the selected input tier; M3.1
synthetic rates are never a public price or formal support claim. ADR-0041
records the data-only activation, external trusted-origin boundary and timing.
