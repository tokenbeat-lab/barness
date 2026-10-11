# Frozen pi-ai 1.1.0 differential oracle

The sole runtime baseline is `@earendil-works/pi-ai@1.1.0`, fixed source
commit `abe508e1b89912adde45528136c3221eb69acdd7` at
[official v1.1.0](https://github.com/earendil-works/pi/tree/v1.1.0).
Both npm packages omit gitHead; `npm_git_head` is empty, never inferred.
Two independent full byte-identical source rebuilds establish equivalence.

| Input | Pin |
| --- | --- |
| AI / telemetry | exact `1.1.0`, including telemetry override |
| AI integrity | `sha512-1T7LAkc/5Bvc0v6w4vAGVdCrli0o/E0pEmYKTnixu95vSFArBjvbhS/G4ZwI0RUePgf0Imcu0VyqlM4EcXxqfw==` |
| Telemetry integrity | `sha512-8gAK05/2pPozZZz6hCLOi8x2vsqzvsmO9gG92kbvuvTPm0qMHaFlaXX98ndXsrOi5pmzha7vNLguuG4+0Rgxjw==` |
| Model schema / generatedAt | `6` / `2026-10-07T22:01:28.515Z` |
| Model structureHash | `080cfcf6bdd13064ce2362f26e4902c31503b675fd06a8c4685c04a0333bc465` |
| Complete data SHA-256 | `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8` |
| Node SDKs | OpenAI `7.19.0`, Anthropic `0.129.0`, Google GenAI `2.21.0` |
| All transitive inputs | package-lock.json and model-data.sha256 |

The hash covers `<sha256(file)>  <name>\n` for all 43 sorted files in
`dist/providers/data`, including `.manifest.json`. The runner refuses incorrect
version, commit or model hash before provider import; runner.test.mjs exercises
all three. Go validates the returned version and data identity on every call.

[Reproducible source proof](../../../../../.scratch/barness-ai-thinking-sampling/release-verification/README.md)
records two clean temporary builds on 2026-10-11, each **827 AI + 24 telemetry**
files, zero differences. Official tag identity and archive hash replace the
unavailable Git transport; no npm gitHead is fabricated. Only released generated
data is copied into frozen source; there is no online model generation.
Nine mutation checks and two I/O boundaries fail safely; audits are clean.
The oracle lock has the proof's exact dependency graph, with runner-specific root
name and engine declaration. Historical 1.0.0 evidence remains archival only.

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
node --test ai/internal/testkit/pioracle/node/runner.test.mjs
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -count=1
```

Registry access is required only for installation. The Go child gets an empty
environment; the runner refuses credential, endpoint and proxy variables,
restricts TCP to loopback and rejects IPC. Synthetic keys enter only from case
input. Chat lookup requires type chat, so image/classifier exports cannot enter chat. The oracle
never runs upstream tests. Baseline changes require source proof, new pins,
full zero-pending differential and the complete release contract.
