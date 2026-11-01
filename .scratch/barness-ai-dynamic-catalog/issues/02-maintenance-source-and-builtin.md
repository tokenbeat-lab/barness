# 唯一 Provider 维护源与默认 JSON 迁移

Type: task
Status: ready-for-agent
Blocked by: none

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责 catalog/data/、catalog/maintenance/ 的基础数据模型与确定性生成，以及 ai 默认目录数据替换。冻结旧 ai/release/catalog-snapshot.json 为独立迁移基准，保留 118 条定义及 null/缺省语义。定义 Provider 独立版本、稳定合成版本、字段归属/来源/生命周期的维护类型，生成内置及外部同源 JSON。删除旧字面量维护路径，保留历史快照/pin。先与后续 CLI 同步简洁公共接口。

## Acceptance and verification

T01、T07、T10；独立旧快照与公共 BuiltinCatalog/生成入口比较。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。
