# 工单 14：Google 图像真实接受与目录证据

固定 image / Google / google-interactions，裸型号 gemini-nano-banana-2.1，
官方 v1beta/interactions；公共 Client.GenerateImages 的独立真实进程已完成
生成和一参考图编辑 PASS。2 调用、2 张 JPEG 1024×1024、0 重试，预算为每进程
最多 4 调用/4 张/1K。每次请求认证只在 x-goog-api-key；store=false，单 image
response_format，1K/1:1，无背景/流式/工具/续接，也无版本、型号或协议回退。

初次生成和编辑均以 400 拒绝显式 delivery=inline；官方虽列该可选枚举，真实
组合不接受。维护者明确批准省略 delivery 并修订规格（ADR-0023）。请求始终
省略该字段，回调添加 inline/uri/null 均拒绝；响应仍只接受 completed、内联
可读图片，无 URI/续读。随后简单圆形提示被版权/复述过滤拒绝，改用原创静物
提示的同组合完整重跑通过。两份 FAIL 原始报告保留，不能解释成 PASS。

真实响应省略 id/请求 ID 响应头，报告 model；诊断 ID 保持 unset，不用授权
型号或本地标识伪造。每个 model_output 都遍历；完整有界捕获（16 MiB）保留
步骤、终态和尾部 usage，图片与签名替换为长度/哈希标记；图片另存且完整解码。
捕获丢失/截断不能证明成功。结果、请求、Observer 与 audit 均在各原始目录中。
审查发现的失败正文脱敏退回原文已修复：请求截断不会跳过响应脱敏；无法解析、
超限或读取中断的正文只保留长度/哈希，短签名、续读令牌和 URI/URL 也清除。
六个公共 Client 失败路径先 RED 后 GREEN，日志保存为 review-red.log / review-green.log。
交付日志仅将本机绝对仓库路径替换为 [REPO]，保留原始失败断言。

两份 PASS 的 output_tokens_by_modality 只有 image=1120。文本思考另报，不在
图片明细内；计入 Output，同时作为 Reasoning 内含分项，TotalTokens 保留厂商值。
原始模态字段不补零、不累加内部模型调用明细；缺文本输出、生成的缺图片输入
明细为 partial。已知成本包含一次思考费，不代表完整账单：生成 0.03981150 USD，
编辑 0.03554250 USD，合计 0.07535400 USD；未报告的非思考文本分项没有猜测计价。
实际形状的离线回放覆盖 captured partial、缺总量 partial、无 usage unreported；
原有合成 fixture 继续覆盖完整明细下思考已含/未含的两种形状。

官方型号、参考图能力、宽高比/尺寸、输入 MIME 和标准费率重新核对于 2026-10-08，
见 ai/e2e/testdata/google-images/official.json 的官方链接。官方价格页与图像指南
对 4K token 数冲突，保留冲突并仅验证 1K。费率按美元/百万 token：text/image
输入 1.50，text/thought 输出 7.50，image 输出 30；该路线未文档化适用缓存价，
不推断、不叠加按张价。内置能力只声明一参考/一输出、1K、1:1，目录 2026-10-08.5。
弃用/关闭型号未纳入；官方更大能力不是本次支持声明。

原始来源：

- delivery-refused：20261008T090854.560734000Z（FAIL，2 调用，0 用量报告）。
- prompt-refused：20261008T091316.585690000Z（FAIL，生成被拒；编辑图片已返回，早期形状断言未通过）。
- live-pass：20261008T091619.605370000Z（两项 PASS）。
- not-run：20261008T091928.075162000Z（无凭据，两项 NOT_RUN）。

凭据由显式宿主 launcher 从 .env 只读取 GEMINI_KEY，再只注入本组合变量。
库不读取 .env、不读取厂商环境变量。账户别名 dotenv-gemini@region-unconfirmed
说明本地来源；API 响应不能证明 key 权限等级或实际区域。CI 仍由宿主提供
本组合低权限账户及准确区域别名。

复跑真实接受（有费用，每进程仍遵守四张上限）：

```sh
python3 .scratch/barness-ai-pi-1.0/google-images-live-evidence/run-live.py \
  --dotenv .env --evidence-dir .evidence/issue14-live \
  --alias dotenv-gemini@region-unconfirmed
```

正式 CI 以自身 secret store 注入 BARNESS_AI_LIVE_KEY_GOOGLE_INTERACTIONS_IMAGE，
同时设置 BARNESS_AI_LIVE=1、BARNESS_AI_LIVE_COMBO=google-interactions-image、
BARNESS_AI_LIVE_ACCOUNT_ALIAS，并运行 go test -tags live ./ai/live -count=1。
无 key 的同一入口为 NOT_RUN，真实拒绝为 FAIL，生成/编辑不能 UNSUPPORTED。

无需厂商凭据的验证：

```sh
go vet ./...
go vet -tags live ./...
go test ./ai/e2e -run '^TestGoogleImages' -count=1
go test -race -tags live ./ai/live -run '^Test(GoogleImagesHarness|ImagesHarness)' -count=1
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
python3 .scratch/barness-ai-pi-1.0/google-images-live-evidence/verify.py
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/google-images-live-evidence
```

每次 E2E 写独立 evidence bundle，manifest 带回放命令，结束自动脱敏审计。
Google 作为原生扩展登记，不计为 pi 差分通过。支持矩阵只由本组合报告合并；
公开 CLI 验证错误路由、缺能力、旧/配置失败、超预算、缺形状、URI/续读、错误
思考关系等批次不能部分写。发布门禁新增本组合及独立 audited live bundle 要求，
本工单不宣称九组合全部发布门禁通过（后续工单 17）。

verification.json、offline-manifest.json、google-cases.json、live-harness-manifest.json、
live-harness-cases.json、trace.json 与 review.md
记录本次回归和双轴审查；source-hashes.json / sha256.json 固定源文件和交付内容。
原始 live 报告的 gitCommit 为启动时 HEAD（实现仍在工作区），不能当作独立已提交
构建的证明；完整离线/race 回归及 source-hashes 补齐可重复的实现关联。
