# barness-ai-openai-decisions 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [candidate-live-evidence/live-authorized-20261013/launcher-used.py](#barness-ai-openai-decisions--candidate-live-evidence--live-authorized-20261013--launcher-used-py) | 历史源码 | 仅原轮次 |
| [chinese-live-evidence/launcher-process.py](#barness-ai-openai-decisions--chinese-live-evidence--launcher-process-py) | 测试 | 仅原轮次 |
| [chinese-live-evidence/run.py](#barness-ai-openai-decisions--chinese-live-evidence--run-py) | 历史源码 | 仅原轮次 |
| [classifier-load-evidence/pack.py](#barness-ai-openai-decisions--classifier-load-evidence--pack-py) | 历史源码 | 仅原轮次 |
| [formal-support-evidence/launcher.py](#barness-ai-openai-decisions--formal-support-evidence--launcher-py) | 历史源码 | 仅原轮次 |
| [jev-baseline/replay.py](#barness-ai-openai-decisions--jev-baseline--replay-py) | 历史源码 | 仅原轮次 |
| [jev-baseline/verify.py](#barness-ai-openai-decisions--jev-baseline--verify-py) | 活动工具 | 现有公共消费者 |
| [mixed-answer-evidence/verify.py](#barness-ai-openai-decisions--mixed-answer-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [release-evidence/collect.py](#barness-ai-openai-decisions--release-evidence--collect-py) | 历史源码 | 仅原轮次 |
| [release-evidence/pack.py](#barness-ai-openai-decisions--release-evidence--pack-py) | 历史源码 | 仅原轮次 |
| [release-evidence/run-gate.py](#barness-ai-openai-decisions--release-evidence--run-gate-py) | 历史源码 | 仅原轮次 |
| [release-evidence/verify.py](#barness-ai-openai-decisions--release-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [runtime-evidence/pack.py](#barness-ai-openai-decisions--runtime-evidence--pack-py) | 历史源码 | 仅原轮次 |
| [text-choice-evidence/verify.py](#barness-ai-openai-decisions--text-choice-evidence--verify-py) | 活动工具 | 现有公共消费者 |
| [ticket-review/originals/ticket-review/verify.py](#barness-ai-openai-decisions--ticket-review--originals--ticket-review--verify-py) | 历史源码 | 封存源码快照 |
| [ticket-review/verify.py](#barness-ai-openai-decisions--ticket-review--verify-py) | 历史源码 | 仅原轮次 |

<a id="barness-ai-openai-decisions--candidate-live-evidence--live-authorized-20261013--launcher-used-py"></a>

## candidate-live-evidence/live-authorized-20261013/launcher-used.py

原件：[.scratch/barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013/launcher-used.py](../../barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013/launcher-used.py)；SHA256 `3962e3080b365c4e1fc0cc69f4b005d8c60d3d7848706fd82d0f7f87eb1c845f`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：launcher-used.py：已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L14 `base.mkdir(parents=True, exist_ok=True)`；L21 `Path('/tmp/barness-decisions-authorized-live.jsonl').write_text(output)`；L23 `Path('/tmp/barness-decisions-authorized-live-receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L21 `Path('/tmp/barness-decisions-authorized-live.jsonl').write_text(output)`；L23 `Path('/tmp/barness-decisions-authorized-live-receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')`；L24 `print(json.dumps({'exitCode':r.returncode,'log':'/tmp/barness-decisions-authorized-live.jsonl'}))`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `(root / '.env').read_text().splitlines()`；L5 `(root / '.env').read_text()`；L14 `base.mkdir(parents=True, exist_ok=True)`；L19 `subprocess.run(cmd, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=210)`；L20 `r.stdout.decode().replace(key, '[REDACTED]').replace(str(Path.home()), '<host-home>')`。

<a id="barness-ai-openai-decisions--chinese-live-evidence--launcher-process-py"></a>

## chinese-live-evidence/launcher-process.py

原件：[.scratch/barness-ai-openai-decisions/chinese-live-evidence/launcher-process.py](../../barness-ai-openai-decisions/chinese-live-evidence/launcher-process.py)；SHA256 `1c212c15f96c3b2ed0470716f12c7fdb709645f20d0ee359a5e41972f08ffa53`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Public launcher regression: rejected transport secrets must never persist。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L56 `(base / ('launcher-' + phase + '.json')).write_text(json.dumps(report, indent=2) + '\n')`；L27 `shutil.copyfile(root / 'ai/live/support-matrix.json', sandbox / 'ai/live/support-matrix.json')`；L29 `target.parent.mkdir(parents=True)`；L30 `target.write_bytes((base / 'run.py').read_bytes())`；L26 `source.extractall(sandbox)`；L51 `shutil.rmtree(sandbox / '.evidence', ignore_errors=True)`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/chinese-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`；L29 `target.parent.mkdir(parents=True)`；L30 `target.write_bytes((base / 'run.py').read_bytes())`；L32 `subprocess.run(['git', 'init', '-q', str(sandbox)], check=True)`；L33 `subprocess.run(['git', '-C', str(sandbox), '-c', 'user.name=Process Fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '--allow-empty', '-qm', 'fixture'], check=True)`。

<a id="barness-ai-openai-decisions--chinese-live-evidence--run-py"></a>

## chinese-live-evidence/run.py

原件：[.scratch/barness-ai-openai-decisions/chinese-live-evidence/run.py](../../barness-ai-openai-decisions/chinese-live-evidence/run.py)；SHA256 `dc83cee7291ae27d5202a1a607543af6567f4ba7d7655d7cf9dae267fbf971a1`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Isolate Luna credentials and preserve a fresh receipt for every execution。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--live', action='store_true', help='enable up to 24 paid Luna calls')` L44；`parser.add_argument('--proxy', type=safe_proxy, default=None, help='explicit loopback HTTP proxy')` L45；`parser.add_argument('--account-alias', default='local-openai@region-unreported')` L46。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L53 `execution.mkdir(parents=True, exist_ok=False, mode=448)`；L59 `(execution / 'build.log').write_bytes(build_output)`；L117 `(execution / 'execution.log').write_bytes(output)`；L126 `(execution / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/chinese-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-openai-decisions--classifier-load-evidence--pack-py"></a>

## classifier-load-evidence/pack.py

原件：[.scratch/barness-ai-openai-decisions/classifier-load-evidence/pack.py](../../barness-ai-openai-decisions/classifier-load-evidence/pack.py)；SHA256 `49026d1b0e6422bc4b0e545c2a74d454014dd4aecc1e4977f175fac4441343a1`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Archive already-executed issue 08 runs and audit the resulting bundle。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('--executions', type=Path, required=True)` L30。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L19 `path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')`；L106 `(output / 'SHA256SUMS').write_text(''.join((f'{digest((output / name).read_bytes())}  {name}\n' for name in names)))`；L84 `(root / 'logs').mkdir()`；L88 `write_json(root / 'manifest.json', {'issue': '08', 'baseline': 'c4cc495', 'runs': summaries, 'measurements': measurements, 'executions': validation, 'live': 'NOT_RUN', 'releaseGate': 'NOT_RUN'})`；L101 `write_json(output / 'manifest.json', {'issue': '08', 'archive': archive.name, 'sha256': digest(archive.read_bytes()), 'runs': summaries, 'measurements': measurements, 'members': members})`；L80 `destination.parent.mkdir(parents=True, exist_ok=True)`；L87 `(root / 'logs' / path.name).write_text(raw)`；L48 `shutil.copytree(manifest.parent, destination)`；L51 `destination.mkdir(parents=True)`；L56 `shutil.copytree(manifest.parent / case['dir'], destination / case['dir'])`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/classifier-load-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-openai-decisions--formal-support-evidence--launcher-py"></a>

## formal-support-evidence/launcher.py

原件：[.scratch/barness-ai-openai-decisions/formal-support-evidence/launcher.py](../../barness-ai-openai-decisions/formal-support-evidence/launcher.py)；SHA256 `95cc5ded69859c8432a4d9be734842c097fca91f1073548cc691734bdff363d6`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：launcher.py：已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L21 `base.mkdir(parents=True, exist_ok=True)`；L24 `source_dir.mkdir(exist_ok=True)`；L38 `(out / 'execution-source-snapshot.json').write_text(json.dumps(source_hashes, indent=2) + '\n')`；L47 `(out / 'live-execution.jsonl').write_text(log)`；L54 `(out / 'live-receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L34 `dest.parent.mkdir(parents=True, exist_ok=True)`；L36 `dest.write_bytes(raw)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/formal-support-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `(root / '.env').read_text().splitlines()`；L10 `(root / '.env').read_text()`；L21 `base.mkdir(parents=True, exist_ok=True)`；L24 `source_dir.mkdir(exist_ok=True)`；L25 `(root / 'ai').glob('*.go')`。

<a id="barness-ai-openai-decisions--jev-baseline--replay-py"></a>

## jev-baseline/replay.py

原件：[.scratch/barness-ai-openai-decisions/jev-baseline/replay.py](../../barness-ai-openai-decisions/jev-baseline/replay.py)；SHA256 `bf9ba47a2848064735562622db440f129a8104cfd2d84d87432c529c25197c0d`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Execute current Jev regressions into a new directory; never reseal history。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--out', type=Path, required=True)` L27。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L30 `out.mkdir(parents=True, exist_ok=False)`；L34 `(out / 'inputs.json').write_text(json.dumps(identity, indent=2) + '\n')`；L48 `(out / 'commands.json').write_text(json.dumps(records, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/jev-baseline`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-ai-openai-decisions--jev-baseline--verify-py"></a>

## jev-baseline/verify.py

原件：[.scratch/barness-ai-openai-decisions/jev-baseline/verify.py](../../barness-ai-openai-decisions/jev-baseline/verify.py)；SHA256 `4936382c94dbd53dd77a6f1a914a73233fe6c4ebc0f12e56ffcc7141a6623e40`。活动工具 / 现有公共消费者。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：Read-only verification of the fixed Jev evidence; no provider calls。

**调用方**：[ai/e2e/jev_baseline_evidence_test.go](../../../ai/e2e/jev_baseline_evidence_test.go) L16（confirmed-public-Go-consumer）。

**输入**：--bundle中的固定manifest及ZIP；支持可选--audit；原manifest/归档/场景/历史pin不变；当前固定输入位于ai/e2e/testdata/retained/jev-baseline。参数：`parser.add_argument('--bundle', type=Path, default=Path(__file__).resolve().parent)` L132；`parser.add_argument('--audit', action='store_true', help='also rerun the existing whole-bundle redaction audit')` L133。

**输出/写入**：stdout核验/拒绝与数字退出；--audit在临时目录解出有界副本及审计结果。实际静态写点：L143 `target.parent.mkdir(parents=True, exist_ok=True)`；L144 `target.write_bytes(bundle.read(member))`。

**副作用**：默认只读原包；--audit会写自有临时副本并go run审计器，可能构建；不执行归档中捕获命令或Node oracle，不访问厂商。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[ai/e2e/jev_baseline_evidence_test.go](../../../ai/e2e/jev_baseline_evidence_test.go)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/jev-baseline`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-openai-decisions--mixed-answer-evidence--verify-py"></a>

## mixed-answer-evidence/verify.py

原件：[.scratch/barness-ai-openai-decisions/mixed-answer-evidence/verify.py](../../barness-ai-openai-decisions/mixed-answer-evidence/verify.py)；SHA256 `29a2086e886292c46e3e9d0cfde690ae4be04883cc58dbe409a9dde88a57fd7f`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify fixed Decisions mixed-answer evidence without accessing any Provider。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--bundle', type=Path, default=Path(__file__).resolve().parent)` L110；`parser.add_argument('--audit', action='store_true')` L111。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L121 `target.parent.mkdir(parents=True, exist_ok=True)`；L122 `target.write_bytes(bundle.read(item))`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/mixed-answer-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-openai-decisions--release-evidence--collect-py"></a>

## release-evidence/collect.py

原件：[.scratch/barness-ai-openai-decisions/release-evidence/collect.py](../../barness-ai-openai-decisions/release-evidence/collect.py)；SHA256 `c3fb84fcb00e5c978c5c86a85d7ecfac7cae626b3d3a92e6c9df64d3afc88ab9`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Collect current live evidence in isolated processes; never overwrite a run。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--out', type=Path, required=True)` L27；`p.add_argument('--live', action='store_true')` L28。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L34 `out.mkdir(parents=True, mode=448)`；L67 `folder.mkdir(parents=True)`；L101 `(folder / 'execution.log').write_bytes(output)`；L116 `(folder / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L118 `(out / 'collection.json').write_text(json.dumps(results, indent=2) + '\n')`；L115 `(folder / 'merge.log').write_bytes(merge.stdout.replace(str(ROOT).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`。

<a id="barness-ai-openai-decisions--release-evidence--pack-py"></a>

## release-evidence/pack.py

原件：[.scratch/barness-ai-openai-decisions/release-evidence/pack.py](../../barness-ai-openai-decisions/release-evidence/pack.py)；SHA256 `3220b758d42facbcf179b1f707c9a092dd796efd4471e0f8e57b8b8a6f872972`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Freeze full ten-route evidence and failure history without relabeling runs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`p.add_argument('--gate', type=Path)` L80；`p.add_argument('--collection', type=Path)` L81；`p.add_argument('--seal', action='store_true')` L82。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L29 `(FOLDER / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L134 `shutil.copyfile(collection, FOLDER / 'collection.json')`；L135 `shutil.copyfile(collection.parent / 'execution-source.zip', FOLDER / 'live-execution-source.zip')`；L137 `logs.mkdir(exist_ok=True)`；L171 `(FOLDER / 'known-pressure-blocker.md').write_bytes(blocker.read_bytes())`；L133 `(FOLDER / name).write_bytes(redact((gate / name).read_bytes()))`；L139 `(logs / path.name).write_bytes(redact(path.read_bytes()))`；L145 `(logs / (record['route'] + '-' + name)).write_bytes(redact(path.read_bytes()))`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="barness-ai-openai-decisions--release-evidence--run-gate-py"></a>

## release-evidence/run-gate.py

原件：[.scratch/barness-ai-openai-decisions/release-evidence/run-gate.py](../../barness-ai-openai-decisions/release-evidence/run-gate.py)；SHA256 `25c9eebde21057e4b084c68de2a36c23b24011102d21e3bc93a5922dd438c417`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Run the actual full gate using a completed isolated evidence collection。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`p.add_argument('--collection', type=Path, required=True)` L11；`p.add_argument('--out', type=Path, required=True)` L12。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L23 `a.out.mkdir(parents=True)`；L24 `(a.out / 'command.json').write_text(json.dumps(command, indent=2) + '\n')`；L32 `(a.out / 'exit.json').write_text(json.dumps({'exitCode': code}) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`；L13 `p.parse_args()`；L15 `a.out.resolve().is_relative_to(root)`；L15 `a.out.resolve()`；L17 `json.loads(a.collection.read_text())`。

<a id="barness-ai-openai-decisions--release-evidence--verify-py"></a>

## release-evidence/verify.py

原件：[.scratch/barness-ai-openai-decisions/release-evidence/verify.py](../../barness-ai-openai-decisions/release-evidence/verify.py)；SHA256 `c8fe330e1914d6a37f7e41338de2fad32f235616aea0d475e9d6c96b57f20fd6`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Check delivered hashes and optionally replay artifact gates without vendor calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`p.add_argument('--replay', action='store_true')` L54。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L107 `(FOLDER / 'verification.json').write_text(json.dumps(result, indent=2) + '\n')`；L47 `target.parent.mkdir(parents=True, exist_ok=True)`；L45 `target.mkdir(parents=True, exist_ok=True)`；L49 `shutil.copyfileobj(source, output, 1 << 20)`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`。

<a id="barness-ai-openai-decisions--runtime-evidence--pack-py"></a>

## runtime-evidence/pack.py

原件：[.scratch/barness-ai-openai-decisions/runtime-evidence/pack.py](../../barness-ai-openai-decisions/runtime-evidence/pack.py)；SHA256 `29fb6ac39f67dd8c81f14c69828f2cdc570b19ac5ba6392dbe7be79a6bc1c2e5`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Package issue 06's existing executions; never manufacture a test verdict。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('--validation', type=Path, required=True)` L24；`parser.add_argument('--history', type=Path, required=True)` L25；`parser.add_argument('--race', type=Path, required=True)` L26；`parser.add_argument('--full', type=Path, required=True)` L27；`parser.add_argument('--log', type=Path, action='append', default=[])` L28。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L19 `path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L34 `write_json(output / 'validation.json', validation)`；L114 `(output / 'SHA256SUMS').write_text(''.join((f'{digest((output / name).read_bytes())}  {name}\n' for name in names)))`；L76 `write_json(root / 'validation.json', validation)`；L78 `logs.mkdir()`；L95 `write_json(root / 'manifest.json', {'issue': '06', 'baseline': baseline, 'runs': summaries, 'live': 'NOT_RUN', 'observerAuditViews': observer_views, 'decisionsParity': 'NOT_RUN', 'pressure': 'NOT_RUN'})`；L109 `write_json(output / 'manifest.json', {'issue': '06', 'archive': archive.name, 'sha256': digest(archive.read_bytes()), 'runs': summaries, 'members': members})`；L72 `destination.parent.mkdir(parents=True, exist_ok=True)`；L82 `(logs / path.name).write_text(path.read_text().replace(str(repo), '<workspace>').replace(str(Path.home()), '<home>'))`；L91 `alias.write_bytes(raw)`；L47 `destination.mkdir(parents=True)`；L58 `shutil.copytree(manifest.parent, destination)`；L55 `shutil.copytree(manifest.parent / name, destination / name)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/runtime-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-openai-decisions--text-choice-evidence--verify-py"></a>

## text-choice-evidence/verify.py

原件：[.scratch/barness-ai-openai-decisions/text-choice-evidence/verify.py](../../barness-ai-openai-decisions/text-choice-evidence/verify.py)；SHA256 `700a85f360432e57cb35e5c9ae37b7f892fa5777e3aa706d967af1e89c8d63cf`。活动工具 / 现有公共消费者。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：Verify fixed Decisions text evidence without accessing any Provider。

**调用方**：[ai/e2e/decisions_evidence_test.go](../../../ai/e2e/decisions_evidence_test.go) L13（confirmed-public-Go-consumer）。

**输入**：--bundle中的固定manifest及ZIP；支持可选--audit；原manifest/归档/场景/历史pin不变；当前固定输入位于ai/e2e/testdata/retained/decisions-text-choice。参数：`parser.add_argument('--bundle', type=Path, default=Path(__file__).resolve().parent)` L110；`parser.add_argument('--audit', action='store_true')` L111。

**输出/写入**：stdout核验/拒绝与数字退出；--audit在临时目录解出有界副本及审计结果。实际静态写点：L121 `target.parent.mkdir(parents=True, exist_ok=True)`；L122 `target.write_bytes(bundle.read(item))`。

**副作用**：默认只读原包；--audit会写自有临时副本并go run审计器，可能构建；不执行归档中捕获命令或Node oracle，不访问厂商。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[ai/e2e/decisions_evidence_test.go](../../../ai/e2e/decisions_evidence_test.go)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/text-choice-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-openai-decisions--ticket-review--originals--ticket-review--verify-py"></a>

## ticket-review/originals/ticket-review/verify.py

原件：[.scratch/barness-ai-openai-decisions/ticket-review/originals/ticket-review/verify.py](../../barness-ai-openai-decisions/ticket-review/originals/ticket-review/verify.py)；SHA256 `7a1c08e9bd66ac7fa4e8a34543618ee5581ee24a82241738c807a662e7642d02`。历史源码 / 封存源码快照。原文档审批核对，包含封存源码副本。

**职责**：Validate the proposed ticket documents; never runs product tests。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原工单Markdown及审批出处/固定文档集合；originals副本仍按原路径语义读取。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：同目录validation.json。实际静态写点：L129 `(HERE / 'validation.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：读取文档并覆盖validation.json；不跑产品测试、不授予审批；封存副本不作现行文档校验器。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/ticket-review/originals/ticket-review`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `Path(__file__).resolve()`；L11 `json.loads((HERE / 'proposal.json').read_text())`；L11 `(HERE / 'proposal.json').read_text()`；L21 `(FEATURE / 'spec.md').read_text()`；L22 `(FEATURE / 'acceptance.md').read_text()`。

<a id="barness-ai-openai-decisions--ticket-review--verify-py"></a>

## ticket-review/verify.py

原件：[.scratch/barness-ai-openai-decisions/ticket-review/verify.py](../../barness-ai-openai-decisions/ticket-review/verify.py)；SHA256 `ef5b777db62f60e3f1ab685f513d0f46ac9610e11fd2f46539f11794ce84afbc`。历史源码 / 仅原轮次。原文档审批核对，包含封存源码副本。

**职责**：Verify published ticket documents and approval provenance; no product tests。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原工单Markdown及审批出处/固定文档集合；originals副本仍按原路径语义读取。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：同目录validation.json。实际静态写点：L196 `(HERE / 'validation.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：读取文档并覆盖validation.json；不跑产品测试、不授予审批；封存副本不作现行文档校验器。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-openai-decisions/ticket-review`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `Path(__file__).resolve()`；L11 `json.loads((HERE / 'proposal.json').read_text())`；L11 `(HERE / 'proposal.json').read_text()`；L12 `json.loads((HERE / 'publication.json').read_text())`；L12 `(HERE / 'publication.json').read_text()`。
