# barness-ai-glm 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [evidence/ticket-01/compare.py](#barness-ai-glm--evidence--ticket-01--compare-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-10/verify-evidence.py](#barness-ai-glm--evidence--ticket-10--verify-evidence-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-11/launcher.py](#barness-ai-glm--evidence--ticket-11--launcher-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-12/launcher.py](#barness-ai-glm--evidence--ticket-12--launcher-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/commands.py](#barness-ai-glm--evidence--ticket-13--commands-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [evidence/ticket-13/corrective.py](#barness-ai-glm--evidence--ticket-13--corrective-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [evidence/ticket-13/glm_launcher.py](#barness-ai-glm--evidence--ticket-13--glm_launcher-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [evidence/ticket-13/index.py](#barness-ai-glm--evidence--ticket-13--index-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/preserve-default-history.py](#barness-ai-glm--evidence--ticket-13--preserve-default-history-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/preserve-history.py](#barness-ai-glm--evidence--ticket-13--preserve-history-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/reconcile-rollback-history.py](#barness-ai-glm--evidence--ticket-13--reconcile-rollback-history-py) | 历史源码 | 仅原轮次 |
| [evidence/ticket-13/rollback.py](#barness-ai-glm--evidence--ticket-13--rollback-py) | 活动工具 | 现有公共消费者 |
| [evidence/ticket-13/run.py](#barness-ai-glm--evidence--ticket-13--run-py) | 活动工具 | 仅测试消费，旧真实轮次关闭 |
| [evidence/ticket-13/seal.py](#barness-ai-glm--evidence--ticket-13--seal-py) | 历史源码 | 仅原轮次 |

<a id="barness-ai-glm--evidence--ticket-01--compare-py"></a>

## evidence/ticket-01/compare.py

原件：[.scratch/barness-ai-glm/evidence/ticket-01/compare.py](../../barness-ai-glm/evidence/ticket-01/compare.py)；SHA256 `169473f13f7708a214e93b3678988cbdc0db2c072007eee88e7a902aa5a1555a`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Compare captured public Client behavior, verifying each bundle's hashes。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('before', type=Path)` L56；`parser.add_argument('after', type=Path)` L57；`parser.add_argument('--output', type=Path, required=True)` L58。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L109 `args.output.parent.mkdir(parents=True, exist_ok=True)`；L110 `args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-01`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-glm--evidence--ticket-10--verify-evidence-py"></a>

## evidence/ticket-10/verify-evidence.py

原件：[.scratch/barness-ai-glm/evidence/ticket-10/verify-evidence.py](../../barness-ai-glm/evidence/ticket-10/verify-evidence.py)；SHA256 `a8008dd17795ba5c9578e50d12c07d75562a0978a36716365eee1b2cb566754d`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify saved ticket10 bundles; no network, credentials or model calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('raw_root', type=Path)` L30；`parser.add_argument('--out', type=Path, required=True)` L31。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L82 `proof_path.write_text(json.dumps(proof, ensure_ascii=False, indent=2) + '\n')`；L89 `args.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-10`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-glm--evidence--ticket-11--launcher-py"></a>

## evidence/ticket-11/launcher.py

原件：[.scratch/barness-ai-glm/evidence/ticket-11/launcher.py](../../barness-ai-glm/evidence/ticket-11/launcher.py)；SHA256 `6273e9001f39fcfb15e4295823416d25b2117aa1d512cacaea0cb6a6713ddc7d`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Explicit Host launcher. Never imports dotenv or reads another provider key。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`ap.add_argument('--execute', action='store_true', required=True)` L20；`ap.add_argument('--route', choices=['chat', 'responses', 'messages'], required=True)` L20；`ap.add_argument('--binary', required=True)` L20。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L22 `OUT.mkdir(parents=True, exist_ok=True)`；L22 `base.mkdir()`；L38 `(base / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L40 `(base / 'SHA256SUMS.json').write_text(json.dumps(inventory, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-11`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=pathlib.Path('/Users/cyber/RestoX/barness')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `pathlib.Path(__file__).resolve()`。

<a id="barness-ai-glm--evidence--ticket-12--launcher-py"></a>

## evidence/ticket-12/launcher.py

原件：[.scratch/barness-ai-glm/evidence/ticket-12/launcher.py](../../barness-ai-glm/evidence/ticket-12/launcher.py)；SHA256 `2581d484786e63490f6acc0249590f703646c270ac522dc5c6abdb7c671dc9d8`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Explicit Host launcher. Never imports dotenv or reads another provider key。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`ap.add_argument('--execute', action='store_true', required=True)` L20；`ap.add_argument('--route', choices=['chat', 'responses', 'messages'], required=True)` L20；`ap.add_argument('--binary', required=True)` L20。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L22 `OUT.mkdir(parents=True, exist_ok=True)`；L22 `base.mkdir()`；L38 `(base / 'receipt.json').write_text(json.dumps(receipt, indent=2) + '\n')`；L40 `(base / 'SHA256SUMS.json').write_text(json.dumps(inventory, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-12`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=pathlib.Path('/Users/cyber/RestoX/barness')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `pathlib.Path(__file__).resolve()`。

<a id="barness-ai-glm--evidence--ticket-13--commands-py"></a>

## evidence/ticket-13/commands.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/commands.py](../../barness-ai-glm/evidence/ticket-13/commands.py)；SHA256 `e5d8ba1bcdfeb478aa870caeac7201e7829c926fe732b79c9308fc53ff5de90a`。活动工具 / 仅测试消费，旧真实轮次关闭。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：Selected-secret subprocess isolation and finite command receipts。

**调用方**：[ai/e2e/glm_corrective_launcher_test.go](../../../ai/e2e/glm_corrective_launcher_test.go) L23（confirmed-public-Go-consumer）。

**输入**：ticket13固定源码/轮次目录、commands模块、单路线key、评审/构建/资格原件；普通Go测试复制到合成仓库。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次source/binaries/命令/账本/捕获/矩阵/门禁报告；当前测试只写私有fixture。实际静态写点：L26 `path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L85 `json.dump({'state': 'STARTING'}, stream)`；L99 `dump(receipt, record)`；L118 `(self.out / (name + '.log')).write_bytes(output)`；L119 `dump(receipt, record)`。

**副作用**：commands导入即查询Git common-dir；prepare/check/gate可构建测试，live可读.env并请求厂商，merge写矩阵；活动身份仅指现有普通拒绝测试，旧真实轮次不可重启。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L15 `Path(__file__).resolve()`；L16 `subprocess.check_output(['git', 'rev-parse', '--git-common-dir'], cwd=ROOT, text=True).strip()`；L16 `subprocess.check_output(['git', 'rev-parse', '--git-common-dir'], cwd=ROOT, text=True)`；L17 `(ROOT / COMMON).resolve()`。

<a id="barness-ai-glm--evidence--ticket-13--corrective-py"></a>

## evidence/ticket-13/corrective.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/corrective.py](../../barness-ai-glm/evidence/ticket-13/corrective.py)；SHA256 `3e325c2d93dbd10352d41310d5c391918b30bf2b46d65d0e2bf5e62c9c394402`。活动工具 / 仅测试消费，旧真实轮次关闭。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：One separately reserved 36-call source-corrective round; no prior balance。

**调用方**：[ai/e2e/glm_corrective_launcher_test.go](../../../ai/e2e/glm_corrective_launcher_test.go) L23（confirmed-public-Go-consumer）。

**输入**：ticket13固定源码/轮次目录、commands模块、单路线key、评审/构建/资格原件；普通Go测试复制到合成仓库。参数：`parser.add_argument('mode', choices=('plan', 'prepare', 'check', 'live', 'merge', 'gate', 'summarize'))` L160；`parser.add_argument('out')` L161。

**输出/写入**：原轮次source/binaries/命令/账本/捕获/矩阵/门禁报告；当前测试只写私有fixture。实际静态写点：L49 `self.out.mkdir(parents=True, exist_ok=False, mode=448)`；L57 `c.dump(self.out / 'source-at-start.json', dict(files=files, gitCommit=commit))`；L58 `c.dump(self.out / 'authorization-scope.json', plan())`；L76 `c.dump(self.out / 'binaries.json', {n: c.digest(self.out / n) for n in ('live.test', 'chineseeval', 'visualeval')})`；L77 `c.dump(self.out / 'prepared.json', dict(state='PREPARED', credentialsRead=False))`；L111 `c.dump(self.out / 'consumption.json', dict(components=rows, total=totals))`；L142 `c.dump(self.out / 'live-summary.json', dict(commands=codes, total=self.summarize(), allCommandsSucceeded=passed))`；L66 `c.dump(self.out / ('visual-' + route + '-authorization.json'), proof)`；L121 `json.dump(dict(maxCalls=36, maxAttempts=36, automaticReruns=False), stream)`；L127 `c.dump(self.out / 'live-summary.json', dict(commands=codes, total=self.summarize(), allCommandsSucceeded=False))`；L69 `c.dump(self.out / 'protocol-authorization.json', proof)`。

**副作用**：commands导入即查询Git common-dir；prepare/check/gate可构建测试，live可读.env并请求厂商，merge写矩阵；活动身份仅指现有普通拒绝测试，旧真实轮次不可重启。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-glm--evidence--ticket-13--glm_launcher-py"></a>

## evidence/ticket-13/glm_launcher.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/glm_launcher.py](../../barness-ai-glm/evidence/ticket-13/glm_launcher.py)；SHA256 `c0b68699919fc2bc2dc9e833955cb914cf4225439a3b4f3e41a4f6c14845ed35`。活动工具 / 仅测试消费，旧真实轮次关闭。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：One fixed formal GLM route at the reviewed current source; one selected key。

**调用方**：[ai/e2e/glm_corrective_launcher_test.go](../../../ai/e2e/glm_corrective_launcher_test.go) L23（confirmed-public-Go-consumer）。

**输入**：ticket13固定源码/轮次目录、commands模块、单路线key、评审/构建/资格原件；普通Go测试复制到合成仓库。参数：`p.add_argument('out')` L17；`p.add_argument('route', choices=('chat', 'responses', 'messages'))` L17。

**输出/写入**：原轮次source/binaries/命令/账本/捕获/矩阵/门禁报告；当前测试只写私有fixture。实际静态写点：L23 `base.mkdir(exist_ok=False, mode=448)`；L33 `c.dump(base / 'receipt.json', receipt)`；L60 `c.dump(base / 'SHA256SUMS.json', inventory)`；L67 `c.dump(out / ('glm-' + a.route + '-summary.json'), summary)`；L57 `c.dump(base / 'receipt.json', receipt)`。

**副作用**：commands导入即查询Git common-dir；prepare/check/gate可构建测试，live可读.env并请求厂商，merge写矩阵；活动身份仅指现有普通拒绝测试，旧真实轮次不可重启。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-glm--evidence--ticket-13--index-py"></a>

## evidence/ticket-13/index.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/index.py](../../barness-ai-glm/evidence/ticket-13/index.py)；SHA256 `bff632c714b89841a0cea60e0bbafbc0021e0a264c3a1ae9b021b1b388f03f3e`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Index sealed current release evidence and record owned temporary cleanup。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L45 `shutil.copyfile(before_path, E / 'user-untracked-before.json')`；L46 `dump(E / 'user-untracked-after.json', after)`；L52 `link.unlink()`；L58 `dump(E / 'owned-cleanup.json', dict(oracle=oracle, ownedPythonCacheRemoved=not cache.exists(), userUntrackedFiles=len(after), userUntrackedBytesUnchanged=True, rootHEAD=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=R, text=True).strip(), rootBranch=subprocess.check_output(['git', 'branch', '--show-current'], cwd=R, text=True).strip(), r`；L71 `dump(E / 'final-source-applicability.json', dict(finalExecutionCommit=seal['sourceCommit'], completeFrozenFiles=len(freeze['files']), allFrozenBytesPreserved=True, sourceFreezeSHA256=digest(R / '.evidence/glm13-corrective-current-r1/source-at-start.json'), measuredGLMCommit='34853063396ad58476e77dfe52463b3a9154ee77', glmSupportConsumerCommit='adc69`；L86 `shutil.copyfile(Path(__file__), D / 'index.py')`；L87 `shutil.copyfile(D / 'seal.py', E / 'seal.py')`；L88 `shutil.copyfile(D / 'reconcile-rollback-history.py', E / 'reconcile-rollback-history.py')`；L89 `shutil.copyfile(D / 'index.py', E / 'index.py')`；L90 `shutil.copyfile(D / 'preserve-history.py', E / 'preserve-history.py')`；L91 `shutil.copyfile(D / 'preserve-default-history.py', E / 'preserve-default-history.py')`；L101 `dump(E / 'durable-evidence-files.json', pins)`；L149 `dump(E / 'final-index.json', record)`；L150 `dump(D / 'index.json', record)`；L28 `path.write_text(json.dumps(value, indent=2) + '\n')`；L55 `shutil.rmtree(cache)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L9 `W = Path('/tmp/barness-glm-13')`；L10 `R = Path('/Users/cyber/RestoX/barness')`；L41 `before_path = Path('/tmp/barness-glm-user-untracked-before.json')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L34 `subprocess.check_output(['git', 'ls-files', 'ai'], cwd=W, text=True).splitlines()`；L34 `subprocess.check_output(['git', 'ls-files', 'ai'], cwd=W, text=True)`；L43 `(R / p).stat()`；L46 `dump(E / 'user-untracked-after.json', after)`；L52 `link.unlink()`。

<a id="barness-ai-glm--evidence--ticket-13--preserve-default-history-py"></a>

## evidence/ticket-13/preserve-default-history.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/preserve-default-history.py](../../barness-ai-glm/evidence/ticket-13/preserve-default-history.py)；SHA256 `3935d945798b49abad78a025510eec5d8f2e3a318ebda709b1fa992a441c4c36`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Preserve owned default E2E outputs without touching ROOT's shared directory.  This separately sealed operation keeps the already executed 11-tree historical sealer and its hash immutable. Both receipts remain independently repeatable。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L55 `out.mkdir(exist_ok=False)`；L57 `inv.write_text(json.dumps(before, indent=2) + '\n')`；L99 `(E / 'default-e2e-history-preservation.json').write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L13 `W = Path('/private/tmp/barness-glm-13')`；L14 `R = Path('/Users/cyber/RestoX/barness')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L48 `json.loads((W / '.evidence/glm13-corrective-current-r1/gate-summary.json').read_bytes())`；L48 `(W / '.evidence/glm13-corrective-current-r1/gate-summary.json').read_bytes()`；L50 `stat.S_IMODE(SOURCE.lstat().st_mode)`；L50 `SOURCE.lstat()`；L55 `out.mkdir(exist_ok=False)`。

<a id="barness-ai-glm--evidence--ticket-13--preserve-history-py"></a>

## evidence/ticket-13/preserve-history.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/preserve-history.py](../../barness-ai-glm/evidence/ticket-13/preserve-history.py)；SHA256 `4aee8abc94a6bee7f5ee8ab774937c39e30e60110cb374514ba1cedb3218cbdf`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Seal remaining ticket13 historical trees after the actual final CLI ends.  These are immutable failed/superseded receipts, never current release proof. Archive special negative fixtures as lstat metadata, without following links or opening FIFO/device/socket contents. Authoritative report validators are intact。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L70 `target.mkdir(exist_ok=True)`；L136 `dump(E / 'remaining-history-preservation.json', dict(schema=1, scope='all otherwise-unsealed ticket13 worker evidence trees; historical only', allWorkerGLM13TreesAccountedFor=True, finalReleaseEvidence=False, noLinkTargetsOrFIFOContentsRead=True, trees=records, separatelyPreserved=OTHER, tool=dict(source='preserve-history.py', sha256=digest(Path(__`；L41 `path.write_text(json.dumps(value, indent=2) + '\n')`；L79 `dump(inv, before)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L14 `W = Path('/private/tmp/barness-glm-13')`；L15 `R = Path('/Users/cyber/RestoX/barness')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L65 `json.loads((W / '.evidence/glm13-corrective-current-r1/gate-summary.json').read_bytes())`；L65 `(W / '.evidence/glm13-corrective-current-r1/gate-summary.json').read_bytes()`；L70 `target.mkdir(exist_ok=True)`；L74 `stat.S_IMODE(source.lstat().st_mode)`；L79 `dump(inv, before)`。

<a id="barness-ai-glm--evidence--ticket-13--reconcile-rollback-history-py"></a>

## evidence/ticket-13/reconcile-rollback-history.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/reconcile-rollback-history.py](../../barness-ai-glm/evidence/ticket-13/reconcile-rollback-history.py)；SHA256 `e12d17db0b8523bba86b6096a48402b285aceb8fd0fbbaa0dbd83caf6fa50f06`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：One-time read-only reconciliation of the baseline's selected release evidence。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L5 `OUT.mkdir(parents=True, exist_ok=False)`；L75 `dump(OUT / 'selected-file-hashes.json', verified)`；L76 `(OUT / 'baseline-evidence-index.json').write_bytes(raw)`；L77 `dump(OUT / 'reconciliation.json', dict(passVerdict=True, baselineCommit=BASE, baselineTree=subprocess.check_output(['git', 'rev-parse', BASE + '^{tree}'], cwd=ROOT, text=True).strip(), originalTrackedIndex=indexpath, trackedIndexSHA256=hashlib.sha256(raw).hexdigest(), releaseReport=dict(path=str((base / 'gate/release-report.json').relative_to(ROOT)`；L13 `p.write_text(json.dumps(v, indent=2) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L4 `ROOT=Path('/Users/cyber/RestoX/barness'); BASE='f5cf5362610a8bdb8f3e170f408f3858c149e540'`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L5 `OUT.mkdir(parents=True, exist_ok=False)`；L15 `json.loads(raw)`；L27 `parent.glob('*/manifest.json')`；L65 `subprocess.check_output(['git', 'ls-tree', '-r', '-z', source], cwd=ROOT)`；L69 `subprocess.run(['git', 'cat-file', '--batch'], cwd=ROOT, input=''.join((oids[n] + '\n' for n in names)).encode(), stdout=subprocess.PIPE, check=True)`。

<a id="barness-ai-glm--evidence--ticket-13--rollback-py"></a>

## evidence/ticket-13/rollback.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/rollback.py](../../barness-ai-glm/evidence/ticket-13/rollback.py)；SHA256 `a0a8a7b9d5966e3d0c491c9a525559cc217daba52c583c380ee7b03ca602ed89`。活动工具 / 现有公共消费者。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：Verify and execute a complete baseline restoration outside all worktrees。

**调用方**：[ai/e2e/glm_corrective_launcher_test.go](../../../ai/e2e/glm_corrective_launcher_test.go) L23（confirmed-public-Go-consumer）；[ai/e2e/glm_full_rollback_test.go](../../../ai/e2e/glm_full_rollback_test.go) L30（confirmed-public-Go-consumer）。

**输入**：CLI新仓库相对out、完整Git基线对象/树、rollback-probe.txt；GLM还需commands.py模块。参数：`p.add_argument('out')` L87。

**输出/写入**：out内baseline.tar/完整blob与mode库存/探针二进制/命令日志/回滚报告；系统临时目录中的恢复树。实际静态写点：L89 `out.mkdir(parents=True, exist_ok=False, mode=448)`；L90 `c.dump(out / 'root-before.json', before)`；L125 `c.dump(out / 'root-after.json', after)`；L127 `c.dump(out / 'rollback-report.json', dict(passVerdict=True, baselineCommit=BASELINE, baselineTree=TREE, archiveSHA256=c.digest(archive), completeTrackedFiles=len(entries), restoredTreeVerified=True, publicNoGLMRefusal=True, oldProviderControlledE2E=True, rootUnchanged=True, ownedTemporaryRestorationRemoved=True, baselineBinarySHA256=c.digest(out / `；L96 `restored.mkdir()`；L103 `c.dump(out / 'restored-tree.json', inventory)`；L104 `probe.write_bytes((HERE / 'rollback-probe.txt').read_bytes())`；L119 `c.dump(out / 'commands.json', receipts)`；L121 `c.dump(out / 'module-pins.json', {p: c.digest(restored / p) for p in ('go.mod', 'go.sum')})`；L124 `c.dump(out / 'baseline-release-index.json', {p: inventory[p] for p in selected})`；L102 `tar.extractall(restored, filter='data')`；L115 `(out / (name + '.log')).write_text(receipt_text(proc.stdout.decode(errors='replace'), Path(temp)))`。

**副作用**：读取Git并导出旧完整树；解包临时恢复、go build及一个受控旧Provider E2E，清理自有临时树。是普通测试辅助，不是生产回滚操作或新的live资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[ai/e2e/glm_full_rollback_test.go](../../../ai/e2e/glm_full_rollback_test.go)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L18 `(c.ROOT / c.COMMON).resolve()`；L19 `Path(__file__).resolve()`。

<a id="barness-ai-glm--evidence--ticket-13--run-py"></a>

## evidence/ticket-13/run.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/run.py](../../barness-ai-glm/evidence/ticket-13/run.py)；SHA256 `97f7d6507c4c69040f35fd9db8c94cde456a74c978a78d0502b76b846c1dfd8c`。活动工具 / 仅测试消费，旧真实轮次关闭。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：One preregistered current-catalog legacy refresh; no previous-task balance。

**调用方**：[ai/e2e/glm_corrective_launcher_test.go](../../../ai/e2e/glm_corrective_launcher_test.go) L23（confirmed-public-Go-consumer）；[ai/e2e/glm_release_launcher_test.go](../../../ai/e2e/glm_release_launcher_test.go) L98（confirmed-public-Go-consumer）。

**输入**：ticket13固定源码/轮次目录、commands模块、单路线key、评审/构建/资格原件；普通Go测试复制到合成仓库。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'live', 'merge', 'gate', 'summarize'))` L207；`parser.add_argument('out')` L208。

**输出/写入**：原轮次source/binaries/命令/账本/捕获/矩阵/门禁报告；当前测试只写私有fixture。实际静态写点：L37 `self.out.mkdir(parents=True, exist_ok=False, mode=448)`；L45 `c.dump(self.out / 'source-at-start.json', dict(files=files, gitCommit=commit))`；L46 `c.dump(self.out / 'authorization-scope.json', dict(reference='barness-ai-glm-ticket13-r1', source='current implement-spec R21 and issue13; required current-hash release renewal', maxCalls=302, maxAttempts=302, previousTaskBalanceCarried=False, automaticReruns=False, providerCaps={'openai': 116, 'anthropic': 32, 'google': 36, 'deepseek': 64, 'typesa`；L70 `c.dump(self.out / 'binaries.json', {n: c.digest(self.out / n) for n in ('live.test', 'chineseeval', 'visualeval')})`；L71 `c.dump(self.out / 'prepared.json', dict(state='PREPARED', credentialsRead=False))`；L77 `c.dump(self.out / 'live-harness-accepted.json', dict(exitCode=code))`；L139 `c.dump(self.out / 'live-summary.json', dict(commands=codes, total=self.summarize(), allCommandsSucceeded=all((v == 0 for v in codes.values()))))`；L171 `c.dump(self.out / 'consumption.json', dict(routes=rows, total=total))`；L183 `(self.out / 'matrix-before.json').write_bytes(matrix.read_bytes())`；L185 `(self.out / 'matrix-after.json').write_bytes(matrix.read_bytes())`；L199 `c.dump(self.out / 'gate-summary.json', dict(exitCode=code, passVerdict=passed, reportAvailable=available))`；L60 `c.dump(self.out / ('visual-' + route + '-authorization.json'), proof)`；L121 `json.dump(dict(maxCalls=302, maxAttempts=302, automaticReruns=False), stream)`；L63 `c.dump(self.out / 'protocol-authorization.json', proof)`。

**副作用**：commands导入即查询Git common-dir；prepare/check/gate可构建测试，live可读.env并请求厂商，merge写矩阵；活动身份仅指现有普通拒绝测试，旧真实轮次不可重启。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现有合成测试/固定编译合同；真实旧轮次仍受原源码/历史/一次性授权pin约束；recovery7暂停优先。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="barness-ai-glm--evidence--ticket-13--seal-py"></a>

## evidence/ticket-13/seal.py

原件：[.scratch/barness-ai-glm/evidence/ticket-13/seal.py](../../barness-ai-glm/evidence/ticket-13/seal.py)；SHA256 `266eefafd7c3d2e429a0faf151f2352e9013ff70efa2283824ab63933490ca22`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Independent final evidence sealing; requires the actual complete gate PASS。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L239 `dump(OUT / 'corrective-paid-index.json', dict(sourceCommit=SOURCE, canonicalRoot=str(ROOT / COLLECTION), total=consumption, providerCaps={'typesafe': 24, 'openai': 12}, allNineProducerVerifierCommandsPass=True, cumulativeTicket13ActualHTTP=282, originalRoundActualHTTP=246, newCorrectiveHTTP=36, automaticReruns=False, fullCollection=trees['complete-`；L249 `dump(OUT / 'consumer-diagnostic-r2-index.json', dict(finalReleaseEvidence=False, sourceCommit=SOURCE, originalBundleSource='dbc6c4e048e8578997839134af2203f3d102ddeb', scope='NONFINAL leaf-ownership projections, not corrected actual test execution', canonicalRoot=str(projection), tree=trees['nonfinal-consumer-diagnostic'], pins={n: digest(projection`；L269 `dump(OUT / 'current-effect-summary.json', dict(sourceCommit=SOURCE, independentEffects=4, fixedDatasetOnly=True, newQualityThreshold=False, productionAccuracyClaim=False, effects=effects))`；L307 `dump(OUT / 'complete-trace-receipts.json', trace_receipt)`；L308 `shutil.copyfile(WORK / 'ai/release/catalog-snapshot.json', OUT / 'catalog-snapshot.json')`；L309 `shutil.copyfile(WORK / 'ai/live/support-matrix.json', OUT / 'support-matrix.json')`；L310 `shutil.copyfile(WORK / 'ai/live/testdata/glm-candidate/sources.json', OUT / 'glm-price-source-snapshot.json')`；L311 `shutil.copyfile(WORK / 'ai/live/testdata/glm-formal/known-usage.json', OUT / 'known-usage-snapshot.json')`；L312 `shutil.copyfile(gate / 'release-report.json', OUT / 'release-report.json')`；L313 `shutil.copyfile(gate / 'release-report.md', OUT / 'release-report.md')`；L314 `dump(OUT / 'current-release-seal.json', dict(passVerdict=True, sourceCommit=SOURCE, trees=trees, bundles=bundle_records, commands=report['commands'], gates=report['gates'], audits=len(report['audits']), snapshots=report['snapshots'], trace=dict(items=len(items), cases=len(required), allPass=True, receipt='complete-trace-receipts.json', sha256=diges`；L335 `shutil.copyfile(Path(__file__), OUT / Path(__file__).name)`；L34 `path.write_text(json.dumps(value, indent=2) + '\n')`；L227 `dump(inv, after)`；L223 `shutil.copytree(src, dst, copy_function=shutil.copy2, symlinks=True)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/barness-ai-glm/evidence/ticket-13`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**固定宿主路径**：L14 `WORK = Path('/tmp/barness-glm-13')`；L15 `ROOT = Path('/Users/cyber/RestoX/barness')`。只说明原轮次定位，不改成本机可执行入口。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L106 `seal['payloadPins'].items()`；L94 `owned_payloads.items()`；L182 `(gate / mode).glob('*/manifest.json')`；L227 `dump(inv, after)`；L230 `tarfile.open(archive, 'w:gz', compresslevel=1)`。
