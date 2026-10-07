# 02: 核验 pi-ai 1.0.0 发布件

**What to build:** 给维护者一份可重复核验的 pi-ai 1.0.0 发布件证明，使后续冻结基线切换能依据确切源码、依赖完整性和模型数据，而不是在线刷新目录或依赖研究工作树。

**Blocked by:** None (can start immediately).

**Status:** resolved

**依据：** 规格“基线迁移”、实施设计第 4 节，以及 ADR-0004。核验成果交给工单 03；本工单完成时运行中的唯一 oracle 仍为原基线。

- [x] 在隔离的临时核验环境使用 npm @earendil-works/pi-ai@1.0.0，以及配套 pi-telemetry 发布件；固定 tag v1.0.0、目标 commit a13d35a742c6ef8462812a28fbe1d8c8b7431c32 与实际 npm gitHead，并解释任何不相等之处。
- [x] 记录发布包 integrity、精确依赖锁定信息、生成模型数据的逐文件 SHA-256 与完整模型数据哈希；核验安装跳过脚本，生成数据来自冻结发布件。
- [x] 按 ADR-0004 的既有等价核验步骤，用冻结源码加发布数据重建 ai 与 telemetry，逐字节比较发布产物；报告比较范围、文件数、差异数和复现步骤。
- [x] 先列核验失败方式：错版本、源码与 gitHead 差异未解释、缺数据、数据哈希不符、构建产物不等价、依赖未锁定。出现任一问题即明确报告失败，不写成已核验通过。
- [x] 核验材料和脚本输入可由另一个干净环境重放；实际用于切换的 provenance 只在工单 03 启用，不在仓库保留第二套长期 oracle 或 runner。
- [x] 交付有完整性清单和脱敏审计结论的核验证据，明确这是发布件等价证明，还不是 barness 在新基线下的差分通过证明。

## Comments

**2026-10-08 — implemented**：交付 [完整证明与复现说明](../release-verification/README.md)、
[固定输入](../release-verification/inputs.json)、[脱敏证据](../release-verification/evidence/report.json)
与 [材料完整性清单](../release-verification/evidence/SHA256SUMS)。

- tag、目标 commit 与 ai / telemetry 两个实际 npm gitHead 都为 `a13d35a742c6ef8462812a28fbe1d8c8b7431c32`。
- 两个独立的干净临时环境按同一锁安装并重建；ai **811/811**、telemetry **24/24** 个完整 dist 文件逐字节一致，**0** 差异；模型数据共 **43** 文件，完整哈希 `8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e`。
- 9 个篡改拒绝场景 PASS；证据审计 `findings: []`，安装使用 `--ignore-scripts`，未在线生成数据。
- 仅保留验证输入和静态证据；临时安装与源码自动删除。运行中的唯一 oracle 仍为 **0.87.1**；新基线的 barness 差分为 **NOT_RUN**，由工单 03 接续。
- 完整 Go 套件 2877 个场景 PASS（含原基线差分 464 个），两种 vet 与审计 PASS；代码审查发现的下载限额 / 失败清理问题已修复，并补充 2 个边界故障场景。双轴复审最终均为 0 项，见 [验证与审查记录](../release-verification/verification.md)。
