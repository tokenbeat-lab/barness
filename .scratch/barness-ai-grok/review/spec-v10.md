## Spec

No findings. The v9 P2 is substantively resolved.

Issue 16:23 requires “并保留typed捕获所需身份/stop/error/Usage/CallMetadata约束的反例.” The positive fixture now includes the exact `grokhistory.Scenario` ID (`ai/e2e/generic_capture_decoder_test.go:152,172`), thereby entering `chat_capture.go:58–68`. Its Host request and provenance are retained, and the public Client sends the history through the loopback provider; the actual encoded request and CrossModel=2 are verified (`generic_capture_decoder_test.go:185–208`).

The six new cases mutate only the corresponding mandatory-history result: provider, API, model, stopReason, message.errorMessage and outer error (`:56–73`). They reseal inventory, then require rejection both through native audit and the compiled public CLI (`:81–116`); they do not merely add case names. The unchanged positive case passes capture audit before the unknown synthetic route is refused for qualification. Existing attribution, Observer-terminal, negative usage, budget and inventory counterexamples remain.

Read-only inspection of the retained GREEN attempt confirmed 12 native cases PASS, numeric exit 0, audit findings empty, and source hashes matching the reviewed test and production files. Receipt SHA256 `1ca4f1854937211f53b1854e49f21cdfa491f6e0a4ca07d401c3e8c4227c1946` and summary-v2 SHA256 `d5a832ab208396896fddd28922a3654b98628e3c09b59d96a84b63077d1efeb3` match their originals. The summary accurately distinguishes this directly GREEN coverage extension from the original decoder RED/GREEN.

The complete diff remains scoped to the consumer-local DTO, offline regression and evidence summaries. Public JSON contracts, authorization and formal qualification rules remain unchanged. Pending full ordinary validation/refreeze/live are not claimed complete.

Reviewed `git diff 3a44c514414f8d1753ae2f66a07814a44bd27cef...fd02b56aa6177106ae1ef348fec0c3a4920b8a57`. No tests, builds, vendor requests or source writes performed.

Findings: 0; worst severity: none.
