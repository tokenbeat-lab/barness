# 门禁拆分实施交付

实现已完成，稳定源码 `0433459` 的双vet、一次完整普通测试和普通原生证据消费通过。**race NOT_RUN；完整发布尚未认证，当前不具 READY_FOR_RACE 条件。** 唯一当前状态见 [status.json](status.json)，精确计数见 [verification.json](verification.json)。分支 `codex/manual-race-gate`。

## 已实施

- 项目验证策略定义普通阶段、审查时机、证据与人工race授权；AGENTS只指向该策略。通用 implement-spec 仅修改“读取项目策略”和“TDD遵循项目任务分类”两处，原文与diff可独立回退。
- releasegate 默认 check：输入预检 → 双vet → 完整普通 → 原消费者和审计；默认失败停止。collect-all仅收集可装配check诊断；普通通过仍不启动race。
- 真正普通阶段就绪返回 READY_FOR_RACE / 退出2 / release false；显式race从原check目录续接，重新验证来源、输入、工具、回执和原生bundle；独立race产物通过后调用原十门禁。完整发布PASS才退出0。
- 输入采用有界正规文件读取，保留矩阵原严格解析；命令回收、整体期限、有界失败尾部、实时日志和JSON事件失败均拒绝认证。来源集合包含新增/删除、相关测试、工具及可执行证据脚本。
- 并发与取消既有覆盖核对未发现需重复新增的场景，见 [coverage.json](coverage.json)。人工race清单见 [manual-race.md](manual-race.md)。

## 最终稳定来源的一轮验证

凭据环境只传 PATH/HOME/TMPDIR/GOCACHE/GOPATH/GOROOT；普通开启 BARNESS_AI_PIDIFF=1 与 BARNESS_AI_PRESSURE=1。没有传厂商变量或子测试race开关，没有真实模型调用。普通命令为 `go test -json -count=1 -timeout=20m ./...`，未缩小包集合。

| 命令/检查 | 数字退出 | 秒 | 回执 |
| --- | ---: | ---: | --- |
| vet | 0 | 0.799 | [实际命令](../../.evidence/race-optimization/validation-0433459/vet-command.json) |
| vet-live | 0 | 0.539 | [实际命令](../../.evidence/race-optimization/validation-0433459/vet-live-command.json) |
| ordinary | 0 | 633.110 | [实际命令](../../.evidence/race-optimization/validation-0433459/ordinary-command.json) |
| build-releasegate | 0 | 2.168 | [实际命令](../../.evidence/race-optimization/validation-0433459/build-releasegate-command.json) |
| preflight | 1 | 12.626 | [实际命令](../../.evidence/race-optimization/validation-0433459/preflight-command.json) |
| ordinary-consumption | 1 | 91.602 | [实际命令](../../.evidence/race-optimization/validation-0433459/ordinary-consumption-command.json) |

普通Go命令633.110秒；主E2E包631.409秒，9个有测试的包全部通过。主bundle **12,508 PASS / 0 FAIL / 0 NOT_RUN**，包括 **868条差分，pending 0**；原生 `p0-offline` 和 `differential` 门禁均PASS，普通bundle审计零发现。混合压力、六种撤权、两公开入口图片负载及六种MiniMax设计负载均由原消费者判为PASS。

[普通原件](../../.evidence/race-optimization/validation-0433459/ordinary/20261019T221447.296837000Z/manifest.json)、[运行时审计](../../.evidence/race-optimization/validation-0433459/ordinary/20261019T221447.296837000Z/audit.json)、[只读原生消费报告](../../.evidence/race-optimization/validation-0433459/ordinary-consumption/release-report.json)保留实际结果。只读消费退出1是预期完整发布false：它不声明执行过CLI命令，且race/当前外部资格未满足；不能把这些独立回执伪装成check成功。

54,438条Go事件全部可解析，计时分析见 [performance.md](performance.md)。顶层最大项为Pi差分178.590秒；主包最后测试至包结束88.503秒，仍需分段测量Finish/Audit。包并行与父子测试时间不叠加为墙钟。没有执行额外race取得性能比例。

## 当前发布资格及开放项

[最终预检](../../.evidence/race-optimization/validation-0433459/preflight/preflight-report.json)12.626秒拒绝，未启动任何vet/普通/race命令。17份自身支持中16份有效；Jev1.14自身支持及两份视觉效果的原生来源规则拒绝改动后的发布源码，两份中文效果仍PASS。来源影响详见 [source-impact.json](source-impact.json)。原证明、支持矩阵、selector及历史hash没有回写；没有付费续期。

完整发布验收继续开放：受影响证明的授权续期、有效输入下正式check的READY回执、人工完整race及最终Evaluate、race性能放大比例。本次可以交付实现和普通离线结果，不关闭这些要求完整发布证明的验收项。

## 审查及实现来源

[双轴审查](review.md)分别记录Standards和Spec，最终未修项均0；先行失败与通过日志保留在evidence/。提交 `87c9907`（策略/技能）、`01d01da`（阶段拆分/计时/续接）、`90d827d`（完整追溯/有界读取/进程期限等复审修复）、`0433459`（矩阵原严格消费）。本索引及后续文档提交保持这些执行字节不变。

仓库外技能：[当前 implement-spec](/Users/cyber/.agents/skills/implement-spec/SKILL.md)、[原文](evidence/implement-spec.before.md)、[差异](evidence/implement-spec.diff)。自带验证器缺PyYAML，未安装依赖；已核对frontmatter完全不变且只有两处批准的正文改动，见 [检查记录](evidence/skill-review.json)。
