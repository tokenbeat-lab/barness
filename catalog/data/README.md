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

The historical 2026-11-03 adopted baseline explicitly withdraws 37 OpenAI chat identities.
Only the existing GPT-5.6, GPT-6 and GPT-6.1 identities remain active: eight
Responses and seven Chat Completions identities. OpenAI content version is 39;
other Provider versions and all retained definitions are unchanged. The runtime
contains 88 definitions (82 chat, three image, three classifier). Withdrawn
entries remain in State with original definitions and provenance. The complete
review baseline and exact checked snapshot are adopted together. See
`.scratch/catalog-openai-retirement/status.md` for the actual verification and
`artifacts/catalog/README.md` for the versioned delivery.

The earlier 2026-11-04 retirement baseline withdraws nine Anthropic chat identities below
version 4.8. Seven remain: Fable 5/5.1, Opus 4.8/5/5.5, Sonnet 5/5.5; there is
no qualifying Haiku offering in the existing source. Anthropic content version
is 10; every other Provider version and all retained definitions are unchanged.
The runtime contains 79 definitions (73 chat, three image, three classifier).
All withdrawn definitions and provenance remain in State. See
`.scratch/catalog-anthropic-retirement/status.md` and `artifacts/catalog/README.md`.

The earlier 2026-11-04 Haiku baseline adds only
`chat/anthropic/anthropic-messages/claude-haiku-5-5`: 80 definitions (74 chat,
three image, three classifier), with Anthropic content version 11 and eight
active chat identities. Every prior definition, price and relative order remains.
The original models.dev response contains PDF, which has no runtime mapping;
the CLI rejected automatic mapping. A full manual official-source supplement
admits text/image and existing managed adaptive effort, plus complete strictly
above-100k whole-prompt prices. Provider source, exact checked runtime bytes and
the full review baseline are adopted together. See `.scratch/catalog-haiku55/status.md`.

The earlier 2026-11-04 baseline adds only
`chat/deepseek/openai-responses/deepseek-v4-pro`, the preferred documented entry
for Pro. Explicit Binding selection remains; Chat and Flash definitions stay
independent. DeepSeek content version is 2 with four active identities; the
catalog has 81 definitions (75 chat, three image, three classifier). All prior
80 definitions/prices/order remain. Official source captures and manual
field decisions declare text/function/reasoning/history, current effort mapping,
and a conservative peak-rate price snapshot, not calendar-based invoice rates.
ADR-0014's exclusion was corrected; historical measurements retain their original
identity. See `.scratch/catalog-deepseek-pro-responses/status.md`.

The current 2026-11-04 baseline adopts DeepSeek V4.1 Flash 1017 through
`deepseek-flash`, the official rolling API ID. Responses is revised first, then
Chat: only the two display names change; all IDs, capabilities, prices and
relative order stay unchanged. DeepSeek content version is 4 with four chat
identities; the complete catalog stays at 81 definitions (75 chat, three image,
three classifier). Official captured upgrade provenance supplements the original
serving associations. This label does not pin a permanent checkpoint or create
new vendor measurement. See `.scratch/catalog-deepseek-flash1017/status.md`.
