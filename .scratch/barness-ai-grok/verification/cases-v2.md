# Grok frozen ordinary cases v2

- Execution/test source: `8f77ecc8d126ce8d1fc9006c3171bd63f79c493d` in managed `codex/grok-review-fixes`.
- **565 exact required IDs = 515 XAI +50 regression; 65 test roots; 67 families (23 literals +44 dynamic).**
- V1 retained unchanged. V2 replaces10 noncanonical omitted-effort labels `//` with `/default/` and adds one main public evidence-consumer case.
- Wire effort omission remains empty; only test/evidence label is `default`.
- All565 IDs are canonical normalized relative POSIX paths: no empty, dot or parent segments, no absolute IDs, no backslashes.
- No tests/build/vendor/live/race/pressure or repository edits performed for this enumeration.
- Exact `requiredCases`, `expectedTests`, splits, per-test IDs, families, source hashes and provenance: `/tmp/barness-grok-freeze-cases-v2.json`.

## Renamed exact IDs

| V1 ID | V2 ID |
| --- | --- |
| `XAI-E04-mapped-options-full//0` | `XAI-E04-mapped-options-full/default/0` |
| `XAI-E04-mapped-options-full//1` | `XAI-E04-mapped-options-full/default/1` |
| `XAI-E04-mapped-options-full//128000` | `XAI-E04-mapped-options-full/default/128000` |
| `XAI-E04-mapped-options-full//15` | `XAI-E04-mapped-options-full/default/15` |
| `XAI-E04-mapped-options-full//16` | `XAI-E04-mapped-options-full/default/16` |
| `XAI-E04-mapped-options-simple//0` | `XAI-E04-mapped-options-simple/default/0` |
| `XAI-E04-mapped-options-simple//1` | `XAI-E04-mapped-options-simple/default/1` |
| `XAI-E04-mapped-options-simple//128000` | `XAI-E04-mapped-options-simple/default/128000` |
| `XAI-E04-mapped-options-simple//15` | `XAI-E04-mapped-options-simple/default/15` |
| `XAI-E04-mapped-options-simple//16` | `XAI-E04-mapped-options-simple/default/16` |

## New public regression

- `TestXAIOptionEvidenceConsumption` / `XAI-E04-option-evidence-consumer` runs the actual bounded60-case `TestXAIReasoningAndBudget` collector and a compiled releasegate. It asserts native source/E2E acceptance while paid/human stages still deny release.
- Main inventory contribution is one case; nested producer cases are referenced artifacts, not extra main cases.

## Read-only existing provenance

| Existing managed attempt | Source-expected | Manifest cases | Missing / extra / duplicate |
| --- | ---: | ---: | --- |
| mapped-options-and-new-consumer | 61 | 61 | 0 / 0 / 0 |
| final-bounded-all-selected | 565 | 565 | 0 / 0 / 0 |

- Existing bounded producer IDs equal all60 source-derived mapped-option IDs. Existing native consumer source/E2E checks pass; live/race/pressure and release remain false. This read-only comparison creates no runtime receipt.

## Full ordinary comparison

- Full ordinary package manifest contains additional tests; require all65 expectedTests and565 exact requiredCases as a subset with no duplicate selected IDs.
- Expected sets split515 XAI and50 prior regression cases. XAI includes30 E17 process cases whose IDs do not start with XAI-.
- Select affected prior tests by exact function names; GLM/Kimi package prefixes also belong to unselected broader tests.
- Retain raw JSON punctuation, double slashes and full t.Name() portions; no ID normalization.
- Do not count child native/live bundles copied as fixtures as main ./ai/e2e IDs.
- Existing bounded worker manifest equality is provenance only; final full ordinary source/input binding and receipts remain separate.
- No source enumeration or controlled process PASS establishes real E17, race/pressure, formal support or release.pass.
- V2 replaces ten noncanonical omitted-effort // IDs with /default/ while leaving wire effort empty; old IDs must be absent from selected manifest.
- New native evidence consumer contributes one main case; its real60-case nested collector bundle is referenced provenance, not60 additional main cases.

- E17 real vendor qualification, final full-package E18, race/pressure and release.pass remain separate from this selected inventory and controlled evidence-consumer check.
