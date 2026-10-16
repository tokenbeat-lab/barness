# Jev question-local instruction images

Host candidate catalogs can explicitly enable `instructions` on fixed
`jev-1.14.0`, alongside any implemented `context` and `state` locations. The
Binding must allow classifier images, the Client must have finite classifier
image budgets, and the call must select `TypeSafeOptions{ExperimentalImages:true}`.
The built-in catalog and formal support status are unchanged. This slice has
controlled HTTP evidence; real-service protocol and visual quality are NOT_RUN.

```go
req := ai.ClassifierRequest{
    State: json.RawMessage(`{"ticket":"example"}`),
    Questions: map[string]ai.ClassifierQuestion{
        "match": ai.BoolQuestion{Instructions: json.RawMessage(`["Compare this example",null,"with the ticket"]`)},
        "urgent": ai.BoolQuestion{Instructions: json.RawMessage(`"Is the ticket urgent?"`)},
    },
    Images: []ai.Image{exampleImage}, // Valid inline PNG/JPEG, with MIME.
    ImageReferences: []ai.ImageReference{{
        ImageIndex: 0, Placement: ai.ImagePlacementValue,
        Pointer: "/questions/match/instructions/1",
    }},
}
```

Choice, Score and Bool support the same Instructions shapes. A referenced
Instructions root can be `null` for an image alone. Other null/empty Instructions
remain invalid. Objects/arrays retain their text format requirements through
`DescriptionFormats`; a root image is governed by `ImageLocations` independently.
Pointers use RFC 6901 escaping, select existing unique null values and never
replace ordinary text. Criteria images remain rejected until their own slice.

The Adapter inserts a private whole-value reference at that exact local slot;
other questions receive no added local reference and State stays in its original
position when there is no context placement. Explicit context still wraps State
as documented by the context implementation. A source uploads once, but every
occurrence, including reuse across questions, consumes reference count and
expanded decoded-byte budgets. Generated escaped keys count toward the full
question budget before slot expansion. Decisions rejects this request at
capability before credentials; it cannot preserve local scope.

Final Hook validation uses final native questions and actual image occurrences.
ImageInputs lists distinct final assets in first-use order: State (including
context envelope order), then sorted question names, then Instructions object
keys in sorted order and arrays in input order. It contains only MIME and decoded
SHA-256; its indexes need not match Images. Legal renames, retyping and image
replacement are reflected in answer validation and these identities.
An existing context envelope may retain an explicit empty `images:[]` when a
Hook removes global uses and keeps local images; missing/null envelope fields
remain invalid.
Answers and ImageInputs publish atomically; failures preserve valid reported usage.
Shared Observer metadata excludes image contents, private references and hashes.

Disable the Binding, remove the model allowlist entry or revoke classifier image
permission to stop new calls. Already pinned calls keep their snapshot; code
rollback requires no migration. See [ticket 06 verification](../../.scratch/jev-1.14-multimodal-box/06-verification.md).
