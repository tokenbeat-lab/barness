# 10：正式 Catalog 与第十组合独立支持

候选 09 自己真实 9/9 PASS、完整审计后，正式纳入
`classifier/openai/openai-decisions/gpt-6-luna`。BuiltinCatalog 版本
`2026-10-13.1`，Hash
`sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。
正式报告以源码 `f9bd770` 和此 Hash 真实重跑 9/9 PASS，经公共 merge 命令更新
第十行；没有宣布全十组合可发布。

[正式报告](live/live-report.json)、[manifest](live/manifest.json)、
[audit](live/audit.json)、[完整性清单](live/decisions-integrity.json) 及完整 wire
捕获、结果、Observer、真实图片都在 `live/` 和可独立解包的证据 ZIP。
[执行收据](live-receipt.json)、[两轴复审](REVIEW.md)、[最终验证](validation.json)
明确区分真实 API、受控反例、历史失败及全仓检查。

正式共 9 Logical Call / 9 Attempt / 13 题 / 480 state bytes / 3 图 /
1161 decoded bytes，零重试。八次正常响应实际 Model=gpt-6-luna，usage complete，
已报告 input=1608/output=0；取消为成功响应头之后 aborted、usage unreported。
已报告估价合计 $0.0001608；被取消调用的用量及厂商最终账单未知，不把缺省零值
当作免费消耗。区域 premium、缓存账单、alias 权重固定性和服务容量上限均未测。
账户、区域、权限、专用凭据与厂商消费限额按 09 用户授权如实 unknown。

| 正式必交能力 | 结果 | Calls / Attempts | 问题 / 图片 |
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

`catalog-before.json` 与 `catalog-formal.json` 对照后，移除新增 Luna 并恢复旧目录
版本可精确取得旧 Hash；Jev 省略态、同 ID 聊天内容和其他所有旧条目保持原样。
`matrix-schema2-before.json` → `matrix-schema3-migrated.json` →
`matrix-schema3-after-live.json` 展示一次性迁移和自己的报告合并。迁移九行逐值
一致，第十行从九能力 NOT_RUN 开始；无运行时迁移或双写。

[nine-route-applicability.json](nine-route-applicability.json) 逐行标明旧九组合历史
schema 2 不能替代新 Hash 的当前证据。其状态、能力、时间、SDK 与 notes 中历史
Hash 未改写，工单 13 收集正式新 Hash 所需重跑。Jev 默认评估配置只更新目录 pin
和配置版本到 zh-eval-jev-v3，旧原件保留为 jev-evaluation-config-before.json；
没有重新取得或改写 Jev 真实效果，没有执行 Luna 中文业务评估。

首次真实执行（dbb4660 工作区源码）在 ZIP 的 initial-live/，仅报告 token 聚合
缺省为零；实际捕获有已报告 input=1608，原报告未改写。第二轮 faed800 报告聚合
已正确，在 publication-live/；正常运行实际完整通过，但复审发现其暂存文件仍能
单独被 Report decoder 接受。第三轮最终 f9bd770 在 live/，覆盖修复后的独立
暂存封套和原子发布。三轮分别 9 calls/9 attempts，均在自己的进程预算内；
没有重试厂商 2xx 或为了制造拒答/429/504反复请求。各轮捕获和源码独立保留，
不换写 Git 身份、Hash 或历史时间。

公共 TDD 接缝是 Catalog、Merge/CLI 和独立进程。反例含错身份（矩阵与报告同时错）、
候选/旧 Hash、缩集/重复/缺项、UNSUPPORTED 图片、所有累计预算及负 token、
跨行 batch 拒绝、NOT_RUN/FAIL 历史、source 篡改、外来 key、endpoint、厂商故障、
审计失败、两处 abrupt exit。原子发布在完整 audit/inventory 后执行；审计前的
pending 是独立封套，公共严格解码器拒绝。最终 no-key 9 NOT_RUN、零消费，公共
merge 后正式 PASS 文件字节不变，见 not-run-final-receipt.json。

完整 ZIP 同时保存 scoped merge/process/race 的全部 artifacts 和原 RED 反例。
全仓检查包只保存各执行包的 manifest/audit 及完整 Go JSON 日志，其他大型全仓
case artifacts 在本地 `.evidence/formal-final-verified-full`；不声称全部在 ZIP 内。
首次 full 因旧 Jev config pin 失败的原始输出与清单也保留，没有只保留重跑成功。

离线重放（不会注入厂商 key）：

```sh
go test ./ai/e2e -run '^TestDecisionsFormal' -count=1
go test -tags live ./ai/live -run '^TestDecisionsFormalProcess$' -count=1
go test -race -tags live ./ai/e2e ./ai/live -run '^TestDecisionsFormal' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -tags live ./... -count=1
go vet -tags live ./...
```

ZIP 校验与包审计：

```sh
(cd .scratch/barness-ai-openai-decisions/formal-support-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/formal-support-evidence/formal-support.zip
formal_unpack_dir="$(mktemp -d)"
unzip -q .scratch/barness-ai-openai-decisions/formal-support-evidence/formal-support.zip -d "$formal_unpack_dir"
go run ./ai/internal/testkit/audit/cmd/auditbundle "$formal_unpack_dir"
```

再次实际 API 测试必须显式执行 launcher.py，会重新花费；只提取 `.env:OPENAI_KEY`，
仅注入本组合 key，显式 loopback HTTP 代理，固定官方 HTTPS endpoint 与 TLS 校验，
沿用 12 calls/12 attempts/24 题/4 MiB state/8 图/4 MiB decoded 图片、
180 秒进程（留 10 秒收尾）/30 秒调用/零 Binding 重试。重跑应写新 artifact 目录，
不得覆盖本包记录。候选原始 report 和来源 pins 仍在工单 09 的独立包。
