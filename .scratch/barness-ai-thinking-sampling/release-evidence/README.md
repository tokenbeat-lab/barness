# 当前候选版本的完整证据

采样实现已完成；正式发布仍受当前真实调用和独立效果失败阻断。
所有 PASS、FAIL、UNSUPPORTED、NOT_RUN 保留原状态，不用历史目录证据放行。

`manifest.json` 索引 12 个当前完整包：完整离线、完整 race、九个真实组合、
独立中文效果。每份 tar.gz 保存全部原始受控请求、帧、事件、结果、断言、
Attempt、pidiff、fixture、manifest、审计及报告；仅 tar 的主机所有者与时间归零。
`source-hashes.json` 固定实际测试源码，`live-harness.json` 如实记录真实调用
二进制的早期 Git revision 与 `vcs.modified=true`，没有改写报告里的 commit。

失败历史另有完整包：两次 catalog 红灯、快照/采样/duration 红灯、首次基线
迁移、新场景耗时差异、生命周期 fixture 错误，以及第一次全套离线失败、第二次完整 gate 的压力失败和 race 审计超时。
第一次 gate 的 race 被主动中断，原始文件保留但标为 NOT_RUN；它不能替代
最终完成的 race。原始命令日志中的本机路径替换为 `[REPO]` / `[HOME]`。
单独十轮压力复现的完整成功包另行归档；它未重现普通运行的自然失败，不能替代
失败 gate。S01–S30 与 R01–R12 的具体公共 case 索引在 scenario-artifacts.json。

验证与门禁重放不产生真实调用：

```sh
python3 .scratch/barness-ai-thinking-sampling/release-evidence/verify.py --fault-tests
```

检查每份完整归档和逐 case artifacts，重新运行审计，再按固定输入重算七个
证据门禁。三项命令门禁在重放中为 NOT_RUN，原始全套执行结果独立保留。
成功验证打印“recorded release FAIL”，表示证据完整而发布仍未通过。
`verification.json` 是实际复核收据；`FAILURES.md` 记录先行列出的失败边界。

重新生成一套当前结果（下面的 `--live` 会使用 root `.env` 凭据并产生付费调用）：

```sh
npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node
python3 .scratch/barness-ai-thinking-sampling/release-evidence/run-release.py \
  --live --out .evidence/barness-ai-thinking-sampling/fresh-release
```

每个真实 smoke 子进程只读取自己的一个凭据；独立效果进程只读取 TypeSafe。
必须使用新输出目录；账户别名明确标记区域/权限未确认。省略 `--live` 可重新
执行离线、压力、差分、vet 与 race，但没有新真实证据时不能通过整版发布。
新的真实报告按既有命令合并回各自 support-matrix 行；旧 FAIL 不被伪装为 PASS。

重建有限差分登记时，先解包完整 migration-red，然后显式传它的 manifest：

```sh
python3 .scratch/barness-ai-thinking-sampling/update-ledger.py \
  --migration <extracted-migration-red>/manifest.json
```

代码与验收两轴的独立审查见 `REVIEW.md`。源代码改动后必须重新运行完整门禁
和 pack.py；验证器会拒绝使用漂移源码重编工具解释旧证据。
