# 脱敏真实捕获的公共 Client 回放

Type: task
Status: claimed
Blocked by: 10

## Contract

[spec](../spec.md) R21/R24、E17；[SOP](../../../docs/agents/verification.md)。仅在现有公共 Client 与受控 HTTP Provider 接缝回放本路线真实捕获，不授予正式资格。

## Failure ways and public seam

- 回放夹具或来源索引缺失、文件哈希/来源/场景错配必须失败；完整集合必须覆盖18场景24响应。
- 不完整终态、正文/原生推理丢失、工具数量/参数或终态变更、用量双计必须被公共返回断言发现。
- 持久化恢复后密文条目、call/item配对或图片顺序改变必须被下一次实际请求与原捕获的比对发现。
- 密钥、HTTP头、cookie、账户/会话/trace标识和原密文不得进入源码；替换保持跨响应/续接的一致性。

原公共库行为已由实际18/18 live PASS证明。本工单先建立回放消费测试，记录尚无脱敏夹具的真实assembly RED，再接入来源可追溯数据；不虚构生产协议故障。

## Answer

24份成功真实响应已脱敏并接入18场景公共Client回放，另有来源完整性case；新增2个顶层入口、19个必交case。实际assembly RED后，最终定向GREEN/exit0、19case PASS、audit0；[证据指针](../evidence/real-capture-replay-v1.json)。生成器重放25个JSON文件哈希完全一致；原密文/opaque值零残留，所有HTTP头未复制。

续接从公共Result经持久化恢复与TrustNativeState重建，工具结果使用实际返回的ToolCallID；全部请求与捕获独立比对，PNG/JPEG顺序与调用配对不丢失。缓存仅把期望中的脱敏placeholder替换为独立固定测试向量，实际请求无字段删除。

本工单不更改运行时/live harness/目录；原生source是否适用由原生规则判定，不伪造新付费证据。根代理负责独立复核、完整冻结、唯一状态和最终关闭，故本票仍claimed。
