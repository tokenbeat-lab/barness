# 13 智谱 Responses GLM-5.2 原生续接缺少最终文本

Status: claimed
Type: task

## Context

[07 最终验收](07-final-acceptance.md)、[规格 R06 / R18 / R20](../spec.md)及[验证策略](../../../docs/agents/verification.md)。唯一当前状态在[spec](../spec.md#当前状态)；本票保留故障和后续处理要求，不把候选成功视为完整发布。

## Observed failure

已获明确批准的固定续验2在源码 `b384951448aa74c88f2e1342e7623af565922aad` 下，智谱中国 Responses 完整22场景实际21 PASS /1 FAIL，22 Calls /22 HTTP、零重试、child exit1、audit零发现。唯一失败是 `glm-5.2/native-history` 的“required text entry produces final text”。自身原生来源、可信归属、真实 InputReasoning object续接、零降级、HTTP200和合法completed均通过。

实际原件只有reasoning项，最终 `response.completed.output` 也没有message/output_text/refusal；Client保存Thinking5字符、Text0，事件无text且没有溢出或截断。当前GLM-5.3同协议及旧GLM-5.2自身成功原件均具有独立message/output_text。新旧请求仅自身先前推理文本不同，结构、控制和可见助手答案一致。没有依据把reasoning转为可见文本、修改parser或删去最终输出断言，也不能从一次结果推断永久不支持或随机性。

[只读诊断与原件Hash](../../../.evidence/provider-sites/final-acceptance-20261020/glm-responses-diagnosis/report.md)；[停止事实及预算索引](../../../.evidence/provider-sites/final-acceptance-20261020/round2-stopped/index.json)。原生完整报告位于索引所列目录，失败、父UNKNOWN和整笔USD预留均保持原样。

## Resolution requirements

先明确可验证的原因及最小处理；没有证据时不改协议形状或削弱合同。若需改代码，先在公共Client/CLI用自身原件重现，再最小修复、独立复审、冻结相关完整来源并判断真实重验范围。旧成功、另一型号或站点不能补位。

若需要再次付费验证，须先准备具体有限计划并按验证策略取得新的明确范围批准；原续验2明确禁止自动进程重跑。真实完整Responses22场景及后续自身证明通过前，07与feature保持开放，没有新READY或正式race。原九服务资格及最终一次race选择不重复询问。

## Answer

离线诊断完成，未发现可证明的解析或请求结构回归；本轮上游实际输出不满足最终可见文本合同。未修改运行代码、断言、原件或账本，未新增厂商调用。完整验收仍受这条真实失败阻断。

## Comments

2026-10-20：用户直接明确批准“智谱中国 Responses 完整22场景再跑一次，保留失败原件，总预算仍为 US$30；通过后继续剩余验收和最终完整 race”。[实际追加授权](../../../.evidence/provider-sites/final-acceptance-20261020/responses-retry-authorization.json)保留原文；仅固定一次补充进程，旧失败/UNKNOWN/整笔预留不变。通过真实b384隔离checkout、冻结原二进制及完整245来源，复用已成功Chat；不改Go/请求/断言或原报告提交。固定控制器追加入口须先公共流程RED/GREEN和独立复审，再由根任务执行；当前结果只在spec维护。
