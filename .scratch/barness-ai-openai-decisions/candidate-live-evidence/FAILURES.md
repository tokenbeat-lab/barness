# Issue 09 — failures before implementation

Confirmed seams from issue 09: a public Client assembled with a fixed host
candidate Catalog, and the live test executable as an isolated process. Network
fixtures replace only the external HTTP Provider; source mutations replace only
filesystem inputs. No production endpoint override is added.

- Missing key/switch: all mandatory scenarios NOT_RUN; no vendor fallback.
- Foreign combo key, endpoint override, missing account/region/permission proof,
  non-local proxy: config FAIL, zero reserved calls and zero HTTP attempts.
- Missing/altered source or candidate: config FAIL before any paid request.
- Candidate must remain absent from BuiltinCatalog, and never produce the formal
  live-report.json. Own source dates/hashes and candidate version/hash are saved.
- Normal text choice/score/bool, mixed, real PNG/JPEG/WebP, in-flight cancellation
  and recovery require own captures, typed results, usage presence, actual model,
  attempt/vendor IDs, attribution, Observer terminal records, and asset hashes.
- Refusal is a typed valid result; no request is repeated to manufacture it.
- Missing/partial/zero usage differ; protocol completeness is independent.
- 429/504, malformed 2xx or precision counterexample: FAIL with original capture,
  zero retries; vendor failures never become UNSUPPORTED.
- Reserve cumulative calls/attempt slots/questions/state/decoded images before
  every call, including cancellation and failures; overspend is FAIL/budget.
- Narrowed execution cannot become complete. Every missing mandatory scenario
  remains NOT_RUN. Audit failure prevents acceptance; unknown billing/cache/
  regional premium remains unknown. Context and service maxima are not probed.

Evidence is produced by `go test -tags live ./ai/live -run
'^TestDecisionsCandidateProcess$' -count=1`; each child bundle includes manifest,
assertions, replay, report and audit. Controlled evidence cannot grant live support.
