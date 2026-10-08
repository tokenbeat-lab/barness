# 工单 15：混合操作验收计划

基点：`211ab89d8a60e1741efb8d25e39c39d20aa6b520`。测试 seam 已由工单指定：
公共 Client 的 Complete/GenerateImages/Classify、正式本地/云端宿主入口、
宿主 Binding/Credential resolver、HTTP Provider、Admission、Observer 与既有资源探针。
不新增核心接口、跨操作 fallback、兼容 shim 或依赖。

## 编码前失败清单

1. 跨租户同名 Binding、同型号 ID 跨操作：请求、账户、答案、图片、用量或 Observer 串归属。
2. 同账户共享 CredentialRef 的 chat/image Binding：错误扩大操作授权，或误拒合法共享。
3. 旧零值 chat Binding、未知操作、跨入口、错误协议选项：读取凭据或发送请求后才拒绝。
4. 解析期间轮换/撤销：自动重解析；重试改用新 key、端点、账户、操作或 Provider；新调用仍用旧快照。
5. 取消、截止、响应失败、重试：释放别人的许可、提前释放自身许可、污染其他调用结果；
   未读完响应或未完成解析就允许下一次请求。
6. 混合准入：突破租户/进程上限，无界等待者、等待超时不清理，取消误释放许可。
7. 修改构造输入目录、子策略、Binding 切片或调用输入/选项：改变 Client 或在途请求；
   参考图、mask、问题 map/raw JSON、结果用量发生可变别名。
8. 已知超大输入先复制/base64/JSON 编码，证据/Provider 副本主导被测内存。
9. 输出大于 SSE 帧上限被误拒；单图、总图、响应、数量超限半发布；失败/取消残留 body/许可。
10. Observer 泄漏提示、base64、分类状态/答案、错误正文；拥塞、慢消费者、错误或 panic 改变结果。
11. 禁用 TypeSafe、OpenAI image、Google image：新调用仍能发送、影响其他操作，
    或错误承诺撤销已发请求/丢弃尝试消耗。
12. 宿主把请求自报 tenant 当身份；本地共享凭据跨账户/Provider；业务阈值由库擅自决定。

## 设计与实施切片

- 先验收有限混合策略，再添加独立 local multi-binding assembly 与 cloud unary 宿主入口。
- 正式策略专门按混合负载缩小并发；旧聊天策略的容量不自动成为图像默认值。
- 同一个 Client 配置四条真实协议路线（chat、OpenAI image、Google image、TypeSafe），
  使用按租户/Binding 路由的 Provider 脚本；公共 E2E 逐项留证。
- 压力由已有 BARNESS_AI_PRESSURE=1 启用，默认 NOT_RUN；生成参数与长度/哈希代替大正文证据。
  客户端 Body.Read 在交付所有响应字节后暂停 EOF，记录完整 body 和已发布结果的 GC 可达堆；
  调用期 TotalAlloc 保守覆盖采样间暂存副本，Provider 捕获副本、JSON/base64、吞吐与用量占比均记录。
  字节限额占比 <=75%，并发按设计占满，明确本地 256 MiB/云端 512 MiB 混合缓冲预算。
- 修订 ADR-0002/0003/0008/0009/0017、示例说明与 T/C/H 追溯，保留审计、race 和完整回归证据。

## 验证

每切片运行定向 E2E/Go 编译；最后 go vet（含 live）、开启差分/压力的完整测试，
混合操作 -race；执行 code-review 的 Standards/Spec 两轴并修复发现后提交当前 main。

## 实测收敛

审查前仅 GC live 指标不足以覆盖短命副本；新增调用期 TotalAlloc 上界后，32 KiB chat
输出触发 284/567 MiB 总分配红灯，超过 256/512 MiB 预算。保持图片输入/输出和 4/8
并发，将 chat 输出设计负载降为 16 KiB；新边界、每项限额余量和最终实测见
[证据包](mixed-operations-evidence/README.md)。独立租户请求/正文/图像/答案/用量及
全部 Observer 快照断言已补齐，实现审查 Standards 1 项、Spec 2 项已修复；交付脚本审查再补齐真实命令记录、
源码提交与默认开关 run 关联，累计 Standards 3 项、Spec 5 项全部修复并独立复核。
