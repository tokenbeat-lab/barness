# Frozen pi-ai copy for the differential oracle

This directory is barness's own, rebuildable copy of the compatibility
baseline: pi-ai `0.87.1`, commit `898ab804050730e9dcefb4443875d5a932aa6a32`
(spec "Solution" and Testing Decisions §2). It does not read or modify the
research checkout, and CI does not need it.

## What is pinned

| Item | Value |
| --- | --- |
| Package | `@earendil-works/pi-ai@0.87.1` (exact, `package.json`) |
| Tarball integrity | `sha512-X/3PfQBnnoeVdO9Cv8zHghUMglzlgNZYGNzoPnbRoGnHl3Rw3TlA2UKSUB7BRHUOxMryHXYa8dnjWZlbRheDZA==` |
| npm `gitHead` | `f07218c4d4bbc12bef056a7058c3dd49dfe41abe` |
| `@earendil-works/pi-telemetry` | `0.87.1` (override; `^0.87.1` upstream) |
| All transitive packages | `package-lock.json` (versions + integrity) |
| Generated model data | `dist/providers/data/`, `.manifest.json` `generatedAt` 2026-09-22T19:31:44.346Z, `structureHash` `c6acf0a1…0095` |
| Model data hash | `373fa856ca2590733b90f0b6177c932dffeec4370bdd50835087b91ab2a18d13` (see below) |
| `openai.json` sha256 | `3c52c8587e7e4a1829ed98ec6e0e3d7bbb2baf47e8618556ed1e95a9c0362835` |
| `anthropic.json` sha256 | `474a010cd96c759d60419a6ba745747d9a87f172276dc85b7568da9d072b4e75` |
| `google.json` sha256 | `328822707e554d5977974d4b24695c2878aad48e5ca527c1efb0284141ff708c` |
| `@google/genai` (pi's Google SDK) | `2.21.0` (`package-lock.json`) |

`provenance.json` holds the values the Go side checks on every run: the runner
reports the installed version and model data hash, and `pioracle.Run` refuses
a copy that does not match. The model data hash is SHA-256 over the lines
`<sha256(file)>  <name>\n` for every file in `dist/providers/data`, sorted by
name, `.manifest.json` included.

## Why the npm release is the frozen commit

The frozen commit is not the published one, so the equivalence was checked
on 2026-10-01:

1. `git diff --name-only f07218c4 898ab804` shows that within `packages/ai`
   and `packages/telemetry` only `CHANGELOG.md` changed. The 4 commits in
   between touch `packages/durable`, root scripts and other changelogs.
2. A clean clone checked out at `898ab804`, `npm ci --ignore-scripts`, had the
   tarball's `dist/providers/data/` copied into the missing
   `packages/ai/src/providers/data/`. `npm run check:model-data` printed
   "Generated model data is valid." Then chord, tui and telemetry were built,
   followed by `packages/ai` `npm run build:offline`.
3. `diff -r` of the resulting `packages/ai/dist` against the tarball's `dist`
   found 0 differences across 770 files, and the same held for
   `packages/telemetry/dist`.

So this copy is exactly the frozen source plus the generated data that was
published with it. Commit `898ab804` cannot regenerate the data itself:
`generate-models` fetches live catalogs, and the result would drift.

## Rebuild

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
```

This needs npm registry access once. Running the oracle needs no network.
Install scripts are skipped because the runner needs none (no native
addons are loaded).

To re-verify the equivalence, repeat steps 1–3 against a fresh clone of
`github.com/earendil-works/pi`.

## Run

```sh
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -count=1 -run TestPiDifferential
```

`runner.mjs` executes one case per Node.js process. Isolation is enforced at
three levels:

- The Go side starts the process with an empty environment, and the runner
  exits if any credential, endpoint or proxy variable is present.
- Before pi loads, every socket connect is restricted to loopback. A
  non-loopback target throws inside `net.Socket#connect` before any byte is
  sent. This was checked by pointing a copy at `https://api.openai.com`.
- The API key comes only from the case, and pi's provider-env lookup
  receives an explicit empty map.

pi's own test suite is never run.

## Updating

Changing the pinned version means changing the compatibility baseline. Update
the spec first, then repeat the equivalence check for the new commit and
update `provenance.json` and this file. Run the differential last and record
any new differences in `ai/e2e/testdata/pidiff/ledger.json`.
