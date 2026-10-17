# 10: 固定 GLM 的 pi 差分与扩展边界

Status: ready-for-agent
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** 维护者能从冻结 pi 基线和公共请求/结果证据判断每条 GLM 路线的兼容范围，明确区分真实差分通过与基线没有路由的扩展。

Blocked by: 09

依据：GLM 接入 spec R05、R20；验收组 A14；User Stories 57、58、62。

- [ ] 保持唯一 pi-ai 1.1.0 oracle、来源 pins 和既有合法扩展，不升级基线或引入第二套 oracle。
- [ ] 逐一核实中国区 zhipu Provider 与三个 API 是否有适用的冻结路由；不能凭国际站名称、模型同名或相似基址认定同一服务。
- [ ] 存在真实适用路线时比较公共请求/事件/Result/候选目录；不存在时按 ADR-0014 登记 Extension Route、官方 fixture 来源和跳过原因，不计为 pi 差分通过。
- [ ] 针对形状、字段 presence、原生状态、思考、工具、usage/error 等差异按 case/path 记录 fixed/extension/pending；未解释差异保持 pending，不豁免整个 body 或协议。
- [ ] 受 01–09 影响的 OpenAI、DeepSeek 两协议和原生 Anthropic 请求/结果回归与适用差分完整通过。
- [ ] 覆盖对象型 content 不丢失和合法 GLM thinking/工具回放，证明不能用原厂 adapter 的成功冒充 GLM 接入成功。
- [ ] 产出固定基线、逐路由判定、差分/扩展账本和可重复运行索引；为正式目录纳入提供确定的来源/比较规则，不要求先有收费 live 报告。

**独立验收：** 运行固定 oracle 与 GLM fixture，逐路由展示适用 parity 或有来源的 extension，全部已实现差异有界且零未解释 pending。

**依赖理由：** 差分要覆盖 09 前已完成的所有 GLM 行为和跨协议修复；不依赖候选付费验收。

