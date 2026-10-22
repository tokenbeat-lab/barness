# 第 1 节复审：Standards

模型：GPT-6 Astra。固定范围：`3ddc774084fe502ca1a13ab8b41b208053e24378...fce5f387e375fa3f7901700f5f525fdb64c6804f`。结论：负责的四项中三项关闭、一项部分关闭；未发现独立新增回归，仍有一项 P1。

| 原问题 | 判断 | 证据 |
| --- | --- | --- |
| 2：混合成功包掩盖失败、未知或授权 | RESOLVED | `cleanup_scan.py:41–88` 检查全部 receipt、manifest、audit 与文件覆盖；mixed 四类拒绝进程场景通过。 |
| 3：未索引失败的传递依赖 | PARTIAL | `cleanup_scan.py:124–139` 已加入失败/未知引用根，既有间接 JSON 场景通过；相对目录仍遗漏，见下。 |
| 4：状态写失败破坏恢复依据 | RESOLVED | `receipt_files.py:44–57` 同目录临时写、fsync、replace；文件限额失败保留 MOVED 原状态，公开 restore 成功。 |
| 5：后代继承管道导致挂起 | RESOLVED | `run_process.py:102–141` deadline 覆盖进程和管道；脱离 session 的后代、超时、真实小 Go 进程取消均通过。65 MiB 输出触发 CAPTURE_FAILED/-9，日志止于 64 MiB。 |

**[P1] 剩余遗漏：相对目录依赖被静默丢弃，仍可永久删除。** [cleanup_scan.py:150](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/cleanup_scan.py:150) 在规范化前仅接受带特定后缀或 `evidence` 字样的路径。失败 receipt 的 `"input":"../unused"` 因而没有引用记录；成功目录 unused 被标 QUARANTINE。临时假仓库中公开 scan→plan→move→verify 全成功，模拟八天观察期后 purge 也成功，失败原件留下而依赖被删，状态 PURGED。这是原第 3 项的未覆盖变体。应先按来源规范化路径、保护落入候选根的文件及目录；歧义须 BLOCKED，并加入目录与非 JSON 工件引用回归。

复现：[脚本](/tmp/barness-section1-relative-dependency-repro.py)、[结果](/tmp/barness-section1-relative-dependency-repro.jsonl)。依据：[plan.md §6.1(3–4)、§6.5](/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/plan.md)、[AGENTS.md 原则9](/Users/cyber/RestoX/barness/AGENTS.md)、[ADR-0051 失败依赖保留](/Users/cyber/RestoX/barness/docs/adr/0051-scoped-verification-receipts.md)。同时遵循 verification、issue-tracker、domain、GLOSSARY 与 ADR-0050 的失败事实及授权保留规则。

执行范围：11 项现有 cleanup 场景及 3 项生命周期场景共 14 项通过（4.691 秒），另做上述依赖及输出限额复现；[进程摘要](/tmp/barness-section1-standards-20261022/processes.jsonl)。所有清理仅用 TemporaryDirectory；Go 父子环境固定 GOENV=off、GOFLAGS 空。只读核对普通、41 场景及最终判定回执哈希均匹配，实际受测 SHA 为 `3aafd2d`；普通 exit0、判定 exit1。未重跑完整 Go、race、压力、付费 live，未调用真实证据 cleanup，未修复或执行后续节。
