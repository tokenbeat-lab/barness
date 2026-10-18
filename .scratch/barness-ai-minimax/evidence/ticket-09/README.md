# Ticket 09 — evidence curation checkpoint

> 2026-10-18国际站转向：当前交付以[交付状态](../../delivery-status.md)和[ticket11国际站实现/离线证据](../ticket-11/README.md)为准。下列CN来源、追溯、401和当时缺口保留为4295ee8历史；不再等待中国区Key修复。国际站真实验收继续由10/09完成，历史记录不授予国际站支持。

当前整体交付为[PARTIAL](../../delivery-status.md)。下列追溯表保留其原始历史来源；最新验证与剩余义务由交付状态和ticket10证据分别记录。

Final acceptance is **BLOCKED by ticket10**. This checkpoint curates documents
and existing observations; it does not execute vendor calls, change runtime
sources, or close ticket09, ticket10 or the spec. Integration base is
`01cfc5a4eadfe31c35feb93500b3b3e270274d75`.

## International research checkpoint

[Independent read-only audit](international-research-audit.md) passes3,439integrity/accountingchecks for [source38b rawresearch](../ticket-10/international-research-38b7098.json). Global authentication and actualM3 are observed. Chatmirror parsing and Responsesnative localencoding remain runtime gaps; Messagesresearch3PASS is incompletecandidate evidence. Stage10calls/9Attempts/10reserved includesCN2/2;formal0. This audit neither fixes runtime nor certifies support. [Provenance](international-research-audit-provenance.json).

## Finite traceability

The following tables copy the exact finite `requiredCases` selections from
`ai/release/traceability.json` at ticket10 commit
`9481a6ad8028025a7f35da19a3e488935c6cd897`. That contract's SHA-256 is
`f827130b57c44abef2a66d9ae1c8da6aff26ace3f0a27e82701ec7975cdc4d4e`.
The source remains separately pinned even while ticket10 has not merged.

| Spec identifiers | Exact finite selections |
| --- | --- |
| R01–R28 | [28 requirements](traceability/requirements.json) |
| A01–A20 | [20 acceptance groups](traceability/acceptance.json) |
| US01–US80 | [80 user stories](traceability/user-stories.json) |

Their union contains 1,304 actual case IDs. Every one occurs in the original
1,307-case controlled offline bundle from commit
`ecbbeeb01007052b81052c256ebb9940d9b2e203`, finished at
`2026-10-17T23:46:22.409964Z`. Three additional observed cases have explicit
[supplementary links](traceability/supplemental-cases.json); they do not amend
the runtime release contract. The [curation receipt](curation-checkpoint.json)
records counts, source pins and outstanding requirements.

The complete observed case inventory is split into four finite files:
[001–400](traceability/cases/01.json),
[401–800](traceability/cases/02.json),
[801–1200](traceability/cases/03.json), and
[1201–1307](traceability/cases/04.json).
Each entry preserves its original status, test name, replay command and original
artifact/fixture hashes, and points to the actual public E2E test definition.
The [test source index](traceability/source-test-index.json) pins all 59 source
files; their bytes match the original run commit. Curation checked all 9,922
distinct original artifact files, totaling 86,466,083 bytes, against the original
manifest. The [original manifest](observed-manifest.json.gz) is losslessly
compressed, and the [original audit](observed-audit.json) is unchanged.

All original statuses in that one offline run are PASS. This is a historical
observation, with the original catalog and source applicability. These tables
are not fresh release inputs or complete raw-capture archives. Selected raw
cases are preserved by ticket10's curated capsules; final acceptance must retain
and verify its own complete ordinary/race bundles. The original public replay
commands create fresh controlled observations without a vendor credential.

`LIVE` selections remain required for the applicable items. A finite offline
mapping cannot satisfy them. Source changes, a later catalog activation, or
new captures do not update any historical status in these files.

## Historical Responses preservation

[Ticket03](../ticket-03/README.md#preserved-historical-observations) now retains
all 33 original controlled red/green bundles. The
[preservation receipt](ticket03-preservation.json) verifies all 29,012 archived
files, original status/source metadata, original audits and archive/index pins.
No historical capture was rewritten to appear current or live.

## Historical CN delivery gaps at4295ee8

The actual M3 Chat research returned HTTP401 `authorized_error` twice: two
logical calls and two actual Attempts. Research remains FAIL/incomplete; its
dependent native continuation is NOT_RUN because the tool probe failed.
Other M3 actual routes await the repaired, already authorized China standard
credential. Ticket10 owns the sealed failed captures and shared budget history.
M3.1 has no authorized M Plan entitlement, so its three online candidate states
remain NOT_RUN; all three offline model constraints are still mandatory.

The [independent authentication-failure audit](actual-auth-failure-review.md)
passes twenty checks for accounting and retention, while model support remains
FAIL/incomplete. The [capsule pointer](actual-auth-failure-pointer.json) pins
ticket10's 197,572-byte, 58-file portable archive. Ticket10 owns its tracked
archive; this checkpoint does not duplicate it. The
[ticket10 evidence guide](../ticket-10/README.md) is the portable capture entry point. The failed
capture retains commit `9481a6a`; subsequent harness corrections do not refresh
its source applicability or convert it to successful evidence.

The first whole ordinary run was interrupted at the root agent's request for
a known harness correction, after 536.7 seconds with exit -15. Its test body
had finished with 12,304 PASS / 52 FAIL; it has no whole-suite PASS conclusion.
The [read-only timing investigation](whole-ordinary-timing-investigation.md)
and [original-note provenance](whole-ordinary-timing-provenance.json) preserve
that diagnosis. Later targeted repairs, MiniMax-only complete runs and an
independent derived audit are separate observations documented in the current
delivery status; they do not rewrite the original exit status or failures.

Required work before closing ticket09 includes three own M3 candidate reports,
data-only formal activation, three own M3 formal reports, fourteen renewed
legacy support identities, four renewed business effects, and the actual
release gate with ordinary/race/pressure/frozen-pi evidence, both vets,
per-bundle redaction and rollback. The final independent Standards and Spec
reviews use fixed base `ecfd4b4`; the preliminary source review does not replace
them. A source change requires the existing applicability checks.

## Curation verification

These checks read only local documentary artifacts:

```sh
(cd .scratch/barness-ai-minimax/evidence/ticket-03 && shasum -a 256 -c SHA256SUMS)
(cd .scratch/barness-ai-minimax/evidence/ticket-09 && shasum -a 256 -c SHA256SUMS)
```

[curation-before.json](curation-before.json) records the six missing curation
inputs before this work. [validation.json](validation.json) checks the final
finite mapping and documentary pins; its PASS applies only to curation
integrity. It confers no runtime, actual live, formal support or release PASS.
