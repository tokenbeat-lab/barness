# 09 — mixed image runtime failure inventory

Baseline: `58ec73c`; dependencies 07 and 08 are resolved. Written before
implementation. The already approved seams are Client.Classify and
HookedClient.Classify with the controlled HTTP Provider. No live-service or
visual-quality claim follows from these tests.

Failure modes to exercise, using the same contract in normal and race runs:

1. Same-named Binding, source indices and question names leak images, credentials,
   private keys, answers, usage or prices between concurrent tenants/protocols.
2. Repeated context, State/local Instructions/Criteria, four Jev answer kinds or
   three Decisions kinds disappear from the declared finite load.
3. Caller/config/catalog/policy/options, Hook-borrowed trees or returned answers
   alias a pinned request, final image identity, legend or price snapshot.
4. Source count, single/total decoded bytes, reference count/expanded bytes,
   Base64/data-URL/JSON framing, projected documents or output budgets are
   bypassed at entry, construction or final Hook boundaries. Known violations
   allocate encoded copies before rejection. The initial tracer targets Hook
   reference amplification before serialization, including raw and typed trees.
5. Cancel/deadline before sending, while reading, waiting for process/tenant/Host
   admission or sleeping for retry sends again, leaks a body/permit/waiter, changes
   another call's lease or omits an Observer terminal.
6. Default retry is nonzero, explicit retry is unbounded, callbacks run per
   Attempt, or retry changes bytes, headers/private keys, credentials, binding or
   catalog. Successful HTTP with malformed answers is replayed or loses usage.
7. Disabling Binding, removing its model or revoking images affects a pinned
   call, fails to deny a fresh call or falls back to text/another Provider.
8. Observer publishes image bytes, coordinates, private key, Pointer, hash,
   credential or upstream error text; failed Attempt history or legal usage is
   lost. Observer failure/congestion changes the call or blocks indefinitely.
9. Memory/resource records are timing-only or omit recorder overhead; failures
   are hidden by reruns, reduced load or larger limits. Every load barrier must
   prove the expected active HTTP bodies/permits and finish with zero resources.
10. Evidence omits exact fixture/image hashes, inputs, final requests/responses,
    results/errors, Attempts/Observer, assertions/configuration or audit/inventory;
    evidence verifier cannot detect corruption or offline PASS is called live PASS.

Design load: one shared Client, two tenants, two protocols, four simultaneous
calls (one of each protocol per tenant), two distinct 280×280 PNG/JPEG files per
call, context order [1,0,1], plus Jev State, Instructions and Criteria uses of
those files; four Jev questions (Box/Bool/Choice/Score), three Decisions questions.
Fixed 64 KiB request/16 KiB output, 8 references, 64 KiB source/128 KiB expanded
decoded budgets; four process/two tenant permits. Memory ceiling: 64 MiB live
growth and 64 MiB allocation for this bounded batch, including HTTP recording.
These are test policy values, not deployment defaults. Large rejected borrowed
Hook values have a separate sub-512 KiB post-callback allocation ceiling.

Only boundaries confirmed by these E2E failures may be repaired. Existing
release defects retain their own issue identity/status.
