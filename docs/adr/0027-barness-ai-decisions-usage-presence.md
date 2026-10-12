---
status: accepted
date: 2026-10-12
---

# Decisions 六项 presence 与 inclusive input-only 计费

工单 04 落实 ADR-0026 的独立 usage 转换边界。ADR-0010 的公共 Usage、Attempt Reporting 和
Price Snapshot 继续复用，Jev 与聊天转换保持原义。Decisions 不使用聊天的缓存拆桶。

完整报告必须同时明确 input/output/total、cached/cache_write、reasoning 六项非负 int64 整数。
缺失/null 的计数保持 partial，Reasoning 只在明确值存在时设 Nullable Value；缺失/null usage
为 unreported。错误类型、重复键、大小写别名、溢出或矛盾使整份 usage 无效，而不发布部分前缀。
缓存与已知 input、reasoning 与已知 output、全报 total 之间的一致性在协议边界验证。

Input 保留 inclusive input_tokens；canonical cache 桶为零。合法 total 保留，缺失时 checked sum
已知 Input+Output。完整有效 usage 先于答案登记，答案失败仍可由 Host 核对消耗与输入估算。
原始 presence/cache 细节不成为新的 Observer 契约，测试 fixture/capture 已保留它们。

Estimated Cost 按授权 Classifier Model 的目录快照计算，只提取输入费用，阶梯应用于全请求输入。
复用 ModelCost 的算术与 Host 自带费率，不硬编码正式型号、修改内置目录或借用同名聊天价格。
合成目录使用 .1/.2 USD/百万、272000 严格大于阈值；ResponseModel 不改变授权身份或价格。

代价是缺细节的合法响应只能标 partial，不能据零估算声称免费。全球标准处理估算未核对区域 premium
或真实账单；合成非零缓存不是厂商证据。候选和正式目录的费率来源、真实核验留给工单 09/10。
回滚撤销 Decisions Binding/目录授权；已发请求的合法用量保持可核对，不自动改走 Jev 或聊天。
