# Jev Box and independent existence answers

An explicitly configured Host image candidate for fixed `jev-1.14.0` can declare
`ClassifierQuestionBox`. Both `Client.Classify` and `HookedClient.Classify` support
Box mixed with Choice, Score and Bool in one batch. This is controlled HTTP
protocol evidence; real-service protocol and visual quality remain NOT_RUN.
The built-in catalog and formal support status are unchanged.

```go
req := ai.ClassifierRequest{
    State: json.RawMessage(`{"photo":"the shared search image"}`),
    Questions: map[string]ai.ClassifierQuestion{
        "location": ai.BoxQuestion{Instructions: json.RawMessage(`"Locate the mug in the shared image"`)},
        "exists": ai.BoolQuestion{Instructions: json.RawMessage(`"Is the mug present in the shared image?"`)},
    },
    Images: []ai.Image{photo}, // Valid inline PNG/JPEG. Omitted refs use context.
}
result, err := client.Classify(ctx, scope, target, req,
    ai.TypeSafeOptions{ExperimentalImages: true})
if err != nil { return err }
box := result.Answers["location"].(ai.BoxAnswer)
exists := result.Answers["exists"].(ai.BoolAnswer)
confidence, hasConfidence := box.Confidence.Get()
// box.Confidence.IsNull() identifies explicit null; numeric zero has
// hasConfidence=true. The Host chooses how to use existence and confidence.
_ = exists; _ = confidence; _ = hasConfidence
```

Box Instructions must be a nonempty JSON string, object or array, or an explicit
root image reference. Box has no criteria. Each Box must see at least one image
in context, State or its own Instructions; images only in another question's
Instructions or Criteria do not qualify. This proves availability, not which
image is the search image. Host retains target/image association and may limit
search images to remove ambiguity.

`BoxAnswer.Box` has four finite normalized edges: Top/Bottom are fractions of
height and Left/Right of width, measured from the top-left. All four fields are
required and range from 0 to 1; Left < Right and Top < Bottom. Confidence is
required, accepting explicit null or a finite number from 0 to 1; missing,
wrong-type, out-of-range, degenerate and inverted answers fail the whole call.
A returned Box is a value, with public JSON `type:"box"`, `box` and `confidence`.
Serializing an unset Confidence or invalid rectangle fails instead of inventing
an explicit null. No image ID, not-found result or existence probability is added.

A valid Box remains returned when the separate BoolAnswer reports low or zero
existence. The library applies no threshold and performs no conditional
execution. Host owns whether to consume the Box, the actual display orientation,
image dimensions and crop/resize transformations. These semantics follow the
[official Box contract](https://docs.typesafe.ai/primitives/box); the required
nullable confidence follows the OpenAPI/SDK evidence in the
[research record](../../.scratch/jev-1.14-multimodal-box/research.md).

Candidate catalogs require the fixed Jev version, image input and at least one
visible location (context, state or instructions) before advertising Box.
Image policy, finite budgets, MIME/location capabilities, Binding image authority
and the explicit experimental option all apply. Decisions and older Jev reject
Box at capability before credentials; no model/protocol fallback occurs.
An original Box without a visible image fails `invalid_request/scope`.

Final Hook questions are frozen independently from the public request. Legal
retyping among supported kinds follows the final names, instructions, criteria
and images; removing the last image visible to a final Box fails
`callback_failed/request`. Permission and experimental-header guards remain in
force. ImageInputs describe final assets by MIME/file SHA-256 and deterministic
first use; their indexes are not Box image attribution. Any invalid answer
withholds all Answers and ImageInputs, retaining legal usage/response identity.
Observer and shared diagnostic metadata omit content, hashes and private keys.

Disable the Binding, revoke image permission or remove model authorization to
stop new calls. Already pinned calls retain snapshots. Code rollback requires
no data migration. See [ticket 08 verification](../../.scratch/jev-1.14-multimodal-box/08-verification.md).
