# 工单 15 双轴审查

基点 `211ab89d8a60e1741efb8d25e39c39d20aa6b520`，初次实现 `62d6120`，
修复与复核 `817595c`。两个独立代理按 code-review 技能审查 Standards 和 Spec。

## Standards

初审 1 项 P2：服务端 Write/Flush 不能证明客户端已读完，测量边界不满足 AGENTS 第 9 条
与 ADR-0017。修复为客户端 Body.Read EOF 屏障，数据与 EOF 同返时先交付数据；
逐字节计数及 4/8 许可持有断言在完整读取后通过。

实现复核：0 项剩余问题。TotalAlloc 上界补足 GC 采样间的暂存峰值，ADR 和设计负载同步；
新增租户差异 fixture、结果及全部 Observer 断言没有规范违反或需修复的代码异味。

## Spec

初审 2 项 P2：相同租户 fixture 无法检出交换正文/答案/用量；仅测最后 GC 的 live 堆
可能漏掉读取和解析之间的短命副本。

先增加区分断言得到真实红灯，再分别引入独立请求标记、正文、PNG、分类答案与 31/47
输入用量，并比较结果、全部 Attempt 及 Observer。压力增加客户端 EOF 测量边界和
调用期 TotalAlloc 上界。原 32 KiB chat 设计触发 284/567 MiB 总分配红灯，超出
256/512 MiB 预算；保留图片负载与并发，将 chat 输出降为 16 KiB 后通过。

实现复核：0 项剩余问题。两租户归属可独立识别，暂存副本受预算约束，文档将 GC 可达堆
明确记为捕获下界。没有新增范围外行为或与规格不符的实现。

## 交付脚本复核

Standards 新增 1 项、Spec 新增 2 项 P2：命令退出码被写为常量且允许日志缺失；
输入 manifest 的测试提交与当前 checkout 源哈希未绑定。先用失败命令/缺失日志/旧提交
构造采集调用，确认原脚本错误接受，再改为读取真实 command-results.json，要求全部
必需命令退出码、运行前后提交和日志哈希一致，并核对 manifest 与逐项 Git 对象源哈希。
verify 同样检查交付日志与提交对象。实际脚本红绿验证保留在 delivery-negative.json。

随后发现默认压力证据仍指向固定旧目录（Standards 1 项、Spec 1 项残留）：已改为取本次
默认命令的 evidence_bundle，核对该 run 提交、审计、两条 NOT_RUN 与 manifest 哈希；
错误默认 bundle 的拒绝也留证。最终两位代理实际运行 verify，确认所有关联与哈希通过。

最终复核结论见 review-results.json。累计 Standards 3 项、Spec 5 项，分别记录；
公共 E2E 红灯及最终断言见 red.json 与 mixed-cases.json。
