# 工单 17：九组合发布门禁与完整证据

2026-10-08 完整门禁 **十项全部 PASS**，退出码 0，受测提交
`0dca06c003c7d30a43c44aef6bf961ec6f5e2dbe`。普通与 race 均执行 4,448 个用例：
离线 3,825、pi 1.0.0 差分 623、pending 0；两种 vet 通过，76 条需求追溯全部 PASS，
十二个完整证据包脱敏审计零发现。普通测试命令耗时 220.92 秒，race 553.47 秒。
交付目录审计和 [便携复核](verification.log) 均通过：安全解包、逐文件哈希、十三个
完整包（含失败历史）重新审计、独立效果核验与七个证据门禁重算一致。

门禁只消费证据，九条真实路线各在自己的进程中执行，单次子进程只收到本组合 key。
本轮六条聊天、TypeSafe 分类、OpenAI Images 和 Google Interactions 图像均取得自己的
完整报告；DeepSeek 两行在修正历史探针残留后再次独立重跑，并通过 CLI 原子合并。
账户别名明确标记区域和权限未确认，不能据此宣称专用低权限账户已获认证。
全部必交能力通过；OpenAI Chat 的 reasoning-history 按协议明确 UNSUPPORTED。
OpenAI mask 本轮实际通过。TypeSafe 有界上下文探针实际返回 400，422 未确认。
九个当前报告共 62 场景（61 PASS、1 明确 UNSUPPORTED），87 次调用/HTTP 尝试、0 环境重试。

## 阅读与复核

| 材料 | 内容 |
| --- | --- |
| [release-report.md](release-report.md)、[JSON](release-report.json) | 十道门禁、实际命令结果与耗时、协议/live/业务效果分别呈现、完整需求追溯 |
| [manifest.json](manifest.json)、[verification.json](verification.json) | 完整包路径/哈希/用例数、受测提交与各 live harness 构建来源 |
| [support-matrix.json](support-matrix.json) | 九组合的 operation/provider/API、当前型号/SDK/能力/时间；自身完整报告合并结果 |
| [catalog-snapshot.json](catalog-snapshot.json)、[fixtures.json](fixtures.json) | 正式目录和价格内容哈希、来源与抓取日期、fixture 逐文件完整性 |
| [ledger.json](ledger.json)、[traceability.json](traceability.json) | 已批准差异、扩展 skip、P0/P07–P09/T/C/H 实际证据选择 |
| [pressure.json](pressure.json)、[rollback.json](rollback.json) | 本地/云端实际混合设计负载与资源归零；三路线独立停用、轮换和旧聊天 Binding 兼容证据索引 |
| [adr-review.json](adr-review.json) | 所需 ADR 修订位置与完整内容哈希 |
| [source-hashes.json](source-hashes.json)、[sha256.json](sha256.json) | 全部受测源码集合/提交/内容及交付材料完整性 |
| [FAILURES.md](FAILURES.md)、[REVIEW.md](REVIEW.md) | 实现前误放行输入与两轴审查修复/复核结论；各 red-*.log 保留实际红灯 |

十二个 `.tar.gz` 保存离线、race、九条 live 和独立中文效果的完整原包。
不删请求/响应捕获、消费结果、Attempt 消耗、Observer、断言、图片或原始清单；
tar 只去除宿主所有者身份，原文件字节与哈希不变。逐个用例 artifacts 可重验。
`live-commands/` 保留各进程及合并日志；归档清单保留原仓库相对路径与实际构建提交。

从当前干净 checkout 根目录执行，不需要厂商 key 或再次发起真实调用：

```sh
python3 .scratch/barness-ai-pi-1.0/nine-route-release-evidence/verify.py
```

核验器先检查完整源码集合、受测提交和所有哈希，再安全解包、逐文件校验、重新审计
十二个完整包及失败历史，验证独立中文报告，并重算七个证据门禁。
`releasegate -bundle` 不执行新命令，因此其三个命令门禁必须为 NOT_RUN；
记录的十道 PASS 来自保存的完整真实命令运行，不能把回放当成新执行。
拒绝 Python `-O`，避免完整性断言被禁用。

重跑全部真实路线和完整门禁需要仓库 Go 工具链、Node.js 与冻结 oracle 依赖：

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
python3 .scratch/barness-ai-pi-1.0/nine-route-release-evidence/run-release.py \
  --live --out .evidence/<fresh-run>
```

脚本为该运行构建独立 harness，逐组合调用 `run-live.py`、合并自身报告后启动
不持有厂商 key 的门禁进程；不覆盖已有输出。`--live` 会消耗有界真实账户预算；
没有 key 或没有启用时明确 NOT_RUN，不能用历史通过替代当前运行。
官方端点返回具有不确定性，重跑结果据实际报告判定。

## 基线、压力与业务范围

唯一 oracle 固定 pi-ai 1.0.0、commit
`a13d35a742c6ef8462812a28fbe1d8c8b7431c32`、模型数据哈希
`8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e`。
[此前发布件等价证明](../release-verification/evidence/report.json) 为 835 个 dist 文件零差异；
本轮全量差分由该唯一 oracle 执行。TypeSafe 可表达子集具有差分，宽 JSON 等扩展以及
DeepSeek Responses、两条原生图像的独立离线证明与明确 skip 分列，不冒充 pi 支持。
正式目录/价格版本为 `2026-10-08.5`，内容哈希为
`sha256:824422bc176c76054cefdbe517af3301f448ca51a3418ead35bdea2134b62406`；
当前九个自身报告均与该快照一致。

本地/云端混合设计负载分别占满 4/8 并发，缓冲增长预算为 256/512 MiB；
GC 采样下界和调用期 TotalAlloc 保守上界均须在预算内，每项字节限额用量 <=75%，
base64 膨胀与资源释放由实际报告证明。报告是合成协议负载验收，不代表厂商像素计算容量。

本轮普通完整套件的实际读数（受测提交 `0dca06c`；精确时间不作跨机器判据）：

| 策略 | 峰值许可 | GC 捕获堆增长 | 调用期分配上界 | 预算 | 吞吐 |
| --- | ---: | ---: | ---: | ---: | ---: |
| 本地 | 4 | 9.33 MiB | 135.56 MiB | 256 MiB | 67.95 次/秒 |
| 云端 | 8 | 22.00 MiB | 271.27 MiB | 512 MiB | 119.42 次/秒 |

两项负载的十一类限额均 <=75%，全部六类资源读数归零。
十一项独立回退/轮换/零值聊天绑定证据均 PASS；完整断言和实际请求仍在离线原包中。

独立中文包保留固定合成任务 24/24 正确、四类各 6/6、24 次调用/尝试与完整校准统计。
它不设发布准确率阈值，不代表生产总体或通用中文准确率；协议通过与此业务结果分别核验。
任务、固定数据集、消费结果、Observer 与真实 COMPLETE 报告均保留在 `chinese-effect.tar.gz`。

## 失败历史

[history/1/release-report.json](history/1/release-report.json) 保留此前完整门禁 FAIL：
`E08-policy-pressure-cloud-interactive-design-load` 的 32 路中 9 路发生读取中断；
当时 DeepSeek 矩阵还残留已删除探针，live 门禁同样拒绝。完整失败离线包和审计保存在
`history/1/offline.tar.gz`，不会用它证明当前通过。
同压力场景独立进程定向执行连续三次通过，完整诊断包与零发现审计在 `pressure-diagnosis/`。
该历史包缺少底层逐路记录，后续诊断见 [工单 18](../issues/18-pressure-loopback-interruption.md)；
没有放宽预算、负载或失败判据，也不据定向通过宣称原因已修复。


## 工单 18 后续诊断（2026-10-08）

[压力回环证据](../pressure-loopback-evidence/README.md) 保存一次带完整关联的自然失败：
15 路客户端 unexpected EOF 全部对到 fixture TCP Write ENOBUFS，客户端当时未取消
或 Close。直接原因在回环 fixture 的写入边界，修复对 ENOBUFS 限时续写未发送后缀，
不重放请求/Attempt；七分支公共 Client 回归先红后绿，另修复证据内存保留。
原压力负载、预算与失败判定不变。最终源码 100 轮、10 个独立进程、完整普通
与 race 各 4,461 用例 PASS；14 份归档共 43 包哈希与脱敏审计通过，工单 18 已解决。
旧工单 17 没有底层逐路记录，不能追溯其
九次失败的具体 errno，也不据此覆盖旧失败历史。

本节只更新说明及 README 哈希，原失败/成功包、门禁报告和九组合 live 结论仍属于
原受测提交。本次未发起真实厂商调用，不能作为当前源码新的完整九组合发布门禁。
旧 verify.py 仍要求原源码快照，当前新增测试使它拒绝当前源码；历史复核应使用源码
匹配的 checkout（例如本次父提交 a6c163a）。本次最终源码/归档使用新目录 verify.py，
新的离线通过不能替代 live 认证。
