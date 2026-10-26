# MiniMax 自适应推理正例探针

Type: task
Status: claimed

## Contract

[Spec](../spec.md) R23/E18；[SOP](../../../docs/agents/verification.md)。本票只修正既有MiniMax-M3原生thinking-on正例任务不足以稳定引出推理的问题。保持真实观察到非空推理的正例断言、模型身份、固定场景ID/完整集合、SDK、预算、账户/endpoint与正式消费者不变。原始FAIL不改成PASS或UNSUPPORTED，不使用模型升级、effort档位、增加重试/调用、来源降级或宽松消费避过问题。

## Failure

v4同源续跑第8条国际MiniMax Chat：14逻辑calls/13HTTP，13场景PASS（含xai-origin-history）、thinking-on FAIL。真实请求thinking.type=adaptive/reasoning_split=true/2048输出，仅要求“Synthetic protocol probe. Reply with the integer 42.”，服务器HTTP200原始SSE无reasoning_content/details，usage.reasoning_tokens=0；native断言own reasoning actually observed失败。同批full-stream同提示曾输出8推理token且适配器正确保留，暂无产品解析丢失证据。完整原件及7条有效资格/后9NOT_RUN由[当前状态](../spec.md#唯一当前状态)和独立停止检查点保存。

## Ownership

单一implementer拥有最小MiniMax native探针请求构造、公开进程/E2E回归及必要说明/证据摘要；在独立现有工作树，先基于集成分支同步。不得改root唯一状态/map、05工单、既有真实原件/授权/grant、目录、资格消费者或其他Provider行为。Root负责保存真实失败、后续来源冻结与付费/人工选择。

## Verification

先在公开harness/进程边界建立无厂商的受控RED，复现合法自适应响应对直接回答42任务产生零推理、正例探针失败；保留真实命令/exit/bundle/audit。随后只为thinking-on使用固定、有实际推理需求、输入/输出与时限有界的任务；继续要求该case自身实际非空推理，保留请求type=adaptive和现有响应完整性检查。正/反例证明任务变化会触发探针、空推理仍拒绝、原Simple/Full/工具/历史场景不被意外更改，按影响复用现有公共native受控E2E，避免实现镜像式单元测试。不能使用真实Key/live/race/压力，不建新框架、不复刻生产代码、不借自报true取得正式资格。文件保持500行以内。

源码/探针变更需要重新冻结实际来源、二进制、输入与完整普通；历史7条当前来源资格保留为历史，不冒称新来源已验收。新真实调用另取明确授权，不复用本次已claim grant或closed ledger。若复现揭示不同原因，先报告证据，不强行照本票假设改码。

## Answer

待实现并独立复核。

## Comments

- 2026-10-27：实际付费失败后建立离线修复前沿；本票不产生新的付费或人工阶段权限。
