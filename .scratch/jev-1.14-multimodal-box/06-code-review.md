# Ticket 06 two-axis review

Fixed point: `4adf8c2b4786a083e719cef57032f5b8e3cff9b3` (start of work).
Initial implementation: `60f33a8`; final implementation: `5449e59`.
The implement skill's required code-review skill ran independent parallel,
read-only Standards and Spec sub-agents. Both reviewed the correction separately.
Existing user README/GLOSSARY edits and future specification files stayed outside
the committed diff. The comparison and commit list were:

```sh
git diff 4adf8c2b4786a083e719cef57032f5b8e3cff9b3...5449e59
git log 4adf8c2b4786a083e719cef57032f5b8e3cff9b3..5449e59 --oneline
```

## Standards

No findings. Reviewed the committed diff against AGENTS.md, domain/issue-tracker
conventions, relevant ADRs, and the supplied smell baseline. No concrete hard
violations or actionable judgment smells found. Public/native boundaries,
owned snapshots, shared projection behavior, finite capacity checks, E2E evidence,
and documented rollback remain consistent with repository standards.

Follow-up: no remaining Standards findings in `60f33a8...5449e59`. The fix
preserves explicit envelope topology, documents the nil-versus-empty distinction,
and adds public E2E coverage for empty, missing, and null image lists.

## Spec

Initial review found 1 P2: removing the last context occurrence wrongly rejected
valid remaining Instructions images. R10 requires “Hook 后重建题目、图片出现、MIME、
预算和答案依据”; the failure inventory includes moving/replacing/adding/removing
images. A Hook preserving the State envelope and setting images to [] left an
asset fully used by Instructions, but the scanner rejected the empty context
array before reaching Instructions.

The parent first reproduced the failure through public Classify E2E, then
corrected it in `5449e59`. Explicit empty arrays preserve builder provenance and
continue scanning local uses; missing/null envelope arrays remain invalid.
`06-logs/review-context-red.txt` and `review-context-green.txt` retain the
reproduction and related passing regressions. The independent follow-up review
confirmed the original P2 resolved and no remaining Spec findings.

No additional concrete defects or scope creep were found. Root-image format
handling, native ownership, deterministic traversal, authorization, capacity
checks and atomic publication otherwise match ticket 06.

Standards: 0 findings / 0 remaining. Spec: 1 P2 corrected / 0 remaining.
