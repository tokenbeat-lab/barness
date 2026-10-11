# 01：核验 pi-ai 1.1.0 发布件与迁移差异

Status: resolved
Type: task
Depends on: none
Updated: 2026-10-11

## What to build

为唯一 1.1.0 oracle 取得可复现的发布来源、生成数据、源码等价证明和迁移清点。该工单负责来源核验与证据文档，不切换正式 oracle，不修改采样生产逻辑。

目标固定为 `@earendil-works/pi-ai@1.1.0`，commit `abe508e1b89912adde45528136c3221eb69acdd7`。以官方发布数据补齐固定源码缺失的生成目录，禁止运行会抓取当前线上模型的 generate-models。

## Ownership and context

负责本 feature 的 `release-verification/`、源码/发布件逐文件清单、可重复构建说明与迁移 inventory；大临时下载和依赖不入库。保留内容哈希与来源链接，不输出本机路径、真实凭据或环境秘密。协作时不要回退他人修改。

依据：[规范 R09–R10](../spec.md)、[设计 §6](../design.md#6-冻结基线与目录迁移)、[研究未核验范围](../../pi-1.1.0-thinking-sampling/research.md#9-barness-ai-接入映射与建议尚未实现)、[ADR-0004](../../../docs/adr/0004-barness-ai-frozen-pi-oracle.md)。

## Failure scenarios first

- npm gitHead 与 tag 不一致、integrity 不一致、缺生成数据或依赖版本来源不明确。
- 使用线上生成数据而不是冻结发布数据；只比较少数函数或文件便声称源码等价。
- 构建产物字节差异、第二个干净环境无法重现；仅忽略差异或改比较规则使核验通过。
- 未清点旧的六个差分子集、原生扩展豁免、SDK 错误规则或既有型号硬约束。
- 研究报告未核验发布模型默认，被误写成“1.1.0 全部内置型号没有默认值”。

## Acceptance criteria

- [x] B01：记录 official tarball URL、integrity、gitHead、源码 commit、实际 AI/telemetry 依赖、构建工具链与来源时间。
- [x] 两个干净环境以固定源码与该发布数据重建适用 dist；逐文件内容/数量/哈希与官方发布件核对，无未解释差异。
- [x] 交付发布模型数据清单及哈希；对所有既有聊天条目清点通用/分档采样、level map、已携带 compat 与新增硬约束；不扩大型号集合。
- [x] 交付旧→新差分 inventory：每项有协议/操作、原因、目标行为、existing extension 是否仍成立、解决工单；不把未解决项目登记为广泛 extension。
- [x] 复现命令不依赖研究时的本机 checkout，不含秘密；核验 artifacts 经既有审计或等价的受控来源审查。

## Verification and handoff

按[验收 B01](../acceptance.md#5-基线与发布验收)写出来源报告、两环境逐文件比较结果、生成数据 manifest 与 inventory。实际版本/hash 从执行中取得，不预填研究之外的数值。若不能证明源码等价，明确记录 FAIL、原因和阻断，不交付“已核验 oracle”。

交给[工单 02](02-migrate-single-oracle.md)的必要输入：已验证 lock 来源、模型数据哈希、工具链、等价证明，以及所有迁移必需修正的范围。工单完成要求这些输入齐备；有未完成外部条件时保留证据和未执行项，不标已完成。

## Comments

2026-10-11：方案阶段创建，尚未执行发布件核验。

2026-10-11 实施：两环境來源等价 PASS，827+24 文件零差异、43 文件模型数据、9+2 失败检查与完整逐 case 迁移清点已交付。 命令、实际 artifacts、失败历史和未完成项见[专项证据](../evidence.md)与[来源核验](../release-verification/README.md)。
