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
