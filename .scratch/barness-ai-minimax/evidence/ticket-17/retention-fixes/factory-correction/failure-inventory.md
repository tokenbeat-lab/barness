# Sole remaining resource-constructor finding

RESOURCE-FACTORY-STANDARDS-01 / RETENTION-FIXES-SPEC-01 concern the new private negative-fixture constructor at e1f3b84. The previous four P2 findings are independently closed; original/native and both negative capsules remain valid and immutable. Retain the e1 source/reports and exact public before/after history.

Confirmed seam: real `resource-probe.py --common-root <original archive root> --value false --out <fresh private output>` executed through runpy actual __main__/argv. OS audit injection only replaces private file entries at real standard open boundaries; never mock reads/helpers and never mutate original files/shared inodes.

Read inventory, fixed before changes:

1. Private `index.json`: >1MiB valid JSON with trailing whitespace. Old unbounded parse continues to inventory and full construction; eventual current verifier refusal is not oversized native acceptance. New existing1MiB bounded regular FD read must stop first.
2. Private `inventory.json.gz`: >8MiB valid original gzip plus zero padding, with matching private metadata pins. Old unbounded read/decompress continues. New compressed8MiB bound must reject before expansion.
3. Inventory decoded JSON:40MiB+1 valid JSON/padding, compressed below8MiB and exact private SHA/length pins. Old gzip.decompress materializes all bytes; new existing full-reader decoder must stop at40MiB+1 and refuse the limit.
4. Seal metadata: summary JSON becomes8MiB+1 before its seal open. Old seal reads whole file, rewrites SHA256SUMS and invokes the verifier, which later rejects. New existing8MiB per-file bounded regular read must refuse before seal publication/verifier invocation. Keep seal whole-shape/file/payload limits finite using the same established envelopes.
5. Fresh output archive becomes257MiB sparse zeros immediately before final read/hash, after its earlier stat. Old unbounded pathname hash continues and invokes the verifier, whose later size refusal remains honest exit1. New same-FD nofollow/nonblock/regular initial256MiB cap, cumulative byte cap and before/after stat equality must reject before output pins/public verification.

All RED factory processes may end1 due the existing current resource verifier or later bounds; classify by actual early/late stage and raw open events, never call these malformed native successes. A normal current constructor should return0 while its deliberately counterfeit resource is publicly refused1; historical old accepted-RED receipts stay unchanged. Also read back original ticket17 and both existing full negative capsules. No old8 producer cases, Go, network, Key, native refresh or paid call is needed.
