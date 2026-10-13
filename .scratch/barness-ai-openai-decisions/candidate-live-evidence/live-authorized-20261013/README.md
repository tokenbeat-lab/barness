# 09：授权未知账户信息后的真实候选验收

2026-10-13 使用仓库 `.env` 的 `OPENAI_KEY`，经 `http://127.0.0.1:7890` 调用官方
HTTPS `/v1/decisions`。固定候选 9/9 场景 PASS，报告 `complete=true`、
`auditPassed=true`，审计零发现。用户明确授权权限和区域无法核实时标记 unknown，
不因此阻塞；账户身份、区域、凭据权限、是否专用及厂商消费限额均保持 unknown，
`hostPermissionVerification=null`，没有从变量名或响应推断权限。

[候选报告](candidate-decisions-report.json)、[实际执行收据](receipt.json)、
[脱敏 Go 输出](execution.jsonl) 可直接查阅；
[完整证据包](candidate-live-authorized.zip) 保存全部请求/响应、结果、Observer、
原始图片、逐场景断言与预算、来源和执行源码。[验证收据](validation.json) 区分
真实调用、受控离线检查与历史失败。[两轴复审](REVIEW.md) 各 0 项发现。

| 场景 | 结果 | Logical Call / Attempt | 问题 / 图片 |
| --- | --- | --- | --- |
| text-choice | PASS | 1 / 1 | 1 / 0 |
| text-score | PASS | 1 / 1 | 1 / 0 |
| text-bool | PASS | 1 / 1 | 1 / 0 |
| mixed-questions | PASS | 1 / 1 | 3 / 0 |
| image-png | PASS | 1 / 1 | 1 / 1 |
| image-jpeg | PASS | 1 / 1 | 1 / 1 |
| image-webp | PASS | 1 / 1 | 1 / 1 |
| cancel-in-flight | PASS | 1 / 1 | 1 / 0 |
| recovery | PASS | 1 / 1 | 3 / 0 |

实际共 9 调用、9 Attempt、13 问题、480 state bytes、3 图、1161 decoded bytes，
零重试；均有 HTTP 200 与厂商 request ID。八次正常响应实际 Model 为 gpt-6-luna，
用量完整，已报告 input=1608、output=0。取消发生在成功响应头之后、正文读取之前，
结果 aborted、答案为空，response model 未取得、usage unreported；同 Client 恢复
通过。这里没有把取消调用的未知用量计为零，也没有把目录估价当成最终账单。
正常输出共 6 choice、3 score、3 bool；概率/score 容差未改变。拒答、精度/格式
反例、厂商账单、缓存账单、区域 premium 和服务容量上限未在本次自然观察到。

候选为 `classifier/openai/openai-decisions/gpt-6-luna`，版本
`decisions-candidate-2026-10-13.1`，Hash
`sha256:b00222918033826f65c545519c4b6ad57ec1e619f427472c6c53a32ae41b1503`。
报告 schema=2、kind=candidate-only；账户归属为不声称真实身份的 opaque
`candidate-unverified-account`。正式目录与支持矩阵仍由工单 10 接线。

执行时报告的 `gitCommit=fc6937e` 是当时基线，编译含工作区改动；
[source-snapshot.json](source-snapshot.json) 固定当时每个 harness/fixture 的 SHA-256，
包内 `execution-source/` 与之逐项匹配。该线上运行时后来提交为 `abdb733`。
随后 `1c8afce` 仅调整离线完整性故障夹具，不修改候选运行时和预算；包内另存
`verified-source/`、`fixture-refinement.patch`，区分实际执行源码与最终检查源码。
没有为新的 Git 标识重跑付费调用或改写捕获。

线上运行采用 [launcher-used.py](launcher-used.py)：只提取 `.env:OPENAI_KEY`，不
source 整份文件，以环境变量注入候选自己的 key；子进程只继承基础宿主变量，不
携带其他组合 key。代理显式配置为 loopback HTTP；官方目标和 TLS 校验不变。
未知信息路径须显式 `BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1`，默认仍要求宿主
核验；提供的错误证明、跨组合 key、远端代理和 endpoint override 仍被拒绝。

离线首次 race 的 `integrity-read-failure` 在审计 16 MiB+1 合成文件时超过 20 秒
子进程监督期限：44 条目 PASS、故障子项与父项共 2 FAIL，退出 1；原始输出与
完整父进程包保存于 ZIP 的 `validation/initial-race/`，缺失的子进程 audit 没有
补造。随后用非普通文件触发同一 inventory 读取拒绝，保留 FAIL/evidence 与
complete=false 的行为；孤立 race 2 条目 PASS，完整 race 46 条目 PASS。没有放宽
生产预算或监督期限。更早的配置 FAIL 与无 key NOT_RUN 仍在上级历史包原样保留。

最终源码 `1c8afce` 的全仓检查为 7130 PASS、1118 SKIP、0 FAIL，7 个执行包 PASS；
`go vet -tags live ./...` 通过。可选 PIDiff/pressure 没有在这轮启用，其此前证据在
上级历史包，不将本轮 SKIP 冒充 PASS。

ZIP 内 `live/` 是完整真实 bundle；归档外目录的 manifest/audit/integrity 是同一原始
文件的便利副本，其路径相对于解包后的 `live/`。ZIP 根 manifest 逐项枚举实际
封装成员，包含 live 的完整 83 项 inventory、最终进程证据、原 race 失败、
日志、两轴复审与源码；全仓仅封装每包 manifest/audit，其他全仓 case artifacts
留在本地 `.evidence/decisions-candidate-authorized-full*`，不冒充包内成员。
两个既有图片 replay 选择测试另生成仅一项的 manifest，没有 audit；保留其缺失
状态，并在 validation 中列明。它们用于核验重放命令，未执行线上图片调用。

校验与离线重放：

```sh
(cd .scratch/barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013 && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013/candidate-live-authorized.zip
candidate_evidence_dir="$(mktemp -d)"
unzip -q .scratch/barness-ai-openai-decisions/candidate-live-evidence/live-authorized-20261013/candidate-live-authorized.zip -d "$candidate_evidence_dir"
go run ./ai/internal/testkit/audit/cmd/auditbundle "$candidate_evidence_dir"
go test -tags live ./ai/live -run '^TestDecisionsCandidateProcess$' -count=1
go test -race -tags live ./ai/live -run '^TestDecisionsCandidateProcess$' -count=1
go vet -tags live ./...
```

如需再次执行真实 API，先从仓库根确认当前凭据与代理，再显式运行上述 launcher；
它会再次付费调用并写入新的 `.evidence` run directory。原运行日志和收据路径是
脚本使用时的固定 `/tmp` 路径，重跑需保留本包中的既有文件。
