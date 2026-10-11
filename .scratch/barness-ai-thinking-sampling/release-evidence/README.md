# 当前候选版本的完整证据

采样实现已完成；正式发布仍受当前真实调用和独立效果失败阻断。
所有 PASS、FAIL、UNSUPPORTED、NOT_RUN 保留原状态，不用历史目录证据放行。

受测 commit：`dba212b095338ba61430bca48525d6cf4bf21ab3`。当前最终 gate 退出 1：

| 门禁 | 实际结果 |
| --- | --- |
| 完整普通 / race | PASS / PASS，各 6,125 case；272.638 / 970.760 秒 |
| 两种 vet、P0、差分、快照 | PASS；815 差分（含目录）、pending 0 |
| 十二份输入脱敏审计 | PASS，零发现 |
| 九组合 live | FAIL；DeepSeek 两条 PASS，其余七条 FAIL |
| 独立中文效果 / 必交 artifacts | FAIL；24/24 HTTP 451，0 可评分 |
| 追溯 | FAIL；依赖 live/effect 的验收未满足 |

因此 B01–B03 已满足，B04 与工单 06 保持开放。历史压力异常原因未确认，
工单 07 保持开放；当前同负载重跑通过不构成修复证明。


`manifest.json` 索引 12 个当前完整包：完整离线、完整 race、九个真实组合、
独立中文效果。每份 tar.gz 保存全部原始受控请求、帧、事件、结果、断言、
Attempt、pidiff、fixture、manifest、审计及报告。逐 case 原始字节与 Hash 不变；
tar 的主机所有者与时间归零，顶层失败命令报告及命令日志中的主机路径脱敏，展示日志去掉空白行的尾空格。
`source-hashes.json` 固定实际测试源码，`live-harness.json` 如实记录真实调用
二进制的原 SHA 与不含 VCS 字段的 buildInfo，不能证明构建 commit；manifest 的
commit 为运行时 checkout HEAD，没有改写成最终构建来源。普通/race 的源码内容
与 dba212b 的对应由逐文件 Hash 验证。

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

检查每份完整归档和逐 case artifacts，重新审计整个归档目录（含没有原 manifest
的中断 race，临时验证标记不会认证原运行），再按固定输入重算七个
证据门禁。三项命令门禁在重放中为 NOT_RUN，原始全套执行结果独立保留。
成功验证打印“recorded release FAIL”，表示证据完整而发布仍未通过。
`verification.json` 是实际复核收据；`FAILURES.md` 记录先行列出的失败边界。
本次完整复核退出 0，耗时 297.801 秒：23 份归档的整个目录树审计通过，
24 个 manifest bundle（含历史 race）逐 case Hash 和原判定一致，12 项负面
边界均拒绝，七个证据门禁重放一致。中文效果完整性检查退出 0、业务结果
仍为 FAIL；复核没有产生真实调用。
解析前检查原始 tar header 的总展开预算，仅接受本打包器产生的文件、目录和
单个 PAX path 记录；超大 PAX、global-PAX、GNU 扩展及扩展链的拒绝保留红绿证据。

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
；验证器会拒绝使用漂移源码重编工具解释旧证据。pack.py 固定本次失败快照的
输入与历史，是该交付包的重建脚本；后续发布需使用新运行的输入和原判定。
