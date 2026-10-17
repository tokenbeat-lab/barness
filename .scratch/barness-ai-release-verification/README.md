# 当前发布续修

用户授权修复 Decisions 清单和完整 normal/race 的实际失败，按原压力 #07 验收定位，
真实服务继续 NOT_RUN。不读取 `.env`、不继承路线凭据，不发起收费请求。

- `investigation.md`：实现前失败清单、假设、RED/GREEN、自然字节因果与保留的失败历史。
- `native-pressure/`：独立自然反例的有界诊断装置、600 Go 源文件守卫与诚实限制。
- `code-review.md`：两轴独立复核；诊断/封存脚本发现及修复也分别记录。
- `verification.md`、`evidence.json`：最终实际结论及完整性回执。
- `seal-evidence.py`：独立封存当前续修，不改写上轮 `14-evidence.json`。

完整门禁的零凭据重现入口：

```sh
python3 .scratch/jev-1.14-multimodal-box/run-14-verification.py .evidence/NEW_FOLLOWUP --gate
```

输出目录必须全新。它运行两种 vet、11 条路线预检、4 个效果包和独立 verifier、
live-tag 受控 harness，以及完整 normal/race、冻结 Pi 差分和原压力负载。
收集脚本退出 0 不是发布 PASS；必须检查 commands.json 和 gate/release-report.json。
真实支持和效果缺失时总体仍 BLOCKED，不借用历史包或修改旧失败结论。

最终封存后从仓库根目录校验：

```sh
python3 .scratch/barness-ai-release-verification/seal-evidence.py --verify
```

回执固定完整 Go 文件集合、配置/目录/Pi 账本、命令、各主包清单与审计、全部本次
历史目录和16个明确所属的原生诊断/RED目录的精确文件集合及存储字节 Hash。
只有证据根目录中的 live.test/chineseeval/visualeval 三个可重建执行文件排除；
记录在 receipt 中。故意损坏的嵌套 fixture 和链接仅作原始字节/链接目标身份保留，
绝不把它们递归识别为成功生产包。partial 和审计缺失仍保持原状态。

原压力工单仅在完整本版验收结束后转换；启动时的 blocker 快照和 pass=false 保留，
另记转换前后原文/Hash/时间。验收后的状态不能反向改写该次 CLI 报告。
