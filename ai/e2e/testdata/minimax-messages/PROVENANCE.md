# MiniMax Messages synthetic protocol fixtures

Authored 2026-10-18 for spec R15/R16/R19/R20. These files are independent,
controlled Messages event scripts, not real MiniMax captures or copied Claude
fixtures. `MODEL` is replaced with the independently selected candidate ID.
All prompts, output, signatures, tool IDs, usage counts and request IDs are
synthetic. Candidate prices are explicitly synthetic test rates.

The protocol field vocabulary and request shapes were checked against the
2026-10-18 official OpenAPI snapshot:
<https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json>
(SHA-256 `57111d5dddd2ada7f481e5c373d88c745eba0b852b60bf1ca20096c61f28b77c`),
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
[Ticket04 evidence](../../../../.scratch/barness-ai-minimax/evidence/ticket-04/README.md).
