# Ticket 13 failure inventory

Preparation baseline: `9c25b5583621fa18cc286726ed115298b0ad1fa8` on `main`.
Written before any ticket-13 implementation or new test. This is a planned
inventory, not a claim that its E2E cases ran or that the ticket is complete.

## Admission prerequisite

The ticket requires complete, auditable candidate protocol and independent
visual quality evidence before considering fixed Jev 1.14 catalog inclusion.
Tickets 11/12 are resolved as implementation work but their real-route results
are explicitly NOT_RUN. Their controlled results do not meet this prerequisite.
Do not alter historical packages, infer authorization from another task, or
publish the catalog to manufacture a successful admission test.

## Confirmed seams and evidence

The approved spec/design fixes the runtime seam at `Client.Classify` and
`HookedClient.Classify`. Catalog discovery/snapshots are observed through public
catalog/client APIs. Formal producers, independent verification and support
matrix merging are exercised through their public process commands; controlled
HTTP proves those commands only. No private-helper unit tests are planned.

Each implementation slice must first retain a failing public E2E, then its green
result. Save fixtures/image hashes, public inputs, final request/response,
results/errors, Attempt/Observer, assertions, finite config, resource release,
inventory and redaction audit. Record normal/race and real-service provenance
separately; new formal-Hash evidence must come from a new run.

## Failure cases to cover

| ID | Failure | Required observable outcome |
| --- | --- | --- |
| F01 | Candidate protocol package absent, controlled, NOT_RUN, failed or incomplete | No formal inclusion or real PASS claim |
| F02 | Independent effect package absent, controlled, NOT_RUN, wrong candidate/model/route or unlinked to its protocol evidence | No formal inclusion; classification, existence and Box quality stay separate |
| F03 | Fixture, annotation, configuration, source seal, inventory or audit altered/missing | Explicit verification rejection; preserve the original package |
| F04 | Runtime/HTTP schema validity substituted for accuracy, Box confidence treated as existence, failures removed from denominators or labels changed to improve results | Reject evidence or retain honest incomplete result; no production accuracy promise |
| F05 | Moving alias, inferred model capability or another fixed model automatically gains images/Box | No discovery/authorization expansion; explicit unsupported result before credentials/Attempt |
| F06 | New model omits or falsely advertises any implemented context/state/instructions/criteria location, PNG/JPEG, finite count or Box kind | Catalog/config rejected or capability accurately restricted |
| F07 | Price lacks a pinned primary source, changes silently, or image occurrence is billed twice | Do not publish an unsupported rate; estimated cost follows reported tokens and recorded price snapshot |
| F08 | Borrowed catalog/query slices mutate image locations, kinds, MIME, limits or already pinned calls | Public snapshots remain independently owned; content change requires new version/Hash |
| F09 | Image policy, model whitelist, tenant permission or explicit experimental selection missing; Hook changes any of these | Both Classify entries reject with the defined capability/request error and zero unauthorized send |
| F10 | Invalid MIME/APNG/dimensions, source/expanded counts or bytes, State/question/body/output limits exceeded | Existing phase/code and finite resource behavior hold for the formal model, including final Hook guard |
| F11 | Revocation or concurrent tenant mutation changes a pinned call or leaks image identity/usage | New calls obey revocation; pinned calls finish under their snapshot; tenant data remain isolated |
| F12 | Malformed Box, wrong legend or final Hook image identity leaks partial answers | Atomic result rejection; independently reported valid usage retained |
| F13 | Jev 1.14 report is accepted under the existing 1.13 identity or overwrites its row | Model-specific identity/capability validation rejects the mismatch; original 1.13 row and raw package retained |
| F14 | Formal report names wrong operation/provider/API/model/endpoint, old/candidate Hash or stale source | Public verifier/merge refuses it; no matrix mutation |
| F15 | Mandatory capability omitted, duplicated, relabelled, narrowed by test filter or borrowed from another route | Public producer/verifier/merge rejects an accepting verdict |
| F16 | Calls/Attempts/questions/images/bytes/time/retries exceed ticket-11/12 per-round ceilings, disagree with captures, or are hidden as NOT_RUN | Pre-send finite reservation and independent accounting reject/stop; retain failures and unsent tasks |
| F17 | Missing/forged final capture, Attempt/Observer terminal data, assertion, RequestID or usage | Reject report even if outer checksums are refreshed |
| F18 | Bad inventory/audit, path escape, symlink, extra/missing file or oversized evidence input | Bounded explicit rejection before merge; no credential/data exposure |
| F19 | Candidate/old bundle is rewritten or re-labelled as formal, or a controlled package is promoted by metadata alone | Strict real/formal verification refuses it; originals stay auditable in their own source/catalog context |
| F20 | Internal schema changes leave producer/consumer/migration out of sync or introduce runtime compatibility/dual-write paths | Atomic controlled migration with all consumers updated; historical raw bundles preserved |
| F21 | New Hash silently refreshes other rows' model/status/time or makes historical PASS current | Historical rows unchanged; current-catalog acceptance remains incomplete until ticket 14 |
| F22 | Jev 1.13 text, Decisions images, frozen Pi projection or usage behavior regresses | Public regression suite rejects change; no alias/fallback or old evidence promotion |
| F23 | Missing dedicated approved route credentials/authorization or ambient mixed-route credentials | Zero real calls/Attempts; explicit NOT_RUN/config failure; never read another task's key |
| F24 | Documentation, issue, proposed ADR or parent spec marks overall work done before real/formal acceptance | Keep incomplete state explicit; ticket 14 owns overall release conclusion |

## Resume order

1. Obtain the explicitly approved route configuration or existing authentic
   real protocol/effect bundles and verify them with the existing strict commands.
   Do not rerun for a higher score or silently change labels/budgets.
2. Confirm the pinned primary-source rate/limits and enumerate the formal
   model-specific required capabilities without changing historical identities.
3. Implement one public E2E red/green slice at a time: catalog/client behavior,
   formal producer/consumer/validator and controlled migration, then tamper cases.
4. Collect fresh authorized protocol and required effect evidence under the
   formal catalog's own Hash within the respective per-round ceilings. Never
   relabel candidate packages or inherit their PASS.
5. Run focused normal/race checks, full suite once at the end, independent
   Standards/Spec review, and commit the implementation. Synchronize delivery
   documentation while leaving proposed ADR/parent completion to full acceptance.
