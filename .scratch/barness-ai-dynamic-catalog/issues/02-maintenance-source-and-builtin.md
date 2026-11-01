# 唯一 Provider 维护源与默认 JSON 迁移

Type: task
Status: resolved
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

2026-11-01：公共验证边界为 maintenance.BuiltinState/Compose/Generate → ai.BuiltinCatalog/公共 Client；先核对独立历史快照及 JSON 原件。失败方式：迁移丢项/重排/改变 null 或费率、未保留来源缺口、生成不确定、Provider 版本串升、下架被补回、返回副本共享、重复身份/位置和错误维护数据被接纳。

## Answer

已迁移唯一 Provider 数据源、维护类型和确定性生成。内置默认只读取生成的 JSON，原字面量和独立 runtime activation 路径已删除。保留原历史快照、三份历史 activation 原字节及现有 Kimi/xAI runtime 约束；旧 proof 工具仅保留受控提案或只读资格消费。默认和外部生成保留 118 条记录及原顺序、所有 typed 默认/null/价格语义；新 Provider/合成版本不冒充旧版本。

验证在独立 worktree 基准 4eae987a0a0ee32083cdfd186af80654ac1763be 上执行，运行时 dirty payload 的实际 source/selectedInputs 由下列回执完整绑定；冻结集成验证另由 05 执行。

- 初始公共迁移红：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T093858.062585Z-bfv4fwhp/receipt.json`，FAIL，进程 exit 1（维护包尚不存在）。
- 开发生成：`.../20261101T093947.162542Z-v3r4tlad/receipt.json`，INVALID_EVIDENCE，进程 exit 0；命令生成了受控来源文件，故不能充当测试通过证据，原件保留。
- 迁移绿：`.../20261101T094019.557295Z-brvu366r/receipt.json`，PASS，exit 0。
- 旧 activation 退役合同红：`.../20261101T094122.763709Z-z2yap0zi/receipt.json`，FAIL，exit 1。
- 退役编译失败：`.../20261101T094139.264753Z-828j13cm/receipt.json`，FAIL，exit 1；删掉 writer 后 import 不匹配，已修复，原件保留。
- 最终定向绿：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T094204.929586Z-1bw5afpm/receipt.json`，PASS，exit 0。真实入口 `go test -json ./ai/e2e -run '^Test(MaintenanceCatalogMigration|BuiltinCatalogReleaseSnapshot|RegionalPromotionBoundary|KimiQualificationPromotionBoundary|RegionalCatalogComposition)$' -count=1`。实际完成四项：TestMaintenanceCatalogMigration（CATALOG-T01-maintenance-migration）、TestBuiltinCatalogReleaseSnapshot（catalog-release-snapshot）、TestRegionalPromotionBoundary（RP18-promotion-boundary）、TestKimiQualificationPromotionBoundary（KIMI-K16-promotion-proof-boundary）。请求的 RegionalCatalogComposition 不匹配真实 TestRegionalQualificationCatalogComposition，故该项未执行，不计覆盖。
- 主包 `bundles/20261101T094211.041010000Z/manifest.json` SHA256 `60485b353898fedc0d16dfc64ea17bc9fb41410e6481f49251cbbb6c6adc5e30`；audit SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。各源码/数据/生成命令文件用 selectedInputs 显式绑定，具体 argv 和哈希见原回执。

未执行 live、race、压力及完整普通回归；未声明厂商资格或 release.pass。最终 broad 回归、source applicability、审查和状态索引在 05 集成处理。
