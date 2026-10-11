# 04：在 Chat 与 Responses 请求边界按 canonical level 应用采样

Status: resolved
Type: task
Depends on: 03
Updated: 2026-10-11

## What to build

让两个消费 API 的请求构造复用一个小的采样语义模块：依据入口输入与 Model 能力选档，在具名 JSON 之后依次应用通用、当前分档、调用覆盖。保留 full/simple wire reasoning 的差异与既有 DeepSeek 分支。

## Ownership and context

负责共享采样构造职责、Chat/Responses 的 options/request 边界及对应公开 E2E fixture。既有 `applySamplingParams` 从 Responses 专属位置提取时删除旧实现，不留 shim；不把整个 Model/Profile 塞进 DTO。协作时不要回退他人修改。

依据：[规范 R02–R05、R07](../spec.md)、[设计 §3–§5](../design.md)、[ADR-0025](../../../docs/adr/0025-barness-ai-thinking-level-sampling.md)。

## Failure scenarios first

- 使用 requested level 或厂商 wire value 查 profile；off=null、非推理、xhigh/max、全 null 选错。
- simple 预先合并导致重复物化与输入修改；两个 builder 复制不同规则。
- 具名 temperature 覆盖模型默认、null/0/false 丢失、嵌套递归合并或空档清空默认。
- 为“统一实际强度”自动夹取 full wire，summary-only 从 off 起算，省略 reasoning 却补 effort。
- DeepSeek disabled/默认被改写，或 Anthropic/Gemini 开始发送新 map。
- 调用 map 修改 reasoning 后重新选档，导致优先级或请求不确定。

## Acceptance criteria

- [x] 先写 S01–S17 公共失败场景与 full/simple 对照，再实现请求映射。
- [x] Chat full 从 effort/off 起算；Responses full 从 effort/summary-medium/off 起算；simple 保留既有转换，采样只应用一次。
- [x] 复用已有 clamp，canonical 与 wire 分离；不新增公开 resolver、策略引擎或 Provider 注册。
- [x] 四条消费路线、四个 Client 入口与核心 WithHooks 无修改变体符合验收矩阵；Stream/Complete 请求采样一致。
- [x] full low→sampler high、summary-only wire medium→sampler high 等分离行为按目标 oracle 核对，文档不声称最终 wire 一定相同。
- [x] DeepSeek enabled/disabled/default 与原生忽略协议保持原合同；DeepSeek Responses 自身请求证据明确为扩展路线。
- [x] 三层浅覆盖保留所有合法 JSON 值；不变更输入数据，OnPayload 之后不 reselection。

## Verification and handoff

每个 case 保存真实捕获请求、响应脚本、终态与断言；适用场景开启唯一 1.1.0 oracle，比较完整 payload 和既有事件/结果，不只比较采样键。原生扩展明确 skip 并自证。

交给[工单 05](05-isolation-and-differential.md)采样请求路径和全部功能证据；任何未解释差异登记 pending，不在本工单增加无依据 extension。文件保持小于 500 行，测试不依赖私有 helper。

## Comments

2026-10-11：方案阶段创建，尚未执行目标请求映射场景。

2026-10-11 实施：共享 canonical clamp 在两个请求 builder 单次应用，所有消费路线/四入口/核心 Hooks、full/simple 入口差异和 native ignore PASS。 命令、实际 artifacts、失败历史和未完成项见[专项证据](../evidence.md)与[来源核验](../release-verification/README.md)。
