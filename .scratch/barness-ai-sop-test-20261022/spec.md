# 2026-10-22 SOP 普通回归

请求：按照测试 SOP 发起一次测试。

依据：[测试流程 SOP](../../docs/agents/verification.md)、[ADR-0051](../../docs/adr/0051-scoped-verification-receipts.md)、[测试流程重置状态](../barness-ai-test-process-review/status.md)。本次是当前 HEAD 的新普通回归，不修改实现，也不替换原功能的历史验收事实。

公共入口：`go test -json ./ai/e2e ./ai ./ai/internal/testkit/release ./ai/release/cmd/releasegate ./ai/internal/testkit/supportmatrix -count=1 -failfast -timeout=20m`，通过现有 `tools/run.py --stage e2e` 采集。

范围依据：最新变更是测试 SOP 启用及此前的回执、原生消费者修复；复用上一完整普通验收的五包范围和全部 8,636 份历史输入。执行文件集合与上一回执相同；按当前完整 SHA 新跑，不把旧结果充作新执行。预先冻结上一主 bundle 的 12,086 个 PASS case ID；核对本次实际发现集合、各辅助 bundle 库存、原件哈希及脱敏 audit。上一回执的 214 个 NOT_RUN 节点单独比对，不能写为完成。

本次不改厂商行为、型号、协议或支持资格，live 不适用；没有新的负载行为变更，压力不适用。本次未选择人工 race，记录 NOT_RUN；race 候选为相同五包加 `-race`，须交付实际来源、输入和用例后由人明确选择。压力候选仍按原功能状态维护，本轮不覆盖其未完成事实。

完成条件：真实命令退出 0、预先选定 case 全部 PASS、主原生包唯一、所有辅助原件和 audit 有效、来源及选定输入前后相等。失败即停止并保留 attempt，不自动修复或重跑。最终保存唯一当前状态、不可变索引和只读发布判定；缺人工 race 时 `release.pass=false`。

回滚：本次仅新增测试记录和独立证据，没有产品代码回滚；保留新旧失败和依赖，不执行证据清理。
