# 22: DeepSeek × Chat Completions（P06）

**What to build:** 租户可通过独立的 DeepSeek Chat binding 完成文本、工具往返与 reasoning_content 回放；与 DeepSeek Responses 可经授权引用同一账户凭据，但两个协议的 binding 与模型配置独立（spec I4、P06）。

**Blocked by:** 20

**Status:** ready-for-agent

- [ ] reasoning_content 增量归一，以及带工具调用/结果的 assistant 回放
- [ ] [DONE] 前 usage 不遗漏
- [ ] 不混用 Responses 配置；同账户双协议 binding 用例证明 Provider/API/model 与 AccountScopeID 正确
- [ ] thinking 模式下不支持强制 tool_choice=required：强制工具用例在关闭 thinking 时运行，thinking + 工具自动选择与推理历史回放另跑，二者不互相证明
- [ ] 复用 E01–E09、E11 适用用例全部通过；适用部分接入 pi 差分无待处理差异

## Comments

**2026-10-02 — handover from issue 20.** The Chat adapter implements only the standard OpenAI compat (ADR-0013 决策三). pi's `detectCompat` gives DeepSeek (provider `deepseek` or a `deepseek.com` base URL) `max_tokens` instead of `max_completion_tokens`, no `store`, no developer role, `thinkingFormat: "deepseek"` (`thinking: {type: enabled|disabled}` plus `reasoning_effort`) and `requiresReasoningContentOnAssistantMessages` (an empty `reasoning_content` on every replayed assistant message of a reasoning model); none of these exists yet. Usage already reads `prompt_cache_hit_tokens`, and `reasoning_content` streaming and replay are in place.
