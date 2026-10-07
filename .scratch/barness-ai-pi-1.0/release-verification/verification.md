# 工单 02：最终验证与代码审查

2026-10-08 完成。工作起点与审查固定比较点为
`4f1ef33853e0e4853359caed7d242add12736c3c`；变更提交到当前 `main` 分支。

## 验证

| 验证 | 结果 |
| --- | --- |
| 干净安装 / 冻结源码重建 | 独立临时环境重放 PASS；最终脚本再次完整运行并重新生成提交材料 |
| 发布产物逐字节比较 | ai 811/811、telemetry 24/24、合计 835 文件，0 差异 |
| 模型数据 | 43 文件的名称集合、逐文件 SHA-256 与完整模型哈希均一致 |
| 篡改拒绝 | 9 项 PASS，包括秘密审计；实际材料被修改后拒绝，再恢复原字节 |
| 边界故障 | 2 项 PASS：下载读取中超限取消；证据写入失败时清理并标记 FAIL |
| Node 脚本语法检查 | 四个 `.mjs` 文件全部 PASS |
| Go 编译与定向审计测试 | `go test ./ai/internal/testkit/audit/... -run '^$'`；`go test ./ai/internal/testkit/audit/... -count=1` PASS |
| 完整 Go 套件 | `BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test ./... -count=1` PASS |
| 完整 E2E 证据 | 2877 个场景全部 PASS，含原 0.87.1 基线差分 464 个 PASS |
| vet | `go vet ./...` 与 `go vet -tags live ./...` 均 PASS |
| 脱敏审计 | 提交发布件证据与完整 E2E 包均 `findings: []` |
| 完整性清单 | 脚本输入与提交证据的 SHA-256 清单均通过 |
| 当前 oracle | `git diff` 核对 `ai/internal/testkit/pioracle/node` 无变更，保持 0.87.1 |

完整 Go 套件的可重复证据按仓库惯例保存在 git-ignored 目录：
[manifest](../../../.evidence/barness-ai-pi-1.0/02-tests/20261007T212422.784440000Z/manifest.json)、
[audit](../../../.evidence/barness-ai-pi-1.0/02-tests/20261007T212422.784440000Z/audit.json)。
重跑可用：

```sh
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 \
  BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-pi-1.0/02-tests-replay" \
  go test ./... -count=1
```

`pioracle/node` 的原依赖需先按现有 lockfile 安装；本次机器已有该安装。
这条控制验证使用 0.87.1，不是 barness 在 1.0.0 下的差分证明。
新基线差分仍为 **NOT_RUN**，发布件核验脚本的重放与交接见 [README](README.md)。

## Standards

初审发现 2 项文档标准违反，均关联 `AGENTS.md` 原则 4 / 5：
下载应在读取中限额，证据写入失败时也应保证临时资源释放。
先把失败方式与 CLI 边界故障场景加入材料，再修正脚本；下载场景先观察到未取消的红灯，
随后通过；证据最终化的失败分类也先观察到不正确的阶段，再修正为 `evidence-finalization`。

复审结论：两项均已解决；没有剩余文档标准违反或 baseline smell。
读取器逐段计数并取消 / 释放，证据最终化由嵌套 `finally` 无条件清理。
最终化错误以非零状态退出并将结论更新为 FAIL；即使该失败报告也无法写入，清理仍执行。
CLI 场景对网络与文件系统边界施加受控故障，符合失败先行规则，没有生产测试 hook 或推测性抽象。
审查只读，未重跑在线构建。

## Spec

0 项发现。

工单 02 的发布身份、包 integrity、依赖锁、43 个模型数据文件哈希、
两包完整 dist 比较、失败拒绝场景、重放说明与脱敏证据均已交付。
证据明确记录 ai 811/811、telemetry 24/24、0 差异与新基线差分 NOT_RUN。
审查代理独立核对了脚本输入与证据的 SHA-256 清单，并确认当前 package / provenance 保持 0.87.1。
边界修正属于必要的验证加固，没有第二套模型 runner、未提前启用 1.0 provenance，
未发现缺失、错误实现或未经要求的功能范围。

审查汇总：Standards 初审 2 项已修复，最终 0 项；Spec 最终 0 项。
