# 13: 闭合十组合发布门禁、迁移文档与回滚证据

**What to build:** 发布命令实际消费十组合、P10 离线/固定差分、新图片容量以及 Jev/Luna 独立效果，缺失任一当前适用证据都拒绝；Host 能按文档撤销新授权，并获得可重放的最终 verdict。

**Blocked by:** 12: 取得 Luna 当前正式目录下的独立中文效果报告

**Status:** claimed
**Type:** task
**Source tickets:** 09
**Requirements:** R17, R19, R20
**Acceptance cases:** OD53, OD57, OD58
**Updated:** 2026-10-13

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

- [x] 先写发布 CLI/artifact 的删项、错身份/Hash、缩集、同包复用、预算、缺审计和 null 指标反例，再修改固定合同。
- [x] 固定十组合和旧要求；P10 同时进入批准 P0、差分 Required、场景选择器及 R01–R20 追溯；原六差分子集/三 Extension Route 保留。
- [x] 删除全部 P10 差分或仅剩普通 P10 PASS 必失败；Decisions 是有 upstream 的第七差分子集，不可用 extension skip 放行。
- [x] 新分类图片设计负载验证器实际被调用，count/decoded 总量/完整 JSON/占用率/内存/资源字段缺失或错误必失败，旧 Jev/九路线要求不删。
- [x] Jev 与 Luna 效果分身份各自必交、各自校验 catalog/config/dataset/code/manifest/audit；同包两次、交换身份、fixture/NOT_RUN/FAIL、过期证据拒绝。
- [x] 保持历史 Jev 报告的独立核验用途；当前十组合发布要求两份当前适用的真实效果，合法全拒答的完整报告能导入且质量为 0/null，如实呈现而不自设门槛。
- [x] schema 3 十组合自身完整报告与正式快照一致；原九组合因 schema/Hash 不适用时实际重跑，不能只改旧 Hash/时间，不共享 key 或 PASS。
- [x] 实际正常/race 两种运行均启用固定差分与压力，完成两种 vet；逐包核验所有 offline/race/live/evaluation manifests、来源与零发现审计。
- [x] 导入前置授权撤销证据：禁 Binding、移除型号或图片权限阻止新调用，已 pin 调用保持快照/费用，无 fallback；缺撤销证据阻断交付。
- [x] 文档覆盖公共 score 2–10、JSON 说明差异、三图片条件、RefusalAnswer、usage/cost、最终 Hook、错误、迁移、服务 alias 及实际回滚。
- [x] ADR-0026 实装且证据成立后才 accepted；旧 ADR 的历史不覆写，GLOSSARY 的发布组合数和相关 docs 同步为十组合。
- [ ] 每票给出实际状态与 artifact，完整 release report 分开协议/live/容量/效果；缺 key 或旧未解失败保留并阻断最终完成。

## Completion

使用实际发布命令收集并判定，不自动部署或发布。外部条件不足时保存拒绝 verdict、本票未完成；只有完整当前证据成立才能写功能交付完成。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 09

2026-10-12：计划 ready-for-agent 不等于已发布；本任务只写设计与实施工单。


### 实施与当前阻断（2026-10-13）

实现提交 `ee3bf76`、审查修复 `807ccba`：固定十组合、P10 第七差分及原三 Extension Route、
R01–R20、normal/race 各自完整容量及撤销、Jev/Luna 独立效果，全部通过实际 CLI 导入。
发布边界先 RED 后 GREEN；审查发现的空 JSON、live inventory 删项和诊断问题已修复，
两轴复审 Standards 0 / Spec 0。用户已有 GLOSSARY 定义与规划文件保留。

最终自动门禁实测退出 0：normal/race 各 7,616 PASS（普通离线 6,755，固定差分 861，
pending 0），两种 vet 通过，109 追溯 PASS，14 包脱敏审计零发现。正式当前目录下十组合
各自真实 smoke 已重跑；Jev 新包和 Luna 原包均独立 COMPLETE，24/24，24 Call / Attempt、
零 retry。原始 Hash、时间、捕获及费用保留。首次迁移测试错误导致的 normal/race FAIL
与修复后的完整成功分别归档。

**本票仍 claimed，未完成交付。**
[旧压力流完整性缺陷](../../barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)
仍未确认根因与修复。额外执行原票 10 轮复现全部 PASS（70.997 秒），但不足以关闭缺陷。
因此最终 delivery verdict 为 FAIL，ADR-0026 保持 proposed；不能据自动门禁通过宣布可发布。
最后一项验收保持未勾选，待原缺陷有确认的修复证据、重新完成最终验收后再解决本票。

完整原包、逐票状态、实际命令、失败历史、审查、Hash 及无厂商调用重放见
[发布证据](../release-evidence/README.md)、[自动报告](../release-evidence/release-report.md)
和[最终交付 verdict](../release-evidence/delivery-verdict.json)。
