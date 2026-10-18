# 14 — 修复证据映射 hash 与付费前置契约预检

Type: task
Status: claimed
Updated: 2026-10-18

## 问题与边界

票13实际 `8fbd170` 完整 ordinary 结果保留12448 cases、12439 PASS、8 FAIL、1 NOT_RUN和audit0，其中 `MM10-release-contract`、`P0-GLM13-release-trace-complete` 报“approved evidence mappings changed”。`ad20469` 为区分实际pi差分与目录元数据而完成七项元数据case改名，trace已更新，固定映射hash未同步。

实现包含两处有限修复：同步 `ai/internal/testkit/release/plan.go` 的 `approvedMappingHash`，以及在 routed paid preflight 增加现有公开契约检查。前者在独立复核七项改名且无必需集合缩水后，固定为 `f1de4ce6b8a81f08b0ce27c56bc01f6c8fae9e8a8ca51c114c44d5a55acd21b6`。不改trace、selector、required cases、断言、来源映射、importer、预算或捕获。

## 已约定公开测试边界

沿用 `TestMiniMaxAdditiveReleaseContract` 与 `TestGLMReleaseTraceContract` 的现有public CLI / release contract E2E，先在worker `a3a9097` 保存RED，再做一行修复，复跑相同ordinary与race及双vet。GLM完整/缺requirement/缺acceptance/缺story/缺required-case模式全部保留，未新增内部测试。

失败清单：基线仍拒绝当前已审查映射；错误接受删除requirement/acceptance/story/required case；减少自身17身份/固定8差分/有限MiniMax集合；替换了非此常量；把旧A2失败包改写为成功；把post-fix不再适用的来源pin刷新为当前；擅自发起新付费轮，均不可接受。

## Comments

- 此修复与已冻结的票13实际A2分开；A2当时8 FAIL保持原样。当前只在独立worker工作树修复，不在A2 race期间改primary HEAD、源码或输入。
- 无Key读取、无API请求、无新付费授权。凡选中 `plan.go` 的已有证据，修复后按原始source hash自然过期；具体影响清单保留后续审查结果，不刷新旧pin。

## 预检防回归

新增范围仅在现有 routed `check` 追加一次固定无Key公开契约命令：`go test ./ai/e2e -count=1 -run '^(TestMiniMaxAdditiveReleaseContract|TestGLMReleaseTraceContract)$'`，含GLM四类删除/缺失负向模式。与双vet、no-key-live及visual-proxy合计五项数值退出码；任一失败写入checks=false并拒绝paid intent，成功不自动授予授权。旧已冻结checks与已关闭round不改，新round也未创建。

公开CLI失败先行已证明旧预检仅四项、从未启动这两项契约测试，且控制契约失败时旧check错误返回通过。临时fixture凭据路径为目录；追加后须证明真实调用该命令、无凭据继承/读取，失败时即使提供独立精确source review和checks hash，仍无native paid command或canonical intent。

## 来源适用性与后续合同

独立诊断 `/tmp/barness-routed-release-plan-gap.json` 固定当前基线8fbd；旧包和原消费545/545保持原样。此修复令完整1047来源freeze过期；GLM217与M3各263的已选native来源不含plan.go或Python工具，因此自身映射不改，GLM Responses实际FAIL也不改变。两路visual323含plan.go；protocol114及派生Jev的861来源含plan.go，因此对修复树过期。两路中文原始860完整snapshot不改；现有261执行选择排除release工具，因此此限定修复不改变既有中文执行适用性。

十项generic live native没有Go执行来源映射；原读入策略接受非空measured commit并不构成1047完整测量的新鲜度例外。后续若要求当前完整source验收，按独立审查的保守解释需新固定3 GLM + 完整15 L合同，现有单进程上限合计374 calls /398 Attempts；这是未来提案，不是现在的授权。不借用旧未消费上限、不刷新旧hash、不剪裁15项、不创建新round工具或paid intent。M3现有native适用则不重跑，M Plan仍NOT_RUN。

## 修复证据留存

代码提交 `c2604db6` 完成公开Go契约ordinary/race各6 PASS、公开CLI RED/GREEN、工具60项及双vet。c260的实际无Key预检在补齐6个selector固定的历史ZIP后五项均0；首次缺ZIP的12 FAIL完整保留，未混入真正代码RED或覆盖历史结果。模板保持PENDING，原18 canonical hash前后不变，无新round/paid intent。固定原始归档、逐项hash、27,412个artifact引用校验及14个拒绝反例见[票14无Key修复留存](../evidence/ticket-14/repair/README.md)。源代码审查锚点仍为c260，留存文档提交不刷新原始capture来源。
