# 第 1 节修复后复审

日期：2026-10-22。后续修复与最新验收以 [当前状态](../status.md) 为准；以下结论保留固定审查时点。两路独立审查均使用 **GPT-6 Astra**，分别审查 Standards 与 Spec；本报告汇总其结论并补充只读回执核验。

固定对比：`3ddc774084fe502ca1a13ab8b41b208053e24378...fce5f387e375fa3f7901700f5f525fdb64c6804f`，27 个文件。普通验收实际受测提交为 `3aafd2d41436d222449f5b5f0e8c0596ed1154e9`；其后提交仅有本功能状态、计划说明与验证索引等六份元数据变更，来源边界和选定输入均相同。

**不能确认全部解决。原七项中五项关闭、两项部分关闭；追加脱敏项关闭、日志一致性项部分关闭。合计六项 RESOLVED、三项 PARTIAL；发现两项 P1 残留和一项新引入的 P2。** 本次新增失败复现不会使已保存普通验收原件失效，但说明现有成功验收覆盖不足。冻结的 status.md 中“七项已修复”为修复轮当时的判断，本复审结论以本报告为准；未改写被不可变索引绑定的旧状态或原件。

## Standards

**[P1，原问题 3 残留] 失败原件的相对目录依赖仍可被永久删除。** [cleanup_scan.py:150](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/cleanup_scan.py:150) 在路径规范化之前，用字符串和后缀过滤引用，忽略失败 receipt 中的 `"input":"../unused"`。临时假仓库的公开 scan→plan 将其依赖目录标为 QUARANTINE，refs 为空；move→verify 成功，模拟八天观察期后 purge 成功，失败原件保留而依赖目录消失。应先按引用来源解析路径，保护落入候选范围的文件与目录，对歧义保守阻断，并覆盖相对目录及非 JSON 工件引用。违反计划的失败原件传递依赖保留和 ADR-0051 的清理边界。

复现：[脚本](20261022-section1-rereview-repros/relative-dependency.py)、[结果](20261022-section1-rereview-repros/relative-dependency.jsonl)。完整依据：[Astra Standards 报告](20261022-section1-rereview-standards.md)。

Standards：**1 项 P1**；未发现独立新增回归。

## Spec

**[P1，原问题 6 残留] 普通 chat 仍可凭零调用证据通过 live 资格与发布判定。** [record_live.go:44](/Users/cyber/RestoX/barness/ai/internal/testkit/release/record_live.go:44) 接入了原生 Merge，但该消费者缺少普通 chat 的调用、预算和固定模型等资格约束；[live_manifest.go:73](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_manifest.go:73) 又只在 Attempts>0 时要求请求原件。临时 google-gemini 六个场景仅有 PASS assertions、`calls=attempts=0`、`budget={}`、虚构型号与 endpoint，实际 `-live-bundle` 退出 0、pass=true，随后合成完整记录的 `-record` 得到 release.pass=true。应补齐普通 chat 原生资格检查，并用此公共 CLI 失败场景回归；接入现有 Merge 本身不足以兑现工具说明中“拒绝零调用”的承诺。该 true 只属于临时合成复现，当前真实发布判定仍为 false。

复现：[脚本](20261022-section1-rereview-repros/zero-call-chat.py)、[结果](20261022-section1-rereview-repros/zero-call-chat-summary.json)。

**[P2，新引入，追加问题 9 部分关闭] 消费者最终回执写入失败时，保存的日志与真实退出不一致。** [live_consumer.go:132](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_consumer.go:132) 提前写出 stdout/stderr 与 exit-code；[最终写入:155](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_consumer.go:155) 失败后，CLI 还会输出错误并退出 1。临时文件限额 2904 字节允许 stdout 写完、令最终 live-consumer.json 写失败；保存 exit-code=0、stderr 为空，实际退出 1、stderr 为 `releasegate: native live consumer failed`。正常成功与资格拒绝场景已一致，终结 I/O 失败尚未覆盖。应由外层进程捕获最终真实输出与数字退出，并保留失败记录，避免提前宣告成功。

复现：[脚本](20261022-section1-rereview-repros/consumer-final-write.py)、[结果](20261022-section1-rereview-repros/consumer-final-write-summary.json)。完整依据：[Astra Spec 报告](20261022-section1-rereview-spec.md)。

Spec：**1 项 P1、1 项 P2**；P2 为本轮修复中新引入的收尾失败问题。

## 逐项关闭情况

编号沿用七条 review 意见，8、9 分别为修复收尾发现的脱敏与日志一致性问题。

| 项 | 结论 | 依据与限制 |
| --- | --- | --- |
| 1 持久 Go 配置隐式启用 race、后代重建环境 | RESOLVED | 强制 GOENV=off、空 GOFLAGS、普通隔离 HOME；父子配置注入场景通过，只运行 go env，未执行 race。 |
| 2 混合成功候选掩盖失败、未知文件或授权 | RESOLVED | 检查全部 receipt、manifest、audit 与库存；失败、NOT_RUN、未知字节、grant/claim/ledger 均保守阻断。 |
| 3 未索引失败原件的传递依赖 | PARTIAL | 已加入失败及未知引用根；相对目录依赖仍被丢弃，见 Standards P1。 |
| 4 移动后状态保存失败破坏恢复依据 | RESOLVED | 同目录临时写、fsync、replace；注入写限额后保留有效 MOVED 状态，公开 restore 成功。 |
| 5 脱离 session 后代持管道导致超时挂起 | RESOLVED | deadline 覆盖进程及管道，定向取消与输出限额场景有界失败。 |
| 6 默认 live 完整资格及自报消费者结果 | PARTIAL | 原生构造器、Merge 与实际消费者回执已接入；普通 chat 仍接受零调用，见 Spec P1。 |
| 7 原生 live 的 tag/timeout 包发现 | RESOLVED | 保留原字符串、唯一主包匹配；合成生产者覆盖有效形式和错包/歧义拒绝。 |
| 8 隔离 HOME 后宿主 HOME 漏脱敏 | RESOLVED | 注册真实宿主 HOME，先失败后通过原件及 Host 审计结果一致，未放松审计规则。 |
| 9 保存 stdout/stderr/数字退出与实际进程一致 | PARTIAL | 正常成功和资格拒绝已一致；最终写失败仍不一致，见 Spec P2。 |

## 普通验收与回执核验

本次父级独立只读核对了 **104,574 个 SHA-256**：1,244 份执行源码、8,636 份选定输入、94,660 份原生工件及命令/状态/索引文件均匹配；源文件与输入集合相同，无遗漏或新增未登记项。普通 Go 日志的 12,890 个通过节点与回执测试列表一致；原生 **12,086 PASS / 214 NOT_RUN**、audit 零发现，通过案例集合与上一轮保持一致。

最终工具回执的 **41 组进程场景**、普通 Go、两种 vet、live-tag 仅编译，以及最终只读判定的命令、实际数字退出、日志文件哈希与来源保持一致。普通实际 exit0；最终只读判定实际 exit1，source / E2E / live 不适用检查通过，race 与正式压力缺失，**release.pass=false**。详见[本次只读核验 JSON](20261022-section1-rereview-receipts.json)及[原冻结验证索引](../verification.json)。这些既有成功结果不覆盖上述三种新复现场景。

Astra 本次另执行 14 项 Standards 定向场景、6 项消费者及 2 项配置/包发现普通进程场景，均通过；另完成三个失败复现及输出限额复现。所有复现只使用临时合成仓库，Go 相关命令隔离持久配置。复现脚本和摘要已复制到[复现清单](20261022-section1-rereview-repros/manifest.json)，保留 SHA-256；脚本依赖该受审版本 tools/tests 辅助代码及其原临时输出目录。

本次只新增 reviews/ 下审查材料；未修改实现、计划、冻结状态、验证索引或旧证据。未重跑完整普通 Go，未运行 race、实际压力、付费 live，未执行真实证据清理或第 2–8 节。
