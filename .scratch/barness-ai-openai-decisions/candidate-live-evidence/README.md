# 工单 09：候选 Decisions 独立证据

候选 harness 已实现，正式目录和支持矩阵未纳入 Luna。真实厂商调用尚未执行，
本票未完成：用户指定 `.env:OPENAI_KEY` 与本地代理 `127.0.0.1:7890`，但没有提供
测试账户/区域别名、专用低权限 key 和权限核验记录。实际凭据配置检查在网络前
返回 FAIL/config；无凭据路径独立记录 NOT_RUN。不能把受控 fixture PASS 写成真实支持。

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
4 MiB 累计 decoded 图片，180 秒总与 30 秒单调用，Binding 零重试。每次调用之前
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
或厂商 key，不在命令行/日志写 key。须提供以下宿主配置：

```sh
BARNESS_AI_LIVE=1 \
BARNESS_AI_LIVE_COMBO=openai-decisions-candidate \
BARNESS_AI_LIVE_ACCOUNT_ALIAS=<已核验测试账户别名> \
BARNESS_AI_LIVE_REGION_ALIAS=<已核验区域别名> \
BARNESS_AI_LIVE_PERMISSIONS_FILE=<脱敏宿主核验记录路径> \
BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890 \
go test -tags live -timeout 180s ./ai/live -run '^TestCandidateDecisions$' -count=1
```

权限记录格式（示例值不能当成实际权限证明）：

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
配置的核验来源。文件以有界单一 JSON 读取，拒绝未知/重复字段、尾随内容和未来核验
时间。代理必须是显式 loopback HTTP URL；真实目标和普通 TLS 校验保持官方 endpoint。
只有 candidate 自己全部 PASS 且审计成功，才可把本票转为 resolved 并进入工单 10。

最终验证收据、两轴复审和固定包校验将在本目录保存；未观察的服务精度/格式反例
仍未知。若将来遇到真实反例，原样保存 capture、usage 和 protocol FAIL，归因到
Decisions，涉及 ADR 的取舍需显式修订；本次没有改变概率/评分容差。
