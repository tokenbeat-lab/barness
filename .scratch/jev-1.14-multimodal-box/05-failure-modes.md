# Ticket 05 failure inventory (before implementation)

Baseline: `641fc8475dbede0bff59cb7c04b04982cfdddb13`, current branch `main`.
Ticket 04 is complete; see its verification and sealed evidence. Confirmed public
seam: `Client.Classify` / `HookedClient.Classify`, real runtime and controlled
HTTP Provider. Work in vertical red/green slices; no private-helper unit tests.

- Catalog/capability: Jev state declaration rejected or only metadata widened;
  context-only models accept State values; state-only models require context;
  Decisions accepts a valid local request; instructions/criteria advertised early;
  defaults, builtin catalog/hash or formal support silently change.
- Projection: root/nested object/array null not replaced, ordinary null/string/
  object keys mistaken for images, template-looking objects interpreted, source
  reuse duplicated/merged, question/choice names changed, exact JSON number
  spellings lost, caller data mutated, public pointers interpreted after wrapping.
- Mixed context: business state/images overwritten, context order lost, no-context
  State wrapped, State appearances multiplied by question count, numeric object
  keys and escaped slash/tilde/empty keys treated as array indexes.
- Scope rejection: duplicate target, nonexistent path/source, unused source,
  invalid RFC escape, leading-zero/signed/out-of-range/append array index,
  scalar/non-null/container/identity target, unreferenced root null accepted.
- Capacity: source/single bytes/count, occurrence count/expanded bytes, projected
  State including escaped keys and envelope, native body/Host byte bounds bypassed;
  known oversized projections allocated before rejection; finite inherited limits
  ignored; shared images counted per question or per source instead of occurrence.
- Final native freeze: original pointer plan trusted after Hook; replaced/added/
  deleted/reordered assets or moved State nodes produce stale ImageInputs; map
  iteration controls first-use order; unused assets, substring, object/identity
  keys, question text and unsupported local positions accepted; raw/typed Hook
  payload bypasses budgets, model location restriction or pinned image authority.
- Atomicity/observability: image metadata published on a bad answer, validated
  usage lost, private keys/file hashes/paths/content leak into shared observations;
  final file hashes differ from actual wire, retries regenerate keys/bytes or
  invoke Hook again; 2xx failures retried; permits/bodies not released.
- Evidence: missing fixture hashes, public input, final wire/response, results/
  errors, Attempts/Observer, finite config, assertions, manifest or redaction
  audit; controlled HTTP claimed as live PASS. Real service remains NOT_RUN.

Implementation boundary: Jev context + State value only. Question-local images,
visual criteria/legends and Box remain explicitly unsupported. No new dependency,
public abstraction, catalog promotion or paid API execution is needed.
