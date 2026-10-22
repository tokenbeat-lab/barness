# 05 — Host 装配与价格来源

Type: task
Status: resolved
Blocked by: 01

## Context

- [Spec](../spec.md)
- [设计](../design.md)
- [研究](../research.md)
- [唯一当前状态](../spec.md#唯一当前状态)
- [测试 SOP](../../../docs/agents/verification.md)

## Scope

R02、R10、R12；K02、K11、K15。拥有 kimihost 示例、六路线官方目标/凭据隔离、留存价格来源及回滚说明。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：foundation implementer 在 `codex/kimi-host` 认领，基线集成 `e40b1a93aa2c85a12aefc16155491be8666edc94`。保留现有 HostConfig URL 校验，Kimi Purpose 独立；普通公共 Client/Host 装配 E2E，不启动真实厂商、race、压力。

## Answer

Host 与共享边界实现已完成，普通本地 scope 和 Chat/Messages 四入口已绿；Responses 四入口成本接线等待票03合并，故暂不关闭本票。实际 argv、数字退出、受测来源与原生 bundle/audit 指针见[票05回执](../evidence/ticket-05.md)。live/race/压力未执行，release.pass=false。

2026-10-22 最终普通验收：已合并三协议，完整六路线 Host 四入口、十二并发隔离、精确 hook 数值与共享成本均完成。冻结受测提交 `e8d42f90fc78e7ec6d1402cc2d8a73c467277696`；通过范围与 argv/exit/原件 SHA 见[票05完整收尾回执](../evidence/ticket-05.md#final-six-route-ordinary-closure)。普通实施验收完成，live/race/压力 NOT_RUN，不产生正式支持或 release.pass。
