# Ticket 08 failure plan (before implementation)

Confirmed seam: spec T01 public Client Stream/Complete/full/simple/WithHooks,
controlled HTTP Provider, Host Bindings/credentials and independent candidate
catalogs. No paid calls, credential reads or deliberate upstream failures.

1. HTTP 401/403/400/429/5xx lose stable code/phase/status/request ID.
2. Responses string code, Chat numeric code and Messages nested error object
   are confused; billing fixtures are synthetic contract cases, not observed
   Messages billing wire. Malformed/unknown code uses HTTP fallback.
3. 1113/1308/1309/1311 billing/entitlement failures retry, including upstream
   x-should-retry:true. Other known permanent auth/parameter/policy failures
   must also suppress retries. 1302/1305 require explicit bounded policy.
4. SDK defaults retry; retries change credential or post-hook payload; no
   retry budget; stream error/EOF/cancel replays partial generation.
5. Sensitive upstream body/prompt/thinking/key echoes enter errors/messages;
   a provider code itself may contain unsafe arbitrary text.
6. HTTP200 error, Responses failed, Chat sensitive/network_error/context stop,
   Messages error object, or premature EOF appears successful.
7. Known usage is lost on interruption; omitted usage looks complete/free;
   complete zero looks unreported; cache/reasoning are double counted.
8. Responses absent total_tokens loses known sum; absent total stays partial;
   failed-response usage accidentally changes its legacy unreported contract.
9. Messages initial zero replaces later actual counts, absent cache creation
   looks reported zero, or missing reasoning tokens is invented.
10. X-Request-Id loses attribution, fake/DeepSeek/local ID fills an absent ID;
    observer contains raw upstream errors rather than metadata.

Each vertical test slice precedes its corresponding implementation. Existing
public lifecycle, usage and frozen protocol differential tests provide native
provider regression coverage. Each run emits repeatable audited evidence.

11. Parent review found Messages start fabricates CacheWrite1h=0 and raw
    Anthropic-shaped 1h detail activates a 2x-input TTL premium for GLM.
    Public presence/cost red cases precede the narrow provider-gated fix.
