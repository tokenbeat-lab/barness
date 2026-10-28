# Spec review v17 — issues21/22

Fixed base: `e1f6986677c905b48d80a6b01fc0c0f125427c82`. Reviewed pin: `a80bfe1b6ebe57d71db702ead6a4bc009e66b9be`. Exact command: `git diff e1f6986677c905b48d80a6b01fc0c0f125427c82...a80bfe1b6ebe57d71db702ead6a4bc009e66b9be`.

**New findings:0. Prior v16 P2:closed.**

Issue22 line18 requires “拒绝 worktree、外置目录及符号链接逃逸；仅写自己的私有 bundle/store”. The new guard at `ai/e2e/testdata/process-grant-fixture/main.go:50–61` checks the claim-store root with Lstat, requires a real directory and validates its physical path. Rejection occurs before output allocation(line92) and `livegrant.Claim`(line95), closing SPEC-21-22-001’s external-store write path.

The public helper regression at `tools/verification/tests/test_live_release.py:58–105` covers external/internal symlinks and a regular file, asserting rejection, unchanged target bytes and absent output. Its positive case retains a normal owned directory and sentinel while producing typed claim evidence. Preserved RED shows both symlink cases previously returned0 and changed their synthetic targets. The later targeted GREEN and complete52-test Python result record exit0; all three refusal targets remain unchanged and the owned-directory positive succeeds. No real store was used by this review.

The full E1→current code-file set remains the eleven paths examined in v16. Current-pin delta from `c82a24753cc970265f37377e5ee3e2a975b7fa6a` contains only this helper guard, its public regression and three documents preserving the original v16 failure. The unchanged portions retain v16’s independently reviewed scope; no additional missing/partial requirement, unrequested scope or incorrectly implemented requirement found.

This is an implementation review, not completed frozen ordinary or formal qualification. Existing v8 FAIL, unused authorization bytes and pending stages remain separate facts; global-source changes do not authorize automatic pin rebinding. This review ran no tests, builds, consumers, claims, key/store inspection or vendor calls and edited no repository files.
