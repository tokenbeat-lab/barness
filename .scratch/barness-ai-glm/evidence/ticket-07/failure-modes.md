# Ticket 07 failure modes — before implementation

Confirmed seam: spec T01 public ai.Client, controlled local Provider, trusted Host.
The implement-spec instruction explicitly calls tdd; this seam is already approved.
No paid calls or .env reads in this ticket.

1. Exact model-specific proven 5.2 enabled/budget1024/max4096 is rejected, inherits display/beta, or silently changes its budget. Canonical levels and all other explicit shapes must stay rejected.
2. Returned thinking/signature/tool JSON fragments lose content, order or ID; Host validation cannot execute a call; tool_result fails to preserve pairing and final answer.
3. Trusted unsigned thinking becomes text or gains an empty/fabricated signature instead of field absence. A stored unsigned block bypasses native envelope checks, allowing another account/model/API or an untrusted record to replay native state.
4. Signed thinking fails direct and Host-restored replay; redacted thinking is sent as proven GLM support despite no own evidence. Untrusted redacted data must disappear on downgrade.
5. Forced/any/none choices, arbitrary budget/effort, nested extensions or callback overrides reach network. Authorization errors must retain precedence over capability errors. A schema/input property named display must remain ordinary tool data.
6. Tool JSON deltas at different indexes mix; invalid event sequence, delta kind, duplicate/open blocks or after-terminal content count as success. Keep native Anthropic's pinned parser behavior intact.

Work proceeds one public behavior slice at a time; every changed behavior must have a retained red run before implementation. Raw wire sources are own Messages bundles 03-tools-thinking and 04-default-unsigned, indexed by messages-exploration.
