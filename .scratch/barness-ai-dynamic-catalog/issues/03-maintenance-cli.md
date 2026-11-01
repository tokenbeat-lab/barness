# 确定性维护 CLI、来源映射和生命周期

Type: task
Status: ready-for-agent
Blocked by: 02

## Context

[Spec](../spec.md)、[设计](../design.md)、[当前状态](../status.md)、[验证 SOP](../../../docs/agents/verification.md)。

## Ownership and outcome

负责 cmd/catalogctl/、catalog/maintenance/ 的来源冻结、四层映射、候选存储、预览确认及新增/批量改价/下架/恢复/回退/版本历史。公共规则和 Provider 映射明确且保留缺失/有损/冲突/确认项；人工价格不被覆盖，缺价不补零。所有变更有只读 dry-run/help、JSON 与明确确认，原子保存且候选/历史不覆盖。实现 check/apply 的公共接口供 04 完成，不以自报 PASS 放行。

## Acceptance and verification

T02、T03、T04、T06、T07；真实 CLI 进程与固定来源、文件前后对照。
先在已确认公共边界建立失败场景，再实现。遵循 TDD 技能及仓库 SOP，失败原件保留。单文件不超过 500 行。不得执行未获单独授权的 race、压力或真实厂商请求。

## Comments

2026-11-01：从已确认 spec 建立实施依赖图。
