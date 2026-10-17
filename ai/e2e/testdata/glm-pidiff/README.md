# China standard GLM extension evidence

Frozen pi-ai **1.1.0**, commit `abe508e1b89912adde45528136c3221eb69acdd7`,
model-data SHA256 `8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8`
and the existing SDK lock/source pins remain the only oracle. No custom GLM
model is injected into pi to manufacture same-route parity.

| barness Provider/API | Standard base | Frozen native route | Classification |
| --- | --- | --- | --- |
| zhipu/openai-completions | `https://open.bigmodel.cn/api/paas/v4` | absent | ExtensionRoute |
| zhipu/openai-responses | `https://open.bigmodel.cn/api/v1` | absent | ExtensionRoute |
| zhipu/anthropic-messages | `https://open.bigmodel.cn/api/anthropic` | absent | ExtensionRoute |

Each row covers independent `glm-5.3` and `glm-5.2` Host catalog entries.
The entire 43-file frozen native data scan has no `zhipu` identity or any of
these three standard bases. Its `zai` provider uses international Coding
`https://api.z.ai/api/coding/paas/v4`, Chat only, with both target models.
`zai-coding-cn` uses China Coding
`https://open.bigmodel.cn/api/coding/paas/v4`, Chat only; 5.3 is present and 5.2
absent. All four native China-host entries use that Coding base. Similar host,
model name, compatibility detection or generic adapter knowledge does not
establish the same Provider/account/service/API contract.

`TestGLMFrozenRouteIdentity` first asks the unchanged sole oracle to verify
all dist, model-data and SDK pins, then inventories **all** installed native
data. The oracle runner's narrower APIS query table alone is insufficient
proof of absence. [Exact hashes/source excerpts](../../../../.scratch/barness-ai-glm/evidence/ticket-10/sources/)
and [case artifacts/replay index](../../../../.scratch/barness-ai-glm/evidence/ticket-10/README.md)
retain the audit. Existing three extension route registrations and all native
differential decisions are preserved. ADR-0014's registration rule applies.

## Own fixtures and skipped native comparison

All of these fixtures intentionally remain outside native pi inference
comparison because the corresponding standard route is absent. None generates
`PIDIFF-GLM` cases; a successful own public test is **extension acceptance**,
never `piParity=PASS`.

| Own fixture | Independent source and limit |
| --- | --- |
| `../glm-chat/{text,tools}.json` | Official China Chat OpenAPI/model/thinking facts; synthetic SSE, two-model literals. No paid-wire claim. |
| `../glm-responses/{text,tools}.json` | Official China Responses Input/Output schemas; synthetic object/array/done/tool variants, separately marked own observed captures where present. Object output is schema evidence; early real output was array. 5.3 synthetic native-array history is not a real-service replay claim. |
| `../glm-messages/text.json`, `tools-thinking.json` and referenced request/SSE files | Independent model-specific bounded Messages observations, hashed unchanged. 5.2 default tool response has no thinking; exact enabled/budget1024/max4096 capture supplies its separate thinking scenario. Signature absence is a tested Host transformation of signed history, not observed unsigned output. |
| `../glm-failures/{http,stream,usage}.json` | Own protocol error/usage schemas and actual Messages usage placement; explicitly labeled synthetic failures/cache mutations. No deliberately paid billing/rate/error probes. |

The own public Client suite includes full/simple entries, hooks, request guards,
original Native State, complete stateless history, tools/results, usage presence,
errors and cross-protocol isolation for both models. OpenAI/Anthropic/DeepSeek
native differential PASS remains separate evidence. It cannot stand in for one
of these GLM tests.

## Finite case/path acceptance ledger

[ledger.json](ledger.json) uses the existing parser/classification vocabulary,
with `zhipu/<api>` scope and **exact public case IDs and leaf paths**. These are
source-backed behavior records for own extension acceptance, not invented
Pi/Barness observations. Paths use the public Result envelope (`result.message`
and `result.metadata`); request paths describe the captured wire. The referenced
public cases independently assert literals or own previously observed wire.
The evidence index maps every listed case to its manifest and hashed artifacts.

- `fixed`: object-content loss was corrected by04; recurrence stays pending.
- `extension`: the exact own GLM behavior is approved by the spec/source.
- `pending`: unresolved behavior blocks acceptance; there are currently zero
  such records. Unknown case/path or native protocol scope remains pending.

No whole request body, result, event list or protocol is exempted. Configuration
acceptance probes the bounds without pretending those probes are actual pi
findings. New observed differences require a new bounded record and public
assertion, or remain pending. The ledger is separate from central native
comparison decisions to prevent it from weakening their gates.

The predecessor Responses tool fixtures use delimiter-free `call_id`/item IDs.
Their recorded success does not establish arbitrary-string composite-ID safety.
The pinned InputFunctionCall.call_id and OutputFunctionCall.{call_id,id} are
plain strings with no `pattern`, `format` or alphabet excluding `|`; exact
schema fields are retained in the Responses source excerpt. The coordinated05 source-led correction has separate focused public regressions
for delimiter, percent, Unicode and ambiguous concatenation IDs, trusted/restored
state and cross-origin conversion. Their exact58 case IDs and four call/result
leaf paths are now fixed records; recurrence remains pending. Additional
cross-API collision and empty-origin-ID cases retain their own05 evidence scope.
See [follow-up evidence](../../../../.scratch/barness-ai-glm/evidence/ticket-05-followup-ids/README.md).

No built-in catalog, formal route registration, support matrix or live evidence
is promoted here. Candidate catalog rates remain synthetic. Tickets11–13 own
standard-key live, formal catalog/matrix and complete release evidence.
