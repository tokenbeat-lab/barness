# 04 自身证据与发布追加门禁

Status: resolved
Type: task
Blocked by: 01, 02, 03

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R02 R16 R17 R18 R20。

## Ownership

现有 live、支持矩阵、差分扩展、release 消费者追加新路线/型号与身份/来源校验，保留全部旧门禁和原件；CLI/消费者失败场景优先。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。


## Answer

2026-10-20：实现完成，生产字节冻结于 `af03f53ffd0025b1c01614e3fb78d5f722d4371e`（`codex/regional-evidence`，已融合 01–03、06 与 Host 的过期来源拒绝修订）。追加六条自身支持路线，保留原 17 条与原八项差分，发布合同现为 23 条路线、十项真实冻结差分及四条精确新扩展；追加 `REGIONAL-R01..R20` / `REGIONAL-RP01..RP20` 的独立有限 case ID 集与不可缩减映射 Hash。

新增独立候选/正式 live 入口与 keyless 原生消费者。grant 绑定 mode/stage、真实 Provider、standard 用途、账户/地区别名、型号、端点、当前源码、目录、SDK、有效期与固定预算，一次性 `.used` 在发送前独占创建。执行记录完整 source/build/binary SHA、SDK wire 请求/流、Result、Observer、工具往返、原生 reasoning、事件和资源；消费器重新核对完整有限用例及库存、实际型号/端点/归属、预算与用量、终态、工具配对和原始授权回执。源码过期、兄弟站点/型号、候选、受控 replay、缺失或篡改材料不能认证正式支持。首次受控 producer 的 source 顺序、取消 fixture、编码预算及 ID 解码失误的真实失败均保留，未刷新历史原件或 pin。

后续人工资格验证入口、grant 字段与 `-ldflags` 源码注入方式见 `ai/live/REGIONAL.md`。keyless 命令：`go test ./ai/e2e -run '^TestRegionalProviderRelease'`；`go run ./ai/live/cmd/regionalverify -source`；`go run ./ai/live/cmd/regionalverify -grant-template <combo>`。真实/正式运行必须按该文档另行获得自身明确授权；本票没有读取实际 key、访问厂商或运行 race。

证据索引：根工作区 `.evidence/provider-sites/release/index.json`。最终源码冻结后的三条实际 PASS 回执分别为 `27-integrated-release/receipt.json`（新 Release E2E 与原发布合同/CLI/当前目录回归）、`28-consumer-packages/receipt.json`（release 与 supportmatrix 包完整测试）、`29-live-tag-keyless/receipt.json`（live tag 编译与无 opt-in 的 NOT_RUN 行为）。六条受控原生 bundle 均位于永久证据目录，覆盖九个新型号身份的 81 次受控 HTTP Attempt；材料的 mode 保持 controlled，独立消费及篡改拒绝通过。

边界：新增九个身份自身真实 live、目录晋升及晋升后的新 formal 运行均 `NOT_RUN`，新矩阵保持候选未认证；全局正式验收与 release 仍未通过。旧来源适用性由原生 freshness 检查拒绝；本票未把旧捕获重新标为当前证明。root 后续统一执行原始 checkout 的全量普通门禁与双轴审查。map.md 由 root 统一维护。
