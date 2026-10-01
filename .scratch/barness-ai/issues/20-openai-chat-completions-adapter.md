# 20: OpenAI × Chat Completions（P04）

**What to build:** 租户可通过 OpenAI Chat Completions binding 完成文本、工具往返与模型支持的推理/图片，以协议终态判定成功，通过全部适用横向场景（spec I4、P04）。

**Blocked by:** 18

**Status:** ready-for-agent

- [ ] 使用 OpenAI Go SDK 的 Chat 能力，关闭默认重试；核查 WithJSONSet / ExtraFields / ExtraBody / 受控 RoundTripper 的实际可用性，显式配置 pi 默认字段并以实际请求验证
- [ ] 工具 JSON 分片、finish_reason 映射、reasoning 扩展字段
- [ ] 流末 usage chunk：结束信号前不遗漏 usage；缺结束信号为错误
- [ ] samplingParams 走 OpenAI-compatible 覆盖语义
- [ ] 图片支持与占位降级分别覆盖
- [ ] 复用 E01–E09、E11 适用用例全部通过；pi 差分无待处理差异
