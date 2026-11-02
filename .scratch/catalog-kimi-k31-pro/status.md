# Kimi K3.1 Pro 维护当前状态

状态：六身份已检查并应用；最终普通验证与交付归集待完成。
基准：`ceabec8b58c42377f4f643ccb87c1bc1d4beb456`。
工作批次：`artifacts/catalog/work/20261102T023400Z-c9d70d4f/`。
维护库：`artifacts/catalog/work/store`，原 `.catalog` 和 `local-catalog.json` 保留。

影响：`ai/kimi_catalog.go` 需要接受新 ID；随后维护源/外部快照及对应测试。
本次先完成外部 JSON；尚未采用内置维护源、快照或基线。
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
