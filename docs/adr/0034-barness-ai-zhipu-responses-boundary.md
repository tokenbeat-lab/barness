---
status: accepted
date: 2026-10-17
---

# GLM Responses 在原始 JSON 边界转换解析副本，单独保留原生 reasoning

GLM 官方 Responses schema 允许 reasoning.content 为对象、message.content 为对象或数组；锁定的 Go SDK 可以无错误解码对象却丢掉正文。因此沿用现有 Responses adapter 的生命周期，在 SSE JSON 进入 SDK 前按结构校验并把已知内容转换为解析用数组，同时按 output_index 保存未经转换的 reasoning 条目供 Native State 使用。解析副本不补造 summary、encrypted_content 或 OpenAI ID；原生签名保留厂商的对象或数组形状。直接认定 Unmarshal 成功意味着内容完整、字符串替换或在 transport 改写响应都会掩盖边界问题，故不采用。后续历史回放须在请求边界按 GLM InputReasoning 转换，不能将解析副本当作原生状态；工单05负责这一消费端。

权威 output_item.done 结束每个内容块；GLM 独立处理 text.done/reasoning_text.done 和仅终态输出，终态不再重复已结束的块。合法 completed/incomplete 不依赖 [DONE]；缺终态 EOF、畸形结构及与事件不符的状态明确失败。保持 OpenAI/DeepSeek 原有解析行为和 pi 契约。

Provider 协议能力只自动发送 store=false 和隔离后的缓存路由键；不带 summary、encrypted include、service tier、缓存保留期或亲和头。GLM max_output_tokens 按官方 schema 限于1..min(型号上限,131072)，不沿用 OpenAI 的16下限。所有采样层的有效预算保留原始 JSON 数字至最终边界，以共享精确整数校验拒绝浮点舍入后看似合法的分数；回调后再次校验。型号级完整档位在授权后、读取凭据前拒绝，simple 按独立候选目录夹取。

5.3 的 Responses generic effort schema 与型号页有冲突，首期按 spec 限制 low/high/max，省略 effort 则保留服务默认；wire low 不等于已测得的 Chat low 计算强度。独立 Responses 候选目录的1,000,000上下文是保守边界，131072是本协议 schema 精确输出整数；价格为明确标记的合成测试值。两个真实基础样本均为数组；对象失败的红绿证据来自官方 schema 的合成公共 Client E2E，不能称为真实对象样本或正式支持。

证据与重放见[工单04](../../.scratch/barness-ai-glm/evidence/ticket-04/README.md)；正式目录、价格与独立真实验收由后续票提供。
