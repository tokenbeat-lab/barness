# Claude Haiku 5.5 目录接入

2026-11-04 用户要求使用 Catalog CLI 增加 Claude Haiku 5.5。
范围仅 chat / anthropic / anthropic-messages / claude-haiku-5-5。
只新增这一固定 ID，不恢复旧 Haiku 4.5，不更新其他模型的价格或正式验证目标。

官方规格：文本/图片输入，1,000,000 context、128,000 output，adaptive thinking。
5 个 effort 等级：low/medium/high/xhigh/max；服务默认 medium。
禁发 temperature；已有托管 effort 与 drop_block 路径遵守前缀约束。
本库选择始终 adaptive 的已实现子集，支持逐轮 effort；不声明 PDF、浏览器/电脑原生工具、Batch 或关闭推理能力。
基本 USD/百万 token：input 0.1、output 0.5、cacheRead 0.01、cacheWrite 0.125。
提示词（普通输入+缓存读+缓存写）严格超过 100,000 token 时整请求费率为 0.5/2.5/0.05/0.625；1h 写按该请求 input 的 2 倍计费。

保存 models.dev 与官方原始来源及实际抓取时间。models.dev 的 PDF 模态当前无映射，不修改来源来绕过校验；使用 CLI 官方补充入口及完整人工定义/决定。
先公开发现测试 RED，再 add 预览/确认 → check → apply 预览/确认 → export 完整 review。
同步 Provider 源、原字节内置快照、完整 baseline，交付版本化 JSON、changelog、代码/分发判定及摘要索引。

历史下架测试只验证其冻结快照内的身份，允许后续独立接入；新增模型的公共 E2E 验证原 79 条定义保持、全部能力与跨缓存阈值计价。
旧评估配置不改写，新建 v8 仅更新文档版本和目录 pin。
最终目录 80 条（74 chat、3 image、3 classifier），Anthropic 8 条、内容版本 11。
不执行厂商请求、软件发布或部署，不宣称新厂商资格。
