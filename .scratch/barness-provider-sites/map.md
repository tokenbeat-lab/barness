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

## Fog

正式资格和验收的未完成项只在 [spec 当前状态](spec.md#当前状态) 维护；Key 配置变更不授予付费或 race。后续入口与人工命令见 [验证交接](evidence/verification.md#人工-race-与自身-live-交接)。
