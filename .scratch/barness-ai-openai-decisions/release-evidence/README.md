# 工单 13：十组合发布证据

**最终交付：FAIL，工单 13 未完成。** [delivery-verdict.json](delivery-verdict.json)
保留[旧压力流完整性缺陷 07](../../barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)
的阻断：根因和修复尚未确认，不能用后续通过关闭。ADR-0026 保持 proposed，本票保持 claimed。

**本次自动门禁：PASS，实际退出 0。** [完整报告](release-report.md)、
[机器报告](release-report.json)、[命令](command.json)、[日志](command.log)记录实测结果；
自动检查通过不等于上述历史阻断已解决。

受测源码 `807ccbaa53b6764e2d03299cc582ef560fe87165`，正式目录 `2026-10-13.1`，
Hash `sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。

| 核验 | 实际结果 |
| --- | --- |
| normal 完整套件 | 7,616 PASS；其中普通离线 6,755，固定差分 861；362.250 秒 |
| race 完整套件 | 7,616 PASS；同样开启 Pi 差分和压力；1,225.044 秒 |
| 两种 vet | 普通及 `-tags live` 均 PASS |
| 固定合同 | 十组合、原六差分加 P10、原三 Extension Route、109 条追溯全部 PASS，pending 0 |
| 容量 | normal/race 各自验证旧混合负载与分类图片 count/decoded/JSON/headroom/heap/resources |
| 回滚 | normal/race 各六种初始/最终 Hook 撤销；新调用拒绝、已 pin 重试及费用保留、恢复与资源释放 |
| 真实 smoke | 十个自身 schema 3 完整报告均 PASS；OpenAI Chat 原有 reasoning-history 为明确 UNSUPPORTED |
| 审计 | 全部 14 包：normal、race、十个 live、两个 evaluation，均零发现 |
| 历史压力复现 | 原命令额外 10 轮 PASS，70.997 秒；未重现，不代表原因已确认或缺陷已修复 |

首次完整运行 normal/race 都因旧迁移测试错误比较“当前矩阵与迁移前历史”而 FAIL；
修复为比较保留的迁移快照，仍检查当前十组合身份。原报告及完整原包保存在 history*.zip。
没有改写失败为 PASS，也没有降低负载、删除旧协议或复用别的路线的报告。

## 独立效果与凭据边界

| 效果包 | 状态 | Calls / Attempts / retry | 正确率 | 已知估价 USD |
| --- | --- | --- | --- | --- |
| Jev，本轮新采集 | COMPLETE | 24 / 24 / 0 | 24/24 | 0.000539406 |
| Luna，工单 12 原始包 | COMPLETE | 24 / 24 / 0 | 24/24 | 0.000623800 |

这是固定 24 项任务的结果，不保证生产质量，估价不等于账单。门禁不设置业务质量阈值；
完整合法全拒答可导入，准确率 0，条件准确率/ECE/Brier 为 null。两路线不共享质量结论。
历史 Jev 的 recorded-snapshot 独立验证用途保留。

`collect.py` 从 `.env` 每次仅选取相应厂商 key，分别启动十个 smoke 和一个 Jev 进程；
每个进程只有自己的凭据变量。固定显式 loopback HTTP proxy `127.0.0.1:7890`，
官方 endpoint 与 TLS 校验不变。账户、区域、权限、key 是否专用、厂商支出限额均未独立核实，
如实记录 unknown。每次失败保留，输出目录不可复用。Luna 原包不修改 Hash 或时间。

保留原始完整 code-snapshot / CodeHash；gate 比较 runtime、评估引擎/CLI、requestid、
testkit 依赖及 go.mod/go.sum 的全部适用非测试源码集合，含新增/删除。测试、live harness、
release 工具及无关 examples 不影响模型执行适用性。任务/config/catalog 另行固定；
公共 verifier 重放捕获并核验 manifest、Observer、文件 Hash 和 audit。

## 归档与重放

[manifest.json](manifest.json)逐包保留原始 manifest/audit SHA、HEAD 标签、位置和状态。
`offline.zip` / `race.zip` 含全部 7,616 个 case；`live-evaluation.zip` 含十个完整 smoke
及 Jev/Luna；`history*.zip` 含首次完整失败、RED/GREEN、live harness 和压力复现。首次 normal/race
各自分包保留，以控制单个归档文件大小，未删任何 case。
`live-execution-source.zip` 对应真实采集时的工作树；`tested-source.zip` / `source-hashes.json`
保存最终受测 Go/JSON 源码与模块文件。HEAD 标签不冒充二进制构建证明；采集收据另存 binary SHA。
case 原始字节完全保留，仅顶层命令诊断中的本机路径替换为 `[REPO]` / `[HOME]`。

从仓库根核验与重放，不持有 key、不调用厂商：

```sh
python3 .scratch/barness-ai-openai-decisions/release-evidence/verify.py --replay
```

核验外层完整 inventory/SHA、原历史缺陷及失败归档 SHA、受测源码、全部原 manifest/audit，
再由公开发布 CLI 导入解压后的完整包，重放全部非命令 gate、追溯及证据状态。
`-bundle` / `-race-bundle` 不执行测试，命令 gate 按原契约为 NOT_RUN/FAIL；脚本核对这一差异，
绝不把重放当作再次运行 normal/race。`verification.json` 的成功表示归档一致，交付仍为 FAIL。

重新采集并执行真实门禁（厂商调用会产生费用）：

```sh
python3 .scratch/barness-ai-openai-decisions/release-evidence/collect.py --live --out .evidence/new-ten-route-live
python3 .scratch/barness-ai-openai-decisions/release-evidence/run-gate.py --collection .evidence/new-ten-route-live/collection.json --out .evidence/new-ten-route-gate
```

`run-gate.py` 无 key，执行两种 vet 和 normal/race 实际命令，均开启固定差分及压力。
历史缺陷必须独立解决后才能重新作出交付决定；单次自动报告 PASS 不解除该条件。

## 每票状态与证据

| 工单 | 状态 | artifact |
| --- | --- | --- |
| 01 Jev 基线 | resolved | [基线](../jev-baseline/README.md) |
| 02 文本 choice | resolved | [文本](../text-choice-evidence/README.md) |
| 03 混合答案/拒答 | resolved | [混合](../mixed-answer-evidence/README.md) |
| 04 usage/价格 | resolved | [用量](../usage-evidence/README.md) |
| 05 inline 图片 | resolved | [图片](../image-evidence/README.md) |
| 06 运行时隔离/撤销 | resolved | [运行时](../runtime-evidence/README.md) |
| 07 固定 Pi 差分 | resolved | [差分](../frozen-differential-evidence/README.md) |
| 08 分类图片负载 | resolved | [容量](../classifier-load-evidence/README.md) |
| 09 候选 live | resolved | [实际授权执行](../candidate-live-evidence/live-authorized-20261013/README.md) |
| 10 正式目录/矩阵 | resolved | [正式报告](../formal-support-evidence/README.md) |
| 11 Host/效果工具 | resolved | [公开工具](../host-tooling-evidence/README.md) |
| 12 Luna 真实效果 | resolved | [独立效果](../chinese-live-evidence/README.md) |
| 13 最终发布 | claimed / 未完成 | [本次实测](release-report.md)、[最终拒绝 verdict](delivery-verdict.json) |

两轴复审见 [REVIEW.md](REVIEW.md)：代码 Standards 0、Spec 0；历史交付阻断另行保留。
下一步须在原压力缺陷工单确认原因、形成先红后绿回归、保持原负载修复并重跑最终发布验收。
