# xAI图片目录激活独立审查

范围：工单06；仅源码目录激活与必要中文评估当前配置修订。原实测、付费授权和正式资格保持。Standards初审1项P2，由同一实现者修复；复核0遗留项。Spec和配置增量两轴均0项。实际普通结果另见唯一当前状态。

---

## Standards

部分完成：目录激活的静态审查发现 1 项行动问题。

- **[P2] 硬违反：当前目录测试绑定历史激活版本。** `ai/e2e/xai_images_builtin_test.go:46` 同时要求当前 `cat.Version` 等于固定 `2026-10-29.1` 和历史 `facts.ActivationVersion`。第 176 行再次固定结果的当前版本。未来只增加其他 Provider 的型号并更新根目录版本时，这两个 xAI 测试仍失败。这违反 ADR-0055 的按路线判断适用性要求，也违背 `AGENTS.md` 原则 9 的长期可验证性。请沿用 `ai/e2e/catalog_snapshot_fixture_test.go:11` 的 `currentCatalogVersion(t)`。原激活日期和候选版本/hash继续独立固定，不得反向约束当前根目录版本。

已验证：静态读取 `07de7542..cec2bd09`。排除根版本和新 xAI 图片型号后，发布快照的其他型号元数据完全相等。候选 JSON、版本/hash和历史原件没有改动。新增内置价格按调用独立构造，既有目录 clone 继续独立复制指针和输出价目行。四个受影响 Go 文件均少于 500 行。重复价格文字附有候选独立演进原因，符合项目复用规则。

未验证：未运行测试或厂商请求。本结论不宣称普通验收或发布通过。

Standards 共 1 项；最严重项为当前目录版本与历史来源耦合。

## 修复复核

已完成：静态复核 `43821cb67e4682824d0e2c1a4303b714bb5ac69b`。

原行动问题修复。当前目录与结果使用 `currentCatalogVersion(t)`。历史激活版本与原目录版本独立记录。没有新增硬问题。

未验证：未运行测试或厂商请求。

---

## Spec

已完成：`07de7542..cec2bd09a874ef66e2c46ed4cab19df4466d57cb` 的 7 文件增量审查。可行动缺陷：0。

- 工单 06 要求「激活 xAI 目录」并限制自身实测范围。`ai/catalog.go:61–64` 将独立 xAI 图片条目加入默认 BuiltinCatalog。`ai/catalog_images.go:39–49` 固定 5 参考图、2 输出、三档分辨率、low/medium/auto 和 1:1/16:9，未放入候选 10 输出或全部比例。
- 规格要求「编辑 N 仅允许省略或 1」和封闭类型选项。增量没有改选项转换或守卫。`ai/xai_images_options.go:48–59` 继续限制编辑单输出、单图比例及目录授权。默认 Client 离线重放证明无需自带 Catalog 即可调用获准 Binding。
- 规格第 180–181 行要求自身证据、独立来源和价格快照。逐字节核对 `official.json`、`live-contract.json` 与父任务独立提取文件相等。再独立读取原生 bundle：6 份 HTTP 请求/响应、状态/请求头、7 张图片哈希/字节/MIME、收费、ticks、UsageReporting、CostEstimate 与响应身份全部相等。原 ticks 合计 5100000000；token 实际未报告。
- `ai/e2e/xai_images_builtin_test.go:145–177` 验证冻结 wire、图片哈希与独立收费。候选 JSON 原字节未改，版本/hash 仍为 `xai-images-candidate-2026-10-29` / `72f8…3670`。价目来源保留 2026-10-29。比较新旧发布快照，既有聊天、图片和分类条目全部相等；仅总版本/hash 与新增图片条目变化。
- 原生 provenance 仍为 mode=live、stage=candidate。增量不改原授权、报告、正式资格器、race、压力或发布结论。

已验证：只读核对既有回执 `20261029T014640.050665Z-pt0b7kof`。退出 0，1044 个测试事件 PASS，audit findings 为 0。7 份增量文件与实测来源哈希一致。默认发现、6 个原生回放和相邻路线检查通过。

本次没有运行测试或厂商调用。此复核不产生正式资格或 release.pass=true。

Spec：0 项。

---

## Standards

已完成：中文评估目录固定值修复的静态审查。行动问题为 0。

范围：`43821cb6..9549cf4c` 的七个文件。依据：AGENTS.md、测试流程 SOP、ADR-0055 和本次目录激活合同。

已验证：静态差异与字节比较确认以下事实。

- 三份 `v1` 配置与 Git `b8080341` 的原字节相同。
- 三份新增 `v2` 配置只改变配置版本、目录版本和目录 hash。数据集、型号、任务说明和有限预算相同。
- CLI 默认值及当前 E2E helper 选择 `v2`。显式配置仍走原有严格读取流程。
- `Plan.Load`、历史包校验、型号价格和网络调用路径没有变化。旧固定值继续拒绝当前目录。没有新增兼容分支或自动升级。
- 新公共 E2E 检查保留原配置哈希，并检查当前配置加载与任务范围相等。根工作区 README 的当前说明与 `v2` 默认值一致。
- 四个受影响 Go 文件均少于 500 行。本次变化限于配置版本及选择，没有扩大职责或新增抽象。

未验证：未运行测试或厂商请求。未在根工作区执行 Git 操作，未修改代码、状态或工单。

Standards 共 0 项，无待修复问题。

---

## Spec

已完成：`43821cb6..9549cf4c960587036ced36130fb7fee8f60ba268` 的 ChineseEval 目录 pin 修复审查。可行动缺陷：0。

- 根目录激活更换整体目录版本/hash，当前默认配置必须随新快照显式修订。本增量只增加三个 v2 配置，并修改 CLI 默认和当前 E2E helper；属于激活的必要相邻修复，没有扩大模型能力或预算。
- 独立比较 v1/v2 原始 JSON：只有 `version`、`catalogVersion`、`catalogHash` 三项变化。新 pin 为 `2026-10-29.1` / `f44d…5348d`；模型、路由、数据集、提示、单位和有限预算全部相等。三个 v1 文件与基线 Git 字节完全一致。
- 规格要求价格快照可追溯，历史不覆盖。`ai/examples/chineseeval/cmd/chineseeval/main.go:62–66` 只选择新默认文件。`ai/examples/chineseeval/plan.go:102–103` 的严格 Load 校验没有改动；显式旧配置仍拒绝当前目录，不自动换 pin。
- `main.go:50–51` 的历史 verify 分支先于默认配置选择。`cmd/chineseeval/verify.go:20–28` 仍读取 bundle 自带配置/目录，`recorded.go:17–40` 继续使用原快照或唯一旧归档。历史验证不取得当前适用性。资格消费者、白名单与发布判定没有改动。

已验证：只读核对既有冻结绿回执 `20261029T020714.918159Z-ypaf148b`。命令 `go test -json ./ai/e2e -run '^TestChineseEvaluation' -count=1` 退出 0。98 个测试事件 PASS，audit findings 为 0。7 份增量文件与回执来源哈希相等。当前 Jev/Luna 默认与历史过程检查通过。

失败原件保留：`20261029T020415.894476Z-aggog5d9` 退出 1，含 `evaluation: task, model, dataset or price snapshot drift`；`20261029T020549.924379Z-ttw_tjh0` 退出 1，含三份 v2 配置缺失。两份回执均保留 `invalid case outcome`。

本次没有运行测试或厂商调用。此修复不产生新正式资格、付费权限、race/压力事实或完整发布通过。

Spec：0 项。


---

## Standards

已完成：工具快照场景的单文件静态审查。行动问题为 0。

范围：`3d29262c..7d4edddd` 的 `ai/e2e/tool_round_trip_test.go`。未运行测试、厂商请求或 Git 命令。

已验证：静态读取与原文件比较确认以下事实。

- `FramingLF` 每个 fixture 事件对应一个 chunk。`AfterChunk` 在首个参数 delta 写出并 flush 后阻塞，因此终态帧保持未发送。
- `paused` 确认服务端阻塞。测试同时检查 `Pending` 与原 `ToolCallNotExecutable` 断言，没有放宽公共合同。
- `sync.Once` 允许重复释放。defer 先释放服务端，再调用 `Close`。检查失败及超时路径同样执行该顺序。
- 释放后，测试等待终态，并确认同一个 PartialView 等于最终消息。终态工具调用必须通过校验。
- 修改只控制本场景的已存在公共 HTTP 测试边界。没有修改产品、共享 provider/helper 或新增生产职责。

## SOP 验证范围

拟定范围符合 SOP。没有必须再次完整运行 `./ai/e2e` 的具体遗漏。

最终命令中的 `TestToolRoundTrip$`、`TestToolCallTruncation$` 和 `TestStreamLifecycle$` 覆盖失败场景及相关快照合同。目录、xAI、相邻图片路线和中文评估按实际修改面重验。

静态测试清单确认：`./ai` 只有 `TestThinkingBudgets` 与 `TestCostEstimate`；正则包含二者。两个 chineseeval 包没有测试函数，命令仍编译它们。同次命令提供唯一主 E2E 包及辅助包，不依赖缺主包的额外 collector。

原完整采集继续保持 FAIL/exit1。其他 PASS 事实只按原始事件和原件引用。新索引只声明本次实际完整通过的选定集合；必须核对必需 case、来源、退出和 audit。不得合成新的全包 PASS。

未验证：本次最终冻结命令尚未执行。本审查不声明运行或发布通过。

Standards 共 0 项，无待修复问题。


---

## Spec

已完成：`3d29262c..7d4edddddfcc3edb8485b6ed889ba3e2c3287e9b` 的测试时序增量审查。可行动缺陷：0。范围只有 `ai/e2e/tool_round_trip_test.go`。

- 根因对应正确。`ai/partial.go:9–16` 明确 PartialView 是读时最新视图，可以超前于排队事件；终态后等于最终消息。旧测试把排队 delta 误当成当前消息仍 pending 的证明。`ai/tool_validate.go:91–103` 只拒绝尚未完成或非 ToolUse 消息，终态 ToolUse 返回可执行符合合同。
- `ai/e2e/tool_round_trip_test.go:119–144` 找出首个 arguments.delta，并用既有 AfterChunk 在该帧刷新后阻止后续帧。FramingLF 每事件一 chunk，因此终态未发送，pending 前提确定。
- `:151–161` 先等待上游暂停，再读取同一 PartialView。新增 Pending 断言后保留原 ToolCallNotExecutable 断言；没有删除或放宽原验收要求。
- `:164–175` 释放上游后，等待最终结果，核对同一视图达到 ToolUse、等于最终消息且工具可执行。覆盖的是视图按时间推进的完整合同。
- `:133–146` 使用 sync.Once 释放 channel。defer 始终先释放上游，再 Close；等待使用既有有界 helper。断言失败、FailNow 和正常结束均有释放路径。
- 产品代码、协议夹具、资格器与发布合同均无修改。本增量修复既有普通测试的时序假设，不扩大目录激活的业务范围。

已验证：只读核对失败回执 `20261029T022715.894874Z-82pbwikr`。命令 `go test -json ./ai/e2e -run '^TestToolRoundTrip/live-view-not-executable$' -count=1` 退出 1，关键输出为 `live snapshot refused: err=<nil>`。该确定性红证据与所述根因一致。原全包失败保持。

未验证：本次修复的运行结果。本次没有运行测试或厂商调用。最终冻结普通集合及结果由父任务登记，不能用静态复核替代。

Spec：0 项。

