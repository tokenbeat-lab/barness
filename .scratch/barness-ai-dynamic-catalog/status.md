# 模型目录维护与本地加载：当前状态

Updated: 2026-11-01
Implementation: IN_PROGRESS
Ordinary verification: NOT_RUN
Release: false

## Scope and source

- 授权：用户调用 implement-spec 实施本功能；仅普通受控开发与验证，不授权付费、race 或压力。
- 基准完整 HEAD：9a43ec0827b0ab794501682a7e66037a592a2f8d。
- 开始工作区：只有未跟踪的本功能 spec.md；原分支 codex/barness-ai-dynamic-catalog。
- 集成分支：codex/barness-catalog-integration。
- 实际影响：新增本地隔离加载、维护组件/CLI、候选执行与文件保存；默认目录从字面量迁为同源生成 JSON。ai 协议和调用生命周期保持既有合同，实际差异在审查中逐项核对。
- 公共边界：维护 CLI/子进程 → 候选/check/apply → 本地文件 → 公共加载 → Config.Catalog/公共 Client；损坏文件直接通过公共加载入口验证。此边界已由 spec 的 Testing Decisions 确认。
- 普通范围：新增组件和 CLI 进程 E2E、受影响 ./ai/e2e 与 ./ai，以及新增包编译/vet。具体测试发现集合与必需 case 在各工单中固定；tools/verification/run.py 留存独立 attempt。
- 输入：旧 ai/release/catalog-snapshot.json 是迁移基准；实际源码/新增组件/CLI/维护数据/规则/来源固定字节显式枚举，不能依赖验证工具默认覆盖。
- 来源政策：本轮不直接请求上游；固定 fixture 支持来源映射验收，实际维护可显式钉住输入，历史研究不冒充完整来源原件。

## Task graph

01 本地加载、02 唯一维护源并行；02 → 03 CLI；01 + 02 → 04 合约 runner/检查框架，04 的完整应用闭环须整合 03；全部 → 05 审查验证。见 [issues](issues/)。

## Stages

| 阶段 | 状态 | 范围/依据 |
| --- | --- | --- |
| ordinary | NOT_RUN | 实现前失败场景、随后受影响普通 E2E/编译/vet；按独立 attempt 记录 |
| live | NOT_APPLICABLE (pending review) | 迁移既有定义及受控本地接纳不取得新厂商事实；若实际协议变化需重新评估，不执行付费调用 |
| race | NOT_RUN | 维护存储、执行取消等并发路径候选；普通完成后列出实际范围供人工选择 |
| pressure | NOT_RUN | 加载/执行资源与规模影响待最终评估；不隐式执行压力 |
| release gate | NOT_RUN | 未完成人工适用阶段；不声明完整发布通过 |

## Evidence

工单 02 已合并：f8e234c928c5d901361707322739ed3425e29563（功能提交 cd83be49e99a19dd423dbda95b8434d4f603a963）。唯一维护源、同源生成及旧激活路径退役完成。全部 118 条定义/顺序/null/缺省语义由独立旧快照对照；厂商实测历史未改写。

定向普通证据：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T094204.929586Z-1bw5afpm/receipt.json`，PASS/exit0；实际完成 TestMaintenanceCatalogMigration、TestBuiltinCatalogReleaseSnapshot、TestRegionalPromotionBoundary、TestKimiQualificationPromotionBoundary 四项。受测源码是工单工作树，完整来源与输入见原件；不是最终集成回执。此前失败/编译失败及初始通过共六个 attempt 已保留。工单 01/03/04 尚在实施，完整普通验证仍 NOT_RUN。

架构细化：候选/存储接口已明确后，04 可先并行建立独立 runner 和证据消费；完成前必须合并 03 并验证真实 CLI 闭环。此调整不改变规格行为或验收范围。

最终记录受测 SHA、来源/输入、完整场景、audit、当前适用性与回滚方式；历史回执不覆盖。
