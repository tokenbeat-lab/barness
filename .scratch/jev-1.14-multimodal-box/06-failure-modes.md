# Ticket 06 failure inventory — before implementation

Baseline: `4adf8c2b4786a083e719cef57032f5b8e3cff9b3`; dependency 05 resolved.
Agreed seams: public `Client.Classify` and `HookedClient.Classify`, using the
existing controlled HTTP Provider, Host resolver and evidence recorder.

- Projection: choice/score/bool root images, nested objects/arrays and ordered
  text/image fragments may fail, fill the wrong slot, wrap State unnecessarily,
  leak local images into State/context/another question, alter ordinary nulls,
  text, keys, criteria or exact numeric spellings, mutate borrowed input, or
  upload repeated sources multiple times. Escaped, empty, numeric and Unicode
  keys must resolve literally; reference-list order must not determine final
  ImageInputs order.
- Public validation: missing/non-null/container targets, duplicate targets,
  malformed RFC escapes, noncanonical/out-of-range array indexes, missing
  questions, unused/out-of-range sources and protocol/identity/container fields
  must fail before credentials/Attempt. Unreferenced null/empty/missing
  instructions remain invalid; an image counts as nonempty. Duplicate JSON keys
  cannot make an ambiguous slot valid.
- Capabilities: root images must not require a fictitious null text format;
  object/array containers still require their DescriptionFormats. Missing local
  location, MIME/input capabilities, policy, experiment, Binding authorization,
  tenant/actor identity must not gain authority. Decisions rejects the same
  otherwise-valid local input before credentials with no conversion or split.
  Criteria and Box stay unimplemented, including catalog declarations.
- Capacity: source count/bytes, single file, per-occurrence count/decoded bytes,
  projected full question bytes (including escaped generated keys), State and
  complete body apply independently before known expansion. Reuse in two
  questions counts twice even though uploaded once. Hook final bytes, source,
  occurrence and expanded limits must be rechecked before Attempt.
- Final Hook: legal renaming/retyping/reordering questions, moving/replacing/
  adding/removing images, including from text input, must use final independent
  native questions and deterministic State then sorted-question instruction
  order. Empty/null/malformed final instructions, substring/object-key/private
  identity/criterion uses, unused/overlapping keys, URL/part/non-image asset and
  unauthorized location/MIME/count/header/model/target/credential/experiment
  changes must fail. No stale public pointer/index or borrowed Hook state may
  determine results or image metadata.
- Results/runtime: invalid/stale/missing/extra answers fail atomically with no
  Answers/ImageInputs; valid usage survives, no private key/bytes leaks to
  results or shared Observer/Attempt metadata. Retries reuse frozen bytes,
  images and headers, never replay a bad 2xx; cancellation/response/body/admission
  release, snapshot isolation and no-image/context/State/Decisions regressions
  remain intact.
- Evidence: fixture/image decoded SHA-256, input/config, final request/response,
  result/error, Attempt/Observer, assertions, resource release, inventory hashes
  and zero-finding redaction audit must be replayable. Controlled HTTP PASS
  cannot imply real-service, visual-quality, built-in catalog or formal support.

Implementation proceeds in vertical red/green slices: projection; validation
and capacity; final Hook/result/runtime. No implementation-afterward unit tests.
