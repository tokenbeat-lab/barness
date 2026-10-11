# 06：同步公共契约并交付完整 1.1.0 发布证据

Status: ready-for-agent
Type: task
Depends on: 01, 02, 03, 04, 05
Updated: 2026-10-11

## What to build

在实现与零 pending 差分齐备后，同步正式文档、基线合同、目录版本、追溯与发布选择器，完成完整当前版本发布包。该工单既负责使用说明，也负责明确专项完成与整版发布完成的区别。

## Ownership and context

负责 `docs/barness-ai/`、主 barness-ai spec、ADR-0004/0025 的基线生效说明、`ai/release/` 的 traceability/固定合同与正式快照、发布 evidence 索引。保留 ADR-0004 的历史 1.0.0 核验与证据，在新节记录 1.1.0，不改写旧通过记录。协作时不要回退他人修改。

依据：[规范 R09–R12](../spec.md)、[设计 §7](../design.md#7-发布失败处理与回滚)、[验收 B04](../acceptance.md#5-基线与发布验收)、[现行 ADR-0017](../../../docs/adr/0017-barness-ai-release-gate.md)。

## Failure scenarios first

- 用户示例缺必要 Host 装配，或将合成 temperature/top_k 数值称为厂商推荐。
- 文档声称“最终实际强度选参”，但未写 full wire、summary-only、off=null 或回调覆写边界。
- 主规范/ADR/发布校验版本不一致，正式目录/价格 hash 与当前代码不一致。
- 新采样需求没有 case/artifact 追溯；靠删除原要求或宽泛正则放松固定发布合同。
- 复用旧 Catalog hash 的 live PASS、只跑受影响聊天组合便宣称九组合整版通过。
- 缺凭据或业务效果证据却填 PASS，泄漏原始值，或回滚只回退 oracle 一部分。

## Acceptance criteria

- [x] 用户文档包含可执行 Host 示例、模型级字段与调用覆盖、合并顺序、七个 canonical key、full/simple 区别、支持/忽略协议及有效 JSON 边界。
- [x] 示例使用已支持、已授权模型或受控 Provider；数值标明来源或“合成演示”，不增型号与支持声明。
- [x] 正式契约、主 spec、ADR、oracle/provenance、目录/价格快照、ledger、traceability 和固定发布合同一致；只在核验完成后记录 1.1.0 生效。
- [x] 每条 R01–R12 与 S/B 场景有实际 artifacts；新增选择器和合同 hash 经审查，原有 P0/六差分子集/三扩展路线/压力/live/效果要求完整保留。
- [ ] B04：现行完整 releasegate 执行通过，全部九组合 live/evaluation 输入匹配当前正式目录、价格、SDK 和身份，离线与 race 的差分/压力、vet、审计齐备。
- [x] 记录整包回滚步骤：实现、目录、oracle lock/provenance、账本和合同恢复匹配的先前版本，Host 同步回退配置；无双写/shim。
- [x] 缺外部条件时逐项标 NOT_RUN 并保持发布未完成；只有实际通过的项目写 PASS，不由历史证据或需求映射填结果。

## Verification and closure

运行现有 releasegate，按实际 CLI 提供每条 live 与独立 evaluation 包；测试凭据只通过既有受控入口注入，不写入命令或文档。保留完整请求/响应/结果、artifact hashes、审计、历史失败与当前报告。

本 feature 的完成条件为 B01–B04 全通过、所有目标需求可追溯且无未完成验收。完成后更新 ADR-0025 的目标状态和各工单 Comments，记录执行命令、commit/版本、证据路径与结果；不得只因生产代码完成而关闭发布工单。

## Comments

2026-10-11：方案阶段创建，未执行目标 live 或完整发布门禁。

2026-10-11 实施：文档、ADR、目录/价格、traceability 与固定计划已同步，示例编译通过。当前真实中文 24 次均 HTTP 451，smoke 同样外部失败，完整发布未完成；保留 FAIL，未关闭工单或降低门槛。 命令、实际 artifacts、失败历史和未完成项见[专项证据](../evidence.md)与[来源核验](../release-verification/README.md)。

2026-10-11 当前完整候选：受测源码 dba212b；普通与 race 各 6,125 case 全 PASS，
两种 vet、固定计划、目录快照、全部差分零 pending 和原始审计通过。九组合实际
执行后 DeepSeek 两条路线 PASS，七组合 FAIL；独立中文效果 24 次 HTTP 451、
0 可评分，B04 未通过，继续保持工单开放。另保留[压力缺陷 07](07-pressure-stream-integrity.md)
的历史原失败与未确认根因，重跑通过不视为修复。完整可携证据与原状态重放见
[发布证据包](../release-evidence/README.md)。
