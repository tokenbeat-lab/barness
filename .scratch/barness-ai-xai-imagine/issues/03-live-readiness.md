# 原生验证准备与候选支持边界

Type: task
Status: resolved
Blocked by: 01, 02

## Scope

补充本路线的有界 live 场景与捕获/资格准备；先受控进程失败后实现；不发放授权、不请求厂商、不晋升正式目录。明确精确场景和预算提案。

完整契约：[spec](../spec.md)。验证遵循 [SOP](../../../docs/agents/verification.md)。

## Comments

2026-10-29：implement-spec 根据现有规格建立工单依赖图。

## Answer

候选验证准备完成；实际真实调用和正式资格不属于此完成声明。接线提交 `4a0c8fc3`、`f1205580`、`6ab4cbc8`；普通受测提交 `2a6d1aeb65e1bc12b7c04416a1bec3bf118f951b`，包含集成分支及协议 `e2d81cd5`。

- 精确入口 `^TestXAIImagesCandidate$`，`-test.count=1`，live 固定 `-test.timeout=1200s`。先于旧 runner 选择独立模式，使用既有 Client/Host 装配和 unary 生命周期。
- 六个固定场景见 [live-plan](../live-plan.md)。最大 6 调用/6 HTTP、8 输入/7 请求输出；零重试。目录预留总额 USD0.53、固定停止阈值 USD0.60；每调用180s，进程1200s，120s留证据。
- `supportmatrix.ProcessScope("xai-images", mode, "candidate", binary)` 复用既有 `loadProcessAuthorization`、livegrant 严格解码/Check/Claim/权威共同Git存储。scope 绑定最终源码、二进制、候选目录版本/hash、价格JSON及实际WebP输入字节。过期/错scope/重用claim在发送前拒绝。
- recorder 只增加受信任的有限 `responseLimit` 配置。本路线32MiB捕获保留完整JSON、图片及精确整数ticks，不调用Google的float64压缩器。请求/响应、结果、图片、Observer、原授权/claim、源库存和实际退出均留在原生包。
- 候选始终 `formalQualified=false`，formal scope 拒绝；正式消费者拒绝controlled候选。没有Builtin晋升，没有复制资格/发布控制器。

### 验证与保留

公共接缝为隔离二进制的真实进程退出与报告。先失败后实现，所有环境从空白构造，无真实凭据、零厂商HTTP。

命令：`python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e -run '^TestXAIImagesLiveProcess$' -count=1 -timeout=6m`。

- 原始缺入口/报告红证据：`.evidence/barness-ai-xai-imagine/e2e/20261029T000611.164633Z-73vxuynf`，FAIL/exit1。
- 合入尚未完成协议的编译失败：`20261029T001103.808792Z-o6o5d3q_`，FAIL/exit1；原错误为 Usage/UsageReporting 返回赋值顺序，随后合入协议修正。
- 首次完整绿证据：`20261029T001233.979472Z-1mdx6c6v`，PASS/exit0。
- 最终定向绿证据：`20261029T001329.923027Z-y_mixpyo/receipt.json`，PASS/exit0；19个测试事件，墙钟10.952s。receipt SHA256 `43b2eed7c0279616a58dafc9ada8fe2a129c277ba568f0b1fa0ee2b02e982c9a`。唯一主包 `bundles/20261029T001331.166152000Z/manifest.json` SHA256 `05e1ec69c61179fd01966487d9846601b428c3204868a1ba5aafc3fc1fd42277`；audit SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- 覆盖15组无凭据/无grant/身份/目录/来源/二进制/预算/expiry/alias/入口/混合凭据拒绝；六场景实际公共图片调用；大于512KiB的完整raw和尾部ticks；USD0.53独立费用/预留；一次性claim拒绝重用；取消；真实EISDIR报告写失败；正式CLI拒绝controlled候选。
- 相邻与全受影响普通包由工单04在最终集成来源统一执行。此工单定向通过不冒充最终冻结普通回执。
- live/race/压力均NOT_RUN；无新增付费授权、无真实支持资格，release.pass=false。未来执行必须使用新的本路线授权，沿用已有流程；此工单没有签发授权。
