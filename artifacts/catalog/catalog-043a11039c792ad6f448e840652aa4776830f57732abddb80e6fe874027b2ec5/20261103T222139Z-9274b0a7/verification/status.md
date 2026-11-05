# Anthropic chat 下架当前状态

状态：完成（2026-11-04）。目录已同步内置；软件发布/部署未执行。
起点、当前 HEAD 和实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`。本次改动尚未提交；实际工作区源码和输入由普通回执逐文件固定。
已有未跟踪 `.scratch/barness-ai-observation-reading/` 保留。

本次仅保留 Anthropic 四个系列中 4.8 及以上的现有 7 条身份：
Fable 5、5.1；Opus 4.8、5、5.5；Sonnet 5、5.5。Haiku 没有符合型号。
其余 9 条逐一完成 withdraw dry-run/confirm → check → apply dry-run/confirm。
Anthropic 内容版本 1 → 10，其他 Provider 版本不变。总目录 79 条：73 chat、3 image、3 classifier。
所有保留定义、价格、相对顺序和其他 Provider 原字节保持；下架定义与追溯保留。

Catalog.Version：`catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5`。
Catalog.Hash：`sha256:40d4c5fe7551fb06e79210b676e69e96a47b6b1b200dde5f6c07e243f970b2f0`。
JSON 文件 SHA-256：`44d7caa8f36ce1b5b7914d014215be1e3625b453677cabe54d891b59ac5a9bb8`。
维护 revision：`sha256:2047fad03a8020c866481cb8549e9e0a28f4a0a767646e65f4e13b41a381349f`。

9 次候选检查实际 PASS/exit0；最终候选逐身份验证全部 79 条定义的查询、请求/响应、用量、计价和适用能力。
公开 Client 下架场景先出现预期 RED/exit1，采用后 GREEN/exit0：目录不再发现旧型号，显式允许列表也不能使其发出 HTTP。
全部失败/检查/重跑独立保留于 verification/attempts.json；没有其他修复性失败。

完整普通范围：go test -json -failfast ./ai/e2e ./ai ./ai/examples/chineseeval/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen -count=1 -timeout=40m。
由 tools/verification/run.py 分两段采集，实际表达式分别为 ^Test([N-Z]|LoopbackTransportDiagnostics$) 与 ^Test[A-M] -skip ^TestLoopbackTransportDiagnostics$。
两段实际 PASS/exit0，执行源码和全部显式输入集合相等。已发现 934 个顶层测试，实际选择并集精确相等。
墙钟跨度 884.20 秒；2181 个源码文件、82 个显式输入。固定必需场景 54 项全部 PASS，全部原生 audit 0。
跳过/NOT_RUN 保持未执行。各段真实 argv、数字退出、耗时、原生包及回执 SHA 见 verification/ordinary-summary.json。
最终同源 go generate、build、vet、live 标签编译及 vet 均 exit0；后两项没有启动厂商请求。

核心运行代码：NO；运行依赖：NO；schema：NO。完整基准文件集合及 SHA-256 与当前运行代码相等，包括外部 catalog loader。
变化仅为维护源、原字节运行 JSON、完整基线、相关 E2E、示例配置 v7、验证工具型号和文档。
正式验证默认 Haiku 4.5 调整为 Sonnet 5.5；历史报告与测量不改写，不授予新型号真实资格。
ChineseEvaluation v7 仅改配置版本与目录 pin；旧 v1–v6 保留原字节，v6 能按原快照读取并拒绝静默绑定新目录。

受测程序已有外部 JSON 加载能力，最终原 JSON 在本次核心代码上通过兼容检查：可仅分发 JSON，Host 显式加载后创建新 Client；现有 Client 不热更新。
BuiltinCatalog 使用者必须重建并分发程序才能更新默认。其他旧程序版本兼容性未验证。
live：NOT_APPLICABLE，无新厂商事实需求。race：NOT_APPLICABLE，无共享状态/锁/取消路径变化。
压力：NOT_APPLICABLE，目录收缩，无负载或资源策略变化。软件发布/部署：NOT_RUN，本次未请求，不声明完整产品发布。

[交付说明](../../artifacts/catalog/catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5/20261103T222139Z-9274b0a7/delivery.md) · [不可变完成索引](../../artifacts/catalog/catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5/20261103T222139Z-9274b0a7/delivery.json)。
完整 review 库存 34 个文件，已与采用基线逐文件核对。
运行 JSON 可独立使用；证据重放依赖 external-originals.json 和候选检查索引所指外部原件，不声明整包证据自包含。
工作原件：`artifacts/catalog/work/20261103T222139Z-9274b0a7`。
回滚：对 Anthropic 历史 revision 创建 rollback 候选，或逐身份 restore，重新 check/apply、导出及采用整组来源/快照/基线；示例 pin 另建 revision。
本次未执行回滚；保留追加历史和旧快照。
