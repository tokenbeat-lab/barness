# 实施与实际证据

2026-10-11，以用户确认的 `39ab6a5` 为起点。公共采样实现和单一 1.1.0 oracle 已完成；整版发布仍待当前完整真实门禁，不能由离线成功关闭工单 06。审查固定点同为 `39ab6a5`。

| 验收 | 实际结果与入口 |
| --- | --- |
| B01 / R10 | PASS：[两个独立来源等价包](release-verification/README.md)，AI 827 + telemetry 24 dist 零差异，模型 43 文件完整哈希；9 项误输入与 2 项 I/O 失败检查通过 |
| B02 / R10 | PASS：唯一 oracle exact lock、provenance、version/commit/data 三项启动拒绝；TestThinkingSamplingBaseline 把两份完整 proof 和 guard 执行带入 release E2E |
| B03 / R05/R10 | PASS：`BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^TestPiDifferential$|^TestCatalogInclusion$' -count=1`，814 差分 + 25 目录用例、全部 839 PASS，pending 0 |
| S01–S17 / R01–R05 | PASS：四条消费者 thinking-sampling.json，四个公共入口及核心 WithHooks；TestNativeThinkingSamplingIgnored 和 native differential 验证忽略 |
| S18–S25 / R06/R08/R09 | PASS：TestCatalogThinkingSampling、TestThinkingSamplingValidation、TestThinkingSamplingSnapshots、TestThinkingSamplingSnapshotCalls；全部档位校验、三层深拷贝、双 map 顺序 Hash、全部既有目录字段核对 |
| S26–S30 / R07/R08 | PASS：TestThinkingSamplingCallbacks / Retry / Concurrent；最终授权/限额拒绝、原地/替换回调、冻结重试字节、取消、发送失败、共享租户与同 ID 独立客户端 |
| durationMs | PASS：TestChatDuration 四入口成功/失败；每条聊天差分先验证非负整数及唯一终态相等，再对有限数值差异登记用户批准的 extension |
| R12 Host 示例 | PASS：thinking-sampling.md 完整 Go 代码块提取后 `go test <generated-go-file>` 编译通过，无真实调用 |
| B04 / R11 | 待完整 gate 完成。当前中文评估 24/24 HTTP 451 失败，前三条 smoke 失败；保存 FAIL，不降低门槛 |

## 需求与场景追溯

`ai/release/traceability.json` 新增 TS-R01–TS-R12；`plan.go` 固定这些 ID 与 normalized mapping hash，保留所有旧 P0、六个差分、三条扩展路线、九组合 live、压力与独立中文效果要求。矩阵映射不是 PASS 的来源；状态从实际 artifacts 判定。

- S01、S09、S10、S11：完整载荷浅合并、空/缺 map、JSON 值；核心七类在每条消费路线覆盖四入口及 Keep Hook。
- S02–S08、S12–S16：literal 夹取矩阵与 full/simple wire，summary-only 和 DeepSeek 原开关；S14/S15 分别对应 S06/S07 与 S04/S05/S12 的 DeepSeek Chat 展开。
- S17/S25：native 忽略与全结构校验。
- S18–S20：构造/Hash 相同错误位置，所有 canonical 档、每个 API 保留键和非法 JSON；参数值不进错误。
- S21–S23：目录/返回值三层修改，真实后续载荷与 Hash；双层 map 插入顺序与值变化。
- S24：TestCatalogInclusion 的整个 Model 逐字段比较，与 release-verification/model-inventory.json 对照；五项真实价格变化同步，没有默认填充或新硬约束。
- S26/S27：thinking-sampling-callback-*；S28/S30：E05-thinking-sampling-*；S29：E06-thinking-sampling-concurrent-*。
- B01/B02：E03-thinking-sampling-baseline-*；B03：TestPiDifferential 全子集；B04：完整 releasegate 的命令、live/effect、压力、审计和固定计划。

## 失败历史与可重复命令

实现前的元数据、快照、采样和 duration 红灯分别保留在当前证据归档；新差分最初仅出现耗时差异。首次基线迁移保存完整 622 个既有差分，发现终态缺 duration 和三个上下文预算差异；[逐 case 清点](release-verification/migration-inventory.json)记录处理，当前这些 case 全部零 pending。未改比较规则或删字段。

fixture-cases.py 重建 literal 合成矩阵；update-ledger.py 仅更新有限耗时 case 和三个 fixed 预算项。完整 release-evidence 包将保存原始受控请求、帧、终态、断言、Attempt、pidiff、manifest 和审计，不将日志摘要替代原包。

```sh
python3 .scratch/barness-ai-thinking-sampling/fixture-cases.py
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
node --test ai/internal/testkit/pioracle/node/runner.test.mjs
go test ./ai/e2e -run '^TestThinkingSampling|^TestCatalogThinkingSampling|^TestChatDuration' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^TestPiDifferential$|^TestCatalogInclusion$' -count=1
go run ./ai/release/cmd/releasegate -write-snapshot
```

真实 smoke 使用 release-evidence/run-live.py，每个子进程仅提取自己组合的 root .env 字段，运行固定有界套件。凭据不进入 argv、命令记录、Observer 或 artifacts；账户别名明确区域/权限未确认。run-evaluation.py 独立只提取 TypeSafe 凭据，当前 config pin 升为 zh-eval-jev-v2，数据集、问题与预算不变。

回滚恢复完整先前实现、目录、oracle lock/provenance、账本和固定发布合同，Host 同时撤回分档元数据配置；无数据库迁移，无内部兼容双路径。当前失败真实请求不覆盖历史成功，也不以历史目录 hash 的报告证明当前发布。
