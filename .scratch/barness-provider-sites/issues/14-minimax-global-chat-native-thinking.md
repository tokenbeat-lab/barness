# 14 MiniMax 国际 Chat 原生推理真实验收失败

Status: claimed
Type: task
Blocked by: 15

## Context

[07最终验收](07-final-acceptance.md)和[规格R06/R18/R20](../spec.md)。当前状态唯一入口为spec；本票保留本次新故障，不撤销此前GLM重验和六formal成功。

## Actual failure

来源600a4ec的国际MiniMax-M3 Chat完整13场景12 PASS/1 FAIL，实际13 Calls/12 HTTP，thinking-on未观察到原生推理。显式使用http://127.0.0.1:7890。原始流只有content/role，没有reasoning字段，usage.reasoning_tokens=0；最终请求与旧自身成功逐字相同。当前公共Client离线两次精确重放均RED，旧阳性响应同Client为GREEN，同进程工具推理也正常保留。无依据修改parser、请求形状或削弱断言。[只读诊断](../../../.evidence/provider-sites/final-acceptance-20261020/minimax-chat-diagnosis/report.md)。

父UNKNOWN、整笔US$0.20、原nonce/ledger与FAIL均保留。跨轮实际303/302、领取456/492、整笔US$10.70，17进程未启动；[停止事实](../../../.evidence/provider-sites/final-acceptance-20261020/remaining-stopped/index.json)。无新READY/race，release.pass=false。

## Acceptance

若新增完整进程，先落实15固定恢复授权/源码及预算，再取得具体新付费范围批准；不能重用已消费nonce、原grant、剩余7calls或历史ID冒充完整重跑。新完整自身证明及剩余23支持/四效果齐备后才继续07。保留原失败，不从一次无reasoning推断永久不支持或已证实随机根因。
