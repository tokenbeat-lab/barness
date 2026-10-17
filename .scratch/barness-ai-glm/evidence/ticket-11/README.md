# Ticket 11: six real candidates passed

All six model×protocol candidates independently passed their11 required calls. Three direct China standard API processes used66 Logical Calls/66 reservations/66 actual HTTP requests, zero reruns, with all audits and integrity checks clean. Candidate evidence does not publish formal BuiltinCatalog, support matrix or route support.

| Protocol | GLM-5.3 | GLM-5.2 | Calls / HTTP | Duration |
| --- | --- | --- | --- | --- |
| Chat | PASS11/11 | PASS11/11 |22/22|131.53s|
| Responses | PASS11/11 | PASS11/11 |22/22|111.79s|
| Messages | PASS11/11 | PASS11/11 |22/22|108.99s|

[Fixed pre-paid plan](plan.md), [exact scenarios](../../../../ai/live/testdata/glm-candidate/plan.json), [sources](../../../../ai/live/testdata/glm-candidate/sources.json), [six-entry USD snapshot](../../../../ai/live/testdata/glm-candidate/candidate.json), [execution source freeze](execution-source-freeze.json), [archive index](index.json), [ticket12 handoff](handoff-12.md).

Paid source06975520f8f1a76278644dd26d7926bfef24a1d2 and binary87cf6b3ed542e7132ca6b399a74305ea0f0a3928a4e08ad88a09addac75f79d7 were frozen before the Host read only named GLM_KEY. Source/plan manifest837a83c4711dc4914b8110341403c0d6e43cac254b3101f601d6d538b209c907. The delimiter follow-up was merged before freezing; ticket10's later ledger/docs/tests merge occurred only after all paid runs were sealed. Original live reports/ZIPs retain the frozen source.

Each own report captures final request/rawSSE/events/Result/Attempts, raw usage presence, actual request-ID source, Observer, truncation, budgets and USD estimate. Raw captures and ZIPs are durable under root `.evidence/barness-ai-glm/ticket-11/`; tracked reports and Host receipts are alongside this README. All343-file route inventories verify; no capture truncated. Cancellation is intentional partial read, separately marked.

Known reported total estimate is$0.08818744; it is not an invoice or complete charge. Canceled Chat/Responses usage is unreported; Messages remains partial on all11 calls/model because vendor fields are absent. Missing fields remain unknown. Fixed official RMB8/28/2 perM and CFETS6.7220 produce the dated USD snapshot; cache-write uses explicit ordinary-input computation inference. Context1000000 and Messages output128000 are conservative barness bounds, not vendor-exact or load tested. Account/region/permissions/billing stay unverified Host aliases.

Responses5.3's own InputReasoning-object replay now passes;5.2 rawlow/max are safe single-field Host controls, not canonical intensity equality. Chat temperature0/include_usage compare all other final request fields. Messages own signed and unsigned replay pass;5.2 keeps enabled1024/max4096 in full/native/unsigned cases and its default-omission control changes only thinking;5.3 controls unsigned signature and explicitauto. Canonical Messages maps remain null.

The exact captures also replay through the public Client without real keys or network. [Replay plan](replay-plan.md) and [pinned capture index](../../../../ai/live/testdata/glm-candidate/live-captures.json) preserve the originals; Mode=replay has a separate report,0 network requests and22 replay attempts/route. Every final payload, full Result and Observer metadata is compared; only exact Host local IDs/clocks/durations normalize. Provider IDs, raw signatures and tool/result pairing remain exact. Fresh PartialView event snapshots are retained as observations.

```sh
go test -tags live -timeout 1800s ./ai/live -run '^TestGLMCaptureReplayProcess$' -count=1
```

Controlled public child checks cover22 candidate configuration/failure variants, all22 synthetic calls on each route, and10 exact-capture replay variants (three success; wrong/missing/mutated archive, credentials, narrowed entry and altered expected provider-ID/signature failures). Raw child evidence is copied into parent bundles. Original REDs, read-bound rejection and normal/race GREENs are retained in the archive index. Final normal PASS11.36s, race PASS190.50s and live vet clean;32 process variants plus2 assembly cases. Checks and audits are recorded there.
