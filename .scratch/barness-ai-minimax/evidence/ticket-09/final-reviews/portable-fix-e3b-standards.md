# Standards review: portable retention gzip correction

Verdict: PASS — prior P2 closed; zero open documented-standard or judgment-smell findings. Confidence: high for the pinned fix.

Base: `3d3a1fb28e143a2c29d684f78edcad94c0217243`.
Target: `e3b90a3fc7fdd774a1a4f0faa82ea7278b3182d9`.
Comparison: `git diff 3d3a1fb...e3b90a3`.

Applied AGENTS.md, retained domain/issue-tracker/GLOSSARY/ADR standards and all twelve smell heuristics; tooling-enforced concerns excluded.

At `original-preflight-source-proof/verify_q.py:39–46` and `actual-corrective-1/verify_actual.py:147–155`, each inventory reader now reads at most8MiB+1 compressed bytes, rejects compressed overflow, authenticates that same immutable buffer, then reads at most32MiB+1 expanded bytes and rejects overflow before JSON parsing. Concatenated streams share the same finite expansion ceiling. This closes the allocation-before-limit and Q authentication-order defects without changing production limits, runtime or launcher code.

The retained public CLI evidence distinguishes six passing current refusal checks from historical red results. Genuine old/current positive outputs differ only by one added bound check; the original failed preflight and PARTIAL actual qualification remain unchanged. The resource calibration and intermediate failures are explicitly preserved rather than relabeled.

Independent read-only byte comparison confirmed the nine-file historical archive matches original Git bytes, all seven original parent metadata files are unchanged, and both corrected verifier hashes/lengths match `retention-bounds/current-verifier-index.json`. That index and its README explicitly distinguish current program provenance from historical parent records, satisfying AGENTS.md principles8/9 without rewriting original evidence.

The other new3d3 portable verifier sites contain no further eager gzip inventory decoder. The optional uncommitted closed-L reader inspected separately uses a128MiB+1 gzip read and checks its limit before JSON; this is a narrow same-family observation, not approval of that unpinned capsule. No actionable abstraction or compatibility-smell finding applies to immutable historical proof.

All1,034 frozen generation files remain unchanged. Actual acceptance blockers and the ongoing release gate are unaffected. No merge, repository writes, tests, environment/credential access, network or vendor calls were performed.
