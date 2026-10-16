# Fixed Jev 1.14 formal support

The built-in catalog `2026-10-17.1` includes fixed `jev-1.14.0` alongside the
unchanged `jev-1.13.0` text model. Images/Box remain experimental: explicitly
set `TypeSafeOptions{ExperimentalImages: true}`, authorize the model in the
classifier Binding, enable `AllowClassifierImages` and choose finite image and
reference policies. Catalog discovery grants no permission. Moving aliases are
excluded. Decisions still rejects local image positions and Box.

Supported Jev positions are context, State, own-question Instructions and
Criteria. The catalog's 24 sources is a bounded barness subset, not a vendor
count limit. Only static PNG/JPEG, 280–1568 pixels per edge, are supported;
other limits and the distinction between source and vendor-processed capacity
are pinned in [sources.json](sources.json). Input costs $0.042/M tokens, including
reported image tokens, with free output; this is an estimate, not a bill.
Primary sources: [models](https://docs.typesafe.ai/models),
[images](https://docs.typesafe.ai/images), [Box](https://docs.typesafe.ai/primitives/box),
read 2026-10-17. No image token formula or production accuracy guarantee is added.

## Separate identities

The existing `typesafe-classifier` report/row remains fixed to 1.13.
`typesafe-classifier-jev-1.14.0` has an independent required twelve-scenario set:
default context, explicit context, ordered reuse, State/Instructions,
visual Criteria/legend and Box/independent existence, each through plain/hooked
Classify. Schema 3's existing route reports gain an optional `jev` evidence
member for this new identity; historical report/matrix fields remain unchanged.
Current producers/consumers are updated together, with no compatibility execution
path, schema migration or relabelling of historical packages.

Formal acceptance requires both a complete formal protocol package and a complete
independent formal effect package at the current built-in catalog Hash and source
snapshot. The effect task retains its original candidate provenance/fixture pins;
the enclosing formal package separately verifies the new formal protocol, rather
than interpreting that old provenance receipt as formal PASS. Old packages are
verified in their own recorded checkout; changed source cannot borrow a current
PASS. Integrity receipts are not origin signatures: the Host must preserve trusted
raw live packages. Account/region/permissions can remain explicitly unverified;
an isolated credential process does not introspect provider privileges.

## Produce each package once

From the repository root, build the existing commands:

```sh
go test -tags live -c -o /tmp/barness-jev-probe ./ai/live
go build -o /tmp/barness-visualeval ./ai/examples/visualeval
```

Use the candidate guide's isolated approved authorization and credential variables
for the protocol, additionally selecting `BARNESS_MM_FORMAL=1`. From `ai/live`,
run `/tmp/barness-jev-probe -test.run='^TestMultimodalFormal$' -test.timeout=180s`.
The exact entry is mandatory; mixed credentials and unknown formal routes fail
before sending. Candidate mode remains a separate explicitly selected workflow.

For effects, the visual evaluation authorization must name the **formal** catalog
Hash. In an isolated process with only `BARNESS_VISUAL_KEY`, run:

```sh
/tmp/barness-visualeval -formal -mode live -route jev -entry plain \
  -authorization /trusted/formal-effect-authorization.json -out NEW_EFFECT_PACKAGE
```

The protocol has ≤12 calls/12 Attempts, 32 questions, 24 sources/32 occurrences,
8 MiB decoded sources/16 MiB expanded/4 MiB State, 30 seconds per call/180 seconds
per process including 10 seconds cleanup. Effects retain 12 fixed samples,
≤12 calls/12 Attempts, 48 questions/24 occurrences, 8 MiB sources/16 MiB expanded,
30 seconds per call/240 seconds per process including 15 seconds cleanup.
Both have zero retries. Do not rerun samples for higher scores. Failed/NOT_RUN
raw processes retain their terminal evidence and cannot produce formal support.

## Seal, verify and merge without provider calls

Go and Python 3 are required. These commands only verify retained evidence and
replay captured responses through loopback HTTP; they load no route credential:

```sh
go run ./ai/live/cmd/jevformal -protocol NEW_PROTOCOL_PACKAGE \
  -effects NEW_EFFECT_PACKAGE -out NEW_SUPPORT_PACKAGE -require-live
go run ./ai/live/cmd/jevformal -verify NEW_SUPPORT_PACKAGE -require-live
go run ./ai/live/cmd/supportmatrix -matrix ai/live/support-matrix.json NEW_SUPPORT_PACKAGE
```

The output directory must be new; protocol/effect packages are copied unchanged.
The outer manifest, final report, independent audit and full inventory bind both
subpackages. Every verify/merge checks actual captures, identities, mandatory
capabilities, source set, fixtures, budget, usage, Observer, metrics, audit and
inventories. Standalone self-reported JSON cannot establish support. Duplicated
JSON fields, symlinks, unlisted/missing files and refreshed checksums on semantic
corruption are rejected. A batch containing any refused report leaves the matrix
unchanged. Ordinary verification accepts controlled pipeline evidence; strict live
verification and matrix merging always reject it.

Other model rows retain their exact historical states and timestamps. Their old
Hash is not proof for this current catalog. Overall release acceptance belongs to
ticket 14, and existing external release blockers remain independent. Rollback
removes the explicit 1.14 Binding/model authorization; it never falls back to 1.13
or another provider and never erases already reported usage or raw evidence.
