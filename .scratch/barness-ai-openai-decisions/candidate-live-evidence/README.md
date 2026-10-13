# 工单 09：候选 Decisions 独立证据

2026-10-13 用户明确授权使用 `.env:OPENAI_KEY`，经 `127.0.0.1:7890` 执行，
权限和区域未知如实记录，不再阻塞。固定候选 9/9 真实场景 PASS，9 次调用/9 次
Attempt、13 问题、3 张图片、零重试，complete=true，审计零发现。
[真实报告、收据及完整捕获](live-authorized-20261013/README.md) 已保存；账户/区域/
权限/是否专用/厂商消费限额为 unknown，未伪造权限证明。正式目录和矩阵仍未纳入。

下述 `candidate-live.zip` 与 validation.json 是此前离线实现和配置拒绝的历史证据，
保留原 NOT_RUN/FAIL，不以新 PASS 改写历史。新的 candidate report schema=2，以
nullable 权限证明和显式 unknown 状态区分实际执行与账户核验。

公共接缝为 `Client.Classify` / `HookedClient.Classify` 与独立测试进程。
[失败方式](FAILURES.md) 先于实现；进程反例覆盖来源/候选篡改、跨组合凭据、代理/
权限配置、重复字段/尾随 JSON、缺失场景、429/504、损坏/精度 2xx、逐题及全拒答、
缺失/部分/明确零用量、实际 response alias、预算耗尽。受控 Provider 是 TLS 边界替身，
报告 mode=controlled；生产入口没有 endpoint override。

完整管线为文本 choice/score/bool、混合题、PNG/JPEG/WebP、成功响应头到达后取消、
同 Client 恢复，共 9 Logical Call、9 Attempt、13 个问题、3 张图片。取消时正文未读，
usage 维持 unreported；不捏造厂商耗用。图片为真栅格，完整解码并保存字节、尺寸、
MIME、SHA-256、base64/data URL 长度。正常结果核验自己的问题键/候选/有序 rubric、
分布和 Decisions 精度、实际型号、调用归属、Attempt/厂商 ID、Observer 四记录。
合法 RefusalAnswer 不导致重复请求；拒答/429/504 不主动在线制造。

每进程上限 12 调用/12 Attempt 预留/24 问题/4 MiB 累计 state/8 张累计图片/
4 MiB 累计 decoded 图片，180 秒总与 30 秒单调用，Binding 零重试。其中预留 10 秒用于终态摘要、Observer、
manifest/audit/integrity 落盘，执行使用进程期限及 Go 测试外层期限中较早者；不等待
硬超时 panic 才退出。每次调用之前
原子预留完整输入预算；故障和取消不退回预留。state/图片字节反例还验证累计耗尽，
不是仅检查单次大小。请求捕获最多 8 MiB，响应捕获沿用 512 KiB；截断不能 PASS。
协议 pipeline complete 与 UsageReporting 分开，usage 缺失也不编造 token 计数。

固定来源及候选在 [source README](../../../ai/live/testdata/decisions/README.md)。
独立报告 `candidate-decisions-report.json` 使用 candidate-only 身份及版本/Hash，
包含来源日期/Hash、权限核验记录/Hash、预算、完整必交集合和未观察项目。
`candidate-integrity.json` 校验包内报告、manifest、audit 和逐场景完整文件。
正式 `live-report.json` 不会产生；正式加载器拒绝候选 schema。权限记录是宿主
核验说明，不冒称来自 OpenAI 权限自省 API。

重放离线进程与现有协议/差分：

```sh
go test -tags live ./ai/live -run '^TestDecisionsCandidateProcess$' -count=1
go test -race -tags live ./ai/live -run '^TestDecisionsCandidateProcess$' -count=1
BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run 'Decisions|JevDifferentialMetadata' -count=1
go vet -tags live ./...
```

真实运行由宿主从 `.env` **只提取 OPENAI_KEY**，通过进程环境注入
`BARNESS_AI_LIVE_KEY_OPENAI_DECISIONS_CANDIDATE`；不 source 整份文件，不注入其他组合
或厂商 key，不在命令行/日志写 key。默认要求宿主核验；本次用户授权的未知信息路径使用以下配置：

```sh
BARNESS_AI_LIVE=1 \
BARNESS_AI_LIVE_COMBO=openai-decisions-candidate \
BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1 \
BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890 \
go test -tags live -timeout 180s ./ai/live -run '^TestCandidateDecisions$' -count=1
```

提供核验材料时的权限记录格式（示例值不能当成实际权限证明）：

```json
{
  "accountAlias": "ci-decisions-lowpriv",
  "regionAlias": "us",
  "dedicatedTestKey": true,
  "model": "gpt-6-luna",
  "permissions": ["POST /v1/decisions"],
  "spendLimitUSD": 1,
  "verifiedAt": "2026-10-12T00:00:00Z",
  "verificationReference": "维护者核验的脱敏配置记录编号"
}
```

宿主记录须与别名一致，权限只含本路线，并说明真实模型访问、专用 key 与有限消费
配置的核验来源。文件以有界单一 JSON 读取，拒绝未知/非规范大小写/重复字段、尾随内容和未来核验
时间。未提供材料且显式 ALLOW_UNKNOWN_ACCOUNT=1 时，用 unknown 记录，不填写虚构材料；
提供的错误材料仍拒绝，跨组合 key、非本地代理和 endpoint override 均不获放行。
代理必须是显式 loopback HTTP URL；真实目标和普通 TLS 校验保持官方 endpoint。
只有 candidate 自己全部 PASS 且审计成功，才可把本票转为 resolved 并进入工单 10。

复审另以真实 Go 测试期限和文件系统边界验证：进程期限到达后仍保存完整失败账本；
integrity 读取/写入失败会撤销 complete 并以 evidence 失败退出；权限字段大小写
别名拒绝于网络前。原复审 RED 与修复 GREEN 原样保留。

最终源码 `a7c15ec`：候选进程普通/race 各 41 个测试条目 PASS；全仓最终
7125 PASS、1118 SKIP、零 FAIL。未开启 PIDiff/pressure 的可选项没有被冒充通过。
更早 `85ea917` 的完整 pressure+PIDiff 全仓为 8172 PASS、67 SKIP、零 FAIL，
自己的 Decisions 公共 E2E/差分 1092 条目 PASS。`go vet -tags live ./...` PASS，
所有已执行包审计无发现；两轴剩余发现各 0。准确命令、源码归属、退出状态和日志
SHA-256 见 [validation](validation.json)，初始发现及复核见 [REVIEW](REVIEW.md)。

固定 `candidate-live.zip` 保留最终普通/race 进程完整证据、1092 项协议/差分、
实际配置 FAIL/config、无 key NOT_RUN、复审 RED/GREEN、两轮全仓 manifest/audit
与混合容量适用 artifacts，以及起点变更的修复前/后源码。其他全仓 case 原始
artifacts 仍保留在本地 `.evidence`；包内对应完整 manifest 如实列出其状态。
包内 root manifest 对实际封装成员逐项 SHA-256，不把省略的全仓 case 当成封装成员。

固定包核验：

```sh
(cd .scratch/barness-ai-openai-decisions/candidate-live-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/candidate-live-evidence/candidate-live.zip
candidate_evidence_dir="$(mktemp -d)"
unzip -q .scratch/barness-ai-openai-decisions/candidate-live-evidence/candidate-live.zip -d "$candidate_evidence_dir"
go run ./ai/internal/testkit/audit/cmd/auditbundle "$candidate_evidence_dir"
```

包内 manifest 与报告、图片、逐场景 assertions/replay、每个候选进程独立的
candidate-integrity.json 均可重复核对。原始开发失败和复审失败保留，没有以后续
PASS 改写历史。未观察的服务精度/格式反例仍未知。若将来遇到真实反例，原样保存 capture、usage 和 protocol FAIL，归因到
Decisions，涉及 ADR 的取舍需显式修订；本次没有改变概率/评分容差。
