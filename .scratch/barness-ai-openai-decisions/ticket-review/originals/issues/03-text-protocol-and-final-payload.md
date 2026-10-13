# 03 — Decisions 单选文本调用与最终 payload 闭环

Status: ready-for-agent
Type: task
Depends on: 02
Updated: 2026-10-12

## What to build

首个真实 vertical slice：公共 Classify 完成一次 Decisions choice 请求、严格 choice/单题 refusal 结果及基础已报告用量。接入 fixed route、shared unary、native 编码、Hook 后 guard/freeze；score/bool/图片在此阶段 capability 明确拒绝且不读凭据，下一工单补齐。依据 [R01–04/R08/R09/R10/R13](../spec.md)、[设计 §3/4](../design.md)。

## Ownership and context

负责 classifier/callRuntime 的实际路由和固定分派、新 Decisions orchestration/request/guard/首个 response 切片；既有 Typesafe wire 不动。使用宿主合成目录和 loopback Provider，不改正式 catalog，不增加 SDK/client/registry。协作不回退他人编辑。

## Failure scenarios first

- 新 API 误路由到 TypeSafe 或 chat，使用环境 key 或不同 endpoint。
- State 被双重编码，JSON 数字经 float64；map 随机顺序导致冻结请求变化。
- 单候选或结构化说明送到 OpenAI；错 Options、未实现的题型在发送后才拒绝。
- payload hook 换 model/认证/目标，guard 仍读取 state/map 或验证原题。
- 返回合法 choice/refusal 不能被解析，或返回旧 candidate/missing/duplicate 时被接受。
- 注册空 parser/返回假成功形成不可用的中间 API。

## Acceptance criteria

- [ ] 先写 OD02–04/09–12/14/16/22–23/25 的单选适用变体，并写 choice、refusal 和损坏响应闭环 E2E。
- [ ] 只允许 classifier/openai/openai-decisions，白名单/目录交集在凭据前；固定快照和 x-request-id 有效。
- [ ] input text、string value、null/空说明省略、排序、unique names、native strict shape 按设计成立。
- [ ] Hook 合法换题按最终 frozen record 验证；越权/错误 schema 零 Attempt；原输入独立。
- [ ] 正常 choice 严格检查名称/候选/最高概率/finite/range/sum，单题 refusal stop/nil；非 2xx 有界脱敏。
- [ ] shared sendUnary 完整读/parse/close/release；正常和失败结果真实可用，没有 success stub。
- [ ] 阶段性已实现 kind 检查在 capability；不声称所有能力已交付，04/05 用实际实现替换该支持集合。
- [ ] 01 Jev 对照继续通过，旧通用错误优先级和 native 回调不改变。

## Verification and handoff

保存真实 public request/result/usage/Attempt/hook 次数/归属/资源 artifact，并运行 target E2E 和 Jev 对照。基础 usage 使用本阶段声明的合成 fixture；复杂 presence/tiers、score/bool/mixed 的待交付项显式关联 04，不能记正式支持矩阵 PASS。

## Comments

2026-10-12：设计阶段；先完成 choice 的真实闭环，而不是先创建多个空协议 Module。
