# 2026-10-20 最终验收轮次

本页保存固定来源和本次实际执行事实；唯一当前状态在 [spec](../spec.md#当前状态)。此前完整普通、双 vet 与双轴审查见 [开发整合验证](verification.md)。本轮由用户明确要求完整 race，并另行批准新增六路线各一次 candidate、总预算 US$5，确认对应站点 standard API 资格。

## 冻结与完整 race

实际起始提交 `804dc4e6f52bbe761f7ed843c19fb16659bc2ef2`。1,230 份完整执行文件、原普通回执、32 份原生 audit、已索引输入原件与 Go/Node/Python/Git 工具身份重新核对通过。完整源码清单 Hash `470682d37e2d1c839bf5609f6898f9a1e4dafbd3dbfcd2cb77f2048a530f3e05`，原生资格执行 Hash `f38486193de4d19980123968616286259eb97a6cc500a29f8eee7a18ae1a28bc`。两种范围不同，保留各自完整清单。[实际重核](../../../.evidence/provider-sites/acceptance-20261020/recheck.json)。

```sh
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 \
BARNESS_AI_EVIDENCE_DIR=<本轮新 race 目录>/bundles \
go test -json -race ./... -count=1 -timeout=75m
```

实际执行 **exit 0 / PASS**，UTC `2026-10-20T06:40:40.287684+00:00` → `2026-10-20T07:23:50.234916+00:00`，墙钟 **2589.635s（43 分 10 秒）**。源码前后完整清单相等，没有 race detector 警告。主 `./ai/e2e` 原生 **13,244 / 13,244 PASS**；另一个 `./ai/examples/visualeval` 包为 23 PASS，分别计数。Go 节点 14,306 PASS / 4 SKIP，包 9 PASS / 27 无测试包 SKIP；四个节点 SKIP 与前次普通相同，是两个私有子进程入口及两个可选诊断，不是必需差分或压力跳过。

**32 份真正原生 Run 的 audit 均零发现**；889 份 pi 差分全部 PASS、pending 0，其中地区新增 21 份；204 个名字含 pressure 的场景通过，包含实际压力及导入拒绝控制，不能将它们全部称为负载测量。父子 Go 节点与预期失败的授权子进程不并入主包 PASS 数。实际命令、数字退出、UTC、原生 Run 清单、清单 Hash 和审计分别见 [race receipt](../../../.evidence/provider-sites/acceptance-20261020/race/receipt.json) / [原生汇总](../../../.evidence/provider-sites/acceptance-20261020/race/native-summary.json)；[主 manifest](../../../.evidence/provider-sites/acceptance-20261020/race/bundles/20261020T064049.554056000Z/manifest.json) Hash 为 `3ac51027442dcf49b46384f03007eeb0f2cddc981b4bc06ed7fa8d31fb5f5da8`。

race 的旧 HTTP 工件比较 **3,077 / 3,077 PASS**，覆盖 2,480 个实际 HTTP 请求及 body/零发送投影；比较仍只规范化 ADR-0049 指定的派生键值和七个明确并发案例的跨流调度顺序，其他 wire 字段 / 有无 / 各流顺序严格比较。[命令回执](../../../.evidence/provider-sites/acceptance-20261020/race-wire-audit/receipt.json) / [逐项比较](../../../.evidence/provider-sites/acceptance-20261020/race-wire-audit/comparison.json)。整体墙钟使用命令单调时钟，不将包耗时相加或将尾部差额标作独立审计时长。

本轮是人类明确选择的完整开发 race，没有当前 `READY_FOR_RACE` check checkpoint，因此不冒充 `releasegate -stage race -from` 的正式发布证明。正式 release 仍须先补齐下述真实来源与 formal 输入，再按原生 check / race 流程取得完整证明；来源改变时不能跨来源套用本轮事实。

## 新六路线自身 candidate

使用 [人类授权原文与范围](../../../.evidence/provider-sites/acceptance-20261020/candidate-authorization.json)、[精确机器计划](../../../.evidence/provider-sites/acceptance-20261020/candidate-plan.json) 构建源码注入的 live-tag 二进制；每路线一次独立进程、独立 grant 和持久消费记账，仅注入自身 Provider 对应的 Key。其他凭据、其他路由和 ambient proxy 不传入；官方目标直连。授权文件不含秘密，Key 没有进入命令参数、证据或文档。

| 路线 | 场景 PASS / FAIL | Calls / HTTP | 进程退出 | 独立 qualify 退出 | native audit |
| --- | --- | --- | --- | --- | --- |
| [zai-chat](../../../.evidence/provider-sites/acceptance-20261020/candidate/zai-chat/receipt.json) | 18 / 0 | 18 / 18 | 0 | 0 | 0 |
| [zai-responses](../../../.evidence/provider-sites/acceptance-20261020/candidate/zai-responses/receipt.json) | 18 / 0 | 18 / 18 | 0 | 0 | 0 |
| [zai-messages](../../../.evidence/provider-sites/acceptance-20261020/candidate/zai-messages/receipt.json) | 17 / 1 | 18 / 18 | 1 | 1 | 0 |
| [minimax-cn-chat](../../../.evidence/provider-sites/acceptance-20261020/candidate/minimax-cn-chat/receipt.json) | 9 / 0 | 9 / 9 | 0 | 0 | 0 |
| [minimax-cn-responses](../../../.evidence/provider-sites/acceptance-20261020/candidate/minimax-cn-responses/receipt.json) | 9 / 0 | 9 / 9 | 0 | 0 | 0 |
| [minimax-cn-messages](../../../.evidence/provider-sites/acceptance-20261020/candidate/minimax-cn-messages/receipt.json) | 9 / 0 | 9 / 9 | 0 | 0 | 0 |

总计 **81 Calls / 81 HTTP、零重试，80 PASS / 1 FAIL，五条路线 qualify 通过**。六份原生 audit 均零发现；[完整原始回执](../../../.evidence/provider-sites/acceptance-20261020/candidate-results.json) 与 [汇总](../../../.evidence/provider-sites/acceptance-20261020/candidate-summary.json) 保留各自 grant、来源、编译/可执行文件身份、实际命令/退出/时间及原生包。未执行晋升、formal 或额外厂商调用；没有转用已消费 grant 或本轮未花预算。

固定 token 预留上限按各自已核验 USD 费率保守估算为 US$2.282610，低于批准 US$5；返回 Result 用量估价之和为 US$0.020954。两者都不是厂商账单，取消 / 部分用量可能缺少服务端已经计费的工作，不据此声称账单或剩余可用额度。

## 唯一真实失败及离线诊断

[08 资格缺口](../issues/08-zai-messages-glm52-native-qualification.md)：ZAI Messages `glm-5.2/native-history` 实际失败；原进程 17 PASS / 1 FAIL，`native=false`。HTTP 200、合法完成、Host 结果、用量、真实归属与资源释放均通过。全九次该型号 Messages 原响应都没有 thinking/signature，先前工具回合只含 text / tool_use。未回写失败、缩减需求或借用另一型号/协议/站点资格。

[独立诊断](../../../.evidence/provider-sites/acceptance-20261020/diagnosis/report.md) 以同一公共 Client 实际重放捕获：原两回合 RED×2，最小单回合 RED，同一二进制 GLM-5.3 原始对照 GREEN 并保留 thinking 与 signature，降级计数为空。原请求 JSON 与捕获逐对象一致。33→12 帧的最小化材料明确标为合成离线派生，不用于 live 资格。没有证明解析器丢失返回字段，也没有证明服务永久不支持或随机性。

九个原请求均未显式发送 thinking；Chat 文档的默认开启不能独立证明这条 Anthropic-compatible Messages 路线的具体合同。显式形状需要自身协议资料与另行批准的有限研究/完整重验；中国 Zhipu 的 enabled/budget1024/max4096 资格不能移植。没有根据为本次捕获合成 native 数据或扩大协议选项；执行源码保持不变。

## 旧真实证据与正式前置

当前 keyless CLI 对原始 17 live / 四份效果实际执行 preflight，退出 **1**、ready=false、release.pass=false，墙钟 7.956s。十条通用路线仍适用；七条严格自身路线和四份效果被当前来源消费者拒绝：

- Jev 1.14：`formal source set drift`。
- 中国 GLM 三协议：`GLM execution source drift`。
- 国际 MiniMax 三协议：`renewal canonical authority: fixed renewal source/history/caps differ`。
- 两份中文效果：执行源码过期/不完整；两份视觉效果的原生 verifier：`source set drift`。

[原生 preflight](../../../.evidence/provider-sites/acceptance-20261020/legacy-input-audit/preflight/preflight-report.json)、[独立输入审计](../../../.evidence/provider-sites/acceptance-20261020/legacy-input-audit/report.md) 和 [3,539 份旧输入原件核对](../../../.evidence/provider-sites/acceptance-20261020/legacy-input-audit/original-input-pin-comparison.json) 保留确切路径、argv、数字退出、时间和来源。21 个旧输入树逐字节保持，历史 PASS 未被撤销或刷新。两个新审计/诊断目录的保留是新只读结果的字节相等留存，不重导出旧捕获。

Snapshot 与 difference-ledger 通过；新增六条路线尚未晋升 / formal 合并，矩阵保留 NOT_RUN。candidate 通过不认证 formal；上述七旧路线 / 四效果不在本次两服务 candidate 授权内。完整验收要求先解决 08、独立晋升并取得新 formal，以及补齐这些旧严格来源和效果的当前证明，再取得新来源下完整原生发布门禁。 晋升改变目录 / 执行来源后，必须重新消费全部所选旧输入；本轮十通用路线的适用性不能预先替代晋升后的核验，也不能预先抵扣其可能需要的新调用。工单 07、08 和 feature 继续开放，`release.pass: false`。

[本轮机器索引](acceptance-index.json) 绑定原始回执与审计 Hash；早先普通失败、审查 red/green、旧来源与本次付费失败全部保留。收尾仅文档与索引，不改变已测源码。
