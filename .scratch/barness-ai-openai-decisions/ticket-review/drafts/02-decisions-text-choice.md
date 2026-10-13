# 02: 打通 Decisions 文本单选与最终请求闭环

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 01: 固定 Jev 公共行为与可重放回归基线
**Source tickets:** 02, 03
**Requirements:** R01, R02, R03, R04, R05, R08, R09, R10, R11, R13, R14, R19
**Acceptance cases:** OD01, OD02, OD03, OD04, OD05, OD06, OD09, OD10, OD11, OD12, OD14, OD16, OD22, OD23, OD25, OD27, OD28, OD29, OD30, OD31, OD32
**Updated:** 2026-10-12

**What to build:** Host 通过公共 Classify / HookedClient.Classify 和自带文本目录完成 Decisions 单选，得到严格普通答案或类型化拒答；合法 Hook 换题按最终请求作答，越权在发送前拒绝。

## Responsibility

负责首个可用的协议路径、最少必需公共契约和文本能力发现：API、空 Options、RefusalAnswer、Input/MinChoices/DescriptionFormats、目录校验/复制/Hash、native 编码/guard/冻结及单选结果。图片契约随图片切片交付。保留 TypeSafe 独立转换，复用共享 unary，直接替换单协议分派，不新增 SDK、registry、运行时或兼容 shim。协作时保留他人改动。

## Failure scenarios first

- 同名聊天型号、外国租户或错误 Binding 取得分类权限；未授权时读了凭据。
- 新能力字段省略使旧 Jev 目录无效，或返回的目录 slices 污染下一次调用。
- 单候选、结构化说明或错 Options 被送出；typed nil 导致 panic。
- State 多加引号、经 float64 丢精度，数字键及 map 枚举顺序漂移。
- Hook 换 model、认证、目标或新增未支持的 native kind；换题后仍按旧题解析。
- 单选缺题、重复名称/字段/候选、坏概率或错误 discriminator 被部分发布。

## Acceptance criteria

- [ ] 先列全部失败变体、写红色公共 E2E，再增加公共契约与生产路径；先得到一次实际调用闭环。
- [ ] 四维型号身份、Binding 操作/Provider/API/白名单/目录交集在凭据前授权；原 Jev scope 错误顺序和构造范围不变。
- [ ] nil、空值和非 nil pointer Options 使用正确协议默认；typed nil 和错协议为 invalid_request/capability，无 Attempt。
- [ ] 旧省略态保持 text-only、最少 1 候选和宽 JSON；新文本目录显式声明最少 2 候选及 string 说明；新 slices 独立，Hash 与 NewClient 共用校验。
- [ ] State string 变原文，object/array 精确 compact；choice value 保持字符串，null/空说明省略；问题和候选冻结顺序确定，响应以 name 匹配。
- [ ] 无图 native 请求只有授权 model、input string、questions array；复用固定凭据、header → payload → response 和共享 unary，不读取环境 key。
- [ ] 入口和最终 Hook 同时检查问题数、每题完整编码大小及完整 JSON 限额；越权 tenant_denied/request、坏 schema callback_failed/request，零 Attempt。
- [ ] 初始尚未实现的 score/bool/图片能力在凭据前拒绝；Hook 无法启用未实现的题型或 message/image 形状。后续切片直接扩充真实实现。
- [ ] 单选严格检查完整名称/候选集、精确字段、finite/range、1e-6 概率和与最高概率并列；单题 refusal 只输出 type=refusal、stop/nil error。
- [ ] 基础已知 usage 和输入成本真实保留；不伪造 complete，不用 placeholder parser。任何损坏答案整体清空，2xx 后不重放，body/permit 释放。
- [ ] Jev 基线持续通过，公共请求/Hook 的可变引用在冻结后修改不影响发送与结果；完整 artifact 包含 SHA-256 manifest、审计及准确重放命令。

## Completion

演示真实文本单选和拒答闭环，尚未扩展的能力保持拒绝。候选/正式目录与 live 结论留给后续切片，不把本票的合成目录记为内置支持。
