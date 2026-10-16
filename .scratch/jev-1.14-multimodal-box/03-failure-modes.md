# Ticket 03 — failure inventory and agreed public seams

Baseline: `1319e1f48d3b01d4f9e63890376dc381a8a5e2fa`.
Written before implementation. Ticket 02 is resolved. The ticket explicitly
authorizes E2Es through `Client.Classify` and `HookedClient.Classify`; these are
the agreed seams. JSON checks cover the returned public metadata contract.

## Identity and ownership

- Hashing Base64 text, a data URL, an image header only, or the original request
  instead of the decoded final file yields the wrong identity.
- Reordering sources instead of occurrences, collapsing duplicate files or
  references, or assuming source and final indices match loses identity.
- Hook replacement/addition/removal/reorder (including image-to-text and
  text-to-image) leaves stale MIME/hash/order or an empty field in text results.
- Mutable caller images, borrowed Hook containers, later results or concurrent
  tenants change an already frozen result; repeated blocks share mutable state.
- Metadata leaks bytes, native resource keys, or hashes into Observer/diagnostics.

## Closed metadata codec

- Missing/null/wrong-type/aliased/unknown/duplicate fields are silently accepted.
- Noncanonical or unsupported MIME, uppercase/short/long/nonhex SHA-256 values
  marshal or unmarshal successfully; failed decoding partially mutates a value.
- Valid PNG/JPEG/WebP records fail to round-trip; text calls serialize the field.

## Final request and failure atomicity

- A Hook sends malformed/truncated/trailing/escaped inline data, bad signature,
  unsupported MIME, external URLs or resource IDs without final validation.
- Hook additions bypass pinned image permission, model limits, occurrence,
  expanded decoded bytes, per-file bytes or complete encoded request budgets.
- Invalid Hook input creates an Attempt or HTTP request, or exposes partial hashes.
- One bad answer after a valid answer, wrong question identity, duplicate answer,
  invalid/duplicate usage, malformed body, HTTP failure, callback failure or
  cancellation exposes Answers or ImageInputs.
- Valid unique reported usage or ResponseModel disappears on an answer failure;
  invalid usage is invented, retries recalculate content, or a 2xx failure replays.
- Valid refusal is mistaken for failure or loses final image metadata.

## Resource lifetime and evidence

- Hashing buffers the decoded image, is unbounded, ignores cancellation, or
  caches image content across calls; rejection leaks body/permit/call resources.
- Evidence omits fixtures/hashes, public input, final HTTP request/response,
  result/error, Attempt/Observer, assertions, finite configuration or audit.
- Controlled HTTP success is mislabeled as real-provider PASS.

Implementation proceeds in vertical E2E slices: final identity, strict codec,
Hook/failure atomicity, then lifecycle/ownership. Existing protocol and resource
E2Es remain regression controls. No post-implementation unit tests are added.
