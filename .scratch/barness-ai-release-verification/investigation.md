# 当前发布失败：续修调查

2026-10-17，起点 `b61c6a5`。用户要求修复 Decisions 清单与 normal/race 失败，
按原压力 #07 的因果验收定位，再跑完整门禁。真实服务始终 NOT_RUN、零实际请求。
保留上轮全部失败与封存回执，不改写旧证据。沿用已确认的公开 Client/HookedClient、
受控 Provider 与生产者/发布 CLI E2E 边界；不在实现后补单元测试。

## 修改前失败清单

- Decisions 声明场景缺 case、重复/额外 case、NOT_RUN 却伪造调用；报告、清单、审计不一致。
- 修复 NOT_RUN 生产分支却削弱实际执行分支、错误处理或预算保留；绕过严格消费校验。
- 调用 deadline 与读闲 timer 竞态，已到期调用误报尝试 timeout，或反向掩盖真正读闲超时；
  终态、错误阶段、原子空结果、Attempt、body/许可/Observer 与恢复调用不一致。
- 压力脚本与 socket/客户端字节不一致；字节相同但帧或工具累积不一致；请求身份错配；
  ENOBUFS 未报告已消费字节，错误续写导致重复或遗漏；显式失败被吞掉或无界重试。
- 只保留成功回合，遗失首个失败、handler 结束/连接/写入边界/资源与审计证据。
- 减并发、缩内容、加限额、重试完整场景、改变原比较或认证授权，造成虚假通过。
- 调试产生无界正文副本/日志、泄露秘密、改变内存预算，或不能正确取消并释放资源。
- 收集进程继承真实凭据、超时后遗留子进程，或将收集退出 0 误认发布通过。

## 顺序与验收

先建立原命令反馈并保留每次独立包；读请求关联的写入/读取指纹找到最早偏离。
每个确认的边界先写公开 E2E 红灯，再实施，再回到原复现。
分别报告故障注入反例和自然失败因果，不用一次专项 PASS 替代原 #07 或本次整版门禁。
修复后运行 vet、live-tag harness、原压力十轮、完整 normal/race 和零凭据预检/效果校验。
预期发布仍因真实证据 NOT_RUN 保持 BLOCKED；其他 gate 和原 #07 按实际验收判断。

## Decisions：假设与反馈

排序：①生产者在创建场景 case 前整体 Skip；②finalizer 丢弃已经创建的 case；
③严格消费者将合法未执行清单误拒。现有 `TestDecisionsFormalProcess/not-run` 扩展到
同一公开 AuditBundle 消费者后，0.07s 红灯，报 live artifact integrity verification failed。
生产代码在 cfg.notRun 分支只修改九个声明的 Note，随即 Skip，尚未调用 scenario factory；
因此并无九个 case 供 finalizer 保留。消费者要求声明与 case 一致的合同正确。
修复只在该分支逐场景记录明确原因并 Skip，不创建请求、结果、Observer 或消耗。

## unary：确认的期限竞争

原公开用例使用单 P 独立 race 进程，第 162 轮实际重现 error/ReadIdleTimeout，
所有回合分别保留在 `.evidence/release-followup/unary-single-p/`。
只记录时序的临时 overlay 第 19 轮捕获：idle 触发时调用 deadline 已过 3,918,125ns，
该 Read 已等待 102,501,041ns，而 parent.Err 与 child Cause 仍 nil。
这支持较早调用期限的回调尚未传播，排除该次读闲回调提前作用于新 Read 的假设。

早期 single-P busy-body 原型没有业务红灯，且捕获输出触发路径脱敏审计；它已从源码删除，
`unary-order-red` 是原型/审计失败，不能冒充终态回归 RED。
最终 `TestUnaryDeadlineOrder` 用合法的 Host Context/AfterFunc 有界传播延迟，经真实受控 HTTP，
Client 与 HookedClient 的较早 deadline 两例均 RED；读闲较早的反向对照两例 PASS。
统一 watchedTransport 只为确实早于调用 deadline 的尝试限额计时；正常/真正读闲分类不变。
定向 normal/race 均通过，仍须返回原复现和完整门禁。

## 压力：已捕获的字节边界

起点原命令 7.1s 再现 14 条流中断；每个 handler、requestId、connectionId、失败 socket
write 与客户端完整前缀/首个缺失字节均保留。两条自然连接实际读取比 handler 报告写出
分别多 65,536 / 697,522 bytes，证明 ENOBUFS 可隐去本次已入队进度；
`natural-boundary.json` 固定其原包位置与首个缺失帧，不能据此重试尾部。

只恢复旧续写的临时 overlay 在原负载下 20 轮没有复现，未视为修复证明。
本机有界竞争缓冲的另一诊断捕获六条自然字节偏离，未注入错误或改变脚本文字/并发：
正文偏移 1,318,335 / SSE frame 4098 offset 262214，以及 1,580,656 / frame 4099
offset 262210，JSON 结束处被重复的 `c` 参数字节覆盖。与同连接第 194 次（部分为 193）
ENOBUFS/0-reported-write 后的同 Hash 续写关联，已在进入 SDK/SSE 解析前发生偏离。
详见 `native-full-comparison/commands.json` 对应的原始 native 包。该轮两个竞争缓冲
实验发生重叠，因此仅作诊断观察，不能当独立单变量比较或生产 PASS；随后独立重复
同一单进程 128-pair 条件未复现。追加的更强系列已中止，部分证据保留，不计验收。
没有声称精确重放或反向认证历史 call-04/call-11。

当前修复在默认受控 TCP Provider 接收连接时请求 64 KiB send buffer，以普通 TCP 背压
限制夹具自身的内核队列增长；保持所有 32 个许可、24 个 64K-token 文本、8 个
256 KiB 工具脚本、认证授权、生产资源策略和全部比较不变。自定义故障 listener 保持
原故障语义，任何实际 ENOBUFS 仍立即失败，绝不恢复或重试。此值不是生产 SDK 的默认值。
原设计及十轮命令已通过；最终独立回归、完整门禁及原 #07 剩余验收仍按实际记录。

Darwin 源码佐证：TCP 在输出前入队
([tcp_usrreq.c](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/netinet/tcp_usrreq.c))；
用户设置 SO_SNDBUF 会清除 SB_AUTOSIZE
([uipc_socket.c](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/uipc_socket.c))。
源码解释不能替代上述本机字节证据。

资源竞争实验未结束时过早启动的一轮回归出现额外网络/恢复失败，保留于 `pressure-green`
及其 condition.json。已终止竞争进程，后续 `isolated-regressions` 独立执行；
不以相同目录覆盖失败，不将实验返回码或后续成功冒充那轮 PASS。
