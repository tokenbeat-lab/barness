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

`go generate ./ai` regenerates the embedded artifact from reviewed source.
Runtime `ai` imports neither the maintenance package nor these source files.
Catalog changes use the candidate/check/apply maintenance workflow; code
compatibility checks do not grant vendor support, tenant permission or release
qualification. Historical activation wrappers remain under
`ai/release/catalog-activations/`; the former runtime activation writers have
been deleted.
