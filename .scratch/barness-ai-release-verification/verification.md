# Decisions / normal / race：当前发布续修验收

执行起点 `b61c6a5`；本次只完成用户授权的受控修复与门禁，真实服务/效果保持 NOT_RUN。
源码修复：`4e4d99d` Decisions 清单、`f7c3aff` deadline 归属、`cc78cfd` 夹具发送缓冲。
诊断归档提交 `f29703b`。本版完整命令在该提交启动，之后没有修改 Go 文件。

## 修复与因果

Decisions 的零凭据分支曾在创建场景 case 前整体 Skip，报告声明九个场景而清单缺项。
现在逐场景保留未执行原因与 NOT_RUN case；不构造请求/结果/Observer、不消耗预算。
严格发布消费者没有放宽；公开 formal-process 的未执行包先 RED，再修生产分支得到 GREEN。

unary 的原 single-P race 第162轮实际重现：call deadline 已过，但取消回调尚未传播，
后到的 read-idle timer 先写下 timeout 原因。时序 overlay 捕获 deadline 已过3,918,125ns。
公开 Client/HookedClient E2E 用合法 Host Context/AfterFunc 有界延迟形成两个 RED，
并保留两条真正 read-idle 更早的反向对照。运行时只为严格早于调用截止的尝试限额计时，
正常/真正读闲错误语义不变。当前回归及原单P独立 race 300轮均通过。

原压力设计自然重现14条流截断；两条连接的实际读取比 handler 报告写出多65,536 /
697,522字节。历史 ENOBUFS 不安全续写会重复已经入队但未报告的字节，不能据返回0续传。
两轮独立普通TCP竞争缓冲诊断均定位解析前首差 body1,318,335 / frame4098 offset262,214，
并与同连接seq194、wire offset1,060,329、783,784-byte write、ENOBUFS/0-reported、
同Hash续写关联。不注入 errno/正文、不改原32并发、24条64K-token文本、8条256KiB工具流。
当前实现保持错误立即失败，并只在默认受控 TCP 夹具设置64KiB发送缓冲，形成有限内核背压；
没有修改生产SDK的预算、认证、限额、重试或比较，自定义故障listener保持原语义。

同竞争条件的安全对照实际 FAIL：22显式错误、10成功，32handler/trace，每请求一次Attempt，
零首差、资源0、audit0。最终归档脚本正常路径又独立得到18错误、14成功，仍为实际FAIL；
故意耗尽不是生产PASS。公开旧续写 RED 重放仍失败，当前隔离回归70case通过。
原设计单跑与十轮压力均通过。全部原始记录、未复现回合和部分失败保留。

## 完整执行

```sh
python3 .scratch/jev-1.14-multimodal-box/run-14-verification.py .evidence/release-followup/final-f29703b --gate
```

主命令通过环境白名单清除路线凭据；没有读取.env或进行真实服务/收费调用。
运行两种vet、11条独立路线预检、两文本及两正式视觉效果包、四个独立verifier、
完整live-tag harness、完整normal/race、冻结Pi差分、原压力负载和当前图片设计负载。

最终原始报告为`.evidence/release-followup/final-f29703b/gate/release-report.json`，
SHA-256 `085add91f8ec6c2cefe54c74770298164a7cbd44e86b64252c9d1f9e7f5ee313`。

| 验收 | 实际结果 |
| --- | --- |
| 完整normal | 命令成功，493.945s，9,098 PASS / 0 FAIL，audit0 |
| 完整race | 命令成功，1,849.728s，9,098 PASS / 0 FAIL，audit0 |
| 两种vet、P0、冻结差分、聚合审计 | 六项受控gate均PASS，17个发布输入审计0发现 |
| UC14 / Pi | 两阶段各33 / 861 PASS |
| 原cloud压力 / unary deadline失败场景 | 两阶段均PASS |
| 当前混合图片负载 | 两入口×两阶段均PASS；各4调用、双租户/协议、原64MiB限额、资源0 |
| live-tag主harness | 204 PASS / 62 NOT_RUN，audit0 |
| 真实路线 / 效果 | 11路线、4效果包均NOT_RUN，真实Calls/Attempts0 |

600个Go文件Hash与cc78cfd固定集合一致。完整runner耗时2,491.544s，未超时；
发布CLI实际exit1、收集脚本exit0，不能把后者写成发布PASS。

整体仍BLOCKED。原报告live、traceability和required-artifacts因真实/效果NOT_RUN阻断；
snapshots还保留启动时原#07未关闭的真实快照。Decisions逐场景清单已被严格消费者接受，
没有再出现清单完整性失败。独立Standards/Spec均无剩余发现；#07因果、回归及完整本版
受控条件已满足，已在报告结束后将其自身及续修01/02转换为resolved。
转换前后原文、Hash、时间和依据见`.evidence/release-followup/pressure-tracker-transition.json`；
原report.pass=false和旧snapshot不变，验收后状态不能追认该次报告PASS。
真实支持/效果未执行，本次没有整体发布PASS或部署声明，整合票14保持needs-info。

最终回执`evidence.json`通过`seal-evidence.py --record`建立；从仓库根目录执行
`python3 .scratch/barness-ai-release-verification/seal-evidence.py --verify`可独立复核
完整源文件集合、实际命令/门禁、主包清单/审计、原始存储字节及精确目录覆盖。
两个故意中止的live-harness子包仍auditMISSING，作为子进程历史保留，不计完整通过。

## 历史与可核验材料

- `.evidence/release-followup/`：本次全部原命令、专项、压力历史和最终门禁；
  `history-summary.json`保留实际FAIL及原型audit2发现，不能写成全历史零发现。
- `.evidence/pressure-integrity/native-20261017T022612.488595Z` 与
  `native-20261017T022622.224455Z`：两个独立自然RED。
- `native-20261017T022910.882286Z` 与 `native-20261017T023345.189293Z`：安全对照实际FAIL。
- `.evidence/pressure-integrity/red-replay-20261017T023049.132600Z`：既有两个公开RED重放。
- `investigation.md`、`native-pressure/README.md`与`code-review.md`：失败清单、限制和两轴复核。

早期竞争实验重叠、强条件中止、冷编译/设置超时、编译失败、busy-body原型没有业务RED，
以及竞争未结束时过早启动的回归FAIL均明确保留，不计有效独立PASS。
不声称反向认证历史call-04/call-11原始字节；本次证明相同原负载下可重复的自然续写损坏机制。
