# MiniMax 交付状态

Delivery: PARTIAL
Updated: 2026-10-18
Integration branch: `codex/minimax-global`

已按指定路由完成本轮固定验证：中国区GLM、DeepSeek直连，国际OpenAI、Anthropic、Google、TypeSafe通过`http://127.0.0.1:7890`。其余15项身份/效果全部通过；GLM Chat、Messages通过，Responses仍失败。默认17身份矩阵已原子导入，记录16项PASS和1项FAIL。完整公共发布门禁已生成`pass=false`报告，spec、10、09保持开放。

| 范围 | 实际结果 | 依据 |
| --- | --- | --- |
| 国际站M3三协议自身正式 | 三项PASS，39 calls/36 HTTP Attempts；原263文件映射仍适用 | [正式证据](evidence/ticket-10/international-formal-glm/README.md)、[来源核对](evidence/ticket-13/current-source-impact.json) |
| M3.1三协议线上候选 | 三项NOT_RUN，0/0；没有获准M Plan资格，离线约束仍必交 | [本轮候选状态](evidence/ticket-13/actual-legacy/README.md) |
| 本轮GLM三路 | Chat、Messages各22场景PASS；Responses为19 PASS/3 FAIL；共65 calls/65 Attempts | [GLM实际证据](evidence/ticket-13/actual-glm/README.md) |
| 其余15项身份/效果 | 15个collection PASS，180 calls/180 Attempts；155个原生场景/样本为82 PASS、1 UNSUPPORTED、72 answered | [实际证据与Jev支持包](evidence/ticket-13/actual-legacy/README.md) |
| 默认17身份矩阵 | 导入成功，最新16项PASS、GLM Responses FAIL；旧allPassedAt保留 | [矩阵](../../ai/live/support-matrix.json) |
| 完整普通/race | 各12439 PASS、8 FAIL、1 NOT_RUN，审计0；868个真实PIDiff、29个Pressure case均PASS | [完整门禁证据](evidence/ticket-13/final-gate/README.md) |
| 四项实际命令与发布报告 | vet=0、live vet=0、ordinary=1、race=1；CLI exit1，报告pass=false | [完整门禁证据](evidence/ticket-13/final-gate/README.md) |
| 后续离线修复 | 固定映射Hash已在工作树修正并复核；可选诊断注册顺序和严格inventory测试夹具正在补正 | [票14](issues/14-release-preflight-gaps.md) |

GLM Responses的`glm-5.3/tool-call`和`glm-5.3/control-max`各约90秒后`deadline_exceeded`，各1call/1Attempt且未记录HTTP响应；依赖的`tool-result`保留FAIL和0/0。没有足够证据进一步归因到DNS、TCP、TLS或服务端处理。两个型号native-history本轮均PASS；原R1缺正文失败保留为历史。

Responses捕获和完整性清单齐全，但`complete=false`，没有成功capturePins。严格三路raw/known-usage/support views续期拒绝，四个旧selector字节不变。GLM的六项正式导入/用量离线case因此仍失败。另两项离线case因固定映射Hash漂移失败；可选OD57跳过还进入了mandatory evidence inventory。完整release包另有测试夹具缺少文件Hash清单。[只读诊断](evidence/ticket-13/postgate-gap-2.json)区分这些离线缺口及实际GLM阻断；后续修复不能改写原全量结果。

本轮18个付费进程已全部关闭：17 PASS/1 FAIL，新增245/245，上限374/398；此前本次MiniMax续期300/300保留，累计545/545。GLM原轮/R1、本轮以及更早独立GLM任务分别记账。M3候选历史49/45、正式39/36另列。未知usage不记免费，未用额度不转移，不自动新增轮次。

GLM实际来源51c3433，其余15项及完整门禁来源8fbd170，冻结执行文件逐字一致。后续票14修复处于独立工作树，尚未合入；不能把旧全量运行授予修正后源码。两轴独立政策复核确认：按既有受影响组合规则，拟复用的10项generic与2份文本效果相关执行范围保持则可继续使用；需新验的固定6进程为GLM三路、protocol114及两份视觉效果，上限108calls/132Attempts。GLM还将按[票16](issues/16-glm-host-deadline-alignment.md)把专属Host响应头等待对齐到已有120秒调用上限，并独立验证；不证明上游会在120秒成功。此范围须经最终源码复核，详见[票15](issues/15-source-scoped-revalidation.md)，尚未请求或获得新增付费预算。

原完整门禁启动尝试因`go run`改变PATH绕过退出码观测器，由操作者停止：exit-15、无报告。保留原失败后，构建并直接执行同一公开CLI完成本轮全量；新执行没有采用旧ordinary/race bundle，也没有新增付费。全部四项数字退出码和最终false报告均已留存。

[两轴审查](evidence/ticket-13/reviews/README.md)逐份保留源码、计划与实际归档审查，审查通过不替代真实支持或完整发布认证。大原始包保留在主仓库`.evidence/minimax-retained/`，各证据入口提供逐文件Hash和无Key有界复核。用户原有session-analysis与其他功能human-review目录保持。

[上轮状态快照](evidence/ticket-13/pre-routed-delivery-status.md)保留300/300、其余15项2PASS/13FAIL、3d3a1fb全量失败及当时无报告的原结论；[旧审查链](evidence/ticket-09/final-reviews/README.md)、[GLM R1](evidence/ticket-12/actual-corrective-1/README.md)与中国区401历史均不改写。机器可读的[当前状态](evidence/ticket-09/international-delivery-status.json)与历史快照分开。
