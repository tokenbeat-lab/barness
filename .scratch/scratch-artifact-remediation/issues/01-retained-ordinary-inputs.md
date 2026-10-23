# 01：普通测试固定输入随 Git 交付

Status: resolved
Type: task
Updated: 2026-10-24

依据：审查 F01 与用户明确授权第1项。目标、验证和边界见[spec](../spec.md)、[plan](../plan.md)，实际状态只见[status](../status.md)。

验收：独立干净工作区先 RED 后 GREEN；受影响普通包真实退出0，实际必交场景完整、audit通过，输入均跟踪/固定哈希，未调用厂商或继承负载开关。历史 zip/manifest/脚本原件不改。

## Comments

2026-10-24：开始依赖盘点及干净工作区 RED；未实施其他治理阶段。

## Answer

2026-10-24：原三份缺输入与追加GLM固定归档缺口均按独立RED/GREEN修复，冻结来源在干净工作区完整五包PASS；1909来源、8757历史输入、58必交case及所有原生audit已核对。实际14,071 PASS/214 NOT_RUN，退出0，446.40秒。完整结果与不可变原件只见[唯一当前状态](../status.md)。人工race未选，完整发布false；其他治理阶段未开始。
