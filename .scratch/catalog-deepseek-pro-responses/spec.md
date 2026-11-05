# DeepSeek V4 Pro 优先接入 Responses

2026-11-04 用户纠正 ADR-0014 line 36：Pro 也必须支持 Responses，且优先支持。
同日澄清：凡厂商为模型提供 Responses 协议，barness 都优先适配、接入和验证其 Responses 身份，并优先推荐 Responses 调用入口；原则适用于所有 Provider 和型号。Pro 是本次具体补齐项。
撤销 Pro 被排除的错误结论；优先补齐 chat/deepseek/openai-responses/deepseek-v4-pro，并以 Responses 作为 Pro 的推荐调用入口。
Host 仍显式选择 Binding.API，不新增自动协议路由、隐式 fallback 或 SDK 型号分支。
原 Pro Chat、Flash 双接口、其他模型与旧测量保留；不修改正式 live 契约的 Flash 资格目标。

官方原生身份：deepseek-v4-pro，文本输入、1,000,000 context、384,000 output。
Responses 采用当前官方 effort 映射：off→none，minimal/low→low，medium/high/xhigh→high，max→max；服务默认 high 作为 metadata。
选择已有无状态 Responses 子集：文本、函数工具、推理与完整历史；不声明本轮未接入的 web_search/custom apply_patch。
价格冻结峰值 USD/百万 token：input 1.32、output 3.96、cacheRead 0.044、cacheWrite 0。
低谷价为峰值一半，当前 ModelCost 不表达时段/节假日；上述为保守峰值估算，不声称等同任意时刻账单。

公开 Client E2E 先 RED 后 GREEN；再 CLI add 预览/确认 → check → apply 预览/确认 → export 完整 review → 三处同步。
原 80 条定义和顺序保持；总目录 81（75 chat、3 image、3 classifier），DeepSeek 4 条，内容版本 1→2。
修订 ADR-0014 与对应历史工单注解，保留原文快照和旧厂商测量，不把文档纠正冒充实测。
原 Haiku 接入 E2E 只核对其原始身份投影，允许后续独立添加。
ChineseEvaluation 新建 v9，仅更新配置版本与目录 pin，v1–v8 原字节保留。
交付版本化 JSON、changelog、代码/分发结论、原件摘要和完成索引。
