# 23: 真实 API 冒烟与支持矩阵

**What to build:** 维护者在发布前或 SDK/模型升级后，以同一 Client 对六个首期组合执行真实官方 API 冒烟，结果写入支持矩阵；日常 `go test ./...` 不受影响（spec Testing Decisions §2、§5）。

**Blocked by:** 19, 20, 21, 22

**Status:** ready-for-agent

- [ ] live 测试需独立 build tag（或显式命令入口）加 `BARNESS_AI_LIVE=1` 双开关；真实 key 仅由 CI secret store 注入到对应 Provider×API 的测试进程，不读开发者默认账号或进程环境回退
- [ ] 每个组合执行：短文本 Stream 与 Complete、模型工具调用 → 测试宿主结果 → 下一次生成、首帧后取消，以及受支持模型的推理/签名历史与图片
- [ ] 断言结构、关联、终态、非空有效内容与 usage，不断言随机文本全文；限制 token 与调用次数
- [ ] 结果仅为 PASS / FAIL / NOT_RUN / UNSUPPORTED：缺 key 为 NOT_RUN，厂商故障为 FAIL/环境故障并在既定预算内重试确认，不改成 skip
- [ ] 每次记录模型、SDK 版本、厂商 request ID、耗时、错误类别；输出先脱敏
- [ ] 支持矩阵保存 Provider、协议、模型、SDK、测试账户/区域别名、能力与最后通过时间；共享 adapter 不连带标记通过

## Comments

**2026-10-02 — from issue 21.** P05's offline fixtures infer from OpenAI's Responses schema what DeepSeek's guide leaves out; the DeepSeek Responses smoke should confirm: reasoning.effort `high` and `max` (from pi's Chat level map; only `none`/`low` were sent live), the reasoning item (`content[{type: reasoning_text}]`) and function call item id shapes, the error body shape and whether a vendor request id header exists (ADR-0014 Consequences). Maintainer decision 2026-10-02 (ADR-0014 决策三): a level the service refuses becomes null in `builtinDeepSeekResponsesModels` under a new catalog version, and the P05 fixtures are corrected to the observed shapes.
