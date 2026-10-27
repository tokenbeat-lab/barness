# 工具公开 CLI 夹具补齐 Grok 历史证明

Type: task
Status: claimed

## Contract

[Spec](../spec.md)与[测试 SOP](../../../docs/agents/verification.md)。修复工具公开只读消费正例中的旧 Google 合成捕获：夹具必须满足当前原生完整必交合同及真实形状的 host/provenance/跨型号证明，不能仅补场景标签或放宽消费者取绿。

## Actual failure and feedback

21 的冻结工具回归实际运行49项，48 PASS、1 FAIL；`test_live_release.test_complete_chat_captures_are_consumed_readonly` 被编译后的原生消费者拒绝。真实原件：[工具回归](../../../.evidence/barness-ai-grok-glm-tool-process/e2e/20261027T020833.044598Z-9_0tuxcl/result.json)。现有 `live_fixture.py:chat_bundle` 仍生成旧六场景 Google 夹具，缺少当前强制 Grok 外来历史场景及独立捕获证明。此公开 CLI 正例的实际失败是反馈 RED，保留原件；标签存在不等于证明完整。

## Ownership

单一 implementer 复用21的 native-proof 独立工作树，最终同步集成 tip。你不是唯一工作者，不得回退他人编辑。拥有 `tools/verification/tests/live_fixture.py` 和既有 `test_live_release.py` 的最小正反例：先保存 label-only/provenance 篡改仍被拒绝的公共 CLI 证明，再补齐完整七场景合成夹具。以当前类型合同及捕获校验为依据，不复制其他 Provider 证明冒充 Google。不修改生产消费者、原生 harness、目录、模型、SDK、预算、来源选择器或已冻结授权。文件小于500行。

不拥有 root Spec/map/05/21/22 状态；不读 `.env` 或真实 Key，不写真实 grant/store，不发厂商请求，不执行全完整普通、race、压力或正式镜像构建。root 负责合并、双轴审查、冻结完整普通及授权适用性。

## Verification

保留上述真实 RED。先提交缺证明/证明篡改的公共消费者拒绝场景，再实现夹具修正；重跑完整公开 Python 工具套件和新增反例，保持原49项负例及有限胶囊约束。21 的原始两 GLM 入口、32辅助、四项取消/时序及 MM18 既有完成证据按实际来源关系交付；只有新改动影响它们时才重跑。最终差异须独立确认四个原生来源 map、八镜像及17 grant 字节未变，不能预判或自动重绑 v8 来源。

## Comments

- 2026-10-27：从21工具完整回归发现并拆票；仅离线合成夹具修复，v8 仍0真实调用/0 claim，原17权限未使用。
