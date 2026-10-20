# GLM Messages glm-5.3/cancel 单用例复测

Type: task
Status: resolved
Blocked by: None

## Question

2026-10-19 用户要求：重新进行 `glm-5.3/cancel` 在 Messages 路线的测试，通过后及时更新上下文。

## Scope

本次新指令单独授权这一实际用例，最多 1 logical call / 1 HTTP Attempt，输出上限 4096 token，调用期限 120 秒，无重试、无自动下一轮。中国区 `open.bigmodel.cn` 直连，不继承代理，不跨区回退。复用原固定 prompt、Messages options、公共 Client 与收到首个 TextDelta/ThinkingDelta 后主动取消的断言。

保留票15原 120001ms 超时失败；原固定六项及其 647/647 消费仍然关闭。新测试独立记账，不复用旧 intent 或未消费额度。单用例 PASS 只更新该失败场景的最新观察，不拼接旧 22 场景报告，不授予 Messages 全路线、默认矩阵或完整发布认证。

## Acceptance

- 独立、固定单用例入口；受控失败先行验证实际只运行此用例并限制 1/1；原正式入口不能任意缩小验收集合。
- 执行前绑定已提交来源、实际二进制、明确计划与受控检查；只在该实际执行子进程中读取 `.env` 所选 `GLM_KEY`，不得输出凭据。
- 保存原始请求、事件、断言、用量状态、命令退出、来源与完整性信息；不改变原失败证据和 selector。
- 立即更新人类与机器交付上下文，明确单场景结果及剩余完整验收缺口。

## Comments

2026-10-19：继续用户已调用的 implement-spec 工作流。实现者在独立 worktree 中增加最小诊断入口；本票之外的用户目录不参与变更。

## Answer

2026-10-19 单用例真实复测PASS，来源6a24ba3；861ms，收到非空ThinkingDelta后取消，`canceled/stream`、HTTP200、StopReason=`aborted`。中国区直连，1call/1Attempt、零重试，usage partial未知部分不作免费。新1/1授权已关闭，续期累计648/648；原失败及wire request、selector、矩阵均核对保留。单用例结果不替代完整formal22/发布认证。

[本次证据](../evidence/ticket-18/README.md)、[机器状态](../evidence/ticket-09/international-delivery-status.json)。
