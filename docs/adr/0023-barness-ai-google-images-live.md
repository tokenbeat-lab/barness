---
status: accepted
date: 2026-10-08
---

# Google Interactions 图像：显式省略 delivery，响应强制内联

工单 14 固定 Google × google-interactions × image、v1beta 和裸型号
`gemini-nano-banana-2.1`，通过公共 Client 的本组合独立真实进程验证。
官方参考列出可选 `delivery: inline`，但 2026-10-08 本组合生成与编辑均以
HTTP 400 `Image delivery mode is not supported.` 拒绝该字段。原失败报告保留。
维护者在本次会话明确批准省略 delivery，修订规格后继续真实验证。
这重新打开 ADR-0022 的显式 inline 请求固定项；同步、无状态和禁止外部资源的边界不变。

请求始终省略 delivery，且回调不能加入该字段（包括 inline、uri 和 null）。
固定 store=false、单个 image response_format，省略 background/stream；不进行
版本、型号、协议或失败后的隐式请求回退。只有 completed 且全部 model_output
图片含可校验内联 data、无 URI、无 continuation_token 才成功。内置能力纳入仍
须本组合生成、参考图编辑的真实 PASS；拒绝或缺凭据不能声明支持。

考虑保留显式 inline：其失败阻止首期必须交付的能力；请求省略可选 delivery
符合官方图像指南的原生请求示例。将来恢复显式 delivery 必须另立协议变更、
更新 fixture 并通过该组合真实冒烟，不能因官方枚举仍存在便恢复。

官方来源（抓取 2026-10-08）：[参考](https://ai.google.dev/api/interactions-api)、
[图像指南](https://ai.google.dev/gemini-api/docs/image-generation)。
本组合实测、usage 计价结论及能力目录范围见
[交付证据](../../.scratch/barness-ai-pi-1.0/google-images-live-evidence/README.md)。

## 真实接受、用量与支持范围

生成及一参考图编辑两项真实 PASS，2 调用/2 张/0 重试、JPEG 1024×1024。
两份响应都省略 id 和请求 ID 头，报告 model 与 completed；不能为满足报告字段
而伪造 ID。output_tokens_by_modality 只含 image=1120；文本思考另报，因此该
图片明细不含思考。Input 取总输入，Output=原生总输出+thought，Reasoning 是
Output 内分项，TotalTokens 保留厂商值。计价保留原始模态明细并将 thought 以
文本输出费率计一次；未报文本输出及未报输入 image 不填 0，不累加内部模型
调用明细，两项为 partial。估算仅已知部分，不是完整账单；complete/unreported
仍按既有契约，离线实际形状回放覆盖这三个轴。

官方标准费率（美元/百万 token）为 text/image 输入 1.50，text/thought 输出
7.50，image 输出 30。该无状态路线没有文档化适用缓存价格；不猜测，也不叠加
按张费用。价格页与图像指南对 4K token 数矛盾，保存事实并仅验证 1K。
首批内置能力只声明一参考/一输出、1K、1:1，目录为 2026-10-08.5。
