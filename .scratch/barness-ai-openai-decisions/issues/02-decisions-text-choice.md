# 02: 打通 Decisions 文本单选与最终请求闭环

**What to build:** Host 通过公共 Classify / HookedClient.Classify 和自带文本目录完成 Decisions 单选，得到严格普通答案或类型化拒答；合法 Hook 换题按最终请求作答，越权在发送前拒绝。

**Blocked by:** 01: 固定 Jev 公共行为与可重放回归基线

**Status:** resolved
**Type:** task
**Source tickets:** 02, 03
**Requirements:** R01, R02, R03, R04, R05, R08, R09, R10, R11, R13, R14, R19
**Acceptance cases:** OD01, OD02, OD03, OD04, OD05, OD06, OD09, OD10, OD11, OD12, OD14, OD16, OD22, OD23, OD25, OD27, OD28, OD29, OD30, OD31, OD32
**Updated:** 2026-10-12

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

- [x] 先列全部失败变体、写红色公共 E2E，再增加公共契约与生产路径；先得到一次实际调用闭环。
- [x] 四维型号身份、Binding 操作/Provider/API/白名单/目录交集在凭据前授权；原 Jev scope 错误顺序和构造范围不变。
- [x] nil、空值和非 nil pointer Options 使用正确协议默认；typed nil 和错协议为 invalid_request/capability，无 Attempt。
- [x] 旧省略态保持 text-only、最少 1 候选和宽 JSON；新文本目录显式声明最少 2 候选及 string 说明；新 slices 独立，Hash 与 NewClient 共用校验。
- [x] State string 变原文，object/array 精确 compact；choice value 保持字符串，null/空说明省略；问题和候选冻结顺序确定，响应以 name 匹配。
- [x] 无图 native 请求只有授权 model、input string、questions array；复用固定凭据、header → payload → response 和共享 unary，不读取环境 key。
- [x] 入口和最终 Hook 同时检查问题数、每题完整编码大小及完整 JSON 限额；越权 tenant_denied/request、坏 schema callback_failed/request，零 Attempt。
- [x] 初始尚未实现的 score/bool/图片能力在凭据前拒绝；Hook 无法启用未实现的题型或 message/image 形状。后续切片直接扩充真实实现。
- [x] 单选严格检查完整名称/候选集、精确字段、finite/range、1e-6 概率和与最高概率并列；单题 refusal 只输出 type=refusal、stop/nil error。
- [x] 基础已知 usage 和输入成本真实保留；不伪造 complete，不用 placeholder parser。任何损坏答案整体清空，2xx 后不重放，body/permit 释放。
- [x] Jev 基线持续通过，公共请求/Hook 的可变引用在冻结后修改不影响发送与结果；完整 artifact 包含 SHA-256 manifest、审计及准确重放命令。

## Completion

演示真实文本单选和拒答闭环，尚未扩展的能力保持拒绝。候选/正式目录与 live 结论留给后续切片，不把本票的合成目录记为内置支持。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 02

2026-10-12：待实施，公共字段目前尚未增加。

### 原票 03

2026-10-12：设计阶段；先完成 choice 的真实闭环，而不是先创建多个空协议 Module。


### 本票实施（2026-10-12）

已通过公共 Classify / HookedClient.Classify 打通合成宿主目录的 Decisions 文本单选与单题 RefusalAnswer，初始和最终 native 题目/JSON 都有完整限额，答案以最终冻结题目校验。新增 API、空 Options、文本能力、独立复制与 Hash；在固定凭据前完成四维身份及 capability 授权。尚未实现的 score/bool 在初始和 Hook 后拒绝，图片公共契约按工单 05 交付。内置目录与支持矩阵未改。

先列失败清单再做公共 E2E；文本/拒答、Hook 越权和限额、答案字段别名均有红→绿记录。独立包含 653 个最终公共 PASS（P10 195），6 次红色历史、SHA-256 manifest、审计、精确重放与缺包/篡改/重写拒绝。固定 Jev 包重新核验 655 case；相关 race 653 PASS，完整开启 Pi 的测试退出 0：6,426 E2E case，6,417 PASS、9 既有 NOT_RUN，815 项差分全部 PASS；两种 vet 通过。9 项压力/循环 opt-in 未执行，不表示 Decisions live 或完整 releasegate 通过。

两轴 code-review：Spec 0；Standards 发现验证器重开 ZIP 的 TOCTOU 和无界 manifest 读取两项边界问题，先补公共进程红色回归再修为有界不可变字节快照，最终复审记录见 [REVIEW](../text-choice-evidence/REVIEW.md)。审查修复不改库生产源码，修复后的六项证据 process E2E 与 race 独立重跑。

实现提交 `0b6a011`；证据、最终源码适用性、命令/退出码及审查修复见 [文本证据](../text-choice-evidence/README.md) 和 validation 收据。没有厂商真实请求；后续真实请求遵循用户要求使用本地代理。
