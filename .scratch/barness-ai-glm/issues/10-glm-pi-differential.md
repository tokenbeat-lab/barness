# 10: 固定 GLM 的 pi 差分与扩展边界

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** 维护者能从冻结 pi 基线和公共请求/结果证据判断每条 GLM 路线的兼容范围，明确区分真实差分通过与基线没有路由的扩展。

Blocked by: 09

依据：GLM 接入 spec R05、R20；验收组 A14；User Stories 57、58、62。

- [x] 保持唯一 pi-ai 1.1.0 oracle、来源 pins 和既有合法扩展，不升级基线或引入第二套 oracle。
- [x] 逐一核实中国区 zhipu Provider 与三个 API 是否有适用的冻结路由；不能凭国际站名称、模型同名或相似基址认定同一服务。
- [x] 存在真实适用路线时比较公共请求/事件/Result/候选目录；不存在时按 ADR-0014 登记 Extension Route、官方 fixture 来源和跳过原因，不计为 pi 差分通过。
- [x] 针对形状、字段 presence、原生状态、思考、工具、usage/error 等差异按 case/path 记录 fixed/extension/pending；未解释差异保持 pending，不豁免整个 body 或协议。
- [x] 受 01–09 影响的 OpenAI、DeepSeek 两协议和原生 Anthropic 请求/结果回归与适用差分完整通过。
- [x] 覆盖对象型 content 不丢失和合法 GLM thinking/工具回放，证明不能用原厂 adapter 的成功冒充 GLM 接入成功。
- [x] 产出固定基线、逐路由判定、差分/扩展账本和可重复运行索引；为正式目录纳入提供确定的来源/比较规则，不要求先有收费 live 报告。

**独立验收：** 运行固定 oracle 与 GLM fixture，逐路由展示适用 parity 或有来源的 extension，全部已实现差异有界且零未解释 pending。

**依赖理由：** 差分要覆盖 09 前已完成的所有 GLM 行为和跨协议修复；不依赖候选付费验收。


## Answer

2026-10-17，已完成。实现/来源登记 `af49ce0`，完整证据索引 `99d2318`，不透明 ID 有界记录 `4325d2d`；已合并最新稳定集成 `2222b9476f107843f23797ae6a094f08f231b71f`（本分支合并点 `2b8c1a6`）。

- 保持唯一 pi-ai 1.1.0、commit `abe508e1b89912adde45528136c3221eb69acdd7`、完整 43 文件 model-data SHA、827 文件 source SHA 和原 SDK lock；原 317 条决定与 3 条扩展未改。全量冻结数据证明 `zai` 是国际 Coding、`zai-coding-cn` 是中国 Coding，两者均不是中国标准服务。三个 `zhipu` 标准协议 × 两型号均登记为 ExtensionRoute，从不计为 native pi parity PASS。
- GLM 自身账本 43 条、752 case/path 对、420 个独立公共场景，零 pending；fixed 复现仍失败、未知 case/path/原生协议保持 pending。对象内容、原始/native reasoning 与请求转换、工具/历史、usage/error、跨协议隔离均使用自身协议/型号的公共 Client 证据。新增 plain-string ID 来源风险由05按先 RED 后实现修复；普通捕获仍限于其原始 delimiter-free 范围，58 个有界 ID 场景列为 fixed，额外跨 API/空 ID 场景保留独立记录。
- 当前验收：原完整 GLM 1,733 PASS + DeepSeek Responses 64 PASS；适用冻结差分 726 PASS（OpenAI Responses 335 / Chat 203、DeepSeek Chat 118、Anthropic 70），零 pending。未启用 oracle 的批次有 1 个明确 NOT_RUN 来源核验，单独 opt-in 5 PASS。合并05稳定修复后，90 ID + 5 来源/账本共 95 PASS；05自身另有 288 回归 / 124 race / 335 冻结 Responses PASS。未重复无关重型门禁。
- [运行、来源与归档索引](../evidence/ticket-10/README.md)、[机器索引](../evidence/ticket-10/index.json)、[逐成员 ZIP 校验](../evidence/ticket-10/archives.json)。durable raw 位于根 `.evidence/barness-ai-glm/ticket-10/`；每个 case 的请求、SSE、事件、Result、Attempt/Usage、断言、重放命令和 SHA 均核验。RED 与源查询 null 的 harness 诊断分开保留，不计通过。

本票不读取 Key/.env、不做收费调用，不修改 BuiltinCatalog、正式路由或 Support Matrix；正式标准 Key live、候选价格/内置纳入与完整发布证据继续由11–13负责。
