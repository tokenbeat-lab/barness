# 01 — 身份、候选目录与成本范围

Type: task
Status: resolved

## Context

- [Spec](../spec.md)
- [设计](../design.md)
- [研究](../research.md)
- [唯一当前状态](../spec.md#唯一当前状态)
- [测试 SOP](../../../docs/agents/verification.md)

## Scope

R01–R03、R09–R10；K01、K02、K10–K11、K14–K15。拥有 Provider 常量、Kimi 候选目录/验证、Attempt 成本值、共享成本解析与调用快照接线。

## Acceptance

在已确认的公共 Client E2E seam 先留存失败、再实现并转绿；使用 SOP run.py 保留命令、退出与原生 bundle。文件低于 500 行；保留其他 Provider 合同。完成时记录提交、回执指针及未执行项。

## Comments

2026-10-22：从已批准实施的 spec/design 派生；任务依赖是集成顺序，独立协议可在同一基线并行开发，合并前同步集成分支。

2026-10-22：由 foundation implementer 在 `codex/kimi-foundation` 认领；基线 `8678bc951ddaed20b0e5d16314ef791ed19ed156`。公共 Client E2E seam 及验证范围沿 spec；不启动 live/race/压力。

## Answer

已交付独立 Provider 身份、K3 元数据上界、六份独立候选记录、官方完整原文/双哈希、原币与历史 FX 估算记录、值类型 CostEstimate、严格共享计数与实际 TTL 成本计算、Attempt/Observer 同步快照及安全错误分类。候选可用性不激活正式目录。

实施提交：`d8a8fac`、`5a488dc` 及后续本票收尾提交。仅普通公共 Client E2E：目录红例 `20261022T140509.860514Z-f52_2h3h` / 追加约束红例 `20261022T140937.772146Z-lz_566wf`；最新目录与候选快照绿例 `20261022T141515.819793Z-rcxbsaf_`（exit 0）。Messages 模型采样红例 `20261022T141417.307795Z-h8mnorju`，同一最新绿例已覆盖。共享成本边界红例 `20261022T141305.559353Z-u8nehw8b`（exit 1），待协议接线合并后转绿；本票不提前声明成本最终验收。

以上原件保留在主工作区 `.evidence/barness-ai-kimi-foundation/e2e/<attempt>/receipt.json`，每份均自含 argv、数字退出、来源文件与原生 bundle/audit 哈希。Chat usage/audit 与 Messages hook panic 的独立红例分别由票02/04保留；最终相关协议普通包由集成分支统一验收。真实 live、race、压力均 NOT_RUN；无正式晋升与发布资格，`release.pass=false`。回滚为停用 Kimi Binding/移除候选并回退本票提交，不重写历史。

2026-10-22 最终普通验收：已合并三协议，完整六路线 Host 四入口、十二并发隔离、精确 hook 数值与共享成本均完成。冻结受测提交 `e8d42f90fc78e7ec6d1402cc2d8a73c467277696`；通过范围与 argv/exit/原件 SHA 见[票05完整收尾回执](../evidence/ticket-05.md#final-six-route-ordinary-closure)。普通实施验收完成，live/race/压力 NOT_RUN，不产生正式支持或 release.pass。
