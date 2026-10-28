# 测试流程 SOP

编写测试、选择验证范围、执行验收、判定发布和维护证据时遵循本 SOP。工程规则见 [AGENTS.md](../../AGENTS.md)，当前来源与资格适用性按 [ADR-0055](../adr/0055-route-scoped-evidence-applicability.md)，此前阶段规则见 [ADR-0051](../adr/0051-scoped-verification-receipts.md)。本次工具归属与新入口见[ADR-0053](../adr/0053-verification-tool-ownership.md)及[迁移当前状态](../../.scratch/scratch-artifact-remediation/status.md)；原流程调整的实施事实见[原功能状态](../../.scratch/barness-ai-test-process-review/status.md)；具体功能另维护自己的唯一当前状态。

## 1. 确定变更与验证范围

开始前读取功能 spec/issue、当前状态和相关 ADR，检查完整 HEAD、已修改及未跟踪文件。把以下内容写入该功能当前状态：变更影响、公共测试入口、普通范围、live 路线或不适用依据、race/压力候选范围、实际输入与已有证据。pi 作为架构参考，行为依据 barness 的公共契约和自身证据。

| 变更 | 普通验证 | 大功能后续范围 |
| --- | --- | --- |
| 生命周期、队列、取消、共享状态、资源限额 | 公共 E2E 覆盖隔离、受控时序、取消与释放 | 对应并发路径、旧 Provider 回归、受影响设计负载 |
| 协议、历史、编码和用量 | 对应协议 E2E、已脱敏捕获回放 | 每条受影响自身 live 路线及资源路径 |
| CLI、发布判定、证据消费或维护 | 公共 CLI/子进程，含文件系统失败、取消与超时 | 对应执行、取消及并发读取路径 |
| SDK、模块、构建标签或难以限定的共享变更 | 扩大普通包与编译/vet | 按影响列出真实路线、race 和压力 |
| 纯文档、展示清单、审查记录 | 引用、格式、状态、来源/选定输入集合相等性 | 沿用来源适用的事实，不伪造新测试 |

参与运行、价格、追溯或证据消费的数据按行为变更验证。测试范围由影响决定；合并不自动触发全库重跑。**完成条件：每项变更都有验证去向，未执行项与不适用项分开记录。**

## 2. 日常开发：先失败，再实现

1. 在公共 Client 或 CLI/进程边界建立能复现问题的失败场景，保留实际命令、退出和原件。复杂功能优先 E2E；确需隔离时先列全失败方式。遵循 AGENTS.md 的测试原则。
2. 修改实现，重跑该失败场景，再运行受影响普通包。并发使用受控服务、hook/channel 确定时序与取消点，容量验证留到独立压力阶段。
3. CLI/消费者/harness 变更增加对应无厂商请求的进程检查；普通消费者读取留存夹具，真实负载生成器与消费者放在独立压力阶段。

在仓库根按功能设置变量，以下为命令模板：

```sh
TOOLS="$PWD/tools/verification"
FEATURE='<功能 slug>'
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage e2e -- \
  go test -json ./ai/e2e -run '<受影响测试表达式>' -count=1
# 交付前覆盖完整受影响包；CLI/证据消费者变化时追加对应包。
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage e2e -- \
  go test -json ./ai/e2e ./ai -count=1 -timeout=20m
```

每次生成新的 attempt，检查实际测试集合及必需 case ID，不能把零匹配、父测试 PASS 或缺场景当成完成。普通回执包含唯一主 `./ai/e2e` bundle，辅助包也保留完整库存与 audit。只跑工具进程场景的补充命令回执须与同来源普通证据一起交付，不冒充主 E2E 原生包。

`run.py` 固定 GOENV=off、空 GOFLAGS，清除继承的差分、压力、live、路线和凭据配置；非 live 隔离 HOME，并保留宿主 HOME 脱敏。独立构建/元数据命令也须隔离这些配置。无 Key 检查选择已知受控入口；不要直接在继承环境中启动整个 live-tag 测试包。

**完成条件：原失败场景转绿、受影响普通测试实际完成、命令退出0、断言与 audit 通过，父子进程没有隐式启动 race、压力、厂商请求或冻结 oracle。失败即停，保留独立失败 attempt 后再修复。**

## 3. 审查并冻结大功能来源

完成最终 payload、原生历史、用量、资源释放、取消、超时与租户隔离审查；修复后复核，再提交并冻结源码、测试、夹具、依赖、工具及实际输入。记录实际受测 SHA，不只记录 HEAD。

工具内建来源覆盖 `ai/`、本维护工具目录、go.mod/go.sum，含 Git 忽略的源码；不包含 Markdown、生成缓存和整个仓库。跨出该范围的实际代码、脚本、历史夹具、目录快照、授权及参与消费的文档，按文件枚举，用重复的 `--input <仓库相对文件>` 固定哈希；凭据不进入 argv 或输入快照。

每个阶段固定自己的实际来源和输入；阶段之间用原件引用及实际差异连接，不要求授权文书、普通夹具和消费者输入集合相同。相关产品行为改变时取得对应新普通回执；仅工具/测试变化验证对应过程，不自动使厂商资格过期。历史回执、pin、时间和状态保持，当前适用性另留结论。live 原输入可按原测量提交核验留存字节，不要求文档当前版本等于原版本；未提交的授权、捕获等仍须在不可变路径保留，找不到原字节即拒绝沿用。普通阶段仍核对当前实际来源和输入。

**完成条件：审查结果、完整来源/输入集合和冻结普通回执一致。随后按 live → race → 压力顺序推进；前项未通过或缺少依据时停在当前阶段。阶段完成、修复及合并均不自动启动下一阶段。**

## 4. 受影响 live：独立授权与原生资格

先区分原测量、当前适用性和新执行权限。目录核对本路线实际使用的型号，来源按实际产品差异判断；新增无关型号、补测试或修消费者不扩大 live。消费者改变先离线重验完整捕获，缺项只列对应路线。共享协议/生命周期/SDK改变按实际影响扩大，未归类产品变更不得声称无影响。需要新厂商事实时逐路线列出身份、场景、原件缺口、预算与来源，再取得明确付费授权；无新增厂商事实需求时写有依据的 NOT_APPLICABLE。

每次只持有所选路线凭据，使用新 grant/claim 及该路线原生预算；失败或取消仍消耗该次授权。保留原始授权、账本、构建 argv/退出、二进制 SHA 和来源绑定；候选、正式、controlled/replay 分别判定。地区六路线的冻结构建、source 注入、精确入口和正式资格按 [REGIONAL.md](../../ai/live/REGIONAL.md) 执行；其他路线按自身原生合同，环境转发见[工具合同](../../tools/verification/README.md#原生-live-环境合同)。

```sh
# 仅在该次范围、资格和预算获准后执行原生冻结二进制。
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage live --route '<已选路线>' -- \
  '<原生冻结二进制>' '<该路线的精确测试入口>' '<该路线时限>'
```

MiniMax 国际 Messages 仅允许固定十三加一消费：原完整失败包的十三项贡献，加绑定原 manifest 的 thinking-on 单项。原包仍 FAIL/exit1，新单项 PASS/exit0 但 complete=false；完整 VerifyCapture 对二者单独均拒绝。组合必须核验完整两包、原授权及权威账本、真实 producer（含 env -C 启动）、来源与 selectedInputs 原字节，逐项输出覆盖，不能删失败或合并预算。两包 mode 一致；controlled 不取得真实资格。单项精确入口/count/180s、1调用/1HTTP、零重试、2048输入/输出、USD0.01 在 claim/追加/发送前分别约束。Scope 的两个可选补项字段不改变旧 grant 的规范字节和 claim key；旧时限仍由各路线构造器固定1800s。未来付费另批，本轮不签发。消费者及测试来源变化只需离线重验和重建未批准二进制，不自动牵动所有 Provider。详见[实施合同](../../.scratch/minimax-thinking-supplement/spec.md)。


MiniMax 国际 Messages/M3/standard 单项在 wire 无 cache_control 时，独立写费允许按固定 M3 自动缓存依据标为 NOT_APPLICABLE；creation 缺失保持缺失及 UsagePartial，不补零。其他实际计费counter必须有效，显式null/正写/TTL/缓存控制或其他身份继续拒绝。新运行终态FAIL/audit先固定再生成最终seal，后续只读；旧两次FAIL/exit1包不重写/重封存，离线回放可用于回归不等于已获得完整组合资格。详见ADR-0055和[本次实施范围](../../.scratch/minimax-cache-billing-finish/spec.md)。

当前资格统一由已编译 releasegate 的原生路由校验，再由外层采集一次真实退出；保留各厂商独立 wire/用量/历史合同，不重复嵌套资格命令：

```sh
python3 "$TOOLS/consume_live.py" --binary '<releasegate 绝对路径>' \
  --bundle '<仓库相对原生 bundle>' --out '<仓库相对的新消费者目录>'
```

`consumers` 引用外层 schema 2 的 live-consumer.json；它实际采集唯一原生子进程，绑定来源、全部输入、资格结果、argv、时间、日志和真实数字退出。最终写失败或采集未完成保持非通过。该命令不请求厂商，也不发放 live 权限，不能用旧自采集格式或自报 `true` 替代。

可回放的已脱敏捕获收为对应路线夹具，保留来源与哈希。补夹具只验证对应普通范围。已覆盖正式完整合同的真实候选测量可经当前离线校验取得新的正式资格结论；原 stage/mode/测量时间不改，受控或回放不能补造真实场景。

Kimi 固定九加一组合按 REGIONAL.md 保留两个原生包及各自真实进程退出，清单只引用原件；外层消费者绑定两个包、生产进程回执/日志/二进制和可信账本的传递库存。单项范围固定场景、原 manifest 与独立一次性授权；普通受控组合不产生付费授权或真实资格，不牵动其他 Provider 复验。

**完成条件：所有受影响路线的构建、实际调用、完整必交场景、audit 与正式原生资格均有效，或 live 有依据地不适用。离线/受控通过不产生真实支持资格。**

## 5. 按影响选择人工 race

先判断并发影响。纯文档、展示或无共享状态变化的只读消费者可有依据地 NOT_APPLICABLE；claim、锁、进程取消和共享资源改变须列对应并发范围。适用时在普通及所需 live 完成后交付具体 argv、输入与预计耗时，等待人类选择。未授权的适用范围写 NOT_RUN；实施批准不代替该选择。

```sh
# 待人工选择该实际范围；可按影响缩小，须列出保留/未跑项。
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage race -- \
  go test -race -json ./ai/... -count=1 -timeout=75m
```

**完成条件：实际退出0、选定测试完成、stdout/stderr 无 DATA RACE，所有期望原生包与 audit 有效，来源及输入适用。race 不顺带开启压力/live；失败重跑或扩大范围需新的明确选择。**

## 6. 人工压力或有依据的不适用

race 通过后，队列、取消、共享状态、内存、限额、复杂度或大载荷变化由人选择受影响负载；其余情况写 `NOT_APPLICABLE` 和影响依据。默认普通消费者只读夹具，压力阶段才设置 BARNESS_AI_PRESSURE=1。

```sh
# 待人工选择；这是一份全集合模板，不是每次改动的默认命令。
python3 "$TOOLS/run.py" --feature "$FEATURE" --stage pressure -- \
  go test -json ./ai/e2e -run 'Pressure|DesignLoad|ToolArgumentsScaling' \
  -count=1 -timeout=20m
```

选择范围时对照实际测试发现集合：Policy、AdmissionBurst、Mixed、ClassifierImages、MiniMax/GLM SharedDesignLoad、Multimodal RuntimeDesignLoad、夹具写背压和工具写完整性等。工具参数复杂度变化包含 TestToolArgumentsScaling；生成器与依赖它的消费者同次运行。TestPressureLoopback 还需人选择有界循环数，未选则明确记未跑。

**完成条件：实际退出0，选定必交场景实际 PASS，资源/复杂度/完整性原件与 audit 有效，政策数值、来源和输入匹配。部分 PASS 或仅设置开关不足以完成。**

## 7. 更新当前状态并只读判定发布

每功能一份唯一可变当前状态；保留既有 Markdown 格式，并引用独立、不可覆盖的版本化 JSON 索引。状态包含影响依据、实际受测提交、当前提交、来源/输入、各阶段范围/argv/实际状态、原件路径与哈希、尚未执行项。

1. 核对所有阶段原件与当前适用性，保存真实 argv、数字退出（启动失败无原生退出码）、起止/墙钟、工具身份、日志和精确 `bundles/<run-id>/manifest.json` / audit 哈希。父子耗时不相加冒充墙钟。
2. 保存旧当前状态的快照；更新当前状态并计算内容哈希，在新的 record.json 中绑定它及实际阶段回执、impact、selection 与有限完整 requiredCases。字段以 [record.go](../../ai/internal/testkit/release/record.go) 和[工具合同](../../tools/verification/README.md#只读发布-cli)为准；合成测试记录不是实际阶段事实。
3. 在新的报告目录只读判定，保留实际命令/退出和报告。需要精确 CLI 退出码时使用编译后二进制，go run 会包装程序的非零退出。

```sh
releasegate -record '<不可变回执索引.json>' -out '<新的报告目录>'
```

| 五项事实 | 通过条件 |
| --- | --- |
| 提交/来源 | 当前状态哈希、完整执行来源和选定输入适用；收尾沿用有集合相等性及两个提交的记录 |
| E2E | 实际退出0，必需场景完成，唯一主包及所有辅助原件完整、audit 通过 |
| live | 各受影响路线的实际执行与原生正式资格/外层消费者有效，或 impact 与原因支持不适用 |
| race | 本次人类选择与实际 argv 一致且完整回执通过，或 impact 与原因支持不适用 |
| 压力 | 本次人类选择与负载原件一致并通过，或 impact 与原因支持不适用 |

退出0才表示 release.pass=true；缺失、过期、失败、NOT_RUN 或无依据的不适用都拒绝。只读判定不启动测试、vet、Node、厂商或任何后续阶段。修改已绑定的当前状态后，须新建索引并重新只读判定，不能悄悄改变旧索引所指字节。

**完成条件：实际判定与当前状态一致，报告和命令原件可追溯。普通验证完成、流程启用、等待人工与完整发布通过分别记录。**

## 8. 证据保留与维护

每次使用 `.evidence/<feature>/<stage>/[<route>/]<attempt>/`；原生包在 attempt 内 bundles/<run-id>。失败、START_FAILED、超时、中断、未知/缺退出、当前完整 attempt、传递依赖、授权/claim/账本及尚未变成有效夹具的捕获均保留。成功重跑不覆盖旧失败。scratch 中新增验收说明仅保存命令、退出、哈希与原件指针，不复制整棵胶囊或压缩包。

真实维护独立于测试和发布，开始前读取已审[清理方案第 6 节](../../.scratch/barness-ai-test-process-review/plan.md#6-存量清理)。按 scan → plan/清单审阅 → move → verify → 至少7×24小时观察 → inspect/purge 操作确切批次及清单哈希；首次 verify 成功才开始观察期，恢复/内容变化重新验证计时。失败/未知/有依赖/授权记录不能因年龄、大小、名称或一次成功被清理。相对目录、二进制与无扩展名依赖同样保留。

工具持独占锁、原子保存状态并重新核对引用、基线和库存；任一步失败即停，恢复拒绝覆盖不同内容。purge 内部重查全部条件，任何前置失败均零删除；仅删本批 payload，永久保留 metadata、引用与移动/恢复日志。不用 LATEST、裸删除或修改正文引用掩盖误隔离。

**完成条件：真实隔离有批次/哈希、移动与 verify 回执、新增缺失0；真实清理还需观察期满后的新鲜核验、PURGED、实际回收量及剩余依赖复核。未执行就分别记 NOT_RUN；临时合成清理自测不算真实清理。**

## 9. 收尾交付

交付变更、验证范围、实际 argv/数字退出、受测 SHA 与当前提交关系、回执入口、未跑/不适用依据及回滚方式。原测量按原固定版本核验，当前消费者独立核对路线适用性；固定续期/恢复控制器从活动实现删除。旧十门禁、冻结差分和历史控制器只保留原版本，不构建兼容 shim。旧批准、消费和暂停状态均不复活。

**完成条件：活动代码/CLI/文档使用同一合同，当前状态只有一份，全部声明有证据；若缺人工阶段则明确 release.pass=false。** 新流程的首次大功能闭环、真实隔离和永久清理分别按[计划第 8 节](../../.scratch/barness-ai-test-process-review/plan.md#8-完成条件与-review-对照)记录，不能用文档完成代替未发生的执行。
