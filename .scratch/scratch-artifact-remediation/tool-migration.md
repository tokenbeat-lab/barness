# 活动验证工具迁移映射

归属与来源合同见[ADR-0053](../../docs/adr/0053-verification-tool-ownership.md)；唯一当前执行状态见[status](status.md)。当前操作按[测试SOP](../../docs/agents/verification.md)及[新工具合同](../../tools/verification/README.md)。本页只描述路径/职责与历史解析，不另维护预算、轮次或发布结论。

## 当前归属

- `tools/verification/`：run/环境/文件/consumer/cleanup与显式进程测试；三种职责继续独立，普通命令不自动启动下一阶段。
- `ai/internal/testkit/jevformal/verify_protocol.py`：Jev结构核验。Go原生producer/consumer显式固定其新路径/字节；正式资格仍要求来源、公共回放和独立效果。
- 独立Python/Go发布来源边界均为`ai/`、`tools/verification/`和go.mod/go.sum；含忽略执行源码，文档/缓存仅按已有规则排除。

## 历史字节与后继路径

旧脚本、README及固定Git原字节已保存；活动旧路径删除，无wrapper/shim。下表旧路径是originalPath标识，不作为当前执行入口；快照必须按原基址/Git阅读，不能从归档目录导入执行。完整模式/原哈希/当前哈希见[逐文件映射](tool-migration.json)。

| 原路径身份 | 当前入口 | 原字节快照 |
| --- | --- | --- |
| `.scratch/barness-ai-test-process-review/tools/README.md` | [tools/verification/README.md](../../tools/verification/README.md) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/README.md) |
| `.scratch/barness-ai-test-process-review/tools/cleanup.py` | [tools/verification/cleanup.py](../../tools/verification/cleanup.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/cleanup.py) |
| `.scratch/barness-ai-test-process-review/tools/cleanup_scan.py` | [tools/verification/cleanup_scan.py](../../tools/verification/cleanup_scan.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/cleanup_scan.py) |
| `.scratch/barness-ai-test-process-review/tools/cleanup_state.py` | [tools/verification/cleanup_state.py](../../tools/verification/cleanup_state.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/cleanup_state.py) |
| `.scratch/barness-ai-test-process-review/tools/consume_live.py` | [tools/verification/consume_live.py](../../tools/verification/consume_live.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/consume_live.py) |
| `.scratch/barness-ai-test-process-review/tools/live_environment.py` | [tools/verification/live_environment.py](../../tools/verification/live_environment.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/live_environment.py) |
| `.scratch/barness-ai-test-process-review/tools/receipt_files.py` | [tools/verification/receipt_files.py](../../tools/verification/receipt_files.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/receipt_files.py) |
| `.scratch/barness-ai-test-process-review/tools/run.py` | [tools/verification/run.py](../../tools/verification/run.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/run.py) |
| `.scratch/barness-ai-test-process-review/tools/run_process.py` | [tools/verification/run_process.py](../../tools/verification/run_process.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/run_process.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/fixture_child.py` | [tools/verification/tests/fixture_child.py](../../tools/verification/tests/fixture_child.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/fixture_child.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/live_fixture.py` | [tools/verification/tests/live_fixture.py](../../tools/verification/tests/live_fixture.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/live_fixture.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/process_support.py` | [tools/verification/tests/process_support.py](../../tools/verification/tests/process_support.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/process_support.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/test_cleanup.py` | [tools/verification/tests/test_cleanup.py](../../tools/verification/tests/test_cleanup.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/test_cleanup.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py` | [tools/verification/tests/test_live_release.py](../../tools/verification/tests/test_live_release.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/test_live_release.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/test_ordinary.py` | [tools/verification/tests/test_ordinary.py](../../tools/verification/tests/test_ordinary.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/test_ordinary.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/test_release.py` | [tools/verification/tests/test_release.py](../../tools/verification/tests/test_release.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/test_release.py) |
| `.scratch/barness-ai-test-process-review/tools/tests/test_run.py` | [tools/verification/tests/test_run.py](../../tools/verification/tests/test_run.py) | [原件](history/phase5-before-20261024/barness-ai-test-process-review/tools/tests/test_run.py) |
| `.scratch/jev-1.14-multimodal-box/verify-11-evidence.py` | [ai/internal/testkit/jevformal/verify_protocol.py](../../ai/internal/testkit/jevformal/verify_protocol.py) | [原件](history/phase5-before-20261024/jev-1.14-multimodal-box/verify-11-evidence.py) |

## 旧文档与旧资格

第3项脚本分类、第4项追溯、旧review/plan/execution-log及其currentEntry字段均保留原审查版本。涉及上述18个旧入口时按本表或固定原Git读取，不通过旧路径运行；历史工单里的命令不因原件可读而获得新授权。GLM/MiniMax及地区控制器本体不迁移，不改暂停index、旧map/pin或已消费账本。

Jev完整原生来源含新Python路径，地区原生选择器还包含被修改的Go消费者/测试；旧严格来源不能凭搬迁刷新。逐路线影响及原生只读消费结果见唯一状态引用的本轮记录。GLM/MiniMax实际producer范围按各自原生选择器检查，不能因公共辅助路径变化自动宣布全部路线失效或全部不适用。

本项只验证普通/受控进程入口。合成live/race字段是工具拒绝/采集fixture，不是实际厂商、人工race或容量压力；清理测试只操作临时合成repo，真实清理未执行。recovery7保持暂停；回滚代码/调用方/选择器/测试/活动文档成组回退，原件与新失败/新回执继续保留。
