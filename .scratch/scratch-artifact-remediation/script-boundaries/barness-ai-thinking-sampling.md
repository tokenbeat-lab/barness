# barness-ai-thinking-sampling 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [fixture-cases.py](#barness-ai-thinking-sampling--fixture-cases-py) | 历史源码 | 仅原轮次 |
| [pressure-stream-integrity-evidence/native-causality/reproduce.py](#barness-ai-thinking-sampling--pressure-stream-integrity-evidence--native-causality--reproduce-py) | 诊断装置 | 仅原轮次 |
| [pressure-stream-integrity-evidence/native-causality/tool-red.py](#barness-ai-thinking-sampling--pressure-stream-integrity-evidence--native-causality--tool-red-py) | 诊断装置 | 仅原轮次 |
| [pressure-stream-integrity-evidence/replay-red.py](#barness-ai-thinking-sampling--pressure-stream-integrity-evidence--replay-red-py) | 诊断装置 | 仅原轮次 |
| [pressure-stream-integrity-evidence/verify.py](#barness-ai-thinking-sampling--pressure-stream-integrity-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [release-evidence/pack.py](#barness-ai-thinking-sampling--release-evidence--pack-py) | 历史源码 | 仅原轮次 |
| [release-evidence/run-evaluation.py](#barness-ai-thinking-sampling--release-evidence--run-evaluation-py) | 历史源码 | 仅原轮次 |
| [release-evidence/run-live.py](#barness-ai-thinking-sampling--release-evidence--run-live-py) | 历史源码 | 仅原轮次 |
| [release-evidence/run-release.py](#barness-ai-thinking-sampling--release-evidence--run-release-py) | 历史源码 | 仅原轮次 |
| [release-evidence/run-smokes.py](#barness-ai-thinking-sampling--release-evidence--run-smokes-py) | 历史源码 | 仅原轮次 |
| [release-evidence/verify.py](#barness-ai-thinking-sampling--release-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [release-verification/check.mjs](#barness-ai-thinking-sampling--release-verification--check-mjs) | 历史源码 | 仅原轮次 |
| [release-verification/failures.mjs](#barness-ai-thinking-sampling--release-verification--failures-mjs) | 测试 | 仅原轮次 |
| [release-verification/faults.mjs](#barness-ai-thinking-sampling--release-verification--faults-mjs) | 测试 | 仅原轮次 |
| [release-verification/run.mjs](#barness-ai-thinking-sampling--release-verification--run-mjs) | 历史源码 | 仅原轮次 |
| [update-ledger.py](#barness-ai-thinking-sampling--update-ledger-py) | 历史源码 | 仅原轮次 |

<a id="barness-ai-thinking-sampling--fixture-cases-py"></a>

## fixture-cases.py

原件：[.scratch/barness-ai-thinking-sampling/fixture-cases.py](../../barness-ai-thinking-sampling/fixture-cases.py)；SHA256 `febb7673123d742adac220e1609d933752c580e0b522832a43bac273c44a0037`。历史源码 / 仅原轮次。一次性夹具作者程序，顶层覆盖写入，无当前调用方。

**职责**：Synthetic public Client fixtures. Expected profile keys are literal worked examples from acceptance.md, never obtained by importing the implementation. Event scripts/result semantics reuse existing protocol acceptance fixtures。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：固定相对ai/e2e/testdata中的基础fixture、acceptance worked examples与脚本内字面期望。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：responses/chat/deepseek-chat/deepseek-responses下的thinking-sampling.json。实际静态写点：L130 `(root / route / 'thinking-sampling.json').write_text('{\n"case":' + json.dumps(out['case']) + ',\n"source":' + json.dumps(out['source']) + ',\n"scenarios":[\n' + ',\n'.join((json.dumps(s, ensure_ascii=False, separators=(',', ':')) for s in scenarios)) + '\n]}\n')`。

**副作用**：顶层直接读取并覆盖4个fixture，无main guard/显式out/覆盖保护；import也会写。只留历史，不运行修补现行夹具。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L130 `(root / route / 'thinking-sampling.json').write_text('{\n"case":' + json.dumps(out['case']) + ',\n"source":' + json.dumps(out['source']) + ',\n"scenarios":[\n' + ',\n'.join((json.dumps(s, ensure_ascii=False, separators=(',', ':')) for s in scenarios)) + '\n]}\n')`；L11 `json.loads((root / route / 'text-basic.json').read_text())`；L11 `(root / route / 'text-basic.json').read_text()`；L14 `json.loads((root / route / 'text.json').read_text())`；L130 `json.dumps(out['source'])`。

<a id="barness-ai-thinking-sampling--pressure-stream-integrity-evidence--native-causality--reproduce-py"></a>

## pressure-stream-integrity-evidence/native-causality/reproduce.py

原件：[.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality/reproduce.py](../../barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality/reproduce.py)；SHA256 `a457579d323403e002e35f3da1344d5afd3a5c4c83605b069915e93582a1850e`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Diagnostic-only historical retry overlay. No injected socket errors or gate verdicts。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原source-base/pin、历史重试实现/overlay、可选负载/输出参数和受控OS资源。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：独立诊断目录的overlay/补丁副本/编译物/自然反例/日志/receipt。实际静态写点：L9 `dest.mkdir(mode=448)`；L11 `(dest / 'condition').mkdir()`；L144 `overlay.write_text(json.dumps({'Replace': replaces}))`；L147 `(dest / 'build.log').write_bytes(build.stdout + build.stderr)`；L162 `(dest / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L172 `(dest / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L136 `target.write_text(patched)`；L140 `(dest / (source.name + '.txt')).write_text(patched)`；L152 `ambient_source.write_bytes((Path(__file__).parent / 'ambient.go.txt').read_bytes())`；L155 `(dest / 'ambient-source.go.txt').write_bytes(ambient_source.read_bytes())`。

**副作用**：可编译运行临时Go overlay、开loopback socket并产生压力/资源负载；历史600-Go等来源守卫限制适用性，不能当当前门禁或正式厂商证据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`；L9 `dest.mkdir(mode=448)`；L11 `(dest / 'condition').mkdir()`；L15 `write.read_text().replace('import (', 'import (\n"errors"\n"syscall"\n"time"\n"os"', 1)`；L15 `write.read_text()`。

<a id="barness-ai-thinking-sampling--pressure-stream-integrity-evidence--native-causality--tool-red-py"></a>

## pressure-stream-integrity-evidence/native-causality/tool-red.py

原件：[.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality/tool-red.py](../../barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality/tool-red.py)；SHA256 `d60c47dbe34d734abd9eb154d930f82e7670854be5eb5930c921d383057bcc00`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Replay the removed retry through a temporary Go overlay; never edit runtime source。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原source-base/pin、历史重试实现/overlay、可选负载/输出参数和受控OS资源。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：独立诊断目录的overlay/补丁副本/编译物/自然反例/日志/receipt。实际静态写点：L37 `base.mkdir(mode=448, exist_ok=False)`；L38 `(base / 'overlaid-write-trace.go.txt').write_text(patched)`；L65 `(base / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L44 `overlay_source.write_text(patched)`；L46 `overlay.write_text(json.dumps({'Replace': {str(source): str(overlay_source)}}))`；L53 `(base / 'execution.log').write_bytes(output)`。

**副作用**：可编译运行临时Go overlay、开loopback socket并产生压力/资源负载；历史600-Go等来源守卫限制适用性，不能当当前门禁或正式厂商证据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/native-causality`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`；L8 `source.read_text()`；L35 `patched.replace('import (', 'import (\n"errors"\n"syscall"\n"time"', 1)`；L37 `base.mkdir(mode=448, exist_ok=False)`；L38 `(base / 'overlaid-write-trace.go.txt').write_text(patched)`。

<a id="barness-ai-thinking-sampling--pressure-stream-integrity-evidence--replay-red-py"></a>

## pressure-stream-integrity-evidence/replay-red.py

原件：[.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/replay-red.py](../../barness-ai-thinking-sampling/pressure-stream-integrity-evidence/replay-red.py)；SHA256 `1f12cf2ec7534430f1123ef0e6b3b5d0504acb530203875e4e9fb6c3477033fd`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Replay the removed retry through a temporary Go overlay; never edit runtime source。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原source-base/pin、历史重试实现/overlay、可选负载/输出参数和受控OS资源。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：独立诊断目录的overlay/补丁副本/编译物/自然反例/日志/receipt。实际静态写点：L37 `base.mkdir(mode=448, exist_ok=False)`；L38 `(base / 'overlaid-write-trace.go.txt').write_text(patched)`；L65 `(base / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L44 `overlay_source.write_text(patched)`；L46 `overlay.write_text(json.dumps({'Replace': {str(source): str(overlay_source)}}))`；L53 `(base / 'execution.log').write_bytes(output)`。

**副作用**：可编译运行临时Go overlay、开loopback socket并产生压力/资源负载；历史600-Go等来源守卫限制适用性，不能当当前门禁或正式厂商证据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L6 `Path(__file__).resolve()`；L8 `source.read_text()`；L35 `patched.replace('import (', 'import (\n"errors"\n"syscall"\n"time"', 1)`；L37 `base.mkdir(mode=448, exist_ok=False)`；L38 `(base / 'overlaid-write-trace.go.txt').write_text(patched)`。

<a id="barness-ai-thinking-sampling--pressure-stream-integrity-evidence--verify-py"></a>

## pressure-stream-integrity-evidence/verify.py

原件：[.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/verify.py](../../barness-ai-thinking-sampling/pressure-stream-integrity-evidence/verify.py)；SHA256 `fee12422c882755e0931cc216ffb192b6b7b0d04f28d9b76a0f751108b773bc0`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify preserved pressure evidence, including FAIL runs, without network。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `Path(__file__).resolve()`。

<a id="barness-ai-thinking-sampling--release-evidence--pack-py"></a>

## release-evidence/pack.py

原件：[.scratch/barness-ai-thinking-sampling/release-evidence/pack.py](../../barness-ai-thinking-sampling/release-evidence/pack.py)；SHA256 `11a271f78bb8b770ab635ad626c51c03063f8e78c4197f6e640a3eac012f8639`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Freeze complete bundles, including failures; never change test artifacts。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L117 `shutil.copyfile(ROOT / '.evidence/barness-ai-thinking-sampling/pressure-reproduction/command.json', FOLDER / 'pressure-reproduction-command.json')`；L122 `shutil.copyfile(ROOT / '.evidence/barness-ai-thinking-sampling/host-example.go', FOLDER / 'host-example.go.txt')`；L124 `logs.mkdir(exist_ok=True)`；L32 `(FOLDER / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L119 `(FOLDER / name).write_bytes(redact_paths((CURRENT / 'gate' / name).read_bytes()))`；L121 `shutil.copyfile(CURRENT / name, FOLDER / name)`；L134 `(logs / (str(i) + '-' + path.name)).write_bytes(data)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`；L15 `Path(__file__).resolve()`；L119 `(FOLDER / name).write_bytes(redact_paths((CURRENT / 'gate' / name).read_bytes()))`；L119 `(CURRENT / 'gate' / name).read_bytes()`；L124 `logs.mkdir(exist_ok=True)`。

<a id="barness-ai-thinking-sampling--release-evidence--run-evaluation-py"></a>

## release-evidence/run-evaluation.py

原件：[.scratch/barness-ai-thinking-sampling/release-evidence/run-evaluation.py](../../barness-ai-thinking-sampling/release-evidence/run-evaluation.py)；SHA256 `fa74e3b05b18c620e0b94d59ed8d023dc5f5391679967a1cd032b82386bc2ccd`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Run the independent Chinese evaluation with one isolated TypeSafe key。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--live', action='store_true', help='enable paid real evaluation')` L13；`parser.add_argument('--out', type=Path, default=Path('.evidence/barness-ai/thinking-sampling-evaluation'))` L14。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L18 `log_base.mkdir(parents=True, exist_ok=True)`；L60 `log.write_bytes(output)`；L47 `log.write_bytes(diagnostics)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L15 `parser.parse_args()`；L16 `Path(__file__).resolve()`；L18 `log_base.mkdir(parents=True, exist_ok=True)`；L22 `(root / '.env').open()`；L24 `line.strip().removeprefix('export ').partition('=')`。

<a id="barness-ai-thinking-sampling--release-evidence--run-live-py"></a>

## release-evidence/run-live.py

原件：[.scratch/barness-ai-thinking-sampling/release-evidence/run-live.py](../../barness-ai-thinking-sampling/release-evidence/run-live.py)；SHA256 `4a0288115fa6238512d1e0e968ddcfc32ec93f4827aa6fbb9e0e0af15fac05d6`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：One explicit combination per process; the release gate never handles keys。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--combo', required=True, choices=['openai-responses', 'openai-chat', 'anthropic-messages', 'google-gemini', 'deepseek-responses', 'deepseek-chat', 'typesafe-classifier', 'openai-images', 'google-interactions-image'])` L12；`parser.add_argument('--live', action='store_true')` L13；`parser.add_argument('--binary', type=Path, required=True, help='go test -c -tags live -o <binary> ./ai/live')` L14；`parser.add_argument('--out', type=Path, required=True)` L15；`parser.add_argument('--alias', required=True, help='host-confirmed account/region alias, or explicit unconfirmed alias')` L16。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L42 `args.out.mkdir(parents=True, exist_ok=True)`；L49 `(args.out / 'run.log').write_bytes(output)`；L52 `(args.out / 'command.json').write_text(json.dumps(result, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `parser.parse_args()`；L18 `args.out.resolve()`；L19 `Path(__file__).resolve()`；L28 `(root / '.env').open()`；L30 `line.strip().removeprefix('export ').partition('=')`。

<a id="barness-ai-thinking-sampling--release-evidence--run-release-py"></a>

## release-evidence/run-release.py

原件：[.scratch/barness-ai-thinking-sampling/release-evidence/run-release.py](../../barness-ai-thinking-sampling/release-evidence/run-release.py)；SHA256 `874a3e8a6ce0d6b8f464bf5644b4ad6a9b3a8a23965d3fff74abfeef3b048edf`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Re-run the complete gate, with live calls only when explicitly enabled。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--live', action='store_true', help='enable the nine bounded smokes and independent evaluation')` L12；`parser.add_argument('--out', type=Path, required=True)` L13。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L20 `out.mkdir(parents=True)`；L34 `(out / 'gate-command.json').write_text(json.dumps({'argv': command, 'commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()}, indent=2) + '\n')`；L37 `(out / 'gate-exit.json').write_text(json.dumps({'exitCode': result.returncode}, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`；L10 `Path(__file__).resolve()`；L14 `parser.parse_args()`；L15 `args.out.resolve()`；L20 `out.mkdir(parents=True)`。

<a id="barness-ai-thinking-sampling--release-evidence--run-smokes-py"></a>

## release-evidence/run-smokes.py

原件：[.scratch/barness-ai-thinking-sampling/release-evidence/run-smokes.py](../../barness-ai-thinking-sampling/release-evidence/run-smokes.py)；SHA256 `65e7b725c7742268a21278a438b507992f20272d93a1f47ff980ca610cd526c4`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Nine isolated live processes. Each report is merged only into its own row。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--live', action='store_true')` L7；`parser.add_argument('--out', type=Path, required=True)` L8；`parser.add_argument('--binary', type=Path, required=True)` L9。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L14 `out.mkdir(parents=True, exist_ok=True)`；L29 `(out / 'smoke-commands.json').write_text(json.dumps(results, indent=2) + '\n')`；L27 `(directory / 'merge.log').write_bytes((merge.stdout + merge.stderr).replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `Path(__file__).resolve()`；L10 `parser.parse_args()`；L11 `args.out.resolve()`；L14 `out.mkdir(parents=True, exist_ok=True)`；L21 `subprocess.run(command, cwd=root, env=env)`。

<a id="barness-ai-thinking-sampling--release-evidence--verify-py"></a>

## release-evidence/verify.py

原件：[.scratch/barness-ai-thinking-sampling/release-evidence/verify.py](../../barness-ai-thinking-sampling/release-evidence/verify.py)；SHA256 `37fdce8dc39de94f1bc959e3daec8d340e788330148e92f605b438985c8c8061`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify full archives and replay the recorded gates without vendor calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--fault-tests', action='store_true')` L214；`parser.add_argument('--fault-tests-only', action='store_true')` L215。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L306 `(FOLDER / 'verification.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')`；L101 `tar.extractall(work, filter='data')`；L129 `marker.write_text('{"verificationScope":"whole-archive-tree","originalRunUnchanged":true}\n')`；L194 `file.write_text('initial')`；L196 `file.write_text('changed')`；L203 `file.unlink()`；L135 `marker.unlink()`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `Path(__file__).resolve()`；L18 `Path(__file__).resolve()`；L216 `parser.parse_args()`；L220 `json.dumps({'rejected': fault_results})`；L223 `FOLDER.rglob('*')`。

<a id="barness-ai-thinking-sampling--release-verification--check-mjs"></a>

## release-verification/check.mjs

原件：[.scratch/barness-ai-thinking-sampling/release-verification/check.mjs](../../barness-ai-thinking-sampling/release-verification/check.mjs)；SHA256 `9e5a4a98ca926198eb46650fb93e86aa444f0292503a2d7e3659f739d17c1981`。历史源码 / 仅原轮次。固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**职责**：check.mjs：固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：兄弟run.mjs提供的临时work、out、inputs/locks/model-data及dist完整库存。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：out内等价report/model-data.sha256/dist清单及依赖核验结果。实际静态写点：L15 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L50 `mkdirSync(out, { recursive: true });`；L105 `writeFileSync(join(out, 'model-data.sha256'), dataList);`；L122 `writeFileSync(join(out, `dist-${suffix}.sha256`), inventory.join(''));`。

**副作用**：读完整源码/发布树并写报告；只检查对应固定pi版本的等价性，不是当前barness验收。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `import { readFileSync, writeFileSync, readdirSync, mkdirSync, lstatSync, existsSync } from 'node:fs';`；L15 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L50 `mkdirSync(out, { recursive: true });`；L105 `  writeFileSync(join(out, 'model-data.sha256'), dataList);`；L122 `    writeFileSync(join(out, `dist-${suffix}.sha256`), inventory.join(''));`。

<a id="barness-ai-thinking-sampling--release-verification--failures-mjs"></a>

## release-verification/failures.mjs

原件：[.scratch/barness-ai-thinking-sampling/release-verification/failures.mjs](../../barness-ai-thinking-sampling/release-verification/failures.mjs)；SHA256 `7913c43338c044178a40d813354a13167348c0bc9a4aaffb57cc75395b7c0988`。测试 / 仅原轮次。随固定上游来源等价核验封存的失败/故障子测。

**职责**：failures.mjs：随固定上游来源等价核验封存的失败/故障子测。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：对应固定版run/check与受控安装/重建树、下载/文件写入故障注入。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：failure-checks/boundary-checks及子进程状态，私有临时环境。实际静态写点：L43 `if (changed === null) rmSync(path); else writeFileSync(path, changed);`；L52 `writeFileSync(path, original);`；L53 `rmSync(out, { recursive: true, force: true });`；L61 `writeFileSync(join(auditOut, 'manifest.json'), '{}\n');`；L62 `writeFileSync(join(auditOut, 'capture.json'), JSON.stringify({ sample: secret }) + '\n');`；L72 `rmSync(auditOut, { recursive: true, force: true });`；L74 `writeFileSync(output, JSON.stringify({ status: 'PASS', cases: checks }, null, 2) + '\n');`。

**副作用**：启动Node子进程；failures会篡改再恢复work输入，faults可触发下载并清理临时树。只适用相应封存版，不自动运行。

**执行能力**：network：faults可通过run子进程触发下载；failures以临时work为输入；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：Node子进程及临时树变更/清理；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L3 `import { readFileSync, writeFileSync, mkdtempSync, rmSync, existsSync } from 'node:fs';`；L43 `    if (changed === null) rmSync(path); else writeFileSync(path, changed);`；L44 `    const result = spawnSync(process.execPath, [join(here, 'check.mjs'), work, out], { encoding: 'utf8' });`；L52 `    writeFileSync(path, original);`；L53 `    rmSync(out, { recursive: true, force: true });`。

<a id="barness-ai-thinking-sampling--release-verification--faults-mjs"></a>

## release-verification/faults.mjs

原件：[.scratch/barness-ai-thinking-sampling/release-verification/faults.mjs](../../barness-ai-thinking-sampling/release-verification/faults.mjs)；SHA256 `65ee901ded4b8701441ccf47e92dbaf032698a313219694d3033b8806207895b`。测试 / 仅原轮次。随固定上游来源等价核验封存的失败/故障子测。

**职责**：faults.mjs：随固定上游来源等价核验封存的失败/故障子测。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：对应固定版run/check与受控安装/重建树、下载/文件写入故障注入。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：failure-checks/boundary-checks及子进程状态，私有临时环境。实际静态写点：L59 `if (state?.workspace) rmSync(state.workspace, { recursive: true, force: true });`；L60 `rmSync(testDir, { recursive: true, force: true });`；L63 `writeFileSync(resolve(process.argv[2]), JSON.stringify({ status: 'PASS', cases: results }, null, 2) + '\n');`。

**副作用**：启动Node子进程；failures会篡改再恢复work输入，faults可触发下载并清理临时树。只适用相应封存版，不自动运行。

**执行能力**：network：faults可通过run子进程触发下载；failures以临时work为输入；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：Node子进程及临时树变更/清理；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L3 `import { readFileSync, writeFileSync, mkdtempSync, rmSync } from 'node:fs';`；L19 `      const originalWrite = fs.writeFileSync;`；L24 `      fs.writeFileSync = (path, ...args) => {`；L42 `    const result = spawnSync(process.execPath, ['--import', 'data:text/javascript,' + encodeURIComponent(loader),`；L59 `    if (state?.workspace) rmSync(state.workspace, { recursive: true, force: true });`。

<a id="barness-ai-thinking-sampling--release-verification--run-mjs"></a>

## release-verification/run.mjs

原件：[.scratch/barness-ai-thinking-sampling/release-verification/run.mjs](../../barness-ai-thinking-sampling/release-verification/run.mjs)；SHA256 `8b3a34a2cfe3a3cc0c12b57aad386c88e7b6fd46ddf1b325e9faf82f8bec411e`。历史源码 / 仅原轮次。固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**职责**：run.mjs：固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本目录inputs.json/package-lock/model-data固定pin、官方npm/GitHub发布包和固定commit。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：指定新out的发布包/源码等价报告、命令日志/哈希/审计；临时安装与源码树。实际静态写点：L16 `mkdirSync(out, { recursive: true });`；L23 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L78 `writeFileSync(join(work, name), '');`；L92 `writeFileSync(join(work, `registry-${suffix}.json`), JSON.stringify(selected, null, 2) + '\n');`；L95 `writeFileSync(join(work, `pi-${suffix}-1.1.0.tgz`), tarball);`；L97 `mkdirSync(join(work, 'release'));`；L98 `for (const name of ['package.json', 'package-lock.json']) cpSync(join(here, name), join(work, 'release', name));`；L103 `writeFileSync(join(work, 'source-tag.json'), JSON.stringify(tag));`；L106 `writeFileSync(join(work, 'source.tar.gz'), archive);`；L108 `mkdirSync(source);`；L113 `cpSync(join(work, 'release/node_modules/@earendil-works/pi-ai/dist/providers/data'),`；L133 `writeFileSync(join(out, 'commands.log'), safeLog(logs.join('\n')).trimEnd() + '\n');`；L138 `writeFileSync(join(out, 'inputs.sha256'), inputFiles.sort()`；L152 `writeFileSync(join(out, 'SHA256SUMS'), readdirSync(out).filter(name => name !== 'SHA256SUMS').sort()`；L162 `rmSync(work, { recursive: true, force: true });`。

**副作用**：联网下载、npm ci/install和TypeScript离线构建、Go审计、故障子测及临时树清理；两版本各自封存，不是活动Node oracle或产品TS源码。

**执行能力**：network：直接官方GitHub/npm下载及依赖安装；无模型调用；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：npm安装/TypeScript构建、Node子测试、Go审计；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L2 `// oracle execution occurs. All installs skip lifecycle scripts.`；L4 `import { readFileSync, writeFileSync, cpSync, mkdirSync, mkdtempSync, readdirSync, rmSync, existsSync } from 'node:fs';`；L16 `mkdirSync(out, { recursive: true });`；L23 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L36 `  commands.push({ stage: label, command: [cmd === process.execPath ? 'node' : cmd, ...args].map(s => s.replaceAll(here, '<inputs>').replaceAll(work, '<workspace>').replaceAll(out, '<evidence>')),`。

<a id="barness-ai-thinking-sampling--update-ledger-py"></a>

## update-ledger.py

原件：[.scratch/barness-ai-thinking-sampling/update-ledger.py](../../barness-ai-thinking-sampling/update-ledger.py)；SHA256 `2557187a253e17b6f78762728d674614dcdd2c0268543c15f1b8c5d62b1f41c1`。历史源码 / 仅原轮次。依赖已删除的旧差分ledger；不重建旧环境。

**职责**：Add finite duration scopes; retain all prior decisions and compare every field。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：已删除的ai/e2e/testdata/pidiff/ledger.json及--migration manifest/逐case数据。参数：`parser.add_argument('--migration', type=Path, required=True, help='manifest.json from the full migration-red archive')` L8。

**输出/写入**：覆盖同一旧差分ledger，追加有限duration决定。实际静态写点：L28 `p.write_text('{\n  "decisions": [\n' + ',\n'.join(('    ' + json.dumps(x, ensure_ascii=False, separators=(',', ':')) for x in d['decisions'])) + '\n  ],\n  "routes": ' + json.dumps(d['routes'], ensure_ascii=False, indent=2) + '\n}\n')`。

**副作用**：在parse_args前读旧ledger，help也非安全探针；旧输入已退役，不重建差分环境；当前改动用自身契约和留存fixture。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-thinking-sampling`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `json.loads(p.read_text())`；L5 `p.read_text()`；L9 `parser.parse_args()`；L11 `json.loads(manifest.read_text())`；L14 `json.loads(pcase.read_text())`。
