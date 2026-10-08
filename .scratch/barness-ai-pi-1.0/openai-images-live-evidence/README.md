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

验证进展：live、目标 E2E、TypeSafe 共享格式回归、目标 race 和 go vet -tags live 已通过。
全量 race（pi 差分与压力开关启用）和双轴 review 的最终结论在完成后追加。
