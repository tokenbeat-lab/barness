# Ticket 13 independent review

Fixed implementation point: `75ae5907786a7b55893c6b833bdeb34669210ef8`.
Implementation: `527f633`; reviewed corrections: `e633b3f` and `1721024`.
Diff: `git diff 75ae5907786a7b55893c6b833bdeb34669210ef8...1721024`.
Spec: [ticket 13](issues/13-formal-catalog-and-model-specific-support.md).
Standards: AGENTS.md, GLOSSARY.md, docs/agents, relevant ADRs and the
code-review skill's Fowler smell baseline. Independent Standards and Spec
agents ran in parallel, then rechecked their findings without editing source
or using credentials. Unrelated user edits were excluded.

The preparation-only review of `d344ea0` found zero findings on each axis;
its complete original report remains in commit `75ae590`. That was not feature
acceptance. The review below covers the actual implementation and corrections.
Fresh real formal captures, full-suite execution and final acceptance receipts
were pending when the reviewers issued these reports; their outcomes are
recorded separately in [verification](13-verification.md).

## Standards

Initial review: two P2 findings. A conflicting `MODEL` alias could be silently
accepted alongside `model`; a FIFO report could block before validation.
Both reproduced through public commands before the fixes in `e633b3f`.
Final follow-up (lightly cleaned):

> Standards follow-up: zero remaining findings.
>
> Both prior issues are fixed and independently rechecked: conflicting `MODEL`
> alias now rejects with exit 1; FIFO report immediately rejects with exit 1;
> the matching controlled bundle still verifies with exit 0.
>
> Reviewed exact field validation, captured-response replay and mandatory
> scenario scope checks. No additional documented-standard breaches or
> actionable baseline smells found. Source files and credentials remained
> untouched; no real calls were made.

## Spec

Initial review: one P2 finding. Correlated corruption of the native Box and
saved public result could pass structural equality without runtime decoding.
Public command E2E reproduced it before correction. An additional adversarial
producer check found that a valid Instructions-only payload could retain the
State/Instructions scenario name; the independent mandatory scope contract now
rejects it. Final follow-up (lightly cleaned):

> Zero remaining actionable Spec findings.
>
> The prior invalid-Box finding is resolved: the subordinate verifier validates
> normalized coordinates and confidence, and `ReplayProtocol` decodes captured
> responses through public `Classify` and compares the atomic public result.
> This satisfies R12's explicit finite normalized positive-area coordinates and
> R14's atomic malformed-answer rejection.
>
> Required scenario validation independently checks source counts, image
> placements, question sets and kinds, addressing ticket 13's mandatory
> capability requirement. An additional correlated visual-Bool-to-Choice
> tamper probe was rejected by public replay.
>
> Model-specific identity remains separate from 1.13; controlled evidence
> cannot merge as live support; source/catalog seals and historical rows remain
> protected. No scope creep found.

## Reproducible review checks

- Public command red: `.evidence/barness-ai/20261016T224551.586165000Z`
  (alias, FIFO and correlated invalid Box).
- Direct structural Box red: `.evidence/barness-ai/20261016T224756.033975000Z`.
- Relabelled mandatory State-scope red:
  `.evidence/barness-ai/20261016T225150.274777000Z`.
- Source-matching controlled green:
  `.evidence/barness-ai/20261016T225211.564576000Z/formal-support-package`.
- Independent reviewer confirmed ordinary green verification and exact alias/
  FIFO rejections. No provider calls were made by either review agent.

Standards: 0 remaining findings; Spec: 0 remaining actionable findings.

Final regression delta `e633b3f...1721024` was independently reviewed on both
axes: zero code findings. Current configuration revisions and paired consumer
hashes agree; obsolete candidate injection is deleted; historical raw packages
and matrix records are preserved. The Spec reviewer flagged that the new local
receipt helper trusted summary counters after verifying the support bundle.
It now compares the entire summary with the verified raw protocol/effects
reports and derives all catalog/call/budget/metric fields from those reports.
This helper correction is exercised by the final sealing/recheck commands.

The Spec reviewer independently rechecked the receipt-helper correction and
confirmed it resolved; final result remains zero findings on each axis.

Final acceptance work is now complete: full suite, both focused race runs and
vet passed at `1721024`; fresh formal-Hash protocol and independent effect
packages passed strict verification, and only the new Jev row merged. The final
receipt's public seal/recheck commands passed. Its stored-byte checks account
for the existing recorder's pre-redaction legacy fixture fingerprints; new
UC13 fixtures match both identities. No runtime/source pin changed after the
fresh real runs. Exact outcomes and artifact links are in 13-verification.md.
