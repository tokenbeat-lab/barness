# MiniMax Messages synthetic protocol fixtures

Authored 2026-10-18 for spec R15/R16/R19/R20. These files are independent,
controlled Messages event scripts, not real MiniMax captures or copied Claude
fixtures. `MODEL` is replaced with the independently selected candidate ID.
All prompts, output, signatures, tool IDs, usage counts and request IDs are
synthetic. Candidate prices are explicitly synthetic test rates.

The protocol field vocabulary and request shapes were checked against the
2026-10-18 official OpenAPI snapshot:
<https://platform.minimax.io/docs/api-reference/text/api/openapi-chat-anthropic.json>
(SHA-256 `867ac2ff2f6ceb9da629bdeb2c0172c6af95a25864adc3323e643229eddd45b0`),
plus the frozen local spec/research. Public enum omissions do not establish
that a stream feature is unsupported.

- `text.sse`: message_start input/cache counts, ping, text start/delta/stop,
  authoritative output usage in message_delta and message_stop.
- `tools.sse`: thinking and three signature fragments, text, two tools with
  independently interleaved partial JSON and reverse stop order. This is a
  decoder stress script, not a claim that the vendor interleaves tool blocks.
- Test mutations cover faults, complete start-block input, empty placeholder
  input plus fragments, unsigned/empty signatures and opaque redacted output.

Only trusted signed thinking replays natively. Unsigned visible thinking is
plain text; redacted output is retained but same-origin native replay fails.
Nonempty complete start input plus later partial_json is rejected pending
verified service semantics. Real M3 stream/signature/tool continuation, actual
request ID headers and M3.1 qualification stay separate ticket07 obligations;
these synthetic results cannot certify any live capability.

Replay and immutable assertion/event/Result/Attempt/Observer bundles:
[Current international evidence](../../../../.scratch/barness-ai-minimax/evidence/ticket-11/README.md).
[Ticket04 evidence](../../../../.scratch/barness-ai-minimax/evidence/ticket-04/README.md) retains its original historical CN applicability.

International applicability was rechecked for ticket11 using independently
retrieved2026-10-18 `.io` snapshots. The three OpenAPI structural schemas match
the historical CN shapes after prose metadata is excluded; this is not live
wire/account/pricing proof. Historical CN captures and source hashes remain
immutable in their original evidence archives; no CN native origin qualifies
an international continuation. See ADR-0042 and research-international.md.
