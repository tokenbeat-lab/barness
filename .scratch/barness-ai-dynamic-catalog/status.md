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

定向普通证据：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T094204.929586Z-1bw5afpm/receipt.json`，PASS/exit0；实际完成 TestMaintenanceCatalogMigration、TestBuiltinCatalogReleaseSnapshot、TestRegionalPromotionBoundary、TestKimiQualificationPromotionBoundary 四项。受测源码是工单工作树，完整来源与输入见原件；不是最终集成回执。此前失败/编译失败及初始通过共六个 attempt 已保留。工单 01 已完成（见下文）；03/04/05 尚在实施，最终完整普通验证仍 NOT_RUN。

架构细化：候选/存储接口已明确后，04 可先并行建立独立 runner 和证据消费；完成前必须合并 03 并验证真实 CLI 闭环。此调整不改变规格行为或验收范围。

最终记录受测 SHA、来源/输入、完整场景、audit、当前适用性与回滚方式；历史回执不覆盖。


## 集成进展与独立失败

- 已集成 01 完整加载器（合并 0e9ef483cb132c681e7a09b22e230190b19b9511）与 03/04 的受控检查点；当前还不是最终冻结来源。
- 01 最终定向回执：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T095707.682235Z-1bz74bwa/receipt.json`，PASS/exit0，155 项实际测试、40 个必需 CAT case、audit 通过。完整源码/输入与文档收尾提交集合相等性见 [loader-verification.md](loader-verification.md)。180,582 个原件已复制并逐一核对。
- 01 全包失败原件：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T094358.499706Z-718a6l1a/receipt.json`，FAIL/exit1，643 秒。原生原因 `route product input changed: ai/call_runtime.go`；既有测试随后对缺失 inputs 的类型断言 panic，主 bundle 未完成。不得将前序测试通过当成该 attempt 通过。
- 05 对此先保留组合源码 RED，再审阅四项确切 xAI Images 字节变更（等价路由提取、新定义入口、相同 classifier kind 判断提取、xAI 纯数据函数删除），保持旧原件和旧审阅条目。生成 JSON 只按构造数据分类，仍逐项比较原合同实际使用的模型。定向来源/原件组 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T101100.524194Z-rtd6323w/receipt.json` PASS/exit0，149 个 Go 测试、235 native PASS、audit0；尚待该工作树提交合并。
- 04 初始完整 118 条逐身份合约回执：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T095202.144349Z-2fgyn023/receipt.json`，PASS/exit0，909 个场景。后续非零缓存计价、原件缺项、原件重绑及输出保护等按独立 attempt 扩展，不把初始通过冒充扩展范围已完成。
- 03 已在真实 check/apply 后验证来源管理批量改价、人工覆盖/缺失费率保留、能力不变和 Provider 独立升版；生命周期、回退、批准来源导出与 fresh bootstrap、原始人工输入及非 nullable 类型校验仍在收尾。
- 流程偏差：04 的初始组合 CLI check RED 在实现前，但 apply 存储实现先于其专门 E2E；没有补写单元测试，也不补造专门 apply RED。实际公共成功/拒绝/文件失败原件保留，并由最终审查核对。详见工单 04。

全部阶段原件以各工单记录的完整路径、实际 source/inputs、场景与数字退出为准。最终冻结普通回执、构建/vet、双轴审查和只读发布判定尚未完成，release.pass 仍 false。

## 最终全量回归检查点

- 两轮静态复核已清空原双轴问题，记录见 [implementation-review-v2.md](implementation-review-v2.md)。这不代替全量执行。
- 集成受测 SHA `4265627e58579b22eaf052a359a8a29037a94463` 的完整普通命令发现 35 个中文评测失败事件；按 SOP 发送 SIGTERM，外层最终 CANCELLED/exit-15，主原生包未完成。490.502666 秒，原件及失败清单见 [integration-interrupted-v1.json](integration-interrupted-v1.json)。该 attempt 不算普通验证通过。
- 首个失败为 TestChineseEvaluationCurrentRoutePins/jev；其他场景报 task/model/dataset/price snapshot drift。单一修复代理正在定位旧消费者与目录迁移的关联；修复后重新冻结并验证。
- build/vet 元数据检查已 PASS/exit0，原件位于 `.evidence/barness-ai-dynamic-catalog/metadata/20261101T114832.746962Z-gccunam6/`。它们不代替主 E2E。
- 当前普通阶段：IN_PROGRESS（存在保留的失败/中断，尚无最终全量 PASS）。live/race/压力状态未升级，release.pass=false。
