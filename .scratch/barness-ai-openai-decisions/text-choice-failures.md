# Issue 02：先列失败，再实施

测试边界沿用已批准 acceptance §1：公共 Classify / HookedClient.Classify、Catalog/Lookup/List/Hash/NewClient；HTTP Provider 和 Host resolver 是系统边界。起点 `a8a4686d7b47a48d4514ac1c459f54b95ff2f8e0`，当前分支 main，保留已有未提交规划及 GLOSSARY 改动。

- OD02/03：同名 chat 无 classifier 身份，错误 operation/provider/API、白名单/目录不交集、外国租户、disabled；凭据前拒绝。
- OD04：nil/value/pointer 默认，两个协议错 Options，typed nil 不 panic；无 Attempt。
- OD05/06：旧省略态；未知/重复/缺 text 的 Input、未知/重复格式、负/倒置 MinChoices、已知协议能力错误；Hash/NewClient 一致，所有新 slices 与 tiers 在入口/查询出口独立。
- OD09/10/12：坏 State、无题、空名称、nil question、256 choice/11 score；单候选/结构化说明/未实现 score、bool 在 capability 拒绝且无凭据。
- OD11/14/16：字符串原文，空 string/object/array，compact 精确数字；数字键先数值排序、其他键按冻结 JSON 键序；字符串 true/false，null/空说明省略；wire 仅 model/input/questions。
- OD22/23/25：合法 Keep/Replace 换名/候选/顺序；旧题答案拒绝；入口/Hook 引用在冻结后的同步修改不影响发送和解析；改 model、认证、目标、工具、授权字段及外部引用为 tenant_denied；未知字段、非 string input/message/image、未实现 kind、非 string 候选为 callback_failed。
- 资源：入口已有 scope limits 保持；初始与最终 native 完整题目（含 name）和 JSON 限额、问题数、输入字节超限；零发送，所有资源释放。
- OD27–32：按 name 回配、缺/多/重复/null name、重复/case alias 字段、错误 discriminator、refusal 携普通语义字段；候选缺/多/重复/非字符串、非法 choice、非最高/合法并列、概率与 confidence 缺/null/类型/NaN/overflow/负/大于1、概率和 1e-6 内外；整体发布或整体清空。
- 用量与生命周期：合法已知 input/output 与 input-only cost，缺用量不伪造 complete，合法 usage 在坏答案时保留；2xx 后不重放，header→payload→response 一次，固定凭据，body/permit 释放，Observer 仅元数据。
- Jev：不删/不收窄固定基线、原目录 Hash/省略态/构造范围、精度/舍入/错误顺序/最终换题回归。

图片公共契约、mixed score/bool、完整 usage presence/pricing、live/正式内置目录按后续工单交付；本票真实 HTTP 闭环使用 loopback 合成宿主目录，不代表厂商 live 支持。
