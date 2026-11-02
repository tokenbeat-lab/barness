# Kimi K3.1 Pro 维护当前状态

状态：六身份已检查并应用；内置定向 GREEN 完成，完整普通发现旧资格计数范围问题，待修订后重验。
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

普通计划：新模型公共 Client RED/GREEN，旧 Kimi 回归、固定候选完整目录合约，
随后完整受影响 ai/e2e、ai、catalog/...、cmd/catalogctl、cmd/cataloggen；
如采用内置，则覆盖完整基线/新维护种子和 ChineseEvaluation 严格 pin。
真实 argv、必需 case、受测提交及输入由验证 agent 固定并留存。

live：NOT_APPLICABLE（计划）。本次只声明受控代码兼容，无新增厂商资格主张。
race：NOT_APPLICABLE（计划）。仅显式型号校验和目录数据，无共享状态/取消变化。
压力：NOT_APPLICABLE（计划）。无负载、复杂度或资源边界变化。
各项仍须依最终差异复核；适用但未执行项应写 NOT_RUN。

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
最终完整普通会重新覆盖旧 K3，执行期间不再提交或修改来源。

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

重新冻结后普通计划包含 `./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/...`
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
以及 ChineseEvaluation 当前与历史 pin。受测工作树实际文件需与最终冻结来源
逐项确认；不只看当时仍为 1827f095 的 Git HEAD。

下一轮完整普通使用独立 `run-ordinary-v2.py`、`ordinary-inputs-v2.json`
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
