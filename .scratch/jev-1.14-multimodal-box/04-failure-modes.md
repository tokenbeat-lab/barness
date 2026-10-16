# Ticket 04 failure inventory (before implementation)

Baseline: 490e2f901d711ded59df97f4156fa6ffb43d47f5, current branch main.
Confirmed seam: Client.Classify / HookedClient.Classify, real runtime with
controlled loopback HTTP. Implement in vertical red/green slices. No unit tests.

- Candidate catalog: explicit jev-1.14.0 text+image/context PNG/JPEG rejected,
  missing/duplicate/unsupported locations or MIME accepted; aliases accidentally
  gain capabilities; builtin catalog/hash/support matrix change.
- Projection: sources lost/merged, reference order [1,0,1] changed, repeated use
  uploaded twice, same bytes in distinct sources merged; State state/images
  fields or exact numbers overwritten; questions rewritten; text wire changes.
- Private references: collision with decoded original strings/object keys,
  mutual substring keys, unbounded retry, key limits exceeded, keys regenerated
  by transport retry, private keys leaked into answers/Observer.
- Preflight: missing policy/capability/placement/MIME/count/Binding permission/
  experiment reaches credentials, hooks, admission or HTTP; unsupported local
  references or future Box accepted. Experiment option alone grants authority.
- Files: malformed Base64, mismatched MIME/signature, data URL, APNG, truncated
  PNG chunks, invalid dimensions, edge <256 or >1568, area <78400 or >2458624,
  decoded file >16 MiB; stricter Host single/source/reference/body limits ignored.
  Jev checks leak into Decisions WebP or existing small-image behavior.
- Headers: experiment absent/wrong/duplicated/case-aliased/contradictory;
  Hook adds experiment to an unselected or unauthorized/text-only call, removes
  selected experiment, injects Long Context, changes auth/model/target.
- Final native freeze: invalid/unused assets, unknown fields, URL/part/data URL,
  substring references, overlapping or confusing keys, object/identity keys,
  misplaced images in state/Instructions/criteria accepted; legal asset add/
  delete/replace/reorder rejected; final limits or pinned authority bypassed
  through RawMessage/typed/custom JSON. Context-only subset remains enforced.
- Result: final deterministic first-use order lost, same asset duplicated,
  distinct assets merged, replaced bytes hashed incorrectly; partial Answers or
  ImageInputs on bad answer/usage/cancel/HTTP/response-hook failure; valid usage
  and response model lost, input charged twice, missing usage fabricated.
- Lifecycle: hook/credentials called per retry, bytes/headers/keys change, 2xx
  decode retries, cancellation or deadline leaks bodies/permits; mutable Hook
  containers alias final validation, tenants/calls share assets or metadata.
- Evidence: missing synthetic images/hashes/public input/final wire/scripts/
  result/error/Attempt/Observer/config/release assertions/manifest/audit;
  controlled HTTP reported as real-service PASS. Live stays NOT_RUN.

Jev native subset in this ticket: assets referenced only as whole string values
in the projected context array state.images. No State value, question-local
images, visual criteria/legends or Box; later tickets own those capabilities.
