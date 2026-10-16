# Ticket 12 failure inventory (before implementation)

Baseline: ba68f0cf962941a26307dfd334b80887026dbbdc; dependencies 10/11 resolved.
Authorized seam: Client.Classify / HookedClient.Classify (spec Testing Decisions).
Implementation belongs to the executable Host, not the protocol library. Fixed
source scenes, processed images, annotations, questions, configuration and candidate
catalogs are independent versioned inputs; predictions never read annotation fields.

- Truth or configuration drift: missing/duplicate sample, missing Chinese rationale,
  altered label/box, wrong image bytes/hash/dimensions/orientation, changed questions,
  catalog version/hash, wrong route/model or protocol evidence association.
- Inflated quality: schema success counted as correct, failed/refused/empty/missing
  answers removed, denominator shortened, low existence box hidden, box confidence
  used as existence, Decisions assigned invented boxes, shared subset mixed with
  Jev local tasks, absent scores represented as zero, poor accuracy treated as a
  protocol failure or repaired by changing truth/retrying selected samples.
- Incorrect arithmetic: confusion, false positives/negatives, null vs zero confidence,
  IoU intersection/union, present-target misses, low-existence consumption policy,
  missing/partial usage mistaken for free, estimated cost using another price pin.
- Unbounded spend: >12 calls/attempts, retries, >48 questions/24 image occurrences,
  >8 MiB decoded sources/16 MiB expanded, >30s call, >225s work/240s process including
  15s cleanup, continuation after cancellation or resource exhaustion.
- Missing terminal evidence: HTTP error, malformed/empty/atomic bad answer, refusal,
  version drift, timeouts, cancellation, absent/partial usage, failed observer or
  leaked bodies/permits; unsent samples must remain explicit.
- Authorization failure: implicit credential discovery, foreign route credentials,
  wrong official endpoint/model/budget, missing opt-in/authorization, unsupported
  route, redirect or proxy escape. Missing authorization produces zero-spend NOT_RUN.
- Misleading artifacts: controlled HTTP presented as real vision PASS; missing true
  protocol evidence presented as eligible quality; old candidate promoted after
  catalog changes; missing captures/input/labels/config/source/time/observer/assertions;
  overwritten evidence, traversal/symlink, missing/unlisted files, altered metrics,
  accounting, result, raw wire, image identity, audit or inventory (even resealed).
- Secret exposure: credentials, ambient secrets, host paths/names, response headers,
  vendor error text or image/body content in Observer. Audit must fail closed.

Vertical E2E slices: fixed workflow and hand-worked scores; failure retention and
zero-spend NOT_RUN; process budgets/cancel; independent bundle tamper verification.
Each slice goes red at the public process/Classify seam before its implementation.
