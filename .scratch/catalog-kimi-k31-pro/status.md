# Kimi K3.1 Pro 维护当前状态

状态：进行中，已确认两站，尚未应用或交付。
基准：`ceabec8b58c42377f4f643ccb87c1bc1d4beb456`。
工作批次：`artifacts/catalog/work/20261102T023400Z-c9d70d4f/`。
维护库：`artifacts/catalog/work/store`，原 `.catalog` 和 `local-catalog.json` 保留。

影响：`ai/kimi_catalog.go` 需要接受新 ID；随后维护源/外部快照及对应测试。
内置采用等待本次用户选择。六条新身份、文本实施子集及独立冻结换算见 spec。
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
`research-cn.md` 与 `inputs/`。尚无本次 PASS 或发布结论。
