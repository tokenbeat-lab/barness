# MiniMax固定单项计费与终态封存修复

先修finishMiniMax的single/full后验拒绝收尾，使最终FAIL及audit写入后生成最终seal，之后不再修改被封存数据。再仅对国际Messages/M3/standard、无cache_control的固定单项允许creation计数缺失，独立写费NOT_APPLICABLE；raw缺失与UsagePartial保留，禁止补零或修改全局SDK/产品usage。null/正写/TTL/显式cache_control、其他计费项缺失与其他身份仍拒绝。

实现前失败清单：真实受控进程single/full后验失败时seal必须与最终报告一致；失败写入/库存/封存保持拒绝；M3创建计数省略与显式0通过，保持missing/partial；null/正写/TTL/其他计数缺失/终态缺失/cache_control/错误身份拒绝。只在可信私有Git common域构造controlled授权，写前断言，不接触真实权限。

架构：只复用现有MiniMax原生校验、grant/ledger与CLI；写费适用性为当前验收输出，不修改原测量。新输出应明确cacheWriteFee=NOT_APPLICABLE与原UsageReporting；无新通用框架。新旧来源分别固定，不自动牵动其他Provider。

验证：先公共E2E RED后实现GREEN；完整八包普通范围ai/e2e、ai、release、supportmatrix、minimaxlive、livegrant、jevformal、visualeval，scoped正常vet及live-tag compile-only vet。不启动真实live/race/压力、不激活目录、不恢复暂停。保持工作区静止直至完整普通回执完成。

离线旧捕获判定：只读回放两个FAIL/exit1的thinking-on wire及预算/资源/原授权与producer证据，明确语义适用与完整资格的区别。原seal与最终report不一致、原producer exit1不改；不新增历史兼容入口或例外来凑13+1通过。能作为回归复用则记录，不能用于完整资格则列出原件/合同缺口与最小补项。所有历史证据、授权和8162个公共账本文件保持。
