# Ticket05 Host/shared checkpoint

已实现六条独立 Kimi Host 路线、标准 API 用途、独立凭据元数据和读秘密前校验、四入口/Hook wrapper、精确官方 URL/可信 gateway 校验、本地/云端装配与预算/回滚说明。共享范围同时修复未证实的 request-ID 头、推理与 total 包含关系、有限价格乘法溢出，均通过公共 Client 先红后绿。

基线/受测 Git 提交 `3e25c0e` 加本票工作区变更；实际受测文件字节由各 receipt 的 source.files 绑定。后续提交只记录这些字节与非运行时 scratch 说明。集成后的最终六路线/完整相关包验收由票07绑定，当前 checkpoint 不声明完整发布。候选价格/FX 完整来源见 `ai/internal/testkit/kimi/`，不重复复制原件。

## Retained attempts

主工作区原件位置 `.evidence/barness-ai-kimi-host/e2e/<attempt>/receipt.json`；上述树原样 copy2/copytree 保留，内部 bundle 路径不改写。所有其他失败/运行中断 attempt 同样保留。

- `20261022T141820.716458Z-tvamowlc`: FAIL, exit `1`, receipt SHA-256 `99359717f5f9175b7fe7294a03d316d31bd135672f32d6467567981ba0dda8fb`; argv `["go", "test", "-json", "./ai/e2e", "-run", "^TestKimiHostOfficialHostnameCannotBeGateway$", "-count=1"]`.
- `20261022T141901.073412Z-0j2d_3pf`: FAIL, exit `1`, receipt SHA-256 `a7be7a5a3740db9531ffdf2cc7bd199dcea705dc5734aaf6b1bb107649352c4a`; argv `["go", "test", "-json", "./ai/e2e", "-run", "^TestKimiHostCredentialOwnership$", "-count=1"]`.
- `20261022T142259.431623Z-p8t6dbhk`: FAIL, exit `1`, receipt SHA-256 `cc7ac97754308e7a01ce39043922f3404db2d7bf9924cb0b9d35154ed9b9ba10`; argv `["go", "test", "-json", "./ai/e2e", "-run", "^TestKimiHostUnverifiedRequestID$", "-count=1"]`.
- `20261022T142650.925245Z-rqs8ahs2`: FAIL, exit `1`, receipt SHA-256 `aed3de5d6806b691823b890e624abd5aef8460fe6604a29dab9e68a5ec949792`; argv `["go", "test", "-json", "./ai/e2e", "-run", "^TestKimiHostReasoningTotalContainment$", "-count=1"]`.
- `20261022T143010.040355Z-xgpzu4_g`: FAIL, exit `1`, receipt SHA-256 `cde085ddaf5fb10666184c3df88cff7b28a993a764318e1f733e992459c08265`; argv `["go", "test", "-json", "./ai/e2e", "-run", "^TestKimiHostCostOverflowPreservesLegalUsage$", "-count=1"]`.
- `20261022T143123.495796Z-07g1fzw3`: PASS, exit `0`, receipt SHA-256 `5c3840894395f5f6fc3e9b0ddff43d44d1798527c965c12fb5a59afee7cdc438`; argv `["go", "test", "-json", "./ai/e2e", "./ai/examples/kimihost", "./ai/examples/internal/hostconfig", "-run", "^Test(KimiHost(Configuration|CredentialOwnership|OfficialHostnameCannotBeGateway|UnverifiedRequestID|ReasoningTotalContainment|CostOverflowPreservesLegalUsage|DisableAndRollback)|RegionalProviderHost|KimiUsagePricingBoundaries|KimiChatUsage|KimiMessagesUsage)", "-count=1"]`.
- `20261022T143202.025308Z-2pcw1iwm`: PASS, exit `0`, receipt SHA-256 `e37db653bfdcb1acf3479e91f11590fb9fd1a9601420b504e859342d560cfe73`; argv `["go", "test", "-json", "./ai/e2e", "-run", "^TestKimiHostRoutes/[^/]+/(0|2)/", "-count=1"]`.

最新常规 scope 覆盖 Host 授权/元数据72拒绝、官方目标、gateway、shared request-ID、共享成本边界、既有四 Provider Host 回归和 Chat/Messages usage；另一个最新 scope 覆盖2市场×2已合并协议×4入口，16个实际子用例。无厂商网络请求。

Responses wrapper 已有独立红例 `20261022T142225.769705Z-czh_x1bs`，只缺票03协议成本接线；保留票05 claimed，等集成六路线 wrapper 转绿后完成。live/race/压力均 NOT_RUN，未授予正式支持，`release.pass=false`。恢复方法是停用 Kimi Binding、从加载目录移除 Kimi 身份，再恢复之前软件；不改写历史或其他 Provider。

## Concurrent multi-tenant/dual-market acceptance

追加公共 Client E2E `TestKimiHostConcurrentMarketsAndTenants`：一个 Client 同时持有两租户×两市场×Chat/Messages 八条调用；同账户标签/同 Session 下四份独立凭据，header 和 socket channel 屏障证明八个许可同时持有。逐调用校对归属/封套、实际 key alias/路径、四个无原始标识的不同派生 cache key、独立输入/Result/CostEstimate/Observer 快照、字面 worked USD 范围与资源归零。未发现需要修改产品代码的问题；未运行 race/压力/厂商请求。

最终本片来源为提交 `e0ac913` 加此验收文件；冻结文件字节见 receipt.source.files。普通命令 `go test -json ./ai/e2e -run '^TestKimiHostConcurrentMarketsAndTenants$' -count=1`，exit0，PASS；原件 `.evidence/barness-ai-kimi-host/e2e/20261022T144043.784126Z-41wvzk80/receipt.json`，SHA-256 `43d3f323ef3a8a7eee478f6aa5f8bea5cee9c44c45086921d3c91bcc02eaabbd`。编译修正前的两个失败 attempt 和初版通过同样原样保留。票05仍等待票03合并后的六路线完整 scope。

## Exact hook number investigation (no product defect)

公共 Client 验收覆盖两市场 Chat/Messages 的 json.Number 与 RawMessage：分数 budget、近似固定 top_p 均在零 Attempt/HTTP 之前拒绝；合法精确数值可用，model 越权仍先报 tenant_denied。`marshalJS` 当前直接使用 json.Encoder，不存在猜测中的 float64 round-trip；产品代码无需改动。既有 OpenAI hook 同样保留原始数字拼写。

初始失败 `20261022T145028.378654Z-argp7a2w` 只来自本次测试对既有 OpenAI 数值归一化的错误期待，属于测试设置错误，不作为产品缺陷红例。修正期待后的普通验收 `20261022T145208.513160Z-ga07rikd`：exit0/PASS，命令 `go test -json ./ai/e2e -run '^TestKimiHostHook(ExactNumbers|LegacyNumberPreservation)' -count=1`；receipt SHA-256 `9787a481763d4ef74964a18f3c67de834b3bc14cfd40d6b5176797fc261e5d06`。原件已复制至主工作区 `.evidence/barness-ai-kimi-host/e2e/`。后续03合并后增加 Responses 同组验收。
