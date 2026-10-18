# Ticket17 committed-source applicability

Exact source: `230691742b731c77e8167b94646219e10f78c8f7`; baseline: `7931b163f983a8b362374163b7949d31e68bbec4`. This read-only seal uses the existing complete native selectors; it changes no recorded map or pin. Full per-report paths, original report hashes, file SHA256 values and Git blob IDs are in [the JSON inventory](/tmp/barness-ticket17-current-source-applicability.json).

| Existing evidence | Original selected count | Current complete count | Changed/new selected paths | Current strict equality |
|---|---:|---:|---|---|
| M3 Chat/Responses/Messages, each | 263 | 264 | payload modified; new message helper | false, all3 |
| G Chat/Responses/Messages, each | 217 | 218 | payload modified; new message helper | false, all3 |
| Original text Jev/Luna, each | 261 from immutable860 raw | 262 | payload modified; new message helper | false, both |
| E protocol114 | 862 | 865 | payload and plan modified; new helper and two E2E files | false |
| E visual Jev/Decisions, each | 323 | 324 | payload and plan modified; new helper | false, both |
| A-derived Jev | inherits862/323 | inherits865/324 | same protocol/visual changes | false |

Existing selectors require complete names and bytes. The JSON gives every changed/new path, old/native and current SHA256, Git blob ID, selector pin and report pin. Text's raw860 snapshot stays unchanged; only its existing execution selection is compared.

Generic10 source applicability remains valid under the established affected-combination policy: actual SDK/catalog/models/live producers are unchanged; the edited function has an unchanged non-MiniMax early return, and new private helpers have no globals/init or foreign-provider entry. Original8fbd report times and pins remain. This is explicit producer isolation, not a waiver based on absent HEAD checking; no current verifier was executed. Old launcher freezes remain stale.

Historical M3 PASS is source8900034; generic/text are8fbd; G/E/A/B are7931. Those measurements cannot certify changed2306917. The matrix remains17 rows/16 complete/old Responses FAIL, hash`7fbe72c37324a5700b0cff9416bf2a05543167ca1eed88186741669e39c6190f`. Latest raw Messages FAIL/missing support remains separate. A candidate validation stays1; no rows were imported.

No future count, fee, contract, rerun or approval is inferred. Later metadata/docs need a separate relevant-byte equality seal; no unknown HEAD is assumed. No project/guard/CLI/test/Key/network/B read or repository write occurred.

Original inventory:`ad1188a4dee7643bce3aa1c658eb8e001972bc9dff23b355fc3befa9d2065b2d`. Source review:`fa5d7b937d3a8376246e83403bf9c0291cd4995cb2288195dd4a26528c8da573`.

JSON SHA256:`29d7009a2f31e03199b17225b3f069b78d99777e4c0d8c21ea71f1c9e51f8ff0`.
