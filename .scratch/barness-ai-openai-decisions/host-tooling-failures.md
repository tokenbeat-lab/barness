# Issue 11 failure-first plan

Agreed seams: OD55 CLI/process + serialized bundle verifier; OD56 existing local/cloud Host through public Client.Classify. Fixed baseline efe6d198e301e96e4ccfcd0226d0f3c17eff00a8.

Before implementation, process E2E covers mismatched operation/provider/API/model route, wrong catalog Hash, excess attempts, old Jev default/schema, normal/mixed/all refusal, absent/partial usage, budget exhaustion and cancellation with no subsequent request. Import counterexamples cover changed route/code Hash/counters/authorized model/evidence kind, missing Observer and non-null all-refusal metrics. Historical Jev uses its recorded snapshot, while current Plan rejects its old Hash for applicability.

Red run: .evidence/barness-ai/20261013T015612.722875000Z, exits 1. New route/fixture flags are absent; historical verification fails against current catalog. Existing Jev default passes.

Implementation boundaries: host-owned route/config, fixture transport, task report conversion, descriptive metrics, artifact import. Library adapters and release gate stay outside issue 11. Schema 1 retains Jev meanings; schema 2 explicitly records route and source provenance, per-sample refusal, independent protocol/task/quality conclusions. Fixed dataset bytes remain unchanged. No business threshold or fallback.

Review counterexamples were added before fixes: coherently rehashed question/answer captures; missing/null/negative/excess Decisions reasoning and chat cache; coherent fixture-to-live identity; missing actual captures; dual request-ID headers; zero-Attempt cancellation; HTTP failure capture. All now pass at process/public artifact boundaries. Source red bundles are preserved in host-tooling-evidence/failure-iterations.zip.

Final target run: 97 PASS, audit clear. Full normal run with frozen Pi differential: 7,356 PASS, 187 NOT_RUN, one existing timeout ordering failure, preserved separately. Original single timeout case rerun exits 0; no assertion or protocol change was made. Follow-up is tracked in .scratch/barness-ai-timeout-ordering/issues/01-call-and-read-idle-deadline-order.md. Full race and final verification summary are recorded in host-tooling-evidence/README.md.
