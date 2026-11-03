# OpenAI chat 目录下架当前状态

状态：完成（2026-11-03）。已同步内置目录，未发布软件或部署。
起点/当前 HEAD/实际受测 HEAD：`b041eff5ca74393a88f66916b13ce1aaeec12ee8`；本次数据、测试和示例变更尚未提交。
实际受测工作区源码及全部输入由普通回执逐文件固定，不把 HEAD 当作未提交字节的替代。
已有未跟踪 `.scratch/barness-ai-observation-reading/` 保留。

保留 8 个 OpenAI chat 型号，Responses 8 条、Chat Completions 7 条。
其余 20 个型号、37 条接口身份均逐条经过 CLI dry-run/confirm → check → apply
预览/确认。维护记录与来源保留，OpenAI 内容版本 2 → 39；其他 Provider 不变。
总目录 88 条：82 chat、3 image、3 classifier。全部保留定义、价格及顺序不变。

Catalog.Version：`catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2`。
Catalog.Hash：`sha256:9dff935e34e42b96adc8afeb34c2db689447c936f1c584928ea08b3ed0a7b65a`。
JSON 文件 SHA-256：`e706fdc1dd13aed8799478da342467a810005e769420abce1ed0493dd8c0d9e6`。
维护 revision：`sha256:f5bf85ac328703c01a4019bebec41c5b779b1601b2f02eee0ae533881ed99a3e`。

37 个候选检查实际 PASS/exit0。最终检查逐身份覆盖完整 88 条目录；
每次原生必需场景、来源、退出和原件摘要见交付包 verification/candidate-checks.json。
同步前新增公开下架 E2E 为预期 RED。首次普通和后续定向回归各 FAIL/exit1：
我迁移历史测试时，将两个固定 gpt-4 单模型夹具误换成另一型号，导致 withdraw
返回 unchanged；已恢复它们的固定目标，失败原件保持。没有修改产品校验。另一次普通发现评估配置已更新但测试 Client 仍装配历史目录，
因此被原有漂移检查拒绝；已显式选择当前目录。后续定向回归发现本地示例未选
历史 Catalog，以及正式验证默认型号已下架；均修订其实际装配/验证目标。
另补齐普通诊断/资源夹具的显式历史 Catalog；为避免执行中修改源码，
主动取消过一次互补普通尝试，实际 CANCELLED/exit -15，原件保留。
全部失败、取消和重跑结果完整保存在 attempts.json，不以部分通过冒充完整普通。
协议回归显式装配原批准历史 Catalog；当前默认收缩通过独立公共 Client 测试。
维护生命周期测试使用当前有效型号，历史改价测试仅在私有临时库显式 restore。
正式 OpenAI 验证目标改为 GPT-5.6 Luna、GPT-5.6 Sol 和 GPT-6 Sol，quiet 使用
已声明支持的 off；旧 pro 重新加入目录并请求厂商的探针已删除。旧报告和 pin
保持原字节，旧型号测量不自动成为新目标的真实资格。

完整受影响普通范围：`go test -json -failfast ./ai/e2e ./ai
./ai/examples/chineseeval/... ./catalog/... ./cmd/catalogctl ./cmd/cataloggen
-count=1 -timeout=40m`，由 tools/verification/run.py 分两段采集。
第一段 `-run '^Test([N-Z]|LoopbackTransportDiagnostics$)'`；
第二段 `-run '^Test[A-M]' -skip '^TestLoopbackTransportDiagnostics$'`。两段独立执行，源码集合完全相等，发现的 933 个
普通顶层测试与两段实际选择集合精确相等；每份回执分别含唯一主原生包。
实际 PASS/exit0，墙钟 1109.84 秒；
2176 个源码、83 个显式输入。
各段唯一主原生包及辅助包各自统计、audit、42 个固定必需场景均见普通摘要。
跳过/NOT_RUN 场景保持未执行，不计通过。
最后一段回执：`.evidence/catalog-openai-retirement/e2e/20261103T124519.494506Z-_xbcm5a7/receipt.json`，SHA-256 `dc7eb1d11088521da222319c88c7b9eedb2817246dae329ccb23368160738fc1`。
两段原件、真实 argv/退出与各自耗时均在普通摘要 parts 和交付中的两份 receipt。
生成一致性、受影响构建和 vet 在最终同源工作区均实际 exit0。

核心运行代码：NO；运行依赖：NO；模型 schema：NO。
变更为维护源/快照/完整审阅基线、E2E 装配与断言、示例默认配置 v6、文档，
以及 supportmatrix/process_scope.go、generic_contract.go、generic_qualification.go
和 live/combos_test.go、catalog_test.go、final_budget_process_test.go 的验证目标。
ChineseEvaluation v6 仅改变配置版本和目录 pin；v1–v5 原字节保留。
受测 b041eff 程序支持外部 JSON 加载，且最终快照实际通过代码兼容检查：
可仅分发版本化 JSON，由 Host 显式加载并创建新 Client。现有 Client 不热更新。
使用 BuiltinCatalog 的程序需重建并分发。其他旧版本兼容性未验证。

live：NOT_APPLICABLE，无新厂商事实或资格需求。
race：NOT_APPLICABLE，无共享状态、锁、取消路径变化。
压力：NOT_APPLICABLE，目录缩小，复杂度、资源和设计负载未变。
软件发布/部署：NOT_RUN，本次未请求；不声明完整产品发布或新厂商资格。

完整交付与不可变完成索引：
[catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2](../../artifacts/catalog/catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2/20261103T114943Z-b12de1f9/delivery.md)，
[delivery.json](../../artifacts/catalog/catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2/20261103T114943Z-b12de1f9/delivery.json)。
完整审阅包库存 34 个文件，原字节与采用基线相等。
回放证据依赖索引所指外部原件；运行 JSON 可独立分发。

回滚：对所选历史 revision 执行 OpenAI Provider rollback 候选，或逐身份 restore，
重新 check/apply、导出并同步整组 Provider/快照/完整基线。目录身份变化时示例配置
另建 revision；保留追加历史和旧快照。本次未执行回滚。
