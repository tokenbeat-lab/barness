# DeepSeek Pro Responses 当前状态

状态：完成（2026-11-04）。修订 ADR-0014，优先通过 Catalog CLI 增加 Pro Responses 并同步内置目录。
尚未提交 Git，未发布软件或部署。
当前及实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`。起点含此前未提交的 Anthropic 下架及 Haiku 5.5 接入，前置目录 80 条。
本次实际比较基准由 before-source.json、before-builtin.json 与 before-code/ 原字节固定，不把 HEAD 冒充起点全部字节。
原交付、测量与 `.scratch/barness-ai-observation-reading/` 保留。

新增 chat / deepseek / openai-responses / deepseek-v4-pro，Responses 是 Pro 的推荐入口。
Host 显式选择 Binding.API，不增加自动协议选择或隐式 fallback。原 Pro Chat、Flash 双协议及其他定义保持独立。
目录 81 条：75 chat、3 image、3 classifier。DeepSeek 4 条、内容版本 1 → 2。
原 80 条定义、价格及相对顺序不变。文本输入，context 1,000,000，max output 384,000。
Responses effort 映射：off→none，minimal/low→low，medium/high/xhigh→high，max→max。
服务默认 high 是 metadata，不改变既有 SDK 的省略策略；推荐示例显式指定 high。
使用已有无状态 Responses 子集：文本、函数工具、推理和完整历史。本轮不扩大到 web_search/custom apply_patch。

价格为固定峰值保守估算，USD/百万 token 的 input/output/cacheRead/cacheWrite：1.32/3.96/0.044/0。
官方低谷价为一半；当前 ModelCost 不表达日历/时段，不声称上述估算等同任意时段真实账单。
官方 Responses、价格与推理原捕获及真实抓取时间/hash 均留存。CLI 采用官方 pricing 原件、完整人工定义和逐字段决定。
完成 add dry-run/confirm → check → apply dry-run/confirm → export review，再同步 Provider 源、原字节内置快照及完整 baseline。

Catalog.Version：`catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8`。
Catalog.Hash：`sha256:e6dae4b736a90c2dc7a6a4a0891c66a535af314be77eb780a7bb2090ef81473d`。
JSON 文件 SHA-256：`810061d335f93a2835516234dd9398d65243258e2c733026cca8c1381e9d90b8`。
维护 revision：`sha256:b61ef265a715462f70ad6609588e44c4fbbed9b636a95ed92114cb0abe536257`。

候选完整目录检查 PASS/exit0，全部 81 条逐身份及必需场景通过。
公共 E2E 先因缺 Pro Responses 得到预期 RED/exit1，采用后专项 GREEN/exit0；原失败原件保留。
专项覆盖定义/limits、七级 effort 映射、价格、工具/历史、文本输入/图片拒绝、授权、无状态字段隔离、ServiceTier 拒绝与无隐式 Chat fallback。
旧 Haiku 接入测试改为核对原批准身份的投影，后续独立添加不改变旧身份的历史证明。

完整普通范围：go test -json -failfast ./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen -count=1 -timeout=40m。
由 tools/verification/run.py 两段采集：^Test([N-Z]|LoopbackTransportDiagnostics$) 与 ^Test[A-M] -skip ^TestLoopbackTransportDiagnostics$。
两段实际 PASS/exit0，来源和显式输入集合相等；全部 936 个已发现顶层测试的实际选择并集精确相等。
墙钟跨度 902.75 秒，2191 个源码文件、98 个显式输入；固定必需场景 83 项 PASS，所有原生 audit 为 0。
Go skip/native NOT_RUN 保持未执行。实际 argv/退出/耗时/原生包见 verification/ordinary-summary.json；不把父测试通过当子场景通过。
同源 generate/build/vet/live 标签编译及 vet 均 exit0；仅编译检查，未运行厂商请求。

本次 SDK 核心运行代码：NO；运行依赖：NO；schema：NO。核心源码完整集合与 SHA 相等，包含外部 catalog loader。
Go 变化仅相关 E2E 和示例默认 pin。ChineseEvaluation 新建 v9，仅改配置版本/目录 pin；v1–v8 原字节保留，v8 仍按旧快照加载并拒绝静默改绑。
ADR 当前勘误和原文均留存；旧 Flash 探针/冒烟只证明原来测量的路线，不转为 Pro 的厂商实测资格。
正式 live 目标与原授权未变。

外部 JSON：本次受测程序已有显式加载能力，最终原字节快照通过兼容检查，可仅分发 JSON，由 Host 显式加载并创建新 Client；现有 Client 不热更新。
BuiltinCatalog 默认目录须重建并分发程序；其他旧版本兼容性未验证。
live：NOT_APPLICABLE，仅目录代码兼容接纳，无新增厂商测量需求。race：NOT_APPLICABLE，无并发实现变化。
压力：NOT_APPLICABLE，仅固定一条记录增加，无资源策略/复杂度机制变化。软件发布/部署：NOT_RUN，不声明完整产品发布或新厂商资格。

[交付说明](../../artifacts/catalog/catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8/20261104T030633Z-d3d7bbbb/delivery.md) · [不可变完成索引](../../artifacts/catalog/catalog-2cd1a14e6874a71e10e4c1e6fdd6327c6a102ca05a9428bdaa81c260ff6466b8/20261104T030633Z-d3d7bbbb/delivery.json)。
完整 review 库存 40 个文件，与采用 baseline 逐文件相等。
测试胶囊/CLI/工具仍依赖 external-originals.json 与候选检查所指外部原件；不声明全部证据自包含。运行 JSON 可独立使用。
工作原件：`artifacts/catalog/work/20261104T030633Z-d3d7bbbb`。
回滚：withdraw 新身份，再 check/apply/export 并同步 Provider 源、原字节快照及完整 baseline；目录 pin 另建配置 revision，保留追加历史。本次未回滚。
