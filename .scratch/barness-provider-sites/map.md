# 地区 Provider 实施图

## Notes

[规格与唯一当前状态](spec.md)。集成分支：`codex/barness-provider-sites`；起点：`68dec5a`。工单 Answer 保存各自固定范围的历史事实，[最终验证索引](evidence/verification.md) 关联整合结果与正式交接。

## Decisions-so-far

- [01 协议](issues/01-protocol-profiles.md)：已解决，合并 `77e8796`；私有 codec 配置与实际 Provider 归属、协议/原生历史隔离。
- [02 隔离](issues/02-runtime-isolation.md)：已解决，合并 `11475c6`；缓存 v2 固定向量、四服务多租户准入、受控并发与资源边界。
- [03 Host 与候选](issues/03-host-catalog.md)：已解决，合并 `3e63d79`；四服务装配、新九候选、独立官方来源与固定 FX；后续整合补齐 budget 和凭据目标校对。
- [06 冻结差分](issues/06-frozen-differential.md)：已解决，合并 `1a8220e`；两条真实适用路线及三型号比较，另四条 extension；冻结数据与旧八子集保留。
- [04 自身证据与发布](issues/04-evidence-release.md)：已解决，合并 `6833018`；独立 live namespace、keyless 原生消费者、23 行目标与追加追溯门禁；后续审查修复见下项。
- [05 整合验证与迁移](issues/05-integration-validation.md)：已解决；最终受测 `d4b63d9`，双轴审查与同一 implementer 的修复、旧 golden 迁移、完整普通、双 vet、来源/原生审计及人工交接。固定事实见[最终索引](evidence/verification.md)，[审查原文](evidence/review.md) 保留发现和复审。
- 01 / 02 / 03 分工并行；04 / 06 依赖前三项；05 依赖所有实现。测试接缝保持公共 Client、Host 装配和 CLI / 原生证据消费者。
- 未增加依赖、升级 SDK 或冻结 pi，也未改 BuiltinCatalog。内部旧实现直接删除，无兼容 shim。

- [07 完整 race 与最终验收](issues/07-final-acceptance.md)：完整开发 race PASS，真实候选按独立 US$5 授权完成 81 Calls / 81 HTTP，五路线 qualify PASS、一条失败；正式验收继续开放。[本轮事实](evidence/acceptance.md)。[唯一当前状态](spec.md#当前状态)。

- [08 ZAI Messages GLM-5.2 原生推理资格](issues/08-zai-messages-glm52-native-qualification.md)：07 的实际真实失败；独立公共 Client 离线复现表明捕获无原生字段，保留资格缺口和后续授权边界。

- [11 地区目录受控晋升](issues/11-regional-catalog-promotion.md)：inactive 生产 hook、六 own candidate 严格 consumer/独占 intent、不可激活的 controlled 提案和两态 Host 组合已实现；候选正式显示名与 `.2` snapshot 固定。实际 live 晋升仍等待 08 资格及最终审查，状态见 spec。

- [10 有限采集与来源发布](issues/10-final-round-collection.md)：固定 33 完整进程＋原始研究 4/4，跨阶段一次 intent/美元保留、选中 Key 隔离、最终 wire guard、原生来源闭包、GLM 新原件及 23 行公共 importer 已实现并保留普通 RED→GREEN；待父任务独立审查和真实调度。实际批准/资格/最终验收见 spec。

- [12 ZAI Chat 固定原生来源](issues/12-zai-chat-native-source.md)：本轮上游工具响应未返回 reasoning，预先固定同轮 full-stream 原生阳性消息并保留工具链；受控修复与新的有限重验准备由 spec 记录。

- [14 MiniMax Chat](issues/14-minimax-global-chat-native-thinking.md)、[15固定恢复3](issues/15-fixed-recovery-authority.md)：各自范围已解决；实际6b87a584三M3各完整13 PASS，授权机制及失败即停已验证。后续阻断是[16 ZAI Responses固定自身推理来源](issues/16-zai-responses-native-source.md)，具体当前验收/金额只见spec。

- [17固定恢复4](issues/17-fixed-fourth-recovery.md)：准备完整closed3历史、固定ID4及24进程新有限计划；所有实际新增批准/来源/预算/验收仍只由spec当前状态与其原件指针记录。本票不授予新调用。

- [21 中国MiniMax Responses固定原生来源](issues/21-minimax-cn-responses-native-source.md)：自身原SSE公共回放证实工具响应未带native；预定自身文本源与严格consumer同步，保留独立工具链和所有实际失败。来源资格、后续有限准备/预算只见[spec当前状态](spec.md#当前状态)。

## Fog

正式资格和验收的未完成项只在 [spec 当前状态](spec.md#当前状态) 维护；Key 配置变更不授予付费或 race。后续入口与人工命令见 [验证交接](evidence/verification.md#人工-race-与自身-live-交接)。
