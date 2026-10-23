# barness-ai-test-process-review 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [reviews/20261022-section1-rereview-repros/consumer-final-write.py](#barness-ai-test-process-review--reviews--20261022-section1-rereview-repros--consumer-final-write-py) | 诊断装置 | 仅原轮次 |
| [reviews/20261022-section1-rereview-repros/relative-dependency.py](#barness-ai-test-process-review--reviews--20261022-section1-rereview-repros--relative-dependency-py) | 诊断装置 | 仅原轮次 |
| [reviews/20261022-section1-rereview-repros/zero-call-chat.py](#barness-ai-test-process-review--reviews--20261022-section1-rereview-repros--zero-call-chat-py) | 诊断装置 | 仅原轮次 |
| [tools/cleanup.py](#barness-ai-test-process-review--tools--cleanup-py) | 活动工具 | 现行SOP |
| [tools/cleanup_scan.py](#barness-ai-test-process-review--tools--cleanup_scan-py) | 活动工具 | 现行SOP |
| [tools/cleanup_state.py](#barness-ai-test-process-review--tools--cleanup_state-py) | 活动工具 | 现行SOP |
| [tools/consume_live.py](#barness-ai-test-process-review--tools--consume_live-py) | 活动工具 | 现行SOP |
| [tools/live_environment.py](#barness-ai-test-process-review--tools--live_environment-py) | 活动工具 | 现行SOP |
| [tools/receipt_files.py](#barness-ai-test-process-review--tools--receipt_files-py) | 活动工具 | 现行SOP |
| [tools/run.py](#barness-ai-test-process-review--tools--run-py) | 活动工具 | 现行SOP |
| [tools/run_process.py](#barness-ai-test-process-review--tools--run_process-py) | 活动工具 | 现行SOP |
| [tools/tests/fixture_child.py](#barness-ai-test-process-review--tools--tests--fixture_child-py) | 测试 | 显式合成测试 |
| [tools/tests/live_fixture.py](#barness-ai-test-process-review--tools--tests--live_fixture-py) | 测试 | 显式合成测试 |
| [tools/tests/process_support.py](#barness-ai-test-process-review--tools--tests--process_support-py) | 测试 | 显式合成测试 |
| [tools/tests/test_cleanup.py](#barness-ai-test-process-review--tools--tests--test_cleanup-py) | 测试 | 显式合成测试 |
| [tools/tests/test_live_release.py](#barness-ai-test-process-review--tools--tests--test_live_release-py) | 测试 | 显式合成测试 |
| [tools/tests/test_ordinary.py](#barness-ai-test-process-review--tools--tests--test_ordinary-py) | 测试 | 显式合成测试 |
| [tools/tests/test_release.py](#barness-ai-test-process-review--tools--tests--test_release-py) | 测试 | 显式合成测试 |
| [tools/tests/test_run.py](#barness-ai-test-process-review--tools--tests--test_run-py) | 测试 | 显式合成测试 |

<a id="barness-ai-test-process-review--reviews--20261022-section1-rereview-repros--consumer-final-write-py"></a>

## reviews/20261022-section1-rereview-repros/consumer-final-write.py

原件：[.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/consumer-final-write.py](../../barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/consumer-final-write.py)；SHA256 `04377b6d59f0d3b78ccee4fdf77a8ccaed40fd3af6f909f8df4aa5ac7391e61d`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：consumer-final-write.py：原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L17 `(ART / 'io-failure-actual.stdout').write_text(result.stdout)`；L17 `(ART / 'io-failure-actual.stderr').write_text(result.stderr)`；L18 `write_json(ART / 'io-failure-summary.json', summary)`；L19 `shutil.copytree(bad, ART / 'io-failure-saved', dirs_exist_ok=True)`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L2 `ROOT=pathlib.Path('/Users/cyber/RestoX/barness'); ART=pathlib.Path('/tmp/barness-section1-spec-rereview')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `(ART / 'io-failure-actual.stdout').write_text(result.stdout)`；L17 `(ART / 'io-failure-actual.stderr').write_text(result.stderr)`；L18 `write_json(ART / 'io-failure-summary.json', summary)`；L16 `(bad / 'stderr.log').read_text()`；L20 `json.dumps(summary)`。

<a id="barness-ai-test-process-review--reviews--20261022-section1-rereview-repros--relative-dependency-py"></a>

## reviews/20261022-section1-rereview-repros/relative-dependency.py

原件：[.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/relative-dependency.py](../../barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/relative-dependency.py)；SHA256 `b62ecbee4e58eccde12b29bc79adb47fe1390a0b93c4a9b0d3b6c3a47799a285`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：relative-dependency.py：原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L10 `write_json(failure, {'status': 'FAIL', 'exitCode': 7, 'input': '../unused'})`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L3 `sys.path.insert(0, '/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `write_json(failure, {'status': 'FAIL', 'exitCode': 7, 'input': '../unused'})`；L12 `json.loads((case.meta / 'plan.json').read_text())`；L13 `json.loads((case.meta / 'snapshot.json').read_text())`；L12 `(case.meta / 'plan.json').read_text()`；L13 `(case.meta / 'snapshot.json').read_text()`。

<a id="barness-ai-test-process-review--reviews--20261022-section1-rereview-repros--zero-call-chat-py"></a>

## reviews/20261022-section1-rereview-repros/zero-call-chat.py

原件：[.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/zero-call-chat.py](../../barness-ai-test-process-review/reviews/20261022-section1-rereview-repros/zero-call-chat.py)；SHA256 `328d229825b2c9b580ab73b680177728a42404fb5e4ad9700087cb60b166c458`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：zero-call-chat.py：原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定失败模型及私有源/索引/文件系统副本，部分构造器/probe包含合成外部程序。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：私有fixture、探针日志与有限断言；如处在raw目录则仅为封存源码。实际静态写点：L21 `write_json(bundle / 'manifest.json', manifest)`；L21 `write_json(bundle / 'live-report.json', report)`；L24 `write_json(receiptpath, receipt)`；L30 `write_json(ART / 'zero-call-chat-summary.json', summary)`；L31 `shutil.copytree(t.repo, ART / 'zero-call-chat-repo', ignore=shutil.ignore_patterns('.git'), dirs_exist_ok=True)`；L17 `write_json(bundle / ident / 'assertions.json', {'case_id': ident, 'status': 'PASS', 'assertions': [{'pass': True}]})`；L14 `shutil.rmtree(c)`。

**副作用**：可能启动公共CLI/构造器、修改/删除私有文件或建立FIFO；只用于原问题的诊断，不是通用维护入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/reviews/20261022-section1-rereview-repros`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L2 `ROOT=pathlib.Path('/Users/cyber/RestoX/barness');ART=pathlib.Path('/tmp/barness-section1-spec-rereview')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `json.loads((bundle / 'manifest.json').read_text())`；L21 `write_json(bundle / 'manifest.json', manifest)`；L21 `write_json(bundle / 'live-report.json', report)`；L22 `json.loads(receiptpath.read_text())`；L24 `write_json(receiptpath, receipt)`。

<a id="barness-ai-test-process-review--tools--cleanup-py"></a>

## tools/cleanup.py

原件：[.scratch/barness-ai-test-process-review/tools/cleanup.py](../../barness-ai-test-process-review/tools/cleanup.py)；SHA256 `79a878af47fdb18577791eb2eb2efde92650413722badbd53446b084fd990651`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Maintain an explicitly reviewed evidence batch; no tests or vendor execution。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests/test_cleanup.py](../../barness-ai-test-process-review/tools/tests/test_cleanup.py) L19（literal-reference）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：明确审阅的repo/batch/plan-sha256、引用根/候选库存、锁与批次状态。参数：`parser.add_argument('action', choices=('scan', 'plan', 'move', 'verify', 'restore', 'inspect', 'purge'))` L163；`parser.add_argument('--repo', type=Path, default=Path.cwd())` L164；`parser.add_argument('--batch')` L165；`parser.add_argument('--plan-sha256')` L166。

**输出/写入**：scan/plan/移动恢复日志、verify/inspect结果与永久metadata；purge仅限该批payload。实际静态写点：L22 `meta.mkdir()`；L28 `(meta / name).write_text(''.join((json.dumps(row, sort_keys=True) + '\n' for row in rows)))`；L150 `shutil.rmtree(payload)`；L66 `destination.parent.mkdir(parents=True, exist_ok=True)`；L68 `original.rename(destination)`；L116 `original.parent.mkdir(parents=True, exist_ok=True)`；L118 `payload.rename(original)`。

**副作用**：scan/plan会写批次；move/restore重命名真实文件；purge可永久删除。需独立审阅、verify和至少7天观察；本项真实维护NOT_RUN。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--cleanup_scan-py"></a>

## tools/cleanup_scan.py

原件：[.scratch/barness-ai-test-process-review/tools/cleanup_scan.py](../../barness-ai-test-process-review/tools/cleanup_scan.py)；SHA256 `022f286151f783129938e33524647a2b3f48251fe87e1caaedabc44f3b31b286`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Conservative reference inventory for repository-local evidence maintenance。

**调用方**：[.scratch/barness-ai-test-process-review/tools/cleanup.py](../../barness-ai-test-process-review/tools/cleanup.py) L12（import）；[.scratch/barness-ai-test-process-review/tools/cleanup_state.py](../../barness-ai-test-process-review/tools/cleanup_state.py) L9（import）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：明确审阅的repo/batch/plan-sha256、引用根/候选库存、锁与批次状态。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：scan/plan/移动恢复日志、verify/inspect结果与永久metadata；purge仅限该批payload。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：scan/plan会写批次；move/restore重命名真实文件；purge可永久删除。需独立审阅、verify和至少7天观察；本项真实维护NOT_RUN。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--cleanup_state-py"></a>

## tools/cleanup_state.py

原件：[.scratch/barness-ai-test-process-review/tools/cleanup_state.py](../../barness-ai-test-process-review/tools/cleanup_state.py)；SHA256 `dfc5d81a49a6eb5e2c4be5f9a472c37b6d9abdb05eab578e36f19b0564160ca0`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Durable batch identity, inventory checks and append-only movement receipts。

**调用方**：[.scratch/barness-ai-test-process-review/tools/cleanup.py](../../barness-ai-test-process-review/tools/cleanup.py) L13（import）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：明确审阅的repo/batch/plan-sha256、引用根/候选库存、锁与批次状态。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：scan/plan/移动恢复日志、verify/inspect结果与永久metadata；purge仅限该批payload。实际静态写点：L21 `parent.mkdir(parents=True, exist_ok=True)`；L31 `path.unlink()`。

**副作用**：scan/plan会写批次；move/restore重命名真实文件；purge可永久删除。需独立审阅、verify和至少7天观察；本项真实维护NOT_RUN。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--consume_live-py"></a>

## tools/consume_live.py

原件：[.scratch/barness-ai-test-process-review/tools/consume_live.py](../../barness-ai-test-process-review/tools/consume_live.py)；SHA256 `4bc231fe6ffa78acc86f751bb1839e69e94d519d5d71bde8856cdb6d5580a674`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Capture one read-only native live consumer; never run a verification stage。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L10（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L69（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L195（literal-reference）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：已编译releasegate的--binary、原生--bundle、全新--out和有界timeout。参数：`parser.add_argument('--binary', required=True)` L15；`parser.add_argument('--bundle', required=True)` L16；`parser.add_argument('--out', required=True)` L17；`parser.add_argument('--timeout', type=float, default=120)` L18。

**输出/写入**：新消费者目录的schema2 live-consumer.json、native资格原件、argv/时间/真实退出和日志。实际静态写点：L31 `out.mkdir(mode=448)`；L45 `(out / 'exit-code').write_text(str(receipt['exitCode']) + '\n')`。

**副作用**：执行一次只读原生资格子进程并写新目录；无厂商生产请求，不写支持矩阵；仍须原生完整资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--live_environment-py"></a>

## tools/live_environment.py

原件：[.scratch/barness-ai-test-process-review/tools/live_environment.py](../../barness-ai-test-process-review/tools/live_environment.py)；SHA256 `3f7e149cd2b1241d2ca8f7d6302b7d1c9c655d51775848000b570b9ebc380fec`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Forward the selected native producer's contract without acquiring authority。

**调用方**：[.scratch/barness-ai-test-process-review/tools/run_process.py](../../barness-ai-test-process-review/tools/run_process.py) L10（import）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：恰好一条已选原生路线及对应环境变量；原生grant/ledger资格另行验证。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：单路线转发环境与用于脱敏的敏感值集合；不落盘秘密。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：读取环境中所选凭据/路线字段，返回给run_process；不自行请求厂商或发放权限。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[.scratch/barness-ai-test-process-review/tools/README.md](../../barness-ai-test-process-review/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--receipt_files-py"></a>

## tools/receipt_files.py

原件：[.scratch/barness-ai-test-process-review/tools/receipt_files.py](../../barness-ai-test-process-review/tools/receipt_files.py)；SHA256 `de2f290b03f0d8864bae02b50f5fea2d7568020c30c7d0b1c21551cfbff687f3`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Bounded receipt IO and execution-source snapshots, independent of test policy。

**调用方**：[.scratch/barness-ai-test-process-review/tools/cleanup.py](../../barness-ai-test-process-review/tools/cleanup.py) L15（import）；[.scratch/barness-ai-test-process-review/tools/cleanup_scan.py](../../barness-ai-test-process-review/tools/cleanup_scan.py) L9（import）；[.scratch/barness-ai-test-process-review/tools/cleanup_state.py](../../barness-ai-test-process-review/tools/cleanup_state.py) L10（import）；[.scratch/barness-ai-test-process-review/tools/consume_live.py](../../barness-ai-test-process-review/tools/consume_live.py) L9（import）；[.scratch/barness-ai-test-process-review/tools/run.py](../../barness-ai-test-process-review/tools/run.py) L16（import）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：仓库根、来源选择器、选定路径及原生bundle；调用方传入的JSON/文件。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：有界读写、原子JSON、完整源码快照、文件/包身份与核验结果。实际静态写点：L54 `os.replace(temporary, path)`；L57 `temporary.unlink(missing_ok=True)`。

**副作用**：读Git/文件元数据；save写临时文件、fsync和replace；模块API并非独立操作者CLI。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[.scratch/barness-ai-test-process-review/tools/README.md](../../barness-ai-test-process-review/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--run-py"></a>

## tools/run.py

原件：[.scratch/barness-ai-test-process-review/tools/run.py](../../barness-ai-test-process-review/tools/run.py)；SHA256 `4ef759de59672cf6cffa1280ff488f31eb5ee0055a92cf178a0565ba97fedff6`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Capture one explicitly selected verification process; never schedule another。

**调用方**：[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：人选定的单条argv、阶段/路线、实际--input文件、源码和允许环境；凭据不进入argv/输入快照。参数：`parser.add_argument('--feature', required=True)` L22；`parser.add_argument('--stage', choices=('e2e', 'race', 'pressure', 'live'), required=True)` L23；`parser.add_argument('--route')` L24；`parser.add_argument('--timeout', type=float)` L25；`parser.add_argument('--input', action='append', default=[], help='selected repository-relative input path')` L26；`parser.add_argument('command', nargs=argparse.REMAINDER)` L27。

**输出/写入**：全新.evidence/<feature>/<stage>/<attempt>；命令/日志/退出/源码与输入哈希/原生包指针。实际静态写点：L61 `base.mkdir(parents=True, exist_ok=True)`；L65 `bundles.mkdir()`；L88 `(out / 'exit-code').write_text(str(receipt['exitCode']) + '\n')`。

**副作用**：创建attempt并执行所选子进程；超时/取消终止本进程组；live可转发单路线凭据，重负载必须另选；子命令可构建/联网，工具不是授权。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--run_process-py"></a>

## tools/run_process.py

原件：[.scratch/barness-ai-test-process-review/tools/run_process.py](../../barness-ai-test-process-review/tools/run_process.py)；SHA256 `9ebd8fec669871721c184ab4f0c141ca98efcebb3e68b8e215cc3a2c83cd0295`。活动工具 / 现行SOP。现行SOP明确使用的工具或其内部模块；清理仍需独立范围。

**职责**：Bounded process capture with group cancellation and credential redaction。

**调用方**：[.scratch/barness-ai-test-process-review/tools/consume_live.py](../../barness-ai-test-process-review/tools/consume_live.py) L10（import）；[.scratch/barness-ai-test-process-review/tools/run.py](../../barness-ai-test-process-review/tools/run.py) L17（import）；[docs/agents/verification.md](../../../docs/agents/verification.md) L1（current-operational-contract）。

**输入**：父工具提供的argv/cwd/阶段/路线、环境与超时、日志输出目录。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：白名单环境、隔离HOME、脱敏stdout/stderr/退出及采集结果。实际静态写点：L46 `home.mkdir(mode=448)`。

**副作用**：读取允许环境；创建隔离目录、启动/取消子进程组、写日志；能力受实际子命令约束，不能推断任意命令无网络。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[.scratch/barness-ai-test-process-review/tools/README.md](../../barness-ai-test-process-review/tools/README.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--tests--fixture_child-py"></a>

## tools/tests/fixture_child.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/fixture_child.py](../../barness-ai-test-process-review/tools/tests/fixture_child.py)；SHA256 `397198e84415aecb745611aca5d6341c7b54fa01882fecd778c2d53175b967f4`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：A synthetic external process producing native-shape bundles, never Go/live。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests/process_support.py](../../barness-ai-test-process-review/tools/tests/process_support.py) L67（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L54（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L59（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L128（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L167（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L187（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L192（literal-reference）；[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L21 `case.mkdir(parents=True)`；L22 `(case / 'assertions.json').write_text('{"assertions":[{"passed":true}]}\n')`；L32 `(bundle / 'manifest.json').write_text(json.dumps(manifest))`；L43 `Path('policy.json').write_text('{"limit":99}')`；L46 `Path('ai/README.md').write_text('Prose only; no execution input changed.\n')`；L49 `(base.parent / '.DS_Store').write_bytes(b'synthetic-desktop-cache')`；L50 `Path('ai/.DS_Store').write_bytes(b'synthetic-after')`；L62 `Path('descendant.pid').write_text(str(child.pid))`；L34 `(bundle / 'audit.json').write_text('{"findings":[]}' if mode != 'bad-audit' else '{"findings":[{"rule":"synthetic"}]}')`；L36 `(case / 'assertions.json').write_text('{}')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L21 `case.mkdir(parents=True)`；L22 `(case / 'assertions.json').write_text('{"assertions":[{"passed":true}]}\n')`；L32 `(bundle / 'manifest.json').write_text(json.dumps(manifest))`；L32 `json.dumps(manifest)`；L34 `(bundle / 'audit.json').write_text('{"findings":[]}' if mode != 'bad-audit' else '{"findings":[{"rule":"synthetic"}]}')`。

<a id="barness-ai-test-process-review--tools--tests--live_fixture-py"></a>

## tools/tests/live_fixture.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/live_fixture.py](../../barness-ai-test-process-review/tools/tests/live_fixture.py)；SHA256 `98f8b94dced655984810b2b510026bdda5b697ee658b41b7426250874e24c078`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Synthetic protocol documents for the read-only CLI seam; no vendor requests。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L4（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L85（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L125（import）；[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L43 `write_json(bundle / 'live-report.json', report)`；L44 `write_json(bundle / 'manifest.json', {'package': package, 'git_commit': test.source['commit'], 'started_at': stamp, 'finished_at': stamp, 'versions': {'live_combo': combo, 'live_sdk': 'direct HTTP', 'model_catalog_version': report['catalogVersion'], 'model_catalog_hash': report['catalogHash']}, 'cases': cases})`；L56 `write_json(command_path, command)`；L58 `write_json(path, receipt)`；L98 `write_json(bundle / 'live-report.json', report)`；L101 `write_json(bundle / 'manifest.json', manifest)`；L112 `write_json(bundle / 'manifest.json', manifest)`；L116 `write_json(path, receipt)`；L27 `write_json(case / 'assertions.json', {'case_id': ident, 'status': 'PASS', 'assertions': [{'pass': True}]})`；L87 `write_json(case / 'attempt-1-call-result.json', {'error': '', 'result': {'metadata': metadata, 'message': {'usage': {'input': 1, 'output': 1}}}})`；L89 `write_json(case / 'attempt-1-call-exchanges.json', [{'method': 'POST', 'url': endpoint + '/models/' + model + ':streamGenerateContent', 'requestBody': '{}', 'status': 200, 'responseHeaders': {}, 'responseBody': '{}'}])`；L92 `write_json(case / 'observations-attempt-1-call.json', [{'kind': 'call_finished', 'call': metadata}])`；L93 `write_json(case / 'assertions.json', {'case_id': ident, 'status': 'PASS', 'assertions': [{'pass': True}]})`；L73 `shutil.rmtree(path)`；L33 `write_json(case / filename, [] if filename.startswith('observations') else {'synthetic': True})`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--tests--process_support-py"></a>

## tools/tests/process_support.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/process_support.py](../../barness-ai-test-process-review/tools/tests/process_support.py)；SHA256 `3d26b11420f442f8d72191ec24c42142497bc5af517c7704a972ed145da84572`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：Public process fixtures; synthetic repositories never use retained evidence。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests/live_fixture.py](../../barness-ai-test-process-review/tools/tests/live_fixture.py) L3（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_cleanup.py](../../barness-ai-test-process-review/tools/tests/test_cleanup.py) L4（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L3（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L16（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L60（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L86（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L126（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L137（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L148（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L192（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L202（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_ordinary.py](../../barness-ai-test-process-review/tools/tests/test_ordinary.py) L8（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_release.py](../../barness-ai-test-process-review/tools/tests/test_release.py) L9（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py) L3（import）；[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L21 `path.parent.mkdir(parents=True, exist_ok=True)`；L22 `path.write_text(json.dumps(value, sort_keys=True, indent=2) + '\n')`；L37 `(self.repo / 'ai').mkdir()`；L38 `(self.repo / 'ai/example.go').write_text('package ai\n')`；L39 `(self.repo / 'go.mod').write_text('module example.test/fixture\n\ngo 1.26.2\n')`；L40 `(self.repo / 'state.md').write_text('Synthetic current status\n')`；L53 `target.mkdir(parents=True, exist_ok=True)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`。

<a id="barness-ai-test-process-review--tools--tests--test_cleanup-py"></a>

## tools/tests/test_cleanup.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/test_cleanup.py](../../barness-ai-test-process-review/tools/tests/test_cleanup.py)；SHA256 `49828cc68ca87c30c9a6f503f7b78ab57562ece8efa4f7d9e28b4e388386e267`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：test_cleanup.py：现有显式Python合成进程测试/支持模块，未因本次标注执行。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L10 `(self.repo / '.gitignore').write_text('/.evidence/\n')`；L13 `write_json(self.repo / '.evidence/minimax-retained/receipt.json', {'status': 'PASS', 'exitCode': 0})`；L14 `(self.repo / '.evidence/unused').mkdir(parents=True)`；L15 `(self.repo / '.evidence/unused/body.txt').write_text('synthetic body\n')`；L16 `write_json(self.repo / '.evidence/unused/receipt.json', {'status': 'PASS', 'exitCode': 0, 'files': {'body.txt': digest(self.repo / '.evidence/unused/body.txt')}})`；L38 `write_json(path, state)`；L42 `(base / 'tdd').mkdir(parents=True)`；L43 `(base / 'tdd/final.log').write_text('linked\n')`；L44 `(base / 'handoff.md').write_text('[log](tdd/final.log)\n[run](../../../.evidence/used/index.json)\n')`；L45 `write_json(self.repo / '.evidence/used/index.json', {'path': '../indirect/receipt.json'})`；L46 `write_json(self.repo / '.evidence/indirect/receipt.json', {'status': 'PASS', 'exitCode': 0})`；L53 `(self.repo / 'notes.md').write_text('Historical `.evidence/old/missing.json`.\nExample `.evidence/NEW_RELEASE_CHECK`.\n')`；L61 `(self.repo / 'new.md').write_text('Retain `.evidence/unused/body.txt`.\n')`；L69 `(self.repo / 'new.md').write_text('Restore `.evidence/unused/body.txt`.\n')`；L78 `write_json(self.repo / '.evidence/unused/receipt.json', {'new': 'different'})`；L92 `lock.parent.mkdir(parents=True)`；L93 `lock.write_text('owned\n')`；L95 `lock.unlink()`；L97 `required.unlink()`；L102 `write_json(required, {'path': '../missing-current/index.json'})`；L126 `write_json(self.repo / '.evidence/unindexed-failure/receipt.json', {'status': 'FAIL', 'exitCode': 7, 'input': '../dependency/index.json'})`；L128 `write_json(self.repo / '.evidence/dependency/receipt.json', {'status': 'PASS', 'exitCode': 0, 'files': {}})`；L131 `write_json(dependency, {'input': '../transitive/receipt.json'})`；L132 `write_json(self.repo / '.evidence/dependency/receipt.json', {'status': 'PASS', 'exitCode': 0, 'files': {'index.json': digest(dependency)}})`；L134 `write_json(self.repo / '.evidence/transitive/receipt.json', {'status': 'PASS', 'exitCode': 0})`；L163 `write_json(self.repo / '.evidence/unindexed-failure/receipt.json', {'status': 'FAIL', 'exitCode': 7, 'inputs': ['../unused', '../binary/payload.bin', '../plain/payload']})`；L172 `dependency.write_text('[next](../transitive)\n')`；L173 `write_json(self.repo / '.evidence/disposable/receipt.json', {'status': 'PASS', 'exitCode': 0})`；L174 `write_json(self.repo / '.evidence/unused/receipt.json', {'status': 'PASS', 'exitCode': 0, 'files': {'body.txt': digest(dependency)}})`；L176 `write_json(self.repo / '.evidence/transitive/receipt.json', {'status': 'PASS', 'exitCode': 0})`；L12 `write_json(self.repo / '.evidence/provider-sites/final-acceptance-20261020' / p, {'synthetic': True})`；L109 `write_json(base / 'good/receipt.json', {'status': 'PASS', 'exitCode': 0})`；L167 `base.mkdir()`；L168 `(base / artifact).write_bytes(b'synthetic capture')`；L169 `write_json(base / 'receipt.json', {'status': 'PASS', 'exitCode': 0, 'files': {artifact: digest(base / artifact)}})`；L111 `write_json(base / 'old/manifest.json', {'cases': [{'id': 'original-failure', 'status': 'FAIL'}]})`；L113 `(base / 'unclassified.bin').write_bytes(b'unknown payload')`；L115 `write_json(base / 'native/manifest.json', {'cases': [{'id': 'case', 'status': 'PASS', 'artifacts': {}}]})`；L117 `write_json(base / 'grant.json', {'id': 'synthetic-authority', 'status': 'closed'})`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--tests--test_live_release-py"></a>

## tools/tests/test_live_release.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py)；SHA256 `72d5823f75580c5708efe44993a1cc721431365b66f8c2fe3407d90e42aa6c83`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：test_live_release.py：现有显式Python合成进程测试/支持模块，未因本次标注执行。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L24 `launcher.parent.mkdir()`；L25 `launcher.write_text('#!' + sys.executable + '\nimport os,resource,signal,sys\nsignal.signal(signal.SIGXFSZ, signal.SIG_IGN)\nresource.setrlimit(resource.RLIMIT_FSIZE, (' + str(limit) + ',' + str(limit) + '))\nos.execv(' + repr(str(self.binary)) + ', [' + repr(str(self.binary)) + '] + sys.argv[1:])\n')`；L29 `launcher.chmod(448)`；L175 `write_json(proof_path, original)`；L178 `write_json(bundle / 'extra.json', {'changed': True})`；L225 `alias.symlink_to(bundle, target_is_directory=True)`；L115 `write_json(bundle / 'live-report.json', report)`；L172 `write_json(proof_path, changed)`；L212 `write_json(bundle / 'live-report.json', report)`；L104 `(case / 'attempt-1-call-result.json').unlink()`；L105 `write_json(case / 'attempt-1-call-exchanges.json', [])`；L109 `write_json(p, value)`；L110 `write_json(case / 'observations-attempt-1-call.json', [])`；L113 `p.write_bytes((case / p.name).read_bytes())`；L97 `p.unlink()`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--tests--test_ordinary-py"></a>

## tools/tests/test_ordinary.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/test_ordinary.py](../../barness-ai-test-process-review/tools/tests/test_ordinary.py)；SHA256 `51a9d61e28f7e01d36423b19654ce5f68e7a3a050f0a2d703b1567fcf25b47cc`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：test_ordinary.py：现有显式Python合成进程测试/支持模块，未因本次标注执行。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L17 `(root / 'go').write_text('#!/bin/sh\ncase " $* " in *TestMiniMaxSharedDesignLoad*) echo hidden-load > ' + shlex.quote(str(root / 'started')) + '; exit 91;; *) exec ' + shlex.quote(shutil.which('go')) + ' "$@";; esac\n')`；L18 `(root / 'go').chmod(493)`；L77 `(Path(artifact) / 'home-redaction-attempt.txt').write_text(str(receipt_path.relative_to(PROJECT)) + '\n')`；L32 `dest.mkdir(parents=True, exist_ok=True)`；L33 `(dest / 'ordinary.stdout.log').write_text(result.stdout)`；L34 `(dest / 'ordinary.stderr.log').write_text(result.stderr)`；L35 `(dest / 'ordinary.exit-code').write_text(str(result.returncode) + '\n')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--tests--test_release-py"></a>

## tools/tests/test_release.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/test_release.py](../../barness-ai-test-process-review/tools/tests/test_release.py)；SHA256 `fabd649abcbbfe6291744324e24c199973b2e39c225f32585f68454c268ac4c6`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：test_release.py：现有显式Python合成进程测试/支持模块，未因本次标注执行。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests/live_fixture.py](../../barness-ai-test-process-review/tools/tests/live_fixture.py) L4（import）；[.scratch/barness-ai-test-process-review/tools/tests/test_live_release.py](../../barness-ai-test-process-review/tools/tests/test_live_release.py) L5（import）；[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L46 `write_json(bundle / 'fixture-case/assertions.json', {'assertions': [{'passed': True}]})`；L47 `write_json(bundle / 'manifest.json', {'package': './ai/e2e', 'git_commit': self.source['commit'], 'cases': [{'id': 'fixture-case', 'test': 'TestFixture', 'status': 'PASS', 'dir': 'fixture-case', 'artifacts': {'assertions.json': digest(bundle / 'fixture-case/assertions.json')}}]})`；L50 `write_json(bundle / 'audit.json', {'findings': []})`；L55 `write_json(attempt / 'command.json', {'toolchain': toolchain, 'schema': 1, 'argv': argv, 'stage': stage, 'route': None, 'timeoutSeconds': 1500, 'environmentNames': ['PATH']})`；L57 `(attempt / 'stdout.jsonl').write_text('\n'.join((json.dumps(x) for x in [{'Action': 'run', 'Package': 'example.test/fixture/ai/e2e', 'Test': 'TestFixture'}, {'Action': 'pass', 'Package': 'example.test/fixture/ai/e2e', 'Test': 'TestFixture'}, {'Action': 'pass', 'Package': 'example.test/fixture/ai/e2e'}])) + '\n')`；L61 `(attempt / 'stderr.log').write_text('')`；L62 `(attempt / 'exit-code').write_text('0\n')`；L70 `write_json(attempt / 'receipt.json', receipt)`；L78 `write_json(self.record_path, self.record)`；L89 `bin_dir.mkdir()`；L91 `go.write_text('#!/bin/sh\necho forbidden > started-verification\nexit 91\n')`；L92 `go.chmod(493)`；L114 `target.write_text('{}')`；L116 `(self.repo / 'ai/example.go').write_text('package changed\n')`；L125 `write_json(self.record_path, self.record)`；L127 `self.record_path.write_text(raw)`；L131 `(attempt / 'stderr.log').write_text('WARNING: DATA RACE\n')`；L135 `write_json(receipt_path, receipt)`；L146 `path.write_text(saved)`；L153 `write_json(path, receipt)`；L170 `write_json(command_path, command)`；L172 `write_json(path, receipt)`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-test-process-review--tools--tests--test_run-py"></a>

## tools/tests/test_run.py

原件：[.scratch/barness-ai-test-process-review/tools/tests/test_run.py](../../barness-ai-test-process-review/tools/tests/test_run.py)；SHA256 `e591d0f598e5ff05b925b95ef95b68a5aa54afd3cf430313fe0f648e336219a0`。测试 / 显式合成测试。现有显式Python合成进程测试/支持模块，未因本次标注执行。

**职责**：test_run.py：现有显式Python合成进程测试/支持模块，未因本次标注执行。

**调用方**：[.scratch/barness-ai-test-process-review/tools/tests](../../barness-ai-test-process-review/tools/tests) L1（explicit-Python-test-selection; not automatic Go suite）。

**输入**：测试自行建立的私有临时仓库、合成key/授权/捕获、假外部程序及失败注入；测试支持模块由测试加载。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：unittest/进程断言、临时子进程日志和合成回执；不构成真实资格。实际静态写点：L52 `(self.repo / 'policy.json').write_text('{"limit": 4}\n')`；L69 `(self.repo / '.gitignore').write_text('ai/hidden.go\n')`；L70 `(self.repo / 'ai/hidden.go').write_text('package ai\n// ignored code is still compiled\n')`；L81 `(self.repo / 'ai/cancel_test.go').write_text('package ai\nimport("testing";"time")\nfunc TestControlledCancellation(t *testing.T){time.Sleep(15*time.Second)}\n')`；L138 `(self.repo / 'ai/.DS_Store').write_bytes(b'synthetic-before')`；L149 `config.parent.mkdir(parents=True)`；L150 `config.write_text('GOFLAGS=-race\n')`。

**副作用**：创建/修改/删除私有fixture，启动受控子进程；部分测试会编译Go或做本地Git操作。不是全库Go默认自动执行项，需显式选定测试入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-test-process-review/tools/tests`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。
