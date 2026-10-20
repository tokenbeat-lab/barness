# 13 — 按区域固定路由的缺口验证

Type: task
Status: resolved
Updated: 2026-10-18

## 授权与范围

用户本轮明确要求“继续完成缺口部分的验证；其中，中国区走直连，国际站通过本地代理（127.0.0.1:7890）进行访问”。据此建立新的有限验证合同，旧票10及票12的已关闭进程、消费和失败保持不可变。本轮不沿用旧剩余额度，不自动再开下一轮。

中国区 GLM、DeepSeek 直连；OpenAI、Anthropic、Google、TypeSafe 国际服务只通过 `http://127.0.0.1:7890`，不更换服务端点，不跨区回退，不继承环境代理。M3 国际站三份正式 PASS 仍适用则不重跑；M3.1 缺少 M Plan 权益仍为线上 NOT_RUN。

本轮固定18个完整进程：GLM三路各22场景，以及其余15项身份/效果（包括需对当前来源重新证明的两项 DeepSeek）。沿用既有单进程预算，总上限374 calls / 398 HTTP Attempts，其中 GLM72/96、其他302/302；零自动进程重跑，逐进程失败保留。旧消费300/300另列，总保守上限674/698。此数值是本次请求的执行上限，不声称用户逐字确认过该数值。原GLM专项的更早历史亦独立保留。

## 验收

- 在既有公开 launcher / CLI / Client E2E 边界失败先行验证计划固定、路由隔离、来源及 binary 变化拒绝、缺审查拒绝、单次 intent 跨目录不可重用、失败预算不重置。不得改写旧APPROVED记录或旧selector。
- 复用现有显式 proxy 参数，优先只改辅助工具，避免改变 M3 / GLM 选定执行来源。真实执行前冻结 committed source、binaries、plan，并经独立源码审查及无Key检查。
- 先做无凭据 DNS/TCP/CONNECT/TLS 诊断，保留有限收据；该检查不证明账户资格或模型支持。
- 三路 GLM 自身同来源完整 PASS 才能发布新 raw / known-usage / support views；任何 FAIL 只如实保留，不拼接。
- Jev支持包、17身份及四效果仅在实际来源/断言满足时更新。最终公共发布 CLI 真正执行新全量 ordinary/race、双 vet、冻结差分/压力并持久化退出码；已知阻断不得伪报通过。
- 保留可独立复核的有限归档/Hash/运行收据，更新交付状态和两轴审查。实现按 implement-spec 使用独立工作树，合并后清理自建工作树。

## Comments

- 本票是用户新增路由指令后的前瞻合同；票12的“无自动R2”仍约束旧合同。本票仅固定一次新路由验证，不给予无限重试或后续轮次。


2026-10-18实际完成：固定18进程已全部关闭，17PASS/1FAIL，新增245calls/245Attempts，连旧300/300累计545/545。其余15项collection均PASS（180/180），GLM Chat/Messages PASS、Responses19PASS/3FAIL（三路65/65）；严格三路续期拒绝，旧四selector不变。17身份默认矩阵原子导入16PASS/1FAIL，Jev自身支持包完成。8fbd170完整CLI数字退出0/0/1/1，ordinary/race各12439PASS/8FAIL/1NOT_RUN/audit0，报告pass=false。原中断与全部实际失败保留。完整门禁新暴露的离线缺口转[票14](14-release-preflight-gaps.md)；本票归档与两轴复核仍在收敛，不授权自动下一轮。当前结果见[交付状态](../delivery-status.md)。

## Answer

本轮固定实际验证、原始失败、全量false报告和全部有界留存均已完成两轴审查，并于 `7231fdad` 合入。按有限执行和如实报告范围关闭；17PASS/1FAIL及545/545不改写，spec、10、09仍开放。[证据入口](../evidence/ticket-13/README.md)保留各自测量来源。后续离线修复由14/16完成，新同源真实重验与新完整门禁由[15](15-source-scoped-revalidation.md)承接，仍待独立预算；当前自建工作树继续用于15，完成该工作后统一清理。
