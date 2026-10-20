# Messages glm-5.3/cancel 单用例复测

2026-10-19，**PASS**。中国区直连，已提交来源 `6a24ba3308ade8712dd4b0d20d67d9442603ad67`，实际 mode=`live`。本次用户新指令独立授权该单用例；最多1call/1HTTPAttempt、4096输出token、120秒调用期限，无重试或自动下一轮。

| 观察 | 实际结果 |
| --- | --- |
| 目标 | Zhipu / anthropic-messages / glm-5.3 / cancel |
| 端点 | `https://open.bigmodel.cn/api/anthropic/v1/messages?beta=true`，DIRECT |
| 请求 | 与原失败的wire request逐字段相等；原固定整数提示词及4096上限 |
| 事件 | 收到非空 `ThinkingDelta` 后立即取消 |
| 取消结果 | `canceled` / `stream`；HTTP200；StopReason=`aborted` |
| 用时 | 场景duration861ms（最终消息duration862ms，两个测量字段均原样保留） |
| 消费 | 1 logical call / 1 HTTP Attempt，零重试；续期647/647→648/648 |
| 用量 | partial，不能将零已报计数当作免费 |
| 审计 | 原生及launcher均PASS；native完整性文件逐项哈希复核通过 |

[机器结果](result.json)、[实际命令](actual-command.json)、[一次性intent](paid-intent.json)、[固定计划](plan.json)、[直接用户授权](authorization.json)、[冻结来源](source-at-start.json)、[二进制](binaries.json)、[执行后来源守卫](post-source-guard.json)。执行仅向子进程注入所选GLM_KEY，不继承其他凭据或代理；未保存Key/环境文件。`prepared.json`是执行前快照，实际关闭状态以 `paid-intent.json` 为准。

原正式22场景入口不能筛选，本次增加独立诊断入口并直接复用原 `glmSession.execute`、固定plan、options及取消断言。受控测试先保留真实RED1及第二份完整拒绝子进程RED1，普通/race各9/9 PASS、live vet0；[受控来源与记录](controlled/source-proof.json)。源码两轴各0发现：[Standards](standards-review.json)、[Spec](spec-review.json)。实际提交后二进制又执行9项受控检查，最终 `checks-temp.json` 为PASS；最初同二进制放在用户目录中，9项场景通过但进程审计因命令记录含home路径而exit1，原件不改写。将同SHA256二进制复制到系统临时目录后复验exit0。另有执行前review字段名检查KeyError（Spec用verdict、Standards用decision），发生于intent及Key读取之前，0/0，见 `preparation-error.json`。这些均未引发付费请求。

[留存索引](index.json)和gzip全文件inventory指向主仓库 `.evidence/minimax-retained/ticket18-messages-cancel-actual/raw.tar.gz`，包括首次审计拒绝、最终无Key检查、原始真实请求/SSE/事件/断言、完整性报告及所有收据；可用标准gzip/tar读取，无需运行收费入口。实际报告原件在 `.evidence/glm-messages-cancel-20261019-1/actual/20261018T230218.566865000Z/diagnostic-glm-messages-cancel-report.json`。二进制按独立SHA留在原输出目录，未收入Git或原件压缩包。

原120001ms超时失败所有文件哈希仍匹配 [原失败pins](original-failure-pins.json)；四个正式selector及默认矩阵仍匹配 [执行前pins](publication-before.json)。没有拼接原21个通过场景、生成新的formal22支持报告或修改矩阵。**此PASS只更新该单场景的最新观察**；当前来源下的全Messages正式报告、11组严格来源资格、GLM严格续期/四报告装配以及完整发布门禁仍开放。交付保持PARTIAL，本轮1/1授权与进程均已关闭。

最终实际结果与上下文在提交 `3227b5a` 接受独立[Standards](final-standards-review.json)及[Spec](final-spec-review.json)复核，两轴均0发现。该复核覆盖本次单场景、留存与上下文，不授予完整正式验收认证。[清理记录](cleanup.json)确认本轮两个自建工作树和临时二进制已移除，原冻结二进制及原件继续留存，两个用户目录保持。
