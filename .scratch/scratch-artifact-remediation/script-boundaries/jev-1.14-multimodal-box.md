# jev-1.14-multimodal-box 脚本边界

[分类入口](../script-boundaries.md) · [逐文件机器清单](../script-boundaries.json)

本页是静态审查说明，不是执行脚本清单。路径/行号按审查Git 99d0749 的原字节解析；L号是源码行，压缩源码的L号指解压后的文本。CLI、模块、测试支持与顶层程序分开看。全部副作用描述为代码能力，本次均未运行。

| 文件 | 身份 | 使用边界 |
| --- | --- | --- |
| [14-remaining-real/recheck.py](#jev-1-14-multimodal-box--14-remaining-real--recheck-py) | 历史源码 | 仅原轮次 |
| [14-remaining-real/recover.py](#jev-1-14-multimodal-box--14-remaining-real--recover-py) | 历史源码 | 仅原轮次 |
| [14-remaining-real/run.py](#jev-1-14-multimodal-box--14-remaining-real--run-py) | 历史源码 | 仅原轮次 |
| [14-remaining-real/seal.py](#jev-1-14-multimodal-box--14-remaining-real--seal-py) | 历史源码 | 仅原轮次 |
| [14-typesafe-real/run.py](#jev-1-14-multimodal-box--14-typesafe-real--run-py) | 历史源码 | 仅原轮次 |
| [14-typesafe-real/seal.py](#jev-1-14-multimodal-box--14-typesafe-real--seal-py) | 历史源码 | 仅原轮次 |
| [native-snapshot-evidence/replay.py](#jev-1-14-multimodal-box--native-snapshot-evidence--replay-py) | 历史源码 | 仅原轮次 |
| [run-14-verification.py](#jev-1-14-multimodal-box--run-14-verification-py) | 历史源码 | 仅原轮次 |
| [seal-11-evidence.py](#jev-1-14-multimodal-box--seal-11-evidence-py) | 历史源码 | 仅原轮次 |
| [seal-12-evidence.py](#jev-1-14-multimodal-box--seal-12-evidence-py) | 历史源码 | 仅原轮次 |
| [seal-13-evidence.py](#jev-1-14-multimodal-box--seal-13-evidence-py) | 历史源码 | 仅原轮次 |
| [seal-14-evidence.py](#jev-1-14-multimodal-box--seal-14-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-02-evidence.py](#jev-1-14-multimodal-box--verify-02-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-03-evidence.py](#jev-1-14-multimodal-box--verify-03-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-04-evidence.py](#jev-1-14-multimodal-box--verify-04-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-05-evidence.py](#jev-1-14-multimodal-box--verify-05-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-06-evidence.py](#jev-1-14-multimodal-box--verify-06-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-07-evidence.py](#jev-1-14-multimodal-box--verify-07-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-08-evidence.py](#jev-1-14-multimodal-box--verify-08-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-09-evidence.py](#jev-1-14-multimodal-box--verify-09-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-10-evidence.py](#jev-1-14-multimodal-box--verify-10-evidence-py) | 历史源码 | 仅原轮次 |
| [verify-11-evidence.py](#jev-1-14-multimodal-box--verify-11-evidence-py) | 活动工具 | 现有公共消费者 |

<a id="jev-1-14-multimodal-box--14-remaining-real--recheck-py"></a>

## 14-remaining-real/recheck.py

原件：[.scratch/jev-1.14-multimodal-box/14-remaining-real/recheck.py](../../jev-1.14-multimodal-box/14-remaining-real/recheck.py)；SHA256 `4d69c7d21896f4b3692c55bc180de2b8c04c4ae16a1eaf2e1144d541551a4fa6`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：One same-source DeepSeek Chat confirmation inside unchanged task caps。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L55 `r.dump(job.out / 'task-consumption-final.json', total)`；L58 `r.dump(job.out / 'accepted-deepseek-chat.json', dict(accepting=accepting, selectedBundle=str(r.bundle(job.out / 'deepseek-chat-recheck').relative_to(r.ROOT)), originalFailedBundle=str(r.bundle(job.out / 'live/deepseek-chat').relative_to(r.ROOT)), calls=calls, attempts=attempts, oldFailureRetained=True))`；L38 `json.dump(dict(maxCalls=32, maxAttempts=32, taskMaxCalls=550, providerMaxCalls=128, previousConsumption=before, previousProviderCalls=deepseek, previousProviderAttempts=deepseek_attempts, noAutomaticResend=True, sourceUnchanged=True), marker, indent=2)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/14-remaining-real`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`；L12 `spec.loader.exec_module(recovery)`。

<a id="jev-1-14-multimodal-box--14-remaining-real--recover-py"></a>

## 14-remaining-real/recover.py

原件：[.scratch/jev-1.14-multimodal-box/14-remaining-real/recover.py](../../jev-1.14-multimodal-box/14-remaining-real/recover.py)；SHA256 `529039fc6ff6d27caa7ab2d25a00c5c2f24538fb6c8c73fd2d749786704d68a8`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：The one reserved post-fix full round; retain every initial verdict。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('mode', choices=('prepare', 'check', 'live', 'merge', 'gate'))` L238；`parser.add_argument('out')` L239。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L41 `self.out.mkdir(parents=True, exist_ok=False, mode=448)`；L47 `r.dump(self.out / 'authorization-scope.json', reservation | dict(reason='fixed optional direct transport; one full post-fix evidence round', providerCaps={'openai': 116, 'anthropic': 32, 'google': 36, 'deepseek': 64, 'typesafe': 54}, accountRegionPermissions='UNVERIFIED', dollarCap=None, proxy=None))`；L51 `r.dump(self.out / 'source-at-start.json', dict(files=r.sources(), gitCommit=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, env=r.ENV, text=True).strip()))`；L72 `r.dump(self.out / 'prepared.json', dict(state='PREPARED', credentialsRead=False))`；L79 `r.dump(self.out / 'live-harness-accepted.json', dict(command='live-harness', exitCode=code))`；L153 `r.dump(self.out / 'consumption-all.json', dict(routes=rows, total=total))`；L154 `r.dump(self.out / 'task-consumption.json', task)`；L188 `r.dump(self.out / 'live-summary.json', dict(commands=codes, total=total, allCommandsSucceeded=passed))`；L216 `(self.out / 'matrix-before.json').write_bytes(matrix.read_bytes())`；L218 `(self.out / 'matrix-after.json').write_bytes(matrix.read_bytes())`；L231 `r.dump(self.out / 'gate-summary.json', dict(exitCode=code, passVerdict=r.doc(self.out / 'gate/release-report.json')['pass']))`；L46 `json.dump(reservation, marker, indent=2)`；L60 `r.dump(self.out / ('visual-' + route + '-authorization.json'), proof)`；L164 `json.dump(dict(maxCalls=302, maxAttempts=302, noAutomaticResend=True), marker)`；L214 `json.dump({'wholeBatchAtomic': True}, marker)`；L64 `r.dump(self.out / 'protocol-authorization.json', proof)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/14-remaining-real`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`；L14 `spec.loader.exec_module(r)`。

<a id="jev-1-14-multimodal-box--14-remaining-real--run-py"></a>

## 14-remaining-real/run.py

原件：[.scratch/jev-1.14-multimodal-box/14-remaining-real/run.py](../../jev-1.14-multimodal-box/14-remaining-real/run.py)；SHA256 `6a4e219ff600616e958a42e39f929161ae667aedfb778a628a05660564ec366d`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：One bounded remaining-provider round; credentials never enter argv or builds。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('mode', choices=('prepare', 'live', 'merge', 'gate'))` L291；`parser.add_argument('out')` L292。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L39 `path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L150 `self.out.mkdir(parents=True, exist_ok=False, mode=448)`；L151 `dump(self.out / 'source-at-start.json', dict(files=sources(), gitCommit=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, env=ENV, text=True).strip()))`；L154 `dump(self.out / 'authorization-scope.json', dict(reference='ticket14-remaining-20261017-r1', source='direct user instruction; selected root .env variables', initialMaxCalls=248, initialMaxAttempts=248, taskMaxCalls=550, taskMaxAttempts=550, recoveryMaxCalls=302, recoveryRounds=1, providerInitialCaps={'openai': 116, 'anthropic': 32, 'google': 36, 'd`；L161 `dump(self.out / 'visual-authorization.json', dict(route='decisions', model='gpt-6-luna', service='https://api.openai.com/v1/decisions', catalogHash=doc(ROOT / 'ai/release/catalog-snapshot.json')['hash'], accountAlias='openai-local-unverified', regionAlias='unverified', reference='ticket14-remaining-20261017-r1', dedicatedCredential=True, maxCalls=1`；L173 `dump(self.out / 'prepared.json', dict(state='PREPARED', credentialsRead=False))`；L211 `dump(self.out / 'consumption.json', dict(routes=rows, total=totals))`；L248 `dump(self.out / 'live-summary.json', dict(commands=codes, total=totals, allCommandsSucceeded=all((v == 0 for v in codes.values()))))`；L256 `(self.out / 'matrix-before.json').write_bytes((ROOT / 'ai/live/support-matrix.json').read_bytes())`；L270 `(self.out / 'matrix-after.json').write_bytes((ROOT / 'ai/live/support-matrix.json').read_bytes())`；L271 `dump(self.out / 'merge-summary.json', dict(commands=codes, allAccepted=len(codes) == 9 and all((v == 0 for v in codes.values()))))`；L284 `dump(self.out / 'gate-summary.json', dict(exitCode=code, passVerdict=doc(self.out / 'gate/release-report.json')['pass']))`；L104 `json.dump({'state': 'STARTING'}, intent)`；L219 `json.dump(dict(maxCalls=248, maxAttempts=248, noAutomaticResend=True), marker)`；L255 `json.dump({'noAutomaticReplay': True}, marker)`；L120 `dump(receipt, record)`；L141 `(self.out / (name + '.log')).write_bytes(output)`；L143 `dump(receipt, record)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/14-remaining-real`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L18 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--14-remaining-real--seal-py"></a>

## 14-remaining-real/seal.py

原件：[.scratch/jev-1.14-multimodal-box/14-remaining-real/seal.py](../../jev-1.14-multimodal-box/14-remaining-real/seal.py)；SHA256 `73945c94dcb56844db20ed72e50f21ba8a10fb238565572144d0b7880adc7b07`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Seal full release evidence, initial failures and the single paid recovery。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('--record', help='completed recovery root relative to repository')` L325；`parser.add_argument('--verify', action='store_true')` L326。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L194 `INDEXES.mkdir(exist_ok=True)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/14-remaining-real`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--14-typesafe-real--run-py"></a>

## 14-typesafe-real/run.py

原件：[.scratch/jev-1.14-multimodal-box/14-typesafe-real/run.py](../../jev-1.14-multimodal-box/14-typesafe-real/run.py)；SHA256 `a7055f21bb4f03409a799308ec3df57f400bc82278eeae0819f980e0a8661738`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：One authorized TypeSafe round; no credential inheritance or paid retries。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('mode', choices=('prepare', 'live', 'gate'))` L19；`parser.add_argument('out', help='new repository-relative evidence directory')` L20；`parser.add_argument('--authorization-reference', default='ticket14-user-typesafe-20261017')` L21。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L30 `path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')`；L132 `OUT.mkdir(parents=True, exist_ok=False, mode=448)`；L133 `dump(OUT / 'source-at-start.json', {'gitCommit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, env=ENV, text=True).strip(), 'files': sources()})`；L140 `dump(OUT / 'protocol-authorization.json', proof)`；L143 `dump(OUT / 'visual-authorization.json', proof)`；L144 `dump(OUT / 'authorization-scope.json', dict(source='user instruction; root .env:TYPESAFE_KEY', providers=['typesafe'], models=['jev-1.13.0', 'jev-1.14.0'], maxCalls=54, maxAttempts=54, componentCalls={'protocol113': 6, 'text113': 24, 'protocol114': 12, 'visual114': 12}, accountRegionPermissions='UNVERIFIED', dollarCap=None, proxy=None, note='Dedica`；L156 `dump(OUT / 'prepared.json', {'state': 'PREPARED', 'noCredentialsRead': True})`；L209 `dump(OUT / 'live-summary.json', {'commands': codes, 'allCommandsSucceeded': all((v == 0 for v in codes.values())), 'noOtherProviderCredentialRead': True, 'accountRegionPermissions': 'UNVERIFIED'})`；L240 `dump(OUT / 'gate-inputs.json', {'live': [str(p.relative_to(ROOT)) for p in inputs], 'realTypeSafeOnly': True, 'overall': report['pass'], 'collectionExitDoesNotMeanReleasePass': True})`；L67 `json.dump({'state': 'STARTING'}, intent)`；L165 `json.dump({'state': 'STARTED', 'maxCalls': 54, 'noAutomaticRerun': True}, marker)`；L206 `(OUT / 'matrix-before.json').write_bytes((ROOT / 'ai/live/support-matrix.json').read_bytes())`；L208 `(OUT / 'matrix-after.json').write_bytes((ROOT / 'ai/live/support-matrix.json').read_bytes())`；L218 `json.dump({'state': 'STARTED', 'providerCredentials': False}, marker)`；L77 `dump(receipt, record)`；L99 `(OUT / (name + '.log')).write_bytes(output)`；L101 `dump(receipt, record)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/14-typesafe-real`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L16 `Path(__file__).resolve()`；L22 `parser.parse_args()`。

<a id="jev-1-14-multimodal-box--14-typesafe-real--seal-py"></a>

## 14-typesafe-real/seal.py

原件：[.scratch/jev-1.14-multimodal-box/14-typesafe-real/seal.py](../../jev-1.14-multimodal-box/14-typesafe-real/seal.py)；SHA256 `7fc03eccad216915859f68967ebc81c22ad2345b60a3c3ec88aec4b009c3803a`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Seal the authorized TypeSafe round without changing earlier evidence。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`g.add_argument('--record', action='store_true')` L401；`g.add_argument('--verify', action='store_true')` L402。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L259 `index.parent.mkdir(exist_ok=True)`；L339 `RECEIPT.write_text('{\n' + ',\n'.join(('  ' + json.dumps(k) + ': ' + json.dumps(v, ensure_ascii=False, separators=(',', ':')) for k, v in receipt.items())) + '\n}\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/14-typesafe-real`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L12 `Path(__file__).resolve()`；L13 `Path(__file__).resolve()`；L23 `spec.loader.exec_module(b)`。

<a id="jev-1-14-multimodal-box--native-snapshot-evidence--replay-py"></a>

## native-snapshot-evidence/replay.py

原件：[.scratch/jev-1.14-multimodal-box/native-snapshot-evidence/replay.py](../../jev-1.14-multimodal-box/native-snapshot-evidence/replay.py)；SHA256 `fdae89cea5baff52090afca6a828fbdd8b612a1c29818862f5e11eaa041b18fa`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Replay issue 01 offline; preserve each run and hash every delivered file。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('--out', type=Path, required=True)` L59；`parser.add_argument('--phase', choices=('baseline', 'final'), required=True)` L60；`parser.add_argument('--baseline', type=Path)` L61；`parser.add_argument('--finish-only', action='store_true', help='seal already completed commands without rerunning tests')` L62。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L21 `path.write_text(json.dumps(data, indent=2, sort_keys=True) + '\n')`；L71 `out.mkdir(parents=True, exist_ok=False)`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：所属工作流可启动live或转发所选凭据；仅合成/暂停/历史身份不授予真实执行；具体模式见argv；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box/native-snapshot-evidence`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L11 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--run-14-verification-py"></a>

## run-14-verification.py

原件：[.scratch/jev-1.14-multimodal-box/run-14-verification.py](../../jev-1.14-multimodal-box/run-14-verification.py)；SHA256 `0f33f7cc1cd97ffb42c0566acd918081d6f64cfc5162df3b7e77884f562e049a`。历史源码 / 仅原轮次。已完成轮次的原执行器，无当前日常调用接线；可能隐式构建/整轮/付费。

**职责**：Reproduce ticket 14 without inheriting route credentials or making paid calls。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已固定的命令/路线/源码与out；可含--live/--gate/--mode等，旧参数不构成新授权。参数：`parser.add_argument('out', help='new repository-relative output directory')` L16；`parser.add_argument('--command-timeout', type=float, default=3600, help='positive overall command cap in seconds (at most 3600)')` L17；`parser.add_argument('--gate', action='store_true', help='run full release command after preflights')` L18。

**输出/写入**：原轮次命令日志/编译物/原生包/矩阵或门禁报告，部分路径固定不可复用。实际静态写点：L22 `OUT.mkdir(parents=True, exist_ok=False)`；L108 `(OUT / 'independent-notrun-verification.json').write_text(json.dumps(independent, indent=2) + '\n')`；L63 `(OUT / 'commands.json').write_text(json.dumps(commands, indent=2) + '\n')`；L136 `(OUT / 'multimodal-capacity-summary.json').write_text(json.dumps(capacity, indent=2) + '\n')`。

**副作用**：执行构建/测试/消费者，可能启用旧差分、race、压力；含live时可能读所选.env并付费。零凭据不等于低成本/只读，不作为日常入口。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：原实现清除路线凭据，生成NOT_RUN预检；仍会编译/重负载，不能作为轻量日常入口；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L14 `Path(__file__).resolve()`；L19 `parser.parse_args()`；L22 `OUT.mkdir(parents=True, exist_ok=False)`；L69 `execute('build-live', ['go', 'test', '-tags', 'live', '-c', '-o', OUT / 'live.test', './ai/live'])`；L70 `execute('build-text-effects', ['go', 'build', '-o', OUT / 'chineseeval', './ai/examples/chineseeval/cmd/chineseeval'])`。

<a id="jev-1-14-multimodal-box--seal-11-evidence-py"></a>

## seal-11-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/seal-11-evidence.py](../../jev-1.14-multimodal-box/seal-11-evidence.py)；SHA256 `3cad97b185979fceaa4e4e3e51302fe513ea8bc84f0877788398f254c3a6b6aa`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Recheck and seal the completed ticket-11 runs from the repository root.  Only local artifacts are read. Reconstructed child packages retain original bytes/timestamps and refuse differing existing files. The four legacy failure fixtures record pre-redaction input hashes; final artifact hashes seal their redacted copies independently. New runs must retain new IDs, never edit these。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L53 `dest.mkdir(parents=True, exist_ok=True)`；L98 `output_path.write_text(encoded)`；L58 `p.parent.mkdir(parents=True, exist_ok=True)`；L67 `p.write_bytes(raw)`；L62 `p.write_bytes(data)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L15 `hashlib.sha256(p.read_bytes()).hexdigest()`；L15 `p.read_bytes()`；L75 `receipt['runs'].append(inventory(ROOT / '.evidence/barness-ai/20261016T115408.429237000Z'))`；L78 `receipt['runs'].append(inventory(root))`；L83 `json.loads(safe_path(root, 'multimodal-report.json').read_bytes())`。

<a id="jev-1-14-multimodal-box--seal-12-evidence-py"></a>

## seal-12-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/seal-12-evidence.py](../../jev-1.14-multimodal-box/seal-12-evidence.py)；SHA256 `2753c08906731de5cacbecde9c2b0368fea1980d70e7bfa39280a10a3d6de147`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Verify ticket-12 artifacts and write an append-only receipt, from repo root。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('--run', action='append', required=True, type=Path)` L82；`parser.add_argument('--live', action='append', required=True, type=Path)` L83；`parser.add_argument('--out', required=True, type=Path)` L84。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="jev-1-14-multimodal-box--seal-13-evidence-py"></a>

## seal-13-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/seal-13-evidence.py](../../jev-1.14-multimodal-box/seal-13-evidence.py)；SHA256 `425d6fb7cf14a28de9f28c932f1d1cea865a738f28dff031e21229c1be917129`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Seal or recheck ticket-13 receipts from the repository root, without keys。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('--checks', type=Path)` L59；`parser.add_argument('--formal', type=Path)` L60；`parser.add_argument('--out', type=Path)` L61；`parser.add_argument('--verify', type=Path)` L62。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L170 `args.out.write_bytes(encoded)`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

<a id="jev-1-14-multimodal-box--seal-14-evidence-py"></a>

## seal-14-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/seal-14-evidence.py](../../jev-1.14-multimodal-box/seal-14-evidence.py)；SHA256 `fb05a91c9f3e26d7fc98a72b18831427119b61293ff215db5a5228314020eb0d`。历史源码 / 仅原轮次。原已关闭执行的整理/封存/派生源码，不重新封存或变更pin。

**职责**：Seal/recheck ticket-14 evidence; never turn a failed release into PASS。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：原轮次已结束的执行目录/清单/日志与固定来源；record/freeze/capture不是重新测试授权。参数：`parser.add_argument('--record', help='final verification base, relative to repository')` L111；`parser.add_argument('--verify', action='store_true')` L112。

**输出/写入**：原脚本常量或CLI指定的归档、inventory/index/SHA256SUMS/receipt及派生快照。实际静态写点：L167 `indexes.mkdir(exist_ok=True)`；L219 `RECEIPT.write_text('{\n' + ',\n'.join(('  ' + json.dumps(k) + ': ' + json.dumps(v, ensure_ascii=False, separators=(',', ':')) for k, v in receipt.items())) + '\n}\n')`；L178 `index.write_text(json.dumps(files, sort_keys=True, separators=(',', ':')) + '\n')`。

**副作用**：创建/复制/压缩/封存文件，某些脚本覆盖同目录索引或启动Go审计/消费者；不宜在原件目录重跑。完整写点和子命令见逐文件数据。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：存在压力/负载材料引用或选择线索，须检查具体argv。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-02-evidence-py"></a>

## verify-02-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-02-evidence.py](../../jev-1.14-multimodal-box/verify-02-evidence.py)；SHA256 `90623f580ca9bc923ebdc6f89baf1a60e5cacf1f9052b3e60030f0f74d52a4b1`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify ticket 02's recorded source and offline evidence SHA-256 values。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L9 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-03-evidence-py"></a>

## verify-03-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-03-evidence.py](../../jev-1.14-multimodal-box/verify-03-evidence.py)；SHA256 `f438aa68f54945325ff4156d42b12d862d609fce2a7a9d63a868bc4dac4e2291`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Verify ticket 03 source, artifact inventory, audit and final image identities。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：存在网络/loopback API静态调用，见networkSites；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-04-evidence-py"></a>

## verify-04-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-04-evidence.py](../../jev-1.14-multimodal-box/verify-04-evidence.py)；SHA256 `24b7b946f1ee23dca75506fc0427446ed24570ef5aadad82f1d7a769d72f89f1`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Record/verify ticket 04 artifacts and recompute final native image identities.  Usage: --record FULL_MANIFEST RACE_MANIFEST to seal a completed run; no arguments verifies the sealed inventory. The bundles remain local under ignored .evidence。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L115 `REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L15 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-05-evidence-py"></a>

## verify-05-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-05-evidence.py](../../jev-1.14-multimodal-box/verify-05-evidence.py)；SHA256 `567d96048287bb3cf728b5aca3f0745f1d7ce8aa8a07e1a8e842e98ebabcfc2f`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Record/verify ticket 05 artifacts and recompute final native image identities.  Usage: --record FULL_MANIFEST RACE_MANIFEST to seal a completed run; no arguments verifies the sealed inventory. The bundles remain local under ignored .evidence。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L134 `REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L15 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-06-evidence-py"></a>

## verify-06-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-06-evidence.py](../../jev-1.14-multimodal-box/verify-06-evidence.py)；SHA256 `5fe66e1e31cb2a9f7b7f95346f16f99500868e0a64527c8b30b10625ff3a9d9b`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Record/verify ticket 06 artifacts and recompute final native image identities.  Usage: --record FULL_MANIFEST RACE_MANIFEST to seal a completed run; no arguments verifies the sealed inventory. The bundles remain local under ignored .evidence. This verifier deliberately preserves the prior ticket's inventory schema: each ticket keeps its own historical seal while its final traversal evolves here。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L145 `REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-07-evidence-py"></a>

## verify-07-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-07-evidence.py](../../jev-1.14-multimodal-box/verify-07-evidence.py)；SHA256 `7dbc5dbef599bffed117d6206a23fd0444bbc5606f343335bcb054bedb2ee3be`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Record/verify ticket 07 artifacts and recompute final native image identities.  Usage: --record FULL_MANIFEST RACE_MANIFEST to seal a completed run; no arguments verifies the sealed inventory. The bundles remain local under ignored .evidence. This verifier deliberately preserves the prior ticket's inventory schema: each ticket keeps its own historical seal while its final traversal evolves here。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L166 `REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-08-evidence-py"></a>

## verify-08-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-08-evidence.py](../../jev-1.14-multimodal-box/verify-08-evidence.py)；SHA256 `173e46bd855316837daf53898b3e9e211c4f414d939f2c3a66fc5fd2b66ccffb`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Seal/verify ticket 08 controlled-HTTP evidence. --record FULL RACE manifests.  Ticket 07's identity traversal remains a historical implementation; import its pure identity/hash helpers so this ticket can add Box checks without modifying an already sealed verifier or duplicating image traversal rules。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L160 `REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：存在网络/loopback API静态调用，见networkSites；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `Path(__file__).resolve()`；L23 `spec.loader.exec_module(prior)`。

<a id="jev-1-14-multimodal-box--verify-09-evidence-py"></a>

## verify-09-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-09-evidence.py](../../jev-1.14-multimodal-box/verify-09-evidence.py)；SHA256 `edf63f0568363268143c0d695bde04aaaf36b2c5e18981e1f1574e94399df698`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Seal/verify ticket 09 evidence: --record FULL_MANIFEST RACE_MANIFEST.  The ignored local bundles remain reproducible with each case's replay command. Historical failures are sealed separately and never converted to passing runs。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：无argparse声明；常量/父模块/顶层参数见原件。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L217 `REPORT.write_text(json.dumps(report, indent=2, sort_keys=True) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：存在网络/loopback API静态调用，见networkSites；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件有原生子进程/execute调用，详见processSites；race：存在race字面范围，必须按实际模式/argv另选，不代表已运行；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L17 `Path(__file__).resolve()`；L23 `spec.loader.exec_module(prior)`。

<a id="jev-1-14-multimodal-box--verify-10-evidence-py"></a>

## verify-10-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-10-evidence.py](../../jev-1.14-multimodal-box/verify-10-evidence.py)；SHA256 `470f72d0c0c65e715deabca9232fc02dddc1db1f36dc1cf1ea5b8dcd5b3a84e1`。历史源码 / 仅原轮次。按原固定材料消费的历史核验/辅助源码，无当前执行接线。

**职责**：Seal/recheck ticket 10's public E2E and standalone Host bundles, offline。

**调用方**：未发现现行Go/SOP/合成工具测试的执行接线；原历史调用/导入在JSON的historicalCallers/localReferences中保留。

**输入**：本文件原轮次index/manifest/归档/哈希、可选external-map/common-root等参数；按原pins读取，不读取为当前授权。参数：`parser.add_argument('--record', nargs='+', type=Path)` L124。

**输出/写入**：原轮次完整性/出处/质量核验与stdout/退出；存在写入点时另列具体表达式，不能仅凭verify名假定只读。实际静态写点：L138 `RECEIPT.write_text(json.dumps(receipt, indent=2) + '\n')`。

**副作用**：范围限原封存材料；不执行捕获命令。下列原生子进程/写点是静态能力，旧PASS/FAIL只适用原受测来源，不产生当前发布资格。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：仅原轮次/原固定源码或归档，须逐项满足脚本内pins及原索引；未核实当前适用，不因本轮Git/hash登记续期。当前替代入口：[docs/agents/verification.md](../../../docs/agents/verification.md)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。

**导入边界**：有顶层路径/Git/读写/参数候选；不做import或help探针。L10 `Path(__file__).resolve()`。

<a id="jev-1-14-multimodal-box--verify-11-evidence-py"></a>

## verify-11-evidence.py

原件：[.scratch/jev-1.14-multimodal-box/verify-11-evidence.py](../../jev-1.14-multimodal-box/verify-11-evidence.py)；SHA256 `0ad4b561cfe0407d8ee2943aa6543d4c56ea0f0e895e772ff111d4699a78f08e`。活动工具 / 现有公共消费者。Go公共测试或原生消费者仍有实际执行/复制接线，目录年代不表示退役。

**职责**：Independently read candidate or explicitly selected formal protocol structure.  No network, credentials, or working-source equality required for historical runs. Use --require-live to reject controlled HTTP packages. NOT_RUN/FAIL exit nonzero. Formal support additionally requires the public Classify replay and independent effect verification performed by the jevformal command; this is one stage only。

**调用方**：[ai/internal/testkit/jevformal/verify.go](../../../ai/internal/testkit/jevformal/verify.go) L203（confirmed-public-Go-consumer）；[ai/live/jev_formal_process_test.go](../../../ai/live/jev_formal_process_test.go) L53（confirmed-public-Go-consumer）；[ai/live/jev_formal_tamper_test.go](../../../ai/live/jev_formal_tamper_test.go) L142（confirmed-public-Go-consumer）；[ai/live/multimodal_process_test.go](../../../ai/live/multimodal_process_test.go) L102（confirmed-public-Go-consumer）。

**输入**：候选protocol bundle，可选--formal-catalog与--require-live；必须符合原生结构、字段和来源规则。参数：`parser.add_argument('--require-live', action='store_true')` L356；`parser.add_argument('--formal-catalog', type=Path)` L357；`parser.add_argument('bundle', type=Path)` L358。

**输出/写入**：stdout协议结构核验/拒绝及退出；不写原生包。实际静态写点：本文件未观察到；模块间接能力见调用方与JSON。

**副作用**：纯Python本地读取/有限解码；无子进程/网络/凭据。独立Python通过不等于正式资格，Go jevformal还检来源、回放和效果消费者。

**执行能力**：network：本文件未观察到网络API；子进程/父工作流另看下项；vendor：本文件未观察到厂商生产入口；不等于对子进程的无网络保证；buildOrTest：本文件未观察到原生子进程调用；模块/动态加载看所属调用方；race：本文件未观察到race选择，子命令边界另核对；pressure：本文件未观察到压力选择，子命令边界另核对。

**来源/替代**：现行SOP与实际调用方选择的完整来源/输入；本轮文件哈希只固定审查对象，不替代原生资格。当前替代入口：[ai/live/cmd/jevformal/main.go](../../../ai/live/cmd/jevformal/main.go)；活动文件继续由现有调用方按SOP选定；历史生产/整轮命令由分阶段SOP替代，历史核验无通用等价替身，原件本身继续保留。原相对基址 `.scratch/jev-1.14-multimodal-box`；字面pins、相邻原索引及读/进程调用点逐项见JSON，不能用本轮哈希代替原资格。
