---
status: accepted
date: 2026-10-29
---

# xAI Imagine：保留 image 领域边界，独立转换原生图片契约

用户于 2026-10-29 确认整体设计：Imagine 经 barness-ai 的 image 路径接入，复用 GenerateImages、ImagesRequest/ImagesResult、ImagePolicy 和统一 unary 生命周期，以独立 xai-images 协议及 XAIImagesOptions 处理厂商请求与结果。accepted 表示本次设计共识成立，不表示代码已实现、真实资格或发布已通过，也不自动开始实施。完整范围见[设计方案](../../.scratch/barness-ai-xai-imagine/design.md)，事实见[官方研究](../../.scratch/barness-ai-xai-imagine/research.md)。

两家虽然使用相同 images/generations、images/edits 路径，xAI 编辑的 JSON image/images 资源对象、response_format、比例/分辨率/质量和收费语义均不同。将其塞入现有 OpenAIImagesOptions 会把原生 OpenAI 契约变成含厂商分支的字段集合；新协议承担这部分变化，但不复制认证、凭据、重试、准入、取消或 HTTP 客户端。该建议延续 ADR-0020 的操作授权和 ADR-0022 的有界同步内联图片边界，补充 ADR-0054 之外的 xAI 图片路线，不使聊天 Binding 获得图片权限。

用户已确认本次包含按张估算、选项省略沿厂商默认、报告收费使用独立 USD 浮点金额。ImagePricing 增加与 token 费率互斥的按张分支，保持旧字段单位及历史快照含义；目录估算继续用 Attempt.CostEstimate。xAI 报告的 cost_in_usd_ticks 按官方定义属于本请求已收费事实，以独立 Attempt.ReportedCostUSD 保存，原捕获保留 ticks；token 缺失不能补零或伪装完整。Q1–Q7 均已确认，已确认术语记录在根 GLOSSARY；代码和既有运行时合同未修改。

首期只针对全球 grok-imagine-image-2.0。候选设计上限为生成10张，编辑5张有序参考/输出1张，分辨率1k/1.5k/2k，所有上限受型号与宿主策略进一步收紧；正式支持范围须以自身证据为准。质量省略沿厂商默认，auto 缺乏实际质量依据时只给有限区间或未知估算。单图编辑暂拒绝显式比例，以保守边界处理官方资料差异；放宽须另有该路径证据并更新合同。

替代方案包括：复用 OpenAI Images 全部 DTO/默认（请求和计价事实不成立）；把图片包装成聊天工具输出（改变用户指定的操作与授权边界）；另建 Imagine Client 或异步作业系统（当前同步图片用例不需要第二套生命周期）；把图片张数换算成 token（破坏已公开的用量单位）。

官方依据：[编辑指南](https://docs.x.ai/developers/model-capabilities/images/editing)、[Images REST](https://docs.x.ai/developers/rest-api-reference/inference/images)、[定价](https://docs.x.ai/developers/pricing)、[费用跟踪](https://docs.x.ai/developers/cost-tracking)。查阅日期：2026-10-29。

## 2026-10-29：按自身实测收紧内置目录

用户在真实候选六场景完成后要求“激活xAI目录”。内置目录独立声明 grok-imagine-image-2.0，生成最多2张、编辑输出1张、最多5参考图，分辨率1k/1.5k/2k，质量low/medium/auto，显式比例只开放1:1/16:9。这个范围取自身真实捕获与协议约束的交集，不继承候选10输出与16比例；单图编辑仍拒绝显式比例。原始候选目录、测量stage/mode、授权与正式资格事实保持。

内置价格保留同一带日期按张快照，生产目录不依赖testkit。候选目录保留独立字节，供原测量追溯。默认Client离线回放六份原始捕获以核对最终wire、图片与收费。目录版本变化通过三个新v2配置显式传入中文评估示例；旧v1及严格目录核对保持，不自动升级历史配置。源码激活、正式原生资格和完整发布分别判断。详情及唯一当前状态见[激活范围](../../.scratch/barness-ai-xai-imagine/catalog-activation.md)和[规格](../../.scratch/barness-ai-xai-imagine/spec.md#唯一当前状态)。
