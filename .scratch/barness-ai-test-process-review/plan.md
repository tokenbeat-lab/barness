# 测试与证据清理执行方案

日期：2026-10-22（Asia/Shanghai）。第 1 节入口迁移、七项及后续三项 Astra 复审修复的普通验收已完成；实际来源、回执、执行偏差及未运行范围统一见 [status.md](status.md) 和 [execution-log.md](execution-log.md)。第 2–8 节未启动，真实证据清理未开始。依据：[测试过程评审](research.md)、[测试系统地图](system.md)。

执行顺序：第 1 节完成入口迁移与验收，再启用第 2–5 节的新流程；第 6 节独立执行清理，观察期结束后才允许永久删除；最后按第 8 节逐项收尾。修改本计划不构成上述执行，也不授予 race、压力或付费调用权限。

## 决定

1. 去掉差分。pi 只作为架构和设计参考，不再作为行为规格。`BARNESS_AI_PIDIFF`、冻结 pi oracle 和 differential 发布条件退出活动代码及当前操作文档。
2. 保留公共 Client E2E 和 live。live 捕获能离线重放的，收成 E2E 夹具后再回归；离线夹具不产生真实支持资格。
3. 保留压力和 race，只在大功能的 E2E 通过后由人启动。改变厂商行为时，live 也必须先通过；没有厂商行为则写明不适用。各阶段失败即停止，不能自动扩大范围或重试付费进程。
4. 每次执行独立保存回执，失败原件不覆盖。只有当前状态、历史失败及其依赖均不再需要的成功包才能清理。
5. 2026-10-21 测得整棵 `.evidence` 为 189G，其中 `barness-ai` 61G、`provider-sites` 15G、`barness-ai-glm` 12G、`ticket14` 7.5G、`minimax-retained` 5.6G。这是历史测量，不是本次预期可释放量。

## 1. 准备、入口迁移与规范切换

### 1.1 执行前记录

执行时先读 `docs/agents/verification.md`。记录 `git status --short`、完整 HEAD、相关未提交/未跟踪源码、测试、依赖和配置；`git status` 退出 0 不代表工作区干净。先保存并提交参与执行的修改，再冻结来源，不能只凭 HEAD 认定实际受测字节。

确认没有其他 `go test`、`releasegate`、live 或证据写入者正在使用目标目录。清理工具必须实际独占取得 `.evidence/_quarantine/LOCK`，不能只检查锁是否存在；锁冲突停止，不自动移除别人的锁。中断后由维护者核对原进程和移动日志，再恢复工作。

开始时创建 `execution-log.md`，第一行记 `HEAD <完整 sha> 开始执行 plan.md`，随后记录每步时间、命令、来源、退出码和原件指针。地区站点仍由 `.scratch/barness-provider-sites/spec.md` 及其恢复 7 当前索引维护状态；本计划不恢复该轮、不发行新授权、不把历史失败改成通过。

### 1.2 先写失败场景，再改入口

以下为入口迁移的失败场景与验收合同，实际执行记录见当前状态。先在公共 CLI/进程或公共 Client 边界建立失败场景，留下可重复运行的原件，再修改实现；确需隔离测试时先列全失败方式。此处不先运行旧完整发布链。

| 边界 | 必须先覆盖的失败方式 | 验收原件 |
| --- | --- | --- |
| 普通入口 | 顶层未开压力但子进程偷偷开启；普通命令启动 race/live/Node oracle；继承环境误开开关 | 受控父子进程 argv、环境变量名、启动轨迹及数字退出码 |
| 发布判定 | 五项缺失、FAIL/NOT_RUN、空测试集合、过期来源、错路线、缺 manifest/audit、伪造不适用 | 公共 CLI 拒绝/接受回执，不能仅测试内部布尔值 |
| 阶段回执 | 重跑覆盖日志；非零退出、超时、中断、启动失败、缺/多个匹配 bundle 被写成 PASS | 独立 attempt、真实退出状态、发现清单和哈希 |
| 引用扫描 | Markdown 相对链接、代码中的真实路径、JSON 路径键值、同名不同目录日志、占位符、尾部标点、间接索引 | 合成目录中引用来源→规范路径及分类理由 |
| 清理/恢复 | 既有缺失误判为本次损坏；移动一半失败；扫描后引用变化；恢复覆盖原件；锁冲突 | 临时目录进程 E2E、移动日志和恢复前后哈希 |
| 永久删除 | 校验非零、观察期不足、批次错误、缺元数据、新引用指向隔离内容仍继续删除 | 拒绝场景非零退出且零删除；有效合成批次才允许删除 |

清理自测只用独立临时目录，禁止把真实 `.evidence` 当破坏性测试夹具。每项保留一份可重放的最终索引，不为每次断言制造胶囊。

### 1.3 迁移实际代码与测试合同

实现时先列出活动入口、调用方和依赖，再逐项修改，不能仅修改文档或顶层环境变量。

- 删除 `ai/internal/testkit/pioracle` 的活动依赖、差分测试/断言、冻结 npm 安装入口及只服务差分的账本/发布字段；公共行为仍需验证的断言改为独立 E2E 期望。历史 ADR、报告及原件不改写，不为旧内部路径增加 shim。
- 改造 `ai/release/cmd/releasegate` 和 `ai/internal/testkit/release`：发布判定按第 4 节五项回执进行，只读消费，不启动测试、vet、Node、race、压力或 live。新的调用合同为 `releasegate -record <回执索引.json> -out <新报告目录>`；旧自动执行阶段、强制环境开关和旧十门禁必交选择器退出活动入口，不保留兼容执行分支。
- 保留仍被使用的原生来源、manifest/artifact 哈希、脱敏和 live 资格消费者；发布时按本功能影响范围调用。它们是五项证据有效性的检查，不再隐式要求全库差分、所有历史路线重买或完整压力。输出区分证据无效、失败、未跑和不适用。
- 核查 `ai/e2e/minimax_release_contract_test.go` 的 `TestMiniMaxReleaseLoadConsumer`、`ai/internal/testkit/release/minimax_load_fixture_test.go` 及所有类似子进程：普通消费者测试改用已保存的受控夹具；实际负载生成及其完整消费者检查移入显式压力集合，取消隐藏启动。
- 核查 `ai/e2e` 所有 `pressureCase` 调用、`BARNESS_AI_PRESSURE`/循环开关、名称含 Pressure/DesignLoad 的入口及子进程。大型设计负载须受显式开关约束；保留低成本、确定性的取消/隔离/释放 E2E。删除过时合同断言，保留仍有效的安全与公共行为检查。
- CLI/发布工具、live harness 或依赖变化时，增加对应无 Key 公共进程检查；不能以 Client E2E 覆盖不到工具为由省略验证，也不能自动扩成完整 live 包自测。

验收：未设置开关的一次完整普通 E2E，其父子轨迹均无压力、race、厂商调用、冻结 oracle；只读发布 CLI 不启动验证子进程。源码检索中的剩余旧名称只能是明确标为历史的材料或拒绝旧入口的负例，不能存在活动依赖。缺少证明时不得切换规范。

### 1.4 实现两份有界维护工具

本节的 `.scratch/barness-ai-test-process-review/tools/run.py` 与 `tools/cleanup.py` 已实现并通过普通进程验收；清理测试只使用临时合成目录，真实清理仍未开始。后文命令保留为操作合同。执行时先核查已有工具/依赖，复用 `evidence.Run`、bundle 发现及审计能力；只补阶段回执和安全清理，不新增逐代验收控制器、厂商编排器或授权系统。实现文件遵守 500 行上限，按职责拆分。

`run.py` 的合同：

1. 接受 `--feature <slug> --stage <e2e|live|race|pressure>`，live 另有 `--route <route>`，`--` 后为真实 argv；验证 slug/route，拒绝路径穿越。只有相应阶段可启用对应开关；普通阶段清除继承的差分、压力、live、路线及凭据配置。
2. 每次独占创建 attempt：`.evidence/<feature>/<stage>/[<route>/]<attempt-id>/`。已存在则拒绝，不能用 `mkdir -p`、固定文件名重用或覆盖旧 attempt。
3. 保存 `command.json`、`stdout.jsonl` 或 `stdout.log`、`stderr.log`、`receipt.json`、实际数字 `exit-code`、起止时间、来源/工具链/选定输入摘要与文件哈希。启动失败无子进程退出码，写 `START_FAILED`；中断/超时保存真实返回码和状态；不能补写 0。
4. 设置 `BARNESS_AI_EVIDENCE_DIR=<attempt>/bundles`。现有 `NewRun` 仍写 `<attempt>/bundles/<run-id>/manifest.json`，不改变其合同。按 manifest 的 package 和本次独占基目录发现 bundle，将精确路径和哈希写入回执，禁止挑“最新一个”。
5. E2E/压力期望恰好一个 `./ai/e2e` bundle；race 记录各包 bundle，主 E2E 仍须唯一；live 记录所选路线唯一原生 bundle。无匹配、有歧义、缺 manifest/audit、必交场景未跑或审计失败均不能 PASS。
6. 非零退出或原件不完整仍落失败回执，保存已有日志及局部 bundle。外层超时覆盖构建、执行与收尾，记录取消和子进程释放；阶段结束显式判断回执，失败不能继续下一阶段。
7. live 只传递经路线合同选定的凭据/环境，凭据由可信 Host 注入；argv、回执和诊断不保存密钥值。工具不读取其他路线凭据，不申请或重用授权。

`cleanup.py` 只处理第 6 节的引用快照、批次、移动/恢复及删除，不启动测试或 live。其子命令和拒绝条件以第 6 节为准。

### 1.5 验收后切换规范

入口、回执工具及清理工具各自完成第 1.2 节相关普通进程 E2E 后，记录变更文件、实际 argv、退出码、原件和复审结论；本阶段不因实施批准自动启动 race、压力或 live。真实重负载验收留在第 3 节，未跑如实标记。

再同步 `docs/agents/verification.md`、`docs/barness-ai/README.md`、`ai/README.md`、`ai/live/REGIONAL.md` 的当前入口与门禁描述，更新 `GLOSSARY.md` 的差分/发布术语；新增 ADR 明确取代 ADR-0004、ADR-0017、ADR-0048 中冲突的当前流程，保留历史决定及证据。实现/普通验收先提交，规范切换再单独提交，写明差分退出及压力、race 改为人工阶段。

验收：活动代码、公共 CLI、文档和负例测试使用同一合同；规范切换提交引用实现和普通验收 SHA。发布状态仍由实际五项回执决定，不能把“规范已切换”写成发布通过。回滚以实现和规范两项变更一起回退，不恢复或删除已消费授权，也不改写新旧失败记录。

## 2. 日常改动：普通 E2E

本节仅在第 1 节切换完成后适用。每项行为改动先在 `ai/e2e` 建立本地受控 Provider 的公共 Client 失败场景，再改实现；纯函数规则确需隔离时先写失败方式。工具改动另覆盖受影响公共 CLI/进程。

```sh
# 以下命令均在仓库根执行；TOOLS、FEATURE 在每次会话显式设置。
TOOLS="$PWD/.scratch/barness-ai-test-process-review/tools"
FEATURE='<功能目录 slug>'
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage e2e -- \
  go test -json ./ai/e2e -run '<受影响测试前缀>' -count=1
# 合并前完整运行受影响包；ai 同包规则包含在这里。
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage e2e -- \
  go test -json ./ai/e2e ./ai -count=1 -timeout=20m
```

验收：真实退出码为 0、选定测试实际运行、对应断言与 audit 通过；父子进程没有开启差分、压力、race 或 live。纯文档变更只做引用、状态、格式及来源检查。默认本地试跑包不自动进入当前索引，但失败保留；需要交付回执时使用独立 attempt。状态写完后，成功且无引用的试跑包才能按第 6 节隔离。

## 3. 大功能完成

依次执行 3.1 → 3.2 → 3.3 → 3.4。前项不是 PASS 或有依据的不适用即停止。执行前在当前状态写明范围、输入及来源，人工 race、压力、付费调用分别记录明确选择，计划批准不代替该选择。

### 3.1 E2E 全量回执

冻结提交及完整执行来源后，按第 2 节完整包命令生成独立 E2E attempt。除退出码 0 外，要求发现恰好一个 package 为 `./ai/e2e` 的 manifest，位于 `bundles/<run-id>/`，已记录精确路径、哈希、实际测试集合和零发现 audit。不能检查 `<stage>/manifest.json` 或只凭目录存在放行。

更新功能唯一当前状态：受测 SHA、来源/输入摘要、attempt 路径、命令、退出码及 manifest 路径。后续相关源码、测试、依赖、目录或输入变化时重做本节；文档和索引收尾只有在完整相关来源集合相等时才能沿用，并注明受测提交与当前提交的关系。

### 3.2 live，或记不适用

影响厂商路线、型号、协议形状或支持资格时，先列出影响映射和每条路线的完整必交场景，再取得范围、账户资格与预算的明确授权。共享协议或目录变化须评估所有受影响路线，不能只按文件名选一家。无厂商行为变化时写 `live: 不适用` 并说明依据。

每条路线沿用自己的原生合同，单进程、单路线、唯一 Key、新 grant。地区路线严格按 `ai/live/REGIONAL.md`，不能用通用 `go test -tags live` 代替冻结构建：

1. 用 `regionalverify -source` 保存完整来源及 `sourceSha256`，确认与 3.1 来源一致。用 `-grant-template <combo>` 生成候选模板；正式阶段加 `-stage formal`，不得把 PENDING 模板当授权。
2. 在该次 live 的准备材料中保留源码摘要、构建命令、构建退出码、二进制 SHA 和来源绑定。用下述合同注入真实摘要，任何构建失败即停止：

   ```sh
   go test -c -tags live \
     -ldflags "-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=${SOURCE_SHA}" \
     -o "$BINARY" ./ai/live
   ```

3. 可信 Host 在清除其他路线、凭据和 endpoint 覆盖的环境中注入 `BARNESS_AI_LIVE=1`、所选 `BARNESS_REGIONAL_ROUTE/SERVICE/GRANT/KEY`；正式阶段另有 `BARNESS_REGIONAL_STAGE=formal`。通过 `run.py --stage live --route <combo> --` 执行冻结二进制，精确入口为 `-test.run=^TestRegionalCandidate$` 或 `-test.run=^TestRegionalFormal$`，`-test.timeout=1800s`。构建材料须被本次回执按路径及哈希关联。
4. 地区每进程最多 24 Calls/24 HTTP Attempts、零重试；其他路线使用自身合同，不能套用地区预算。保存 grant、授权回执、账本和原生文件；失败或取消也不能重用 grant。
5. 对精确发现的原生 bundle 运行 `regionalverify -qualify <bundle>`，正式资格另须 `-formal <bundle>`；保存消费者命令、退出码、日志及来源。其他路线使用自身消费者。消费者失败即停止，不能仅凭 Go 测试退出 0 或报告标题 PASS 继续。

验收：构建、live、所需原生消费者均成功；本路线全部必交场景 PASS，来源/目录/型号/SDK/预算及审计一致。候选通过不能代替正式发布资格，controlled/replay 不能认证 live；失败、未跑与不适用分别记录。

可离线重放的已脱敏捕获收进 `ai/e2e/testdata/<route>/`，先建立回放失败场景再完成接线，保留捕获来源及哈希；夹具 E2E 必须变绿。优先在最终冻结前完成已有捕获的夹具收录。新捕获落库若改变 3.1 来源，先提交并重跑 E2E，再用原生来源规则判定旧 live 是否仍适用：集合相等可引用原件；不相等则标记过期，重新取得所需授权，不能刷新旧 pin 或自动重买路线。

### 3.3 人工 race

E2E 和所需 live 均通过或 live 有依据地不适用后，交付实际命令、来源、输入、已有回执和预计耗时，等待人明确选择该范围。未选择时写 `race: NOT_RUN`、`release.pass: false`。

```sh
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage race -- \
  go test -race -json ./ai/... -count=1 -timeout=75m
```

验收：命令退出 0、所选包实际完成、stdout/stderr 无 `DATA RACE`、期望 bundle 和 audit 有效；与 3.1 完整相关来源及输入一致。race 不开启差分、压力或付费 live，子进程也不得偷开。失败及扩大范围需新的明确选择；相关源码变化先回到 3.1，不能复用旧普通回执冒充当前通过。

### 3.4 人工压力

队列、取消、共享状态、内存、限额或工具参数复杂度/大载荷路径改变时，人在 race 通过后选择压力范围；其他情况写 `pressure: 不适用` 及影响分析。

```sh
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage pressure -- \
  go test -json ./ai/e2e -run 'Pressure|DesignLoad|ToolArgumentsScaling' \
  -count=1 -timeout=20m
```

`run.py` 只在此阶段显式设置 `BARNESS_AI_PRESSURE=1`。上述为完整压力入口选择器，至少涵盖 Policy、AdmissionBurst、Mixed、ClassifierImages、MiniMaxShared、GLMShared、MultimodalRuntime 的负载及 `TestToolArgumentsScaling`；循环型 `TestPressureLoopback` 需另选有界循环次数，未选不冒充已跑。

实现迁移时列出所有压力开关、`pressureCase` 调用与生成负载的父子入口，对照实际测试发现集合，确认每个真实压力场景都在选择器或显式附加命令中。工具参数处理变更必须包含 `TestToolArgumentsScaling`。允许按影响缩小范围，但须列出保留/未跑项及原因；需要前置生成捕获的消费者与生成器须同次运行，不能拆开后仅因没有匹配测试而通过。

验收：退出码 0、预计测试和场景均实际 PASS、必需负载/资源/复杂度原件齐全且 audit 通过。来源、输入和政策数值须匹配本功能冻结集合；数字退出 0、部分子测试 PASS 或开关已设置均不单独构成验收。

## 4. 发布记录

每个功能保留一份唯一当前状态；原状态为 Markdown 时继续使用它，引用不可变回执索引。新 `-record` 输入是该状态指向的证据投影，不另维护第二份可变“当前状态”。

| 项 | 通过条件 |
| --- | --- |
| 提交/来源 | 当前发布 SHA、实际受测 SHA、完整相关来源及选定输入摘要；沿用须证明集合相等 |
| E2E | 退出码 0、实际测试集合、唯一主 manifest、artifact 哈希和 audit 有效 |
| live | 每条受影响路线执行及原生消费者成功、正式资格有效，或有依据的不适用 |
| race | 人工选择记录、退出码 0、范围/来源匹配、日志及原生证据有效 |
| 压力 | 人工选择记录、退出码 0、必交压力集合及原件有效，或有依据的不适用 |

索引为版本化 JSON，包含上述五项、功能/当前状态指针、影响范围、各阶段 PASS/FAIL/NOT_RUN/NOT_APPLICABLE、原因、attempt/receipt/manifest 路径及哈希；不直接信任调用者填写的 PASS。只读发布 CLI 须重新读取实际回执和原件，核对来源及路线，不启动验证命令；全部成立才退出 0 并输出 `release.pass: true`，缺失、过期、失败或未跑均非零退出。输出目录必须是新目录。

发布判定不再要求差分或隐式全路线重验，也不能绕过选中证据的完整性、脱敏与真实来源校验。历史十门禁报告仅证明历史来源。第 1.5 节规范切换前保持原发布规则，不能混用新旧条件宣称通过。

## 5. 今后的证据保留

功能根保持 `.evidence/<feature>/{e2e,live,race,pressure}`，阶段下是独立 attempt，live 再按 route 分组。attempt 内 `bundles/<run-id>` 是原生证据，外层命令回执与它绑定，不能只留下 manifest 文件名或阶段级退出码。

| 必须保留 | 满足无引用条件后才可隔离 |
| --- | --- |
| 当前状态点名的完整 attempt、manifest、audit、回执、日志及传递依赖 | 同一阶段更早的成功重跑 |
| 非零退出、START_FAILED、超时、中断、缺报告/退出码的原件 | 成功且没有当前/失败依赖的普通本地试跑 |
| 尚未成功收成夹具的 live 捕获；失败捕获继续保留 | 夹具回归已绿且不再承担真实资格、来源或失败依赖的成功捕获 |
| 开放功能当前索引、停止原件、seal、授权及消费者依赖 | 无保留引用的重复成功日志/胶囊 |

“无引用”须通过第 6 节统一扫描判定，不能凭大小、日志为 `ok`、文件年龄或目录名判断。失败原件连同命令回执保留，后续成功不能覆盖。`.scratch/<feature>/evidence/` 新增内容仅存命令、退出码、哈希和原件指针，不再复制 zip/tar.gz 或整棵胶囊；已有忽略规则保持。

## 6. 存量清理

第 1.2 节清理进程 E2E 通过后才可操作真实目录。候选范围仅为 `.evidence` 中运行包，以及 `.scratch/**/evidence/` 下 `*.log`。Git common directory 中的 grant/claim/账本不是候选；`.evidence/_quarantine` 元数据也不递归清理。

### 6.1 构建引用、分类与移动前基线

实现 `cleanup.py scan`，只读扫描并在独立批次写元数据，不能用 `git grep -l` 的文件名输出判断正文引用：

1. 枚举 tracked 文档/代码/索引及相关未跟踪当前功能文档，记录每份输入路径及内容哈希。直接读取正文；扫描错误或截断须报告，不能默认为无引用。凭据文件不进入输出。
2. 解析 Markdown 链接、代码片段中实际路径、JSON 路径键值及哈希清单。Markdown 相对路径以引用文档为基准；JSON/清单按已有格式的基准转换；绝对仓库路径规范化。解码链接转义、剥离 fragment，裸文本尾部标点按来源语法处理，不能用一个宽正则吞下整句。
3. 每条记录保存引用源/行号或字段、原始路径、规范仓库相对路径及分类：`CONCRETE`、`EXAMPLE_OR_PREFIX`、`PREEXISTING_MISSING`、`UNRESOLVED`。模板变量、占位符、纯目录写入前缀须有分类理由；不能把所有代码块路径都当示例。歧义先解决或保守保护候选，不静默丢弃。
4. 从当前状态、失败原件及索引/manifest/seal/哈希清单递归展开依赖。已引用 bundle 保留完整原生目录及外层回执；明确整体引用的目录保护整树。须识别相对和跨目录依赖，不能只看 Markdown 中直接出现的 `.evidence/` 字串。
5. 保存移动前确实存在的具体路径基线及类型/大小/哈希或目录库存，另存既有缺失清单。当前索引及必需依赖缺失直接阻断；其他历史缺失逐项记处置理由，不伪装成清理新增损坏。

此前 review 的只读演算得到 5,924 条粗提取路径，其中 652 条在移动前已不存在，包含示例和带标点路径；这不是本次可复用清单，必须重新分类。不能要求这些旧缺失“从隔离区移回”或重写历史文件消除它们。

批次元数据至少有 `refs.jsonl`、`classifications.jsonl`、`preexisting-missing.json`、`baseline.json`、保留依赖集合、输入哈希、候选清单及哈希。地区三个强制保护文件为：

```text
.evidence/provider-sites/final-acceptance-20261020/recovery7-preparation/index.json
.evidence/provider-sites/final-acceptance-20261020/recovery6-stopped/index.json
.evidence/provider-sites/final-acceptance-20261020/closed-cohort-6/seal.json
```

同时保护 `.evidence/minimax-retained` 及当前索引实际依赖。移动前断言这些文件存在、哈希已记录且完整依赖在保留集中。

### 6.2 生成并审阅候选清单

`cleanup.py plan --batch <batch-id>` 生成 `KEEP/QUARANTINE/BLOCKED` 清单、原因、预计字节及 Git 跟踪状态。按第 5 节保护失败/未知状态，不能把“无直接文档链接”当唯一删除依据。

- `.evidence` 可按一级目录或独立 run 分类；父目录整体移动与子目录移动不能同时入选，父目录引用不能被子目录细分绕过。
- scratch 日志使用第 6.1 节解析的精确规范路径，不能只按 basename 判断。`handoff-root.md` 的 `tdd/final-durable-verification.log` 须解析为 GLM ticket-13 原件并保留；同名不同目录分别判断。
- 引用集、基线和候选须一致；未解决引用若可能命中候选则标为 BLOCKED。未引用但状态不明的旧包先保留，不能自动当成功包。

验收：人工能按每个候选的依据确认它是无保留依赖的成功中间包；`new_missing: 0`，当前必需文件和依赖全部存在。完成本批清单审阅才允许 move；审阅后内容哈希变化则审阅失效。

### 6.3 隔离与移动日志

```sh
# cleanup.py 实现且自测通过后才可执行；batch-id 使用本次 scan 输出。
python3 "$TOOLS/cleanup.py" scan --repo "$PWD"
python3 "$TOOLS/cleanup.py" plan --batch '<本次 batch-id>'
# 清单审阅后使用确切哈希，不能用可变 LATEST。
python3 "$TOOLS/cleanup.py" move --batch '<本次 batch-id>' --plan-sha256 '<清单 sha256>'
python3 "$TOOLS/cleanup.py" verify --batch '<本次 batch-id>'
```

每批独占 `.evidence/_quarantine/<batch-id>/`，`metadata/` 留清单、基线、状态及追加移动日志，`payload/` 按原仓库相对路径保留内容。目标须不存在且处于允许根内；拒绝路径穿越、符号链接歧义、批次错误或跨文件系统隐式复制。

move 持锁后重新扫描，核对引用输入、保留集及清单哈希；新引用命中候选或来源变化即拒绝，不能沿用旧清单。每次 rename 前落移动意图并同步，成功后追加完成记录；错误即停，不移动后续项目、不继续删除。中断恢复先核对源/目标和日志，不能猜测已移动或补成功状态。

同文件系统 rename 不复制字节，但隔离内容仍在 `.evidence` 内，此时不释放整树空间。完成条件是移动日志与实际源/目标一一对应，已有失败原件及保留依赖未受损；不能把“mv 没报错”当完整验收。

### 6.4 校验、恢复与观察期起点

verify 须检查：移动前已存在的所有保留路径仍在原位且内容/目录库存有效；三个强制文件及依赖完整；当前引用没有新指向隔离 payload；新增缺失数为 0。scratch 引用也须校验，不能只查 `.evidence/`。历史既有缺失单独报告，不混入新增缺失。

校验非零或中断时，批次记 `VERIFY_FAILED` 并非零退出，禁止永久删除。按移动日志执行 `restore --batch <id>` 恢复受影响项目；原位置有不同内容时拒绝覆盖，交付冲突路径。恢复后原件和引用重新校验，通过前保持阻断。

仅在 verify 成功时落带 UTC 时间的 `VERIFIED` 状态，记录元数据/候选/内容摘要，并把批次、时间、结果写进执行日志。从该时刻开始至少观察 7×24 小时；期间发生恢复或内容变化，记新事件，剩余内容重新校验并重算观察期。有新引用需求先恢复，不改正文链接掩盖误隔离。

### 6.5 观察期后永久删除

唯一删除入口为 `cleanup.py purge`，不提供裸 `rm -rf`、`--force`、跳过校验或按 `LATEST` 选批次。下列 inspect 失败不能继续 purge；purge 自身仍须独立重查全部条件，不能信任外层 shell 或旧成功文件：

```sh
python3 "$TOOLS/cleanup.py" inspect --batch '<确切 batch-id>' && \
  python3 "$TOOLS/cleanup.py" purge --batch '<确切 batch-id>' --plan-sha256 '<清单 sha256>'
```

purge 须持锁并依次检查：批次身份/规范路径及元数据完整；状态为 VERIFIED、距观察期起点及最后一次恢复/内容变化满 7 天；移动日志和 payload 库存匹配；重新从当前 tracked/未跟踪文档和索引扫描引用、展开依赖；没有新引用指向候选；所有基线保留路径及强制文件仍在原位且完整。

任一检查失败、扫描不完整或状态变化，非零退出且**不开始删除任何 payload**。发现新引用先恢复相应原件，再生成/审阅新清单；不能用 7 天前的 `keep-paths` 放行。全部通过后只删该批 payload，保留 metadata；删除中 I/O 错误记 `PURGE_FAILED`，保留未删内容及日志，不补记完成。

验收：进程 E2E 证明第 1.2 节拒绝场景零删除；真实批次删除前有新鲜校验回执，删除后记录实际回收量、`.evidence` 体积、剩余依赖复核及 `PURGED` 状态。payload 完整删完且保留集通过才记完成；不删 refs、基线、移动/恢复日志和失败记录。

### 6.6 今后的日常清理

当前状态更新后，仅把已被替代、实际退出码 0、完整成功回执且无保留引用的 attempt 纳入新批；默认试跑同理。缺退出码、局部 manifest、失败/中断、仍承担资格或依赖的目录一律保留。不得用阶段级 `exit-code`、`find -mtime` 或旧扫描清单直接删除。仍走 scan → plan → move → verify → 观察期 → purge，每批元数据独立保留。

## 7. 停止使用与历史保留

第 1.5 节切换完成后，新的工作停止使用：

- `BARNESS_AI_PIDIFF=1`、冻结 pi oracle、旧 differential 发布合同及 Node 安装要求。
- 每次修复或普通 E2E 隐式启动整库 race、压力或 live 的父子入口。
- `releasegate` 旧自动 check/race 执行入口；新入口只校验五项回执。
- 新一代 `.scratch/barness-provider-sites/tools/recovery*_round.py`；恢复 7 保持暂停，不新增恢复 8。
- 为同一次验证复制普通、race 和外层胶囊；固定阶段目录覆盖日志及退出码。
- `git grep -l`/basename 判断引用、未分类宽正则保留集、可变 LATEST 及校验后无条件删除。

历史 ADR、失败、已消费授权和旧报告保持原事实；原生只读消费者仍需读取的材料进入保留集。旧内部执行路径直接删除，不新增兼容执行层。

## 8. 完成条件与 review 对照

实施状态分开记录，不因下一个功能尚未开始而伪造验收：

- **入口迁移完成**：第 1.2 节相关失败场景先于实现，普通公共进程 E2E 通过并留可重放原件；实际父子入口不隐式运行重负载；只读发布判定及文档/ADR/术语一致。真实 race/压力/live 尚未人工选择时仍为 NOT_RUN。
- **新流程首次验收完成**：一次真实行为改动按第 2 节先失败后通过；下一个大功能依第 3–4 节取得实际回执或有依据的不适用，人工选择有记录。未发生前标待验收，不能据此宣布发布。
- **隔离完成**：执行日志有来源、确切批次/清单哈希、移动回执、`new_missing: 0`，三个强制文件、scratch 日志引用及传递依赖完整，观察期已开始。
- **存量清理完成**：满 7 天后重新扫描、检查并成功 purge，metadata 永久保留；执行日志有真实回收量、删除后体积及保留集复核。只隔离尚未删除时不能勾选。

| Review | 修订位置 | 验收要点 |
| --- | --- | --- |
| 1 日志引用误判 | 1.2、6.1–6.2、6.4 | 按正文及相对基准解析；GLM 已链接日志保留，同名日志不混淆 |
| 2 校验失败仍删除 | 1.2、6.5 | 工具内部先完整校验再删；校验失败场景零删除 |
| 3 文档未覆盖执行入口 | 1.3、1.5、7 | 迁移 releasegate、负载子进程及差分合同，再切规范 |
| 4 清理前已存在缺失 | 6.1、6.4–6.5 | 既有缺失分类、有效基线、仅检查新增损坏，删除前重新扫描 |
| 5 固定目录覆盖失败 | 1.4、5、6.6 | 每次独占 attempt，失败回执及 bundle 不被成功重跑覆盖 |
| 6 地区 live 构建不完整 | 3.2 | 冻结来源、注入哈希、精确入口、新授权及原生消费逐步成功 |
| 7 manifest 位置错误 | 1.4、3.1 | 在独立 bundles 基目录按 package 唯一发现，记录精确路径 |
| 8 压力集合漏项 | 1.3、3.4 | 包含 ToolArgumentsScaling，逐项核对发现集及必交场景 |
