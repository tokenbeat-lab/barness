# 12：正式 Luna 独立中文效果

2026-10-13 10:49:27–10:49:35（Asia/Shanghai）实际执行一次固定 `zh-support-v1`：
24 条、四类各六条合成中文支持请求。Luna 自己的 live 报告为 **COMPLETE**，协议状态
COMPLETE；这表示任务与证据完整，没有设置或降低准确率门槛，也不构成发布门禁通过。

| 本次结果 | 数值 |
|---|---:|
| 全任务准确率（正确 / 全部样本） | 24 / 24 = 100% |
| 普通答案条件准确率 | 24 / 24 = 100% |
| 拒答 / 失败 / 未执行 | 0 / 0 / 0 |
| 拒答率 | 0% |
| confidence 均值 / ECE | 0.9854166667 / 0.0145833333 |
| chosen probability 均值 / ECE | 0.9883333333 / 0.0116666667 |
| 多类 Brier | 0.0006916667 |
| 延迟 mean / p50 / p95 | 339.870 / 282.498 / 376.854 ms |
| Logical Call / Question / Attempt / retry | 24 / 24 / 24 / 0 |
| 已报告 input / output tokens | 6,238 / 0 |
| 已知估算费用 | $0.0006238 USD |
| partial / unreported 用量 | 0 / 0 |

混淆矩阵四个对角项 `account`、`billing`、`delivery`、`technical` 各为 6，无误判。
两种校准统计分别落在自己的 `[0.9, 1]` bin，各 24 条；其余 bin 的均值和准确率保留 null。
逐条 confidence、完整概率分布、实际答案与 gold、延迟和 Attempt 消耗在
[完整报告](live/evaluation-report.json)，没有从拒答或缺省用量生成零概率、置信或费用。
本次未观察到拒答，拒答行为由独立受控流程验证，不从本次观测推断不存在拒答。

这 24 条合成任务不能保证生产分布准确率或未覆盖置信区间的校准；没有与 Jev 比较胜负。
费用按正式目录 input-only 费率估算，不是厂商账单。无日期 `gpt-6-luna` alias 的权重
固定性、账户/区域、凭据专用性、权限范围与厂商消费限额均未核实，明确为 unknown。

## 身份、预算和证据

- 正式 route：`classifier/openai/openai-decisions/gpt-6-luna`。
- 目录版本 `2026-10-13.1`；Hash `sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。
- 执行源码 commit `d8442da`；内容 Hash `sha256:998632e85d38a77714681346205717f7760fcc9bd380a35a00158314aaaaa528`。
- 固定数据集 Hash `sha256:3380b6a4033680008446320daefd516d92e38ac9f04d618e6bdc9af25de6f0eb`。
- route/config 的独立 Hash、正式 catalog 精确匹配与 460 个源文件当前匹配见 [适用性核验](applicability.json)。
- 实际 ResponseModel：全部 `gpt-6-luna`；没有用实际 alias 覆盖授权 ModelID。
- 一个子进程仅注入 `.env:OPENAI_KEY` 对应的 `BARNESS_AI_CHINESE_EVAL_OPENAI_KEY`，
  无 Jev/其他厂商 key。账户 alias `local-openai@region-unreported` 只是本地标识。
- 串行、总 600 秒、每调用 15 秒、最多 24 Call/Question/Attempt、零重试；实际 8.157236 秒。
  信号直接送到编译后的进程，取消会保留未执行样本；外部 watchdog 不发重试。

[执行收据](live-receipt.json) 保存准确 build/执行命令、环境变量名、二进制 Hash、起止时间
和预算。[live/](live/) 包含实际完整 requests/responses、消费结果、24 个 Attempt、
96 条 Observer 记录、catalog/dataset/config/code snapshot、manifest、audit 与 files inventory。
公共 verifier 已通过无网络 transport 重放真实捕获。`execution-source.zip` 保留 code snapshot
所列的全部原始源文件，真实运行后没有修改 Go 源码来套用新 Hash。

第一方资料使用正式纳入时冻结的官方 Decisions guide/reference/model 和 Pi 身份快照；
[逐文件 SHA 核验](first-party-source-check.json) 保留原获取日期和 URL，没有宣称重新抓取价格。

## 可重复命令

从仓库根执行以下命令不会请求厂商：

```sh
cd .scratch/barness-ai-openai-decisions/chinese-live-evidence
shasum -a 256 -c SHA256SUMS
cd ../../..
go run ./ai/examples/chineseeval/cmd/chineseeval \
  -route luna -verify .scratch/barness-ai-openai-decisions/chinese-live-evidence/live
go run ./ai/examples/chineseeval/cmd/chineseeval \
  -verify .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live
go test ./ai/e2e -run '^Test(ChineseEvaluation|ClassifierHostMigration)' -count=1
```

需要重新支付调用成本时，显式执行下面命令。每次写新的 `.evidence/issue12-executions/`，
不会覆盖此交付包，不会将失败重跑成同一报告：

```sh
python3 .scratch/barness-ai-openai-decisions/chinese-live-evidence/run.py \
  --live --proxy http://127.0.0.1:7890
```

`run.py` 默认无网络。专用 key 缺失的实际公开命令另见 [NOT_RUN 收据](not-run-receipt.json)：
24 条状态完整、0 Call/Attempt、准确率 null，不称为完成，不写入支持矩阵。

## Jev 历史与后续发布

[Jev 历史核验](jev-history-check.json) 逐文件保存 before/after Hash，原包全部字节未变。
原 `2026-10-08.5` 报告仍可用自身 recorded snapshot 核验；其日期、目录、配置和源码不适用于
当前发布。工单 13 须收集 Jev 自己的新证据，不改旧报告，也不借用 Luna COMPLETE。
`applicability.json` 的判定仅对本次核验时源码有效，未来发布必须重新检查 route/config/catalog/code/date。

支持矩阵执行前后 SHA 完全一致，效果报告没有合并进协议支持矩阵。本票没有宣告十组合发布成功。

## 验证与审查

执行前重新运行工具票的 public process/E2E：105 PASS，含拒答、错 route/Hash、预算、取消、
usage 缺省、旧 Jev 和 verifier 篡改反例。网络直连探测在 10 秒超时后，按同一公开接缝
先取得代理 RED（12 个子场景及父测试均 FAIL），再补最小显式 loopback proxy 参数；GREEN
13 PASS。受控代理只返回 CONNECT 502，无转发、无真实厂商请求，24 失败仍保留和可重放。
新增代理行为不改变评估引擎。专项 race 118 PASS；冻结 Pi 开启的全量 7,991 PASS /
188 SKIP / 0 FAIL，普通和 live-tag 的全仓 vet 通过。SKIP 不视为 PASS。

Standards 初审发现 launcher 收据会保存非法代理的敏感字段，经独立公开进程三例
RED → GREEN 修复，并复审无剩余问题；Spec 独立审查 0 项。`launcher-process.py` 可重复
运行，完全不提供厂商 key。`executed-launcher.py.txt` 保存本次实际执行时的原 launcher，
后续重跑使用修复后的 `run.py`；未修改 Go 源或再次付费执行。

`preflight-artifacts.zip`、`proxy-red-artifacts.zip`、`proxy-green-artifacts.zip`、`race-artifacts.zip`
保存相应全部 evidence；每个子包原始 audit 均无发现。`test-logs.zip` 保存全部披露日志，
`full-suite-manifests.zip` 仅保存全仓的 manifest/audit，完整全仓 case 仍在本地 `.evidence`。
完整全仓测试和双轴审查结果分别记在
`validation.json`、`REVIEW.md`，失败和未运行项不替换成通过。外层 manifest/audit/SHA256SUMS
覆盖本交付目录；内层报告的 manifest/audit/files 保持原字节。
