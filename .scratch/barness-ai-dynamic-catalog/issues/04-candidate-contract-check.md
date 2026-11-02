# 受控公共协议合约与候选检查应用

Type: task
Status: resolved
Blocked by: 01, 02

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责候选 check 的真实进程执行和严格证据消费、CLI check/apply 闭环及每身份公共协议合约。每条最终四维身份都必须有适用合约和实际必需子场景，验证请求/响应、Usage/价格以及声明能力。保留实际代码/冻结来源规则/候选/基准/版本、命令、退出和不可变原件；apply 重新核对来源集合与证据并写原受测快照，不重新取源或重映射。不依赖 ai/internal/testkit 的生产 API，不调用真实厂商。缺能力时明确 code_gap，不放宽 Kimi/xAI 固定约束。

## Acceptance and verification

T05、T06、T11、T12、T13；CLI → 真实受控检查 → apply → 加载 → Client，含失败、取消、超时、篡改和旧证据拒绝。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。

2026-11-01：候选/存储接口已明确，可与 03 并行实施 runner/证据消费；本工单完成前仍须合并 03 并验证完整 check/apply CLI 闭环。设计笔记：`/tmp/barness-catalog-check-notes.md`。

## Failure modes and verification plan

公共边界为编译后的 catalogctl → check/apply 子进程 → 本地 Load → Client。首次红场景验证 check help/dry-run 无写入；随后逐片验证：未确认；定义失败/代码缺口/执行失败；真实退出非零、零匹配、仅父测试通过、自填 PASS；遗漏身份/必需场景；候选/基准/版本/实际源码及输入集合漂移；manifest/audit/标准输出/原件缺失或被改写；源不可再取但原受测快照仍可应用；输出路径与保存失败；受控取消和超时保留失败原件且不推进目录。所有 118 条最终身份都通过最终 Config.Catalog 实际调用，声明能力、Usage 和严格阶梯边界独立核对。

普通执行使用 run.py 的独立 attempt；新增 catalog/cmd 文件逐个 --input。测试实现先红后绿。无真实厂商、race 或压力；根 status.md 由协调者维护。

### Intermediate public flow

- 2026-11-01：候选固定 runner 与严格证据消费者、check/apply 初始闭环完成；工单仍 claimed，后续补充非零缓存与拒绝场景，不宣称最终覆盖。
- 首次 CLI 红：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T094833.374763Z-bmu163fy/receipt.json`，FAIL/exit1（无维护 CLI）。初始逐身份 runner：`20261101T095202.144349Z-2fgyn023`，PASS/exit0、909 原生 case。
- 编译后 check/apply 闭环：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T095708.310357Z-wkx7id6v/receipt.json`，PASS/exit0。内部受控检查有独立 attempt、完整来源字节及最终候选，外层只消费其实际退出和原件。
- 流程偏差：最初红场景覆盖缺少 check CLI，随后同一实现片段写了 apply 存储，再添加专用 apply E2E；专用 apply RED 未先于该段代码。此事实保留，不称其符合逐片先红顺序。后续真实失败（消费者结构/遗漏 GLM 场景）及修复均独立留存；没有补写实现镜像单元测试。

### Expanded contract and refusal receipts

- 非零缓存：`20261101T095844.033570Z-d0k9xz4w` FAIL/exit1 揭示原有未定价/不支持写缓存保护（未修改产品）；`20261101T100118.972704Z-94pzuslm` PASS/exit0 验证可适用读/写价格与原有拒绝，Kimi Chat 未报告 TTL 保留独立上下界。
- 完整必需场景新增后消费者拒绝旧计划集合：`20261101T100155.769384Z-ssfmw9yd` FAIL/exit1；消费者相应固定新增场景，不能以旧通过覆盖新要求。
- apply 拒绝：`20261101T100343.799751Z-kubqs_1s` FAIL/exit1（发现输出可覆盖维护对象/证据及摘要可重绑缺陷）；修复后 `20261101T100504.776405Z-17n026_k` PASS/exit0，含源码新增/内容变化、输出保存失败与候选重绑。
- 原生证据消费者：`20261101T100631.351325Z-ctm10np2` PASS/exit0。派生夹具明确标为 synthetic/derived；原件不修改。父通过、缺身份计划、缺合约、错误原生退出和缺原件均拒绝。
- 取消/超时：`20261101T100833.396214Z-7eue2qoz` FAIL/exit1 揭示超时额外30秒；当前路径失败 `20261101T101045.072851Z-a52gq72d` FAIL/exit1 揭示错误结果误标 applied。修复后 `20261101T101147.103430Z-glgcvkry` PASS/exit0，含实际新数据接纳/未知API code_gap/来源删除后原字节应用及受控 SIGTERM/超时。当前路径错误在预检阶段拒绝，已有外部输出保持；不为推测性后置故障引入 hook。
- 自定义尺寸：`20261101T101238.923915Z-4ywmwbks` FAIL/exit1 揭示能力已声明但未实际发送；`20261101T101501.342723Z-eqs4a1ui` PASS/exit0 验证固定512x512、数量2的独立 wire 预期及自定义边界/图片选项传递。

所有目录均位于 `.evidence/barness-ai-dynamic-catalog/e2e/<attempt>/receipt.json`，保留实际来源/输入、真实退出及原生清单；最新集成来源与完整受影响普通验证由协调者收尾。

- 显式状态查询：`20261101T101645.863881Z-xpml1qua` FAIL/exit1（原 candidate 查询无 check/ref 合同），`20261101T101749.657878Z-r2p3b3kc` PASS/exit0。`candidate --check <check.json> --repo <root>` 对 applicable 重做严格验证；真实失败引用显示 code_gap 等独立状态，自报 PASS 不取得可应用状态。不引入可变候选状态或 floating latest。

## Answer

已实现固定候选公共合约 runner、严格原生证据消费、受测源码/输入原件冻结、check/apply 与显式状态查询。所有生产入口无 ai/internal/testkit 依赖；测试仅用受控 Provider，无真实模型 API、race 或压力。完整接口和使用说明见当前集成文档与协调者状态入口。

组合普通回执：[候选验证索引](../candidate-verification-20261101.json)。实际 argv 为 `go test -json ./ai/e2e ./ai ./catalog/... ./cmd/... -run 'TestCatalogCandidate|TestLocalCatalog|TestMaintenanceCatalogMigration|TestCatalogMaintenance' -count=1 -timeout=20m`；受测提交 `370842791e45e34da8f009ac829f5a4f745f0b1e`；PASS/exit0、1208 完成 Go 测试、1086 原生 case、118 身份/1020 候选子场景，107.84秒。原件 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T102025.819940Z-i1e7_odi/`，manifest/audit 与回执 SHA 见索引。收尾仅本工单/索引文档，实际源码及37项输入集合/字节与受测回执相等。

当前路径错误由预检拒绝并保持既有输出；current 原子 rename 是提交点。外部输出先准备，若 current 安装失败只做一次有界恢复；底层文件系统恢复失败时返回 storage_failed 并要求修复，不承诺跨文件系统事务或目录 fsync 失败后的无限恢复。未引入推测性故障 hook。

完整集成回归、双轴审查与人工阶段判定由工单05继续；本工单的普通通过不等于厂商实测或完整发布通过。所有失败、派生夹具、来源原件及内部/外层回执保留；没有重写旧失败。

### Last active model withdrawal follow-up

根审查发现最终三集合全空时，原 runner 和消费者均拒绝没有身份的计划，但最后一个有效型号的明确撤回应可生成空运行目录。以独立编译的一条 active 型号的 synthetic reviewed source 为启动默认，通过公开 CLI withdraw → check → apply 验证；未直接构造生效 store，未逐一循环撤回118条，未修改根默认目录或03辅助函数。

- `20261101T103545.758748Z-m4q3ik9z` FAIL/exit1：有效空候选被固定 runner 的非空身份计划条件拒绝。
- `20261101T103647.961395Z-xt67p6u_` FAIL/exit1：新增真实 empty child 已通过，原消费者仍拒绝零身份。
- `20261101T103731.654684Z-hof541ks` PASS/exit0：固定全局 `CC-empty-catalog` 合约与真实 check/apply 闭环通过。合约验证完整 Load 空报告、三集合及型号/列表查询为空、三个操作的已授权但目录不存在目标在 capability 阶段拒绝且零 Provider 请求。父通过、缩短计划或缺失 empty 原生 case 的派生夹具均拒绝。

消费者只对实际三集合全空要求精确的全局 empty child，非空候选仍要求全部最终四维身份及必需场景。最终冻结普通验证见 [空目录补充验证索引](../candidate-empty-verification-20261101.json)：受测提交 `229419e6017ce6953ebe8de8494ff27d54d06152`，argv `go test -json ./ai/e2e ./ai ./catalog/... ./cmd/... -run 'TestCatalogCandidateContracts|TestCatalogCandidateLastWithdrawalToEmpty|TestCatalogCandidateEvidenceRefusals' -count=1 -timeout=5m`；PASS/exit0、1141项 Go 测试、1023个外层原生 case，保留默认118身份/1020子场景；内部空候选单独精确包含 plan 与 empty 两个原生 case。外层回执 `20261101T103807.098818Z-omyldv6a`，耗时30.45秒，audit findings为空。收尾仅增加本工单及新索引，实际源码及42项输入集合/字节与受测回执相等。未执行 live、race 或压力；完整集成冻结由协调者更新。

### Fixed entry and failed-output recovery follow-up

根审查指出两项具体缺陷，本工单按公共进程边界补充真实 RED，再修复并冻结验证。

- 固定入口：原消费者只比较 prepared、command、receipt 内互相引用的 argv。新 E2E 先验证完整迁移的 coherent fixture 可消费，再同步更新三份 argv、prepared/command/selected-input 库存并保留真实 native PASS 原件。`20261101T105738.238872Z-8j2ndtcf` FAIL/exit1，证明错误 regex、错误 executable、1201秒超界及 schema2 均被原消费者错误接纳。现在 prepare 与 inspect 共用小型 `fixed_argv(directory, timeout)`，独立重建固定 Go test 入口；inspect 要求 schema 为整数1、timeout 为1..1200整数，拒绝不一致的固定入口。`20261101T105811.878255Z-7ss3o5dv` PASS/exit0。此前无用 import 的编译失败 `20261101T105655.374926Z-9dqngtrx` 另行保留，不冒充行为 RED。
- 输出失败后的重试：旧 Revision.ID 只签 State，但保存正文另含 CandidateID/CheckReference。真实 output 保存失败后留下 orphan；同候选重验到新 check 或同 State/不同 reason 的新候选，再用有效输出 apply，均在 `20261101T105942.218368Z-n_0jrnae` FAIL/exit1 中发生 object_conflict。现在 revisionDigest 签署 ID 为空的完整不可变 Revision（parent/candidate/check/state/snapshot/change），seed、commit 和 read 使用同一身份；删除旧 state-only 内部 helper，无兼容 shim。`20261101T110032.729195Z-zwhy1kxu` PASS/exit0 验证两种重试、原 orphan 字节不变、只追加成功应用的 history、输出等于原受测快照，并通过公共 export 拒绝修改后的 CheckReference。运行 Catalog 的版本/hash 仍独立于应用身份。
- 扩大验证 `20261101T110122.390904Z-j8_36dk9` FAIL/exit1：较早03 peer 的 review export 测试使用嵌套 ev.Fixture 文件名，不能写出 `review/providers/...` 原件；新缺陷场景均已通过。03最终分支已有修复，随后合并最新集成 f22b54f5，未在04重写03 helper。该失败保留。

最终冻结补充验证见 [恢复与固定入口验证索引](../candidate-recovery-verification-20261101.json)：受测提交 `2377f44e23007062a29e7c08fff518dd14bda9a0`，精确 argv/场景/来源/输入/原件见索引。PASS/exit0，1150项 Go 测试、1031个原生 case、默认118身份/1020子场景，169.46秒，audit findings为空；含新固定入口/空目录/原生消费者/两种 retry、03 lifecycle/added rollback/revision integrity/review export/cost update，以及辅助包编译。外层回执 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T110625.007719Z-bgminfx4/receipt.json`，SHA `eccb14283082d9633a203756299435cef4329341df60446694441e03e390b2b1`。收尾仅本工单及新索引文档，实际2142个来源文件及42项输入集合/字节与受测回执相等。所有历史失败和源原件保留；没有删除 orphan 以取得通过。完整集成回归及双轴审查仍由根协调者进行，本补充不执行 live、race 或压力。
