# Standards review: pinned initial bundle selection

Verdict: PASS — all three reported consumer P2 findings are closed; zero open documented-standard or judgment-smell findings. Confidence: high for this pinned source delta and retained consumption chain.

Base: `09fea1b748718795bacc531f6c4037ebe66ba3a9`; target: `7754c7fe9c71c1203a0c9f860da35dbf5973c3d9`. Reviewed `git diff 09fea1b...7754c7f`, reusing previous consumer/full-base reviews. Applied AGENTS.md, domain/issue-tracker/GLOSSARY/ADR context and all twelve smell heuristics; tooling-enforced concerns excluded.

`verified_case.go:29–47` normalizes only the caller-selected root spelling, rejects a non-directory/symlink root, and compares the pre-open, opened-root and post-open inode identities. Mismatch/error paths close the opened root. This closes the initial ABA replacement gap; trailing-slash spellings cannot hide a selected root symlink.

`load.go:28–44` and `live_manifest.go:23–41` now pin that root before reading their initial manifest. `readBundleManifest` uses the same rooted non-symlink regular-file reader and actual64MiB ceiling. The deliberately absent expected hash applies only to this initial declaration; validated case inventory and all subsequent verifiedCase reads still require a nonempty expected artifact hash. No unchecked case-path consumer or long-lived handle is reintroduced.

The earlier closures remain intact: every consumed artifact is read under the pinned bundle identity with bounded bytes and hash authentication; duplicate streamed names are rejected and every declared file must be successfully hashed. Parent/nested ownership, raw identifiers and exact inventory limits remain unchanged. Failure-first public root-selection scenarios cover real directories, direct symlinks and trailing-slash symlinks, alongside the earlier mutation and finite-iterator cases.

The private abstraction serves real existing release consumers; no compatibility shim or speculative hook was added. This delta changes release/e2e source only. Typed M3/Q importers and selected sources remain unchanged; broader L/Jev evidence retains historical applicability. Original A's failed load, suite outcomes and missing release verdict are not upgraded by a successful diagnostic import.

Read-only review: no repository edits, tests, environment/credential access, network or vendor calls. Final normal/race/vet/actual-import executions remain their own evidence; this review does not assert their completion. Earlier review files are unchanged.
