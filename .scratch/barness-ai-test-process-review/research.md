# 三套需求的测试过程复盘

日期：2026-10-21（Asia/Shanghai）。对象是 GLM、MiniMax、地区 Provider 三套需求的**测试过程**，不是三家厂商 API 行为。测量数字只来自本次打开的仓库文件、`du`/`find`/`rg` 计数，以及文末列出的一手网页。文件里没有的耗时写成「未知」。未运行 `go test`。

## 1. 范围与方法

读过：

- 验证政策全文：[docs/agents/verification.md](../../docs/agents/verification.md)。Issue 约定读到「验收写在工单里、每功能一份当前状态」：[docs/agents/issue-tracker.md](../../docs/agents/issue-tracker.md) 第 1–12 行。
- 三份研究只取它们如何定义验证：[.scratch/barness-ai-glm/research.md](../barness-ai-glm/research.md) §7.7（约第 218–225 行）、[.scratch/barness-ai-minimax/research.md](../barness-ai-minimax/research.md) §7.7（约第 252–259 行）、[.scratch/barness-provider-sites/research.md](../barness-provider-sites/research.md) §7 与研究验证记录（约第 190–224 行）。未复述厂商 API。
- 工单：三目录分别为 13、22、26 个 `issues/*.md`（`find` 计数）。GLM 13 张标题与 Answer 都扫过验收句；MiniMax 抽了 01–08 的范围标题，以及 12、15、19、21 的验收/范围；provider-sites 抽了 01–07、06 的 Answer，以及标题含验收、恢复、差分、原生失败的 08、13、14。没有逐行读完全部工单。
- GLM：[live-evidence-plan.md](../barness-ai-glm/live-evidence-plan.md)、[evidence/ticket-02/failure-modes.md](../barness-ai-glm/evidence/ticket-02/failure-modes.md)、[evidence/ticket-04/README.md](../barness-ai-glm/evidence/ticket-04/README.md)、[evidence/ticket-13/README.md](../barness-ai-glm/evidence/ticket-13/README.md) 与 `plan.md` 的预算段、ticket-13 `tdd/` 的 43 个 `*.log` 文件名，以及其中 8 份的头尾（见 §3）。
- MiniMax：[delivery-status.md](../barness-ai-minimax/delivery-status.md)、ticket-19 `final-gate/ordinary/observation.json`、ticket-21 `final-gate/README.md` 与 `final-closure/README.md`。
- provider-sites：[spec.md](../barness-provider-sites/spec.md)「当前状态」与 Testing Decisions、[evidence/verification.md](../barness-provider-sites/evidence/verification.md)、[evidence/acceptance.md](../barness-provider-sites/evidence/acceptance.md)、[evidence/review.md](../barness-provider-sites/evidence/review.md)、[tools/README.md](../barness-provider-sites/tools/README.md)、[tools/FAILURES.md](../barness-provider-sites/tools/FAILURES.md)、[tools/RECOVERY_FAILURES.md](../barness-provider-sites/tools/RECOVERY_FAILURES.md)、[tools/RECOVERY5_FAILURES.md](../barness-provider-sites/tools/RECOVERY5_FAILURES.md)、[ai/live/REGIONAL.md](../../ai/live/REGIONAL.md)。
- 代表用例各读开头：`ai/e2e/glm_chat_text_test.go`、`ai/e2e/minimax_chat_text_test.go`、`ai/e2e/regional_protocol_guards_test.go`；live 驱动读了 `ai/live/glm_harness_test.go`、`ai/live/regional_harness_test.go` 的入口。
- 外部：pi 提交 `73f8de80bc87021375f8b525292ac07492253002`（`gh api` 确认存在，提交说明为 `fix(ai): recover from Copilot request-size overflow`，日期 2026-10-19T20:46:26Z）；[pkg.go.dev/testing](https://pkg.go.dev/testing)；[go.dev/doc/code](https://go.dev/doc/code)；[pkg.go.dev/cmd/go](https://pkg.go.dev/cmd/go) 的 Testing flags；[Google Testing Blog: Test Sizes](https://testing.googleblog.com/2010/12/test-sizes.html)；[Software Engineering at Google 第 11 章](https://abseil.io/resources/swe-book/html/ch11.html)。未打开 ISTQB / ISO 29119，因此不引用。

计数方法：

- `*_test.go`：在 `ai/e2e`、`ai/live`、`ai/internal/testkit` 下，文件名或路径含 `glm` / `minimax` / `regional`（`minimaxlive` 归 MiniMax）。`func Test` 用行首 `func Test` 计数。`zai` 只作为函数名子串另计，且全部落在已计入的 regional 函数里。`zhipu` 函数名命中 0。
- 行数是这些文件的换行计数，含没有 `func Test` 的夹具/helper。
- 体积用 `du -sk`（1024 字节块）。`du -sh` 的 69M / 498M / 1.0M / 15G / 12G 是同一测量的四舍五入，下文用 KiB。
- 日志名：先做子串 `red`/`green`/`race`，发现 `restored` 含 `red`、`trace` 含 `race`，会把证据文件数放大一个数量级。下文「记号」指文件名里独立记号 `(^|[^A-Za-z])red|green|race([^A-Za-z]|$)`。两者都报告，结论只用记号计数。

没读：每个工单的全文、ticket-13 以外的 GLM 日志正文、MiniMax 601 个 `*.log` 的正文、`.evidence` 除体积以外的内容（全树 `du -sh` 为 189G，正文只打开了 `provider-sites` 与 `barness-ai-glm` 相关索引）、pi 的 176 个测试里除下文点名文件以外的正文。

## 2. 三套需求共同的测试方案

三套都把 [verification.md](../../docs/agents/verification.md) 当作流程，而不是另写一套门禁。共同形状是：

1. **先失败、再实现，接缝是公共 Client / Host，不是私有 helper。** 政策第 5 行要求在公共边界先建立失败场景。GLM 研究 §7.7 第 1 步、MiniMax 研究 §7.7 第 1 步、provider-sites 规格 R20（[spec.md](../barness-provider-sites/spec.md) 第 180–184 行）写的是同一句。工单 Answer 反复出现「先 RED 再实现」。
2. **普通验证与付费 live、人工 race 分开。** 政策第 27–33 行：前置 → 双 vet → 带差分和压力的完整普通测试 → 证据消费 → `READY_FOR_RACE`；race 与新厂商调用都要人类另选。三套研究都写了「研究文档不是 live 通过」。provider-sites 研究第 207 行明确本次纯文档不跑这些验证。
3. **证据是指针，失败原件不改写成通过。** [issue-tracker.md](../../docs/agents/issue-tracker.md) 第 12 行；政策第 46–48 行要求每阶段记录 argv、数字退出码、起止时间、来源，且未执行 / 失败 / 输入过期不能互相替代。MiniMax [delivery-status.md](../barness-ai-minimax/delivery-status.md) 把历史 PARTIAL/FAIL 留在下文，当前状态单独一段。provider-sites 规格第 9–11 行声明「当前状态」是唯一入口。
4. **来源一变，旧报告不能冒充当前资格。** 政策第 15–17、31 行。这是后面 recovery / 十一组重验的合同来源，不是事后加的。
5. **墙钟不能把父子测试或并行包相加。** 政策第 46 行。provider-sites 验证记录第 26 行、验收记录第 19 行都按这条记账。

分层落到三套上的差别只是入口名字：GLM/MiniMax 用 `BARNESS_AI_LIVE` 与单组合 Key；地区用 [REGIONAL.md](../../ai/live/REGIONAL.md) 的 grant、`TestRegionalCandidate` / `TestRegionalFormal`，默认普通测试不访问厂商（该文件第 3 行）。

## 3. 分别的用例形状、流程、结果、耗时

### 3.1 用例形状（读到的三份代表）

| 文件 | 形状 | 网络 | 断言 |
| --- | --- | --- | --- |
| [ai/e2e/glm_chat_text_test.go](../../ai/e2e/glm_chat_text_test.go) 第 16–49 行 | `TestGLMChatText`：型号 × `stream/complete/simple` × hooks 的嵌套循环；`run.Case` 案例号 `GLM02-text-…` | 本地 controlled Provider；注释写明只有本地 Provider 被替换，Client 是真的 | `ev.Check` 对回调身份、请求字段；夹具 `text.json` 入队 |
| [ai/e2e/minimax_chat_text_test.go](../../ai/e2e/minimax_chat_text_test.go) 第 13–70 行 | 同样是型号 × 四个入口；案例号 `MM02-text-…` | 合成帧 `enqueue`，不是 live | 省略字段、`reasoning_split`、用量归一、回调顺序 |
| [ai/e2e/regional_protocol_guards_test.go](../../ai/e2e/regional_protocol_guards_test.go) 第 11–74 行 | `TestRegionalProviderPreflightAndFinalGuards`：Provider × API × 五种拒绝模式 | 本地 `regionalWorld` | `checkRejected`：零请求、零 Attempt、凭据读取次数 |

Live 不是这张表的延续。[ai/live/glm_harness_test.go](../../ai/live/glm_harness_test.go) 有 `//go:build live`；`TestGLMCandidate`/`TestGLMFormal` 在未选定路线时 `Skip`，没有 `BARNESS_AI_LIVE` 或 Key 时 `Skip("NOT_RUN:…")`，`Mode==replay` 则安装捕获回放（第 30–79 行）。[ai/live/regional_harness_test.go](../../ai/live/regional_harness_test.go) 同样 build tag；未选路线则 `NOT_RUN`，先 `consumeGrant` 再发请求（第 24–52 行）。场景清单在计划里，不在 Go 表驱动字面量里：GLM 候选计划每协议两型号各 9 个基础调用加对照，合计每协议 22 次逻辑调用（[live-evidence-plan.md](../barness-ai-glm/live-evidence-plan.md) 第 15–19 行）。地区每型号 9 项，ZAI 两型号 18 次、MiniMaxCN 一型号 9 次（[REGIONAL.md](../../ai/live/REGIONAL.md) 第 41 行）。

### 3.2 Go 测试文件（本次计数）

路径归类。`zai` 的 6 个 `func Test` 名字都同时含 `Regional`，已包含在 regional 的 46 里，不另加。`zhipu` 的 `func Test` 名字为 0；智谱身份出现在 GLM 用例的 `ProviderID("zhipu")` 断言里（`glm_chat_text_test.go` 第 48 行），没有单独的 zhipu 测试文件。

| 树 | 流 | 文件 | 行 | 文件内全部 `func Test` | 函数名含该关键字 |
| --- | --- | ---: | ---: | ---: | ---: |
| `ai/e2e` | glm | 67 | 7002 | 103 | 103 |
| `ai/e2e` | minimax | 72 | 6205 | 102 | 102 |
| `ai/e2e` | regional | 45 | 5008 | 46 | 46（其中 6 个名字还含 `ZAI`） |
| `ai/live` | glm | 24 | 3378 | 16 | 16 |
| `ai/live` | minimax | 16 | 2141 | 13 | 13 |
| `ai/live` | regional | 6 | 641 | 2 | 2 |
| `ai/internal/testkit` | glm（`release/glm_contract_test.go`） | 1 | 38 | 1 | 1 |
| `ai/internal/testkit` | minimax（目录 `minimaxlive/` 加 `release/minimax_load_fixture_test.go`） | 35 | 5041 | 28 | 2（仅 `TestMiniMaxRecovery7HarnessDeadline*`；其余 26 个名字是 `Recovery*` / `Renewal*`） |

三树合计 266 个文件、29454 行。e2e 的 251 个 `func Test` 会在循环里展开成多得多的原生案例；下文的 12495、13244 是**整包**原生案例数，不是这 251 个函数的个数。

### 3.3 目录体积与日志名

| 位置 | 文件数 | `du -sk` |
| --- | ---: | ---: |
| `.scratch/barness-ai-glm` | 195 | 71056 KiB |
| `.scratch/barness-ai-minimax` | 2435 | 510452 KiB |
| `.scratch/barness-provider-sites` | 108 | 1032 KiB |
| `.evidence/provider-sites` | 1052262 | 15964372 KiB |
| `.evidence/barness-ai-glm`（工单把原始包指到这里；不是本题点名的目录，但是 GLM 证据主体） | 687645 | 12899200 KiB |

稍后完成的 `du -sh .evidence` 报告整树 **189G**。`du -d 1 -h .evidence | sort -h | tail -40` 只留下最大的 40 项，不是 292 个一级目录的完整清单，也不能相加成这三套需求的合计。其中与本次直接相关的四舍五入体积：`.evidence/barness-ai` 61G（更早的 barness-ai 证据，不全是这三套）、`.evidence/provider-sites` 15G（与上表 15964372 KiB 是同一目录）、`.evidence/barness-ai-glm` 12G、`.evidence/ticket14` 7.5G、`.evidence/minimax-retained` 5.6G、`.evidence/minimax-current-visual-gate-1` 3.3G，以及约 3.1G 的 `minimax-current-eleven-gate-2`、`minimax-global-final-keyless`、`minimax-release-repair-incomplete-gate-1`、`minimax-routed-final-compiled-1`。同一轮 `find .evidence/provider-sites -type f` 得到 1052252，比上表 1052262 少 10；末位会随树变化。

记号计数（文件名，不是正文）：

| 位置 | `*.log` | 记号 red 的 log | 记号 green 的 log | 记号 race 的 log |
| --- | ---: | ---: | ---: | ---: |
| `.scratch/barness-ai-glm` | 43 | 23 | 16 | 5 |
| `.scratch/barness-ai-minimax` | 601 | 134 | 120 | 21 |
| `.scratch/barness-provider-sites` | 0 | 0 | 0 | 0 |
| `.evidence/provider-sites` | 1707 | 67 | 77 | 0 |
| `.evidence/barness-ai-glm` | 129 | 13 | 15 | 8 |

GLM scratch 的 43 个 log 全部在 `evidence/ticket-13/tdd/`：23 个 red、16 个 green、5 个名字含独立 `race`（其中 `prereview-view-bounds-race-red.log` 同时计入 red 与 race，两个 `*race-green.log` 同时计入 green 与 race）。子串计数会把 `.evidence/provider-sites` 的 red 记成 9157、race 记成 2681，因为大量 `restored-*` 与 `wire_trace.json`；那个数不能当作红绿日志爆炸。

provider-sites Python：`tools/test_*.py` 18 个、2449 行。其中 7 个 `test_*fixture.py` 只有 `seed`/`builder`/`approve`，`def test_` 为 0。其余 11 个文件共 107 个 `def test_`。[test_final_round.py](../barness-provider-sites/tools/test_final_round.py) 第 2 行写明 “Socket-free public-process checks; all approvals/keys are synthetic.” [FAILURES.md](../barness-provider-sites/tools/FAILURES.md) 第 1–3 行：测的是固定轮次操作入口，不访问真实服务或凭据。这是证据管道的自测，不是产品协议用例。

### 3.4 流程与读到的结果

**GLM。** 工单 01–13 全部 `resolved`。01–10 是离线公共 Client、隔离和冻结差分；11 才是付费候选。工单 11 Answer：六身份各 11/11，三协议各 22 Logical Call / 22 HTTP，合计 66/66/66，零重跑；进程墙钟 Chat 131.53s、Responses 111.79s、Messages 108.99s；随后无 Key 回放各 22 attempts、0 外部 HTTP；该票记录的 normal 11.36s、race 190.50s（[issues/11-glm-candidate-live-evidence.md](../barness-ai-glm/issues/11-glm-candidate-live-evidence.md) 第 31–37 行）。工单 12：正式目录下再跑 66 次 HTTP、零重跑。工单 13 Answer：公共 CLI 在 `3d3cd0f` 执行 normal/race，十门禁 PASS；并写明本票实际 HTTP 282（66 GLM + 180 原轮 + 36 更正），无自动重跑（[issues/13-glm-release-host-docs-rollback.md](../barness-ai-glm/issues/13-glm-release-host-docs-rollback.md) 第 31 行）。ticket-13 索引写最终无 Key 校验 exit 0、94.934s（[evidence/ticket-13/handoff-root.md](../barness-ai-glm/evidence/ticket-13/handoff-root.md) 第 16 行）。完整发布门禁的墙钟不在本次读到的 GLM 段落里，记为未知。

ticket-13 `tdd/` 抽样（这些是定向 `go test` 日志，不是 66 次 live）：

| 日志 | 读到的结果 | 包耗时 |
| --- | --- | --- |
| `prereview-error-depth-red.log` | `TestGLMDeepErrorEnvelope` 六子测试失败：嵌套 error 被当成合法信封，约 50 015 字节体分配约 1.00–1.01×10⁹ 字节 | 6.943s |
| `prereview-error-depth-green.log` | `ok … ai/e2e` | 3.473s |
| `outer75-red.log` | `TestGLMReleaseCommandEnvelope`：预登记的 race 命令仍是 `-timeout=45m`，不是 75m | 4.202s |
| `outer75-green.log` | `ok … ai/e2e` | 2.488s |
| `formal-mutations-race-current-5m-timeout.log` | `panic: test timed out after 5m0s`，停在 `TestGLMFormalReports/chat`，栈在 `regexp` 与 `supportmatrix.Merge` | 300.653s |
| `prepaid-live-harness-red.log` | `TestDecisionsCandidateProcess` 大批子测试失败，断言是 “candidate leaves formal catalog intact” | 95.838s |
| `cardinality-red.log` | `TestElevenIdentityReleaseCLI`：期望 11 个自身身份，得到 14 | 1.557s |
| `trace-red.log` | `TestGLMReleaseTraceContract`：GLM-R01 起一系列需求 ID 缺失 | 子测试 0.21s 起，全文未逐行加总 |
| `final-durable-verification.log` | 档案校验 PASS | 日志内无秒数 |

多数 green 日志只有 52 字节、一行 `ok`。它们证明的是一次定向包通过，不是全量门禁。

**MiniMax。** 22 张工单全部 `resolved`（含两个 14 号文件）。01–08 是协议、隔离、live harness、Host、差分和验收；12–21 是捕获修正、来源重验和恢复链。当前状态写在 [delivery-status.md](../barness-ai-minimax/delivery-status.md) 第 1–9 行：Delivery COMPLETE（2026-10-19）；来源 `78843` 四条门禁命令退出码 0/0/0/0；普通/race 主包各 **12495 PASS、0 FAIL、0 NOT_RUN、audit 0**；严格来源 11/11；续期累计 **873 calls / 870 HTTP Attempts**。同一文件保留的历史：来源 `01341` 的 race 是 12494 PASS / 1 FAIL，唯一失败 `VQ12-failure-decisions-plain-deadline` 的 `samples/v08/request.json` 零字节（第 23 行）；更早 Messages `glm-5.3/cancel` 为 120001ms FAIL，票 18 单场景 861ms，票 19 完整 22 场景里该场景 1146ms PASS（第 19、31 行）。

ticket-19 `ordinary/observation.json` 有 12495 PASS 和归档 114962711 字节，**没有墙钟字段**。ticket-21 `final-gate/README.md` 第 7–13 行给出普通胶囊 142534 个常规文件 / 1350851478 字节、race 胶囊 142587 文件 / 1346333368 字节，同样没有墙钟。`outer/observation.json` 的 argv 含 `-timeout=20m` 与 `-timeout=75m`，这是上限不是测量。MiniMax 全量普通/race 的墙钟：**未知**。

**provider-sites。** 26 张工单里本次 `rg "^Status:"` 看到 9 张 `resolved`（01–06、13、14、15），其余 `claimed`。规格当前状态（[spec.md](../barness-provider-sites/spec.md) 第 13–24 行，Updated 2026-10-21）：实现 01–06 已交付；恢复 6 在 ZAI Responses 18/18 之后，M3 Chat 的 full-stream 仍为 REQUEST deadline FAIL，后 17 个进程未启动；恢复 7 的完整 native 普通门禁触及 20 分钟上限，**exit 1、墙钟 1203.444s、266 PASS / 12 SKIP，不等于包通过**；`release.pass: false`，正式 race `NOT_RUN`。旧开发 race 的 2589.635s PASS 只证明来源 `470682d3`。

读到的已关闭测量：

- 普通全量：[evidence/verification.md](../barness-provider-sites/evidence/verification.md) 第 18–28 行。`go test -json ./... -count=1 -timeout=20m` exit 0，墙钟 **661.883s**；主 E2E **13244** 个原生场景全部 PASS；889 份 pi 差分 PASS、pending 0；204 个名字含 pressure 的场景通过。双 vet 0.729s 与 0.558s。
- 同文件第 48–50 行保留失败：基线 3376 场景中 3367 PASS / 9 NOT_RUN，且当时没开差分/压力。第一次完整普通（final-v3）主包 **13238 PASS / 6 FAIL**，外层 collector exit 1。
- 人工 race：[evidence/acceptance.md](../barness-provider-sites/evidence/acceptance.md) 第 9–17 行。`go test -json -race ./... -count=1 -timeout=75m` exit 0，UTC 2026-10-20T06:40:40Z 到 07:23:50Z，墙钟 **2589.635s**；主包仍是 13244/13244 PASS。该轮没有 `READY_FOR_RACE` checkpoint，文中写明不冒充正式发布 race。
- 六条 candidate（同文件第 27–36 行）：**81 Calls / 81 HTTP、零重试，80 PASS / 1 FAIL**。唯一失败是 ZAI Messages `glm-5.2/native-history`（17/1）。预留估算 US$2.282610，Result 用量估价之和 US$0.020954；文中写明两者都不是厂商账单。

## 4. 哪些实践有效

以下每条都是测量事实：它抓住的故障写在点名的文件里。建议见 §7，不混在这里。

1. **公共 Client 加合成/本地 Provider，能抓住「成功但内容为空」。** [ticket-04/README.md](../barness-ai-glm/evidence/ticket-04/README.md) 第 5 行：第一次公共 Client 运行在两个型号上复现了 SDK 静默丢失，没有错误、调用完成，对象型 reasoning/message 变成空字符串。同文件最终回归 1978 PASS，冻结 Responses 差分 335 PASS、pending 0。这不是 live 才发现的。
2. **有界 live 能抓住离线夹具覆盖不到的上游形状。** 三份独立失败都写明离线重放没有证明解析器丢字段，因此没有改 parser 或删断言：
   - ZAI Messages `glm-5.2/native-history`：HTTP 200、工具与用量通过，九次响应都没有 thinking/signature（[acceptance.md](../barness-provider-sites/evidence/acceptance.md) 第 40–44 行）。
   - 智谱 Responses `glm-5.2/native-history`：22 场景 21 PASS / 1 FAIL，完成帧没有 message/output_text，Client 保存 Text 0（[issues/13-zhipu-responses-glm52-final-text.md](../barness-provider-sites/issues/13-zhipu-responses-glm52-final-text.md) 第 12–14 行）。后来同断言、不改 Go 的单次重验 22/22 PASS（第 26 行）。
   - 国际 MiniMax-M3 Chat thinking-on：13 场景 12 PASS / 1 FAIL，流里没有 reasoning 字段；旧成功响应用同一 Client 重放为 GREEN（[issues/14-minimax-global-chat-native-thinking.md](../barness-provider-sites/issues/14-minimax-global-chat-native-thinking.md) 第 12 行）。
3. **单场景超时被留下来，而没有被整包绿色盖住。** Messages `glm-5.3/cancel` 在可观察内容前到达 120001ms（[delivery-status.md](../barness-ai-minimax/delivery-status.md) 第 45 行）。票 16 把 GLM Host 响应头从 90 秒对齐到 120 秒调用上限之后，票 18 单场景 861ms、票 19 完整报告里 1146ms，且 partial usage 不记免费（同文件第 19、31 行）。失败原件仍在。
4. **race 抓到普通全绿没抓到的证据损坏。** 来源 `01341`：普通 12495 PASS / audit 0，race 12494 PASS / 1 FAIL，唯一审计是零字节 `request.json`（delivery-status 第 23 行）。票 20/21 修的是「发送前取消仍留下空请求」，不是再改协议。
5. **失败先行的资源探针抓到了解析放大。** `prereview-error-depth-red.log` 在约 5 万字节的嵌套错误体上分配约 10⁹ 字节。审查记录里的独立探针是 100005 字节、9999 层、分配 2103884384 字节、1.96 秒（[evidence/ticket-13/prereview/review.md](../barness-ai-glm/evidence/ticket-13/prereview/review.md) 第 9 行）。随后的 green 日志是 3.473s 的定向包，不是把探针从门禁里删掉。
6. **冻结差分把「上游没有这条路由」和「这条路由不一致」分开。** provider-sites 工单 06：先在公共子进程得到三个身份缺失的 red，再只追加冻结包已经导出的 `ZAI_MODELS` 与 `MINIMAX_CN_MODELS`；18 次双侧比较 + 3 次目录比较，另外 24 个扩展案例明确 `piParity: NOT_RUN`（[issues/06-frozen-differential.md](../barness-provider-sites/issues/06-frozen-differential.md) 第 28–32 行）。定向新路线包耗时 4.495s（第 38 行）。GLM 工单 10 同样把 GLM 记成扩展，而不是 pi parity。
7. **预算与 FAIL 不转移，避免用未用额度把失败刷成通过。** 恢复 6 在第一条失败后停止后 17 个进程（spec 第 16–17 行）。MiniMax 票 15 写明未用 6/30 不转移，120001ms 失败不拼进 22 场景报告（[issues/15-source-scoped-revalidation.md](../barness-ai-minimax/issues/15-source-scoped-revalidation.md) 第 54–63 行）。

## 5. 哪些实践低效或失效

每条都有文件。没有文件的耗时不估计。

1. **来源一变就重跑整包普通/race，检出的新故障很少，重复成本高。** 政策要求相关输入变化后重验受影响范围（verification.md 第 15 行），三套把它做成了多次 `go test ./...` 与 `go test -race ./...`。读到的整包结果：provider-sites 普通 661.883s（13244 场景），race 2589.635s（同样 13244 场景，且只覆盖来源 `470682d3`）。MiniMax 案例数从工单 12 评论里的 12385，到票 15 的 12451（12445 PASS / 6 FAIL），再到 19/21 的 12495。票 19 的 11/11 live 已经通过，完整 race 仍因一个视觉证据文件失败；票 21 只新跑了 36 次 HTTP，却又跑了一遍 12495 的普通和 race。MiniMax 这两次全量的墙钟未知，但胶囊体积已知：普通约 1.35×10⁹ 字节、14 万文件。final-v3 已经花过一整轮才发现 6 个旧断言失败（13238 PASS / 6 FAIL），修完又跑 final-v4 的 661.883s。
2. **恢复轮次在测控制器，不在增加协议故障的检出。** provider-sites 从固定恢复 3 写到恢复 7：每轮一个新的 `recoveryN_round.py`，外加 `test_recoveryN_*.py`。18 个 Python 文件、107 个测试方法、2449 行，断言的是 claim、Key 前拒绝、FIFO、预算不退款（三份 FAILURES 文档的开头）。Go 侧 `minimaxlive` 35 个文件、5041 行、28 个 `func Test`，名字是 `TestRecovery4AuthorityProcess` 这类进程接缝。规格当前状态：恢复 6 付费停在 M3 Chat deadline；恢复 7 还没发行授权，keyless native 包就在 1203.444s 顶满 20 分钟，266 PASS / 12 SKIP 仍是 exit 1。feature 保持 claimed，`release.pass` 仍是 false。
3. **red/green 日志按微断言裂变，多数不增加故障信息。** ticket-13 `tdd/` 43 个日志里，`outer75-red` 用 4.202s 发现预登记命令字符串仍是 45 分钟，`outer75-green` 是 52 字节的 `ok`。`cardinality-red` 与 `trace-red` 失败的是身份个数和需求 ID 映射，不是协议。MiniMax scratch 有 601 个 `*.log`、134 个 red 日志、120 个 green 日志。`.evidence/provider-sites` 另有 1707 个 `*.log`。这些日志没有被收成「一份当前状态」；政策第 48 行禁止为收尾再做逐轮复制器，现有树已经超出这个形状。
4. **证据管道的自测会失败在别的产品的目录不变式上，并占用预付费检查。** `prepaid-live-harness-red.log` 包耗时 95.838s，失败点是 `TestDecisionsCandidateProcess` 的 “formal catalog changed”，不是 GLM wire。`formal-mutations-race-current-5m-timeout.log` 在 5m0s 超时，栈停在支持矩阵合并的正则扫描；[ticket-13/plan.md](../barness-ai-glm/evidence/ticket-13/plan.md) 第 19 行把这记成既有线性脱敏审计的 race 成本，没有 provider 调用。这次 300.653s 没有发现新的协议故障，只发现门禁自己的耗时。
5. **父子时间：没有找到一份把父子耗时相加后当作墙钟发表的错误数字。** 相邻的、已经发生的测量事故是另一件。[verification.md](../barness-provider-sites/evidence/verification.md) 第 50 行：final-v3 后处理把两份示例 manifest 当成 native Run，没先落下子进程退出码。同文件第 26 行与 acceptance 第 19 行写明后来改用命令单调时钟，包耗时不相加。`outer75-red.log` 测的是父命令信封（45m 对 75m）与子 argv 不一致，测试本身 4.202s 结束。所以「误加」是被写成禁令的风险；本次能指认的实测故障是收集器把子进程结果归错类，不是一个加总出来的秒数。
6. **整路线重跑把已经绿的场景又买了一遍。** 这是合同选择，不是记账错误。GLM 候选 66 次 HTTP 之后，正式目录再 66 次（工单 12），发布票又记 282 次 HTTP（工单 13）。MiniMax 续期累计 873/870，其中 GLM 三路在票 12、15、19 多次完整 22 场景。单场景失败（cancel、native-history 缺正文）确实值得保留；同轮里其余 21 个已绿场景被一起重买，是因为消费者拒绝拼接报告（票 15 Acceptance）。检出率的增量集中在那一个失败场景，成本按整路线计。
7. **目录名里的 green/final 不能当通过。** provider-sites 验证记录第 52 行写明：目录名带 green/final 但退出码非零的轮次要按数字退出码保留。这是管道曾经用名字暗示结果、后来被自己的规则否定的证据。

## 6. 与 pi-ai 和一手测试理论的差距

### 6.1 pi `packages/ai`（提交 `73f8de80…`）

树列表：`packages/ai/test/` 下 **176** 个 `*.test.ts`。`package.json` 的 `test` 脚本是 `vitest --run`，devDependency `vitest` 为 `4.1.11`。[vitest.config.ts](https://github.com/earendil-works/pi/blob/73f8de80bc87021375f8b525292ac07492253002/packages/ai/vitest.config.ts) 把环境设为 `node`，`testTimeout: 30000`，注释写 “30 seconds for API calls”。

打开的两份测试：

- [zai-coding-plan-models.test.ts](https://github.com/earendil-works/pi/blob/73f8de80bc87021375f8b525292ac07492253002/packages/ai/test/zai-coding-plan-models.test.ts) 有 3 个 `it`，只调用 `getBuiltinModel`，断言 `baseUrl`、cost、compat。这段没有网络。
- [anthropic-eager-tool-input-e2e.test.ts](https://github.com/earendil-works/pi/blob/73f8de80bc87021375f8b525292ac07492253002/packages/ai/test/anthropic-eager-tool-input-e2e.test.ts) 用 `it.skipIf(!testCase.apiKey)`，通过时 `retry: 2`，函数 `expectToolEnabledRequestAccepted` 调用 `complete(...)`。没有 Key 就不打网络；有 Key 就是真实调用，并且失败会重试两次。

路径里没有 `minimax`、`glm`、`zhipu` 测试文件。`minimax.ts` / `zai.ts` 在 `src/providers/`。GitHub code search 对 `packages/ai/test` 返回 minimax 12 条、zai 21 条；除 `zai-coding-plan-models.test.ts` 外，那些命中文件的正文本次没打开，不能说它们断言了什么。

和 barness 的差距是结构，不是「谁更认真」：pi 默认是进程内目录/转换测试，网络测试按 Key 跳过，超时 30 秒，没有 barness 这种逐身份 grant、脱敏审计和 `release.pass`。barness 的普通 E2E 已经比 pi 的目录 `it` 更接近「真 Client + 假服务」。多出来的成本在全量 `./...`、race 胶囊和恢复控制器，不在「还没有公共 Client 测试」。

### 6.2 Go 工具链自己怎么说

[go.dev/doc/code](https://go.dev/doc/code) 的 Testing 一节：测试是 `_test.go` 里的 `TestXxx`，失败是调用 `t.Error` / `t.Fail`；示例是纯函数的表驱动循环。它写明细节去看 `go help test` 和 testing 包，**这一页没有 `-count`、`-race` 或包级并行**。

[pkg.go.dev/testing](https://pkg.go.dev/testing)：同包测试可以碰未导出标识；外部 `package foo_test` 只能用导出标识，称为 black box。`T.Parallel` 只和其它并行测试一起跑，并等到非并行测试结束；`-test.count` 或 `-test.cpu` 造成的同名多次运行不会互相并行。`T.Run` 在单独的 goroutine 里跑子测试。标准库把测试范围定义成「这个函数失败与否」，没有 small/medium/large，也没有「必须 `go test ./...`」。

[pkg.go.dev/cmd/go](https://pkg.go.dev/cmd/go) Testing flags：

- `-count n`：每个测试、基准和 fuzz seed 跑 n 次，默认 1。`-count=1` 是显式关掉成功结果缓存的惯用写法。可缓存的 flag 列表里**没有** `-count`。
- `-race`：打开数据竞争检测；覆盖模式在 `-race` 下变为 `atomic`。
- `-parallel` 默认 `GOMAXPROCS`，只限制**一个测试二进制里**调用了 `t.Parallel` 的测试。不同包的测试二进制由 `-p` 决定并行，默认也是 `GOMAXPROCS`。

barness 的全量命令已经使用 `-count=1`（provider-sites 验证记录第 20 行），这和 cmd/go 的缓存规则一致。包并行意味着 `go test ./...` 的墙钟不是各包 `ok` 行秒数之和；这和 verification.md 第 46 行是同一件事，标准库把原因写在 `-p` 上。

### 6.3 Google 的尺寸模型

[2010-12-13 的 Test Sizes](https://testing.googleblog.com/2010/12/test-sizes.html) 用资源而不是名字分类。表内：Small 无网络、无多线程、无 sleep，时限 60 秒；Medium 只允许 localhost，时限 300 秒；Large 可以访问外部系统，时限 900 秒以上。博文没有回答时限是单测试还是整组（评论区有人问了，正文没答）。同文要求测试可按任意顺序跑，以便并行。

[SWE book 第 11 章](https://abseil.io/resources/swe-book/html/ch11.html) 把 **size**（资源）和 **scope**（验证了哪些路径）分开。Small：单进程，禁止 sleep、I/O 和网络（脚注允许 hermetic 内存文件系统）。Medium：可以多进程、线程和到 localhost 的阻塞调用，不能访问其它机器。Large：去掉 localhost 限制，留给发布阶段，不要挡在开发循环里。文中给出的粗比例是按案例数约 80% 窄范围、15% 组件交互、5% 端到端，并写明每个团队会不同。它把 large 测试的主要价值写成配置与遗留系统的健全性检查，而不是抓日常逻辑缺陷的主手段。同章不鼓励测试里的条件分支和循环，因为测试本身没有测试。

对照本次读到的 barness：

- 代表 E2E 是真实 Client、进程内假 Provider、会写证据文件、流式路径有并发。按该书是 **medium size、偏宽的 scope**（公共入口 × 型号 × 协议），不是 small。Go 官方示例本身就是表驱动循环，所以「测试里有 for」符合 [go.dev/doc/code](https://go.dev/doc/code)，不符合 SWE 书那条推论。这里两份一手来源不一致；barness 已选择 Go 的表驱动。
- `//go:build live` 加显式开关，符合「large 测试离开日常循环」。恢复 7 把 keyless 的 `ai/live` 整包放进 20 分钟普通门禁并超时，等于把 medium/large 边界上的夹具又拉回每轮必跑。
- 全量 12495 / 13244 案例里，真正打到厂商的是几十次 HTTP（候选 66、地区 81）。按案例数，端到端 far below 5% 这个句子不能从本次文件算出来：原生案例数混有夹具展开，live 案例数是另一套计数，没有同一分母。能说的是：发布门禁的时间主要花在离线整包和 race 插桩，不是花在厂商调用上。provider-sites 普通 661.883s 对 live 六路线的进程（验收表里每条路线单独计调用，没有给出六条进程墙钟之和，故六条合计墙钟未知）是不同的账单。

## 7. 优化方案

方案只调整 [verification.md](../../docs/agents/verification.md) 已经有的三层：**开发期受影响普通 E2E、人工选择的 race、另授权的付费 live**。不新增第四套流水线，不把 `release.pass` 的十项门禁改短。

每条都写：改什么、为什么能降耗时或提高检出、不改什么、怎样证明门禁没放宽。

### P0 开发期停止用全量 `./...` 代替受影响范围

改什么：功能开发、恢复控制器和文档收尾的普通验证，改回政策第 15 行已经写的规则——只跑受影响公共 E2E，命令保持 `go test ./ai/e2e -run '<已有前缀>' -count=1`。完整 `go test ./...`（差分、压力打开）只在进入 `READY_FOR_RACE` 之前跑一次。race 仍只在人类点名时跑，且只对那一份冻结来源有效。

为什么：读到的重复全量是 661.883s 和 2589.635s 这一级，而票 06 的新路线差分包是 4.495s，工单 11 记录的 GLM 定向 normal 是 11.36s。全量重复没有在这些短跑已经绿的协议用例上多抓到协议缺陷；票 19 的 race 新抓到的是一个空请求文件，那一类应留在唯一一次发布 race，而不是每次恢复都跑。

不改什么：发布定义仍是政策第 33 行的完整 `go test ./...`、`go test -race ./...`、双 vet、差分、压力、真实来源资格和逐包审计。历史 FAIL 不改写成跳过。

怎样证明没放宽：一次 keyless 对照，同一来源上 (a) 定向 `-run` 与 (b) 全量的失败集合。若定向集合是全量失败集合的子集且多出来的失败属于未改动的包，定向即可作为开发回执；全量仍是 READY 的唯一 PASS。把这次对照的 argv、退出码和失败案例 ID 记进该功能的唯一状态入口。

### P0 恢复轮次不再复制控制器和它的测试

改什么：停在恢复 7，不要恢复 8。下一轮若还要付费，复用已有一个操作入口，换一份有版本的计划/授权文件。新的拒绝条件加进现有 FAILURES 清单的一条，而不是新的 `test_recovery8_round.py` 加一组 `TestRecovery8Authority*`。`minimaxlive` 里按轮次复制的进程测试不再按 N 增加文件。

为什么：107 个 Python 方法和 5041 行 Go 夹具测的是「Key 前拒绝、预算不退」。恢复 7 的 1203.444s 超时发生在这些 keyless 自测上，后面的付费与正式 race 仍是 NOT_RUN。自测的边际检出是控制器回归，不是厂商故障；控制器回归用一份共享夹具就能守住三份 FAILURES 里已经列出的失败方式。

不改什么：Key 前拒绝、失败不退款、不自动重跑、旧 seal 只读。这些是 [FAILURES.md](../barness-provider-sites/tools/FAILURES.md)、[RECOVERY_FAILURES.md](../barness-provider-sites/tools/RECOVERY_FAILURES.md)、[RECOVERY5_FAILURES.md](../barness-provider-sites/tools/RECOVERY5_FAILURES.md) 里已经证明有效的拒绝，必须仍在读 Key 之前失败。

怎样证明没放宽：对合并后的入口重放三份 FAILURES 里各一条代表性拒绝（未批准、FIFO、超额、旧轮次重开），退出码仍为非零且证据目录没有新的厂商 socket。不要求再生成 43 个 red/green 日志；一条失败记录进现有状态索引即可。

### P1 付费重验按失败场景发新 grant，不拼接、也不默默重买整路线

改什么：运行时源码未变、只有一个场景 FAIL 时，新的人类授权可以只覆盖该场景的新 grant。消费者仍拒绝「缺场景的单份报告」。通过条件改成：旧的已封存同轮原件保持字节不变，加上这一份新场景原件，哈希都对得上。源码或目录哈希变了，仍按政策重验受影响组合，而不是只重跑一个场景。

为什么：cancel 的 120001ms 与 native-history 缺正文，都是一个场景的失败。票 13 的 Responses 重验证明「同一断言再跑一次」可以变 PASS，且不需要改 parser。同路线其余已绿场景的重复 HTTP 没有新的故障记录。GLM 候选本身 66 次、约 109–132 秒量级，是便宜的；贵的是后续把 22 场景和全量门禁绑死之后的 873/870 与多次 12k 案例。

不改什么：FAIL 原件、零重试上限、不能把 NOT_RUN 写成 PASS、不能用另一型号补位。工单 08/13/14 已经证明这些限制挡住了「删断言」。

怎样证明没放宽：keyless 夹具两条都要仍失败——缺一个必交场景的报告被拒；新场景原件与封存兄弟的来源哈希不一致被拒。只有两条都拒，才允许少发 HTTP。

### P1 日志与胶囊收成一份索引

改什么：每个功能只留政策第 48 行要求的一份当前状态。新的 red 只在失败断言和退出码与上一份不同时追加。52 字节的 `ok` 行不再各存一份。全量门禁的原生 manifest 留在 `.evidence` 的一个目录，不再在 scratch 与 `.evidence` 各放一套可运行副本。ticket-21 那种 14 万文件的胶囊保留这一次发布 PASS 的原件，下一次发布用新目录，不把中间微迭代打进胶囊。

为什么：`.evidence/provider-sites` 已是 1052262 个文件、15964372 KiB；`.evidence/barness-ai-glm` 687645 个文件、12899200 KiB。ticket-13 的 23 个 red 日志里，抽样显示多份是命令字符串、身份计数和追溯 ID，信息量小于一次 `go test` 的失败行。体积本身会让下一次「完整读回」变慢；票 21 的留存读者已经要为 2GiB 解码上限写专门限制（final-gate README 第 17 行）。

不改什么：已封存的失败 ZIP、零字节 request、120001ms 的 cancel 原件、恢复 6 的停止索引。有效实践依赖这些原件还在。

怎样证明没放宽：索引里的每个 PASS/FAIL 仍指向原 argv、数字退出码和原件哈希。抽查 ticket-04 的空内容失败与 ticket-19 的零字节 request，哈希与现在的原件一致。少了日志文件不等于少了失败。

### P2 按现有三层标尺寸，不把 live 包的 keyless 自测算进每次普通全量

改什么：在 verification.md 的三层上加尺寸注记，不改阶段顺序。

- 普通开发：公共 Client + 本地/loopback Provider（现在的 `ai/e2e` 与 REGIONAL controlled）。这是 medium。继续用受控服务复现时序，而不是 `testing` 包的 sleep。
- 人工 race：仍是 `go test -race`，只在发布冻结后跑。race 插桩下的整包脱敏正则扫描不要再当作预付费门槛；ticket-13 的 5 分钟超时已经说明它测的是门禁自己。
- 付费 live：保持 build tag、单路线进程、grant。`go test -tags live ./ai/live` 的 keyless 自测若会接近 20 分钟，就从「每次恢复的普通门禁」挪到「操作入口变更时的定向 `-run`」。1203.444s 那次 exit 1 应视为自测预算失败，而不是厂商资格失败。

为什么：SWE 书建议把 large 测试留在发布，pi 用 `skipIf(!apiKey)` 做到了默认不打网络。barness 的 build tag 已经做到默认不打厂商；失效点是 live **包里的进程自测**被当成普通全量。分开之后，协议回归仍由 e2e 的 251 个函数承担，live 自测只在控制器变更时跑。

不改什么：`TestGLMDeepErrorEnvelope` 这类资源上限、`TestRegionalProviderPreflightAndFinalGuards` 的零 Attempt、冻结差分的 `NOT_RUN` 与真实双侧比较、live 报告的脱敏审计。这些是 §4 里抓到过故障的层。

怎样证明没放宽：定向 e2e 仍包含上述三个名字并 exit 0 才允许说普通开发通过；live 自测不跑时，状态写 `live-harness: NOT_RUN`，不能写成 PASS。发布前的那一次全量普通+race 仍是 `release.pass` 的必要条件。

### 不做什么

不把 barness 收成 pi 的 176 个 vitest。pi 的目录断言抓不住 ticket-04 那种「Client 完成但正文为空」，也抓不住 ZAI/智谱/MiniMax 的真实缺字段。不引入 ISTQB 过程，本次没有打开的标准页。不把 80/15/5 写成 barness 的配额：第 11 章写明那是粗比例，而且本次没有同一分母。

## 8. 未核实项

- MiniMax 来源 `01341` 与 `78843` 的全量普通/race **墙钟**。打开的 observation 与 final-gate README 只有案例数、退出码和字节数。
- 六条地区 candidate 各自的进程墙钟，以及 81 次 HTTP 的合计墙钟。验收表有调用数和退出码，没有秒数。
- GLM 工单 13 十项门禁的墙钟。读到的是 HTTP 次数、94.934s 的无 Key 校验，以及若干定向日志的包耗时。
- `.evidence` 全树只有 `du -sh` 的 189G，以及最大 40 项的四舍五入体积，没有整树 `du -sk`，也没有全部一级目录清单。`.evidence/barness-ai` 的 61G 不能算进这三套需求。
- 子串计数里的 9157 / 2681 不是日志爆炸的规模；若有人引用那个数，需要先去掉 `restored` 与 `trace`。
- 记号 red 与记号 green 的交集只对 ticket-13 的文件名逐个看过。MiniMax 601 个 log 的交集未逐个核对；早先子串扫描里有 11 个名字同时含 `red` 与 `green`。
- pi 176 个测试里哪些会在有 Key 时打网络。只打开了 vitest 配置、zai 目录测试和一份带 `skipIf` 的 e2e。code search 的 12 条 minimax、21 条 zai 命中未逐文件打开。
- 没有运行 Go 或 Python 测试，因此不能确认这 266 个文件或 107 个 Python 方法今天仍是绿的。
- 没有厂商账单。文档里的美元是预留或 Result 估价，作者已写明不是发票。
- 没有找到一份「把父测试秒数和子测试秒数相加后写成墙钟」的错误发表值。相关实测是收集器归错类，以及 45m/75m 信封不一致。
- Go 博客表的 60/300/900 秒是「每个测试」还是「一组」，正文没有说。
