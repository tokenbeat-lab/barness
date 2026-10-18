# Paid launcher failure inventory (before implementation)

Public auxiliary seam: executable CLI `tools/paid.py` (T01). All processes in
controlled tests use an isolated synthetic Git repository and fake Go/Python
executables; no vendor endpoints or real credential file are touched.

1. Preparation, no-key checks, previews and summary must never open credentials.
2. Missing/stale/non-APPROVED/exact-source review refuses before acquiring a key.
3. Uncommitted, added, removed or changed execution/config/script source refuses;
   changed built binary refuses, including source changes after a child exits.
4. One selected credential only reaches a child; ambient keys/proxy/endpoints are
   excluded. Selected key, repository and HOME are redacted from output/receipts.
5. Exclusive common-root intent precedes secret acquisition. Changed output
   directories/worktrees cannot restart a route; interrupted/unknown consumes
   the original full cap. No automatic process retries are available.
6. MiniMax uses exactly one route/model/purpose and canonical common-root stage
   ledger. Research (3 fixed cases), full candidate (13 fixed cases + at most
   one tool rerun) and formal remain distinct. M3.1 requires independent M Plan
   review and credential; it cannot be formal or inherit M3 authorization.
7. Legacy selectors/budgets remain 302/302 + GLM72/96 = 374/398, retaining all
   four effects. NOT_RUN differs from missing/invalid report UNKNOWN; current
   report counts and cumulative MiniMax budget are reported separately.
8. Collection resolution supplies exactly 17 own inputs (10 generic, Jev1.14,
   GLM3, MiniMax3). Missing/ambiguous bundles refuse; keyless verification,
   merge and real gate command plans retain four effects and all 17 inputs.
9. Child timeout/interruption/audit failure cannot publish success. A report is
   optional for a failed process; absence cannot borrow a stale report.

Parent owns exact-source approval; this executable never creates APPROVED.
