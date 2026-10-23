# barness-provider-sites 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [tools/final_contract.py](#barness-provider-sites--tools--final_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_execute.py](#barness-provider-sites--tools--final_execute-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_finalize.py](#barness-provider-sites--tools--final_finalize-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_grants.py](#barness-provider-sites--tools--final_grants-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_minimax.py](#barness-provider-sites--tools--final_minimax-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_native.py](#barness-provider-sites--tools--final_native-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_predecessor.py](#barness-provider-sites--tools--final_predecessor-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_process.py](#barness-provider-sites--tools--final_process-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_publication.py](#barness-provider-sites--tools--final_publication-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_retry.py](#barness-provider-sites--tools--final_retry-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_retry_execute.py](#barness-provider-sites--tools--final_retry_execute-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_round.py](#barness-provider-sites--tools--final_round-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/final_state.py](#barness-provider-sites--tools--final_state-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery4_contract.py](#barness-provider-sites--tools--recovery4_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery4_history.py](#barness-provider-sites--tools--recovery4_history-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery4_native.py](#barness-provider-sites--tools--recovery4_native-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery4_round.py](#barness-provider-sites--tools--recovery4_round-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery5_contract.py](#barness-provider-sites--tools--recovery5_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery5_history.py](#barness-provider-sites--tools--recovery5_history-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery5_native.py](#barness-provider-sites--tools--recovery5_native-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery5_retained.py](#barness-provider-sites--tools--recovery5_retained-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery5_round.py](#barness-provider-sites--tools--recovery5_round-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery6_contract.py](#barness-provider-sites--tools--recovery6_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery6_files.py](#barness-provider-sites--tools--recovery6_files-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery6_history.py](#barness-provider-sites--tools--recovery6_history-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery6_native.py](#barness-provider-sites--tools--recovery6_native-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery6_retained.py](#barness-provider-sites--tools--recovery6_retained-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery6_round.py](#barness-provider-sites--tools--recovery6_round-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery7_contract.py](#barness-provider-sites--tools--recovery7_contract-py) | 活动工具 | 真实轮次暂停；合成测试仍消费 |
| [tools/recovery7_files.py](#barness-provider-sites--tools--recovery7_files-py) | 活动工具 | 真实轮次暂停；合成测试仍消费 |
| [tools/recovery7_history.py](#barness-provider-sites--tools--recovery7_history-py) | 活动工具 | 真实轮次暂停；合成测试仍消费 |
| [tools/recovery7_native.py](#barness-provider-sites--tools--recovery7_native-py) | 活动工具 | 真实轮次暂停；合成测试仍消费 |
| [tools/recovery7_retained.py](#barness-provider-sites--tools--recovery7_retained-py) | 活动工具 | 真实轮次暂停；合成测试仍消费 |
| [tools/recovery7_round.py](#barness-provider-sites--tools--recovery7_round-py) | 活动工具 | 真实轮次暂停；合成测试仍消费 |
| [tools/recovery_contract.py](#barness-provider-sites--tools--recovery_contract-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow.py](#barness-provider-sites--tools--recovery_flow-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_execute.py](#barness-provider-sites--tools--recovery_flow_execute-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_files.py](#barness-provider-sites--tools--recovery_flow_files-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_finalize.py](#barness-provider-sites--tools--recovery_flow_finalize-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_native.py](#barness-provider-sites--tools--recovery_flow_native-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_publication.py](#barness-provider-sites--tools--recovery_flow_publication-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_source.py](#barness-provider-sites--tools--recovery_flow_source-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_flow_state.py](#barness-provider-sites--tools--recovery_flow_state-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_history.py](#barness-provider-sites--tools--recovery_history-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_native.py](#barness-provider-sites--tools--recovery_native-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/recovery_round.py](#barness-provider-sites--tools--recovery_round-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [tools/test_closed_fixture.py](#barness-provider-sites--tools--test_closed_fixture-py) | 测试 | 显式合成测试 |
| [tools/test_final_publication.py](#barness-provider-sites--tools--test_final_publication-py) | 测试 | 显式合成测试 |
| [tools/test_final_retry.py](#barness-provider-sites--tools--test_final_retry-py) | 测试 | 显式合成测试 |
| [tools/test_final_round.py](#barness-provider-sites--tools--test_final_round-py) | 测试 | 显式合成测试 |
| [tools/test_recovery4_fixture.py](#barness-provider-sites--tools--test_recovery4_fixture-py) | 测试 | 显式合成测试 |
| [tools/test_recovery4_round.py](#barness-provider-sites--tools--test_recovery4_round-py) | 测试 | 显式合成测试 |
| [tools/test_recovery5_fixture.py](#barness-provider-sites--tools--test_recovery5_fixture-py) | 测试 | 显式合成测试 |
| [tools/test_recovery5_intent_io.py](#barness-provider-sites--tools--test_recovery5_intent_io-py) | 测试 | 显式合成测试 |
| [tools/test_recovery5_io.py](#barness-provider-sites--tools--test_recovery5_io-py) | 测试 | 显式合成测试 |
| [tools/test_recovery5_round.py](#barness-provider-sites--tools--test_recovery5_round-py) | 测试 | 显式合成测试 |
| [tools/test_recovery6_fixture.py](#barness-provider-sites--tools--test_recovery6_fixture-py) | 测试 | 显式合成测试 |
| [tools/test_recovery6_round.py](#barness-provider-sites--tools--test_recovery6_round-py) | 测试 | 显式合成测试 |
| [tools/test_recovery7_fixture.py](#barness-provider-sites--tools--test_recovery7_fixture-py) | 测试 | 显式合成测试 |
| [tools/test_recovery7_round.py](#barness-provider-sites--tools--test_recovery7_round-py) | 测试 | 显式合成测试 |
| [tools/test_recovery_fixture.py](#barness-provider-sites--tools--test_recovery_fixture-py) | 测试 | 显式合成测试 |
| [tools/test_recovery_review_fix.py](#barness-provider-sites--tools--test_recovery_review_fix-py) | 测试 | 显式合成测试 |
| [tools/test_recovery_round.py](#barness-provider-sites--tools--test_recovery_round-py) | 测试 | 显式合成测试 |
| [tools/test_retry_fixture.py](#barness-provider-sites--tools--test_retry_fixture-py) | 测试 | 显式合成测试 |

<a id="barness-provider-sites--tools--final_contract-py"></a>

## tools/final_contract.py

原件：[.scratch/barness-provider-sites/tools/final_contract.py](../../barness-provider-sites/tools/final_contract.py)；SHA256 `f37c001290607f2758f82dadb1c758c33418dc3a78f98b665ccdd95865bd27d6`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：One finite provider-sites collection; argv is constructed here, never imported。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L7（import）；[.scratch/barness-provider-sites/tools/final_finalize.py](../../barness-provider-sites/tools/final_finalize.py) L5（import）；[.scratch/barness-provider-sites/tools/final_grants.py](../../barness-provider-sites/tools/final_grants.py) L5（import）；[.scratch/barness-provider-sites/tools/final_minimax.py](../../barness-provider-sites/tools/final_minimax.py) L3（import）；[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py) L3（import）；[.scratch/barness-provider-sites/tools/final_predecessor.py](../../barness-provider-sites/tools/final_predecessor.py) L8（import）；[.scratch/barness-provider-sites/tools/final_process.py](../../barness-provider-sites/tools/final_process.py) L12（import）；[.scratch/barness-provider-sites/tools/final_publication.py](../../barness-provider-sites/tools/final_publication.py) L2（import）；[.scratch/barness-provider-sites/tools/final_retry.py](../../barness-provider-sites/tools/final_retry.py) L6（import）；[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py) L5（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L8（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L8（import）；[.scratch/barness-provider-sites/tools/recovery4_contract.py](../../barness-provider-sites/tools/recovery4_contract.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery5_contract.py](../../barness-provider-sites/tools/recovery5_contract.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery6_contract.py](../../barness-provider-sites/tools/recovery6_contract.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery7_contract.py](../../barness-provider-sites/tools/recovery7_contract.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery_contract.py](../../barness-provider-sites/tools/recovery_contract.py) L3（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`。

<a id="barness-provider-sites--tools--final_execute-py"></a>

## tools/final_execute.py

原件：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py)；SHA256 `8e821bb60d1a15a89cdf0a9d5253862426221291445155b7d7836411ebd4e39f`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Durable one-use claims, conservative USD reservation and isolated child runs。

**调用方**：[.scratch/barness-provider-sites/tools/final_finalize.py](../../barness-provider-sites/tools/final_finalize.py) L7（import）；[.scratch/barness-provider-sites/tools/final_minimax.py](../../barness-provider-sites/tools/final_minimax.py) L5（import）；[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py) L5（import）；[.scratch/barness-provider-sites/tools/final_retry.py](../../barness-provider-sites/tools/final_retry.py) L126（import）；[.scratch/barness-provider-sites/tools/final_retry.py](../../barness-provider-sites/tools/final_retry.py) L131（import）；[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py) L10（import）；[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py) L23（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L26（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L35（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L53（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L64（import）；[.scratch/barness-provider-sites/tools/recovery_flow_execute.py](../../barness-provider-sites/tools/recovery_flow_execute.py) L112（import）；[.scratch/barness-provider-sites/tools/recovery_flow_native.py](../../barness-provider-sites/tools/recovery_flow_native.py) L57（import）；[.scratch/barness-provider-sites/tools/recovery_flow_native.py](../../barness-provider-sites/tools/recovery_flow_native.py) L94（import）；[.scratch/barness-provider-sites/tools/recovery_flow_publication.py](../../barness-provider-sites/tools/recovery_flow_publication.py) L42（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L23 `root.mkdir(parents=True, exist_ok=True, mode=448)`；L128 `c.dump(out / 'checks.json', result)`；L193 `c.dump(intent, record)`；L196 `c.dump(Path(env['BARNESS_FINAL_BUDGET']), dict(schema=1, round=r.ROUND, maxMicrousd=reserve, reservedMicrousd=0, attempts=0, basis='Conservative final-wire known-rate estimate, not invoice; no release on unknown usage.'))`；L223 `c.dump(intent, record)`；L107 `c.dump(marker(job['name']), c.redact(record), exclusive=True)`；L89 `c.dump(proof, identity, exclusive=True)`；L228 `c.dump(intent, record)`；L231 `c.dump(intent, record)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_finalize-py"></a>

## tools/final_finalize.py

原件：[.scratch/barness-provider-sites/tools/final_finalize.py](../../barness-provider-sites/tools/final_finalize.py)；SHA256 `2b983177644bd66fedbde6eeeaa1b2e07bd64209c9d0352cc80d97e9f5b7db6f`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Exactly23 own current inputs and four effects through the public consumers。

**调用方**：[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L50（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L61 `destination.mkdir(mode=448)`；L62 `c.dump(destination / 'plan.json', c.redact(result))`；L64 `c.dump(destination / 'input-files.json', c.redact(before))`；L75 `matrix.write_bytes(u.file_bytes(c.ROOT / 'ai/live/support-matrix.json'))`；L88 `(destination / 'matrix-before.json').write_bytes(original_bytes)`；L100 `c.dump(destination / 'result.json', c.redact(result))`；L93 `os.chmod(name, target.stat().st_mode & 511)`；L94 `os.replace(name, target)`；L96 `os.unlink(name)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_grants-py"></a>

## tools/final_grants.py

原件：[.scratch/barness-provider-sites/tools/final_grants.py](../../barness-provider-sites/tools/final_grants.py)；SHA256 `aacdc013bedbe3adb148c4ae0a39fd3f1a2a0efec3be842506926bc1b05de9d6`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Only current native component templates and pre-key checks for this round。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L187（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L91（import）；[.scratch/barness-provider-sites/tools/recovery_flow_native.py](../../barness-provider-sites/tools/recovery_flow_native.py) L12（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L15 `c.dump(target, value)`；L23 `c.dump(target, value)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_minimax-py"></a>

## tools/final_minimax.py

原件：[.scratch/barness-provider-sites/tools/final_minimax.py](../../barness-provider-sites/tools/final_minimax.py)；SHA256 `806c0d21eb24a633bf0f489a358f8526be896ec1b8d816183908a20b1549bed1`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Issue only this approved round's fixed native60/72 M3 grant, keylessly。

**调用方**：[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L42（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L25 `c.dump(out / 'native-renewal-facts.json', facts)`；L47 `c.dump(base / 'human-approval.json', human, exclusive=True)`；L48 `c.dump(base / 'source-review.json', review, exclusive=True)`；L54 `c.dump(base / 'grant.json', grant, exclusive=True)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_native-py"></a>

## tools/final_native.py

原件：[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py)；SHA256 `6f17ead63ab5cdeee926838ce23f8e2f3bf57c561461c57b3b200f91f6b5de4e`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Concrete native consumers and canonical byte pins for this final round。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L185（import）；[.scratch/barness-provider-sites/tools/final_publication.py](../../barness-provider-sites/tools/final_publication.py) L3（import）；[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py) L18（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L45（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L152 `receipt_root.mkdir(parents=True, exist_ok=True, mode=448)`；L38 `matrix.write_bytes((c.ROOT / 'ai/live/support-matrix.json').read_bytes())`；L52 `c.dump(e.marker(name), record)`；L64 `matrix.write_bytes((c.ROOT / 'ai/live/support-matrix.json').read_bytes())`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_predecessor-py"></a>

## tools/final_predecessor.py

原件：[.scratch/barness-provider-sites/tools/final_predecessor.py](../../barness-provider-sites/tools/final_predecessor.py)；SHA256 `e6ffeb7170ff9e12d9c06c2d30a006d8265e0af38277d1a576292604be6e6cf3`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Read the one fixed closed first cohort; no research or old authority revival。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L36（import）；[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L74（import）；[.scratch/barness-provider-sites/tools/final_retry.py](../../barness-provider-sites/tools/final_retry.py) L7（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L53（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L118（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L183（import）；[.scratch/barness-provider-sites/tools/recovery_history.py](../../barness-provider-sites/tools/recovery_history.py) L4（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_process-py"></a>

## tools/final_process.py

原件：[.scratch/barness-provider-sites/tools/final_process.py](../../barness-provider-sites/tools/final_process.py)；SHA256 `acc317d65ad7db0af6e8d283889e5dabf7cdb9aa1d985dd0cea5c8df4ade273c`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：One fixed selected-secret child, with exactly-one .env definition policy.  Closed launchers keep their historical parser. This finite round deliberately owns a stricter acquisition boundary; no closed authority or parser is edited。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L210（import）；[.scratch/barness-provider-sites/tools/recovery_flow_execute.py](../../barness-provider-sites/tools/recovery_flow_execute.py) L122（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L42 `c.dump(path, c.redact(record), exclusive=True)`；L48 `c.dump(path, c.redact(record, key))`；L62 `(out / (job['name'] + '.log')).write_text(c.redact(raw.decode(errors='replace'), key))`；L77 `c.dump(path, safe)`；L80 `c.dump(intent, value)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_publication-py"></a>

## tools/final_publication.py

原件：[.scratch/barness-provider-sites/tools/final_publication.py](../../barness-provider-sites/tools/final_publication.py)；SHA256 `96ba91728f799148ba00af3d9115aef620066fada97102ac8f6b84f46bf77c2a`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Only this finite round's new canonical full22 GLM originals may publish。

**调用方**：[.scratch/barness-provider-sites/tools/final_finalize.py](../../barness-provider-sites/tools/final_finalize.py) L59（import）；[.scratch/barness-provider-sites/tools/recovery_flow_finalize.py](../../barness-provider-sites/tools/recovery_flow_finalize.py) L58（import）；[.scratch/barness-provider-sites/tools/recovery_flow_publication.py](../../barness-provider-sites/tools/recovery_flow_publication.py) L51（import）；[.scratch/barness-provider-sites/tools/recovery_flow_publication.py](../../barness-provider-sites/tools/recovery_flow_publication.py) L25（import）；[.scratch/barness-provider-sites/tools/recovery_history.py](../../barness-provider-sites/tools/recovery_history.py) L69（import）；[.scratch/barness-provider-sites/tools/test_recovery_fixture.py](../../barness-provider-sites/tools/test_recovery_fixture.py) L44（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_retry-py"></a>

## tools/final_retry.py

原件：[.scratch/barness-provider-sites/tools/final_retry.py](../../barness-provider-sites/tools/final_retry.py)；SHA256 `f2cefc8e7d37fe2a1017d7d505a623b07f0ff4123fd69eb7236efec003fe06a0`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：One expressly approved Responses amendment; the original33 scope stays closed。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L51（import）；[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L72（import）；[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L181（import）；[.scratch/barness-provider-sites/tools/final_minimax.py](../../barness-provider-sites/tools/final_minimax.py) L36（import）；[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py) L96（import）；[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py) L79（import）；[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py) L4（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L35（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L58（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L149（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L210（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_retry_execute-py"></a>

## tools/final_retry_execute.py

原件：[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py)；SHA256 `c5d300aa26c2ffb37efe45ac0bed91f6aba996de71c960c622ff2009b5505f4b`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Durable accounting for the one fixed Responses retry, with no refunds。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L51（import）；[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L72（import）；[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L181（import）；[.scratch/barness-provider-sites/tools/final_finalize.py](../../barness-provider-sites/tools/final_finalize.py) L55（import）；[.scratch/barness-provider-sites/tools/final_minimax.py](../../barness-provider-sites/tools/final_minimax.py) L36（import）；[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py) L126（import）；[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py) L201（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L38 `c.dump(e.marker(a.NAME), c.redact(record), exclusive=True)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_round-py"></a>

## tools/final_round.py

原件：[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py)；SHA256 `df45d9cae43a22edd19447e259d9d3d7e530dd48c3c22d41696a205d57c6813b`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed provider-sites-final-acceptance-2 operator; no old authority is reused。

**调用方**：[.scratch/barness-provider-sites/tools/test_final_round.py](../../barness-provider-sites/tools/test_final_round.py) L39（literal-reference）；[.scratch/barness-provider-sites/tools/test_final_round.py](../../barness-provider-sites/tools/test_final_round.py) L108（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L154（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'run-selected', 'check', 'summarize', 'final-plan', 'finalize', 'native-grant', 'promote', 'prepare-responses-retry', 'run-responses-retry'))` L14；`parser.add_argument('out', nargs='?')` L15；`parser.add_argument('route', nargs='?')` L16；`parser.add_argument('--phase', choices=r.PHASES)` L17；`parser.add_argument('--inputs')` L18；`parser.add_argument('--execution-root')` L19。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--final_state-py"></a>

## tools/final_state.py

原件：[.scratch/barness-provider-sites/tools/final_state.py](../../barness-provider-sites/tools/final_state.py)；SHA256 `e225cdde380a409093c415ada3694bda775d1cca810851d2a9420d54f965b5c1`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Exact-source freezes and fixed one-round approval pins; never acquires keys。

**调用方**：[.scratch/barness-provider-sites/tools/final_execute.py](../../barness-provider-sites/tools/final_execute.py) L8（import）；[.scratch/barness-provider-sites/tools/final_finalize.py](../../barness-provider-sites/tools/final_finalize.py) L6（import）；[.scratch/barness-provider-sites/tools/final_grants.py](../../barness-provider-sites/tools/final_grants.py) L6（import）；[.scratch/barness-provider-sites/tools/final_minimax.py](../../barness-provider-sites/tools/final_minimax.py) L4（import）；[.scratch/barness-provider-sites/tools/final_native.py](../../barness-provider-sites/tools/final_native.py) L4（import）；[.scratch/barness-provider-sites/tools/final_retry_execute.py](../../barness-provider-sites/tools/final_retry_execute.py) L6（import）；[.scratch/barness-provider-sites/tools/final_round.py](../../barness-provider-sites/tools/final_round.py) L29（import）；[.scratch/barness-provider-sites/tools/recovery_flow_state.py](../../barness-provider-sites/tools/recovery_flow_state.py) L7（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L45 `c.dump(command, value)`；L51 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L52 `c.dump(out / 'source-at-start.json', frozen)`；L55 `c.dump(out / 'predecessor-disclosure.json', predecessor)`；L65 `c.dump(out / 'toolchain.json', c.redact(toolchain))`；L76 `c.dump(out / 'regional-source.json', facts)`；L90 `c.dump(out / 'binaries.json', {name: c.digest(out / name) for name, _ in builds})`；L93 `c.dump(out / 'component-templates.json', component_templates)`；L96 `c.dump(out / 'plan.json', c.redact(plan))`；L101 `c.dump(out / 'prepared.json', prepared)`；L103 `c.dump(out / 'prerelease-source-review.template.json', prepared | dict(decision='PENDING', reviewedBy='', standardsReview=dict(path='', sha256=''), specReview=dict(path='', sha256=''), authorizedRoutes=[]))`；L105 `c.dump(out / 'execution-authorization.template.json', prepared | dict(decision='PENDING', requestText='', requestReference='', eligibilityAuthorizationSHA256=predecessor['authorizationSHA256'], raceAuthorizationSHA256=predecessor['authorizationSHA256'], maxCalls=r.MAX_CALLS, maxAttempts=r.MAX_ATTEMPTS, maxUSD=None, standardCredentialScopes={}, netw`；L59 `c.dump(out / 'execution-context.json', final_retry.context(execution_root))`；L102 `c.dump(out / 'retry-authorization.template.json', final_retry.template(out, prepared))`；L86 `(out / name).chmod(448)`；L87 `c.dump(out / 'frozen-live-copy.json', dict(kind='EXCLUSIVE_FROZEN_BINARY_COPY', source=ctx['live.test'], destinationSHA256=c.digest(out / name), credentialsRead=False))`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery4_contract-py"></a>

## tools/recovery4_contract.py

原件：[.scratch/barness-provider-sites/tools/recovery4_contract.py](../../barness-provider-sites/tools/recovery4_contract.py)；SHA256 `cb4f842d45c50d213441d7b387ce42e78824d5cffa68fe90299956e295a12e73`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed recovery4 execution scope; the unchanged USD guard is policy2 only。

**调用方**：[.scratch/barness-provider-sites/tools/recovery4_history.py](../../barness-provider-sites/tools/recovery4_history.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery4_round.py](../../barness-provider-sites/tools/recovery4_round.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L22（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery4_history-py"></a>

## tools/recovery4_history.py

原件：[.scratch/barness-provider-sites/tools/recovery4_history.py](../../barness-provider-sites/tools/recovery4_history.py)；SHA256 `99bef474636da17a9f7d730c71b39bab6d02f9c732cf20563936285f6e2db848`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed closed3 history; old native PASS is historical, never current qualification。

**调用方**：[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py) L6（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L22（import）；[.scratch/barness-provider-sites/tools/recovery_round.py](../../barness-provider-sites/tools/recovery_round.py) L22（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery4_native-py"></a>

## tools/recovery4_native.py

原件：[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py)；SHA256 `7972e8ac86cf591671cc5464d1d816f40aead890cd88f8d9457ea77c523603dd`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed schema5 projection; M3 authority follows the first own ZAI Responses PASS。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L22（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L60 `c.dump(out / 'native-recovery4-facts.json', value)`；L67 `c.dump(base / 'human-approval.json', human, exclusive=True)`；L67 `c.dump(base / 'source-review.json', review, exclusive=True)`；L70 `c.dump(base / 'grant.json', grant, exclusive=True)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery4_round-py"></a>

## tools/recovery4_round.py

原件：[.scratch/barness-provider-sites/tools/recovery4_round.py](../../barness-provider-sites/tools/recovery4_round.py)；SHA256 `e219d628be6bb8d833bc0a7c24ff48a0cf165edf100562780741ca80a4310a0f`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：The fixed fourth recovery only; pending approval is never paid permission。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L11（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L133（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L177（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L36（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L28（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'check', 'summarize', 'native-grant', 'run-selected', 'final-plan', 'finalize'))` L12；`parser.add_argument('out', nargs='?')` L13；`parser.add_argument('route', nargs='?')` L13；`parser.add_argument('--inputs')` L13。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery5_contract-py"></a>

## tools/recovery5_contract.py

原件：[.scratch/barness-provider-sites/tools/recovery5_contract.py](../../barness-provider-sites/tools/recovery5_contract.py)；SHA256 `93fb17399985217effecc696e0d344d718705daa6f7160b219cfde691e708fa8`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed recovery5: corrected policy2 estimator, nineteen own full processes。

**调用方**：[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py) L8（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery5_round.py](../../barness-provider-sites/tools/recovery5_round.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L26（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery5_history-py"></a>

## tools/recovery5_history.py

原件：[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py)；SHA256 `6147feade08c63e04618e7c11a6d8275c962092b989e04de57ce86bc3d205ff6`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Closed fourth originals and the narrowly justified eight generic retained roles。

**调用方**：[.scratch/barness-provider-sites/tools/recovery4_round.py](../../barness-provider-sites/tools/recovery4_round.py) L22（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery5_retained.py](../../barness-provider-sites/tools/recovery5_retained.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py) L6（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L26（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery5_native-py"></a>

## tools/recovery5_native.py

原件：[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py)；SHA256 `39b5889699b89f2ee383049b3fdd7f6ddaa36509ad1468c2a9f7937437d74980`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed schema6 projection; M3 authority follows the first own ZAI Responses PASS。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L26（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L60 `c.dump(out / 'native-recovery5-facts.json', value)`；L67 `c.dump(base / 'human-approval.json', human, exclusive=True)`；L67 `c.dump(base / 'source-review.json', review, exclusive=True)`；L70 `c.dump(base / 'grant.json', grant, exclusive=True)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery5_retained-py"></a>

## tools/recovery5_retained.py

原件：[.scratch/barness-provider-sites/tools/recovery5_retained.py](../../barness-provider-sites/tools/recovery5_retained.py)；SHA256 `573b65fd901a402db17c7725109a1ea1142e0cea6df369d941e7d556aca6bbdf`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Eight unchanged generic captures, proven against all measured source and wire bytes。

**调用方**：[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py) L95（import）；[.scratch/barness-provider-sites/tools/recovery6_retained.py](../../barness-provider-sites/tools/recovery6_retained.py) L6（import）；[.scratch/barness-provider-sites/tools/recovery7_retained.py](../../barness-provider-sites/tools/recovery7_retained.py) L6（import）；[.scratch/barness-provider-sites/tools/test_recovery5_fixture.py](../../barness-provider-sites/tools/test_recovery5_fixture.py) L63（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L76（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L85（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery5_round-py"></a>

## tools/recovery5_round.py

原件：[.scratch/barness-provider-sites/tools/recovery5_round.py](../../barness-provider-sites/tools/recovery5_round.py)；SHA256 `27d0fe477b14cc5c07a7b9465e302ed247dd01b9d92b3907f5b82e22642b305c`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：The fixed fifth recovery only; pending approval is never paid permission。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery5_intent_io.py](../../barness-provider-sites/tools/test_recovery5_intent_io.py) L43（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery5_io.py](../../barness-provider-sites/tools/test_recovery5_io.py) L41（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L10（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L39（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L28（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'check', 'summarize', 'native-grant', 'run-selected', 'final-plan', 'finalize'))` L12；`parser.add_argument('out', nargs='?')` L13；`parser.add_argument('route', nargs='?')` L13；`parser.add_argument('--inputs')` L13。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery6_contract-py"></a>

## tools/recovery6_contract.py

原件：[.scratch/barness-provider-sites/tools/recovery6_contract.py](../../barness-provider-sites/tools/recovery6_contract.py)；SHA256 `d54702570134783bda55d12bdbde24dce6bb4d8f9ee15ff1a03870e2edd38c01`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed recovery6: corrected policy2 estimator, nineteen own full processes。

**调用方**：[.scratch/barness-provider-sites/tools/recovery6_files.py](../../barness-provider-sites/tools/recovery6_files.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery6_round.py](../../barness-provider-sites/tools/recovery6_round.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L30（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery6_files-py"></a>

## tools/recovery6_files.py

原件：[.scratch/barness-provider-sites/tools/recovery6_files.py](../../barness-provider-sites/tools/recovery6_files.py)；SHA256 `aed1325f26ef0664b34786c79bb27ab7bf6968e280cf62536aeed164abae5488`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：No-follow bounded inventory of the two sealed fifth collection namespaces。

**调用方**：[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py) L55（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery6_history-py"></a>

## tools/recovery6_history.py

原件：[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py)；SHA256 `849618678824f53f94057e35e6b4bd53fff80eae7060a648cf7518496b008654`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Immutable actual fifth history; current native qualification remains separate。

**调用方**：[.scratch/barness-provider-sites/tools/recovery5_round.py](../../barness-provider-sites/tools/recovery5_round.py) L22（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery6_retained.py](../../barness-provider-sites/tools/recovery6_retained.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py) L6（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L30（import）；[.scratch/barness-provider-sites/tools/test_recovery6_fixture.py](../../barness-provider-sites/tools/test_recovery6_fixture.py) L51（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L159（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery6_native-py"></a>

## tools/recovery6_native.py

原件：[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py)；SHA256 `189bf07944e1214c9b68ed399e4ce68ae8d44cd55ce196a75deb6e5d769f27b3`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed schema7 projection; M3 authority follows the first own ZAI Responses PASS。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L30（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L60 `c.dump(out / 'native-recovery6-facts.json', value)`；L67 `c.dump(base / 'human-approval.json', human, exclusive=True)`；L67 `c.dump(base / 'source-review.json', review, exclusive=True)`；L70 `c.dump(base / 'grant.json', grant, exclusive=True)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery6_retained-py"></a>

## tools/recovery6_retained.py

原件：[.scratch/barness-provider-sites/tools/recovery6_retained.py](../../barness-provider-sites/tools/recovery6_retained.py)；SHA256 `e8dcd419502f02583349e3aefee14a7668054c52f46943703c9f000012b2dadf`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Eight exact old08 inputs require a newly pinned84-wire/header/ledger proof。

**调用方**：[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py) L115（import）；[.scratch/barness-provider-sites/tools/test_recovery6_fixture.py](../../barness-provider-sites/tools/test_recovery6_fixture.py) L52（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery6_fixture.py](../../barness-provider-sites/tools/test_recovery6_fixture.py) L64（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L178（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery6_round-py"></a>

## tools/recovery6_round.py

原件：[.scratch/barness-provider-sites/tools/recovery6_round.py](../../barness-provider-sites/tools/recovery6_round.py)；SHA256 `8c05fab53ba62ca0b7d71e7c5a8acc8aa6ff733f23a4166432b33271791deb68`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：The fixed sixth recovery only; pending approval is never paid permission。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L10（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L28（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L34（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L49（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L69（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`parser.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'check', 'summarize', 'native-grant', 'run-selected', 'final-plan', 'finalize'))` L12；`parser.add_argument('out', nargs='?')` L13；`parser.add_argument('route', nargs='?')` L13；`parser.add_argument('--inputs')` L13。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery7_contract-py"></a>

## tools/recovery7_contract.py

原件：[.scratch/barness-provider-sites/tools/recovery7_contract.py](../../barness-provider-sites/tools/recovery7_contract.py)；SHA256 `b1f8fb0263df3bd8e144c41de8146fdc3389c8a934b89a580a9113b6992bea38`。活动工具 / 真实轮次暂停；合成测试仍消费。真实recovery7暂停；仍被合成测试加载。

**职责**：Fixed recovery7: corrected policy2 estimator, nineteen own full processes。

**调用方**：[.scratch/barness-provider-sites/tools/recovery7_files.py](../../barness-provider-sites/tools/recovery7_files.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery7_round.py](../../barness-provider-sites/tools/recovery7_round.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L34（import）。

**输入**：固定recovery7合同及closed6原件、源码/wire证明、合成测试fixture；真实human7/grant7/paidClaims7尚未建立。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：实现定义新轮次plan/claim/原生索引/回执/发布输出；本次均不生成，暂停index是唯一事实入口。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：prepare/check可能构建/运行测试；run-selected可能厂商付费，finalize可能发布。真实recovery7保持暂停，合成测试消费源码不构成继续授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[.scratch/barness-provider-sites/tools/README.md](../../barness-provider-sites/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery7_files-py"></a>

## tools/recovery7_files.py

原件：[.scratch/barness-provider-sites/tools/recovery7_files.py](../../barness-provider-sites/tools/recovery7_files.py)；SHA256 `0c9faba8a49f40caf10a13538ae3bbc5ebb873717dcf42c55480372d2b90730b`。活动工具 / 真实轮次暂停；合成测试仍消费。真实recovery7暂停；仍被合成测试加载。

**职责**：No-follow bounded inventory of the two sealed sixth collection namespaces。

**调用方**：[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py) L48（import）。

**输入**：固定recovery7合同及closed6原件、源码/wire证明、合成测试fixture；真实human7/grant7/paidClaims7尚未建立。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：实现定义新轮次plan/claim/原生索引/回执/发布输出；本次均不生成，暂停index是唯一事实入口。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：prepare/check可能构建/运行测试；run-selected可能厂商付费，finalize可能发布。真实recovery7保持暂停，合成测试消费源码不构成继续授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[.scratch/barness-provider-sites/tools/README.md](../../barness-provider-sites/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery7_history-py"></a>

## tools/recovery7_history.py

原件：[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py)；SHA256 `24b598c3898a1c75607308e0c68b4a64ba9c03b2f81434cffeb33d43c17c0f27`。活动工具 / 真实轮次暂停；合成测试仍消费。真实recovery7暂停；仍被合成测试加载。

**职责**：Immutable actual sixth history, including one interrupted unreported M3 child。

**调用方**：[.scratch/barness-provider-sites/tools/recovery6_round.py](../../barness-provider-sites/tools/recovery6_round.py) L22（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery7_retained.py](../../barness-provider-sites/tools/recovery7_retained.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L34（import）；[.scratch/barness-provider-sites/tools/test_recovery7_fixture.py](../../barness-provider-sites/tools/test_recovery7_fixture.py) L50（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L174（literal-reference）。

**输入**：固定recovery7合同及closed6原件、源码/wire证明、合成测试fixture；真实human7/grant7/paidClaims7尚未建立。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：实现定义新轮次plan/claim/原生索引/回执/发布输出；本次均不生成，暂停index是唯一事实入口。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：prepare/check可能构建/运行测试；run-selected可能厂商付费，finalize可能发布。真实recovery7保持暂停，合成测试消费源码不构成继续授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[.scratch/barness-provider-sites/tools/README.md](../../barness-provider-sites/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery7_native-py"></a>

## tools/recovery7_native.py

原件：[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py)；SHA256 `9402ebe2b674c0cd5ed1d754af4710b5643cefd54e2897324123114b1a0ec19c`。活动工具 / 真实轮次暂停；合成测试仍消费。真实recovery7暂停；仍被合成测试加载。

**职责**：Fixed schema8 projection; M3 authority follows the first own ZAI Responses PASS。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L34（import）。

**输入**：固定recovery7合同及closed6原件、源码/wire证明、合成测试fixture；真实human7/grant7/paidClaims7尚未建立。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：实现定义新轮次plan/claim/原生索引/回执/发布输出；本次均不生成，暂停index是唯一事实入口。实际静态写点：L64 `c.dump(out / 'native-recovery7-facts.json', value)`；L71 `c.dump(base / 'human-approval.json', human, exclusive=True)`；L71 `c.dump(base / 'source-review.json', review, exclusive=True)`；L74 `c.dump(base / 'grant.json', grant, exclusive=True)`。

**副作用**：prepare/check可能构建/运行测试；run-selected可能厂商付费，finalize可能发布。真实recovery7保持暂停，合成测试消费源码不构成继续授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[.scratch/barness-provider-sites/tools/README.md](../../barness-provider-sites/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery7_retained-py"></a>

## tools/recovery7_retained.py

原件：[.scratch/barness-provider-sites/tools/recovery7_retained.py](../../barness-provider-sites/tools/recovery7_retained.py)；SHA256 `f1d9d368043bcf4b3049be5836b679ba0ed1f25e78d7ad668b33ca75f5d58fda`。活动工具 / 真实轮次暂停；合成测试仍消费。真实recovery7暂停；仍被合成测试加载。

**职责**：Eight exact old08 inputs require a newly pinned84-wire/header/ledger proof。

**调用方**：[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py) L126（import）；[.scratch/barness-provider-sites/tools/test_recovery7_fixture.py](../../barness-provider-sites/tools/test_recovery7_fixture.py) L51（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_fixture.py](../../barness-provider-sites/tools/test_recovery7_fixture.py) L56（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L193（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L221（literal-reference）。

**输入**：固定recovery7合同及closed6原件、源码/wire证明、合成测试fixture；真实human7/grant7/paidClaims7尚未建立。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：实现定义新轮次plan/claim/原生索引/回执/发布输出；本次均不生成，暂停index是唯一事实入口。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：prepare/check可能构建/运行测试；run-selected可能厂商付费，finalize可能发布。真实recovery7保持暂停，合成测试消费源码不构成继续授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[.scratch/barness-provider-sites/tools/README.md](../../barness-provider-sites/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery7_round-py"></a>

## tools/recovery7_round.py

原件：[.scratch/barness-provider-sites/tools/recovery7_round.py](../../barness-provider-sites/tools/recovery7_round.py)；SHA256 `251350d37d0f5e8ea5cb29b629439d620d3f13085b3f17614045049162cc25fd`。活动工具 / 真实轮次暂停；合成测试仍消费。真实recovery7暂停；仍被合成测试加载。

**职责**：The fixed seventh recovery only; pending approval is never paid permission。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L10（literal-reference）。

**输入**：固定recovery7合同及closed6原件、源码/wire证明、合成测试fixture；真实human7/grant7/paidClaims7尚未建立。参数：`parser.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'check', 'summarize', 'native-grant', 'run-selected', 'final-plan', 'finalize'))` L12；`parser.add_argument('out', nargs='?')` L13；`parser.add_argument('route', nargs='?')` L13；`parser.add_argument('--inputs')` L13。

**输出/写入**：实现定义新轮次plan/claim/原生索引/回执/发布输出；本次均不生成，暂停index是唯一事实入口。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：prepare/check可能构建/运行测试；run-selected可能厂商付费，finalize可能发布。真实recovery7保持暂停，合成测试消费源码不构成继续授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[.scratch/barness-provider-sites/tools/README.md](../../barness-provider-sites/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_contract-py"></a>

## tools/recovery_contract.py

原件：[.scratch/barness-provider-sites/tools/recovery_contract.py](../../barness-provider-sites/tools/recovery_contract.py)；SHA256 `11b2d35a497cd76a7c8b209c27036cfdfbb6279f54794d86651bb2e1b6db60fa`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed recovery3 execution scope; the unchanged USD guard is policy2 only。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L18（import）；[.scratch/barness-provider-sites/tools/recovery_history.py](../../barness-provider-sites/tools/recovery_history.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery_native.py](../../barness-provider-sites/tools/recovery_native.py) L3（import）；[.scratch/barness-provider-sites/tools/recovery_round.py](../../barness-provider-sites/tools/recovery_round.py) L7（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow-py"></a>

## tools/recovery_flow.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py)；SHA256 `3c4b943dacbace69b7e3a52d9e1cffdb43cfefc3252045c985349539e81d6d81`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Private assembly for the actual fixed recovery workflows, never a registry。

**调用方**：[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py) L23（import）；[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py) L33（import）；[.scratch/barness-provider-sites/tools/recovery4_native.py](../../barness-provider-sites/tools/recovery4_native.py) L54（import）；[.scratch/barness-provider-sites/tools/recovery4_round.py](../../barness-provider-sites/tools/recovery4_round.py) L18（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L23（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L33（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L54（import）；[.scratch/barness-provider-sites/tools/recovery5_native.py](../../barness-provider-sites/tools/recovery5_native.py) L75（import）；[.scratch/barness-provider-sites/tools/recovery5_round.py](../../barness-provider-sites/tools/recovery5_round.py) L18（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L23（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L33（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L54（import）；[.scratch/barness-provider-sites/tools/recovery6_native.py](../../barness-provider-sites/tools/recovery6_native.py) L75（import）；[.scratch/barness-provider-sites/tools/recovery6_round.py](../../barness-provider-sites/tools/recovery6_round.py) L18（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L27（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L37（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L58（import）；[.scratch/barness-provider-sites/tools/recovery7_native.py](../../barness-provider-sites/tools/recovery7_native.py) L79（import）；[.scratch/barness-provider-sites/tools/recovery7_round.py](../../barness-provider-sites/tools/recovery7_round.py) L16（import）；[.scratch/barness-provider-sites/tools/recovery_native.py](../../barness-provider-sites/tools/recovery_native.py) L8（import）；[.scratch/barness-provider-sites/tools/recovery_native.py](../../barness-provider-sites/tools/recovery_native.py) L23（import）；[.scratch/barness-provider-sites/tools/recovery_round.py](../../barness-provider-sites/tools/recovery_round.py) L18（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_execute-py"></a>

## tools/recovery_flow_execute.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_execute.py](../../barness-provider-sites/tools/recovery_flow_execute.py)；SHA256 `1121bfc260842da1c692286b65f0f61721296deede1c4b84f5af1ac653b4acd4`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Shared private fixed recovery execute mechanics; authority values come only from compiled contexts。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L4（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L36 `base.mkdir(parents=True, exist_ok=True, mode=448)`；L131 `self.flow.r.c.dump(intent, record)`；L134 `self.flow.r.c.dump(usd, dict(schema=1, round=self.flow.r.SPENDING_GUARD_POLICY_SELECTOR, maxMicrousd=reserve, reservedMicrousd=0, attempts=0, basis=self.flow.r.LEDGER_BASIS + '; executionAuthority=' + self.flow.r.ROUND + '; spendingGuardSourceSHA256=' + json.dumps(self.flow.r.policy()['spendingGuardSourceSHA256'], sort_keys=True) + '; no refund; op`；L153 `self.flow.r.c.dump(intent, record)`；L107 `self.flow.r.c.dump(self.marker(job['name']), self.flow.r.c.redact(record), exclusive=True)`；L100 `self.flow.r.c.dump(proof, identity, exclusive=True)`；L158 `self.flow.r.c.dump(intent, record)`；L161 `self.flow.r.c.dump(intent, record)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_files-py"></a>

## tools/recovery_flow_files.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_files.py](../../barness-provider-sites/tools/recovery_flow_files.py)；SHA256 `13260a3df872ab081c9c1ca3128f2943e6dbd28ea6d74d4a222362b44b4f9835`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Bounded common-owner claim inventory for the compiled fixed contexts。

**调用方**：[.scratch/barness-provider-sites/tools/final_process.py](../../barness-provider-sites/tools/final_process.py) L78（import）；[.scratch/barness-provider-sites/tools/final_process.py](../../barness-provider-sites/tools/final_process.py) L65（import）；[.scratch/barness-provider-sites/tools/recovery4_history.py](../../barness-provider-sites/tools/recovery4_history.py) L6（import）；[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py) L8（import）；[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py) L8（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L2（import）；[.scratch/barness-provider-sites/tools/recovery_flow_execute.py](../../barness-provider-sites/tools/recovery_flow_execute.py) L8（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_finalize-py"></a>

## tools/recovery_flow_finalize.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_finalize.py](../../barness-provider-sites/tools/recovery_flow_finalize.py)；SHA256 `9626801c0a9974ddb483d650786f3da5ebc48f06ee1ad570fa8b97b54aa3a0aa`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Shared private fixed recovery finalize mechanics; authority values come only from compiled contexts。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L7（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L61 `destination.mkdir(mode=448)`；L62 `self.flow.r.c.dump(destination / 'plan.json', self.flow.r.c.redact(result))`；L63 `self.flow.r.c.dump(destination / self.flow.r.FINAL_SELECTION_FILE, dict(roles=roles, accounting=summary))`；L65 `self.flow.r.c.dump(destination / 'input-files.json', self.flow.r.c.redact(before))`；L77 `matrix.write_bytes(self.flow.files.read(self.flow.r.c.ROOT / 'ai/live/support-matrix.json'))`；L90 `(destination / 'matrix-before.json').write_bytes(original_bytes)`；L103 `self.flow.r.c.dump(destination / 'result.json', self.flow.r.c.redact(result))`；L97 `os.chmod(name, target.stat().st_mode & 511)`；L98 `os.replace(name, target)`；L101 `os.unlink(name)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_native-py"></a>

## tools/recovery_flow_native.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_native.py](../../barness-provider-sites/tools/recovery_flow_native.py)；SHA256 `66876b7d6b11e56c32692a0c2c0953c1b363c748e5acb89d1111fb0713ec0489`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Shared private fixed recovery native mechanics; authority values come only from compiled contexts。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L5（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L77 `self.flow.r.c.dump(self.flow.e.marker(name), record)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_publication-py"></a>

## tools/recovery_flow_publication.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_publication.py](../../barness-provider-sites/tools/recovery_flow_publication.py)；SHA256 `7f821986b6d1e111edd2b53fc97391a6c4a100286005748c446b8b391c08ad74`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Shared private fixed recovery publication mechanics; authority values come only from compiled contexts。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L6（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_source-py"></a>

## tools/recovery_flow_source.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_source.py](../../barness-provider-sites/tools/recovery_flow_source.py)；SHA256 `f36e44801b9a49d6b55bd2fb3d88edfc4beb501bdfcdd86ec5f387c4ef84ff41`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Exact existing runtime/operator selector with bounded nonblocking file hashes。

**调用方**：[.scratch/barness-provider-sites/tools/final_process.py](../../barness-provider-sites/tools/final_process.py) L21（import）；[.scratch/barness-provider-sites/tools/recovery4_contract.py](../../barness-provider-sites/tools/recovery4_contract.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery5_contract.py](../../barness-provider-sites/tools/recovery5_contract.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery5_history.py](../../barness-provider-sites/tools/recovery5_history.py) L10（import）；[.scratch/barness-provider-sites/tools/recovery5_retained.py](../../barness-provider-sites/tools/recovery5_retained.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery6_contract.py](../../barness-provider-sites/tools/recovery6_contract.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery6_history.py](../../barness-provider-sites/tools/recovery6_history.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery6_retained.py](../../barness-provider-sites/tools/recovery6_retained.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery7_contract.py](../../barness-provider-sites/tools/recovery7_contract.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery7_history.py](../../barness-provider-sites/tools/recovery7_history.py) L9（import）；[.scratch/barness-provider-sites/tools/recovery7_retained.py](../../barness-provider-sites/tools/recovery7_retained.py) L7（import）；[.scratch/barness-provider-sites/tools/recovery_contract.py](../../barness-provider-sites/tools/recovery_contract.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery_flow_files.py](../../barness-provider-sites/tools/recovery_flow_files.py) L4（import）；[.scratch/barness-provider-sites/tools/recovery_flow_state.py](../../barness-provider-sites/tools/recovery_flow_state.py) L8（import）；[.scratch/barness-provider-sites/tools/test_recovery5_fixture.py](../../barness-provider-sites/tools/test_recovery5_fixture.py) L18（import）；[.scratch/barness-provider-sites/tools/test_recovery6_fixture.py](../../barness-provider-sites/tools/test_recovery6_fixture.py) L19（import）；[.scratch/barness-provider-sites/tools/test_recovery7_fixture.py](../../barness-provider-sites/tools/test_recovery7_fixture.py) L19（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_flow_state-py"></a>

## tools/recovery_flow_state.py

原件：[.scratch/barness-provider-sites/tools/recovery_flow_state.py](../../barness-provider-sites/tools/recovery_flow_state.py)；SHA256 `722db71e2251f27453706dfa638465aad43a2cfe9b51aae4801fefef670f8b64`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Shared private fixed recovery state mechanics; authority values come only from compiled contexts。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L3（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L38 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L44 `self.flow.r.c.dump(out / 'toolchain.json', self.flow.r.c.redact(toolchain))`；L48 `self.flow.r.c.dump(out / 'regional-source.json', source)`；L55 `self.flow.r.c.dump(out / 'binaries.json', {name: SourceFreeze(self.flow.r.c).digest(out / name) for name, _ in builds})`；L57 `self.flow.r.c.dump(out / 'component-templates.json', templates)`；L58 `self.flow.r.c.dump(out / 'plan.json', self.flow.r.c.redact(self.flow.r.inventory(out)))`；L62 `self.flow.r.c.dump(out / 'prepared.json', prepared)`；L63 `self.flow.r.c.dump(out / 'prerelease-source-review.template.json', prepared | dict(decision='PENDING', reviewedBy='', standardsReview=dict(path='', sha256=''), specReview=dict(path='', sha256=''), authorizedRoutes=[]))`；L66 `self.flow.r.c.dump(out / 'execution-authorization.template.json', auth)`；L67 `self.flow.r.c.dump(out / self.flow.r.ACTUAL_HUMAN_LEAF.replace('.json', '.template.json'), {k: v for k, v in auth.items() if k not in ('standardCredentialScopes', 'componentAuthorizationSHA256', 'networkRoutes', 'predecessor')})`；L137 `self.flow.r.c.dump(out / 'checks.json', result)`；L40 `self.flow.r.c.dump(out / leaf, value)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_history-py"></a>

## tools/recovery_history.py

原件：[.scratch/barness-provider-sites/tools/recovery_history.py](../../barness-provider-sites/tools/recovery_history.py)；SHA256 `3dbfb3d8c76991557570d37ef5924a157f51e7317dc6f3dcac08bf2d8872243e`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Read-only closed2 facts; all old failures and occupied allowance survive。

**调用方**：[.scratch/barness-provider-sites/tools/recovery4_history.py](../../barness-provider-sites/tools/recovery4_history.py) L5（import）；[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L18（import）；[.scratch/barness-provider-sites/tools/recovery_native.py](../../barness-provider-sites/tools/recovery_native.py) L4（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_native-py"></a>

## tools/recovery_native.py

原件：[.scratch/barness-provider-sites/tools/recovery_native.py](../../barness-provider-sites/tools/recovery_native.py)；SHA256 `30659fc543edd85abb253338bb7976beadd3a95225494b57042cd8205bb27610`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed3 native facts and schema4 issuance; historical authority stays unchanged。

**调用方**：[.scratch/barness-provider-sites/tools/recovery_flow.py](../../barness-provider-sites/tools/recovery_flow.py) L18（import）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：L30 `c.dump(out / 'native-recovery-facts.json', value)`；L37 `c.dump(base / 'human-approval.json', human, exclusive=True)`；L38 `c.dump(base / 'source-review.json', review, exclusive=True)`；L41 `c.dump(base / 'grant.json', grant, exclusive=True)`。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--recovery_round-py"></a>

## tools/recovery_round.py

原件：[.scratch/barness-provider-sites/tools/recovery_round.py](../../barness-provider-sites/tools/recovery_round.py)；SHA256 `600259a543bfcd56d981690d7c4c8e96ac0bdfdc5c92831fa008ccca4955ffa7`。活动工具 / 仅测试消费，旧真实轮次关闭。原真实轮次关闭；现有合成测试及recovery7模块仍消费，不能删/迁移或重新操作。

**职责**：Fixed recovery3; no ambient round override, paid restart or old2 authority。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L30（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L28（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L12（literal-reference）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L165（literal-reference）。

**输入**：原固定轮次编译合同、源码快照、旧历史/原生资格、一次性审批/grant/claim/余额、所选输出目录/路线；helper接收父模块数据。参数：`p.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'check', 'summarize', 'native-grant', 'run-selected', 'final-plan', 'finalize'))` L12；`p.add_argument('out', nargs='?')` L13；`p.add_argument('route', nargs='?')` L13；`p.add_argument('--inputs')` L13。

**输出/写入**：轮次专属plan/账本/intent/命令回执/捕获/汇总，部分finalize可写支持矩阵/派生数据。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：模块能力由下方函数/调用点区分；整条工作流可读所选.env、构建/普通测试/付费子进程/发布数据。当前只保留合成进程测试用途，关闭余额和批准不得复活。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_closed_fixture-py"></a>

## tools/test_closed_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_closed_fixture.py](../../barness-provider-sites/tools/test_closed_fixture.py)；SHA256 `fb11c864023efe3c7e4f3fb8e49ca21ced9465cc753a8cbcbddbcb931c20f942`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Synthetic closed first-cohort facts for public process tests, never live proof。

**调用方**：[.scratch/barness-provider-sites/tools/test_final_publication.py](../../barness-provider-sites/tools/test_final_publication.py) L22（import）；[.scratch/barness-provider-sites/tools/test_final_round.py](../../barness-provider-sites/tools/test_final_round.py) L33（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L21 `(root / parent / 'launch.lock').write_bytes(b'')`；L13 `path.parent.mkdir(parents=True, exist_ok=True)`；L15 `path.write_bytes(content)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_final_publication-py"></a>

## tools/test_final_publication.py

原件：[.scratch/barness-provider-sites/tools/test_final_publication.py](../../barness-provider-sites/tools/test_final_publication.py)；SHA256 `44f00d23aaa03f2d5962e932ca9e0e5e385da41be5f6a7d0fc54afee3eb19f39`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public new-round GLM publication refuses every closed canonical fallback。

**调用方**：[.scratch/barness-provider-sites/tools/test_retry_fixture.py](../../barness-provider-sites/tools/test_retry_fixture.py) L4（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L20 `shutil.copytree(HERE, self.root / '.scratch/barness-provider-sites/tools', ignore=shutil.ignore_patterns('__pycache__'))`；L86 `marker.write_text(json.dumps(value))`；L98 `target.write_text(json.dumps(index))`；L99 `(target.parent / 'live-captures.sha256').write_text(hashlib.sha256(target.read_bytes()).hexdigest() + '\n')`；L101 `checkpoint.write_text(json.dumps(state))`；L30 `out.mkdir(parents=True, exist_ok=True)`；L31 `(out / 'processes.json').write_text(json.dumps(self.journal, indent=2) + '\n')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`。

<a id="barness-provider-sites--tools--test_final_retry-py"></a>

## tools/test_final_retry.py

原件：[.scratch/barness-provider-sites/tools/test_final_retry.py](../../barness-provider-sites/tools/test_final_retry.py)；SHA256 `b6d6a0bfefb8ea3c883e47c1e31f884771c5dab7f6f12d91273ee9342f499495`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public fixed amendment processes use synthetic keys and a separate Git owner。

**调用方**：[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L16 `(self.out / 'retry-authorization.json').write_text(json.dumps(auth))`；L19 `(self.root / '.env').write_text('ZHIPU_KEY=SYNTHETIC-OWN\n')`；L46 `(self.root / '.env').write_text('ZHIPU_KEY=SYNTHETIC-OWN\n')`；L48 `(self.root / '.env').unlink()`；L48 `(self.root / '.env').mkdir()`；L55 `(self.checkout / 'ai/sentinel.go').write_text('package changed\n')`；L60 `(self.root / '.env').write_text('ZHIPU_KEY=SYNTHETIC-OWN\n')`；L65 `auth.write_text(json.dumps(value))`；L70 `failed.write_text(json.dumps(value))`；L71 `failed.write_bytes(before)`；L72 `fake.write_text(json.dumps(dict(state='COMPLETED', exitCode=0, budget=dict(state='REPORTED', status='PASS', calls=22, attempts=22))))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_final_round-py"></a>

## tools/test_final_round.py

原件：[.scratch/barness-provider-sites/tools/test_final_round.py](../../barness-provider-sites/tools/test_final_round.py)；SHA256 `20ac3b2fd31d3f12effd30c650a3b30f43ff8f283cdfdd3a9a96c90f3298cd53`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Socket-free public-process checks; all approvals/keys are synthetic。

**调用方**：[.scratch/barness-provider-sites/tools/test_final_retry.py](../../barness-provider-sites/tools/test_final_retry.py) L4（import）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L4（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L20 `shutil.copytree(HERE, self.tools, ignore=shutil.ignore_patterns('__pycache__'))`；L21 `shutil.copytree(REPO / '.scratch/barness-ai-minimax/tools', self.root / '.scratch/barness-ai-minimax/tools', ignore=shutil.ignore_patterns('__pycache__'))`；L22 `(self.root / 'ai').mkdir()`；L23 `(self.root / 'ai/live').mkdir()`；L24 `(self.root / 'ai/sentinel.go').write_text('package controlled\n')`；L25 `(self.root / 'ai/release').mkdir()`；L26 `(self.root / 'ai/release/catalog-snapshot.json').write_text(json.dumps(dict(hash='sha256:controlled', catalog=dict(version='controlled'))))`；L27 `(self.root / 'go.mod').write_text('module controlled\n')`；L28 `(self.root / 'go.sum').write_text('')`；L38 `(self.root / '.env').mkdir()`；L65 `bindir.mkdir()`；L100 `go.write_text('#!' + sys.executable + '\nimport sys,json,pathlib\na=sys.argv[1:]\nif "-o" in a:\n p=pathlib.Path(a[a.index("-o")+1]);p.write_text(' + repr(producer) + ');p.chmod(0o755)\n' + native_source)`；L101 `go.chmod(493)`；L125 `seal.unlink()`；L156 `marker.write_text('{}')`；L158 `marker.unlink()`；L160 `native.write_text('{}')`；L162 `native.unlink()`；L164 `seal.write_text(json.dumps(data))`；L181 `(self.root / 'ai/sentinel.go').write_text('package changed\n')`；L187 `source.parent.mkdir(parents=True)`；L187 `source.write_bytes(b'controlled embedded original')`；L191 `source.write_bytes(b'controlled changed embedded original')`；L193 `source.write_bytes(original)`；L194 `extra.write_bytes(b'controlled added original')`；L196 `extra.unlink()`；L196 `source.unlink()`；L198 `source.write_bytes(original)`；L228 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L241 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L268 `(self.root / '.env').write_text('ZAI_KEY=SYNTHETIC-ZAI\nMINIMAX_CN_KEY=SYNTHETIC-CN\n')`；L326 `selected.write_text(json.dumps(dict(live=live, effects=effects)))`；L333 `selected.write_text(json.dumps(dict(live=broken, effects=effects)))`；L343 `(self.root / '.env').write_text('ZAI_KEY=SYNTHETIC-FIRST\nZAI_KEY=SYNTHETIC-SECOND\nMINIMAX_CN_KEY=UNSELECTED-SECRET\n')`；L353 `(self.root / '.env').write_text('ZAI_KEY=SYNTHETIC-ONLY-SELECTED\nMINIMAX_CN_KEY=SYNTHETIC-NOT-SELECTED\n')`；L378 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L46 `p.mkdir(parents=True, exist_ok=True)`；L47 `(p / 'processes.json').write_text(json.dumps(self.journal, indent=2) + '\n')`；L98 `(self.root / 'ai/catalog_regional.json').write_text(json.dumps(dict(active=False)))`；L225 `proof.write_text('Synthetic independent ordinary process fixture review.\n')`；L323 `p.mkdir(parents=True)`；L323 `(p / 'live-report.json').write_text(json.dumps(dict(combo=name)))`；L325 `p.mkdir(parents=True)`；L325 `(p / ('evaluation-report.json' if name.startswith('text-') else 'report.json')).write_text('{}')`；L239 `grant.write_text(json.dumps(value))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-provider-sites--tools--test_recovery4_fixture-py"></a>

## tools/test_recovery4_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_recovery4_fixture.py](../../barness-provider-sites/tools/test_recovery4_fixture.py)；SHA256 `7d958ae68741accf0dbaab6b7c8149949df9b0951a5a269b9aff6443ade52e40`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Controlled closed3 originals and executable native contracts; no vendor sockets。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L40（import）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L64（import）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L77（import）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L93（import）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L141（import）；[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L200（import）；[.scratch/barness-provider-sites/tools/test_recovery5_fixture.py](../../barness-provider-sites/tools/test_recovery5_fixture.py) L7（import）；[.scratch/barness-provider-sites/tools/test_recovery5_fixture.py](../../barness-provider-sites/tools/test_recovery5_fixture.py) L81（import）；[.scratch/barness-provider-sites/tools/test_recovery_review_fix.py](../../barness-provider-sites/tools/test_recovery_review_fix.py) L73（import）；[.scratch/barness-provider-sites/tools/test_recovery_review_fix.py](../../barness-provider-sites/tools/test_recovery_review_fix.py) L86（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L63 `go.write_text(text.replace(repr(producer), repr(replacement)))`；L72 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L74 `actual.write_text(json.dumps(human))`；L81 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L84 `(test.root / '.env').write_text(''.join((k + '=CONTROLLED-ONLY\n' for k in ('OPENAI_KEY', 'GEMINI_KEY', 'ANTHROPIC_KEY', 'DEEPSEEK_KEY', 'TYPESAFE_KEY', 'MINIMAX_KEY', 'MINIMAX_CN_KEY', 'ZAI_KEY'))))`；L91 `stop_path.write_text(json.dumps(stop))`；L91 `seal_path.write_text(json.dumps(seal))`；L15 `p.parent.mkdir(parents=True, exist_ok=True)`；L15 `p.write_bytes(value if raw else (json.dumps(value) + '\n').encode())`；L71 `p.write_text('Controlled independent ordinary source decision.')`；L78 `dest.write_text(json.dumps(grant))`；L80 `dest.write_text(json.dumps(value))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery4_round-py"></a>

## tools/test_recovery4_round.py

原件：[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py)；SHA256 `88e1359e022e15235d7cbce53cf8b16b03bd0ee0d3d6efe100f3a13161fec1a6`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Socket-free public fixed4 process checks; never actual approvals/keys。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery_review_fix.py](../../barness-provider-sites/tools/test_recovery_review_fix.py) L6（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L29 `seal.parent.mkdir()`；L29 `seal.write_text('{}')`；L111 `jev.mkdir()`；L111 `shutil.copytree(base / 'protocol114/controlled-session', jev / 'protocol')`；L111 `shutil.copytree(base / 'visual-jev', jev / 'effects')`；L111 `(jev / 'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')))`；L112 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L113 `shutil.copytree(self.root / live['zai-responses'], forged)`；L113 `inputs.write_text(json.dumps(dict(live=bad, effects=effects)))`；L113 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L145 `report.write_text(json.dumps(v))`；L146 `claim.write_text(json.dumps(v))`；L146 `stop_path.write_text(json.dumps(stop))`；L152 `p.write_text('{}')`；L165 `selected.write_text(json.dumps(dict(live=live, effects=effects)))`；L193 `auth.write_bytes(original)`；L206 `seal.unlink()`；L206 `seal.unlink()`；L206 `seal.write_bytes(raw)`；L225 `human.write_text(json.dumps(v))`；L226 `auth.write_text(json.dumps(a))`；L235 `native.mkdir()`；L235 `shutil.copyfile(native.with_name('provider-sites-final-acceptance-3') / 'chat.json', native / 'chat.json')`；L243 `shutil.copytree(base, other)`；L162 `p.mkdir(parents=True)`；L162 `(p / 'live-report.json').write_text(json.dumps(dict(combo=name)))`；L164 `p.mkdir()`；L164 `(p / ('evaluation-report.json' if name.startswith('text-') else 'report.json')).write_text('{}')`；L192 `auth.write_text(json.dumps(v))`；L196 `human.write_text(altered)`；L205 `path.write_text('{}')`；L205 `path.write_bytes(raw)`；L208 `path.write_text('{}')`；L208 `path.unlink()`；L190 `auth.write_text(json.dumps(v))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery5_fixture-py"></a>

## tools/test_recovery5_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_recovery5_fixture.py](../../barness-provider-sites/tools/test_recovery5_fixture.py)；SHA256 `fe7944c1d492c2ce8a42c4e7063d994848851160db1f010fcab9285d73558e70`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Synthetic closed4 and executable producer contracts for public fixed5 processes。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery5_intent_io.py](../../barness-provider-sites/tools/test_recovery5_intent_io.py) L9（import）；[.scratch/barness-provider-sites/tools/test_recovery5_io.py](../../barness-provider-sites/tools/test_recovery5_io.py) L5（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L45（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L59（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L67（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L77（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L94（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L107（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L123（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L148（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L158（import）；[.scratch/barness-provider-sites/tools/test_recovery6_fixture.py](../../barness-provider-sites/tools/test_recovery6_fixture.py) L7（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L59 `(root / 'ai/internal/testkit/finalbudget/estimate.go').write_text('package corrected\n')`；L63 `verifier.write_text(verifier.read_text().replace('f5d37cc92abc7d9f4db844e73ad8e00efc63d7c316824ae43afff917280907a7', sha(root / BASE / 'decisions-budget-diagnosis-dev/compatibility-eight.json')))`；L77 `go.write_text(text.replace(repr(producer), repr(replacement)))`；L89 `actual.write_text(json.dumps(human))`；L90 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L96 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L100 `(test.root / '.env').write_text(''.join((k + '=CONTROLLED-ONLY\n' for k in ('OPENAI_KEY', 'GEMINI_KEY', 'ANTHROPIC_KEY', 'DEEPSEEK_KEY', 'TYPESAFE_KEY', 'MINIMAX_KEY', 'MINIMAX_CN_KEY', 'ZAI_KEY', 'ZHIPU_KEY'))))`；L17 `p.parent.mkdir(parents=True, exist_ok=True)`；L17 `p.write_bytes(value if raw else (json.dumps(value) + '\n').encode())`；L87 `p.write_text('CONTROLLED ordinary review')`；L93 `dest.write_text(json.dumps(v))`；L95 `dest.write_text(json.dumps(v))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery5_intent_io-py"></a>

## tools/test_recovery5_intent_io.py

原件：[.scratch/barness-provider-sites/tools/test_recovery5_intent_io.py](../../barness-provider-sites/tools/test_recovery5_intent_io.py)；SHA256 `ed5cd119fa78e0a6bea62b6c3431ee3c65597842151d1747a4353718dee40dd6`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public canonical settlement FIFO cases; no selected credential is supplied。

**调用方**：[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery5_io-py"></a>

## tools/test_recovery5_io.py

原件：[.scratch/barness-provider-sites/tools/test_recovery5_io.py](../../barness-provider-sites/tools/test_recovery5_io.py)；SHA256 `64c9522d223c2adbc3a0c40c51c6c0529c1f62376f496c6caa592a747cb7fe54`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public CLI FIFO swap regressions at selected-key, audit and settlement boundaries。

**调用方**：[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery5_round-py"></a>

## tools/test_recovery5_round.py

原件：[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py)；SHA256 `5d4cc1605830cbdb2f18d770b93570d074a10c77d7b7742c0e2cde35c6797df2`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public fixed5 CLI processes; synthetic repositories only, no vendor sockets。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery5_intent_io.py](../../barness-provider-sites/tools/test_recovery5_intent_io.py) L8（import）；[.scratch/barness-provider-sites/tools/test_recovery5_io.py](../../barness-provider-sites/tools/test_recovery5_io.py) L4（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L35 `seal.parent.mkdir()`；L35 `seal.write_text('{}')`；L69 `compat.write_text(json.dumps(v))`；L69 `compat.write_bytes(raw)`；L70 `p.write_text('{}')`；L85 `verifier.write_text(verifier.read_text().replace(BUDGET_SOURCE_COMMIT, pin))`；L86 `p.write_bytes((repo / name).read_bytes())`；L88 `p.write_bytes(p.read_bytes() + b'\n// unreviewed budget artifact source drift\n')`；L112 `auth.write_bytes(original)`；L112 `human.write_text(json.dumps(v))`；L112 `auth.write_text(json.dumps(a))`；L113 `human.write_text(raw.decode().replace('"decision": "APPROVED"', '"Decision": "PENDING", "decision": "APPROVED"'))`；L118 `p.unlink()`；L118 `p.unlink()`；L118 `p.symlink_to('plan.json')`；L118 `p.unlink()`；L118 `p.write_bytes(raw)`；L138 `jev.mkdir()`；L138 `shutil.copytree(base / 'protocol114/controlled-session', jev / 'protocol')`；L138 `shutil.copytree(base / 'visual-jev', jev / 'effects')`；L138 `(jev / 'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')))`；L142 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L62 `p.parent.mkdir(parents=True, exist_ok=True)`；L62 `p.write_text('controlled unexpected source')`；L63 `p.unlink()`；L83 `p.parent.mkdir(parents=True, exist_ok=True)`；L83 `p.write_bytes(raw)`；L111 `auth.write_text(json.dumps(v))`；L141 `shutil.copytree(source, forged)`；L141 `inputs.write_text(json.dumps(dict(live=bad, effects=effects)))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery6_fixture-py"></a>

## tools/test_recovery6_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_recovery6_fixture.py](../../barness-provider-sites/tools/test_recovery6_fixture.py)；SHA256 `51a57313190963756daecdf29a479e1be467d491cf4662f31a88c9cbd9034c2e`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Disposable closed5 originals and fixed6 executables, never real authority。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L48（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L62（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L70（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L77（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L91（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L107（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L132（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L142（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L154（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L165（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L185（import）；[.scratch/barness-provider-sites/tools/test_recovery7_fixture.py](../../barness-provider-sites/tools/test_recovery7_fixture.py) L7（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L46 `link.parent.mkdir(parents=True)`；L46 `link.symlink_to('seal.json')`；L51 `path.write_text(path.read_text().replace('5e8014e7405cfa86da21637b96e47cb294c4ff8e766e79440fe88ff56eef4f73', sha(sp)).replace('STOP_FILE_COUNT=31152', 'STOP_FILE_COUNT=' + str(len(stop['files']))))`；L52 `path.write_text(path.read_text().replace('8c8342606390d37dff041f11de7ff3b8d82d12d5', commit))`；L64 `path.write_text(path.read_text().replace('0b5c158568df468574cca2e835c41a16079bcebad03d0e7ede6bd6ea13b58306', sha(dest)))`；L73 `go.write_text(text.replace(repr(producer), repr(replacement)))`；L82 `actual.write_text(json.dumps(human))`；L82 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L88 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L14 `p.parent.mkdir(parents=True, exist_ok=True)`；L14 `p.write_text('package controlled\n')`；L18 `p.parent.mkdir(parents=True, exist_ok=True)`；L18 `p.write_bytes(value if raw else (json.dumps(value) + '\n').encode())`；L55 `p.write_text(p.read_text() + '// CONTROLLED legitimate regional repair\n')`；L80 `p.write_text('CONTROLLED review')`；L85 `dest.write_text(json.dumps(v))`；L87 `dest.write_text(json.dumps(v))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery6_round-py"></a>

## tools/test_recovery6_round.py

原件：[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py)；SHA256 `07efdb196352a39d0176184da8d837a5575b4840fba23a9b4c9d55de2ef5ced7`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public fixed6 CLI processes; synthetic repositories only, no vendor sockets。

**调用方**：[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L38 `seal.parent.mkdir()`；L38 `seal.write_text('{}')`；L72 `compat.write_text(json.dumps(v))`；L72 `compat.write_bytes(raw)`；L73 `p.write_text('{}')`；L96 `auth.write_bytes(original)`；L96 `human.write_text(json.dumps(v))`；L96 `auth.write_text(json.dumps(a))`；L97 `human.write_text(raw.decode().replace('"decision": "APPROVED"', '"Decision": "PENDING", "decision": "APPROVED"'))`；L102 `p.unlink()`；L102 `p.unlink()`；L102 `p.symlink_to('plan.json')`；L102 `p.unlink()`；L102 `p.write_bytes(raw)`；L122 `jev.mkdir()`；L122 `shutil.copytree(base / 'protocol114/controlled-session', jev / 'protocol')`；L122 `shutil.copytree(base / 'visual-jev', jev / 'effects')`；L122 `(jev / 'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')))`；L126 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L157 `stop.write_text(json.dumps(v))`；L158 `seal.write_text(json.dumps(v))`；L159 `p.write_text(p.read_text().replace(before, after))`；L65 `p.parent.mkdir(parents=True, exist_ok=True)`；L65 `p.write_text('controlled unexpected source')`；L66 `p.unlink()`；L95 `auth.write_text(json.dumps(v))`；L125 `shutil.copytree(source, forged)`；L125 `inputs.write_text(json.dumps(dict(live=bad, effects=effects)))`；L178 `proof.write_text(json.dumps(v))`；L178 `p.write_text(source.replace(old, new))`；L179 `p.write_text(source)`；L191 `p.unlink()`；L170 `wire.write_text(json.dumps(data))`；L180 `changed[0].write_bytes(changed[1])`；L188 `p.write_text('{}')`；L189 `p.symlink_to('source-at-start.json')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery7_fixture-py"></a>

## tools/test_recovery7_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_recovery7_fixture.py](../../barness-provider-sites/tools/test_recovery7_fixture.py)；SHA256 `3aa43088c0a43492b7544782b716ba97d761631ff15b74dea2bd32b7ca50e475`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Disposable interrupted sixth originals and fixed7 no-socket executables。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L58（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L74（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L82（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L89（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L103（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L119（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L147（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L157（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L169（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L180（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L200（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L211（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L220（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L44 `link.parent.mkdir(parents=True)`；L44 `link.symlink_to('seal.json')`；L50 `path.write_text(path.read_text().replace('fcb781794736eb8c8cc83582d757ed16e0db5091661690cab51f517bcdbad81e', sha(sp)).replace('STOP_FILE_COUNT=30238', 'STOP_FILE_COUNT=' + str(len(stop['files']))))`；L51 `path.write_text(path.read_text().replace('8c8342606390d37dff041f11de7ff3b8d82d12d5', baseline['sourceCommit']))`；L56 `path.write_text(path.read_text().replace('c1bdb5bd6be0f7b27ef82b043d02e0a0d64e7501b0cc88ca512d10aa9983a16d', sha(dest)))`；L66 `go.write_text(text.replace(repr(producer), repr(replacement)))`；L75 `actual.write_text(json.dumps(human))`；L75 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L81 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L14 `p.parent.mkdir(parents=True, exist_ok=True)`；L14 `p.write_text('package controlled\n')`；L18 `p.parent.mkdir(parents=True, exist_ok=True)`；L18 `p.write_bytes(value if raw else (json.dumps(value) + '\n').encode())`；L53 `p.write_text(p.read_text() + '// CONTROLLED own authorization deadline repair\n')`；L73 `p.write_text('CONTROLLED review')`；L78 `dest.write_text(json.dumps(v))`；L80 `dest.write_text(json.dumps(v))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery7_round-py"></a>

## tools/test_recovery7_round.py

原件：[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py)；SHA256 `f8fe16db81f12fb6577d0f75ed7157b40a1e1549a500e427e95a0ba254d5db08`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public fixed7 CLI processes; synthetic repositories only, no vendor sockets。

**调用方**：[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L48 `seal.parent.mkdir()`；L48 `seal.write_text('{}')`；L84 `compat.write_text(json.dumps(v))`；L84 `compat.write_bytes(raw)`；L85 `p.write_text('{}')`；L108 `auth.write_bytes(original)`；L108 `human.write_text(json.dumps(v))`；L108 `auth.write_text(json.dumps(a))`；L109 `human.write_text(raw.decode().replace('"decision": "APPROVED"', '"Decision": "PENDING", "decision": "APPROVED"'))`；L114 `p.unlink()`；L114 `p.unlink()`；L114 `p.symlink_to('plan.json')`；L114 `p.unlink()`；L114 `p.write_bytes(raw)`；L134 `jev.mkdir()`；L134 `shutil.copytree(base / 'protocol114/controlled-session', jev / 'protocol')`；L134 `shutil.copytree(base / 'visual-jev', jev / 'effects')`；L134 `(jev / 'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')))`；L138 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L172 `stop.write_text(json.dumps(v))`；L173 `seal.write_text(json.dumps(v))`；L174 `p.write_text(p.read_text().replace(before, after))`；L216 `ledger.write_text(json.dumps(v))`；L216 `ledger.write_bytes(raw)`；L222 `p.write_text(re.sub("COMPAT_SHA256='[^']+'", "COMPAT_SHA256='PENDING_CURRENT_GO_84_REPROOF'", text))`；L77 `p.parent.mkdir(parents=True, exist_ok=True)`；L77 `p.write_text('controlled unexpected source')`；L78 `p.unlink()`；L107 `auth.write_text(json.dumps(v))`；L137 `shutil.copytree(source, forged)`；L137 `inputs.write_text(json.dumps(dict(live=bad, effects=effects)))`；L193 `proof.write_text(json.dumps(v))`；L193 `p.write_text(source.replace(old, new))`；L194 `p.write_text(source)`；L206 `p.unlink()`；L215 `p.write_text('{}')`；L215 `p.unlink()`；L185 `wire.write_text(json.dumps(data))`；L195 `changed[0].write_bytes(changed[1])`；L203 `p.write_text('{}')`；L204 `p.symlink_to('source-at-start.json')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery_fixture-py"></a>

## tools/test_recovery_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_recovery_fixture.py](../../barness-provider-sites/tools/test_recovery_fixture.py)；SHA256 `fa4933bde41d14bcf74659ad82b08b5e6a8ca37ee9ca11f8f3ed7c7029bdc08f`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Socket-free synthetic234-file prefix and public producer/consumer executables。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery4_fixture.py](../../barness-provider-sites/tools/test_recovery4_fixture.py) L6（import）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L36（import）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L124（import）；[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py) L151（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L44 `(test.tools / 'final_publication.py').write_text('import final_contract as r\nc=r.c\ndef validate(index,views=None):return dict(controlled=True)\ndef require(route):\n name="glm-responses-retry-1" if route=="responses" else "glm-"+route\n p=c.common_root()/".evidence/minimax-launches"/(r.ROUND+"-"+name+".json")\n return c.doc(p),c.digest(p)\n')`；L45 `p.mkdir(parents=True)`；L45 `(p / 'live-captures.json').write_text('{}')`；L47 `(root / 'ai/live/support-matrix.json').write_text(json.dumps(dict(schema=3, rows=rows)))`；L54 `(p / 'support-views.json').write_text(json.dumps(dict(routes=views)))`；L100 `go.write_text(text.replace(repr(producer), repr(producer + extra)))`；L15 `p.parent.mkdir(parents=True, exist_ok=True)`；L15 `p.write_bytes(value if raw else (json.dumps(value) + '\n').encode())`；L50 `dest.mkdir(parents=True)`；L50 `(dest / 'live-report.json').write_bytes(raw)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery_review_fix-py"></a>

## tools/test_recovery_review_fix.py

原件：[.scratch/barness-provider-sites/tools/test_recovery_review_fix.py](../../barness-provider-sites/tools/test_recovery_review_fix.py)；SHA256 `7d17d0f9dd70c80d47b164d455feb40b5f16ac36d25e47259283a83d28e5aeee`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public CLI review regressions, isolated repositories and synthetic approvals only。

**调用方**：[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L62 `(self.root / '.env').write_text('MINIMAX_KEY=CONTROLLED-ONLY\n')`；L75 `archive.write_text(json.dumps(value))`；L96 `log.write_text(log.read_text() + ' ')`；L98 `log.write_bytes(raw)`；L59 `(self.tools / (name + '.py')).unlink(missing_ok=True)`；L39 `target.unlink()`；L50 `target.unlink()`；L50 `target.write_bytes(raw)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_recovery_round-py"></a>

## tools/test_recovery_round.py

原件：[.scratch/barness-provider-sites/tools/test_recovery_round.py](../../barness-provider-sites/tools/test_recovery_round.py)；SHA256 `5bc3630b9d9505b2a594fcb96ff79df8b28ecc1f8b00c955fe654b8abcc85c7e`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public process checks of fixed recovery3; all state is synthetic。

**调用方**：[.scratch/barness-provider-sites/tools/test_recovery4_round.py](../../barness-provider-sites/tools/test_recovery4_round.py) L6（import）；[.scratch/barness-provider-sites/tools/test_recovery5_round.py](../../barness-provider-sites/tools/test_recovery5_round.py) L5（import）；[.scratch/barness-provider-sites/tools/test_recovery6_round.py](../../barness-provider-sites/tools/test_recovery6_round.py) L5（import）；[.scratch/barness-provider-sites/tools/test_recovery7_round.py](../../barness-provider-sites/tools/test_recovery7_round.py) L5（import）；[.scratch/barness-provider-sites/tools/test_recovery_review_fix.py](../../barness-provider-sites/tools/test_recovery_review_fix.py) L5（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L9 `guard.mkdir(parents=True)`；L49 `(base / 'prerelease-source-review.json').write_text(json.dumps(review))`；L51 `actual.write_text(json.dumps(human))`；L59 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L66 `(base / 'execution-authorization.json').write_text(json.dumps(auth))`；L84 `(self.root / '.env').write_text('MINIMAX_KEY=CONTROLLED-SELECTED\n')`；L99 `(self.root / '.env').write_text(''.join((k + '=CONTROLLED-' + k + '\n' for k in {j['sourceVariable'] for j in plan['jobs'].values()})))`；L112 `jev.mkdir()`；L113 `shutil.copytree(base / 'protocol114/controlled-session', jev / 'protocol')`；L113 `shutil.copytree(base / 'visual-jev', jev / 'effects')`；L114 `(jev / 'live-report.json').write_text(json.dumps(dict(combo='typesafe-classifier-jev-1.14.0')))`；L116 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L117 `shutil.copytree(self.root / live['zai-chat'], forged)`；L118 `inputs.write_text(json.dumps(dict(live=bad, effects=effects)))`；L118 `inputs.write_text(json.dumps(dict(live=live, effects=effects)))`；L127 `victim.unlink()`；L128 `victim.unlink()`；L128 `victim.write_bytes(original)`；L129 `extra.write_text('{}')`；L129 `extra.unlink()`；L130 `p.write_text(json.dumps(value))`；L135 `(self.root / '.env').write_text('MINIMAX_KEY=CONTROLLED-ONLY\n')`；L144 `shutil.copytree(base, other)`；L146 `victim.unlink()`；L146 `victim.symlink_to('execution-authorization.json')`；L146 `victim.unlink()`；L146 `victim.write_bytes(raw)`；L147 `(self.root / 'ai/sentinel.go').write_text('package drift\n')`；L10 `(guard / name).write_text('package controlled\n')`；L48 `p.write_text('Controlled independent source decision.')`；L55 `dest.write_text(json.dumps(grant))`；L58 `dest.write_text(json.dumps(value))`；L174 `victim.unlink()`；L183 `victim.unlink()`；L184 `victim.write_bytes(raw)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-provider-sites--tools--test_retry_fixture-py"></a>

## tools/test_retry_fixture.py

原件：[.scratch/barness-provider-sites/tools/test_retry_fixture.py](../../barness-provider-sites/tools/test_retry_fixture.py)；SHA256 `f1164b65c659e5e51953cbb3388c15598d79af520c34c2984e290bf994e0874d`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Independent synthetic stopped-prefix facts; never authenticates live evidence。

**调用方**：[.scratch/barness-provider-sites/tools/test_final_retry.py](../../barness-provider-sites/tools/test_final_retry.py) L5（import）；[.scratch/barness-provider-sites/tools](../../barness-provider-sites/tools) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L16 `p.parent.mkdir(parents=True, exist_ok=True)`；L16 `p.write_bytes(b'[]\n')`；L17 `p.parent.mkdir(parents=True, exist_ok=True)`；L17 `p.write_text('package supportmatrix\nconst GLMSourcePin = "' + t.pin + '"\n')`；L20 `(t.root / 'ai/live/testdata/glm-formal/live-captures.sha256').write_text('unpublished\n')`；L32 `formal.write_text('#!' + sys.executable + '\n' + scanner)`；L32 `formal.chmod(493)`；L41 `live.write_text('#!' + sys.executable + '\n' + producer)`；L41 `live.chmod(493)`；L51 `(t.root / '.evidence/minimax-renewals' / ROUND / 'launch.lock').write_bytes(b'')`；L63 `(t.root / '.evidence/minimax-launches' / (ROUND + '-glm-messages.json')).unlink()`；L63 `shutil.rmtree(old / 'glm-messages')`；L10 `path.parent.mkdir(parents=True, exist_ok=True)`；L10 `path.write_text(json.dumps(value))`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-provider-sites/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。
