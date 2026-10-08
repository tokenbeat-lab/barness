# Issue 11 — OpenAI Images live and catalog evidence

2026-10-08：固定 `gpt-image-2.5-sunburst-2026-09-08`，由公共 Client 顺序执行
内联 JSON 编辑、生成、mask 编辑，三项 PASS。实际 3 次 HTTP 请求、3 张图、0 次
重试；输出 PNG/JPEG/WebP 均完整可读、1024×1024，usage complete。上报 token 的
估算总价 0.029431 USD，区别于账单；失败请求与重试也会预占图片/调用预算。

- [真实报告](live/20261008T071013.136836000Z/live-report.json)、
  [原始包 manifest](live/20261008T071013.136836000Z/manifest.json)、
  [原始包零发现审计](live/20261008T071013.136836000Z/audit.json)。
- 每个 LIVE-P08 子目录保存有界凭据脱敏请求/响应、完整公共结果、Observer、图片
  文件、SHA-256 和解码校验，不以像素相等判定成功。PNG 原始响应超过 512 KiB
  捕获限制，`truncated: true` 保留；公共结果和图片校验未截断。
- [官方能力/价格事实](../../../ai/e2e/testdata/openai-images/official.json) 保存官方
  URL 与 2026-10-08 抓取日期。费率：输入文本 5、输入图片 8、输出图片 30 USD/百万
  token，输出文本不计费；无直接 Images 缓存费率或按张派生价。
- [真实回放 fixture](../../../ai/e2e/testdata/openai-images/live-contract.json) 明确
  PNG 响应由完整 Client 输出和 reported usage 归一化恢复；另两项是原始 JSON。
  归一化不是第二份真实接受性，不能替代原始报告。
- [离线独立核对](verification.json) 检查真实请求/结果、图片哈希、官方费率、目录
  快照与实际 CLI 原子合并；失败批次不得改变原文件，其他矩阵行保持相同。
- FAILURES.md 在实现前列出误判、预算、用量、矩阵和泄漏失败方式；red/green 日志
  记录注册、报告拒绝、公共 HTTPS Client 与目录纳入的验证。

目录 2026-10-08.4 仅纳入固定快照。支持一张参考图/输出、1024×1024、low/medium、
mask 与透明背景；PNG/JPEG/WebP 分别实测。其他官方质量、数量与尺寸未扩大声明。
Sunburst 的 input_fidelity 值未明确、未实测，所以要求省略，不误报 UNSUPPORTED。
矩阵 image/openai/openai-images 行仅由这份报告合并，编辑与生成为必交能力；仅
明确厂商 mask support 拒绝能将可选 mask 标为 UNSUPPORTED。报告新增可选图片字段
沿用 schema 2，TypeSafe/聊天旧语义与自身报告继续独立。

账户别名 `dotenv-openai@region-unconfirmed` 只说明宿主 `.env` 来源；API 响应没有
证明低权限配置/实际区域。CI 须由宿主提供组合专用低权限凭据及真实区域别名。
本地 launcher 仅注入 OPENAI Images key 到独立进程，其他 dotenv key 不进入子进程。
没有 endpoint/model 覆盖、multipart 或其他协议回退。

从仓库根目录复跑离线核对（不需要 key，不会调用厂商）：

    python3 .scratch/barness-ai-pi-1.0/openai-images-live-evidence/verify.py
    go test ./ai/e2e -run '^TestOpenAIImages(BuiltinCatalog|LiveWireReplay)$' -count=1
    go test -tags live -race ./ai/live ./ai/internal/testkit/supportmatrix -run 'TestImages|TestClassifier' -count=1

明确运行付费真实请求（双开关由 launcher 显式设置，每进程最多 4 请求/4 张 1K 图）：

    python3 .scratch/barness-ai-pi-1.0/openai-images-live-evidence/run_live.py

正式 CI 按 ai/live/doc.go 直接注入 BARNESS_AI_LIVE_KEY_OPENAI_IMAGES，不从文件加载。
新的真实报告须经审计后再合并，verify.py 固定核对本次原始证据。

最终验证：全量 `BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1`
退出 0，4007 个 E2E 全 PASS（含 623 个 pi 差分、11 个压力场景和 267 个 P08 场景），
无数据竞争；普通及 live 标签 go vet、目标 live 标签 race 均通过。见
[全量摘要和四个目录/回放断言](full-evidence.json)、[全量 manifest](full-bundle-manifest.json)、
[全量审计零发现](full-bundle-audit.json)、full-race.log 和 vet-results.json。
完整离线捕获留在本地 `full/20261008T073009.099075000Z`（体积大，git 忽略）；
提交的断言、fixture/artifact 哈希和 replay 命令可在新 checkout 重复验证。

首次全量仅旧 responses/usage.json 的目录哈希 pin 失败，保留 full-pin-failure.*
及零发现审计；更新 pin 后完整重跑全量，不以单测通过替代这次全量验收。
双轴复审：Standards 0、Spec 0 未解决发现。

[发布追溯核对](release-evaluation/release-report.json) 以现有完整包评估；
P08 live 行、V6/H2 图片追溯、全部离线 P0、pi 差分、快照和审计均 PASS。
该 `-bundle` 模式不运行 command gates，且只提供本工单 live 包，其余七条路由
未在这次评估传入包，因此全局 release-report 整体为 FAIL、命令预期退出 1。
这份范围核对不宣称 issue 17 的九路发布门禁已经完成。

manifest.json 保存本证据目录已提交产物的字节数和 SHA-256（排除自身、audit.json
和派生 verification.json）；audit.json 对可提交产物再审计，含 `.env` 所有凭据的
已知值检查，零发现。大型 full/ 的每次运行已单独完成审计。

代码审查修正：保存的逐场景 replay 原先只选 generation/mask，无法满足同进程编辑
前置条件。现使用完整 OpenAI Images 组合；此次只更正 manifest/assertions 的 replay
元数据，没有改变真实结果或调用消耗。新增 emitted-manifest 回归。透明输出完整解码后
还须有透明像素；不透明受控响应不能确认透明能力。mask 格式/尺寸错误不能伪装为厂商
不支持 mask。详见 [双轴审查](review.md)、transparency-red.log 和 mask-refusal-red.log。
