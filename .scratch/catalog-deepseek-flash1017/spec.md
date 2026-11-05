# DeepSeek V4.1 Flash 1017 目录采用

用户要求使用 Catalog CLI 增加 DeepSeek V4.1 Flash 1017。
官方 2026-10-17 更新明确 API ID 仍为 deepseek-flash，现有目录已含 Responses 与 Chat 两条身份。
按 Responses 优先原则依次 revise 两条身份：仅 name 改为 DeepSeek V4.1 Flash 1017，并补充官方来源/字段决定。
不创建虚构的日期 API ID，不新增第三方 Provider，不恢复已退休别名，不扩展协议/能力。
文本/图片、1M context、384k output、已支持 low/high/max 和峰值价格均保持。
保留原有保守峰值估算：input/output/cacheRead/cacheWrite 为 0.3/1.2/0.006/0 USD/百万 token。
低谷价为一半，当前不自动按日历换价；原 Chat/Responses 的已实现能力子集保持。
总目录仍为 81 条（75 chat、3 image、3 classifier），DeepSeek 仍为 4 条；两次 revise 分别推进 Provider 内容版本。

失败场景：名称未更新；伪造 wire ID；重复增加身份；除两处 name 外的定义/价格/顺序变化；请求发送错误 ID；协议字段泄漏；图像/工具/授权回归；旧配置静默改绑；审阅库存与内置字节不一致。
先建立公开 Client E2E 并保留 RED，再按 CLI revise/check/apply/export 采用，更新三处编译数据。
历史采用检查忽略可演进的展示名称，仍核对其余定义和顺序；新专项独立核对两处新名称，数据审计严格限定本批只改变这两处 name。
ChineseEvaluation 新建 v10，仅改配置文档版本和目录 pin；旧 v1–v9 原字节保持。
SDK 核心/依赖/schema 不变；旧资格与授权保持，不把滚动别名的旧测量改为 1017 实测资格。
交付不可变版本化 JSON、完整 review、changelog、代码与分发判定、实际验证和原件索引。
