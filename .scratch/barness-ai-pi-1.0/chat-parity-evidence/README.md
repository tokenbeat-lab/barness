# 工单 03 验证记录

2026-10-08。**PASS**：唯一 oracle 已切换至 pi-ai 1.0.0，五项聊天行为与维护者批准的两类 SDK 错误文本已同步。
全量新基线差分 **611 PASS，0 pending**，离线 E2E、压力、race、目录快照、脱敏审计均通过。
详见 [机器可读报告](report.json) 和 [额外错误文本的来源与决定](extra-differences.md)。

| 检查 | 最终结果 |
| --- | --- |
| `go vet ./...`；`go vet -tags live ./ai/...` | PASS |
| runner 版本／模型哈希不符拒绝 | 2 PASS |
| `BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test ./... -count=1` | PASS；3345 E2E 全部 PASS |
| 全量新基线差分 | 611 PASS、0 FAIL、0 pending |
| policy pressure | 7 PASS；额外独立重放 cloud-interactive/design-load PASS |
| `go test -race ./... -count=1` | PASS；E2E 2726 PASS、619 NOT_RUN、0 FAIL |
| 目录快照重新生成与逐字节比较 | PASS |
| 全量 E2E、race 和本记录脱敏审计 | PASS，0 findings |

race 的 619 NOT_RUN 为未开启的差分、压力和冻结目录比较，未记为 PASS；这些 opt-in 场景均在普通完整运行中通过。
本次未运行 live 探针。一次中间版本压力运行出现单个 loopback 流中断，最终代码全量与独立重放均通过；
保留该失败 bundle 于报告，未降低阈值、放宽资源限制或增加重试来掩盖它。

## 基线与行为范围

唯一 oracle 为 `@earendil-works/pi-ai@1.0.0`，提交 `a13d35a742c6ef8462812a28fbe1d8c8b7431c32`。
ai 与 telemetry 精确锁定 1.0.0，沿用工单 02 已核验的依赖图。完整模型数据哈希为
`8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e`。
来源及干净重建等价证明见 [PROVENANCE](../../../ai/internal/testkit/pioracle/node/PROVENANCE.md)。
旧 0.87.1 安装与运行路径已替换；fixture 的旧行为描述保留迁移上下文。

五项修正涵盖有限 Retry-After 与取消、fast/priority/flex 计价、Anthropic 1h 缓存明细覆盖、
full/simple 的型号 samplingParams 合并及终态之后逐工具块完整性检查。
输出索引保持 missing/null/number 独立身份，3×3 匹配矩阵覆盖 full/simple 和 stream/complete。
新增错误对照覆盖 HTTP 对象／数组及具名／未具名 SSE；非对象具名错误在 DTO 解码前抛错，
thread.* 保持原有忽略行为。全部新增项登记为 fixed，原 180 个 extension 决定及其用例列表逐对象一致。
目录按 ADR-0018 更新为 `2026-10-08.1`，报告记录目录哈希；原有型号的字段与价格保持逐字段一致。

## Standards

本轮发现的 P2：Chat 在具名错误判定前解码对象 chunk，可能将非对象错误误报成功。
已先补失败 E2E，再前移错误判定；两轴复核均确认消除。当前 **0 actionable findings**。
提前注册 fixture keys 修复了过滤重放的证据脱敏顺序；thread.* 对照复核无新增 Standards finding。

## Spec

索引缺失与显式 0 合并造成漏报的 P2，以及非对象具名错误的 P2，均先补失败公共 E2E，再修复并复核通过。
HTTP 包装、正文显示和最终 thread.* 控制均经复核，无范围越界，当前 **0 findings**。
维护者已明确批准两类额外文本同步；完整差分零 pending 满足工单 03 的绿色合入要求。

## 重复验证

在仓库根目录运行：

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
node --test ai/internal/testkit/pioracle/node/runner.test.mjs
go vet ./...
go vet -tags live ./ai/...
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test ./... -count=1
go test -race ./... -count=1
go run ./ai/release/cmd/releasegate -write-snapshot -snapshot /tmp/barness-parity-catalog-snapshot.json
cmp ai/release/catalog-snapshot.json /tmp/barness-parity-catalog-snapshot.json
BARNESS_AI_PRESSURE=1 go test ./ai/e2e -count=1 -run '^TestPolicyPressure$/^cloud-interactive$/^design-load$'
```

每个 E2E 输出 `.evidence/barness-ai/<time>/`，包含输入 fixture、线缆请求与帧、双方输出、断言、
差分判定及精确 replay 命令。报告保留最终 bundle 路径、manifest SHA-256 和版本，便于核查本次实际结果。
运行 bundle 为本地证据；源码提交保留可审阅摘要、先红后绿的证据路径和可重复命令。
本记录的 manifest.json 固定 README、机器报告和额外差异说明的 SHA-256。

```sh
go run ./ai/internal/testkit/audit/cmd/auditbundle .evidence/barness-ai/20261007T223425.434379000Z
go run ./ai/internal/testkit/audit/cmd/auditbundle .evidence/barness-ai/20261007T223427.313505000Z
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/chat-parity-evidence
```
