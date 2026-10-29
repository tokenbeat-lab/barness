# 图片候选验证的显式本地代理

Type: task
Status: resolved
Blocked by: none

## Scope

首次真实候选1次transport调用在ConnectTimeout15s失败，余5场景未执行，原授权不能重跑。系统本地代理启用，图片harness却拒绝BARNESS_AI_LIVE_PROXY。复用现有smokeProxy的严格loopback HTTP校验，显式配置native transport并记录选择；不改生产Client默认、不自动读取环境/系统代理、不扩大目标/预算。

## Verification

先公共子进程红，后最小harness修复；以受控loopback代理拒绝CONNECT证明native传输接线，无厂商请求。普通fixture仍完成6场景，非法proxy在claim/HTTP前拒绝。按SOP仅验证本次harness/过程变化，旧完整普通产品证据保留，不自动发起付费重跑/race/压力。

## Comments

2026-10-29：用户回复1批准的首次授权与失败原件见spec唯一当前状态。本工单仅准备后续未授权的新冻结二进制。

## Answer

代码`e10db0f0`、最终分支`fb1dbb85`、集成`5371abdb`。仅修改3个harness/过程测试文件，复用smokeProxy显式设置loopback HTTP代理，并在报告保留选择；生产Client、协议、型号、六场景、预算与零重试均不变。受控proxy模式要求显式合法代理，收到CONNECT响应即终止后续TLS/API；200/502均无认证头和厂商HTTP，缺失/非法proxy在claim/HTTP前拒绝。

公共子进程红：`20261029T010836.302123Z-a08qc8uj`和`20261029T010954.553610Z-vk0roj4j`，FAIL/exit1。最终冻结绿：`20261029T011128.326343Z-huqny4b5`，PASS/exit0，29测试事件。实际argv：`python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e -run '^TestXAIImagesLiveProcess$' -count=1`。根.evidence保留全部原件；receipt SHA256 `9a6c2fd5dfe1071779e2be9cf1771f4c7018ad8ee297e0e7f03a5c30410aa477`。

两个轴独立复核均0行动问题，见[review](../review-proxy.md)。新候选二进制完成构建但未获新付费授权；本工单不运行真实重试，原FAIL和claim保持。唯一当前状态见spec。
