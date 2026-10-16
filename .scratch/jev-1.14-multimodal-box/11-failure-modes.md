# Ticket 11 failure inventory (before implementation)

Baseline: `80e8d36`; dependencies 07 and 08 are resolved. Approved seams from
the ticket: Client.Classify / HookedClient.Classify, isolated probe process,
and its independently verified on-disk evidence package. No unit tests are added.

- Configuration: switch/key/proof/account/region absent; malformed or mismatched
  authorization; foreign/ambient credential; endpoint override; wrong route,
  service, model, candidate version/hash; changed sources or narrowed scenarios.
  Missing prerequisites produce NOT_RUN with zero spend; supplied invalid
  configuration fails closed. No credentials from other jobs or files are read.
- Reservation: calls/attempts/questions/sources/appearances/source bytes/expanded
  bytes/State exceed their cumulative limits; body exceeds 4 MiB; overflow,
  failed/canceled/unreported calls incorrectly refunded; Hook adds unreserved
  resources; retries or redirects exceed reserved attempts; process or call
  deadline is exceeded; finalization has less than its reserved ten seconds.
- Mapping: default versus explicit context changes semantics; [1,0,1] order or
  reuse lost; State wrapper overwrites content; local instruction/criteria image
  scope lost; Choice/Bool/Score standards or ordered legend changed; private key
  or bytes leak; final Hook image identity differs from result; Box and existence
  conflated, nullable confidence required from a nondeterministic real model.
- Protocol: wrong final URL/model/experimental header, malformed/partial 2xx,
  invalid answers/legend/Box, missing actual ResponseModel, usage or Attempt/
  Observer mismatch, provider 429/5xx/timeout, leaked resources. Preserve the
  failure and all reserved versus actual counters, never retry to obtain PASS.
- Evidence: missing/changed/extra files, symlink/path escape, bad manifest or
  audit, failed writes, crash before publication, stale accepting report,
  duplicate IDs, missing assertions, controlled results promoted to live PASS,
  another route/old Jev evidence substituted, repeated run overwrites history,
  source/candidate change reuses old hash or timestamp. Formal catalog/matrix
  remain unchanged; quality/capacity/billing/region unknowns stay explicit.

Real probes use small valid PNG/JPEG, not destructive format/permission probes.
Existing public E2E covers those invalid inputs. Each run has a new directory;
both routes own their candidate, credential, raw evidence and verdict.
