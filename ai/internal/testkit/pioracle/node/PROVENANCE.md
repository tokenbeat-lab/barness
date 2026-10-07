# Frozen pi-ai 1.0.0 differential oracle

The only running compatibility baseline is `@earendil-works/pi-ai@1.0.0`,
[pi tag v1.0.0](https://github.com/earendil-works/pi/tree/v1.0.0), commit
`a13d35a742c6ef8462812a28fbe1d8c8b7431c32`. No research checkout or online
model generation participates in installation or runtime.

| Input | Pinned value |
| --- | --- |
| ai and telemetry | `1.0.0`, exact dependencies; telemetry override also `1.0.0` |
| ai npm gitHead / tag commit | `a13d35a742c6ef8462812a28fbe1d8c8b7431c32` (identical) |
| ai tarball integrity | `sha512-3/W1vdDaVtpeMd23ElvJC12HLA5yS/BGqqcXF+0SK082dN7cbgNcCwguTBRBC258Ke8SzSvUW1B75iAf8w8IxA==` |
| telemetry tarball integrity | `sha512-WjNBj5TYIiPZFQEz2WlULcDwPLaKwIlmsKjVeYM+LJSbnSp38kWsHUJysIDGnu23IcLbKoPtywsvYfUJCeZePA==` |
| Transitive dependency versions and integrity | `package-lock.json`; same resolved package inputs as issue 02 |
| Model schema / generatedAt | `6` / `2026-10-01T18:57:11.882Z` |
| Model structureHash | `235f2f320916ab6b0d7193e0bf66ec7983e9bc05abeddd7264923fb1e7eaf76e` |
| Complete model-data SHA-256 | `8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e` |
| openai.json SHA-256 | `87faafcdd168c9b6fdaef521137834a207437daf0d71641fd0a2f3dddf4d1ecb` |
| anthropic.json SHA-256 | `3ff00b68990382e42626ebd1b65dcc28ae3cad9e6faa8920331ba92f067f1b3a` |
| google.json SHA-256 | `c0f5a633e5e5a5044432db3e537f4bc674d4de18df09325c6349547fcc13cabe` |
| deepseek.json SHA-256 | `10a296fb3e898f7715c80890c8af0af4f5fd58f22dbcd6612fdc5d762157ccdc` |
| Node SDKs | OpenAI `7.19.0`, Anthropic `0.124.0`, Google GenAI `2.21.0` |

The model hash is SHA-256 over `<sha256(file)>  <name>\n` for all 43 files
in `dist/providers/data`, sorted by filename, including `.manifest.json`.
[model-data.sha256](model-data.sha256) pins every file. The runner checks version
and hash before importing any provider. Go also checks returned identity on
every entry. `runner.test.mjs` proves both mismatch refusals.

## Release equivalence

Issue 02's [reproducible proof](../../../../../.scratch/barness-ai-pi-1.0/release-verification/README.md)
ran in two independent clean temporary environments on 2026-10-08:

1. Resolve tag and both npm gitHeads; all equal the target commit above.
2. Verify tarball integrity and install with `npm ci --ignore-scripts` using
   the fixed release and source locks. Copy only the released model data into
   the frozen source; never run `generate-models` or `hydrate:model-data`.
3. Build chord, tui, telemetry, then ai `build:offline`. Compare complete
   filename sets and bytes, including hidden files, JS, declarations and maps.
4. All **811 ai** and **24 telemetry** dist files match, **0 differences**.
   Wrong versions, gitHeads, missing/altered data, output differences, unlocked
   dependencies and invalid integrity are rejected; redaction audit is clean.

The oracle lock retains exactly that proof's resolved dependency graph; only
its root package name and Node engine declaration describe this runner.
The proof's source/dist hash lists remain in issue 02's evidence directory.

## Rebuild and verify

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
node --test ai/internal/testkit/pioracle/node/runner.test.mjs
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -count=1
```

Installation needs registry access once; runtime only uses a local controlled
Provider. The Go child environment is empty; the runner refuses credential,
endpoint and proxy variables, restricts all TCP sockets to loopback and rejects
IPC. Keys come only from synthetic cases. Chat catalogs explicitly require
`type: "chat"`; image/classifier exports cannot enter a chat call. The oracle
never runs upstream tests. Changing the baseline requires a spec update,
release equivalence proof, new provenance and a full zero-pending differential.
