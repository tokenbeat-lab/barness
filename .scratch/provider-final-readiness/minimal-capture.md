# 最小补采范围与下一步

本清单只描述缺失事实，不是授权。所有原报告、授权、账本、失败和暂停状态保留；本轮厂商调用、付费grant、真实claim、race/压力、目录激活及暂停恢复均为0。

## MiniMax 国际 Messages

原测量：`.evidence/barness-ai-grok/live/minimax-messages/20261027T000249.194767Z-krdog6id/bundles/20261027T000249.211139000Z`，仍为formal/live、complete=false、FAIL；原进程exit1。15条Call记录、14个唯一场景；control初次失败已有合法rerun恢复，不增加第二个缺项。

当前既有原生逐项检查对13个成功场景全部通过：full-stream/full-complete、simple-stream/simple-complete、thinking-control-rerun-1、tool-call/tool-result-native、truncate、cache-first/cache-repeat、cancel、preflight-error、xai-origin-history。完整封存/审计、原授权在原时刻有效性、权威ledger、固定plan、当前来源/SDK和实际M3目录投影均通过。此为逐项适用性分析，不是新增真实测量或完整资格；完整消费者仍拒绝该包。

最小新增事实为**一个thinking-on成功场景**，身份仅`minimax / anthropic-messages / MiniMax-M3 / standard`、国际官方Messages端点及原Host账户/tenant。保持原CRT请求、thinking.type=adaptive、stream=true、max_tokens=2048；当前编码后的wire与原308字节请求语义相同。应保留非空自身Thinking、成功可见文本和message_stop/end_turn、完整用量/Result/Observer/原始响应、资源释放、实际进程退出、新授权及账本。仅收到HTTP200或部分推理不能通过，不降低成功终态门槛；流开始后不自动重放。

受控边界包括：同一793字节合法前缀后注入非EOF读错得到transport/stream；普通EOF缺终态得到protocol/stream；补齐合法结束得到成功；初始request EOF得到transport/request。四项均核对当前请求wire、准确归类、原部分推理、单次Attempt/不重放、唯一终态、Observer与资源归零。注入错误串是合成原因，原thinking-on的底层read error没有留存，不能据此认定厂商/代理/SDK/网络某一方有缺陷。

**执行准备仍有一项缺口**：当前MiniMax标准入口固定14项，尚无MiniMax专用单项采集/13+1组合合同；本轮不新增该能力，不以子测试过滤绕过完整合同，不套用Kimi专用消费者。下一项实施应先评估固定单场景补项所需最小改动及费用，再取得单独授权。MiniMax付费调用上限、USD和执行二进制尚未核定；不把“缺一个成功场景”写成已有可执行的一调用授权，也不默认重跑13个成功场景或整组14项。

## Kimi 六份待批准范围

见[kimi-scopes-v2](kimi-scopes-v2)中的六份`KIMI_PAID_SCOPE_PROPOSAL_NOT_A_GRANT`及[不可变索引](../../.evidence/provider-final-readiness/kimi-pending-scopes-v2.json)。它们不含可执行的顶层RegionalGrant schema/ID/有效期，也没有人类批准或claim；每份绑定真实旧baseline manifest/report/seal、已适用九项、当前native source/SDK/候选目录、留存正常live-tag二进制SHA，以及需要Host确认的自身账户/地区/tenant/account scope。

| 范围 | 每路线 | 六路线合计 |
| --- | --- | --- |
| 场景 | 仅kimi-k3/xai-origin-history | 6个新真实场景 |
| calls / HTTP | 1 / 1 | 6 / 6 |
| SDK/环境重试 | 0 | 0 |
| 输入预留 / 输出 | 4096 / 4096 | 不跨路线借余额 |
| 拟议USD上限 | 0.10 | 0.60 |
| 调用 / 进程 | 120s / 240s | 串行原生进程上限24分钟 |
| 首发等待 | 60s | 串行，不新增跨进程调度器 |

价格依据冻结kimi-prices-2026-10-22.1，非实时账单。未来明确批准后才可为各路线签发schema2、purpose=xai-history-supplement的新grant；建议短有效期30分钟、最长24小时，绑定具体source/build/binary/旧baseline。失败、取消、限流、超时或原生拒绝即停止剩余付费工作；不得自动重试、涨预算、重绑/续期旧授权或恢复旧暂停轮次。

六份只有提案，不请求用户本轮批准。执行前仍需Host确认真实standard API账户、地区和身份范围，以及所用价格/SDK/model/source/input仍符合提案。Kimi USD0.60不覆盖MiniMax、xAI请求、完整十项重跑、目录激活、race/压力或任何暂停轮次。

## 29条路径的交付顺序

[矩阵](delivery-matrix-v2.md)按操作×Provider×API去重，TypeSafe1.13/1.14同一路径保留各自证据。22条完整适用、Kimi六条9/10、MiniMax一条13/14；已适用的其他路径不因本轮补测试失效。

1. 当前离线准备完成后，下一次明确授权才能执行Kimi六项；分别只读组合各自旧九项与新一项，保存两次真实来源/时间/预算和传递原件。
2. 单独审阅Kimi自身正式目录晋升；六组真实十项完整且实际型号事实一致时只读取得正式资格。候选stage不改，不另买六组十项formal。
3. MiniMax按上一节完成最小执行方案和独立授权，再补thinking-on成功事实；不得恢复旧Grok17路线批次或借未用grant。
4. 汇合当前适用性与新增原件，按实际并发/资源影响列人工race/压力候选。只读/文档变化不新增全Provider范围；已有产品并发/负载阶段仍待选择。
5. 建立新的当前交付/发布record，保留旧记录原义。缺新真实事实、目录或人工阶段时release.pass=false，不以适用证据数量冒称所有Provider已经交付。

本次准备时发现并修复Kimi producer对source.files/selectedInputs分栏的遗漏：两份README原件均齐全，精确两域一致性校验已补正。公共CLI先失败后通过，六份真实旧producer全部离线通过。新提案v2绑定修正后重建的正常live-tag二进制和当前原生来源；v1提案保留、不授权、不执行。此为消费者修正，不要求重采旧54项或其他Provider。
