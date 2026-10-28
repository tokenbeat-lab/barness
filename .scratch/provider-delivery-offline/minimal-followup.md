# 最小后续验证清单

本清单是离线核验结果，不是付费、race、压力或暂停轮次恢复授权。原始事实、授权、claim 和账本不变。

## 先完成的离线项

| 路线/型号 | 已有核验 | 最小后续动作 | 当前付费范围 |
| --- | --- | --- | --- |
| TypeSafe / Jev 1.14 | 目录投影一致；原件/审计完整；当前公共 Classify 回放 12/12 | 审阅 25 个保守来源差异，区分聊天专属文件、Provider 分支、注释和实际共用路径；用具体反例收窄消费者，再取得标准完整资格回执 | 尚未证明需要新厂商事实，不列入付费 |
| 智谱 Responses / GLM 5.3、5.2 | 目录投影一致；原件完整；当前原生 wire/Result/Observer 检查 22/22 | 审阅 35 个来源差异及当前 Responses 请求/历史/用量分支；需要时追加对应当前公共捕获回放。不能仅用文件哈希漂移要求重买 | 尚未确定，不默认全 22 项重测 |
| Z.AI Responses / GLM 5.3、5.2 | 目录投影一致；原件完整；当前原生逐场景检查及内存捕获回放 18/18 | 审阅 30 个来源差异，说明旧请求与当前语义是否等价，再修正或满足来源适用性检查 | 尚未确定，不默认全 18 项重测 |

这三项当前完整 releasegate 消费仍 exit1，不能把场景检查 PASS 当作正式资格 PASS。下一步应处理离线来源判断，禁止为得到绿灯直接删掉来源校验或静默放行共用产品变化。

## 已确认缺少的新增真实场景

| Provider | 协议路线 | 缺少场景 | 已有事实可复用 |
| --- | --- | --- | --- |
| moonshot-cn | kimi-cn-chat | kimi-k3/xai-origin-history | 原 9/9 候选及本次当前场景检查 |
| moonshot-cn | kimi-cn-responses | kimi-k3/xai-origin-history | 原 9/9 候选及本次当前场景检查 |
| moonshot-cn | kimi-cn-messages | kimi-k3/xai-origin-history | 原 9/9 候选及本次当前场景检查 |
| moonshot | kimi-intl-chat | kimi-k3/xai-origin-history | 原 9/9 候选及本次当前场景检查 |
| moonshot | kimi-intl-responses | kimi-k3/xai-origin-history | 原 9/9 候选及本次当前场景检查 |
| moonshot | kimi-intl-messages | kimi-k3/xai-origin-history | 原 9/9 候选及本次当前场景检查 |

**最小事实缺口为 6 个场景，每条路线 1 个；不是原 54 项失效。** 缺口来自当前 grokhistory 合同增加的外国原生历史与工具 ID 分配覆盖，不是 candidate 标签、间隔变化或目录全局 Hash 漂移。原九项没有 xAI 来源输入，不能据此推导新增场景真实通过。

执行方案需先解决事实缺口与现有入口的粒度差异：当前 TestRegionalCandidate/正式入口按每路线完整 10 项封存，MaxCalls/MaxAttempts 为 10、MaxRequestedOutputTokens 为 40960、MaxReservedInputTokens 为 131072、每调用 120 秒、进程 1800 秒。直接使用现有完整入口的六路线预算上限是 60 调用/60 HTTP，不能把“最少缺 6 场景”冒称为已有可执行的 6 调用方案。 TestMain 还校验精确入口 `^TestRegionalCandidate$` / `^TestRegionalFormal$`，不能通过子测试过滤绕过完整选择；参见 [main_test.go](../../ai/live/main_test.go) 和 [regional_harness_test.go](../../ai/live/regional_harness_test.go)。

如要只补 6 场景，先给出可审阅的最小单场景采集及只读组合消费设计，明确每份旧/新测量独立授权、来源、型号、场景与预算，旧报告不改写；评估节省 54 调用是否值得必要的工具改动。不得为此新建续期控制器、跨轮调度器或回溯账本框架。本轮未实现该能力，也未发放新授权。

在具体方案获准前，不创建可执行的新 grant，不借用旧 ID、旧余额或未使用的旧 grant；不自动安排六路线 full formal。原 MaxCalls=9 的授权保持原义。

## 已有结论与本轮之外

- OpenAI Images、OpenAI Decisions、Google Interactions Image、TypeSafe Jev 1.13 已得到新的标准适用性回执，无新增 live 范围。
- 前次 16 条可沿用结论保持独立引用，不因本次消费者/测试修改重采。
- MiniMax 国际 Messages 的原 incomplete/FAIL 仍是另一个已知缺口；本轮未诊断、补测或恢复它，不计入上述 12 条路线的 6 个新增场景。
- race/压力先核对既有原件和实际影响，再由人选择；本轮均 NOT_RUN。测试隔离修正只改变合成夹具的 Git 存储命名空间，未修改产品 claim/锁算法，不能据此自动要求全部 Provider 并发复验。
