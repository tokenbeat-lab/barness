# Current catalog identity reference audit

Scope: active `ai` Go, JSON and Markdown plus `docs` Markdown, including tests and examples that may execute after the full suite's usage failure. Reviewed from integration `1d1f6de5`; assertion fix is `aba70f3a`. No catalog definitions, historical bytes or qualification policies changed.

Searches used:

```sh
rg -n '2026-10-29\.1|f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d' ai --glob '*.go' --glob '*.md' --glob '*.json'
rg -n 'catalog-snapshot\.json|2026-10-29\.1|f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d' ai --glob '*.go' --glob '*.md'
rg -n 'catalog-snapshot\.json|2026-10-29\.1|f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d' docs --glob '*.md'
rg -n 'CatalogVersion[^\n]*[=:][^\n]*"2026|catalogVersion[^\n]*"2026|Version[^\n]*==[^\n]*"2026|Version[^\n]*!=[^\n]*"2026' ai --glob '*.go'
```

## Current discovery

`ai/e2e/usage_test.go` was the only unresolved current-context reference to the old full snapshot. Its `builtin-catalog-pinned` case read the immutable release snapshot while comparing today's `ai.BuiltinCatalog()`. The recorded usage fixture remains original price provenance; it does not choose the current discovery identity.

The case now reads `ai/catalog_builtin.json` as a complete typed `ai.Catalog`, recomputes and validates its declared hash, then compares public built-in version/hash. It retains stable-copy, changed-price and invalid-price guards. No new catalog version/hash literal is introduced, and future data maintenance need not edit Go directory definitions. `TestMaintenanceCatalogMigration` remains the independent original-118 definition/price/default/order comparison; the generated file is not substituted for that historical source.

Other current consumers already use the right boundary:

- `ai/e2e/catalog_snapshot_fixture_test.go` reads the generated artifact for `currentCatalogVersion`; OpenAI/Google/xAI image tests use it or current public discovery. The xAI label still says “release snapshot”, but its helper resolves the generated artifact and does not compare the old identity.
- `ai/e2e/chinese_current_catalog_process_test.go` compares default reports with `catalog_builtin.json`; the three v3 configs explicitly bind the new composition. Its remaining release-snapshot read is solely the old-v2 pairing/refusal check.
- `ai/README.md`, `ai/README.zh-CN.md`, `docs/barness-ai/README.md`, and `docs/barness-ai/catalog-maintenance.md` identify generated current data and describe the release snapshot as historical.

## Intentionally historical references preserved

| Paths | Contract retained |
| --- | --- |
| `ai/release/catalog-snapshot.json` | Original 2026-10-29.1/f44d identity and all original bytes. |
| `ai/e2e/catalog_maintenance_migration_test.go`, `ai/e2e/catalog_history_fixture_test.go` | Independent initial-118 preservation and historical provenance, with the new composition identity explicitly separate. |
| `ai/examples/chineseeval/testdata/{jev-1.13.0,jev-route,luna-route}-v2.json`, example README, historical pairing checks in `chinese_current_catalog_process_test.go` | Immutable prior configurations; accepted only with their recorded snapshot, rejected with current default. |
| `ai/e2e/testdata/xai-images/official.json`, historical activation check in `ai/e2e/xai_images_builtin_test.go` | Original activation revision, independent of current global composition. No new source date or qualification is claimed. |
| `ai/live/glm_replay_catalog_test.go`, `ai/internal/testkit/evidence/catalog_history.go` | Explicit measured-commit Git reads and exact version/hash validation for historical replay only. |
| `ai/e2e/historical_catalog_test.go`, `ai/e2e/glm_corrective_launcher_test.go` | Synthetic historical repositories and negative/launcher fixtures; they do not select current production discovery. |
| `docs/adr/0017-barness-ai-release-gate.md`, `docs/adr/0044-barness-ai-historical-catalog-replay.md` | Prior design/historical replay records. The former write-snapshot description remains historical context; current maintenance follows the accepted feature contract/ADR-0058 and current guides. |

The historical inequalities in `glm_formal_catalog_test.go` and `xai_images_builtin_test.go` only assert that current content is distinct from an earlier activation. They do not pin current identity to an old value. Other metadata assertions compare the selected public/host-owned catalog, not an old global constant.

No other unresolved current old-version/hash/path consumer was found in this scope. Historical catalog/provenance and all qualification evidence remain unchanged. Ordinary targeted evidence is recorded in `usage-catalog-consumer-verification-v1.json`; this audit does not create vendor support or release qualification.
