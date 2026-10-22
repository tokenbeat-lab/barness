# 第 1 节复审：Spec

模型：GPT-6 Astra。固定对比：`3ddc774084fe502ca1a13ab8b41b208053e24378...fce5f387e375fa3f7901700f5f525fdb64c6804f`；实际普通受测：`3aafd2d41436d222449f5b5f0e8c0596ed1154e9`。结论：不能全部关闭；发现一项资格残留和一项新消费者收尾缺陷。

| 复查项 | 结论 | 依据 |
| --- | --- | --- |
| 1：持久 GOFLAGS / 重建环境后代 | RESOLVED | `run_process.py:19,34–47` 固定 GOENV/GOFLAGS、隔离 HOME；定向父子 go env 注入测试通过。 |
| 6：默认 native live / 自报 true | PARTIAL | `record_live.go:35–45` 补上原生校验与 Merge，`live_consumer.go:161–209` 拒绝旧自报格式；普通 chat 仍可零调用通过，见 P1。 |
| 7：live tag/timeout 包发现 | RESOLVED | `run.py:106–110` 保留真实 package、要求唯一匹配；四种原生形式接受、四种错包/参数拒绝的假生产者测试通过。 |
| 8：隔离后宿主 HOME 脱敏 | RESOLVED | `ai/e2e/main_test.go:33–34` 注册宿主账户 HOME；现有定向 RED/GREEN 与最终 Host audit 拒绝旧包、接受新包的记录相符。 |
| 9：消费者 stdout/stderr / 数字退出 | PARTIAL | 正常成功/资格拒绝测试通过；收尾写失败仍记录假 0，见 P2。 |

**[P1，残留] 普通 chat 仍能以零请求原件取得发布通过。** [record_live.go:44](/Users/cyber/RestoX/barness/ai/internal/testkit/release/record_live.go:44) 调用的 Merge 没有普通 chat 的调用/预算/固定模型检查；[live_manifest.go:73](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_manifest.go:73) 又仅在 Attempts>0 时要求请求原件。临时 google-gemini 六场景仅写 PASS assertions，`calls=attempts=0`、`budget={}`、虚构型号/代理 endpoint，真实 `-live-bundle` 退出0，`-record` 返回 `release.pass=true`。违反 [plan.md:46、162、166](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/plan.md:46) 的真实资格/不可自报 PASS 合同及 [工具合同:19](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/README.md:19)。建议补齐普通 chat 原生资格校验并以公共 CLI 回归。复现：[脚本](/tmp/barness-section1-spec-rereview/reproduce-chat.py)、[结果](/tmp/barness-section1-spec-rereview/zero-call-chat-summary.json)。

**[P2，新引入] 消费者终结失败留下错误退出与 stderr。** [live_consumer.go:132–156](/Users/cyber/RestoX/barness/ai/internal/testkit/release/live_consumer.go:132) 在最终回执成功写出前保存退出0。临时 `RLIMIT_FSIZE=2904` 允许 stdout 写完、令最终 JSON 写失败：实际退出1、stderr 有错误，保存的 exit-code=0、stderr 为空。违反 [plan.md:61、64](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/plan.md:61) 的真实退出/失败原件合同。建议用外层进程捕获真实输出与退出，并覆盖终结写失败；不能提前声明成功。复现：[脚本](/tmp/barness-section1-spec-rereview/reproduce.py)、[结果](/tmp/barness-section1-spec-rereview/io-failure-summary.json)。

本次实际通过6条 live 消费者、2条 GOENV/包发现定向普通进程测试；首次两条测试名拼写错误已纠正，日志均保留在 `/tmp/barness-section1-spec-rereview/`。所有新复现只用临时合成仓库；仅普通编译/只读消费者，无真实 race、压力、付费 live、真实清理或后续节执行。当前真实发布仍为 false，不将未授权阶段未跑列为缺陷。未修改实现、计划、状态或旧原件。
