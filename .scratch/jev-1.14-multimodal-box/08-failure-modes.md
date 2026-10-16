# Ticket 08 failure inventory (before implementation)

Baseline: `8ecfdecb8a653e59199fa22378fa75a63939a38b`, branch `main`.
Dependency 06 is resolved. Agreed seams: public `Client.Classify` and
`HookedClient.Classify`, with the existing trusted Host and controlled HTTP
Provider/evidence recorder. Public JSON is exercised as input to these calls.
No implementation-afterward unit tests. No new dependencies.

1. Closed Box variant missing, incorrect public/native discriminator or fields,
   criteria accepted, empty/missing/null/primitive instructions accepted without
   an explicit root image, noncanonical field aliases/duplicates accepted.
2. Box cannot mix with Choice/Score/Bool, low/zero existence discards a valid
   box, confidence becomes an existence probability, coordinates are rounded,
   clamped or transformed, image ID/not-found/refusal is invented.
3. Missing/null/non-number/overflow/out-of-range coordinate accepted; negative,
   >1, inverted or zero-area edges accepted; wrong/nested/unknown fields,
   duplicates/case aliases or trailing JSON bypass strict validation.
4. Missing confidence becomes null/zero; explicit null becomes zero; valid zero
   disappears; wrong type, overflow, negative or >1 confidence accepted.
5. No image, only another question's instructions/criteria, or Box's forbidden
   criteria supplies visibility. Each Box must independently see context,
   State or its own Instructions, including nested/escaped paths.
6. Box bypasses disabled image policy, Binding image denial, experiment selection,
   MIME/location/file/source/occurrence/question/body budgets or experimental
   header normalization; preflight denials read credentials/create Attempts.
7. Catalog advertises Box on Decisions, Jev 1.13, alias/unconfirmed version or
   text-only candidate. Missing model kind support falls back or reaches pin.
8. Hook removes all images, leaves only other-question images, or changes kind
   without using final questions for validation; Box criteria accepted natively.
   Legal four-kind retyping/renaming/reordering/image replacement is rejected,
   or input raw instructions/native callback data remain aliased.
9. Bad one-of-many Box publishes partial answers/image metadata, loses legal
   usage/response identity, retries a 2xx protocol failure or leaks images, keys,
   file hashes or answers to Observer/diagnostics. Existing ordinary answers,
   rounding and Decisions refusal regress.
10. Evidence lacks fixture/file hashes, input/final wire, responses, outcomes,
    errors, Attempts/Observer, assertions/config/Host reads/resources or audit;
    inventory cannot detect tampering; controlled HTTP claimed as real PASS.

Public contract: sealed BoxQuestion with nonempty Instructions; BoxAnswer has
Box{Top,Bottom,Left,Right} and required Nullable[float64] Confidence. Native
records own final JSON independently. Only explicitly declared fixed
`jev-1.14.0` image candidates may advertise Box. Built-in catalog/support/Pi
remain unchanged. Host owns multi-image attribution, thresholds and transforms.
