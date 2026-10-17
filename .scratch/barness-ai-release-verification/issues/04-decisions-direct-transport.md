# 04：Decisions 正式真实入口拒绝未配置代理的直连

Status: claimed

## Failure modes before implementation

- 未设置可选 BARNESS_AI_LIVE_PROXY 时，被当作非法代理，在任何 HTTP 前拒绝正式和候选入口。
- 只修 allow_unknown 分支，导致带 Host 权限记录的直连仍失败。
- 空代理回退到 ambient HTTP(S)_PROXY，破坏显式 Host 传输边界。
- 修复误放行非 loopback、坏端口、带身份或非 HTTP 的显式代理。
- 受控 PASS 冒充真实支持；修复源码后继续使用旧当前源码证据；覆盖首轮真实失败或低报预算。

## Acceptance / seam

- 使用已经验收的独立公开进程 E2E seam，先扩展 candidate 与 formal process 的直连场景并记录 RED，后修实现。
- 候选带 Host 记录、候选 allow_unknown、正式 allow_unknown 均可无显式代理完成同一有界受控流水线。
- 显式 loopback 及已有非法代理拒绝回归保持通过，账户权限 unknown 不变为 verified。
- 无凭据 live harness 完整回归、实际 Decisions 重采集与最终完整 normal/race 门禁通过后才关闭。
- Go 内容变化后按本轮授权唯一恢复轮最多302 Calls/Attempts 重采集11支持及4效果；初轮失败原包保留。

## Diagnosis

2026-10-17 首轮 `.evidence/ticket14-remaining-20261017073529Z`：
`openai-decisions` CLI退出1，9场景 config FAIL，Calls/Attempts均0，audit零发现。
`configureDecisionsProxy` 无条件调用只接受非空显式 loopback 的 `smokeProxy`。
其他两个调用方先判断非空；已有候选/正式进程夹具总设置合法 loopback，掩盖了直连分支。
这不是服务端权限拒绝，不需要新增厂商授权；固定 endpoint、TLS 和型号不变。
