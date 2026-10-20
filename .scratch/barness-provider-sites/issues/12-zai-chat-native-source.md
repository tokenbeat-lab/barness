# 12 ZAI Chat 原生续接的固定消息来源

Status: claimed

2026-10-20：独立worktree已在公共CLI用当前真实原始SSE重现17/18，随后固定zai-chat full-stream来源并保留完整工具链、独立consumer来源/请求/actual wire核对；源无native仍失败且禁止fallback。普通实施与全部中间FAIL、最终六路线受控cohort、21个实际Frozen PiDiff、消费者与双vet事实见[实施记录](../evidence/zai-chat-native-source.md)。未发新真实请求；继续claimed，等待整合复审与新授权真实候选。
Type: task

## Context

[规格 R06 / R10 / R18](../spec.md) 要求工具往返与原生 reasoning 续接，并精确核对来源；没有要求原生消息必须由工具调用产生。首轮最终候选的 `zai-chat/glm-5.3/native-history` 实际失败：工具 SSE 没有 reasoning，同轮 full-stream 则有自身原生推理，请求与上次成功时相同。不能制造解析器修复或搬用旧成功字段。

根 `.evidence/provider-sites/final-acceptance-20261020/closed-cohort-1/` 保存关闭指纹；`candidate-final/candidate-zai-chat/20261020T104434.930813000Z/` 保留 17 PASS / 1 FAIL、18 Calls / 18 HTTP、零重试及零审计发现。Messages 的独立 18/18 成功保持原事实。

## Required behavior

仅 `zai-chat` 在请求前固定同进程、同完整身份与账户的 full-stream 场景为原生来源，保留真实用户/助手前缀及原工具调用、Host 结果的完整链。其他五路线保持固定来源。不得事后搜索、挑选、回退到兄弟身份或合成 thinking / signature；源无原生仍失败。

报告声明每型号 sourceCase；独立消费者核对固定来源集合、完整 tenant/account/Provider/API/model、请求确切前缀、真实 wire 原生投递及零降级。工具往返、分片、用量、终态、取消和场景全集保留，不增加调用或降低资格要求。

## Ownership / verification

ZAI implementer 在独立 worktree 负责 harness、原生消费者及公共 Client/CLI E2E。先用本次原始 SSE 受控重现失败，再最小实现和定向普通验证，保留 argv、数字退出、UTC、来源及可重放工件。不得真实 Key、厂商调用或 race。

## Acceptance

公共边界 RED→GREEN、来源错配与缺失原生反例通过，独立审查无剩余问题。新来源重新冻结；六候选仍要求同 commit/build/binary，付费重验须新的有限授权。旧失败、grant / claim / 预算和关闭轮次不改写；新资格及完整验收完成前保持 claimed。
