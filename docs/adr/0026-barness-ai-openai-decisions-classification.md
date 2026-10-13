---
status: proposed
date: 2026-10-12
---

# Decisions 与 System One 共用 Classify Interface，协议与逐题拒答独立转换

Host 需要新增 GPT-6 Luna 的 Decisions 分类和图片能力，同时保持 System One/Jev 的已有行为。方案保留公共 Classify 作为主要 Seam，复用固定 Binding/Credential/Catalog、准入、有限重试、取消和观测；两个 Adapter 独立拥有 native 请求、回调后 guard、最终问题快照、答案和用量，不把新的协议条件泄漏成 Jev 的全局变化。

## 决策

- 公共请求新增 inline Images；图片同时由 classifier 图片资源子策略、型号能力和 Binding.AllowClassifierImages 显式开启。三个条件缺一不授予图片访问，旧配置默认 text-only/无图片授权，GenerateImages 策略不启用分类图片。
- 保留现有 State/Questions 和旧 Jev 的宽 JSON 说明、单候选、原始 wire、错误优先级与有界评分舍入。Decisions 的字符串说明/两候选下限由自己 capability 检查；本期公共 score 仍 2–10，是 barness 支持范围而不是 OpenAI 上限，以保留 scope 拒绝顺序。
- 合法逐题 refusal 新增 RefusalAnswer，混合和全拒答仍 stop/nil error；协议损坏仍整体清空 Answers，并保留唯一已报告 usage。Jev 不接受未知 native refusal。Host 的业务规则与拒答路由不进入协议 Module。
- 新增字段省略默认是公共契约，不改写旧 Jev 目录内容或收紧旧元数据已可构造的范围；图片声明与权限不得由新目录或 hook 自动扩大。
- 用现有 unary native HTTP，不先升级全项目 SDK。每协议独立解析 usage presence；Decisions input-only 费率和固定 272000 全请求阶梯不复用聊天缓存拆桶。score 分布保留，TypeSafe 的舍入例外不共享。
- 保持单一 Pi 1.1.0 oracle；refusal、score 详情、严格校验、JSON 精度、用量和重试等差异有限登记。候选自身 live 后才纳入内置目录，正式 Hash 下再形成第十组合证据；旧九组合和 Jev 效果结论不被复用为新路线 PASS。

## 取舍与影响

选择逐题拒答而不复制 Pi 的整次 error，会新增公共答案 variant；Host 切换到新路线时须处理它，nil error 不等于每题有普通答案。选择保留旧公共 score 上限会暂不表达 Decisions 更宽的潜在原生范围，但避免为未出现的真实用例改变旧错误行为。未来若扩大范围，须评估公共校验顺序，而不是内部补偿 shim。

三个图片条件使 Host 在资源、型号与租户权限上分别明确承担责任，代价是多一个 Binding 开关；它避免已有 classifier 白名单仅因目录升级获得新输入权限。目录元数据和价格固定不能保证 gpt-6-luna 无日期 alias 的权重固定，独立效果证据记录日期和实际 ResponseModel。

该 proposed ADR 对 ADR-0021 的“分类目前单协议”以及其 TypeSafe wire 假设提出扩展，对 ADR-0020 的模型能力/Binding 图片权限提出新增；不取代 Jev 的既有结果和历史验证。ADR-0003/0006/0010/0016/0017/0018/0024 的隔离、重试、价格、证据与 Host 责任继续适用。

完整契约与阶段门槛见[spec](../../.scratch/barness-ai-openai-decisions/spec.md)、[design](../../.scratch/barness-ai-openai-decisions/design.md)、[计划](../../.scratch/barness-ai-openai-decisions/implementation-plan.md)。工单 01–12 已完成公共实现、固定差分、正式目录及独立真实证据；最终验收状态见下方交付记录。


## 实施与交付记录（2026-10-13）

工单 13 的十组合合同、P10 第七差分、双路线效果导入、normal/race 容量与撤销校验已实现。
本轮在正式目录 `2026-10-13.1` 下重新取得十条路线各自的 schema 3 smoke；Jev 新效果
与工单 12 原始 Luna 包独立核验，未替换旧 Hash、时间或业务结论。公共迁移与回滚说明见
[Decisions 契约](../barness-ai/decisions.md)，命令、审查、全部原包与实际 verdict 见
[工单 13 证据](../../.scratch/barness-ai-openai-decisions/release-evidence/README.md)。

本 ADR **保持 proposed**。[旧压力流完整性缺陷](../../.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)
仍未确认根因及修复，本轮检查或复现通过不能关闭它。最终交付因此为 FAIL，工单 13 保持
claimed；只有该阻断以原负载和回归证据解决、最终完整发布验收成立后，才能改为 accepted。
既有 ADR 的历史记录和原始失败包保持不变。
