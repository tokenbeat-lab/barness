# Release bundle consumer boundary — preregistered failures

Public seams: `evidence.Run.Case/Finish` → `release.LoadBundle`; the final
diagnostic uses the public release CLI with explicitly adopted historical
`-bundle/-race-bundle` inputs. Parent approved these seams and consumer-only
scope before tests. No producer, runtime, selected M3/GLM source or raw evidence
is changed.

Failures to exercise before the corresponding implementation:

1. Eight real empty-effort IDs end with one `/`; physical directories are
   canonical while the producer deliberately retains the raw identifier.
   Nested child artifacts are recursively declared by the producer. Both must
   import without losing IDs, statuses or parent assertion ownership.
2. Canonical directory aliases (`a` / `a/`), duplicate IDs, traversal, absolute
   paths, dot components, repeated separators and backslash aliases must fail.
3. A symlink at the selected case, an ancestor beneath the bundle, a nested
   directory or artifact must fail; opens must stay rooted even on replacement.
4. Nested missing, extra, omitted or hash-changed files must fail. Nested
   `pidiff.json` is an artifact, never the parent's differential/support proof.
5. Untrusted declared keys, excessive file count, directory fanout/depth,
   oversized actual files and excessive cumulative bytes must fail before
   unbounded enumeration/read/allocation.
6. The two immutable actual bundles must retain 12,391 cases, 12,385 PASS and
   six genuine GLM FAIL each. Successful import is not a fresh suite execution
   or release PASS; original numeric internal command exits remain unknown.

Finite limits preregistered and parent-approved before implementation:
512 artifact files, 2,048 total directory entries, depth 24, 64 MiB per file,
256 MiB aggregate per case. Actual ordinary maxima: 133 files, depth 2,
36,351,001 bytes/file, 63,697,858 bytes/case. Incremental directory enumeration,
descriptor-rooted opens and bounded streaming hashes preserve these limits
under growth/replacement. These are consumer import ceilings, not vendor or
producer budget changes.

## c680 review follow-up (before implementation)

Standards found two P2 gaps despite the original public checks passing:
post-inventory consumers reopened unchecked paths, and a repeated streaming
directory name could inflate file-count equality. Preserve the c680 checks as
historical intermediate evidence, not proof these gaps were closed.

Public RED scenarios: load intact genuine revocation bytes and confirm PASS;
then replace the case, its ancestor or bundle directory with an outside
symlink, replace a consumed file with a same-bytes symlink, or append JSON
whitespace to a consumed file (same semantics, different pinned hash). Each
subsequent public `RevocationEvidence` call must refuse the modified input.

Parent explicitly approved one necessary isolated iterator-boundary exception:
a finite scripted `fs.ReadDirFile` returns `a.json, a.json` while declarations
require `a.json, b.json`. Counting two is insufficient; duplicate entry names
must be refused and the exact successfully authenticated file-name set must
equal the declared set. No new exported API or dependency is allowed. Ordinary
unique iteration and nested parent attribution remain unchanged.
