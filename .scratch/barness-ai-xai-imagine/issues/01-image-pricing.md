# 按张价格、目录与报告收费领域字段

Type: task
Status: resolved
Blocked by: none

## Scope

公共目录边界先失败后实现；增加 APIXAIImages、ImagePricing.PerImage、完整价格校验/深复制/hash、独立候选目录和 Attempt.ReportedCostUSD。保持旧 token 定价合同。

完整契约：[spec](../spec.md)。验证遵循 [SOP](../../../docs/agents/verification.md)。

## Comments

2026-10-29：implement-spec 根据现有规格建立工单依赖图。

## Answer

增加 APIXAIImages、互斥的 ImagePricing.PerImage 分支与独立 ReportedCostUSD。xAI 图像目录限定协议/型号/能力；价格表拒绝缺项、null、重复、非法档位、负数和非有限值。目录按张价格独立复制并参与既有内容 hash。独立 ImagesCandidateCatalog 保留 2026-10-29 候选能力和价格，不修改 BuiltinCatalog。

公共边界为 NewClient、Catalog.Hash、Catalog 查询。验证命令使用 `python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e -run '<对应表达式>' -count=1`。全部原件保留在 `.evidence/barness-ai-xai-imagine/e2e/`：

- `20261029T000247.418077Z-6ihexuwt`：FAIL/1，初次测试误用不存在的 Client.Close，保留编译失败。
- `20261029T000312.354910Z-hgx0_ze6`：FAIL/1，公共入口明确拒绝未实现的 xAI 图像身份。
- `20261029T000344.317248Z-_4axd3z3`：FAIL/1，测试期望遗漏旧目录编码的 models:null，保留失败。
- `20261029T000402.254237Z-dheznpdc`：PASS/0，目录接受与完整 JSON 保留。
- `20261029T000443.714698Z-rcxzltkv`：FAIL/1，新增价格/能力拒绝场景证明确有缺失。
- `20261029T000546.864242Z-d87cnb7a`：PASS/0，目录接受与全部拒绝场景。
- `20261029T000644.489160Z-lb5pdcrl`：FAIL/1，目录价格指针和行共享导致 mutation 污染。
- `20261029T000714.079747Z-sho57u6n`：PASS/0，独立快照、查询、hash 与显式零。
- `20261029T000741.304482Z-eocyk505`：FAIL/1，新候选加载入口尚不存在，保留编译失败。
- `20261029T000808.364556Z-nz307pb7`：PASS/0，最终选定目录检查与相邻目录回归。

ReportedCostUSD 的公共收费先失败场景由工单 02 留存：`20261029T000504.559048Z-rvq0u012`，TestXAIImagineResponseContract 的 cost-only/cost-zero/failure-retains-facts。金额是值类型 Nullable；Observer 标量审计白名单加入 reportedCostUsd。

本工单不授予真实资格。完整受影响普通包及最终来源冻结由工单 04 负责。
