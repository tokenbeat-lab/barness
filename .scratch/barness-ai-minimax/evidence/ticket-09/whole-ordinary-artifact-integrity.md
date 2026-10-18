# Narrow integrity diagnosis: whole-ordinary 9481 / E08

This note diagnoses only `E08-error-body-at-limit-complete-full/failures.json` in the original evidence bundle. No original evidence or repository files were changed, no tests were run, and no full-bundle enumeration or archive work was repeated.

## Finding

The two recorded hashes describe two different byte sequences. The retained file matches `case.artifacts["failures.json"]` exactly. The separate `case.fixtures["failures.json"]` hash matches the checked-in input fixture at both source commit `9481a6ad8028025a7f35da19a3e488935c6cd897` and base `ecfd4b4d6ec54db30453f4b5112989e9fdbd0782` exactly. Applying the registered test-secret replacement to that input produces the retained file byte-for-byte. This is expected source-fixture attribution versus retained-artifact integrity, not evidence of a post-hash mutation.

The relevant write path uses `bytes.ReplaceAll` in `Run.writeFile`; it does not call `RedactJSON`.

## Original evidence reference and observed hashes

Bundle: `/tmp/barness-minimax-10/.evidence/minimax-ticket10-prepay-final/whole-ordinary/20261017T235628.944008000Z`

Manifest: `<bundle>/manifest.json`

- Observed SHA-256: `6fbe36fb26c89dda8b19c5a9572ce4179680f82356dadf460a6b41b2c18dfaac`.
- `git_commit`: `9481a6ad8028025a7f35da19a3e488935c6cd897`.
- Started `2026-10-17T23:56:28.944008Z`; finished `2026-10-18T00:04:32.805397Z`.
- Exactly one case has the exact E08 identifier; the manifest contains zero duplicate JSON object keys.
- Whole manifest retains **52 FAIL and 12,304 PASS**. The present finding does not change those statuses or turn the interrupted whole command into PASS.

| Meaning | Bytes | SHA-256 |
| --- | ---: | --- |
| Original source fixture at 9481 and ecfd4b4; matches E08 `fixtures.failures.json` | 18,248 | `9b94f8b16e28f6a81fd56563e3d5a7a60cb4a6ad77924a9caf7611da1e6945f3` |
| Retained E08 file; matches E08 `artifacts.failures.json` | 18,242 | `1be5afdbeb412ffd9d4b51a253f9fd4e869d2a31c8a4002437490b577689063e` |
| Source after registered replacement | 18,242 | `1be5afdbeb412ffd9d4b51a253f9fd4e869d2a31c8a4002437490b577689063e` |

Only one registered secret occurs in this input. It is replaced once with the test alias `key:tenant-a@v1`: 21 source bytes become 15 retained bytes, explaining the entire six-byte length difference. Secret values and business response bodies are deliberately omitted here.

E08's retained `assertions.json` hashes to `51d73636639b4c45c9d0c94d4a2745db8d1e1a47393332ecafa037ff1cfe5d09`, matching its artifact entry. Its local status is PASS, with 41 assertions and zero failed assertions. That local status is distinct from the whole-run 52 failures. The bundle’s `audit.json` was added after the original process interruption by root’s later, independent derived audit using the existing standalone `auditbundle` CLI at source `1205300` (72.16 seconds, exit 0, no key used). That later audit records zero findings and observed SHA-256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`; the original 9481 run never finished its audit.

## Source evidence and order

All paths below are repository-relative to `/Users/cyber/RestoX/barness`. Inspected relevant files are byte-identical between 9481 and ecfd4b4, including the source fixture, evidence writer/inventory, E08 test/fixture loader, TestMain registration, and offline release loader/inventory.

1. `ai/e2e/failure_fixture_test.go:68-78` reads `testdata/responses/failures.json` and returns the original byte slice alongside its decoded fixture.
2. `ai/e2e/byte_limits_test.go:173-185` selects the error-body scenario and invokes `ev.Fixture("failures.json", failuresRaw)` exactly once for this case. `ai/e2e/limits_fixture_test.go:110-145` creates a distinct case per boundary and entry point; the target identifier is formed at line 126.
3. `ai/e2e/main_test.go:29-34` registers environment and known fixture-key replacements before any case runs. `ai/e2e/world_test.go:21-44` defines the registered known-key set.
4. `ai/internal/testkit/evidence/evidence.go:258-267` computes `sha256.Sum256(data)` and saves it in `c.hashes[name]` **before** calling `writeFile`.
5. `ai/internal/testkit/evidence/evidence.go:194-200` replaces registered secret bytes with aliases, then calls `os.WriteFile`.
6. `ai/internal/testkit/evidence/evidence.go:305` clones the source-fixture hash map into the finished case. `Run.Finish`, lines 123-135, separately computes `caseArtifacts(c.dir)` and serializes the source hashes as `fixtures` and disk hashes as `artifacts`.
7. `ai/internal/testkit/evidence/inventory.go:11-29` explicitly freezes every already-written file at Finish, reading file bytes and hashing them from disk.
8. `docs/adr/0017-barness-ai-release-gate.md:128-131` explicitly states that the full artifact inventory includes fixture files **without changing the original meaning of fixtures**.

Thus `fixtures` and `artifacts` may legitimately have the same filename with different hashes. They are separate semantic stores, not earlier/later artifact snapshots. Merging those maps by filename destroys that distinction.

## Release-verifier behavior

For this offline `./ai/e2e` manifest, `ai/internal/testkit/release/load.go:27-35` decodes `Artifacts` but does not import `Fixtures`. At lines 50-54 it rejects duplicate case IDs/directories and verifies the case's artifact inventory. It does not select a last case entry or merge the two maps.

`ai/internal/testkit/release/inventory.go:12-35` requires an exact file inventory, rejects missing/unlisted/nonregular files, and compares every listed artifact's actual bytes against its artifact hash. It fails on any mismatch. The present E08 retained file satisfies that artifact hash. Source fixture hashes are separate input provenance and are not alternate required hashes for the redacted on-disk file.

The code permits ordinary `os.WriteFile` replacement in general, but this target's code path, unique case/filename entries, exact source hash, and exact registered transformation supply no evidence of a duplicate snapshot, later rewrite after artifact hashing, or external byte mutation. This narrow comparison cannot prove an unrestricted historical negative; it does resolve the reported E08 mismatch completely.

## Repeatable read-only check

Run the following from `/Users/cyber/RestoX/barness`. It prints only lengths, hashes, counts, and booleans; it never prints secret values or fixture bodies.

```python
import hashlib, json, re, subprocess
from pathlib import Path

rev = "9481a6ad8028025a7f35da19a3e488935c6cd897"
base = "ecfd4b4d6ec54db30453f4b5112989e9fdbd0782"
bundle = Path("/tmp/barness-minimax-10/.evidence/minimax-ticket10-prepay-final/whole-ordinary/20261017T235628.944008000Z")
case_id = "E08-error-body-at-limit-complete-full"
fixture_path = "ai/e2e/testdata/responses/failures.json"
show = lambda r, p: subprocess.check_output(["git", "show", r + ":" + p])
sha = lambda b: hashlib.sha256(b).hexdigest()
manifest = json.loads((bundle / "manifest.json").read_bytes())
cases = [c for c in manifest["cases"] if c["id"] == case_id]
assert len(cases) == 1
case = cases[0]
source = show(rev, fixture_path)
retained = (bundle / case_id / "failures.json").read_bytes()
keys = show(rev, "ai/e2e/world_test.go") + b"\n" + show(rev, "ai/e2e/preflight_reject_test.go")
pairs = re.findall(rb'secret:\s*"([^"\n]*)"\s*,\s*alias:\s*"([^"\n]*)"', keys)
transformed = source
replacement_count = 0
for secret, alias in pairs:
    replacement_count += transformed.count(secret)
    transformed = transformed.replace(secret, alias)
checks = {
    "source_matches_base": source == show(base, fixture_path),
    "source_matches_fixture_pin": sha(source) == case["fixtures"]["failures.json"],
    "retained_matches_artifact_pin": sha(retained) == case["artifacts"]["failures.json"],
    "registered_transform_matches_retained": transformed == retained,
    "source_hash_is_not_retained_hash": sha(source) != sha(retained),
}
print({"source_bytes": len(source), "source_sha256": sha(source),
       "retained_bytes": len(retained), "retained_sha256": sha(retained),
       "replacement_count": replacement_count, "checks": checks})
assert all(checks.values())
```

The equivalent probe was executed during this diagnosis and all five checks were true, with replacement_count=1. No test suite, release gate, or archive was rerun.
