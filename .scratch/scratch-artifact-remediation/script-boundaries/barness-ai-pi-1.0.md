# barness-ai-pi-1.0 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [chinese-evaluation-evidence/run-live.py](#barness-ai-pi-1-0--chinese-evaluation-evidence--run-live-py) | 历史源码 | 仅原轮次 |
| [chinese-evaluation-evidence/verify.py](#barness-ai-pi-1-0--chinese-evaluation-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [google-images-live-evidence/capture.py](#barness-ai-pi-1-0--google-images-live-evidence--capture-py) | 历史源码 | 仅原轮次 |
| [google-images-live-evidence/run-live.py](#barness-ai-pi-1-0--google-images-live-evidence--run-live-py) | 历史源码 | 仅原轮次 |
| [google-images-live-evidence/verify.py](#barness-ai-pi-1-0--google-images-live-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [google-interactions-evidence/capture.py](#barness-ai-pi-1-0--google-interactions-evidence--capture-py) | 历史源码 | 仅原轮次 |
| [google-interactions-evidence/verify.py](#barness-ai-pi-1-0--google-interactions-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [google-reference-images-evidence/capture.py](#barness-ai-pi-1-0--google-reference-images-evidence--capture-py) | 历史源码 | 仅原轮次 |
| [google-reference-images-evidence/verify.py](#barness-ai-pi-1-0--google-reference-images-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [mixed-operations-evidence/capture.py](#barness-ai-pi-1-0--mixed-operations-evidence--capture-py) | 历史源码 | 仅原轮次 |
| [mixed-operations-evidence/run_commands.py](#barness-ai-pi-1-0--mixed-operations-evidence--run_commands-py) | 历史源码 | 仅原轮次 |
| [mixed-operations-evidence/verify.py](#barness-ai-pi-1-0--mixed-operations-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [nine-route-release-evidence/capture.py](#barness-ai-pi-1-0--nine-route-release-evidence--capture-py) | 历史源码 | 仅原轮次 |
| [nine-route-release-evidence/run-live.py](#barness-ai-pi-1-0--nine-route-release-evidence--run-live-py) | 历史源码 | 仅原轮次 |
| [nine-route-release-evidence/run-release.py](#barness-ai-pi-1-0--nine-route-release-evidence--run-release-py) | 历史源码 | 仅原轮次 |
| [nine-route-release-evidence/test-run-release.py](#barness-ai-pi-1-0--nine-route-release-evidence--test-run-release-py) | 测试 | 仅原轮次 |
| [nine-route-release-evidence/verify.py](#barness-ai-pi-1-0--nine-route-release-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [openai-editing-evidence/capture.py](#barness-ai-pi-1-0--openai-editing-evidence--capture-py) | 历史源码 | 仅原轮次 |
| [openai-editing-evidence/verify.py](#barness-ai-pi-1-0--openai-editing-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [openai-generation-evidence/verify.py](#barness-ai-pi-1-0--openai-generation-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [openai-images-live-evidence/run_live.py](#barness-ai-pi-1-0--openai-images-live-evidence--run_live-py) | 历史源码 | 仅原轮次 |
| [openai-images-live-evidence/verify.py](#barness-ai-pi-1-0--openai-images-live-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [pressure-loopback-evidence/reproduce-retention.py](#barness-ai-pi-1-0--pressure-loopback-evidence--reproduce-retention-py) | 诊断装置 | 仅原轮次 |
| [pressure-loopback-evidence/run.py](#barness-ai-pi-1-0--pressure-loopback-evidence--run-py) | 历史源码 | 仅原轮次 |
| [pressure-loopback-evidence/verify.py](#barness-ai-pi-1-0--pressure-loopback-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [release-verification/check.mjs](#barness-ai-pi-1-0--release-verification--check-mjs) | 历史源码 | 仅原轮次 |
| [release-verification/failures.mjs](#barness-ai-pi-1-0--release-verification--failures-mjs) | 测试 | 仅原轮次 |
| [release-verification/faults.mjs](#barness-ai-pi-1-0--release-verification--faults-mjs) | 测试 | 仅原轮次 |
| [release-verification/run.mjs](#barness-ai-pi-1-0--release-verification--run-mjs) | 历史源码 | 仅原轮次 |
| [typesafe-live-evidence/run-live.py](#barness-ai-pi-1-0--typesafe-live-evidence--run-live-py) | 历史源码 | 仅原轮次 |
| [typesafe-live-evidence/verify.py](#barness-ai-pi-1-0--typesafe-live-evidence--verify-py) | 历史源码 | 仅原轮次 |
| [unary-failures-evidence/verify.py](#barness-ai-pi-1-0--unary-failures-evidence--verify-py) | 历史源码 | 仅原轮次 |

<a id="barness-ai-pi-1-0--chinese-evaluation-evidence--run-live-py"></a>

## chinese-evaluation-evidence/run-live.py

原件：[.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/run-live.py](../../barness-ai-pi-1.0/chinese-evaluation-evidence/run-live.py)；SHA256 `96bd48874c7dc85d9e5b08d46c885db1942db1d6b8e3cc466460f7cb99423357`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Run the independent Chinese evaluation with one isolated TypeSafe key。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--live', action='store_true', help='enable paid real evaluation')` L13。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L17 `log_base.mkdir(parents=True, exist_ok=True)`；L58 `log.write_bytes(output)`；L45 `log.write_bytes(diagnostics)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `parser.parse_args()`；L15 `Path(__file__).resolve()`；L17 `log_base.mkdir(parents=True, exist_ok=True)`；L21 `(root / '.env').read_text().splitlines()`；L22 `line.strip().removeprefix('export ').partition('=')`。

<a id="barness-ai-pi-1-0--chinese-evaluation-evidence--verify-py"></a>

## chinese-evaluation-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/verify.py](../../barness-ai-pi-1.0/chinese-evaluation-evidence/verify.py)；SHA256 `c6474de247f906d476548ad18806c190410e344b197b01fc3425d796400b714f`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify delivery/source hashes and independent evaluation evidence offline。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `Path(__file__).resolve()`；L10 `json.loads((base / 'manifest.json').read_text())`；L10 `(base / 'manifest.json').read_text()`；L19 `subprocess.run(['go', 'run', './ai/examples/chineseeval/cmd/chineseeval', '-verify', str(base / name)], cwd=repo, check=True)`；L20 `json.loads((base / 'historical-runs.json').read_text())`。

<a id="barness-ai-pi-1-0--google-images-live-evidence--capture-py"></a>

## google-images-live-evidence/capture.py

原件：[.scratch/barness-ai-pi-1.0/google-images-live-evidence/capture.py](../../barness-ai-pi-1.0/google-images-live-evidence/capture.py)；SHA256 `a112aae864d29fd945b4dd7f1723abd2eb5ec1a8d8db8ffcbeb1578d8ee05248`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Capture issue 14's audited artifacts without repeated fixture payloads。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L24 `(folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L44 `shutil.copytree(source, folder / name, dirs_exist_ok=True)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-images-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `Path(__file__).resolve()`；L11 `Path(__file__).resolve()`；L35 `{'delivery-refused': '.evidence/issue14-live/20261008T090854.560734000Z', 'prompt-refused': '.evidence/issue14-live/20261008T091316.585690000Z', 'live-pass': '.evidence/issue14-live/20261008T091619.605370000Z', 'not-run': '.evidence/issue14-not-run/20261008T091928.075162000Z'}.items()`；L46 `write('manifest.json', {'module': 'barness-ai', 'scope': 'issue 14 own live reports and offline evidence', 'base_commit': base, 'bundles': ['delivery-refused', 'prompt-refused', 'live-pass', 'not-run']})`；L49 `(root / sys.argv[1]).resolve()`。

<a id="barness-ai-pi-1-0--google-images-live-evidence--run-live-py"></a>

## google-images-live-evidence/run-live.py

原件：[.scratch/barness-ai-pi-1.0/google-images-live-evidence/run-live.py](../../barness-ai-pi-1.0/google-images-live-evidence/run-live.py)；SHA256 `30850232393f5f62d6550b33e25d7634a8a959ec0799c75018385c5813f210cc`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Explicit host dotenv injection; never reads or prints another provider key。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--dotenv', type=Path, required=True)` L11；`parser.add_argument('--evidence-dir', type=Path, required=True)` L12；`parser.add_argument('--alias', required=True)` L13。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-images-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `parser.parse_args()`；L16 `args.dotenv.read_text().splitlines()`；L16 `args.dotenv.read_text()`；L18 `name.strip().removeprefix('export ')`；L29 `args.evidence_dir.resolve()`。

<a id="barness-ai-pi-1-0--google-images-live-evidence--verify-py"></a>

## google-images-live-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/google-images-live-evidence/verify.py](../../barness-ai-pi-1.0/google-images-live-evidence/verify.py)；SHA256 `34954c3717dbb412ad5fcf344560538a098e28e4443ba57f2ff8571c42d0aa6f`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify checksums, actual report outcomes, raster headers and own matrix row。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-images-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `Path(__file__).resolve()`；L47 `json.loads(exchange['requestBody'])`；L45 `raster.read_bytes()`；L44 `raster.stat()`。

<a id="barness-ai-pi-1-0--google-interactions-evidence--capture-py"></a>

## google-interactions-evidence/capture.py

原件：[.scratch/barness-ai-pi-1.0/google-interactions-evidence/capture.py](../../barness-ai-pi-1.0/google-interactions-evidence/capture.py)；SHA256 `c133b8613c043763fe75f09c107ac518752d2e823ccfa335daf9eb7defae9b92`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Capture issue 12's audited run without copying repeated fixture payloads.  Per-issue capture stays independent so prior immutable deliveries keep their original evidence schema and verification behavior。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L57 `(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())`；L24 `(folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-interactions-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`；L14 `Path(__file__).resolve()`；L15 `(root / sys.argv[1]).resolve()`；L52 `subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'GLOSSARY.md', 'go.mod', 'go.sum'], cwd=root, text=True).splitlines()`；L52 `subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'GLOSSARY.md', 'go.mod', 'go.sum'], cwd=root, text=True)`。

<a id="barness-ai-pi-1-0--google-interactions-evidence--verify-py"></a>

## google-interactions-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/google-interactions-evidence/verify.py](../../barness-ai-pi-1.0/google-interactions-evidence/verify.py)；SHA256 `e7e55a37afb28c47aa4607a9bd8978aeab51d61b0f5becf8eef43f6a3e8c9a51`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify source, delivery and available original-run hashes for issue 12。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-interactions-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `Path(__file__).resolve()`；L8 `Path(__file__).resolve()`。

<a id="barness-ai-pi-1-0--google-reference-images-evidence--capture-py"></a>

## google-reference-images-evidence/capture.py

原件：[.scratch/barness-ai-pi-1.0/google-reference-images-evidence/capture.py](../../barness-ai-pi-1.0/google-reference-images-evidence/capture.py)；SHA256 `8a951af027b4401f789e93680b287903dff4c4c6b462092052b20b11eb42b714`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Capture issue 13's audited run without copying repeated fixture payloads.  Per-issue capture stays independent so prior immutable deliveries keep their original evidence schema and verification behavior。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L57 `(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())`；L24 `(folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-reference-images-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`；L14 `Path(__file__).resolve()`；L15 `(root / sys.argv[1]).resolve()`；L52 `subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'GLOSSARY.md', 'go.mod', 'go.sum'], cwd=root, text=True).splitlines()`；L52 `subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'GLOSSARY.md', 'go.mod', 'go.sum'], cwd=root, text=True)`。

<a id="barness-ai-pi-1-0--google-reference-images-evidence--verify-py"></a>

## google-reference-images-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/google-reference-images-evidence/verify.py](../../barness-ai-pi-1.0/google-reference-images-evidence/verify.py)；SHA256 `0bdd6eedc4462dd4b7365b406b2e26360da3587637f6bde0871557169b01343e`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify source, delivery and available original-run hashes for issue 13。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/google-reference-images-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `Path(__file__).resolve()`；L8 `Path(__file__).resolve()`。

<a id="barness-ai-pi-1-0--mixed-operations-evidence--capture-py"></a>

## mixed-operations-evidence/capture.py

原件：[.scratch/barness-ai-pi-1.0/mixed-operations-evidence/capture.py](../../barness-ai-pi-1.0/mixed-operations-evidence/capture.py)；SHA256 `34ccdb54eac546db92f0d8676eacbc9db5e113a2aad9607119de019e8a9af7d2`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Capture issue 15's audited public E2E evidence without large load payloads.  This delivery keeps its own schema so later tickets cannot change its hashes. Run from the repo with the full run and the existing releasegate output paths。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L126 `(folder / 'manifest.json').write_bytes((bundle / 'manifest.json').read_bytes())`；L127 `(folder / 'run-audit.json').write_bytes((bundle / 'audit.json').read_bytes())`；L128 `(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())`；L26 `(folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L58 `(folder / log.name).write_bytes(log.read_bytes())`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/mixed-operations-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`；L14 `Path(__file__).resolve()`；L58 `(folder / log.name).write_bytes(log.read_bytes())`；L58 `log.read_bytes()`；L59 `write('command-results.json', commands)`。

<a id="barness-ai-pi-1-0--mixed-operations-evidence--run_commands-py"></a>

## mixed-operations-evidence/run_commands.py

原件：[.scratch/barness-ai-pi-1.0/mixed-operations-evidence/run_commands.py](../../barness-ai-pi-1.0/mixed-operations-evidence/run_commands.py)；SHA256 `2db22568510394190ae388172753425ab64ca3f1919fba2b168fb7075bbe40e7`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Run actual verification commands and record exit codes, commits and logs.  Use --checks-only when an independently recorded full run is already available. Keep output below ignored .evidence so command logs cannot expose local paths。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('output')` L17；`parser.add_argument('--checks-only', action='store_true')` L18。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L21 `output.mkdir(parents=True, exist_ok=True)`；L56 `(output / 'command-results.json').write_text(json.dumps({'tested_commit': commit, 'commands': records}, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/mixed-operations-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L15 `Path(__file__).resolve()`；L19 `parser.parse_args()`；L21 `output.mkdir(parents=True, exist_ok=True)`；L22 `subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()`；L22 `subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True)`。

<a id="barness-ai-pi-1-0--mixed-operations-evidence--verify-py"></a>

## mixed-operations-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/mixed-operations-evidence/verify.py](../../barness-ai-pi-1.0/mixed-operations-evidence/verify.py)；SHA256 `6fb6e73ccbc15a97a4f38de8cc629bce0c395fd72781ece8f30398088e669e07`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify source, public run artifacts and the portable issue 15 delivery。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/mixed-operations-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L8 `Path(__file__).resolve()`；L9 `Path(__file__).resolve()`；L29 `subprocess.check_output(['git', 'show', f"{v['tested_commit']}:{p}"], cwd=root)`。

<a id="barness-ai-pi-1-0--nine-route-release-evidence--capture-py"></a>

## nine-route-release-evidence/capture.py

原件：[.scratch/barness-ai-pi-1.0/nine-route-release-evidence/capture.py](../../barness-ai-pi-1.0/nine-route-release-evidence/capture.py)；SHA256 `3ac426c3d73a5b627895e6a5bee62ac0d6e09927eefe17faa2ed6bef95ac7ee9`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Package a completed release without altering any recorded run artifact。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('run', type=Path)` L37；`parser.add_argument('--history', type=Path, action='append', default=[], help='completed failed gate whose full offline bundle must be preserved')` L38。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L27 `(DEST / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L54 `shutil.copyfile(base / name, DEST / name)`；L61 `(DEST / name).write_text(raw.replace(str(ROOT), '[REPO]').replace(str(Path.home()), '[HOME]'))`；L122 `saved.mkdir(parents=True, exist_ok=True)`；L52 `shutil.copyfile(p, DEST / p.name)`；L77 `logs.mkdir(parents=True, exist_ok=True)`；L127 `(saved / source.name).write_bytes(raw)`；L79 `shutil.copyfile(parent / name, logs / name)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/nine-route-release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`；L15 `Path(__file__).resolve()`；L39 `parser.parse_args()`；L40 `args.run.resolve()`；L60 `(ROOT / source).read_text()`。

<a id="barness-ai-pi-1-0--nine-route-release-evidence--run-live-py"></a>

## nine-route-release-evidence/run-live.py

原件：[.scratch/barness-ai-pi-1.0/nine-route-release-evidence/run-live.py](../../barness-ai-pi-1.0/nine-route-release-evidence/run-live.py)；SHA256 `50b7ba4428f84d1a3d8a192410c95cfebe82458dff57bf04807b7340d3bc29e6`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：One explicit combination per process; the release gate never handles keys。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--combo', required=True, choices=['openai-responses', 'openai-chat', 'anthropic-messages', 'google-gemini', 'deepseek-responses', 'deepseek-chat', 'typesafe-classifier', 'openai-images', 'google-interactions-image'])` L12；`parser.add_argument('--live', action='store_true')` L13；`parser.add_argument('--binary', type=Path, required=True, help='go test -c -tags live -o <binary> ./ai/live')` L14；`parser.add_argument('--out', type=Path, required=True)` L15；`parser.add_argument('--alias', required=True, help='host-confirmed account/region alias, or explicit unconfirmed alias')` L16。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L38 `args.out.mkdir(parents=True, exist_ok=True)`；L45 `(args.out / 'run.log').write_bytes(output)`；L48 `(args.out / 'command.json').write_text(json.dumps(result, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/nine-route-release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `parser.parse_args()`；L18 `args.out.resolve()`；L19 `Path(__file__).resolve()`；L24 `(root / '.env').open()`；L26 `line.strip().removeprefix('export ').partition('=')`。

<a id="barness-ai-pi-1-0--nine-route-release-evidence--run-release-py"></a>

## nine-route-release-evidence/run-release.py

原件：[.scratch/barness-ai-pi-1.0/nine-route-release-evidence/run-release.py](../../barness-ai-pi-1.0/nine-route-release-evidence/run-release.py)；SHA256 `fd0c1ed55819ce58bf288fc18400ef63c937b8daa849878936fc9c95bae4ba8a`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Replay nine isolated smokes, atomic matrix merges and the evidence-only gate。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--live', action='store_true', help='explicitly authorize the bounded real smokes; absent keys produce NOT_RUN')` L10；`parser.add_argument('--out', type=Path, required=True, help='fresh ignored output directory')` L11。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L21 `args.out.mkdir(parents=True)`；L24 `binary.parent.mkdir(parents=True, exist_ok=True)`；L49 `(args.out / 'gate-command.json').write_text(json.dumps({'argv': gate}, indent=2) + '\n')`；L51 `(args.out / 'gate.log').write_bytes(run.stdout.replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))`；L52 `(args.out / 'gate-exit.json').write_text(json.dumps({'exitCode': run.returncode}, indent=2) + '\n')`；L44 `(args.out / 'smoke-commands.json').write_text(json.dumps(results, indent=2) + '\n')`；L42 `(out / 'merge.log').write_bytes((merge.stdout + merge.stderr).replace(str(root).encode(), b'[REPO]').replace(str(Path.home()).encode(), b'[HOME]'))`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/nine-route-release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `parser.parse_args()`；L13 `Path(__file__).resolve()`；L14 `args.out.resolve()`；L21 `args.out.mkdir(parents=True)`；L24 `binary.parent.mkdir(parents=True, exist_ok=True)`。

<a id="barness-ai-pi-1-0--nine-route-release-evidence--test-run-release-py"></a>

## nine-route-release-evidence/test-run-release.py

原件：[.scratch/barness-ai-pi-1.0/nine-route-release-evidence/test-run-release.py](../../barness-ai-pi-1.0/nine-route-release-evidence/test-run-release.py)；SHA256 `66b29d5c748849c13c0accdc26000835e5d4074462eedbe17fb0c23f1ca32ae2`。测试 / 仅原轮次。原归档/原控制器的失败清单，未发现当前公共Go或现行SOP执行接线。

**职责**：Controlled process replay: no keys, network or real Provider calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原固定源码/工具/归档和私有篡改副本或合成fixture；支持模块随对应测试使用。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：该历史失败清单的断言/拒绝记录和临时日志；不续期真实资格。实际静态写点：L20 `scripts.mkdir(parents=True)`；L22 `shutil.copyfile(RUNNER.with_name('run-live.py'), script)`；L24 `binary.write_text("#!/usr/bin/env python3\nimport os\nfrom pathlib import Path\np=Path(os.environ['BARNESS_AI_EVIDENCE_DIR'])/'controlled'\np.mkdir(parents=True)\n(p/'live-report.json').write_text('{}')\n")`；L25 `binary.chmod(493)`；L26 `(root / 'ai/live').mkdir(parents=True)`；L36 `stub.mkdir()`；L38 `go.write_text('#!/usr/bin/env python3\nimport json, os, sys\nfrom pathlib import Path\na = sys.argv[1:]\nif a[0] == \'test\':\n    p = Path(a[a.index(\'-o\')+1])\n    p.parent.mkdir(parents=True, exist_ok=True)\n    p.write_text("#!/usr/bin/env python3\\nimport os\\nfrom pathlib import Path\\np=Path(os.environ[\'BARNESS_AI_EVIDENCE_DIR\'])/\'contro`；L51 `go.chmod(493)`；L56 `go.write_text(go.read_text().replace("os.environ['BUILD_RECORD']", repr(str(record))))`。

**副作用**：运行相邻公共CLI或临时程序，写/删私有fixture；部分故障测试会临时替换再恢复输入，仅原环境适用，不在本次执行。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/nine-route-release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`。

<a id="barness-ai-pi-1-0--nine-route-release-evidence--verify-py"></a>

## nine-route-release-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/nine-route-release-evidence/verify.py](../../barness-ai-pi-1.0/nine-route-release-evidence/verify.py)；SHA256 `b7d82caad599af9457f677207bbaf87c3420f99229768b4d7c95fafe3b2e6fd9`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify portable full artifacts, re-audit and recompute evidence-only gates。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L69 `tar.extractall(work, filter='data')`；L105 `tar.extractall(work, filter='data')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/nine-route-release-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`；L14 `Path(__file__).resolve()`；L50 `subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True).splitlines()`；L50 `subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'ai', 'go.mod', 'go.sum'], cwd=ROOT, text=True)`；L54 `subprocess.check_output(['git', 'show', report['commit'] + ':' + name], cwd=ROOT)`。

<a id="barness-ai-pi-1-0--openai-editing-evidence--capture-py"></a>

## openai-editing-evidence/capture.py

原件：[.scratch/barness-ai-pi-1.0/openai-editing-evidence/capture.py](../../barness-ai-pi-1.0/openai-editing-evidence/capture.py)；SHA256 `c6d8020ce92809c714c97845f199da9511d20fb04c66c08ae632f8b2c28d473e`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Harvest issue 10's audited offline run using relative paths and content hashes。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L47 `(folder / 'catalog-snapshot.json').write_bytes((root / 'ai/release/catalog-snapshot.json').read_bytes())`；L17 `(folder / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/openai-editing-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`；L10 `Path(__file__).resolve()`；L11 `(root / sys.argv[1]).resolve()`；L43 `subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'go.mod', 'go.sum'], cwd=root, text=True).splitlines()`；L43 `subprocess.check_output(['git', 'diff', '--name-only', base, 'HEAD', '--', 'ai', 'docs', 'go.mod', 'go.sum'], cwd=root, text=True)`。

<a id="barness-ai-pi-1-0--openai-editing-evidence--verify-py"></a>

## openai-editing-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/openai-editing-evidence/verify.py](../../barness-ai-pi-1.0/openai-editing-evidence/verify.py)；SHA256 `367e4b9ce0f117b7aecac0015701c9c2fe8ca58725ce15bcf46e8ceffa5749ba`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify the committed delivery hashes and locate the repeatable run bundle。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/openai-editing-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `Path(__file__).resolve()`；L8 `Path(__file__).resolve()`；L9 `json.loads((folder / 'delivery-hashes.json').read_text())`；L9 `(folder / 'delivery-hashes.json').read_text()`；L11 `hashlib.sha256((root / path).read_bytes()).hexdigest()`。

<a id="barness-ai-pi-1-0--openai-generation-evidence--verify-py"></a>

## openai-generation-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/openai-generation-evidence/verify.py](../../barness-ai-pi-1.0/openai-generation-evidence/verify.py)；SHA256 `367e4b9ce0f117b7aecac0015701c9c2fe8ca58725ce15bcf46e8ceffa5749ba`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify the committed delivery hashes and locate the repeatable run bundle。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/openai-generation-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L7 `Path(__file__).resolve()`；L8 `Path(__file__).resolve()`；L9 `json.loads((folder / 'delivery-hashes.json').read_text())`；L9 `(folder / 'delivery-hashes.json').read_text()`；L11 `hashlib.sha256((root / path).read_bytes()).hexdigest()`。

<a id="barness-ai-pi-1-0--openai-images-live-evidence--run_live-py"></a>

## openai-images-live-evidence/run_live.py

原件：[.scratch/barness-ai-pi-1.0/openai-images-live-evidence/run_live.py](../../barness-ai-pi-1.0/openai-images-live-evidence/run_live.py)；SHA256 `9ad52d6a493f2db80f04e001992e945607389e0cb1b63d0c5336487fd6260ce2`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Explicit issue-11 live runner: inject only one dotenv credential into Go.  Run from the repository root. Costs at most four 1024-square output requests. Account permissions/region are host configuration, not inferred from a key。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/openai-images-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L13 `Path(__file__).resolve()`；L15 `(root / '.env').read_text().splitlines()`；L15 `(root / '.env').read_text()`；L23 `Path(__file__).resolve()`；L30 `(base / 'live.log').open('w')`。

<a id="barness-ai-pi-1-0--openai-images-live-evidence--verify-py"></a>

## openai-images-live-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/openai-images-live-evidence/verify.py](../../barness-ai-pi-1.0/openai-images-live-evidence/verify.py)；SHA256 `1705cf9ea001913ecd1b2c48e3a62b63505d42917dcde79aedaf8a5901265d61`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Repeatable offline checks of own live evidence, catalog and atomic merge。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L101 `(base / 'verification.json').write_text(json.dumps(summary, indent=2) + '\n')`；L59 `matrix.write_text(json.dumps(initial, indent=2) + '\n')`；L64 `bad_path.write_text(json.dumps(bad))`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/openai-images-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`；L12 `Path(__file__).resolve()`；L14 `json.loads(p.read_text())`；L14 `p.read_text()`；L33 `next(directory.glob('*-output.*')).read_bytes()`。

<a id="barness-ai-pi-1-0--pressure-loopback-evidence--reproduce-retention-py"></a>

## pressure-loopback-evidence/reproduce-retention.py

原件：[.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/reproduce-retention.py](../../barness-ai-pi-1.0/pressure-loopback-evidence/reproduce-retention.py)；SHA256 `2c375b47e1e3fe7dc2ccc9d1a7cd1197110a36665cae26565b99fe1091bf7316`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Replay the confirmed retention defect in a disposable source copy。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原source-base/pin、历史重试实现/overlay、可选负载/输出参数和受控OS资源。参数：`parser.add_argument('--out', type=Path, required=True)` L15。

**输出/写入**：独立诊断目录的overlay/补丁副本/编译物/自然反例/日志/receipt。实际静态写点：L23 `out.mkdir(parents=True, exist_ok=False)`；L51 `(out / 'result.json').write_text(json.dumps(verdict, indent=2) + '\n')`；L54 `(out / 'manifest.json').write_text(json.dumps({'module': 'barness-pressure-reproduction', **verdict}, indent=2) + '\n')`；L38 `(source / 'ai/internal/testkit/evidence/evidence.go').write_bytes(legacy)`；L40 `(out / 'source-hashes.json').write_text(json.dumps(hashes, indent=2, sort_keys=True) + '\n')`；L45 `(out / 'go-test.log').write_bytes(log)`；L36 `target.parent.mkdir(parents=True, exist_ok=True)`；L37 `target.write_bytes(path.read_bytes())`。

**副作用**：可编译运行临时Go overlay、开loopback socket并产生压力/资源负载；历史600-Go等来源守卫限制适用性，不能当当前门禁或正式厂商证据。

**执行能力**：network：存在网络/loopback API静态调用，见networkSites；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/pressure-loopback-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-pi-1-0--pressure-loopback-evidence--run-py"></a>

## pressure-loopback-evidence/run.py

原件：[.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/run.py](../../barness-ai-pi-1.0/pressure-loopback-evidence/run.py)；SHA256 `d282f00a0cc1221aa965e0b3bf1c9086abe84f01c2909328c4da43dac9d04910`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Run unchanged public-Client pressure loads and preserve each process verdict。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--mode', choices=['isolated', 'fresh', 'full', 'race'], required=True)` L14；`parser.add_argument('--rounds', type=int, default=100)` L15；`parser.add_argument('--out', type=Path, required=True)` L16。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L26 `out.mkdir(parents=True, exist_ok=False)`；L39 `(out / 'source-hashes.json').write_text(json.dumps(sources, indent=2, sort_keys=True) + '\n')`；L45 `process_out.mkdir()`；L58 `(process_out / 'go-test.log').write_bytes(log)`；L61 `(out / 'results.json').write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：存在网络/loopback API静态调用，见networkSites；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/pressure-loopback-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-pi-1-0--pressure-loopback-evidence--verify-py"></a>

## pressure-loopback-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/verify.py](../../barness-ai-pi-1.0/pressure-loopback-evidence/verify.py)；SHA256 `be577e832793b791d9dd83e94ad8865315b4d7cc80168567ed257eeb31774bcf`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Recheck archive integrity, complete case artifacts, source and redaction。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L77 `(work / 'manifest.json').write_text(json.dumps({'module': 'barness-pressure-archive-audit', 'archive': name}) + '\n')`；L47 `target.mkdir(parents=True, exist_ok=True)`；L49 `target.parent.mkdir(parents=True, exist_ok=True)`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/pressure-loopback-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-pi-1-0--release-verification--check-mjs"></a>

## release-verification/check.mjs

原件：[.scratch/barness-ai-pi-1.0/release-verification/check.mjs](../../barness-ai-pi-1.0/release-verification/check.mjs)；SHA256 `4b857c498aceee818960ebcc50ee6c7603d5b8febf856dc68e53b9b7e11f6e49`。历史源码 / 仅原轮次。固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**职责**：check.mjs：固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：兄弟run.mjs提供的临时work、out、inputs/locks/model-data及dist完整库存。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：out内等价report/model-data.sha256/dist清单及依赖核验结果。实际静态写点：L15 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L58 `mkdirSync(out, { recursive: true });`；L111 `writeFileSync(join(out, 'model-data.sha256'), dataList);`；L128 `writeFileSync(join(out, `dist-${suffix}.sha256`), inventory.join(''));`。

**副作用**：读完整源码/发布树并写报告；只检查对应固定pi版本的等价性，不是当前barness验收。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L4 `import { readFileSync, writeFileSync, readdirSync, mkdirSync, lstatSync, existsSync } from 'node:fs';`；L15 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L58 `mkdirSync(out, { recursive: true });`；L111 `  writeFileSync(join(out, 'model-data.sha256'), dataList);`；L128 `    writeFileSync(join(out, `dist-${suffix}.sha256`), inventory.join(''));`。

<a id="barness-ai-pi-1-0--release-verification--failures-mjs"></a>

## release-verification/failures.mjs

原件：[.scratch/barness-ai-pi-1.0/release-verification/failures.mjs](../../barness-ai-pi-1.0/release-verification/failures.mjs)；SHA256 `0ef18f154271ab14353d7032644f539906afbfd5ceb4ae8ab5733d8cf06711a5`。测试 / 仅原轮次。随固定上游来源等价核验封存的失败/故障子测。

**职责**：failures.mjs：随固定上游来源等价核验封存的失败/故障子测。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：对应固定版run/check与受控安装/重建树、下载/文件写入故障注入。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：failure-checks/boundary-checks及子进程状态，私有临时环境。实际静态写点：L43 `if (changed === null) rmSync(path); else writeFileSync(path, changed);`；L52 `writeFileSync(path, original);`；L53 `rmSync(out, { recursive: true, force: true });`；L61 `writeFileSync(join(auditOut, 'manifest.json'), '{}\n');`；L62 `writeFileSync(join(auditOut, 'capture.json'), JSON.stringify({ sample: secret }) + '\n');`；L72 `rmSync(auditOut, { recursive: true, force: true });`；L74 `writeFileSync(output, JSON.stringify({ status: 'PASS', cases: checks }, null, 2) + '\n');`。

**副作用**：启动Node子进程；failures会篡改再恢复work输入，faults可触发下载并清理临时树。只适用相应封存版，不自动运行。

**执行能力**：network：faults可通过run子进程触发下载；failures以临时work为输入；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：Node子进程及临时树变更/清理；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L3 `import { readFileSync, writeFileSync, mkdtempSync, rmSync, existsSync } from 'node:fs';`；L43 `    if (changed === null) rmSync(path); else writeFileSync(path, changed);`；L44 `    const result = spawnSync(process.execPath, [join(here, 'check.mjs'), work, out], { encoding: 'utf8' });`；L52 `    writeFileSync(path, original);`；L53 `    rmSync(out, { recursive: true, force: true });`。

<a id="barness-ai-pi-1-0--release-verification--faults-mjs"></a>

## release-verification/faults.mjs

原件：[.scratch/barness-ai-pi-1.0/release-verification/faults.mjs](../../barness-ai-pi-1.0/release-verification/faults.mjs)；SHA256 `65ee901ded4b8701441ccf47e92dbaf032698a313219694d3033b8806207895b`。测试 / 仅原轮次。随固定上游来源等价核验封存的失败/故障子测。

**职责**：faults.mjs：随固定上游来源等价核验封存的失败/故障子测。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：对应固定版run/check与受控安装/重建树、下载/文件写入故障注入。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：failure-checks/boundary-checks及子进程状态，私有临时环境。实际静态写点：L59 `if (state?.workspace) rmSync(state.workspace, { recursive: true, force: true });`；L60 `rmSync(testDir, { recursive: true, force: true });`；L63 `writeFileSync(resolve(process.argv[2]), JSON.stringify({ status: 'PASS', cases: results }, null, 2) + '\n');`。

**副作用**：启动Node子进程；failures会篡改再恢复work输入，faults可触发下载并清理临时树。只适用相应封存版，不自动运行。

**执行能力**：network：faults可通过run子进程触发下载；failures以临时work为输入；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：Node子进程及临时树变更/清理；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L3 `import { readFileSync, writeFileSync, mkdtempSync, rmSync } from 'node:fs';`；L19 `      const originalWrite = fs.writeFileSync;`；L24 `      fs.writeFileSync = (path, ...args) => {`；L42 `    const result = spawnSync(process.execPath, ['--import', 'data:text/javascript,' + encodeURIComponent(loader),`；L59 `    if (state?.workspace) rmSync(state.workspace, { recursive: true, force: true });`。

<a id="barness-ai-pi-1-0--release-verification--run-mjs"></a>

## release-verification/run.mjs

原件：[.scratch/barness-ai-pi-1.0/release-verification/run.mjs](../../barness-ai-pi-1.0/release-verification/run.mjs)；SHA256 `3917f3603713aaba5dc0737a48e68939ef0a543565d95d10fc0c9cb2172221d3`。历史源码 / 仅原轮次。固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**职责**：run.mjs：固定pi发布件来源等价工具；活动oracle/差分已由ADR0051取代。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本目录inputs.json/package-lock/model-data固定pin、官方npm/GitHub发布包和固定commit。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：指定新out的发布包/源码等价报告、命令日志/哈希/审计；临时安装与源码树。实际静态写点：L16 `mkdirSync(out, { recursive: true });`；L23 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L78 `writeFileSync(join(work, name), '');`；L92 `writeFileSync(join(work, `registry-${suffix}.json`), JSON.stringify(selected, null, 2) + '\n');`；L95 `writeFileSync(join(work, `pi-${suffix}-1.0.0.tgz`), tarball);`；L97 `mkdirSync(join(work, 'release'));`；L98 `for (const name of ['package.json', 'package-lock.json']) cpSync(join(here, name), join(work, 'release', name));`；L107 `cpSync(join(work, 'release/node_modules/@earendil-works/pi-ai/dist/providers/data'),`；L127 `writeFileSync(join(out, 'commands.log'), safeLog(logs.join('\n')).trimEnd() + '\n');`；L132 `writeFileSync(join(out, 'inputs.sha256'), inputFiles.sort()`；L146 `writeFileSync(join(out, 'SHA256SUMS'), readdirSync(out).filter(name => name !== 'SHA256SUMS').sort()`；L156 `rmSync(work, { recursive: true, force: true });`。

**副作用**：联网下载、npm ci/install和TypeScript离线构建、Go审计、故障子测及临时树清理；两版本各自封存，不是活动Node oracle或产品TS源码。

**执行能力**：network：直接官方GitHub/npm下载及依赖安装；无模型调用；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：npm安装/TypeScript构建、Node子测试、Go审计；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L2 `// oracle execution occurs. All installs skip lifecycle scripts.`；L4 `import { readFileSync, writeFileSync, cpSync, mkdirSync, mkdtempSync, readdirSync, rmSync, existsSync } from 'node:fs';`；L16 `mkdirSync(out, { recursive: true });`；L23 `const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');`；L36 `  commands.push({ stage: label, command: [cmd === process.execPath ? 'node' : cmd, ...args].map(s => s.replaceAll(here, '<inputs>').replaceAll(work, '<workspace>').replaceAll(out, '<evidence>')),`。

<a id="barness-ai-pi-1-0--typesafe-live-evidence--run-live-py"></a>

## typesafe-live-evidence/run-live.py

原件：[.scratch/barness-ai-pi-1.0/typesafe-live-evidence/run-live.py](../../barness-ai-pi-1.0/typesafe-live-evidence/run-live.py)；SHA256 `2f18e14cc2cf6fdc245853d3be45961601adfcffe9961285cf568f306d310f0c`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Inject only TYPESAFE_KEY into the isolated official classifier smoke。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L29 `(root / '.scratch/barness-ai-pi-1.0/typesafe-live-evidence/live.log').write_bytes(output)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/typesafe-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`；L11 `(root / '.env').read_text().splitlines()`；L12 `line.strip().removeprefix('export ').partition('=')`；L11 `(root / '.env').read_text()`；L12 `line.strip().removeprefix('export ')`。

<a id="barness-ai-pi-1-0--typesafe-live-evidence--verify-py"></a>

## typesafe-live-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/typesafe-live-evidence/verify.py](../../barness-ai-pi-1.0/typesafe-live-evidence/verify.py)；SHA256 `68e7a78de30a61c5eb3e157aa8231015ce3fe576c3c1b1bcbf0c209cb521491e`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify saved issue08 evidence from the repo root; no network or keys needed。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/typesafe-live-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-pi-1-0--unary-failures-evidence--verify-py"></a>

## unary-failures-evidence/verify.py

原件：[.scratch/barness-ai-pi-1.0/unary-failures-evidence/verify.py](../../barness-ai-pi-1.0/unary-failures-evidence/verify.py)；SHA256 `408b5fc9f30e2c24d15e7f2f22acb01310f6a333e05789caf547f5929dd89fa1`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify saved issue07 evidence from the repository root; no network required。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-pi-1.0/unary-failures-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。
