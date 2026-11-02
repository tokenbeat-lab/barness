# Reviewed catalog maintenance source

`providers/*.json` is the unique maintained definition source, grouped by actual
serving Provider. `maintenance.BuiltinState()` reads independent values;
`maintenance.Compose()` orders selected Provider identities before deriving the
composition version, and preserves each operation's explicit entry positions.
`maintenance.Generate()` emits the same runtime envelope for local files and
`ai/catalog_builtin.json`.

The initial records preserve all 118 definitions (112 chat, 3 image and 3
classifier) from the immutable `ai/release/catalog-snapshot.json`, version
`2026-10-29.1`, catalog hash
`sha256:f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d`.
This is historical catalog provenance, not a new models.dev capture. The six
image/classifier records retain an explicit missing serving-source marker;
none borrows another operation's prices. Historical field ownership retains
original accepted values until a new explicit field decision replaces them.

The initial Provider versions are 1. Their sorted versions and active runtime
content determine a new composition identity. No new identity is presented as
the old version/hash. Withdrawn entries remain in State for future recovery but
are omitted from runtime output. Provenance and field-decision edits alone do
not change a Provider's runtime content digest.

The first adopted baseline adds only
`chat/openai/openai-responses/gpt-6.1-terra`: 113 chat, 3 image and 3 classifier
records. OpenAI's content version is 2; all other Provider versions remain 1.
The complete `baseline/review.json` inventory retains the approved revision,
official source capture, full manual definition, field decisions and original
historical catalog bytes. Its `snapshot.json` is adopted without regeneration.
The prior 118 runtime definitions and their relative order are unchanged.
This is code compatibility acceptance, not new live model qualification.

`go generate ./ai` verifies the embedded artifact against reviewed Provider source without writing. Explicit generation selects `cataloggen --builtin|--state|--store`, previews with `--dry-run`, and writes only with `--confirm`.

Local apply changes only the selected store/output. Use `catalogctl export --store <store> --repo <repo> --review-dir <NEW directory> --confirm` to retain one approved revision as Provider files, State, exact runtime snapshot and deduplicated source/manual captures. `review.json` is the completion marker and file inventory; its external store/check references remain original evidence. Adopt exported `providers/*.json`, `snapshot.json` as `ai/catalog_builtin.json`, and the complete review directory as `catalog/data/baseline/` together. Replace the prior baseline inventory, including all input blobs and review.json. Maintenance validates exact files, digests, provenance and the selected Provider/runtime pair, then retains the baseline in the seed revision for future maintenance/export cycles. Inherited references carry the original index digest as `originReview`; this does not assert availability or current qualification of old external evidence. An absent store derives the same runtime Version/Hash; its maintenance revision ID can differ by lineage. Never replace the runtime artifact alone.
Runtime `ai` imports neither the maintenance package nor these source files.
Catalog changes use the candidate/check/apply maintenance workflow; code
compatibility checks do not grant vendor support, tenant permission or release
qualification. Historical activation wrappers remain under
`ai/release/catalog-activations/`; the former runtime activation writers have
been deleted.
