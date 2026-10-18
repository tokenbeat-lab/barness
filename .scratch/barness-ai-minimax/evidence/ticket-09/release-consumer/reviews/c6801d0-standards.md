# Standards review: recursive release-consumer source

Verdict: CHANGES REQUESTED — two P2 documented-standard findings; no actionable judgment-smell findings. Confidence: high for this pinned delta.

Base: `e3b90a3fc7fdd774a1a4f0faa82ea7278b3182d9`; target: `c6801d0441da589cce29a9d4169d195b54207ac7`. Reviewed `git diff e3b90a3...c6801d0`, retaining the earlier full-base `ecfd4b4` reviews. Applied AGENTS.md, domain/issue-tracker/GLOSSARY/ADR context and all twelve smell heuristics; excluded tooling-enforced concerns.

**[P2] Keep consumed verdicts inside the verified rooted boundary.** `ai/internal/testkit/release/load.go:70–72` closes the pinned case root immediately after inventory verification; lines85–90 then reopen `pidiff.json` through ordinary paths. `decisionsCaptureProblems` and later Bundle consumers likewise receive path strings. Replacing a verified case/ancestor with an outside symlink permits the subsequent read to escape; replacing an artifact can substitute bytes after its hash was verified. This breaches AGENTS.md principles3/4/5 and the preregistered replacement boundary. Consume bounded, expected-hash-verified bytes through the pinned root, or retain immutable verified bytes, including subsequent capture consumers.

**[P2] Verify unique inventory membership, not only file count.** `ai/internal/testkit/release/inventory.go:107–160` increments `files` for every streamed entry and compares only its total with `len(artifacts)`. Directory enumeration is not a snapshot under concurrent rename/growth: a repeated declared name can count twice while another declared file is never visited. Keep a seen-name set, reject repeats and compare the exact enumerated keys with the declaration. AGENTS.md principles5/9 require concurrent integrity and verifiable exact artifacts.

Canonical single-trailing-separator handling preserves raw IDs and rejects lexical aliases. The descriptor-rooted walk, nonblocking regular-file opens, streaming hashes, finite file/entry/depth/byte ceilings and descriptor cleanup otherwise follow the documented boundary. The public producer/consumer E2E retains parent ownership and original failed outcomes.

Selected M3/GLM source files and typed importers are unchanged. Broader L/Jev applicability must remain historical after this Go change; importing original3d3 bundles is diagnostic evidence, not a current suite or release certificate. Original12385PASS/6FAIL/audit0 results remain unchanged.

No edits, tests, environment/credential access, network or vendor calls. Both findings were sent to the parent and sole implementer.
