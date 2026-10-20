# 08 ZAI Messages GLM-5.2 原生推理资格缺口

Status: claimed
Type: task

## Context

[规格 R06 / R10 / R18](../spec.md)、[07 最终验收](07-final-acceptance.md)、[验证策略](../../../docs/agents/verification.md)。2026-10-20 获批六条新路线各一次真实 candidate，US$5、standard API、零重试。81 Calls / 81 HTTP 中唯一失败为 `REGIONAL-zai-messages-glm-5.2-native-history`，ZAI Messages qualify 实际 exit 1，其余五路线 qualify exit 0。

## Observed failure

原 Messages 进程 17 PASS / 1 FAIL，HTTP 200、合法 stop、Host 工具结果、用量、真实归属与资源释放均通过；失败断言为 `own native reasoning retained without downgrade: native=false`，检查的是先前工具响应。原始工具 SSE 完整结束，只含 text 与 tool_use，未返回 thinking / signature / reasoning；本轮 GLM-5.2 Messages 九次原始响应均无该数据。不能借 GLM-5.3、另一协议或另一站点补足资格，不能改为 UNSUPPORTED 缩减目标。

## Diagnosis

独立 explorer 按 diagnosing-bugs 建立公共 Client 的确定性离线反馈环：原两轮请求 JSON 与原捕获逐对象相等；原响应两次 RED，最小一轮工具响应 RED；同一解码器与二轮重放的 GLM-5.3 原件对照 GREEN，保留原 151 字符 thinking 与 24 字符 signature，降级计数均为空。33 帧经 45 次离线删除探测收敛为 12 帧的明确合成最小反例；不回写真实捕获。

结论仅为本轮上游原始响应未给出原生推理，没有证明 Client 丢字段；也不证明永久不支持或随机性。全部九个请求均未显式发送 thinking。现有中国 Zhipu 的 enabled / budget1024 / max4096 资格不能移植给 ZAI；官方 Chat 示例中的默认开启不能证明 Anthropic-compatible Messages 的具体形状或默认。没有依据为缺失原生字段伪造修复或降低断言。

[完整诊断与确切命令](../../../.evidence/provider-sites/acceptance-20261020/diagnosis/report.md) / [来源、原件及探针 Hash](../../../.evidence/provider-sites/acceptance-20261020/diagnosis/index.json)。保留根原件的逐字节留存回执为 `diagnosis-retention.json`，两份来源清单仍为普通 / race 受测 `470682d3…` 与原生 `f3848619…`；没有改 ai/、SDK、pins 或原捕获。

## Acceptance

先取得该自身协议的明确思考控制 / 原生返回合同；如需新研究调用，必须单独列出精确形状、型号、用途、范围和预算并获批。再按公共 Client / CLI 先失败后实现，独立审查后以新 grant 完整重取这条路线的两型号候选资格。原始失败、阶段区别与账户边界保持，不转用已用 grant 或未花预算，不刷新旧原件。source / 目录变更后重新按验证策略取得普通与人类选择的 race。通过候选仍须独立晋升与新 formal，最终 feature 状态只在 [spec](../spec.md#当前状态) 维护。

## Comments

2026-10-20：用户已批准完整有限轮次 US$30。四次真实 ZAI 对照完成：GLM-5.2 omitted 无原生块，精确 enabled/budget1024/max4096 返回 thinking97 / signature24；GLM-5.3 两形状均有原生块。4 Calls / 4 HTTP、零重试、合法终态，[完整自身研究](../evidence/zai-thinking-research.md)。这只证实窄控制形状，不表示 native-history 或候选资格完成；下一步公共 Client RED→GREEN、快照与 full 请求更新、独立复审后再采新候选。

2026-10-20：用户要求继续解决本项后完成最终验收。已固定 ZAI 自身两型号 × 两形状（省略 / enabled-budget1024-max4096）的四个原始 Messages 研究请求及原工具 prompt；研究只区分默认/控制形状，不认证原生历史。工具经独立审查修正 redirect、响应头秘密与首建 claim 持久化边界；本轮预算尚待具体批准，没有读取 Key、发请求或修改生产协议。原失败与离线 RED 继续有效。

2026-10-20：离线诊断已完成；资格缺口继续开放，没有执行任何付费重跑。
