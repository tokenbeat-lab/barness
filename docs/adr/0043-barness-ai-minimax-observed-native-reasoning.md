---
status: accepted
date: 2026-10-18
---

# MiniMax 实际推理镜像与原生续接投影

国际站 M3 在源码 `38b7098` 的有界研究中，Chat 首帧同时返回 `reasoning_content` 和 `reasoning_details` 的 `reasoning.text`，两份文字逐字相等；既有“同时出现即冲突”校验因此在读取后续帧和用量之前失败。仅集成同一 delta 中可验证的精确文本镜像，显示一次思考，保留 details 的条目身份、格式和文本；真正不同的值、未知条目和跨帧切换独立表示仍失败。实际捕获只有首帧，离线补齐帧明确标为合成，不能成为完整真实流或零计费证明。

镜像来源写入现有 `Thinking.Signature` 的 MiniMax 专用不透明封装，内部保存原 details 数组；持久化并经宿主担保后，同一精确目标回放完整 details 和等值 content。仅有 details 的原数组签名仍仅回放 details。这样可在恢复后区别两种来源，不依赖显示文字推断，不扩展公共 DTO，也不把内部封装发送给厂商。原生封套的租户、账户、Provider/API/型号校验与降级规则不变，最终 payload 继续核对两字段等值。

Responses 的完整实际工具流还返回非空字符串 `encrypted_content`，完整原条目已保存在思考签名，但旧转换因未知输出键而在续接前拒绝。该字段只作为已观测的字符串输出元数据校验并保留；文档规定的输入没有该字段，仍将全部可见 `reasoning_text` 无损投影为 `summary_text`，不生成摘要、不转发密文、不放宽输入 hook。非字符串密文输出继续失败。本决策不声明加密续接能力；真实完整候选中的自身工具续接仍必须证明这一文档输入投影可用，否则身份验收保持失败。

本决策补充 ADR-0039/0042 的实际 wire 边界，不改变其他 Provider。原研究的 FAIL、NOT_RUN、已知用量、原始密文、来源 pins、账本和 intent 均原样保留，研究不重跑，也不改写为候选或正式 PASS。[来源与失败优先测试](../../.scratch/barness-ai-minimax/evidence/ticket-10/international-convergence/README.md)保留逐字首帧、完整 Responses 工具流及独立合成补齐的归属。
