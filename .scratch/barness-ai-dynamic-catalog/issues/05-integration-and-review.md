# 全规格集成、双轴审查及普通验证

Type: task
Status: claimed
Blocked by: 01, 02, 03, 04

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责集成收尾与公共端到端缺口、用户文档、唯一当前状态和可重复证据。以初始提交为固定点按 code-review 技能双轴审查，统一修复，再冻结普通来源并执行完整受影响普通包/构建/vet，记录真实 SHA、哈希和未执行阶段。不得把受控通过写成厂商资格或完整发布通过。race/压力适用时仅列具体人工选择请求，未授权不执行。

## Acceptance and verification

T01–T13；普通证据、全部必需实际场景及 audit，状态/索引一致。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。

2026-11-01：开始来源适用性与文档子范围。源代码基准 0e9ef483cb132c681e7a09b22e230190b19b9511；03/04 后续仍在实施，完整普通与双轴审查未执行。验证边界为现有公共 releasegate 资格 CLI、控制捕获及来源拒绝场景。当前状态由主 agent 统一维护。

## Checkpoint 1 — 来源适用性与文档

2026-11-01：有限来源审阅见 [source-review.md](../source-review.md)，不可变命令/退出/哈希索引见 [integration-verification-v1.json](../integration-verification-v1.json)。四项精确来源迁移仅适用 xAI Images；原审阅条目、原始测量、pin 与支持矩阵未改写。运行数据新 JSON 按实际型号投影核对，新增验证/helper 仍属于产品来源，未来编辑、改原来源或错路线继续拒绝。

- RED：100651 原历史原件 CLI 失败；100753 新精确迁移/分类失败；100826 首次修复仍暴露 classifier-kind 提取的源码漂移。全部实际 exit1 原件保留。
- GREEN：100924 定向通过；101100 扩展 Evidence/XAIImages 组 PASS/exit0，149 实际 Go 测试、235 原生 PASS、audit0；101407 安全失败报告修订后的聚焦组 PASS/exit0。完整路径/argv/manifest 与 audit 哈希见索引，不以父测试或零匹配代替。
- 文档：新增本地加载/候选检查应用/字段归属/价格/版本与生成指南，修正现有 README/contract 的旧字面量维护入口和默认快照指向。六个相关文档的本地链接核对无缺失。
- 未完成：03 最终 Provider 导出命令的指南补齐、03/04 最终代码合并后的完整普通/编译/vet、主 agent 双轴审查和最终状态/发布只读索引。保持 claimed，不声明 release.pass=true。

## Checkpoint 2 — 最终 CLI 指南准备

2026-11-01：已按 03 已提交实现补齐 `cost-update`、按 Provider 的 `rollback`、`--source-kind official`、冻结本地定义/决定及 `export --review-dir` 完整审阅包；按 04 补齐 `candidate --check --repo` 的严格只读状态查询。指南中的 28 个参数名与实际 CLI/生成器声明核对一致，六个相关文档本地链接无缺失，最大 476 行。尚未执行最终普通阶段；等待 03 合并与主 agent 明确冻结信号，再检查最终帮助/完整普通/编译/vet。当前状态仍只由主 agent 维护。

## Checkpoint 3 — 本地定义的实际能力与诊断

2026-11-01：主 agent 在最终冻结前确认两项具体缺口。分类器结构自洽但超出固定协议子集时现报告 unsupported_definition；非法限额仍为 invalid_definition。Google 声明未实现的 mask/透明背景/custom size/fidelity/size/quality，以及 OpenAI 未消费的 aspect/imageSize 或实际选项 guard 不允许的数量/quality/fidelity 声明，不再伪装加载成功。参考数复用既有协议 guard 常量。

新增公共 Load/Hash/NewClient 红绿场景，完整旧 Hash/NewClient 接受/错误文本保持；原 guard 条件未变、无能力扩展或新厂商规则。追加精确来源审阅只在原 xAI Images 范围生效，旧条目/旧文档保持；见 [v2](../source-review-v2.md)、[v3](../source-review-v3.md)。完整命令、失败/通过、实际退出及 manifest/audit 指针见 [integration-verification-v2.json](../integration-verification-v2.json)。

分类范围 104438 PASS/exit0，最终图片/分类/加载/证据及全 118 合约扩展范围 105117 PASS/exit0；保留 104121、104252、104838 三次独立失败。仍待 03 最终合并后的完整普通/build/vet 和主 agent 双轴审查；不声明完整发布通过。

2026-11-01：双轴审查见 [implementation-review-v1.md](../implementation-review-v1.md)。Standards 2 项（最高 P2）、Spec 4 项（最高 P1），由单一实现代理修复。完整普通验证须在修复复核后冻结执行。

## Review fixes — 单一实现者收尾

2026-11-01：按双轴原审查的五项独立缺陷完成修复；原 Standards 2 / Spec 4 的计数与严重级别独立保留，不改写 [原审查](../implementation-review-v1.md)。修订阶梯在 typed 转换前校验阈值和四项费率，显式零有效；新增映射、关联与刷新共用来源操作规则；完整审阅基线随 Provider/运行快照一起采用，后续种子 revision 留存原件，继承的旧证据引用明确绑定原 review index 而不宣称可取/资格；生成器参数错误走 JSON；重命名安装与目录持久化分别判定，保存失败有界恢复，恢复未完成或持久化不确定时明确报修复要求，保留不可变孤儿对象。

补充闭环：同一坏输入在 128 次公开加载中实际出现两种原因（113/15），已固定原始字段遍历顺序。复审发现新增时的推理接线规则误用于既有修订，公开 RED 覆盖九个旧模型的 27 次名称/价格/空覆盖，现只复用价格存在性检查。Gemini 2.5 Pro 空覆盖又实际暴露数组内对象键顺序造成伪差异，现保留真实数组差异粒度并忽略这类无语义变化。新嵌入基线占位文件先导致单模型合成夹具编译失败；公开 check→留存源码重编译又证明实际 source archive 也缺该编译输入。两条复制/枚举路径均补齐完整基线资产，没有新增扫描框架或依赖。

最终受测代码 `853279be371ab6887b183d85dd147ecd31c5debd`。组合定向普通回执 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T114007.854682Z-0qdbggi8/receipt.json`，SHA256 `e90d82178f357ff96910ab502f03ff971994291911023df685cbed8d4888eefd`，PASS/exit0，346.9136 秒；1747 个实际 Go PASS、1597 个原生 PASS、audit0。覆盖 Catalog/Maintenance/LocalCatalog、相关来源/资格回归，以及 `./ai ./catalog/... ./cmd/...` 编译与匹配测试；完整 argv、精确必需场景、源码/输入和所有独立失败见 [不可变索引](../review-fixes-verification-v1.json)。实际输入固定完整 catalog/cmd 源码与数据（包括嵌入占位文件）和三份来源审查依据。

失败仍保留：初始修订、来源操作、生成器参数、真实 mode0300 的安装后同步故障、采用基线及元数据/原 index 字节追溯、诊断非确定性、首轮组合编译失败、旧推理修订及数组伪差异。初次证据复制曾仅检查 receipt 存在，误纳入一个 RUNNING 树；其全部 5111 个观测文件已原样移动并核验到 `review-fixes-copy-observations/20261101T113211.067083Z-4xxjavl8/payload`，明确不是完整阶段证据。完成后的原 attempt 仍为 FAIL，另按原路径保留；后续只选择终态阶段回执，最终完整依赖复制包含全部派生消费者原件，无删除或 purge。复制完整事实和哈希亦绑定于索引。

本工单继续 claimed。主 agent 统一维护当前状态，并完成合并后的完整普通/build/vet 与发布只读判定。未执行 live/race/压力；维护受控检查不取得新厂商事实，不宣称 `release.pass=true`。

## Final regression follow-up — 中文评估目录绑定

2026-11-01：根完整回归在首次失败后停止，保留原 `CANCELLED/exit-15` attempt `20261101T114714.995418Z-coufjj5h`，不把中断回执写成完整 FAIL 或 PASS。独立定向 RED `115527.293586Z-dcytz408` 确认默认评估配置在请求/凭据之前被严格拒绝：数据集和任务没有漂移，118 条定义的字段值与三集合顺序保持，变化的是 Provider 合成后的全局 Catalog.Version/Hash，而三份 v2 配置仍固定原 `2026-10-29.1/f44d07c...`。代理 CONNECT 计数为零是此前置拒绝的结果。历史流程原包核验可通过，随后当前计划 helper 失败；不需要放宽历史消费者或资格规则。

实现前补充公开 v1/v2 原字节、v2 配旧快照有效/配新默认拒绝、v3 配旧快照拒绝及当前默认绑定场景，RED `115624.869199Z-1kssw8gq` 保留。新增三份显式 v3 配置及内容版本，只改变配置版本与目录身份，任务、型号、数据集、价格和预算保持；CLI 默认和当前测试选择 v3。v1/v2、原 catalog release snapshot、严格 `Load`、`LoadRecorded` 和历史包核验器均保持原字节，无资格豁免。

代码提交 `3c4aa26c`，随后合并集成文档到 `fc80dded20a1efca329bc5959981187ebf2423e3`。全部 ChineseEvaluation / ClassifierHostMigration 公共场景及 `./ai/examples/chineseeval/...` 辅助编译 GREEN `.evidence/barness-ai-dynamic-catalog/e2e/20261101T115649.314914Z-f96u209c/receipt.json`，PASS/exit0、16.0334 秒、125 个 Go 测试、114 原生 PASS、audit0。配置 JSON 自动纳入 ai 来源，额外输入仍为 49 项；提交/合并后的完整源码与输入集合逐项等于实际受测字节。原 RED 和终态 GREEN 共 625 文件、12,588,246 字节已复制到根工作区并逐文件核验，没有覆盖原件。

原件、argv、测量 HEAD 与提交后来源相等性、旧文件哈希、必需场景及复制库存见 [中文目录消费者不可变索引](../chinese-catalog-consumer-verification-v1.json) 和 [复制回执](../chinese-catalog-consumer-copy-v1.json)。原审查修复索引保持不变。本工单仍 claimed，当前状态由主 agent 维护；本轮未执行付费、race 或压力，也不产生完整发布通过。

## Final regression follow-up — 用量测试的当前目录身份

2026-11-01：完整 failfast 回归形成原生终态 `FAIL/exit1`，原 attempt `20261101T120832.859224Z-9yrkrfox` 保持。用量场景执行到 `builtin-catalog-pinned` 后发现测试仍把当前 `BuiltinCatalog()` 与不可变的旧 release snapshot 比较；真实用量/价格场景与新目录身份正确。独立 RED `122358.133223Z-kune5iyd` 在相同公开子场景复现。failfast 的原生命令完整终止不等于未到达场景已完成或完整普通验证通过。

最小修复仅改 `ai/e2e/usage_test.go`，代码 `aba70f3a9a8564442615f4151bd1abcbcc1c6a27`：当前发现读取 `ai/catalog_builtin.json` 完整 typed Catalog，重新计算并核对声明 hash，再比较公共 BuiltinCatalog 的版本/hash；保留价格变更与非法价格 guard。不写入新的目录身份常量；后续仅数据维护无需改 Go 定义。原 118 定义、价格、顺序和缺省语义仍由独立历史源迁移 E2E 比较，历史 release snapshot 未被替换或更新。

已审查 active ai Go/JSON/Markdown 和 docs 的旧版本、完整旧 hash、历史快照路径及日期型版本断言；逐路径结论见 [引用审计](../catalog-identity-reference-audit-v1.md)。未发现其他未处理的当前旧身份消费者。剩余命中属于三份中文 v2 配置、原 catalog/迁移校验、xAI 原激活事实、按实测提交读取的历史重放、合成历史负例及历史 ADR；现有当前图片/中文 helper 已读取生成数据。没有批量替换、历史字节改写或资格豁免。

GREEN `.evidence/barness-ai-dynamic-catalog/e2e/20261101T122442.177998Z-undt6sjm/receipt.json`，SHA `9bfd13a8ff812e5fe9c0e1a28c79c5d7266d6f38153961a486c0f30a376db8c0`，PASS/exit0、5.6122 秒；94 个实际 Go PASS、89 原生 PASS、audit0。实际集合含完整 UsageAndCost、独立初迁移/历史对照、CatalogSnapshot/OperationCatalogHash、当前中文配置、Google 图片目录、xAI 目录/限额/留存回放；完整测试库存见 [不可变索引](../usage-catalog-consumer-verification-v1.json)。RED/GREEN 原件共 587 文件、18,268,809 字节已终态复制并逐文件核验，[复制库存](../usage-catalog-consumer-copy-v1.json) SHA 为 `d2b6633bcf34caaa949accc7998a2bd296b99432c4ee74ec17d3e54d1288391b`。提交/集成文档合并后的实际源码与受测输入相等。

本工单仍 claimed；完整普通/build/vet 由主 agent 在最终合并和双轴复核后冻结重跑，status.md 不由本实现者编辑。本轮仅测试断言变更，未执行付费、race 或压力，也不声明完整发布通过。
