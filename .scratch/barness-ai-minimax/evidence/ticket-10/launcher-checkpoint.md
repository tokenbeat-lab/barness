# Launcher checkpoint

The public CLI and fixed inventory are in `../../tools/README.md`, `paid.py`,
`paid_core.py` and `paid_plan.py`; repeatable controlled executable checks are
`../../tools/test_paid.py`. All files are below500 lines.

Failure inventory preceded implementation: `launcher-failures.md`. Successive
red/green logs cover missing executable/credential-free first slice, untrusted
report redaction, identity budget aggregation, interrupted partial report,
stale output refusal and honest credential acquisition receipts. Final `launcher-acquisition-green.log` records all9 controlled
executable checks passing. Tests use synthetic selected credentials and fake
subprocesses in temporary Git repositories. Real `.env`/credentials were not
read; no paid or vendor request was made. Checkpoint paid consumption is zero.

Parent must review exact source, finish all Go/config/catalog work and perform
its required two-axis review before preparing/approving actual collection.
The tool writes only a PENDING template; it cannot self-authorize. Its optional
research/fullcandidate and formal preparations are distinct snapshots while the
common-root stage ledgers/intents persist across them.

Remaining execution responsibilities: parent-owned real preflight/source review,
actual bounded route invocations, capture validators, Jev support derivation,
exact17input merge, actual release gate plus fresh report verdict verification.
