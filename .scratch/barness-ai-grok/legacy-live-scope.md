# Grok: remaining 22 old-target live routes

Read-only planning against canonical repository `/Users/cyber/RestoX/barness`, requested integration `c70de780`, 2026-10-26. No tests, builds, vendor calls, `.env` access, grant issuance or repository mutations. Only this `/tmp` report is written. Endpoints/models below are the **existing reviewed native-contract constants**, not a fresh vendor-document or account-entitlement check.

## Main finding

**None of the 22 current native smokes proves the new xAI-origin collision/reuse trigger.** Existing history/tool scenes feed the selected Provider's own returned assistant back to itself. `ProviderXAI`/`xar1` appears in live code only in the xAI runner/selector; the old family scene builders do not construct a foreign xAI assistant. Generic “tool round-trip” or “native-history” PASS therefore does not cover E18's changed condition.

**The current formal record requires a complete own native capability report, not a bounded delta-only probe.** `release/record_live.go:16-75` requires a qualified own report, exact fixed expected set, all PASS, current full Builtin hash and SDK. `record_verify.go:211-232` requires exactly one primary live bundle and calls `verifySelectedLive`; `record_live.go:83-93` requires matching record commit and exactly one actual outer native-consumer receipt. Adding only a tiny one-case JSON/receipt cannot satisfy it. A bounded feature-only contract would require an explicit separately reviewed native/record change; it is not current acceptance and must not silently shrink old full capability sets.

Do not buy the 22 routes yet. Prepare the additive scenario, native assertions, fresh per-process authority gaps and final budgets first; then make per-route approval concrete. Existing grants, closed ledgers and candidate evidence confer no permission.

## Exact routes and existing entries

SDK symbols: **O** = `github.com/openai/openai-go/v3 v3.66.0`; **A** = `github.com/anthropics/anthropic-sdk-go v1.75.0 (HTTP; SSE decoded by barness-ai, ADR-0011)`; **D** = `direct HTTP (ADR-0012)`. Versions are from current go.mod and native SDKVersion helpers. The contract tag in each row identifies the exact complete expected set, budgets, authority and native consumer below; none includes the future xAI-origin scene yet.

| Route | Official SDK base | Native models | SDK | Existing fixed contract | Exact current `-test.run` | Only credential ENV |
| --- | --- | --- | --- | --- | --- | --- |
| `openai-responses` | `https://api.openai.com/v1` | `gpt-5-mini; tool extras gpt-5.4,gpt-6-sol` | `O` | `G-R (8)` | `^TestLive$/^openai-responses$` | `BARNESS_AI_LIVE_KEY_OPENAI_RESPONSES` |
| `deepseek-responses` | `https://api.deepseek.com` | `deepseek-flash` | `O` | `G-DR (8)` | `^TestLive$/^deepseek-responses$` | `BARNESS_AI_LIVE_KEY_DEEPSEEK_RESPONSES` |
| `openai-chat` | `https://api.openai.com/v1` | `gpt-5-mini; extras gpt-5.4,gpt-6-sol,gpt-5.4-pro(unavailable)` | `O` | `G-C (10)` | `^TestLive$/^openai-chat$` | `BARNESS_AI_LIVE_KEY_OPENAI_CHAT` |
| `deepseek-chat` | `https://api.deepseek.com` | `deepseek-flash; effort extras deepseek-v4-pro` | `O` | `G-DC (11)` | `^TestLive$/^deepseek-chat$` | `BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT` |
| `zhipu-chat` | `https://open.bigmodel.cn/api/paas/v4` | `glm-5.3,glm-5.2` | `O` | `L-chat (22)` | `^TestGLMFormal$` | `BARNESS_GLM_KEY` |
| `zai-chat` | `https://api.z.ai/api/paas/v4` | `glm-5.3,glm-5.2` | `O` | `R-Z (18)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `minimax-chat` | `https://api.minimax.io/v1` | `MiniMax-M3 / standard` | `O` | `M (13)` | `^TestMiniMaxFormal$` | `BARNESS_MINIMAX_KEY` |
| `minimax-cn-chat` | `https://api.minimax.cn/v1` | `MiniMax-M3 / standard` | `O` | `R-M (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `kimi-cn-chat` | `https://api.moonshot.cn/v1` | `kimi-k3 / standard` | `O` | `K (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `kimi-intl-chat` | `https://api.moonshot.ai/v1` | `kimi-k3 / standard` | `O` | `K (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `anthropic-messages` | `https://api.anthropic.com` | `claude-haiku-4-5; extras claude-opus-4-8,claude-fable-5,claude-fable-5-1,claude-opus-5,claude-opus-5-5` | `A` | `G-A (11)` | `^TestLive$/^anthropic-messages$` | `BARNESS_AI_LIVE_KEY_ANTHROPIC_MESSAGES` |
| `zhipu-messages` | `https://open.bigmodel.cn/api/anthropic` | `glm-5.3,glm-5.2` | `A` | `L-messages (22)` | `^TestGLMFormal$` | `BARNESS_GLM_KEY` |
| `zai-messages` | `https://api.z.ai/api/anthropic` | `glm-5.3,glm-5.2` | `A` | `R-Z (18)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `minimax-messages` | `https://api.minimax.io/anthropic` | `MiniMax-M3 / standard` | `A` | `M (13)` | `^TestMiniMaxFormal$` | `BARNESS_MINIMAX_KEY` |
| `minimax-cn-messages` | `https://api.minimax.cn/anthropic` | `MiniMax-M3 / standard` | `A` | `R-M (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `kimi-cn-messages` | `https://api.moonshot.cn/anthropic` | `kimi-k3 / standard` | `A` | `K (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `kimi-intl-messages` | `https://api.moonshot.ai/anthropic` | `kimi-k3 / standard` | `A` | `K (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `google-gemini` | `https://generativelanguage.googleapis.com/v1beta` | `gemini-3.8-flash` | `D` | `G-G (6)` | `^TestLive$/^google-gemini$` | `BARNESS_AI_LIVE_KEY_GOOGLE_GEMINI` |
| `minimax-responses` | `https://api.minimax.io/v1` | `MiniMax-M3 / standard` | `O` | `M (13)` | `^TestMiniMaxFormal$` | `BARNESS_MINIMAX_KEY` |
| `minimax-cn-responses` | `https://api.minimax.cn/v1` | `MiniMax-M3 / standard` | `O` | `R-M (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `kimi-cn-responses` | `https://api.moonshot.cn/v1` | `kimi-k3 / standard` | `O` | `K (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |
| `kimi-intl-responses` | `https://api.moonshot.ai/v1` | `kimi-k3 / standard` | `O` | `K (9)` | `^TestRegionalFormal$` | `BARNESS_REGIONAL_KEY` |

All R/K and L formal entries use `-test.timeout=1800s`. M original formal also has 1800s/20s native lifetime/reserve; do not select historical fixed recovery7's special 3600s allowance. G has no native whole-process limit; a concrete future grant must choose and enforce an external test/process bound (e.g. 1800s), not infer one from its 4-minute per-scene constant. The `-test.run` expressions above are entry selectors, not authorization.

Sources: `ai/live/combos_test.go:93-168`; `internal/testkit/supportmatrix/regional_contract.go:13-53`, `regional_kimi.go:16-43`; `internal/testkit/glm/{catalog.go:27,messages.go:26}`; `internal/testkit/minimax/catalog.go:41-45`; `internal/testkit/minimaxlive/plan.go:25-43`. File prefixes in this report are under `ai/` unless stated otherwise.

## Existing full expected sets and budgets

Common generic base **B** = `text-stream,text-complete,tool-round-trip,cancel-after-first-frame,reasoning-history,image`.

| Tag | Fixed full expected scene IDs | Calls / attempts / retries | Output and input caps | Native time bounds |
| --- | --- | --- | --- | --- |
| G-R | B + `tool-changes-gpt-5.4,tool-changes-gpt-6-sol` (8) | max32 Logical Calls; HTTP≤calls; max1 environment retry per scene | 4096 output/call; no separate aggregate input reservation/USD cap. Implied max requested output ≤131072 only, not invoice cap | scene240s; Client call180s; no native whole-process cap |
| G-DR | B + `reasoning-level-high,reasoning-level-max` (8) | G | G | G |
| G-C | B + `usage-position,tool-changes-gpt-5.4,tool-changes-gpt-6-sol,pro-model-unavailable` (10) | G | G | G |
| G-DC | B + `usage-position,forced-tool-with-thinking,prompt-cache-long-retention,reasoning-level-high,reasoning-level-max` (11) | G | G | G |
| G-A | B + `tool-changes-claude-opus-4-8,tool-changes-claude-fable-5,effort-changes-claude-fable-5-1,effort-changes-claude-opus-5,effort-changes-claude-opus-5-5` (11) | G | G | G |
| G-G | B (6) | G | G | G |
| L-chat | For each glm-5.3/5.2: nine base kinds below + `control-temperature,control-usage` =22 | planned22/max24 calls; max32 attempt ceiling but one reserved attempt/call and zero retries/replays; actualHTTP≤reservedCalls | max4096 output/call; no native aggregate requested-output/reserved-input/USD cap | call120s, process1800s, reserve20s |
| L-messages | Each glm-5.3/5.2: nine base kinds + `control-unsigned`; plus 5.3=`control-auto`, 5.2=`control-default` =22 | L | L | L |
| M | `full-stream,full-complete,simple-stream,simple-complete,thinking-on,thinking-control,tool-call,tool-result-native,truncate,cache-first,cache-repeat,cancel,preflight-error` (13) | own cumulative max20 calls/24 reserved+actual attempts; fixed finite reruns; original formal stage shared cap60calls/72attempts/3identities | per scene output1024,1024,1024,1024,2048,1024,4096,2048,1,1024,1024,2048,1; max4096/call, cumulative40000 output,120000 reservedinput peridentity | call120s; process1800s; reserve20s (special historical fixed7 excluded) |
| R-Z | `glm-5.3/<nine>,glm-5.2/<nine>` =18 | max24calls/24attempts; zero retries | max4096/call,98304 requestedoutput,120000 reservedinput; input reserve encodedwirebytes+1024,min2048 | call120s; process1800s |
| R-M | `MiniMax-M3/<nine>` =9 | R-Z | R-Z | R-Z |
| K | `kimi-k3/<nine>` =9 | **max9calls/9attempts**, zero retries | max4096/call,36864 requestedoutput,131072 reservedinput; finite positive MaxUSD/binarySHA required, conservative input-side reservation; optional explicitly granted call interval0..60000ms | call120s; process1800s |

Nine base kinds = `full-stream,full-complete,simple-stream,simple-complete,tool-call,tool-result,native-history,cancel,recovery`; Kimi orders native-history before tool-call but IDs are the same. L additionally validates exact source plan and controls; do not substitute a new xAI scene for an existing native-history/tool scene.

Sources: `ai/live/main_test.go:22-43,341-353`, `ai/live/session_test.go:33-38,99-130`; `internal/testkit/release/live.go:14-50`, `chat_qualification.go:49-85`; `internal/testkit/supportmatrix/glm.go:25-49,90-101`, `ai/live/glm_config_test.go:113`; `internal/testkit/minimaxlive/plan.go:62-80`, `ledger.go:113-135`; `internal/testkit/supportmatrix/regional_contract.go:49`, `regional_kimi.go:67-72`.

## Authority, environment and native consumers

All new executions require human scope and one selected route key. No values/keys were read. Common outer producer is `tools/verification/run.py --feature barness-ai-grok --stage live --route <route> -- <frozen native binary> <exact selector> <explicit timeout>`. Follow the current selected_environment forwarding contract; it never creates authorization.

| Family | Current selectors / grant or ledger contract | Native qualification consumption |
| --- | --- | --- |
| G six routes | `BARNESS_AI_LIVE=1`, `BARNESS_AI_LIVE_COMBO=<route>`, `BARNESS_AI_LIVE_ACCOUNT_ALIAS`, exact key from route table; optional `BARNESS_AI_LIVE_PROXY`. **No one-use grant/parser/source-binary binding** for generic G today. PERMISSIONS_FILE/REGION_ALIAS are forwarded by outer tool but generic loadConfig does not consume them. | Producer writes live-report; compiled `releasegate -live-bundle <bundle> -out <newdir>`, captured through `consume_live.py` schema2. Native generic full report/capture qualification applies. |
| L zhipu chat/messages | `BARNESS_AI_LIVE=1`, `BARNESS_GLM_ROUTE=chat|messages`, `BARNESS_GLM_FORMAL=1`, `BARNESS_GLM_KEY`. **No one-use grant input**; default account/region aliases are hardcoded local-glm/cn-standard-endpoint and verification unknown. No outer forwarding for configurable account/region/proxy. Config rejects BARNESS_AI_LIVE_PROXY despite transport helper, so current actual contract is direct-only. | Formal producer independently seals/publishes `live-report.json`; `supportmatrix.NewGLMReport` then generic compiled releasegate/consume_live. `glmformal -route ... -out ...` is historical support-view derivation, not a fresh vendor-run qualifier or permission. |
| M intl MiniMax three | `BARNESS_AI_LIVE=1`, `BARNESS_MINIMAX_ROUTE=chat|responses|messages`, `BARNESS_MINIMAX_MODEL=MiniMax-M3`, `BARNESS_MINIMAX_PURPOSE=standard`, `BARNESS_MINIMAX_FORMAL=1`, `BARNESS_MINIMAX_ACCOUNT_ALIAS`, `BARNESS_MINIMAX_REGION_ALIAS`, `BARNESS_MINIMAX_ENTITLEMENT=confirmed`, key. Optional `BARNESS_MINIMAX_PROXY`. Existing native authority is **common-Git cumulative stage ledger plus historical fixed renewal selectors**, not a new per-process generic grant. Actual live rejects alternate ledger path. Old cohorts are independently closed/pinned; no historical ID/leftover budget may authorize this feature. | compiled `minimaxverify -formal <bundle>` (and publish projection if needed with its existing explicit publish mode), then `releasegate -live-bundle` via consume_live. VerifyCapture checks exact plan/seal/current execution set/SDK/builtin/entitlement. |
| R-Z,R-M,K eleven | `BARNESS_AI_LIVE=1`, `BARNESS_REGIONAL_ROUTE=<exactroute>`, `BARNESS_REGIONAL_SERVICE=zai|minimax-cn|moonshot-cn|moonshot`, `BARNESS_REGIONAL_STAGE=formal`, `BARNESS_REGIONAL_GRANT=<file>`, `BARNESS_REGIONAL_KEY`; optional `BARNESS_REGIONAL_PROXY`. Fresh `RegionalGrant`: schema1,id,mode,stage,endpoint,sourceSHA,catalogVersion/hash,SDK,contract,combo,provider,purpose=standard,account/regionAliases,models,expiry≤24h,exactcall/attempt/output/input caps. K additionally binds binarySHA,positiveMaxUSD,optional pacing. Contracts `provider-sites/<combo>/standard/v1` or `kimi/<combo>/standard/v1`. Template via `regionalverify -grant-template <combo> -stage formal`; immutable common-Git reserve/claim before requests. | compiled `regionalverify -formal <bundle>` -> own full report/recorded source+catalog+grant/capture/independent protocol replay checks -> releasegate/consume_live schema2. NewRegionalReport formal support requires stageformal, not candidate. |

Grant/config paths and raw approval records must be immutable selected inputs alongside case selection; no key in argv/source snapshot. Final schema2 native consumer must be captured at the paid HEAD; later pure prose/index closure can reuse it with exact source/input equality as described in previous qualification notes. All passing record stages require one Source.Commit and identical selected-input maps (`record_verify.go:116`); plan all route grant/config inputs before the final shared freeze, or accept that later source/input changes require a new honest applicability assessment.

Sources: `ai/live/config_test.go:15-18,37-39,86-119`; `ai/live/glm_config_test.go:112-159`; `ai/live/minimax_config_test.go:27-148,188-200`; `internal/testkit/minimaxlive/stage_path.go:13-57`, `renewal_ledger.go:12-42`; `ai/live/regional_config_test.go:32-150,159-165`; `internal/testkit/supportmatrix/regional_grant.go:15-80`; `tools/verification/live_environment.py:7-36`; CLI files under `ai/live/cmd/{regionalverify,minimaxverify,glmformal}`.

## The missing xAI-origin proof

Actual changed code: `ai/opaque_history_ids.go:16-28` opts in only when source classification includes xAI; `:46-114` allocates over the complete transcript; `:139-147` is the added xAI source classification; `:156-159` includes assistant turn in provenance owner. Own-provider history cannot exercise this new source predicate.

Minimal fixed new scene should send **one actual old-destination request** with a prepared Host transcript that combines:

1. Two distinct valid xAI opaque `[call_id,item_id]` pairs in one assistant turn, deliberately arranged to collapse under the destination's old prefix/sanitizer/truncation rule. Use a sufficiently long common prefix for64char Messages/Gemini/Chat normalizers, not random live IDs that might not collide.
2. The identical full xAI pair reused in a later assistant turn, with a different visible result marker. Owner includes turn, so the allocator must distinguish them even for injective MiniMax/Kimi Responses destination codecs.
3. Correct ordered ToolResult associations, a normal visible answer/reasoning marker, and an occupied legacy spelling if desired. The reuse case already exercises collision replacement when generated spelling is occupied by the first owner.
4. A final ordinary user instruction and supported tool policy to obtain a bounded successful fresh result. Synthetic source input is legitimate Host-supplied cross-provider history; **no extra xAI key/request is needed**. Derive source messages from public Client replay fixtures or tracked stored-message fixtures, rather than calling private allocator helpers.

Observe final sent wire and successful own target Result/Observer. Assert nonempty legal unique destination call IDs, each result points to its corresponding call and marker in order, no unauthorized foreign native item/resource fields or ciphertext leak, visible content preserved, source history immutable, actual metadata targets selected Provider/API/model/account, and expected native downgrade accounting. Protocol projections differ; verify their real fields, not one generic text search.

- OpenAI Responses: source `xar1|...` previously collapsed to shared xar1 prefix; require actual distinct call_id/output pairing, no account-specific item id.
- DeepSeek Responses: sanitizer/truncation collisions need the long common-prefix input.
- Chat eight, Messages seven, the one Gemini route: verify actual tool_calls/tool_call_id or tool_use/tool_result or Gemini functionCall/functionResponse IDs according to that codec.
- Four injective MiniMax/Kimi Responses targets: distinct-source parallel IDs alone is insufficient; **the repeated full pair in separate turns is required**. Preserve the existing destination opaque call-association codec; it may encode the complete public foreign ID. Do not confuse that safe correlation value with replaying a foreign native `id` resource field or ciphertext.

Current offline regression `ai/e2e/xai_history_downgrade_test.go:12` provides the OpenAI public pattern (parallel/separate-turns/occupiedlegacy). Current old live scenes (`scenarios_test.go:108,189`, `glm_harness_test.go:195-244`, `minimax_harness_test.go:187-226`, regional own native/tool steps) are exclusively own-source and do not cover it.

## Smallest coherent additive implementation ownership

| Owner | Files/responsibility | Required contract and verification additions |
| --- | --- | --- |
| Shared transcript fixture + protocol proof | New focused `ai/live/*grok_history*` input/proof helper(s), corresponding public E2E fixture/data | One combined scene above; fixed normalized source fixture, per-protocol actualwire assertions and provenance; keyless public RED/GREEN. Reuse existing Client/Host/recorder. Shared semantics have22realconsumers; no generic controller or callback registry. |
| Generic six routes | `scenariosOf`/`combos`, generic run/config/session, `release/live.go`, `chat_qualification.go`/capture checks | Append one fixed xai-origin-history capability at the published default smoke model; retain all base/extras. Existing32call cap hasroom. Add fresh one-process source/binary/catalog/expectedscene/budget/account grant+claim atselectedroute boundary; missing/duplicate/wronggrant refused preHTTP. Keep exact full expected set and reject candidate/controlled/partial proof. |
| GLM two routes | `glm_*` producer, fixed versioned plan/source manifest, `supportmatrix/glm*.go`, public process regressions | Add one combined scene for each of glm-5.3/5.2:22→24scene/call plan, fits current24call cap; revise exact native expected/plan/source pins openly, preserve original22scenes. Own native decoder verifies addedwirecase, no unknown-kind fallthrough. Fresh one-process approval/account binding needed; explicit proxy only if chosen scope needs it. |
| Regional ZAI/MiniMaxCN + six Kimi | `regional_*` producer; `supportmatrix/regional_contract.go`, `regional_kimi.go`, owned request/result/stream/replay consumers; grant/process public E2E | Add one combined scene per required native model: ZAI18→20; MiniMaxCN9→10, fit24caps. Kimi9→10 requires exact new10call/10attempt cap and requestedoutput40960; retain131072inputcap only after bounded fixture reservation is proven tofit. Recompute freshgranttemplates/source/binary/expectedsets. All old nine capabilities stay mandatory. |
| International MiniMax three | `minimax_*` producer + `minimaxlive/plan,ledger,verify`/support projection; public CLI/process E2E | Plan13→14 fits own20/24 and40000/120000 caps, but **new Grok authority must not resume historical stage/cohort**. Integrate a fresh bounded per-process approval/claim with existing own capture/reservation/verification, distinct from closed renewal authority. No new recovery generation, old grant reopening, moved ledger reset or fake current-source historical report. Keep M3 standard andentitlement qualification. |
| Integration/native release | Existing `record_live.go`, manifest/consumer bridge only where needed; inventory/currentstate/docs | Exactly one full own+addedproof bundle perroute and one actual schema2consumer, or explicitly design/review an alternative feature-delta native contract before claiming applicability. No shrinking full registries, UNKNOWN pass, auxiliaryprobepass relabel, or changingoldpins. Finalpayload/review/freeze before anypaidselection. |

The combined scene is a proposed minimal design, not an implemented capability. Literal newnames/scene counts and budgetchanges must be fixed in reviewed code and actual grant templates before approval. The table of existing entries is useful for ownership/routing but is **not yet a launch plan** for the added source condition.

Ordinary verification should include all22 old-target public controlled paths for the shared scene, process grant refusals/caps/claim consumption, producer→nativeconsumer complete/partial/tamper cases, and complete affected packages after finalreview. Retain RED/GREEN and immutable native artifacts. Paid scope should then include each route's complete fixed ownreport plus the additive source-conditionedcase, exactmodel/SDK/endpoint/aliases/keysource/finiteUSD/call/attempt/token/time limits. Stop onfailure; do not turn the old remaining budgets into authorization.
