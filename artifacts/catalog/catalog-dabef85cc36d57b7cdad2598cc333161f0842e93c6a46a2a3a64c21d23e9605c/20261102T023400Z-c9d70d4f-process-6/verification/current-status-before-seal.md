# Kimi K3.1 Pro 维护当前状态

当前阶段：验证已完成；最终交付索引待封存。
六身份及内置已采用；最终完整普通、生成、构建和 vet 全部 PASS。
基准：`ceabec8b58c42377f4f643ccb87c1bc1d4beb456`。
工作批次：`artifacts/catalog/work/20261102T023400Z-c9d70d4f/`。
维护库：`artifacts/catalog/work/store`，原 `.catalog` 和 `local-catalog.json` 保留。

影响：`ai/kimi_catalog.go` 需要接受新 ID；随后维护源/外部快照及对应测试。
用户已明确选择“同步内置，并更新基线”；内置维护源、快照及完整基线已采用。
六条新身份、文本实施子集及独立冻结换算见 spec。
旧 119 条定义、旧 K3 原生捕获和厂商资格不改。

编排：root 负责来源决策、CLI 顺序操作及版本化交付；GPT-6.1 Terra
`terra_builtin_verification` 负责测试先行及受影响验证；`terra_spec_review`
负责中国站及 CFETS 原件研究。

已完成的普通范围：新模型公共 Client RED/GREEN，旧 Kimi 回归、固定候选完整目录合约，
随后完整受影响 ai/e2e、ai、catalog/...、cmd/catalogctl、cmd/cataloggen；
如采用内置，则覆盖完整基线/新维护种子和 ChineseEvaluation 严格 pin。
真实 argv、必需 case、受测提交及输入由验证 agent 固定并留存。

live：NOT_APPLICABLE。本次只声明受控代码兼容，无新增厂商资格主张。
race：NOT_APPLICABLE。最终差异仅显式型号校验、目录数据和只读配置，无共享状态/取消实现变化。
压力：NOT_APPLICABLE。无负载、复杂度或资源边界变化；125 条目录仍在既有容量覆盖内。
普通阶段禁用的压力场景没有执行，不将其跳过计为压力通过。

来源：已取得 models.dev、两站官方文档及 2026-11-02 CFETS 原件，详见工作批次
`research-cn.md`、`inputs/`、`definitions-v1/`。两站 models.dev 原样预览均实际
exit1 / unsupported_definition（video 模态无法映射），没有删字段冒充自动映射。
后续使用官方原件与完整人工定义，逐字段保留确认和依据。

## 已执行事实

核心与新 E2E 已在 `0aede6e89e320b11fe24da5b5b46225186e1d637` 冻结。
产品仅修改 `ai/kimi_catalog.go` 的型号白名单及共用错误提示；依赖和协议不变。
旧程序实际检查为 code_gap / exit1，原件：
`.evidence/catalog-kimi-k31-pro/check-inputs/20261102T024030.660862000Z-2861964047/check.json`。
因此旧 ceabec8b 程序不能只接收新 JSON，必须先升级 SDK / 重建。

新身份公共 RED：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T023726.756749Z-i4b2ixqk/receipt.json`，
FAIL / exit1，六身份被旧型号白名单拒绝；84 个扩权拒绝场景 PASS，audit 0。
改动后的定向 GREEN：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T023906.579411Z-6_t5t1si/receipt.json`，
PASS / exit0，138 个原生场景、audit 0。

旧 K3 定向回归：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T023927.481768Z-jxy1dhxe/receipt.json`，
实际 Go 退出0、2256 原生场景 PASS、audit 0，但运行期间 root 提交导致 HEAD 改变，
collector 保持 INVALID_EVIDENCE；不能作为通过回执。完整源码 2160 项及实际输入
逐项字节相等仅是诊断，不重写原状态。诊断索引在工作批次
`verification/kimi-regression-invalid-diagnostic-v1.json`。
最终完整普通已重新覆盖旧 K3，执行期间没有提交或修改来源。

六条身份已顺序 preview → freeze → check → apply-preview → apply-confirm，
每条实际 check applicable、数字退出0；后续候选均基于上一条实际生效 revision。
最终目录共 125 条（119 chat、3 image、3 classifier），原 119 条定义、价格及
相对顺序相等；moonshot、moonshot-cn 均 v4，OpenAI 仍 v2。机器可读事实及
每次原件引用为工作批次 `audit-final-v1.json` 和 `applied-index-v1.json`。

- Catalog.Version：`catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c`
- Catalog.Hash：`sha256:93a1692f982836d9ad478252313b73e75eeee4904e6d79e3bcc915cd54a44005`
- JSON SHA-256：`6fb2842be21797a6aa448a92244d12b4c528c212d74eff8e16220a883e89818a`
- revision：`sha256:7f0724d5ae599efb2216a22614c775cbce18de88b967b23f48694cd587f8b063`

最终归集目录：
`artifacts/catalog/catalog-dabef85cc36d57b7cdad2598cc333161f0842e93c6a46a2a3a64c21d23e9605c/20261102T023400Z-c9d70d4f-process-6/`。
完整 review 为 34 文件，索引 SHA-256
`8a4a8758cffbcc360144465bde2d120aa6429f046aed612cbfbfb948886c887d`。
前五个版本/过程目录保留中间应用原件；没有完成标记，不可作为本次最终交付分发。
当前尚未写最终 delivery.json，没有发布或新厂商资格结论。

## 内置范围推进

采用前状态原字节保存在工作批次 `status-before-builtin.md`。
外部范围完整普通在 `1827f0956804c37bef8c32594903b895099ee4e5` 启动，
用户明确扩大到内置范围后，由原 collector 受控取消：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T024602.035474Z-3jcz6fs1/receipt.json`，
CANCELLED / 实际退出 -15，部分场景不计为本次完整通过。

Terra 验证 agent 先建立 Builtin 六身份、旧 119 条投影和完整种子来源 E2E，
同时为 ChineseEvaluation v5 当前配置与 v4 历史快照建立 RED。
`terra_spec_review` 仅负责整组默认数据/完整基线采用；
`terra_standards_review` 负责三份 v5 配置和示例 CLI 默认路径；
两者均在 RED 原件固定前不修改产品数据。

重新冻结后的普通范围包含 `./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/...`
以及两个 catalog CLI 包；完整来源/输入、实际 SHA 和执行脚本需新版本，不覆盖
外部范围的旧执行脚本、151 项输入清单或取消回执。

内置/Chinese 公共 RED：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T025153.030653Z-us31ebf4/receipt.json`，
FAIL / exit1，证明原默认目录无六条身份且新 v5 尚不存在。旧 119 项、v3/v4
历史原字节与回放断言未失败。原失败回执保留。

采用命令 `40-adopt-builtin` 实际 PASS / exit0。12 份 Provider、完整 34 文件
baseline 与批准 review 逐字节相同，内置与外部运行 JSON 同为上方完整文件 SHA。
维护 revision 不变；只将批准数据入库，没有重取源或重映射。
三份 ChineseEvaluation v5 仅改内容版本及目录 Version/Hash，示例 CLI 两处默认
路径转向 v5；v1-v4 和实际采用前 119 条快照保留，旧 pin 不会静默接受新目录。

采用后定向 GREEN：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T025230.476314Z-_wb4hku3/receipt.json`，
PASS / exit0，139 个原生场景全部 PASS、audit 0；回执 SHA-256
`d76ec0aea52286a66813fa4daabd4a8a98a0016a3d7ee031bf2d36e36515607b`。
覆盖 Builtin 六身份、请求/用量/费用/推理/授权、完整来源/新种子重导出，
以及 ChineseEvaluation 当前与历史 pin。该受测工作树已与 2194ca51 的 2166 项
来源和80项输入逐项确认相等；不只看当时仍为 1827f095 的 Git HEAD。

第二轮完整普通使用独立 `run-ordinary-v2.py`、`ordinary-inputs-v2.json`
（165 项）和 `required-cases-v2.json`（202 项），三者均在工作批次 verification/。
执行源码及输入重新冻结后，保持 HEAD 和受测文件不变，顺序完成普通、生成、构建/vet。

在 `2194ca515a3188645b3a0d49852ef99839a10dea` 的内置完整普通发现
`TestKimiQualificationPromotionBoundary` 实际断言失败，随后按 SOP 停止：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T025343.292240Z-w42x02ym/receipt.json`，
CANCELLED / 实际退出 -15。观察到的失败仍保留，不以取消隐藏；未启动后续构建/vet。

该历史资格测试已逐条确认旧 K3 的六份运行定义与原 payload 相等，但随后把两个
Provider 的全部型号数固定为 6；当前两站共有 K3 与 K3.1 Pro 的 12 个身份。
仅将该历史计数限定 `ID == "kimi-k3"`，继续要求恰好六条，不改原 payload、资格、
任何产品定义或新六身份合同。修订后先定向复验，再冻结并重新执行完整普通。
后续必需原生集合需显式加入 `KIMI-K16-promotion-proof-boundary`，避免漏验该失败项。

修订已冻结于 `97e3fef86d6948705eb70abc97976c64c725b442`；相对 2194ca51
只改该历史测试及状态文档，没有产品或数据变化。两项增量审查均确认历史资格边界
保持。定向复验
`.evidence/catalog-kimi-k31-pro/e2e/20261102T030650.435730Z-1xea9ijc/receipt.json`
实际 PASS / exit0，188 个原生场景全部 PASS、audit 0，原失败项已实际 PASS。

完整普通 v3 回执：
`.evidence/catalog-kimi-k31-pro/e2e/20261102T030707.838552Z-x6dz4dj5/receipt.json`。
实际来源固定为 97e3fef8，使用不可变 `run-ordinary-v3.py`、165 项
`ordinary-inputs-v3.json` 和 203 项 `required-cases-v3.json`；最终 PASS / exit0。
完整普通期间不再提交或更改受测来源/输入；状态 Markdown 不属于本轮实际输入。

## 最终验证结果

实际受测提交：`97e3fef86d6948705eb70abc97976c64c725b442`。
完整普通 v3 实际 PASS / exit0，墙钟 1304.86 秒。主 bundle 为
`bundles/20261102T030710.087323000Z/`，17,124 PASS、226 NOT_RUN，audit 发现0。
203 个必需原生场景全部实际 PASS，包含原失败项 KIMI-K16；三份离线受控
live-tag 辅助 bundle 分别 32/77/6 PASS、audit 0，不产生厂商实测资格。

普通 receipt SHA-256：
`7aab50187f44c02a2091d062271aa495664c7a3969ae5062a9845668e9a9fa9a`。
完整受测源码 2166 文件、实际 selected inputs 167 项（165 项清单加执行脚本和清单本身）；
完整文件集合与每项摘要由实际回执固定。运行期间 HEAD 与受测字节稳定。

按顺序执行的 `go generate ./ai`、
`go build ./ai/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen` 和相同包范围
`go vet` 均 exit0。原命令、工具与日志在工作批次 `verification/build-v1/`，
唯一脚本为 `verification/run-build-v1.py`。

验证机器索引：工作批次 `verification/result-v1.json`，SHA-256
`2dbde0d1756f9b4382ccd51f42f23aa0ebc6729bd2a67f267e9b2ec6036dbf8a`。
该索引核验六次 CLI 的 status、实际退出、全部 required cases、manifest 和 audit；
最后一次检查的身份集合与最终 125 条快照完全相等。CLI 实测 0aede6e 到最终 97e3
的 12 项实际源码差异逐项记录；核心 Go 与运行依赖没有再变，不能声称全部源码相等。

跳过完整清单：工作批次 `verification/skip-inventory-v1.json`，SHA-256
`5ffd7747cb28875c08542ba1e3acce2350198530ab416172780bec8116bc352c`。
Go skip 252 项：226 压力 opt-in、16 仅国际站 MiniMax 历史账本、1 可选历史包诊断、
1 循环压力 opt-in、8 无测试包。原生 NOT_RUN 226 项全部为压力 opt-in；
203 必需场景没有跳过。Go 事件与原生场景属于两套库存，不能互相替代或相加。

核心代码相对 ceabec8b 确实变化：`ai/kimi_catalog.go` 新型号白名单；依赖不变。
ChineseEvaluation 示例默认 CLI 路径属于其他代码变化，维护源/基线/运行 JSON 属数据。
旧程序实际第一候选（国际 Chat、新增一条后的120条快照）检查 code_gap/exit1；
该检查不是对最终125条文件的旧程序实测。型号校验代码亦说明旧程序不能使用
K3.1 Pro，须升级 SDK 并重建。新程序对最终125条原字节检查已通过；
使用编译内置默认目录的程序同样须重建和分发。Host 仍须显式授权新 ID。

本次没有软件发布、部署、新厂商资格或实际 live/race/压力执行结论。

文档收尾后的来源适用性核验已 PASS：工作批次
`verification/post-doc-equality-v1.json`，SHA-256
`73a5e91abc82008a91e5efb58d372912cd2c5477090ac90108d5fc2f01d0a278`。
完整 2166 项源码和167项实际输入与普通受测回执逐项相等；此时 HEAD 仍为97e3fef8。
