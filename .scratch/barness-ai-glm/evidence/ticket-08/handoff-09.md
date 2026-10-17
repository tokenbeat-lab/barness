# Ticket 09 cancellation handoff

Ticket08 owns fail/EOF/usage/retry. Own-protocol fixture entry points:
`ai/e2e/testdata/glm-failures/{usage,stream,http}.json`, test factories
`glmFailureWorld` and `glmOwnReply` in public E2E package. `normal` usage and
`interrupted-known` use exactly the same legal usage sequence; EndAbort proves
known counts survive body-read interruption in each protocol without retry.

For09 cancellation: Chat can report usage before a later content delta/finish;
Messages message_start input/output0 and message_delta replacement precede
message_stop. Both permit a public content/event barrier after legal usage
and before cancellation. Messages cache_creation remains absent/partial,
Reasoning remains unset. Responses usage only enters on completed/incomplete
terminal: cancellation before any legal terminal has no reported usage;
never invent early usage by copying Chat placement. Ticket09 should own its
post-terminal settlement invariant/cancellation ordering. Do not label a
successful already-settled terminal as a cancellation failure merely because
context later ended.

For metadata: GLM HTTP200 stream failures now retain HTTPStatus200 and actual
X-Request-Id; missing provider IDs stay empty. Business/echo redaction is
provider-gated and Observer schema remains metadata only. Constructor/stream
callsite diffs are surgical; Messages thinking/tool and Responses history
request guards remain owned by07/05 respectively.
