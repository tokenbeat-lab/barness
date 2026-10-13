# 13: 闭合十组合发布门禁、迁移文档与回滚证据

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 12: 取得 Luna 当前正式目录下的独立中文效果报告
**Source tickets:** 09
**Requirements:** R17, R19, R20
**Acceptance cases:** OD53, OD57, OD58
**Updated:** 2026-10-12

**What to build:** 发布命令实际消费十组合、P10 离线/固定差分、新图片容量以及 Jev/Luna 独立效果，缺失任一当前适用证据都拒绝；Host 能按文档撤销新授权，并获得可重放的最终 verdict。

## Responsibility

负责固定发布合同/证据选择器、所有 artifact 导入、两份独立效果参数和验证、实际门禁执行及公共文档/ADR 交付记录。运行时撤销和容量验证器由前置票提供；本票验证它们已纳入门禁。协作时保留他人改动。

## Failure scenarios first

- P10 仅写入 P0/追溯，未进入 Differential.Required；删除其 diff cases 仍能绿。
- 只读取旧 mixed-pressure，缺新分类图片 count/bytes/headroom 仍 PASS。
- 两个效果参数实际指同一个 bundle，错 route/旧 Hash/fixture/全拒答 null 指标绕过或使进程 panic。
- 同时篡改 expected/mapping/矩阵缩小必交能力，缺十组合报告或审计仍绿。
- 全球 schema 升级后门禁/trace 参数读旧形状，历史 PASS 冒充当前运行。
- 文档支持声明超过实装范围，回滚改走另一协议或抹去费用。

## Acceptance criteria

- [ ] 先写发布 CLI/artifact 的删项、错身份/Hash、缩集、同包复用、预算、缺审计和 null 指标反例，再修改固定合同。
- [ ] 固定十组合和旧要求；P10 同时进入批准 P0、差分 Required、场景选择器及 R01–R20 追溯；原六差分子集/三 Extension Route 保留。
- [ ] 删除全部 P10 差分或仅剩普通 P10 PASS 必失败；Decisions 是有 upstream 的第七差分子集，不可用 extension skip 放行。
- [ ] 新分类图片设计负载验证器实际被调用，count/decoded 总量/完整 JSON/占用率/内存/资源字段缺失或错误必失败，旧 Jev/九路线要求不删。
- [ ] Jev 与 Luna 效果分身份各自必交、各自校验 catalog/config/dataset/code/manifest/audit；同包两次、交换身份、fixture/NOT_RUN/FAIL、过期证据拒绝。
- [ ] 保持历史 Jev 报告的独立核验用途；当前十组合发布要求两份当前适用的真实效果，合法全拒答的完整报告能导入且质量为 0/null，如实呈现而不自设门槛。
- [ ] schema 3 十组合自身完整报告与正式快照一致；原九组合因 schema/Hash 不适用时实际重跑，不能只改旧 Hash/时间，不共享 key 或 PASS。
- [ ] 实际正常/race 两种运行均启用固定差分与压力，完成两种 vet；逐包核验所有 offline/race/live/evaluation manifests、来源与零发现审计。
- [ ] 导入前置授权撤销证据：禁 Binding、移除型号或图片权限阻止新调用，已 pin 调用保持快照/费用，无 fallback；缺撤销证据阻断交付。
- [ ] 文档覆盖公共 score 2–10、JSON 说明差异、三图片条件、RefusalAnswer、usage/cost、最终 Hook、错误、迁移、服务 alias 及实际回滚。
- [ ] ADR-0026 实装且证据成立后才 accepted；旧 ADR 的历史不覆写，GLOSSARY 的发布组合数和相关 docs 同步为十组合。
- [ ] 每票给出实际状态与 artifact，完整 release report 分开协议/live/容量/效果；缺 key 或旧未解失败保留并阻断最终完成。

## Completion

使用实际发布命令收集并判定，不自动部署或发布。外部条件不足时保存拒绝 verdict、本票未完成；只有完整当前证据成立才能写功能交付完成。
