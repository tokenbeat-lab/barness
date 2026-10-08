# 工单 18：压力回环中断定位与修复

已确认一次同负载、同症状自然失败的直接原因：回环 fixture 的 TCP Write 返回
ENOBUFS，handler 提前结束，使公共 Client 读取到 `unexpected EOF`。修复在测试连接
边界限时等待并续写未发送字节；产品、原负载、预算和失败断言保持不变。
另以先红后绿回归修复了证据 Run 逐轮保留已结束 fixture 的内存问题。
所有运行均使用合成回环 Provider，没有真实厂商调用。最终普通与 race 各 4,461 用例
通过，同进程 100 轮及 10 个独立进程通过。

## 因果证据与边界

[自然失败完整包](history/write-buffer-interruption.tar.gz) 保存独立进程第一轮 32 路中
15 路中断：15 个客户端 read error 都是 `unexpected EOF`，当时 context 未取消、
BodyClosed=false；按同一合成 requestId，全部对应 fixture 的 TCP Write
`no buffer space available`。32 个 handler 的完成记录齐全，计划 End 均为正常结束；
fixture 返回写错后关闭连接，客户端才看到截断。该轮仅 3.20 秒，未达到 3 分钟截止。
因此本轮不是调用取消、客户端提前 Close、计划 abort 或前序套件残留造成的失败。

该错误对应 ENOBUFS：系统无法分配内部缓冲或网络输出队列已满。
[Apple send(2)](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/send.2.html)
说明这一错误语义；[Go 1.26.2 FD.Write](https://github.com/golang/go/blob/go1.26.2/src/internal/poll/fd_unix.go)
会等待 EAGAIN，但直接返回 ENOBUFS。fixture 的 HTTP writer 保留首次写错，不能在
ResponseWriter 上重试已经失败的 flush，所以处理位于它下方的 net.Conn。
证据确认了此次截断的直接原因，没有测量宿主内核为何在该时刻耗尽缓冲。
旧工单 17 九路失败没有对应底层记录，不能将其逐路追溯为相同 errno。

`write_buffer.go` 仅针对 ENOBUFS：每次 Write 最多等待 1 秒，以 1 毫秒间隔只续写
未写入的后缀，保留部分写入进度；其他错误直接返回，Close 唤醒等待，底层写截止
仍然生效。持续 ENOBUFS 仍使原压力用例失败，不自动重跑用例、HTTP 请求或 AI Attempt。
等待次数按 handler 记录到 `provider-outcomes.json`，不会吞掉诊断信号。
七个公共 Client E2E 分支先出现 4 FAIL、3 PASS，再修复为 7 PASS，覆盖瞬时/部分写入、
持续耗尽、其他 errno、关闭、写截止和调用取消，并验证只有一个 HTTP 请求和 Attempt。

客户端诊断记录 requestId、RoundTrip/Body.Read 阶段、已读字节、原始错误、当时 context
和 Body.Close 状态；fixture 记录相同 requestId、计划 End、已写字节/flush 次数、
等待次数、写错和完成时 context。标记经公开 WithHooks/TransformHeaders 注入，
不按并发到达顺序猜测身份。原始错误仅进入脱敏审计的合成证据，产品错误与 Observer
保持原契约。fixture 到达屏障可取消，完成等待有界。

## 原负载与运行方式

每轮直接复用原 `TestPolicyPressure/cloud-interactive/design-load`：8 租户、32 并发，
每路 2 MiB 历史及两张各 1 MiB 图像，24 路 64K token 文本、8 路 256 KiB 工具参数。
3 分钟截止、2 GiB 存活堆增长预算、字节/队列限额 <=75%、完整输出、资源归零均保留。
跨轮另验证已结束 fixture 的起始存活堆不能累计超过同一 2 GiB 预算。

每轮使用新 Provider/Client/证据目录，首次失败即停止，不用后续成功覆盖失败。
独立循环排除前序套件，套件最后的循环包含前序状态，fresh 模式每轮启动独立进程。
回放命令包含整个循环以保留失败轮的前序状态。原压力发布开关不变；循环独立选择。

## 实际产物

| 归档 | 运行与判定 |
| --- | --- |
| [history/read-interruption.tar.gz](history/read-interruption.tar.gz) | 早期独立第一轮 1 路 EOF，1 FAIL；当时尚无同轮写入记录，不能归因 |
| [history/write-buffer-interruption.tar.gz](history/write-buffer-interruption.tar.gz) | 完整关联诊断的自然失败：15 路 EOF 对应 15 路 ENOBUFS，1 FAIL |
| [history/retention.tar.gz](history/retention.tar.gz) | 修复前内存回归第 14 轮红灯，13 PASS、1 FAIL，起始存活堆增加约 2,158 MiB |
| [retention-replay.tar.gz](retention-replay.tar.gz) | 临时副本重放内存红灯：审计入口修正前/后各 14 轮，26 PASS、2 预期 FAIL；最终脚本与审计 PASS，初次审计失败保留 |
| [history/isolated-after-retention.tar.gz](history/isolated-after-retention.tar.gz) | 仅内存修复后的历史 100 轮 PASS；起始存活堆约 38–39 MiB |
| [history/fresh-before-write.tar.gz](history/fresh-before-write.tar.gz) | socket 修复前 10 个独立进程 PASS；历史结果 |
| [history/full-before-write.tar.gz](history/full-before-write.tar.gz) | socket 修复前完整普通 4,454 PASS，249.249 秒；历史结果 |
| [history/race-before-write.tar.gz](history/race-before-write.tar.gz) | socket 修复前完整 race 4,454 PASS，663.790 秒；历史结果 |
| [diagnostic-regressions.tar.gz](diagnostic-regressions.tar.gz) | 早期读取/取消/关联诊断先红后绿及 live harness 元数据回归；10 包、6 PASS、3 预期 FAIL、1 空包；当时两种 vet PASS |
| [write-regressions.tar.gz](write-regressions.tar.gz) | 7 分支红/绿完整包，红 3 PASS/4 FAIL，绿 7 PASS；修复后两种 vet PASS |
| [isolated-final.tar.gz](isolated-final.tar.gz) | 最终源码同进程 100 轮 PASS，690.477 秒，起始存活堆 38.51–39.20 MiB；自然运行 BufferWaits=0 |
| [fresh-final.tar.gz](fresh-final.tar.gz) | 最终源码 10 个独立进程各 1 轮 PASS |
| [full-final.tar.gz](full-final.tar.gz) | 最终源码完整普通 4,461 PASS；原压力、冻结 pi 差分及末尾 5 轮，246.169 秒 |
| [race-final.tar.gz](race-final.tar.gz) | 最终源码完整 race 4,461 PASS；原压力、冻结 pi 差分及末尾 5 轮，667.564 秒 |

最终全量用例数为原 4,448（含冻结 pi 1.0.0 差分 623）+ 读取诊断 1 + socket 回归 7
+ 末尾循环 5 = 4,461。最终运行判定以保存的实际命令退出码及各用例清单为准。
受测工作树父提交为 `a6c163af07663fed878de3ca4f18e75eeac2bc2a`，不是改动的提交号；
最终四种压力运行均保存实际源码集合/内容哈希，由核验器匹配当前源码。
早期 read-interruption 与 diagnostic-regressions 包没有源码哈希清单，只提供所存
诊断/回归材料，不能据此核验其受测源码；其余历史包保存各自的源码哈希或重放来源。
历史包不冒充最终源码。内存修复后的历史 100 轮还在完整回放、调用标记和堆差值展示
审查修正之前；旧差值展示可能下溢，应以原始 baselineLive 数值为准。

## 独立确认的内存保留

原 evidence.Run 保存 []*Case，Case 引用 testing.T。Go 1.26.2 清理 cleanup slice
时未清零 backing array，srv.Close 闭包继续引用 Provider 及其约 166 MiB 请求捕获。
先添加跨轮公共负载回归，实际第 14 轮超过原 2 GiB 预算后，再将 Run 改为完成时不可变
摘要，不持有 Case/testing.T，并独立复制 fixture 哈希。manifest/audit 语义保留。
此缺陷妨碍长循环；自然读取失败发生于第一轮，没有证据将其归因于内存累积。

## 复核与重跑

完整归档保留原用例文件、清单、审计和已存在的进程日志/收据，tar 只去除宿主所有者
身份，不改变字节或哈希。早期直接运行没有进程收据，以保存的 FAIL/PASS 断言为证，
不补造日志。曾中止的探索运行不列为通过证据。

从仓库根目录复核，不发起厂商调用：

```sh
python3 .scratch/barness-ai-pi-1.0/pressure-loopback-evidence/verify.py
```

检查交付哈希、安全解包、包/状态数量、每个 artifact 哈希、原 fixture 输入来源、
整棵解包树脱敏审计（包括日志/收据）及最终源码匹配。审计入口添加独立汇总 envelope，
原子包 manifest 保持原字节。既有 fixtures 指脱敏前输入哈希，artifacts 指交付副本；
两者不同时，核验器匹配仓库原 fixture，并继续校验副本 artifact 哈希与审计。
完整普通/race 的历史及最终包各有 349 个这样的副本，原包不改。
[核验计划](verification-plan.json)、[实际核验输出](verification.log)、[两轴审查](REVIEW.md)
区分失败历史与最终通过。

输出目录必须不存在，脚本过滤厂商凭证和继承测试开关：

```sh
python3 .scratch/barness-ai-pi-1.0/pressure-loopback-evidence/run.py --mode isolated --rounds 100 --out .evidence/issue18-new-isolated
python3 .scratch/barness-ai-pi-1.0/pressure-loopback-evidence/run.py --mode fresh --rounds 10 --out .evidence/issue18-new-fresh
python3 .scratch/barness-ai-pi-1.0/pressure-loopback-evidence/run.py --mode full --rounds 5 --out .evidence/issue18-new-full
python3 .scratch/barness-ai-pi-1.0/pressure-loopback-evidence/run.py --mode race --rounds 5 --out .evidence/issue18-new-race
```

full/race 启用原压力和冻结 pi 差分，需要已有 oracle 依赖；必要时执行
`npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node`。模式只改变前序环境和
重复次数，每个收据保存实际命令及退出码；不更换单轮负载、预算或失败判定。

可在临时源码副本重现已确认的内存红灯，固定 Go 1.26.2，仅恢复父提交的 evidence.go：

```sh
python3 .scratch/barness-ai-pi-1.0/pressure-loopback-evidence/reproduce-retention.py --out .evidence/issue18-new-retention-red
```

原工作树不变；只有实际跨轮预算失败且审计通过才报告预期红灯成功。这会占用数 GiB
内存，只证明内存保留，不能替代 ENOBUFS 因果记录。socket 七分支回归可直接执行
`go test ./ai/e2e -run '^TestPressureFixtureWriteBackpressure$' -count=1`，无需压力开关。
