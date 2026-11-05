# Claude Haiku 5.5 当前状态

2026-11-05 Git 收尾：本批改动纳入本轮提交。提交号及推送结果以 Git 记录为准。
以下保留 2026-11-04 测量完成时的信息；其中 HEAD 与未提交表述描述原时点。原封存交付包和状态快照保持原字节。

状态：完成（2026-11-04）。通过 Catalog CLI 添加并同步内置目录；尚未提交，未发布软件或部署。
当前及实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`。起点已含上一轮未提交的 Anthropic 下架变更，前置目录为 79 条；
本次比较对象是本批 before-source.json 与 before-builtin.json 固定的实际工作区，不把 HEAD 冒充起点全部字节。
原测量、交付包和 `.scratch/barness-ai-observation-reading/` 保留。

仅新增 chat / anthropic / anthropic-messages / claude-haiku-5-5；无日期别名。
目录 80 条：74 chat、3 image、3 classifier。Anthropic 8 条、内容版本 10 → 11。
原 79 条定义、价格及相对顺序不变，旧 9 条 Anthropic 身份保持下架。
文本/图片输入，context 1,000,000，max output 128,000。五个 effort：low/medium/high/xhigh/max。
本库选用已实现的始终 adaptive 托管 effort 路径，带 drop_block 前缀保护、逐轮 effort 和禁发 temperature。
不声明 PDF、Batch、原生浏览器/电脑工具或关闭推理。服务默认 medium 是 metadata；
ADR-0019 既有 SDK 托管策略在省略 effort 时仍为 high，显式 medium 的公共调用已验证。

USD/百万 token 的 input/output/cacheRead/cacheWrite 基础费率：0.1/0.5/0.01/0.125。
普通输入+缓存读+缓存写严格超过 100,000 时整请求费率：0.5/2.5/0.05/0.625。
等于门槛保持基础费率，1h 缓存写按该请求选定 input 的 2 倍计费。
原始 models.dev 响应含未映射 PDF；CLI dry-run 实际拒绝 unsupported_definition/exit1，未改写来源绕过。
改用官方原捕获、完整人工定义/逐字段决定，完成 add dry-run/confirm → check → apply dry-run/confirm → export review。
官方捕获、URL、抓取时间和原字节摘要见本批 inputs；可分发 review 含 CLI 实际采用的原捕获、定义、决定。
补充官方迁移/价格/effort 文档另归集在 verification/supplemental-sources，models.dev 原件留于外部索引所指原路径。

Catalog.Version：`catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402`。
Catalog.Hash：`sha256:d7ca061de668054669d4721038fa9cb29c94103815a8d2a318e9ee0025df9703`。
JSON 文件 SHA-256：`f29ff70c8b0a70469ee3b924c156e3c33b0bda079a8dbd1de91b25029d59b468`。
维护 revision：`sha256:8e46300c8052478e57982ed633b0682c6e2be4bd0ac9a5247653f9efae6e3c9e`。

候选检查实际 PASS/exit0，逐身份验证完整 80 条目录及必需场景。
新增 E2E 首次有未使用 import 和字符串类型比较两处编译错误，实际 FAIL/exit1；修正后发现缺型号的预期 RED/exit1，采用后 GREEN/exit0。
工作 launcher 复用已存在隔离 HOME 的首次启动在 CLI 前失败，实际 exit1；精确起止未被采集，明确保留为缺失，不编造耗时；改用独立隔离目录后操作成功。
旧失败、原件、启动缺口和重跑分别保留，不覆盖或追认通过。
首次完整普通的 A–M 段在旧 Kimi 收录测试失败：它把当前全目录等同旧批准快照；修正为只核对原批准身份投影，新增 Haiku 不再改变该历史证明。定向修复回归通过后，以新批次重跑完整普通，旧失败和另一段旧成功均保留。
专项 E2E 覆盖固定定义、全部 effort、图像、工具历史、授权、100k below/equal/above、缓存读/写触发换价、混合 5m/1h 写费、adaptive/binding wire 和原 79 条定义保持。

完整普通范围：go test -json -failfast ./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen -count=1 -timeout=40m。
使用 tools/verification/run.py 两段采集：^Test([N-Z]|LoopbackTransportDiagnostics$) 和 ^Test[A-M] -skip ^TestLoopbackTransportDiagnostics$。
两段实际 PASS/exit0，完整源码与显式输入集合相等；已发现 935 个顶层测试，实际选择并集精确相等。
墙钟跨度 895.81 秒，2186 个源码文件、97 个显式输入；固定必需场景 73 项 PASS，全部原生 audit 0。
跳过/NOT_RUN 保持未执行。实际 argv/退出、耗时、原生包、回执路径及摘要见 verification/ordinary-summary.json。
同源 go generate、build、vet、live 标签编译与 vet 均 exit0；仅编译/vet，没有厂商调用。

SDK 核心运行代码：NO；运行依赖：NO；schema：NO。完整运行源码与前置工作区的集合及 SHA 相等，包含外部 catalog loader。
本次 Go 变化仅相关 E2E 与示例默认路径。正式验证目标从前置工作区起未变；不把旧型号报告变成 Haiku 5.5 资格。
ChineseEvaluation v8 仅改配置文档版本和目录 pin，原 v1–v7 不改写；v7 仍按原快照加载并拒绝静默改绑新目录。
已变更的前置未提交 Go 源和目录说明原字节在本批 before-code/ 留存，并与上一轮回执摘要核对。

外部 JSON：本次受测程序已有显式加载能力，最终原字节快照通过兼容检查，可只分发 JSON 并创建新 Client；现有 Client 不热更新。
BuiltinCatalog 使用者须重建并分发程序；其他旧版本兼容性未验证。
live：NOT_APPLICABLE，仅目录代码兼容接纳，无新增厂商测量需求。race：NOT_APPLICABLE，无并发实现变化。
压力：NOT_APPLICABLE，仅固定一条记录增加，无资源策略或复杂度机制变化。软件发布/部署：NOT_RUN，不声明完整产品发布或新厂商资格。

[交付说明](../../artifacts/catalog/catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402/20261104T013101Z-18d7fe28/delivery.md) · [不可变完成索引](../../artifacts/catalog/catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402/20261104T013101Z-18d7fe28/delivery.json)。
完整 review 库存 37 个文件，与采用 baseline 逐文件相等。
测试胶囊、工具、models.dev 响应及其他原件依赖 external-originals.json 和候选检查引用的外部路径；不声明全部证据自包含。运行 JSON 可独立使用。
工作原件：`artifacts/catalog/work/20261104T013101Z-18d7fe28`。
回滚：withdraw 此身份，再 check/apply/export 并同步 Provider 源、原字节快照和完整 baseline；目录 pin 另建配置 revision，保留追加历史。本次未回滚。
