# MiniMax 缓存计费与终态封存当前状态

2026-10-28：实施、受控与完整普通验证已完成。先复现single/full后验拒绝时report与seal不一致，再将最终FAIL/audit写入移到最终seal之前；新进程退出后原件只读。仅国际Messages/MiniMax-M3/standard固定thinking-on单项、wire无cache_control时独立缓存写费NOT_APPLICABLE。原creation缺失保持缺失，UsagePartial保持partial，不补零、不修改全局SDK/usage。显式null、正写入、TTL、其他计费项缺失、显式缓存控制与其他身份仍拒绝。单项inspect和固定13＋1组合分别输出当前计费适用性，不改两份原测量。

[完整验证](../../.evidence/minimax-cache-billing-finish/verification-v1.json)、[旧两次捕获只读判定](../../.evidence/minimax-cache-billing-finish/old-capture-applicability-v1.json)、[保全](../../.evidence/minimax-cache-billing-finish/retention-v2.json)。本文件为本功能唯一当前状态；既有功能状态及全部历史版本保留。

受控24项有限必需case通过，含新12项封存/计费场景、固定13＋1及原11项单项失败边界。完整普通八包实际退出0，主库存{'PASS': 15357, 'NOT_RUN': 214}，回执.evidence/minimax-cache-billing-finish/e2e/20261028T030752.619882Z-7o4zhdo3/receipt.json，墙钟725.697s；所有辅助库存/audit通过。来源2051文件、选定输入187文件与当前/冻结精确相等。normal及live标签compile-only vet退出0。MiniMax正常冻结二进制在完整普通中由受控真实子进程执行，另冻结Kimi及只读工具供后续审阅，不执行厂商请求、不产生新scope或授权。

受测/当前HEAD均69ff4af880e68ef3ca245e5128bdae5e59a2ef7b，既有未提交修改保留；本轮源码差异见verification-v1，SDK/产品ai根目录、依赖和Kimi目录字节未改，各文件小于500行。ADR-0055、SOP和工具合同同步修订，前版本字节保存在initial；没有通用框架、历史兼容入口或历史重封存。

两次旧捕获均通过当前wire/SSE回放、路线来源/SDK、原授权与预算资源检查；当前判定为独立写费不适用、creation未报告、usage partial。可复用为回归及业务观察证据。原两次producer均FAIL/exit1，最终report与旧seal不一致，原封不动保留；当前VerifyThinkingSupplement和成功producer检查均拒绝，正式13＋1不能直接复用。缺口是合规成功结束并完整封存的thinking-on单项，不是其余13项。若后续选择补采，只需另行批准MiniMax这一项1调用/1HTTP/USD0.01，新grant与有效期；旧两份预算和claim不恢复、不合并、不重新封存。Kimi六项及其他Provider资格不因本次消费者修正而自动失效或重测。

历史扫描213269＋8387项无丢失或未解释变更；8162个公共授权/账本及95个旧付费阶段文件集合和字节相同。旧七份v5事实确认修订1及更早scope索引均保持，不覆盖为可执行授权。本轮新增付费授权/真实claim/厂商调用均0；live/race/压力未执行，Kimi active=false，暂停轮次未恢复。race/压力在本次收尾顺序和固定只读规则范围内无新增共享状态/生命周期/负载变化，NOT_APPLICABLE，不替代旧功能阶段。

完整Provider交付仍未成立：MiniMax尚缺可正式消费的thinking-on成功单项，Kimi六路各缺一个真实xai-origin-history补项。当前实现验证通过与全部Provider发布通过分别记录；只读release记录保留live NOT_RUN并应拒绝完整发布。回滚只撤本轮源文件/文档差异并停止使用新冻结工具，保留全部历史证据、授权、账本及新旧失败。
