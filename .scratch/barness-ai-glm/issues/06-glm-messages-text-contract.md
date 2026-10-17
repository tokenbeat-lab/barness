# 06: 建立 GLM Messages 实际契约并完成文本调用

Status: claimed
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 使用自己的 GLM Messages Binding，通过公共 Client 收到两个型号的文本结果；请求仅携带本协议已证实的字段，GLM-5.3 不被默认关闭思考。

Blocked by: 02

依据：GLM 接入 spec R02、R04、R05、R06、R07、R13、R14；验收组 A01、A02、A06；User Stories 4、5、6、7、8、17、18、21、31、40、51、52、53、56。

- [ ] 先写官方基础示例所支持的公共文本 E2E，再实现 GLM Provider 配置；文档缺失字段通过有界候选探针取证，不能复制 Chat/Responses 的 wire 假设。
- [ ] 两型号候选条目与 Messages Binding 独立，最终官方资源地址和 x-api-key 正确；使用现有显式 SDK service、SSE 读取、资源控制和关闭的 SDK 重试。
- [ ] 核对 SDK 默认 anthropic-version、header 与 GLM 基本事件形状；未证实前不自动发 cache_control/TTL、eager、display、interleaved beta、adaptive/block_binding、managed effort/fallback。
- [ ] Stream、Complete、两种 simple、nil/full 与合法 WithHooks 均可通过同一公共 Client 演示；Messages 不消费 SamplingParams/分档 map。
- [ ] 省略思考使用 GLM 自身已证实的默认形状；始终思考型号不能由 simple 默认产生 false/disabled；未接入的显式思考选项按实际支持范围拒绝。
- [ ] 基本 message_start/content/message_delta/message_stop、正文、stop reason 和已报告 usage 形成可重放 wire fixture；缺终态与流内 error 不判成功。
- [ ] 候选探针先登记变量、预期、标准 Key、型号和预算；一次只改一个变量，未有 Key 记录 NOT_RUN，不改用其他身份或扩大额度。
- [ ] 保存现有 SDK 解析与 GLM 实际字段的兼容/差异结论，记录未知项；本票只交付基本文本路径，不把它标为完整 Messages 支持。
- [ ] 产出两型号公共 Client 文本证据和用于 07 的已确认 thinking/工具/usage 协议形状，原生 Anthropic 行为回归保持。

**独立验收：** 两型号经公共四种入口完成 Messages 文本调用；实际请求头和基础 SSE 可重放，并能证明 GLM 未被附加未经证实的 Claude 字段。

**依赖理由：** 仅复用 02 的 Provider 身份和候选 Host 装配；不依赖 Chat 工具或 Responses 实现。

