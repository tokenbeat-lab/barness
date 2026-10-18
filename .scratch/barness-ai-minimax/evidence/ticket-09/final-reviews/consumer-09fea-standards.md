# Standards review: authenticated release consumption

Verdict: CHANGES REQUESTED — both original P2 findings are implemented as closed; one new P2 remains. No actionable judgment-smell findings. Confidence: high.

Base: `c6801d0441da589cce29a9d4169d195b54207ac7`; target: `09fea1b748718795bacc531f6c4037ebe66ba3a9`. Reviewed `git diff c6801d0...09fea1b` and the release artifact consumption chain, reusing earlier full-base reviews. Applied AGENTS.md, domain/issue-tracker/GLOSSARY/ADR context and all twelve smell heuristics; tooling-enforced concerns excluded.

**[P2] Bind the opened bundle root to its pre-open identity.** `ai/internal/testkit/release/verified_case.go:38–44` compares only the two path Lstat identities; it never compares the opened root's `info` with `before`. An ABA replacement can substitute an outside/replacement directory during `os.OpenRoot`, then restore the original path before the second Lstat. The function accepts a different opened inode and can inventory files outside the originally selected bundle. Later reopen checks preserve that wrong initial identity rather than detecting the initial mismatch. Require a non-symlink directory and `SameFile(before, info, after)`, or explicitly pin the resolved target before opening if initial symlinks must be supported. This breaches AGENTS.md principles4/5 and the preregistered rooted replacement boundary.

Original finding1 is addressed: private verifiedCase replaces unchecked caseDirs across differential, decisions, load, multimodal, revocation and generic live artifact consumers. Consumed bytes use rooted bounded reads and expected artifact hashes; same-semantics JSON edits cannot borrow the original inventory verdict. Descriptor closure and the existing4/16/64MiB budgets are retained without long-lived handles.

Original finding2 is addressed: repeated directory names fail before visitation, successful hashes populate a distinct file-name set, and every declared key must be present. Entry/depth/file/aggregate limits remain finite. The isolated finite iterator test has the explicit repository-required exception and failure inventory; the public post-load E2E covers symlink and hash mutation cases.

The shared private abstraction serves multiple existing consumers without a compatibility shim. M3/Q typed importers and producer/raw evidence remain unchanged. Broad L/Jev source applicability is historical; original3d3 executions are not refreshed or claimed as current commands.

No edits, tests, environment/credential access, network or vendor calls. The remaining finding was sent to the parent and sole implementer; original c680 reports remain unchanged.
