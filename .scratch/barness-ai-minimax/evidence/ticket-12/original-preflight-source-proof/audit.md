# Original 9dc failed preflight: independent retention audit

Integrity PASS. The original no-key preflight remains FAIL; this is no source approval or paid authorization.

The existing common-root raw archive is unchanged: 53,271,279 bytes, SHA-256 `56e08cbf7105f4fe1a1751bae42ba83988bcdbb4d1b1d443dd9bcd9a2d62ae3c`. Independent streaming verified all 27,417 regular files, one intentional negative symlink, and all 15,261 main-manifest artifact references (152,447 bounded checks). No duplicate raw archive was created. Expanded inventory is 253,350,201 bytes, including 141 bytes of symlink target text; the symlink was never followed or extracted. Deliberately corrupt nested fixtures remain negative test inputs.

Original outcomes stay 335 PASS / 62 NOT_RUN / 0 case FAIL. Both vets exited 0. The full no-key Go command exited 1 with exactly two environment audit findings in `GLM-corrective-1-capture-large/assertions.json` and `new-capture-replay-exit.json`. Their raw diagnostics remain unchanged in the ignored raw archive. Neither diagnostic is copied into this safe source proof.

Both original source-review and human-approval templates remain PENDING, with exact source, binary-manifest and plan bindings. The original checks receipt remains false. No approved receipt exists in that captured top-level collection; the three corrective canonical intents were absent when this audit ran. No vendor was called and no credential was read by this audit.

The separate `q-9dc-committed-source-proof.tar.gz` contains all 1,033 committed source-file bytes from Git `9dcf896b9bea357fd6c784acbca1b027dc96a177`, checked against the exact original `source-at-start.json`, plus safe failed-check/PENDING metadata. All source archive bytes passed portable verification. It pins the existing raw archive, original index and compressed full inventory as external boundaries; it does not replace or relabel them. Executable `live.test` remains inside the original raw archive, with its original hash and size in that inventory.

Repeatable verification: run `verify_q.py` with `--source-index q-9dc-committed-source-proof.json`, `--raw-archive` pointing to the existing common-root archive, and `--raw-index` / `--raw-inventory` pointing to the original worker index files. The verifier uses only Python's standard library, streams bounded entries, never follows the negative symlink, and never executes captured commands.

Source-proof archive SHA-256: `7fe0d49506b8c31c445f5ebc9f8868e25dbbee4e875e62e43ea226356601bd86`. Repository files, Git state, raw artifacts and outcomes were not modified. Later reviewed or paid collections are outside this historical failed-window claim.
