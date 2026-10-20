# 按区域路由的缺口验证

本轮依用户指令：中国区GLM、DeepSeek直连，国际OpenAI、Anthropic、Google、TypeSafe固定经`http://127.0.0.1:7890`。完整交付仍为PARTIAL；各入口的保留校验通过只证明字节完整，不把实际FAIL、NOT_RUN或历史来源升级为当前支持。

| 入口 | 实际范围与结论 |
| --- | --- |
| [网络诊断](network/README.md) | 无凭据的DNS/TCP/CONNECT/TLS观察，不证明模型或账户资格。 |
| [执行工具](implementation/README.md) | 固定18进程合同、来源/binary/审查/单次intent和区域路由的失败先行验证。 |
| [GLM实际](actual-glm/README.md) | 来源51c3433，65calls/65Attempts，63PASS/3FAIL；Chat/Messages各22PASS，Responses19PASS/3FAIL。严格三路续期拒绝。 |
| [其余15项实际](actual-legacy/README.md) | 来源8fbd170，15个collection PASS，180/180；155个原生场景/样本为82PASS、1UNSUPPORTED、72answered。含四效果、派生Jev支持包和M3.1独立NOT_RUN。 |
| [启动历史](keyless-launch-history/README.md) | 原`go run`启动因PATH观测失效而由操作者终止；原exit-15、无报告及探针RED/GREEN均保留。 |
| [完整门禁](final-gate/README.md) | 同一公开CLI直接编译运行新全量，来源8fbd170，四项数字退出0/0/1/1；报告pass=false。两套各12439PASS/8FAIL/1NOT_RUN/audit0，868真实PIDiff和29压力case均PASS。 |
| [独立两轴审查](reviews/README.md) | Standards和Spec分别保留原始报告、发现和复审；源码审查不替代真实验收。 |
| [原映射Hash缺口](release-plan-gap.json)、[其余离线缺口](postgate-gap-2.json) | 记录完整门禁暴露问题；后续修复由[票14](../../issues/14-release-preflight-gaps.md)承接，不改写原运行。 |

固定18个实际进程已全部关闭，17PASS/1FAIL。本轮新增245calls/245HTTPAttempts，上限374/398；此前本次MiniMax续期300/300保留，累计545/545。各进程原始收据和规范intent不可重用，剩余额度不转移，没有自动新轮次。M3候选/正式及更早独立GLM专项消费分别记账。

GLM Responses两次约90秒deadline_exceeded没有记录HTTP响应；依赖工具结果FAIL0/0。捕获及完整性inventory齐全，complete=false、无成功capturePins。两个型号native-history本轮PASS，原R1缺正文失败依旧保留。当前矩阵的16PASS/1FAIL准确反映最新实际结果；历史allPassedAt没有覆盖失败。

[原状态快照](pre-routed-delivery-status.md)与[机器快照](pre-routed-delivery-status.json)保留上轮300/300、其余15项2PASS/13FAIL、原始完整门禁没有报告的历史。[8fbd执行来源核对](current-source-impact.json)及[早期83检查点](selected-source-impact.json)分别保留；后续代码来源适用性不得通过刷新旧pins取得。

大原始包在主仓库`.evidence/minimax-retained/`按各索引固定Hash保留，有限逐文件清单、复核程序和命令在各入口中。未改动用户原有session-analysis和其他功能human-review。当前实际与后续修复适用边界见[交付状态](../../delivery-status.md)。
