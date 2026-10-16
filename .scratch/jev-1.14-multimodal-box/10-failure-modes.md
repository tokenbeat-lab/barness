# 10 — Host unified image example failure analysis

Recorded before implementation. Test seam already authorized by ticket 10:
public Client.Classify / HookedClient.Classify, exercised by the executable Host
and checked from an external E2E process. No private library seam or post-hoc
unit tests. First red may be absence of the new executable; retain that honestly.

## Failures to cover

- Default execution requires credentials, contacts a remote endpoint, silently
  discovers another key, or reports controlled HTTP as real-service PASS.
- PNG/JPEG fixtures are illegal, nondeterministic, missing, or have mismatched
  MIME, decoded dimensions, hashes, orientation or crop provenance.
- Changing protocol requires rewriting common context input, native assets,
  private keys or data URLs. Explicit reuse/order is lost or budgets only count
  unique sources instead of occurrences.
- State, Instructions or Criteria images leak to another question, or Decisions
  silently lifts unsupported local inputs into global context. Box is silently
  dropped, rerouted, or treated as supported by the built-in Jev catalog.
- Host interprets input indices as final indices; Hook MIME/content/order
  changes leave stale legends, search-image attribution or transformations.
- Score legend references are ignored, interpreted relative to the request,
  resolved to bytes/private keys, or point outside final ImageInputs.
- Box confidence null becomes zero/unset; a separate zero existence answer
  suppresses a valid Box; a refusal becomes false or a fabricated ordinary answer.
- Pixels use the source dimensions before orientation/crop, or infer which of
  multiple images was searched. Example existence thresholds become library defaults.
- Invalid Box/legend publishes partial answers or ImageInputs, discards known
  usage, or retries a successful HTTP response.
- Image permission / binding revocation still sends, fallback occurs, expanded
  budgets are bypassed, experiment-off sends, or another tenant can resolve a key.
- Cancellation sends a new request or retains resources; observer completion is
  lost, records image content, or races with the evidence writer.
- Real mode starts without explicit opt-in/config/named key source, accepts an
  unrelated protocol, persists credentials, or labels unverified output PASS.
- Evidence omits complete input, final native request/response, results/errors,
  Attempts/Observer, configuration, fixture/image hashes, assertions or audit;
  inventory misses files or a repeat run overwrites the original evidence.
- Docs widen support/production claims, change parent spec, or promote proposed
  ADRs. Existing user changes are accidentally included in the commit.

## Intended architecture

A standalone example command under ai/examples/classifierimages owns synthetic
fixtures, Host assembly, result views and evidence. It calls the existing public
Client without modifying runtime/catalog contracts. Controlled Provider and
redaction audit are reused from ai/internal/testkit. Per-call immutable captures
and a synchronized resolver/observer preserve tenant boundaries and revocation.
Real mode is a single explicit Host-configured call using the same input builder.
