# barness-ai-release-verification 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [native-pressure/reproduce.py](#barness-ai-release-verification--native-pressure--reproduce-py) | 诊断装置 | 仅原轮次 |
| [seal-evidence.py](#barness-ai-release-verification--seal-evidence-py) | 历史源码 | 仅原轮次 |

<a id="barness-ai-release-verification--native-pressure--reproduce-py"></a>

## native-pressure/reproduce.py

原件：[.scratch/barness-ai-release-verification/native-pressure/reproduce.py](../../barness-ai-release-verification/native-pressure/reproduce.py)；SHA256 `327da34a2a1fac8ec3c2dd9b6dcc5bcf26e06c9a292fb24f296c8e6ec74d2eb4`。诊断装置 / 仅原轮次。原缺陷/资源/文件系统诊断；依赖原来源与私有fixture，不是产品资格入口。

**职责**：Diagnostic-only historical retry overlay. No injected socket errors or gate verdicts。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原source-base/pin、历史重试实现/overlay、可选负载/输出参数和受控OS资源。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：独立诊断目录的overlay/补丁副本/编译物/自然反例/日志/receipt。实际静态写点：L16 `dest.mkdir(mode=448, parents=True)`；L18 `(dest / 'condition').mkdir()`；L121 `overlay.write_text(json.dumps({'Replace': replaces}))`；L124 `(dest / 'build.log').write_bytes(build.stdout + build.stderr)`；L113 `target.write_text(patched)`；L117 `(dest / (source.name + '.txt')).write_text(patched)`；L146 `(dest / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L135 `ambient_source.write_bytes((Path(__file__).parent / 'ambient.go.txt').read_bytes())`；L139 `(dest / 'ambient-source.go.txt').write_bytes(ambient_source.read_bytes())`；L175 `(dest / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：可编译运行临时Go overlay、开loopback socket并产生压力/资源负载；历史600-Go等来源守卫限制适用性，不能当当前门禁或正式厂商证据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-release-verification/native-pressure`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `sys.argv.remove('--fixed')`；L11 `json.loads((Path(__file__).parent / 'source-base.json').read_text())`；L11 `(Path(__file__).parent / 'source-base.json').read_text()`；L12 `hashlib.sha256(p.read_bytes()).hexdigest()`；L12 `(root / 'ai').rglob('*.go')`。

<a id="barness-ai-release-verification--seal-evidence-py"></a>

## seal-evidence.py

原件：[.scratch/barness-ai-release-verification/seal-evidence.py](../../barness-ai-release-verification/seal-evidence.py)；SHA256 `d6985a587043e247c2dc2c7c472fbe71fa03853f89c33860ee32deec1d2de4a5`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Seal this follow-up's actual outputs; a retained FAIL never becomes PASS。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`choice.add_argument('--record', action='store_true')` L251；`choice.add_argument('--verify', action='store_true')` L252。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L141 `indexes.mkdir(exist_ok=True)`；L183 `RECEIPT.write_text('{\n' + ',\n'.join(('  ' + json.dumps(k) + ': ' + json.dumps(v, ensure_ascii=False, separators=(',', ':')) for k, v in receipt.items())) + '\n}\n')`；L156 `index.write_text(json.dumps(files, sort_keys=True, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-release-verification`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`；L10 `Path(__file__).resolve()`；L20 `spec.loader.exec_module(bytes14)`。
