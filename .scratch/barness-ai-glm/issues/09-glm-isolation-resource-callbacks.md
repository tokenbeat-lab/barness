# 09: 验证共享 Client 的隔离、回调授权与资源释放

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** 本地与云端 Host 共享 Client 并发使用三种 GLM 协议时，租户、账户、原生状态和缓存互不串用；回调不能越权，取消和限额会及时释放资源。

Blocked by: 03, 05, 07, 08

依据：GLM 接入 spec R02、R15、R16、R19；验收组 A11、A12、A13；User Stories 11、12、32、33、34、40、41、42、46、47、50。

- [x] 先写跨租户/账户/协议的真实工具历史、并发取消、回调越权与资源超限 E2E，再做必要修复；使用已确认的公共 Client 接缝。
- [x] 相同 BindingID、ModelID、session 名而不同租户/账户不会串用凭据、缓存 key 或 Native State；RequestID 仍由 Host 保证全局唯一。
- [x] trusted、持久化重新担保、未担保、跨账户和跨 API 状态按合同回放/降级；授权跨模型转换保留可见内容和有效工具配对，最终不留未授权引用。
- [x] header→payload→response 顺序保持；模型/Key/目标/托管工具/原生引用授权和型号硬约束都检查最终请求，具名选项、sampling 与 OnPayload 均不可绕过。
- [x] 发送前拒绝零请求/零 Attempt；OnResponse 等发送后失败保留实际 Attempt/已知用量；重试固定原配置和合法回调字节。
- [x] Admission、首帧前后取消、调用/请求时限、body/帧/输出/队列上限都经过现有策略；最终结果只结算一次，body/goroutine/许可释放可观察。
- [x] 请求/目录/选项/返回 map 与原始 JSON 的复制边界可验证；合法调用后修改不影响运行中或另一租户结果，不承诺调用方并发写正在读取的 Go map。
- [x] Observer 只保留元数据，错误与证据不泄露 Key、提示、思考、工具参数或原生状态；每个证据包通过脱敏审计。
- [x] 产出可重放隔离证据、race/资源结果和设计负载记录；保留已有 Provider 的共享生命周期语义。

**独立验收：** 同一 Client 并发完成三个协议的工具调用与取消/限额故障，按租户账户核验所有结果、重试快照与资源回收。

**依赖理由：** 跨协议真实历史、回调和重试快照需要 03/05/07 的完整往返及 08 的错误重试行为。


## Answer

已完成；公共 Client 验收接缝沿用 spec T01，未读取真实 Key、未发收费请求，未改正式目录/矩阵。

- 三协议/两型号在共享 Client 上完成同名服务绑定、型号与 session 的双租户工具往返、实际跨协议历史、原生信任/Host 重担保/降级、缓存与凭据隔离；并发请求 ID 独立。
- 修复 Chat 最终 native 内容、strict、已授权 hosted-tool 字段及未接入能力绕过；授权仍优先，发送前零请求/Attempt。安全回调仍允许已计划的 stream_options.include_usage 对照。
- 公开失败 E2E 发现宿主回调 panic 会终止进程；仅在 headers/payload/response 宿主函数调用边界恢复，按 callback_failed 结算、保留安全 error cause 并释放资源，不扩大到 adapter 内部。
- 验证 immutable 请求/目录/采样/Raw JSON/回调/事件/Result 边界、Observer 元数据、取消与已知用量、唯一终态、13 种资源/Admission/时限、六路同时持有 HTTP lease 的设计负载；48 个并发故障调用回收且 goroutine 回到相同基线。

实现提交 `6e6cf45`；后续仅证据/测试细化至 `a2031e9`。当前完整09 **359 PASS / 359 race PASS**；受影响 GLM/旧 Provider 回调和资源/Chat **2,794 PASS**；补充 race **523 PASS**；冻结 OpenAI+DeepSeek Chat 差分 **321 PASS、零 pending**；vet、脱敏和归档哈希验证通过。完成前已合入/确认最新集成 `11ef77e`。

详见[证据索引与重放](../evidence/ticket-09/README.md)、[哈希索引](../evidence/ticket-09/index.json)、ADR0005/0033。持久原始证据/日志/ZIP 位于 repo-root `.evidence/barness-ai-glm/ticket-09/`。回调 panic 的 RED 进程原样记为 INCOMPLETE，不伪造完成 manifest/audit；其他 harness/断言修正保留诊断，不计入发布通过。正式 live、价格、目录与发布门禁由10–13继续。
