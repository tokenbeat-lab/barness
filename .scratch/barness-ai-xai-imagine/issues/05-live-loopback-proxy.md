# 图片候选验证的显式本地代理

Type: task
Status: claimed
Blocked by: none

## Scope

首次真实候选1次transport调用在ConnectTimeout15s失败，余5场景未执行，原授权不能重跑。系统本地代理启用，图片harness却拒绝BARNESS_AI_LIVE_PROXY。复用现有smokeProxy的严格loopback HTTP校验，显式配置native transport并记录选择；不改生产Client默认、不自动读取环境/系统代理、不扩大目标/预算。

## Verification

先公共子进程红，后最小harness修复；以受控loopback代理拒绝CONNECT证明native传输接线，无厂商请求。普通fixture仍完成6场景，非法proxy在claim/HTTP前拒绝。按SOP仅验证本次harness/过程变化，旧完整普通产品证据保留，不自动发起付费重跑/race/压力。

## Comments

2026-10-29：用户回复1批准的首次授权与失败原件见spec唯一当前状态。本工单仅准备后续未授权的新冻结二进制。
