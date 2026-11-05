# Flash 1017 当前状态

2026-11-05 Git 收尾：本批改动纳入本轮提交。提交号及推送结果以 Git 记录为准。
以下保留 2026-11-04 测量完成时的信息；其中 HEAD 与未提交表述描述原时点。原封存交付包和状态快照保持原字节。

状态：完成（2026-11-04）。通过 Catalog CLI 采用 DeepSeek V4.1 Flash 1017；尚未提交 Git，未发布软件或部署。
当前及实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`。起点包括此前未提交的 Anthropic 下架、Haiku 与 Pro 接入。
本次实际基准由 before-source.json、before-builtin.json、before-code/ 固定；不把 HEAD 冒充全部起点字节。
原交付、原测量、文档澄清记录和 `.scratch/barness-ai-observation-reading/` 均保留。

官方 2026-10-17 更新明确调用 ID 仍为 deepseek-flash，调用方式与价格不变。
因此按 Responses 优先原则依次 revise 既有 openai-responses 与 openai-completions 身份。
两处 name 由 DeepSeek V4.1 Flash 改为 DeepSeek V4.1 Flash 1017，并补充官方原捕获与 /name 人工决定。
不创建虚构日期 ID，不增加重复条目，不恢复退休别名，不改服务绑定选择或授权。
原 ID、能力、价格与相对顺序保持，模型目录仍为 81 条（75 chat、3 image、3 classifier）；DeepSeek 4 条，内容版本 2 → 3 → 4。
文本/图片、1M context、384k output、已支持 low/high/max 保持；其他 effort 的既有 null/缺省语义不扩展。
峰值 USD/百万 token input/output/cacheRead/cacheWrite 为 0.3/1.2/0.006/0，作为固定保守估算；低谷价为一半，当前不按日历自动换价。
展示 1017 不表示永久固定 checkpoint；deepseek-flash 随官方升级滚动变化。

完成两次独立 revise dry-run/confirm → check → apply dry-run/confirm → export review，Responses 在前、Chat 在后。
官方捕获真实时间/原字节摘要及手工字段决定见本批 inputs。补充官方原件不会改写原 serving source association。
第一次过渡交付未作为编译内置默认采用；最终三处 Provider 源、原字节内置快照与完整 baseline 同步。

Catalog.Version：`catalog-34de22d2e31abb82108b6dde3dccb5def54fbb825e2f7281b862924b8f6860b4`。
Catalog.Hash：`sha256:fd23a12ac65a114116bcfbaf6f6ff335e3ab4cf5b2af255cc00d1cff49a85156`。
完整 JSON SHA-256：`672cb593ac0e3c6894bfbca8c753254162fdcba8c41adf8321a37108c6e31de4`。
维护 revision：`sha256:14525dac0d1c3505be8f3b330f7bdb25f9e29a609779618bcd9f0b9d61ac8ee1`。

两次候选完整检查均 PASS/exit0，各覆盖全部 81 条身份与必需场景。
新公开 Client 测试在旧名称上产生预期 RED/exit1；原测试文件按该失败回执的精确 hash 留存。
首次定向 GREEN 在 Kimi 历史名称比较处 FAIL/exit1；历史比较改为保留原展示名称、核对其余运行契约，修复后的完整定向范围 PASS/exit0。
一次准备未完成时的 launcher 调用因脚本尚不存在而 exit2，没有 Go 测试开始；精确起止/墙钟缺失，明确为未知，见 launcher-failure.json。
失败与成功回执、原件和工具输出分别保留，不覆盖失败。
新增专项覆盖两种 API 的查询、基础请求、缓存用量、工具历史、图像、参数、推理与授权；新名称独立核对，所有能力和价格与前置定义完全相同。
本批数据审计独立证明全目录恰好两处 name 变化；历史检查放开展示标签不放开其他字段或顺序。

完整普通范围：go test -json -failfast ./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen -count=1 -timeout=40m。
由 tools/verification/run.py 两段采集：^Test([N-Z]|LoopbackTransportDiagnostics$) 与 ^Test[A-M] -skip ^TestLoopbackTransportDiagnostics$。
两段均 PASS/exit0；完整源码及显式输入集合相等，实际选择并集与全部 937 个已发现顶层测试精确相等。
墙钟跨度 897.48 秒，2196 个源码文件、103 个显式输入，固定必需场景 100 项 PASS，所有原生 audit 为 0。
Go skip/native NOT_RUN 保持未执行。真实 argv/退出/时间/原生库存见 verification/ordinary-summary.json。
同源 generate/build/vet/live 标签编译及 vet 均 exit0；仅编译检查，没有厂商请求。

本次 SDK 核心运行代码 NO，运行依赖 NO，schema NO；核心完整文件集合与 SHA 相等，包含外部 catalog loader。
Go 改动仅相关 E2E 和示例默认配置路径，完整文件清单见 data-audit.json。一次性辅助脚本仅用于 CLI/验证/归档。
ChineseEvaluation 新建 v10，仅修改配置文档版本与目录 pin；v1–v9 原字节保持，v9 可按原快照加载并拒绝静默改绑。
本次未改正式 live 目标与原授权。旧 Flash 测量保留原时点和来源，不改成 1017 新实测资格。

外部 JSON：最终原受测字节在本次程序上通过兼容检查，Host 可以显式加载并创建新 Client；现有 Client 不热更新，其他旧版本兼容性未验证。
BuiltinCatalog 默认更新须重建并分发程序。live NOT_APPLICABLE：仅目录滚动名称采用，无新增厂商事实需求。
race NOT_APPLICABLE：无并发实现变化。压力 NOT_APPLICABLE：固定元数据修订，无资源/复杂度机制变化。
软件发布/部署 NOT_RUN；不声明完整产品发布或新厂商资格。

[最终交付](../../artifacts/catalog/catalog-34de22d2e31abb82108b6dde3dccb5def54fbb825e2f7281b862924b8f6860b4/20261104T040151Z-cd1addcc/delivery.md) · [不可变完成索引](../../artifacts/catalog/catalog-34de22d2e31abb82108b6dde3dccb5def54fbb825e2f7281b862924b8f6860b4/20261104T040151Z-cd1addcc/delivery.json)。
完整 review 库存 43 个文件，与采用 baseline 逐文件相等。
过渡交付及哈希见 verification/intermediate-delivery.json；它不替代最终内置采用。
测试胶囊/CLI/工具依赖 external-originals.json 与候选检查引用的外部原路径，不声明全部证据自包含。运行 JSON 可以独立使用。
工作原件：`artifacts/catalog/work/20261104T040151Z-cd1addcc`。
回滚：以新 revise 恢复原名称，再 check/apply/export/三处同步并另建配置 pin；保留追加历史，本次未回滚。
