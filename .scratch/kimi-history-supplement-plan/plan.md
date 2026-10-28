# Kimi 两站六路线：只补 xAI 来源历史的采集与组合消费方案

本轮交付设计与离线原件核验，不实施采集器/消费者，也不发放执行授权。建议保留各路线原九项，只为每条增加一次 `kimi-k3/xai-origin-history`；满足下述付费前置条件后，新增上限为 **6 次逻辑调用、6 次 HTTP、0 重试、USD 0.60**。这是新批准范围的提案，不是已有权限。

## 1. 已核实事实与前置条件

六份原 bundle 的 54 个场景全部 PASS，逐场景 artifacts、seal、原报告、授权/命令/外层回执及账本证明引用重新核对完整；实际使用的型号投影与当前候选目录相同。原报告均为 `mode=live, stage=candidate`，预算各为 9 calls / 9 attempts、36,864 输出、131,072 输入上限，不能改成 10 或挪用旧余额。六份精确路径、manifest/report/seal SHA 和原引用见 [路线证据与预算](../../.evidence/kimi-history-supplement-plan/route-evidence-and-budget-v2.json)。

但“原九项捕获有效”与“当前产品来源适用”是两项事实。前次保守来源比较对 Chat/Messages 各列 11 文件、Responses 各列 19 文件；根协议入口范围已修正，部分差异可缩小，仍须针对 Kimi 自身路径审阅，不能照抄 TypeSafe/GLM 的放行条目。**付费前先完成逐项差异审阅、54 项旧捕获的当前公共 Client 请求重建/响应回放及原生语义复核，取得每路线明确的九项适用性结论。** 原生 history/tool-result 必须从该 bundle 的固定实测前序和租户/账户封套恢复；取消保留已观察前缀及原时序边界，不能当成完整成功响应。

未知产品变化、缺原字节、实际语义不等价、SDK/型号能力或价格不适用时，停止该路线的授权准备，单列实际受影响场景；不自动补整组九项或其他 Provider。故“6 次即可补齐”以六份九项证据均适用为前提，目前没有宣称完整当前资格已通过。

当前 `ai/catalog_kimi.json` 仍 `active=false`。候选证据补齐、正式目录数据审阅/激活、当前正式资格、整体发布分别处理；本方案不激活目录。

## 2. 只采一个固定场景

复用 [grokhistory.Request / Verify](../../ai/internal/testkit/grokhistory/history.go)、[wire 校验](../../ai/internal/testkit/grokhistory/wire.go) 和地区 `regionalSession.execute`：固定两个 xAI assistant turn、同轮易碰撞 ID、另一轮重复原身份、三个预置工具结果。Host 历史为合成夹具，不请求 xAI，不实际执行工具；新增真实事实是 Kimi 本路线对转换后历史请求的接受、返回、归属与用量。

拟新增精确入口 `^TestKimiHistorySupplement$`，每进程只运行 `kimi-k3/xai-origin-history`，无前置厂商调用。选择由新 grant 的固定 purpose 推导；不增加任意场景列表环境变量，不用 `-test.run` 子测试过滤绕过 `TestRegionalCandidate` 的完整十项约束。

新采集报告保留真实 `candidate/live` 身份、单项 Expected/Scenarios 和真实时间/来源。`regional.complete` 保持 false（未独自覆盖十项）；由专用单项原件校验从 actual case/Attempt/exit 推导“所选单项完成”，不会写入九个虚构的 NOT_RUN/PASS，也不产生单包正式 `live-report.json`。

必须捕获并独立检查：

- request / immutable-input、history-provenance：输入和两个原可信来源封套前后不变；当前固定夹具、来源 tenant/account/provider/API/model 完整匹配。
- 实际 endpoint、认证形状、型号、options/tools、完整 exchanges/events/Result/Observer；恰好一个 HTTP Attempt，调用/响应归属为自己的 Kimi 身份。
- `grokhistory.Verify` 检查三个有序 call/result 对应、ID 不冲突/合法、参数和结果标记完整；保留可见文本，去除外国密文/资源状态；`NativeStateDowngrades.CrossModel=2`。
- 成功终态 Stop、用量报告及 resolved/bounded 成本范围；原始响应经当前公共 Client 解码与保存 Result 一致。输出耗尽、工具终态、截断、错误或缺原件均 FAIL，不追加第二次调用。
- 资源归零、原预算预留/实际计数、原始 grant/权威 reservation、冻结 source/build/binary/SDK/catalog/价格快照、manifest/audit/seal 与实际进程退出。

## 3. 六份新授权及预算

| 路线 | Provider | 协议 | 账户/地区别名基线 | 新 calls / HTTP | 新 USD 上限 |
| --- | --- | --- | --- | --- | --- |
| kimi-cn-chat | moonshot-cn | openai-completions | KIMI / CN | 1 / 1 | 0.10 |
| kimi-cn-responses | moonshot-cn | openai-responses | KIMI / CN | 1 / 1 | 0.10 |
| kimi-cn-messages | moonshot-cn | anthropic-messages | KIMI / CN | 1 / 1 | 0.10 |
| kimi-intl-chat | moonshot | openai-completions | KIMI / GLOBAL | 1 / 1 | 0.10 |
| kimi-intl-responses | moonshot | openai-responses | KIMI / GLOBAL | 1 / 1 | 0.10 |
| kimi-intl-messages | moonshot | anthropic-messages | KIMI / GLOBAL | 1 / 1 | 0.10 |

同名别名不表示两市场同一账户。未来 Host 须确认每路线的真实 standard API 账户/tenant/account scope；若与原证据不同，不跨账户拼接。仅持当前所选市场/路线凭据，不使用 coding/M-plan，不读 xAI 凭据，不混发六路由。

每份新批准固定：

| 项目 | 上限/约束 |
| --- | --- |
| 场景/型号 | 仅 `kimi-k3/xai-origin-history` / `kimi-k3` |
| 逻辑调用 / Attempt / SDK 与环境重试 | 1 / 1 / 0 |
| 单次及累计 requested output | 4,096（沿用现有该场景合同，含模型计费推理输出；不扩大） |
| input reservation | `max(2048, 实际编码 wire 字节数+1024)`，累计 ≤4,096；不是以模型报告的估算 token 代替前置预留 |
| USD | ≤0.10，先预留最贵输入档 + 全部输出上限，失败/取消不按下界释放批准 |
| 单调用 | 120 秒，包含取消/读取等现有资源控制 |
| 首次调用前等待 | `callIntervalMs=60000`，复用现有 waitInterval；按原生观察时间核验 |
| 原生进程 | 240 秒，覆盖首次等待 60 秒 + 单次 120 秒 + 启动/收尾余量；内层 deadline 从此 profile 推导，manifest 记录真实 timeout |
| 有效期 | 未来实际执行前签发短有效期，建议 30 分钟，仍不超过现有 24 小时上限；本轮不填写 ID/expiry |

六进程串行。现有 pacing 包括首个调用前的 60 秒，串行执行自然保留路线切换间隔，不加账户级调度器/跨进程节流账本。六原生进程时间预算合计 ≤24 分钟，不含离线构建、消费和人工间隔。限流/权限/价格异常时停止剩余付费执行，保留失败与已消费授权；不自动重试、换路由、重发旧 grant 或恢复旧暂停轮次。

### 费用推导及适用范围

使用仓库冻结 `kimi-prices-2026-10-22.1`，不是实时价格/汇率报价。原价格原件和换算见 [prices.json](../../ai/internal/testkit/kimi/prices.json) 与 ADR-0052。

现有 `RegionalReservedUSD` 按 `max(Input, CacheRead, CacheWrite, 2*Input)` 预留输入（覆盖 1h 写价），不靠实际 TTL/缓存命中降低上限。取输入/输出各 4,096：

- 国际：`4096*6/1e6 + 4096*15/1e6 = USD 0.086016`。
- 中国：目录转换后上界约 `USD 0.0853079440646144`；原币上界为 `4096*40/1e6 + 4096*100/1e6 = CNY 0.57344`。美元转换沿用留存 2026-10-16 的 6.7220，不是今日外汇结算承诺。
- 六条 token 上限对应保守预留合计约 `USD 0.5139718321938432`，拟批准总额 **USD 0.60**；实际发送按真实 wire 计算，无需把每条 0.10 全记为实际消耗。

现有受控 StreamSimple/64 夹具输入预留为 Chat 2,594、Messages 2,599、Responses 2,953；这是 4,096 输入限额的初步依据，不能冒充新采集入口的 full/4096 证明。实施后的无 Key E2E 必须用新入口的实际最终 payload 再验证。若过界则发送前拒绝，回到离线设计；不悄悄提高 cap。价格/模型变化同理须重新核定该路线预算。

### 授权合同

复用 `RegionalGrant`、`ReserveRegionalGrant` 与 `livegrant.Consume` 的存储/原子消费算法，不引入授权平台。新增 supplement schema 2、固定 `purpose="xai-history-supplement"`、contract `kimi/<route>/xai-origin-history/v1`；绑定精确单项 scenario IDs、新 binary/source/build/SDK/candidate catalog/价格、上述预算、自己的账户/地区、原 baseline manifest SHA、原批准人类范围。

schema2 对关键范围字段采用现有有界、无歧义的解析与校验，未知/重复范围不回落到旧授权路径。原 full/standard schema 1 及历史九项原件按其原合同核验；它们不能授权 supplement，新的单项 grant 也不能授权 full 十项。新 schema 2 范围全部进入现有规范 scope 哈希；继续使用 `.git/barness-kimi-grants` 的 ID + scope 两把永久 reservation，不建新账本、不迁移/修改旧记录。成功、失败、取消或部分 claim 都消耗该次批准；补签/再次执行需独立新批准。六次成功预期追加六组 reservation（每组 ID/scope 两原件）；旧文件逐字节不变。

本轮只保存设计型预算表，**不生成可被 processgrant 接受的授权文件，不签发 ID、有效期，不 claim，也不触碰凭据**。将来先完成离线实现/冻结，再把完整来源、二进制哈希、六个 baseline manifest 与具体 paid 范围交给人批准。批准实现和批准 USD0.60 真实执行是两件事。

## 4. 一个扁平的组合消费入口

每路线一份新 `kimi-composition.json`，仅包含类型/version、精确 combo/model，以及两个原件引用：

1. `baseline`：本方案预先指定的旧九项原生 bundle + manifest/report/seal 哈希。
2. `supplement`：未来新单项 bundle + manifest/report/seal 哈希。

当前不生成假 supplement 引用。组合不复制胶囊、不拼改原报告、不嵌套组合、不递归寻找历史轮次、不按成功结果自动挑包、不接受任意场景列表。新 grant 绑定的 baseline manifest 必须与此输入一致。

模块接口收敛为 Kimi 专用 `VerifyKimiComposition(dir, formal)`（拟新增），内部使用两个固定校验 profile：

- **baseline-nine**：验证原九项/原 9-call 授权及预算、原时刻到期规则、真实 mode/stage、固定 native-history/tool-result 前序、全部原件/账本，随后应用已完成的当前来源与 used-model/SDK 适用性。它只返回九项贡献，不能独自产生当前正式 PASS。
- **history-one**：验证新 purpose/schema/contract、唯一 xai-origin-history、1-call 预算/权威 reservation、真实退出与完整捕获，拒绝标准/受控/回放材料冒充真实补采。

消费者必须独立计算集合，得到同路线唯一十项：`oldNine ∩ newOne = ∅` 且并集严格等于 `RegionalExpected(route)`；不重复、不缺项、不跨站点/协议/账户/型号借证据。各测量各自核验原 source、目录、授权/账本与实际预算；不要求两次执行整套来源相同，也不能把旧 9 次记进新 1 次的授权/支出。

输出是**当前资格记录**，保留 `qualificationAt` 和两个 `measurements` 的真实提交/时间/来源/阶段/预算，以及十项到原 manifest/case/artifact SHA 的映射。不得造一个“十项同进程”的测量 GitCommit/StartedAt/FinishedAt、统一 grant 或追写原候选报告。只做当前能力集合的内部投影，不回写旧测量。

公共外壳沿用现有 releasegate / consume_live.py，不新增脚本控制器：

- candidate 检查允许从组合记录得到 `CANDIDATE_COMPLETE`，只表示十项完整的当前候选证据。
- `--bundle <组合目录>` 的正式消费识别专用组合类型；未有正式目录数据审阅/激活、当前资格或模型投影不一致则拒绝正式 PASS。
- 正式目录获准激活后，只读核对六个自身型号事实及本组合记录；不因原 stage=candidate 再付费重跑整组。原 stage 不改。必要时在 ADR-0055 与 REGIONAL.md 明确这种受限二来源组合，替代活动文档“必须同一 bundle”的旧要求。
- 外层继续 schema 2：唯一实际 releasegate 子进程、数值退出、binary/source、组合清单和两个被引用包的**完整输入库存**、原资格输出、stdout/stderr/exit 哈希均绑定。不能只盘点组合目录而漏掉外部两包；记录/发布消费同样核验这组传递输入。

## 5. 最小实现范围与验证

本轮未修改以下实现。实施时仅作这些改动，原 `RegionalExpected` 十项标准集合、产品 adapter/SDK 和其他 Provider 合同保持：

| 模块位置 | 必要变化 |
| --- | --- |
| supportmatrix 地区授权/预算模块、processgrant | 固定 supplement profile 和模板 purpose；schema2 的范围/原 baseline pin；复用原存储消费算法，模板仍不授予执行权限 |
| ai/live regional config/harness/finish、main TestMain | 一个精确新入口；依据已验证 grant 选择固定一项及 240s/1-call 预算；原 full 入口严格性不放宽；manifest/ReplayRun 记录并核对精确入口及 240s timeout；单项完成不冒充全十项 |
| supportmatrix Kimi 证据模块（拟独立小文件） | 两个固定原件校验 profile + 一个扁平组合接口；复用现有 raw wire、Result/Observer、用量、资源、seal 和授权检查 |
| release 现有消费/输入库存/record 模块 | 只增加 Kimi 专用分派，消费组合资格；record 使用每份实际测量来源引用，不把资格提交当成单次测量提交；外层命令/退出采集不重写 |
| ADR-0055 / REGIONAL.md | 实施阶段补充组合合同、预算/授权边界和目录晋升条件；本设计阶段不修改已接受决策 |

不增加公共 Client 选项、通用 DAG/任意场景合并、全局资格重置、账户级调度、轮次数据库、旧 grant fallback 或新 claim 算法；不自动扩大为 60 次 full formal。新文件按职责保持 ≤500 行，原文件接近上限则将 Kimi 组合逻辑放入自己的模块。

先列失败场景，再写公共 Client/CLI/子进程 E2E（本轮只列清单，不写/运行测试）：

- 新入口恰好一次请求、首等 60 秒、预算/目录/SDK/官方 endpoint 及实际退出正确；重复发送、超输入/输出/USD、超时/取消、已消费/部分 claim 拒绝，资源归零。
- 旧 9-call 授权不能启动单项；新 1-call 授权不能启动 full；错 purpose/schema/场景/route/Provider/账户/地区/型号/source/binary/baseline、过期、无新授权或非精确 test.run 都在发送前拒绝。
- 单包九项、单包一项均不能得到完整当前正式资格；受控/回放不能填真实缺口。54+6 的受控组合仅验证消费者机制，不产生真实支持。
- 跨站点拼包、重复/缺项、篡改/额外/缺失原件、路径逃逸/symlink、报告/manifest/seal/权威账本/传递输入任一不匹配均拒绝；集合缺项必须精确报对应场景。
- 篡改 Host/封套、实际 call/result 配对/顺序/参数/标记、泄露外国密文/native resource、降级计数、Result/Observer/用量/费用、资源或退出证明，原封不动重封也不能通过。
- 未激活目录时 candidate-complete 可以成立，formal 必须拒绝；激活后仅相应六身份当前资格可通过，其他 Provider 不随动。

先受控进程与定向场景 RED/GREEN，再按 SOP 跑完整受影响普通 E2E、supportmatrix/release/相关 live-tag **无厂商精确入口**、工具进程与 vet；先于后续实现冻结取得完整普通回执。所有合成 grant/claim 在私有 GIT_DIR/common namespace 中，写入前断言隔离并在前后核对真实六处存储。新授权字段/并发调用若涉及 race，列人工候选范围，本轮及实现批准均不自动执行 race/压力。各阶段绑定各自真实输入；将来填入新 grant 的 ID/有效期等批准事实，不因它未参与普通测试就要求重跑全普通。只有实际产品/采集/消费行为或所用输入变化时按影响补验证。

## 6. 推进顺序及停止条件

1. 审阅本设计；后续若获准实施，先完成六份旧九项来源审阅与当前离线回放，形成无 paid 的逐项适用性。
2. 实施单场景采集与组合消费者，完成上述受控/普通验证、传递库存和历史留存核验；不创建可执行 live grant。
3. 冻结新实现、实际 payload 输入预留、SDK、candidate/价格、二进制，形成六份具体待批准范围。任一前置缺口未解决，不进入付费。
4. 人明确批准这六条固定单场景/新账户范围/总 USD0.60 后，才逐路线签发新 grant 并串行采集；失败停止，旧暂停轮次始终不动。
5. 每条仅将原九项与自己的新一项只读组合；六条十项均完整后，单独进行正式目录数据审阅/激活和当前正式消费。再按功能影响判断整体发布及人工阶段，不把本组合 PASS 等同全部 Provider 已交付。

若实际产品差异使某条原九项中个别场景失去适用性，或价格/输入上限不足，本方案的“只补六项”前提不成立：列出精确场景、理由与预算差额供重新审阅；不自动沿用本次提案的额度做其他测试。本轮所有新采集、grant/claim、目录激活、正式组合资格、race、压力和暂停恢复均 NOT_RUN。
