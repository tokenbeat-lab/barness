# barness 全局测试系统

这张笔记只回答测试内容是什么、各层打算证明什么、词在代码和文档里指什么、`.evidence` 与 `.scratch/**/evidence` 怎么互相指，以及编排是不是一套固定可重复的过程。效果好不好见同目录 `research.md`，这里不重复那份评审。体积只引用该文件已经写下、本次又对过句子的整树 `189G`（`research.md` 第 25、78 行）；本次没有再对整树做 `du`。

## 1. 范围与方法

打开并引用的主文件：

- `docs/agents/verification.md` 全文
- `docs/agents/issue-tracker.md` 关于证据落点的约定（第 12 行）
- `docs/barness-ai/README.md` 发布门禁（第 69–94 行）及测试入口 `ai/README.md`（第 444–493 行）
- `ai/live/REGIONAL.md`、`ai/live/doc.go`
- `GLOSSARY.md` 里差分、真实冒烟、支持矩阵、发布门禁、续期授权各条
- `docs/adr/0050-final-acceptance-renewal-authority.md` 对控制器与「不设任意轮次注册表」的决定
- 证据写入：`ai/internal/testkit/evidence/evidence.go`
- 门禁判定与命令：`ai/internal/testkit/release/gate.go`、`plan.go`、`ai/release/cmd/releasegate/command.go`、`execution.go`、`main.go`
- 控制器入口：`.scratch/barness-provider-sites/tools/README.md` 开头，以及 `final_round.py`、`recovery4_round.py`–`recovery7_round.py` 的文档字符串与模式列表
- 功能状态入口（只取定义当前门禁状态的开头，不评效果）：`.scratch/barness-provider-sites/spec.md` 第 9–24 行，`.scratch/barness-ai-minimax/delivery-status.md` 第 1–9 行
- 指针样本：`.scratch/barness-ai-glm/evidence/ticket-13/README.md` 与 `index.json`；`.evidence/provider-sites/final-acceptance-20261020/recovery7-preparation/index.json`；`.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/README.md` 与 `capsule.py`

命令（都没有跑测试套件）：

- 排除 `.evidence`、`.git`、`.scratch` 与冻结 pi 的 `node_modules` 后，`find` 统计 `*_test.go`：640
- `ls -1 .evidence | wc -l`：292 个一级名字。只列了一级；`provider-sites`、`barness-ai-glm`、`minimax-retained` 再各列了一级。这三个根上 `find -maxdepth 1` 没有 README 或 `index.json`
- `ai/live/support-matrix.json` 里 `"combo"` 出现 23 次

未知的不补数字。第 7 节列出没打开或没数过的部分。

## 2. 测试内容地图

仓库是单一 Go module `github.com/tokenbeat-lab/barness`（`go.mod` 第 1 行）。发布门禁的普通与 race 命令都是仓库根上的 `./...`（`ai/release/cmd/releasegate/command.go` 第 34–44 行）。本次排除 `.evidence` 与 `.scratch` 后，640 个 `*_test.go` 全在 `ai/` 下：`ai/e2e` 484，`ai/live` 94，`ai/internal/testkit` 53（含其下子包，未再拆），`ai/examples/visualeval` 4，`ai/release/cmd/releasegate` 3，库根 `ai/` 同包 2。`.scratch` 里的 Python 测试不在这 640 里。

| 层 | 位置 | 意图 | 触发 | 证明什么 | 不证明什么 |
| --- | --- | --- | --- | --- | --- |
| 同包隔离测试 | 库根 `ai/cost_test.go`、`ai/reasoning_budget_test.go`，包名都是 `ai`。门禁纯函数的失败方式在 `ai/internal/testkit/release/gate_test.go`，由 `gate.go` 第 8–10 行说明 | 公共 E2E 到不了的规则：多档成本、思考预算（两份测试各自写「E2E 覆盖不了，这是隔离测试」）。`BARNESS_AI_PIDIFF=1` 时成本与预算期望再与冻结 pi 逐位比。门禁 `Evaluate` 的误放行在隔离测试里先写失败方式 | 随 `go test ./...` 编译进普通套件。不需要 live 标签，默认不访问厂商 | 这些函数在给定输入上的规则，以及差分开关打开时与冻结 pi 的对应结果 | 不证明公共 Client 的一次调用、租户隔离或真实厂商行为。两份「隔离测试」注释各说各的规则，不能读成全仓库只有这两个测试文件 |
| 普通 E2E | `ai/e2e`，包名 `e2e`。`main_test.go` 第 1–4、22–24 行：外部程序，只打公共 `ai.Client`、本地受控 Provider 和宿主替身，不进库内部。`TestMain` 为这次 `go test` 建一个 `evidence.Run("./ai/e2e")` | 验收行为走宿主会用的公共边界。开发阶段先在这条边界写失败场景，再跑受影响的定向 E2E（`docs/agents/verification.md` 第 7、15–17 行） | 开发：`go test ./ai/e2e -run '<前缀>'`。发布 check 阶段把它包含在 `go test -json -count=1 -timeout=20m ./...` 里，并强制打开差分与压力（`command.go` 第 40、110 行） | 受控脚本下公共 Client 的协议、生命周期、取消、隔离与资源释放；每个 case 在 bundle 里留下可重放记录 | 默认不打厂商。未设 `BARNESS_AI_PIDIFF=1` 时差分 case 是 `NOT_RUN` 而不是 PASS。未设 `BARNESS_AI_PRESSURE=1` 时压力 case 同样是 `NOT_RUN`。一次定向 `-run` 不等于十项门禁 |
| 差分 | `ai/e2e/pidiff_test.go` 的 `TestPiDifferential`；oracle 在 `ai/internal/testkit/pioracle`。账本 fixture 在 `ai/e2e/testdata/pidiff/ledger.json`（`pidiff_test.go` 第 257–259 行附近的 `loadLedger`） | 同一 fixture、同一逻辑输入、同一响应脚本，分别驱动冻结 pi-ai 与 barness，比较受控 Provider 收到的请求和两边发出的结果；待决差异让该协议的差分门失败（`pioracle/oracle.go` 第 1–6 行，`pidiff_test.go` 第 22–26 行） | 环境变量 `BARNESS_AI_PIDIFF=1`，且本机有 Node 与 `npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node` 装好的冻结副本。发布 check 与 race 两次都会设置该变量（`command.go` 第 110 行；`docs/barness-ai/README.md` 第 81 行） | 开关打开且 oracle 能打开时，已登记场景与冻结 pi 1.1.0 的差异已分类，pending 为 0。扩展接入路径登记在账本里，不以 pi 差分 PASS 计（`GLOSSARY.md` 第 175–177 行）。当前合同要求恰好 10 条冻结差分（`ai/internal/testkit/release/plan.go` 第 45–51 行） | 开关未开：case 记 `NOT_RUN`，永不记 PASS（`pidiff_test.go` 第 28–30、244–247 行；跳过在 `evidence.Case.finish` 写成 `NOT_RUN`，`evidence.go` 第 298–301 行）。请求了差分但副本打不开则失败，不是跳过（第 249–252 行）。bundle 里零个 `PIDIFF-` case 时，差分门直接失败（`gate.go` 第 387–388 行）。DeepSeek Responses、原生图像、智谱三协议等扩展路线不因这条门通过而被算成 pi 已覆盖 |
| 压力 | `ai/e2e/policy_pressure_test.go` 的 `TestPolicyPressure`，case 前缀 `E08-policy-pressure-*`。设计负载的定义在 `GLOSSARY.md` 第 187–189 行。另有准入突发、混合负载等，门禁文档点名 `E08-admission-burst-pressure-*`、`E08-mixed-pressure-*`（`docs/barness-ai/README.md` 第 130、148、165 行） | 按策略注释里的设计负载把并发打满，再让一个消费者停止读取；用量占限额必须 ≤ 0.75，停读后以 `resource_limit` 结束，而不是一直占着内存（`policy_pressure_test.go` 第 22–36、42–54 行） | `BARNESS_AI_PRESSURE=1`。未设置则 `pressureCase` 跳过，证据状态为 `NOT_RUN`（第 101–108 行）。发布 check 与 race 都会设置该变量。未开时 P0 门失败（同文件第 22–25 行） | 该次运行在注释声明的负载下，限额余量、堆和停读行为与策略注释一致 | 不证明数值是通用默认，也不代替厂商 live（`GLOSSARY.md` 第 187–189 行；`docs/barness-ai/README.md` 第 171 行写明合成压力不替代 Luna 自己的 live）。单次压力有 3 分钟期限（`policy_pressure_test.go` 第 28–30 行），这是期限不是本次实测墙钟 |
| race | 没有单独的 race 包。命令是 `go test -race -json -count=1 -timeout=75m ./...`（`command.go` 第 42–43 行）。检测器覆盖同一次 `./...` 里会编译的测试，包含打开差分与压力后的 E2E | 看并发缺陷。政策把完整 `go test -race ./...` 列为正式覆盖的一项（`verification.md` 第 33 行）。并发场景的定向名字列在 `verification.md` 第 37–42 行，供人工选择定向 race 时参考 | 只有人明确选择 `releasegate -stage race -from <check-dir>`，或在聊天里明确要求该范围（`verification.md` 第 29–31 行；`releasegate/main.go` 第 1–3、47–48 行）。开发、合并、普通通过、方案批准都不启动它 | 该次冻结来源上，race 检测器开启时命令成功，并且其后的十项门禁接受这份独立 bundle | 不自动重跑。旧的开发期 race 只证明它自己的来源。定向 race 也要单独授权。`go test -race` 成功本身还不是 `release.pass` |
| live 包 | `ai/live`，构建标签 `live`。包说明在 `ai/live/doc.go` 第 1–11、66–81 行。地区路线的无 key / 受控 / 真实三级在 `ai/live/REGIONAL.md` | 发布前以及 SDK 或型号升级后，用低权限测试账户连厂商官方 API，喂支持矩阵。一个进程只跑一个操作 × Provider × 协议，只持有该组合的 key | 两道开关：`-tags live` 与 `BARNESS_AI_LIVE=1`，再加该组合的 key 变量。`go test ./...` 不带标签，因此普通套件到不了网络（`doc.go` 第 8–9 行）。地区真实资格还要新的 grant 文件与 `BARNESS_REGIONAL_*`（`REGIONAL.md` 第 19–33 行）。受控流水线用同一 live 标签二进制打 loopback，`mode=controlled` | 该进程自己的场景结果只为 PASS、FAIL、NOT_RUN、UNSUPPORTED 之一；报告可合并进矩阵里属于它的那一行。正式阶段通过的 `live-report.json` 才能进入支持矩阵的对应 row（`REGIONAL.md` 第 47 行） | 无 key 是 `NOT_RUN`，不是跳过成功。厂商故障在预算内重试后仍失败则是 FAIL（`doc.go` 第 69–71 行）。受控模式不能认证 live（`REGIONAL.md` 第 16 行）。兄弟站点、另一个型号、历史捕获都不能互相代替。候选目录通过不等于正式资格 |
| 发布门禁 | 命令 `ai/release/cmd/releasegate`；判定纯函数 `release.Evaluate`（`gate.go` 第 262–294 行）。十个门禁 ID：`go-test`、`go-test-race`、`go-vet`、`p0-offline`、`differential`、`live`、`redaction-audit`、`traceability`、`snapshots`、`required-artifacts`（`gate.go` 第 47–58 行）。人读条件表在 `docs/barness-ai/README.md` 第 83–92 行 | 一次发布判定：命令、离线 P0、差分、live、脱敏审计、追溯、快照、设计负载与四份独立效果都有效才允许发布 | check：`go run ./ai/release/cmd/releasegate -stage check -out <check-dir> ...`（README 第 74–76 行）。内部顺序是预检、`go vet ./...`、`go vet -tags live ./...`、带差分和压力的普通 `go test`、原生证据消费（`execution.go` 第 66 行；`verification.md` 第 27 行）。race 阶段见上一行。`-preflight` 只查外部输入。`-bundle` / `-race-bundle` 只重放已有证据 | 十项全过时 `Evaluate` 的 `Pass` 为真，race 阶段命令退出 0，执行报告 `releasePass: true`（`execution.go` 第 29–35 行） | check 全部通过只得到 `READY_FOR_RACE`、退出码 2、`releasePass` 仍为 false、race 为 `NOT_RUN`（`execution.go` 第 109–110 行；README 第 79 行）。`-bundle` 把命令门禁记为未执行，不能代替真的跑过。`go run` 会把程序的非零退出包成自己的退出 1，读退出码要用编译后的 CLI（README 第 79 行）。词汇表里的门禁句仍写「十三协议组合、十四身份」（`GLOSSARY.md` 第 195–196 行）；当前 `CheckReleasePlan` 要求恰好 23 个 live 身份（`plan.go` 第 32–38 行）。以哪份历史报告为准，看该报告自己的受测源码，不看词汇表旧句 |
| Python 控制器测试 | `.scratch/barness-provider-sites/tools/test_*.py` 18 个，`.scratch/barness-ai-minimax/tools/test_*.py` 7 个（本次 glob `test_*.py`，共 25）。打开的样本：`test_recovery7_round.py` 第 1 行，`test_renewal.py` 第 2 行 | 在合成仓库里测某一代验收控制器的 CLI：claim、拒绝、额度、模式。`test_recovery7_round.py` 写明 synthetic repositories、no vendor sockets；`test_renewal.py` 写明 keyless、captures are synthetic | 手工 `python3 -B .scratch/.../tools/test_<代>.py`。`tools/README.md` 第 43–47 行给出 `test_final_round.py` 与 `test_final_publication.py`，并说明测试结果不授予付费、不启动 race | 该代控制器在合成输入上的拒绝与计数和合同一致 | 不证明真实 HTTP、真实 key、正式 race 或 `release.pass`。方法个数本次没有数 |

旁边还有 4 个 `ai/examples/visualeval/*_test.go`，包名是 `main`（打开了 `unsent_test.go` 第 1 行）。它们测示例程序，不是 `ai/e2e` 那个外部测试包。`verification.md` 第 41 行把视觉取消的公共测试名列成并发覆盖入口；那些名字与这 4 个文件的对应关系，本次没有逐个对函数名。

## 3. 词汇

**差分。** 用冻结的 pi-ai 与 barness 对同一逻辑输入做对照，任何 pending 都过不了差分门；环境变量 `BARNESS_AI_PIDIFF=1` 才执行，否则记 `NOT_RUN`。位置：`ai/internal/testkit/pioracle/oracle.go` 第 1–11、27–39 行；`ai/e2e/pidiff_test.go` 第 22–30 行；门禁消费见 `gate.go` 第 350–388 行。冻结副本是 pi-ai 1.1.0，源码 SHA 与 lockfile SHA 写在 `oracle.go` 第 31–36 行。扩展接入路径是冻结 pi 没有路由的组合，走自身 fixture 与真实冒烟，登记在差分账本，不计为 pi 差分通过（`GLOSSARY.md` 第 175–177 行）。

**控制器。** 某一代验收的 Python 操作程序：准备、检查、按固定名字发放原生 grant、一次选一条路线跑子进程，并在全部成功后做最终计划与发布矩阵。它不自动做研究，也不自动启动 race。位置：`.scratch/barness-provider-sites/tools/README.md` 第 1–5 行（operator；「not a reusable round scheduler」）；同文件第 56 行称这份被审查过的程序为 controller。`docs/adr/0050-final-acceptance-renewal-authority.md` 第 55 行要求「Python控制器仍须实际完成严格原生消费」。`recovery7_round.py` 第 2 行：固定的第七次恢复，pending 批准永远不是付费许可。这不是 UI 控制器。

**live 包。** 带构建标签 `live` 的 `ai/live`：宿主同一个公共 Client，连厂商真实官方 API 做有界冒烟，结果进入 `support-matrix.json`。位置：`ai/live/doc.go` 第 1–16、72–81 行；`GLOSSARY.md` 第 179–181 行称此为真实冒烟。地区合同把受控 loopback 与真实 live 分开（`ai/live/REGIONAL.md` 第 14–16、19–33 行）。

**胶囊。** 一次普通、race 或外层原件的有界只读 tar.gz 留存，以及只流式读取、不解包、不执行其中命令的 reader。位置：`.scratch/barness-ai-minimax/evidence/ticket-13/final-gate/capsule.py` 第 1–6 行写明这是与更早 32MiB/192MiB/32k 协议分开的全套件留存协议，不执行捕获的命令。同目录 `README.md` 第 20–32 行把三份归档放在 `.evidence/minimax-retained/ticket-13-final-gate/`，并称 `capsule.py` 为新的全套件协议。中文「胶囊」在后续留存说明里指这种归档，例如 `.scratch/barness-ai-minimax/evidence/ticket-16/deadline-alignment/README.md` 第 20 行。它不是 `evidence.Run` 当场写出的那个目录 bundle。

**原生 Run。** `evidence.Run`：一次 `go test` 调用拥有的证据 bundle。位置：`ai/internal/testkit/evidence/evidence.go` 第 1–7、35、58–76 行。`NewRun` 在 `BARNESS_AI_EVIDENCE_DIR` 下再建一层运行 ID 目录；变量为空时默认 `<repo>/.evidence/barness-ai/<run-id>`。E2E 在 `TestMain` 里 `NewRun("./ai/e2e")`（`ai/e2e/main_test.go` 第 22–24 行）。live 进程在自己的 `TestMain` 里建 Run，包参数形如 `-tags live ./ai/live`（`ai/live/main_test.go` 第 45–63 行）。操作文档里的 native 还指 Go 侧写出的这份原件及其消费者（`regionalverify`、`releasegate`），与 Python 控制器自己的 claim、索引分开。

**manifest。** 该 Run 在全部 case 清理之后写出的 `manifest.json`：模块、包参数、起止时间、Go 与平台、git commit、版本表、每个 case 的 id、测试名、状态、重放命令、fixture 哈希与 artifact 清单。位置：`evidence.go` 第 117–154 行，`(*Run).Finish`。发布侧用 `release.LoadBundle` 从这份 manifest 读 case 状态（`ai/internal/testkit/release/load.go` 第 25–27 行）。

**grant。** 一次明确批准的有界授权，用完即占用，失败不退回、不能靠改 ID 或换目录重生。地区一次进程的形态是 `RegionalGrant`：模式、阶段、端点、源码 SHA、目录、SDK、账户与地区别名、型号、24 小时内到期、调用与 token 上限（`ai/internal/testkit/supportmatrix/regional_grant.go` 第 12–37 行）。续期形态是 `RenewalGrant`，绑定历史 cohort、人类批准、来源与二进制（`ai/internal/testkit/minimaxlive/renewal.go` 第 21–65 行）。词汇表的「续期授权」见 `GLOSSARY.md` 第 207–209 行。模板不是授权：`REGIONAL.md` 第 14 行写明 `-grant-template` 仍缺 ID、别名和有效期，因此不能发请求。

**普通（ordinary）。** 政策里的普通验证是 race 之前的阶段：预检、双 vet、打开差分和压力的完整 `go test ./...`、原生证据消费，然后停在 `READY_FOR_RACE`（`verification.md` 第 27 行）。实现里这一步的命令 ID 是 `go-test`，证据子目录名是 `offline`（`command.go` 第 40、103 行）。开发期的「定向普通 E2E」是受影响公共用例，范围由变更决定，合并本身不触发全量（`verification.md` 第 7–15 行）。控制器 README 在恢复 4 一节另写了一种 check：双 vet 加上完整 `ai/live` 的无 key 普通验证（`tools/README.md` 第 92 行）。那是该代的无 key 检查，不是十项门禁的 PASS。

**READY_FOR_RACE。** check 阶段没有失败时写入执行报告的状态。同时 race 保持 `NOT_RUN`，`releasePass` 为 false，进程退出 2（`execution.go` 第 29–31、109–110 行）。人还没有选择 race。

**release.pass。** 执行报告字段 `releasePass`（`execution.go` 第 16 行）。只有状态为 `PASS` 且该字段为真时退出码才是 0（第 33–35 行）。check 结束时的就绪不是这个真值。文档里写成 `release.pass: false` 的句子，指这份字段仍假，例如 `verification.md` 第 31 行。

**支持矩阵。** 每个操作 × Provider × 协议一行的真实冒烟状态，文件是 `ai/live/support-matrix.json`。只由该组合自己的报告合并，共享 adapter 的通过不标记其他组合（`GLOSSARY.md` 第 183–185 行；`ai/live/cmd/supportmatrix/main.go` 第 1–10 行）。本次该文件有 23 个 `"combo"` 键，与 `CheckReleasePlan`「恰好二十三个 live 身份」同数（`plan.go` 第 37–38 行）。没有逐行核对每个 combo 字符串是否与 `plan.go` 那一长串逐字相同，也没有在这张笔记里抄每行的 PASS 时间。

**scratch evidence 与 `.evidence`。** `.scratch/<feature>/` 是工单与规格所在（`docs/agents/issue-tracker.md` 第 5–12 行）：每个功能一份当前状态，验证结果记成指针。`.evidence/` 是 git 忽略的原件树（`.gitignore` 第 1–2 行），由 `evidence.Run` 或门禁、控制器写运行目录。两边的关系在第 4 节。

## 4. `.evidence` 与 `.scratch` 里 evidence 文档的关系

`.evidence` 整目录被 git 忽略（`.gitignore` 第 1–2 行）。`.scratch/**/*.zip` 与 `.scratch/**/*.tar.gz` 也被忽略（第 4–6 行），所以进 git 的是 scratch 里的说明、索引、reader 和哈希；大原件留在 `.evidence`，或留在被忽略的压缩包里。

Go 的默认写入点是 `.evidence/barness-ai/<run-id>/manifest.json`（`evidence.go` 第 31–33、64–72、150 行）。`releasegate` 未指定 `-out` 时写到 `.evidence/barness-ai-release/<时间戳>`（`ai/release/cmd/releasegate/main.go` 第 40–41 行）。控制器准备目录用 `.evidence/provider-sites/...`（`recovery7_round.py` 第 14 行的 inventory 默认路径，以及 README 里的 `prepare` 参数）。

本次 `ls -1 .evidence` 得到 292 个一级名字，根上没有总 README。三个被点名的子树根上同样没有 README / `index.json`。能对上归属的，是别处的索引写明的：

- **GLM。** `.scratch/barness-ai-glm/evidence/ticket-13/index.json` 第 5 行 `durableRoot` 为 `/Users/cyber/RestoX/barness/.evidence/barness-ai-glm/ticket-13`。同目录 `README.md` 第 1–9 行把这份索引当作可重复校验入口，原件在 `.evidence` 那一侧。`.evidence/barness-ai-glm` 的一级名字是 `final-review`、`protocol-exploration` 和 `ticket-01` … `ticket-13`，与 scratch 的票号目录平行，但字节在 `.evidence`。
- **provider-sites。** 规格把唯一当前状态指到 `.evidence/provider-sites/final-acceptance-20261020/recovery7-preparation/index.json`（`spec.md` 第 11 行）。打开的索引 `kind` 为 `CURRENT_STATE`，`state` 为 `PAUSED_AFTER_NATIVE_ORDINARY_TIMEOUT`，`releasePass` 为 false，`race` 为 `NOT_RUN`（该文件第 2–15 行）。scratch 规格是指针，机器状态文件在 `.evidence` 里。
- **MiniMax 留存。** `.evidence/minimax-retained` 的一级是一批 `*.tar.gz` 和 `ticket-13` … `ticket-21` 目录（本次只列了一级）。`capsule.py` 所在的 scratch README 把普通/race/verdict 原件指到 `.evidence/minimax-retained/ticket-13-final-gate/`。Git 侧留索引和 reader，原始成员留在那份归档。

`.scratch/barness-ai-minimax/delivery-status.md` 第 1–9 行是另一份功能状态：Delivery COMPLETE，日期 2026-10-19，写明来源 78843 的门禁命令退出 0/0/0/0、报告为真，普通与 race 主包各 12495 PASS。这是该文件自述的 MiniMax 交付，不是 provider-sites 规格里仍为 `release.pass: false` 的地区恢复。两份「当前状态」各管各的功能；仓库没有一份覆盖所有功能的单一 `release.pass`。

同目录 `research.md` 已测整树 189G。本次只核了那句话，没有重测，也没有按文件走进去。

一级名字（292，`ls -1` 原样）：

```text
barness-ai
barness-ai-glm
barness-ai-issue06-gate
barness-ai-issue06-logs
barness-ai-issue13-release
barness-ai-issue13-release.log
barness-ai-pi-1.0
barness-ai-readme-update
barness-ai-release
barness-ai-thinking-sampling
build-images-bundle.py
classifier-images-package
current-source-eleven-precheck-20261019-1
decisions-07-archive-audit
decisions-07-owned
decisions-authorized-archive-replay
decisions-authorized-package
decisions-authorized-source
decisions-candidate-authorized-full
decisions-candidate-authorized-full-final
decisions-candidate-authorized-offline
decisions-candidate-authorized-race
decisions-candidate-authorized-race-final
decisions-candidate-authorized-race-isolated
decisions-candidate-current-config
decisions-candidate-e2e-development
decisions-candidate-final-full
decisions-candidate-full
decisions-candidate-live-authorized
decisions-candidate-live-development
decisions-candidate-not-run
decisions-candidate-package
decisions-candidate-race
decisions-candidate-review-final
decisions-candidate-review-race
decisions-diff-gate-full
decisions-diff-gate-probe
decisions-diff-gate-reviewed
decisions-reading-priority-20261013
decisions-text-choice
decisions-usage-full
decisions-usage-full-final
decisions-usage-race-final
decisions-usage-reviewed
decisions-usage-reviewed-final
finalize-jev-recheck-20261016.py
formal-decisions-live-authorized
formal-development
formal-final-full
formal-final-race
formal-final-scoped
formal-final-verified-full
formal-final-verified-race
formal-jev-config-green
formal-not-run
formal-not-run-final
formal-package-raw
formal-pending-green
formal-pending-red
formal-publication-green
formal-publication-red
formal-token-budget-green
formal-token-budget-red
formal-usage-pin-green
glm-messages-cancel-20261019-1
glm13-consumer-diagnostic-r2
glm13-corrective-current-r1
glm13-current-r4
glm13-legacy-current-r3
issue08
issue12-executions
issue12-full
issue12-preflight
issue12-proxy-green
issue12-proxy-red
issue12-race
issue13-current-live
issue13-final-gate
issue13-gate
issue13-green-cli
issue13-green-inventory
issue13-green-live-proxy
issue13-green-race
issue13-green-review
issue13-known-pressure-reproduction
issue13-live-harness
issue13-red-cli
issue13-red-inventory
issue13-red-live-proxy
issue13-red-live-proxy-actual
issue13-red-race
issue13-red-review
issue14-complete
issue14-final
issue14-full
issue14-full-race.log
issue14-live
issue14-live-race.log
issue14-not-run
issue14-review-green.log
issue14-review-red.log
issue14-trace
issue14-trace.log
issue15-final
issue16-full
issue17-accepted
issue17-capture.log
issue17-confirmed
issue17-current-capabilities-green.log
issue17-deepseek
issue17-delivery-audit.log
issue17-delivery-final-audit.log
issue17-final
issue17-gate-regression-final.log
issue17-live
issue17-live-race-final.log
issue17-live-race.log
issue17-node-final.log
issue17-node.log
issue17-optimized-python.log
issue17-portable-verify.log
issue17-pressure-diagnosis
issue17-release
issue17-runner-final.log
issue17-targeted.log
issue17-tools
issue17-verified
issue18
jev-baseline-20261012
jev-baseline-recheck-20261016
minimax-current-eleven-gate-1
minimax-current-eleven-gate-2
minimax-current-eleven-glm-1
minimax-current-eleven-glm-2
minimax-current-eleven-publication-1
minimax-current-eleven-remaining-1
minimax-current-visual-gate-1
minimax-current-visual-publication-1
minimax-current-visual-renewal-1
minimax-global-candidate-4e8c59a
minimax-global-candidate-complete
minimax-global-catalog-pins
minimax-global-final-keyless
minimax-global-formal-glm
minimax-global-glm-corrective-1
minimax-global-glm-corrective-1-parent
minimax-global-glm-corrective-1-reviewed
minimax-global-glm-corrective-1-reviewed-parent
minimax-global-glm-r1-rejected
minimax-global-glm-r1-rejection-parent
minimax-global-legacy-final
minimax-global-legacy-final-parent
minimax-global-origins
minimax-global-origins-parent
minimax-global-r1-matrix-validation
minimax-launches
minimax-provider-sites-final-round2-raw
minimax-provider-sites-round2-catalog-pins
minimax-release-repair-effects-1
minimax-release-repair-final-1
minimax-release-repair-glm-1
minimax-release-repair-glm-publication-1
minimax-release-repair-incomplete-gate-1
minimax-release-repair-keyless-negative-1
minimax-renewal-home-redaction-1
minimax-renewals
minimax-research-captures
minimax-retained
minimax-routed-final-1
minimax-routed-final-compiled-1
minimax-routed-final-plan-preview.json
minimax-routed-glm-1
minimax-routed-glm-1-parent
minimax-routed-glm-1-pre-review-83ff7d6
minimax-routed-glm-publication-check-1
minimax-routed-glm-publication-rejected-1
minimax-routed-helper-history-83ff7d6
minimax-routed-legacy-1
minimax-routed-legacy-1-parent
minimax-routed-network-diagnostics-1
minimax-routed-observer-check-1
minimax-routed-preflight-retirement-83ff7d6.json
minimax-stages
minimax-three-preparation-control-1
minimax-three-preparation-control-2
minimax-visual-deadline-diagnostic-1
minimax-visual-gate-control-1
minimax-visual-publication-control-1
mixed-answers
mixed-answers-full-race.jsonl
mixed-answers-full-race.stderr
mixed-answers-full.jsonl
mixed-answers-full.stderr
mixed-answers-green.log
mixed-answers-package-audit-final.log
mixed-answers-package-audit.log
mixed-answers-race.log
mixed-answers-red.log
mixed-answers-regression-final.log
mixed-answers-regression.log
mixed-answers-verification.log
mixed-answers-vet-live.log
mixed-answers-vet.log
pre-push
pressure-integrity
provider-sites
race-optimization
release-followup
runtime-06-audit-sample.log
runtime-06-audit-sample.txt
runtime-06-final-race
runtime-06-full
runtime-06-full-race-retry.jsonl
runtime-06-full-race.jsonl
runtime-06-full.jsonl
runtime-06-history
runtime-06-pack-help.txt
runtime-06-package-args.json
runtime-06-race.jsonl
runtime-06-review-race.jsonl
runtime-06-target.jsonl
runtime-06-timeout-post-audit.txt
runtime-06-verify-eys866k4
runtime-06-vet-live.txt
runtime-06-vet.txt
ticket11-final-race.log
ticket11-first-red
ticket11-fixed-race.log
ticket11-full.log
ticket11-live-not-run
ticket11-metadata-inventory-symlink-green-assertion-error.json
ticket11-metadata-inventory-symlink-green.json
ticket11-metadata-inventory-symlink-red.json
ticket11-metadata-symlink-red.json
ticket11-probe-final.test
ticket11-probe.test
ticket11-seal-before-publication.json
ticket11-seal-optimized-result.json
ticket11-seal-receipt-c05a254f2a7d.json
ticket11-seal-receipt-e39ebe2489e1.json
ticket11-seal-result.json
ticket11-seal.py
ticket11-sealed
ticket12-deadline-red
ticket12-debug-1
ticket12-debug-2
ticket12-live-not-run-d4cd265
ticket12-live-not-run-e2c8287
ticket12-live-not-run-e6548ae
ticket12-live-not-run-f7c9bcf
ticket12-provenance-debug
ticket12-targeted-receipt.json
ticket13-doc-baseline
ticket13-final-1721024
ticket13-final-e633b3f
ticket13-live-admission-20261016T222549.801919Z
ticket13-live-formal-20261016T230251.995980Z
ticket13-prerequisites-20261016T221348.597703Z
ticket14
ticket14-decisions-direct-20261017
ticket14-full-recovery-20261017074119Z
ticket14-remaining-20261017073529Z
ticket14-remaining-integrity
ticket14-typesafe-20261017T055619.971418Z
ticket14-typesafe-integrity
ticket14-typesafe-matrix-repro-20261017
ticket14-typesafe-r2-20261017T061932.423266Z
typesafe-native-snapshot
unary-early-red
untracked-evidence-archives
usage-build-final.log
usage-build.log
usage-captures
usage-focused.log
usage-full-final.jsonl
usage-full-final.stderr
usage-full.jsonl
usage-full.stderr
usage-green-final.log
usage-green.log
usage-jev-verification.log
usage-package
usage-package-audit-final.log
usage-package-audit-preliminary.log
usage-race-final.log
usage-race.log
usage-red.log
usage-reviewed-final-green.log
usage-reviewed-final.log
usage-reviewed.log
usage-vet-live.log
usage-vet.log
```


## 5. 编排

从一次行为变更到发布判定，数据经过三层固定程度不同的东西。下面是一条流，不抄控制器 README 的路线表。

1. **开发期普通验证（政策固定，命令范围不固定）。** 行为变更先在公共 Client 或 CLI 上建立失败场景，再跑受影响的定向 E2E（`verification.md` 第 7 行）。生命周期、协议、发布工具、纯文档、共享模块各有一张「普通范围 / 建议 race 范围」表（第 9–14 行）。相关行为或输入变了才重跑受影响范围；合并本身不触发全量（第 15 行）。此阶段不启动 race，也不授予新的厂商调用（第 23、29 行）。
2. **付费与冻结（政策固定，每次批准单独）。** 付费或长时验收前审查最终 payload、历史、取消、超时、用量、释放和租户隔离，然后冻结来源、构建工具和选定输入（第 21 行）。验证计划获批不等于新的厂商调用预算（第 23 行）。地区六条路线还要各自的 grant，不能沿用旧 GLM / MiniMax 授权（`REGIONAL.md` 第 1、19–28 行）。
3. **发布 check（命令固定）。** `releasegate -stage check` 先核对固定输入；通过后按序跑 `go vet ./...`、`go vet -tags live ./...`、`go test -json -count=1 -timeout=20m ./...`（环境只有工具链变量，另加 `BARNESS_AI_PIDIFF=1`、`BARNESS_AI_PRESSURE=1`、`BARNESS_AI_EVIDENCE_DIR=<check>/offline`），再做原生证据消费（`command.go` 第 34–44、47–55、99–110 行；`execution.go` 第 66–110 行）。任一步失败就停。全部通过则报告 `READY_FOR_RACE`，退出 2，`releasePass` false。`-collect-all` 只继续收集 check 诊断。
4. **人工 race（政策要求人另选；命令固定）。** `-stage race -from <check-dir> -out <race-dir>` 复核原 check 的来源、输入和回执，再跑 `go test -race -json -count=1 -timeout=75m ./...`，证据目录为 `<race>/race`，同样打开差分和压力。十项门禁全部通过才退出 0。来源或输入变了就拒绝沿用旧普通结果（`verification.md` 第 31–33 行；README 第 81 行）。
5. **live 证据从旁路进入门禁，不在 check 的 `go test ./...` 里发生。** 人另行启动带标签的 live 进程，bundle 里有 `live-report.json`，再用 `supportmatrix` 合并进矩阵，并把 bundle 目录作为 `-live` 交给 `releasegate`（`doc.go` 第 75–81 行；README 第 74 行）。矩阵行只代表该组合自己的报告。
6. **某一代控制器把「多个付费进程 + 一次最终 check/race」收成该 cohort 的固定顺序。** `final_round.py` 是 `provider-sites-final-acceptance-2`（该文件第 2、14 行）。`recovery4_round.py` 至 `recovery7_round.py` 各自声明只做第 4 至第 7 次恢复，模式都是 `inventory`、`prepare`、`preview`、`check`、`summarize`、`native-grant`、`run-selected`、`final-plan`、`finalize`（四份文件的 `choices` 相同，本次读了第 12 行）。`final_round.py` 多 `promote`、`prepare-responses-retry`、`run-responses-retry`。README 写明这些是有限集合，不是可复用的轮次调度器（第 3 行）；ADR-0050 第 16 行写明不采用调用方自选的新轮次入口。控制器的 `check` 可以通过，仍然要在 `finalize` 之后重建 CLI、冻结来源，再跑一次新的 releasegate check 才可能出现新的 `READY_FOR_RACE`；race 仍要人另选（README 第 29、78、98 行）。

政策第 44–48 行要求每阶段记下真实 argv、数字退出码、起止时间、范围、来源和原件指针；未跑、失败、输入过期不能互相顶替。每个功能只留一份当前状态。

因此：`(a)` 稳定的是 `verification.md` 的阶段顺序、人另选 race、付费另授权、十项门禁的 ID，以及 `evidence.Run` / manifest 的形状。`(b)` 稳定的是 `releasegate` 里那四条 argv（两种 vet、普通 20 分钟、race 75 分钟）和差分/压力/证据目录的环境变量。`(c)` 只对一代 cohort 固定的是 `final_round.py` 与 `recovery4_round.py`–`recovery7_round.py`：每一代有自己的 ID、schema、额度、关闭 seal 和路线顺序，下一代是新文件，不是同一个程序的参数。

工具 README 开头仍写「当前固定入口是 `recovery4_round.py`」（第 1、81 行），并声明权威状态在规格。规格当前状态（Updated 2026-10-21）写的是恢复 7 的准备索引、恢复 6 已关闭、正式 race `NOT_RUN`、`release.pass` false（`spec.md` 第 9–24 行）。两句同时在仓库里。本文不把 README 的「当前入口」改写成规格里的恢复 7，也不把规格改写成 README 的恢复 4。

## 6. 没有单一永恒脚本的地方

没有一个程序从「改一行行为」一直跑到「十项门禁退出 0」，也没有一个轮次参数能生成下一次恢复。ADR-0050 把「无限轮次注册表」明确排除；控制器 README 把各代写成有限集合。`recovery7_round.py` 不会去跑恢复 4 的付费入口，恢复 4 的活动入口在本代 seal 出现后拒绝（`recovery4_round.py` 第 16–17 行一类的关闭判断；恢复 5、6 的文件同样在本代 seal 存在时拒绝非 `summarize` 的模式）。

仍然可重复的是上一节的 `(a)` 和 `(b)`：同一份冻结源码上，`releasegate` 的 check 命令与退出码语义可以再执行一次；人选择之后，race 命令也可以再执行一次。每一代控制器在自己的合同里可重复（同一 `prepare` / `check` / `run-selected` 参数），不能把恢复 7 的 argv 当成恢复 4 或下一次尚未写下来的恢复。Python 的 `test_*.py` 重复的是该代合成拒绝，不是发布判定。

MiniMax 交付状态文件自述 2026-10-19 的完整门禁已通过。provider-sites 规格自述地区恢复仍未取得新的 `READY_FOR_RACE`。全局地图因此是「门禁命令固定，功能 cohort 各有一代控制器」，不是「仓库此刻有唯一一条正在跑的永恒流水线」。

## 7. 未核实

- 没有重跑 `go test`、`releasegate` 或任何 Python 测试。表里的「证明什么」来自打开的注释和判定函数，不是本次运行结果。
- 没有重测 `.evidence` 体积。189G 只来自同目录 `research.md` 第 25、78 行，本次对过句子。292 是本次 `ls -1 | wc -l`。
- 292 个一级名字里，只对 GLM 票 13、provider-sites 恢复 7 准备索引、MiniMax 票 13 胶囊 README 这三处做了指针阅读。其余 `minimax-*`、`decisions-*`、`issue*`、`ticket*` 目录的归属没有逐个打开根上说明；很多根上就没有 README。
- `support-matrix.json` 只数了 23 个 `"combo"` 键，没有逐行比对 `plan.go` 第 32 行那 23 个名字，也没有核对每行 outcome。
- `GLOSSARY.md` 第 195–196 行的十三组合 / 十四身份，与 `plan.go` 的 23 个身份、README 门禁表的 23 条必交记录并存。哪一次历史发布用哪一句，要看该次报告；本次没有重放那些报告。
- `ai/internal/testkit` 的 53 个测试文件没有按子包拆开。`ai/examples/visualeval` 的 4 个 `package main` 测试没有对上 `verification.md` 里的函数名。
- Python 25 个 `test_*.py` 没有数测试方法，也没有读完每一代的合同常量。
- 控制器 README 写恢复 4 为当前入口，规格写恢复 7 为当前状态。哪一个是下一次人手会执行的命令，以规格自己说的「唯一当前状态」为准时是恢复 7 的准备索引，但 README 开头的句子尚未在本次改动；本次没有执行其中任何一个。
