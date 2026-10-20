# Ticket 09 public Client failure plan

Seam: spec T01's already approved Client Stream/Complete/simple/WithHooks;
controlled Host resolvers, official-protocol Provider scripts, ResourcePolicy,
HTTP body tracker, Admission and metadata Observer. No private helper tests.
All RequestIDs are globally unique. Two tenants reuse binding/model/session
names but have different credentials/accounts. Input mutation occurs only after
a visible snapshot barrier, never concurrently with a library map read.

Vertical slices:
1. Chat final capability boundary: text-only native containers, strict at actual
   function container, hosted tools despite Host allowance, developer role,
   stream:false and unintegrated cache/service controls can reach the Provider.
   Assert invalid_request/Request, zero requests/Attempts; model unauthorized
   together with invalid capability must remain tenant_denied. Catalog compat
   must not enable endpoint capabilities. Ordinary tool property names remain
   data. Official Chat OpenAPI retrieved 2026-10-17 is recorded in sources.
2. Shared Client: both models and each GLM protocol independently generate real
   tool history, replay trusted state, downgrade untrusted/cross-account/API/
   model state while retaining visible text and paired tools, reject cross-tenant
   retained references. Same session never shares credential/cache route keys.
3. Hooks: header→payload→response, legal modifications, authorization overrides,
   callback errors/panics and actual post-send Attempt; frozen legal payload on
   retries is inherited from ticket08 and rerun with affected coverage.
4. Lifecycle: cancellation before headers and after valid usage/content; deadline,
   Admission rejection, request/error-body/frame/tool/output/queue limit. Result
   waits independently of consumer, exactly one terminal, body and permit cleanup.
   Responses before terminal has unreported usage; already settled terminal stays
   successful after cancellation. No claim of remote cancellation or billing.
5. Copies: catalog/options/request/hook-retained JSON/results cannot mutate an
   ongoing or sibling call after snapshot. Observer sees metadata only.
6. Design load: six concurrent streams (two tenants × three routes), two models,
   synthetic 1 KiB prompt, ≤4 KiB tool args and ≤32 KiB output, bounded finite
   policy; record memory/headroom, bodies/Admission/goroutines plus race artifacts.
