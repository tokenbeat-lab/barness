# 固定图片任务的独立 Host 效果评估

从仓库根目录运行；沿用 `classifierimages` 的公开图片输入与结果处理流程。
本示例属于 Host，不向 barness-ai 添加业务阈值，也不更新正式目录、支持矩阵或
冻结 Pi。工单 11 的候选协议证据与本评估的效果结论分别保留。

```sh
go run ./ai/examples/visualeval -route jev -entry plain -out .evidence/my-jev-visual
go run ./ai/examples/visualeval -route decisions -entry hooked -out .evidence/my-decisions-visual
go run ./ai/examples/visualeval -verify .evidence/my-jev-visual
```

默认是 **controlled**：真实 Client/HookedClient、协议转换、HTTP 和 Observer
连接受控 Provider，回放 `testdata/*-responses.json`。这些脚本故意含误判且不做视觉识别，
COMPLETE 只表示固定流程完成；`realQuality` 始终为 NOT_RUN。
`-verify ... -require-live` 会拒绝 controlled 和 NOT_RUN。

`testdata/dataset.json` 是独立固定真值，包含中文标注依据、类别、存在性和人工框；
`scenes.json` 只包含绘制说明。`generate.go` 不读标注，运行时请求和响应脚本也不读真值。
`experiment.json` 固定问题、Host 消费策略和上限；`pins.json` 固定所有这些字节及各张
原始 PNG、处理 PNG/JPEG、生成器、响应脚本和候选目录的 SHA-256。

共 12 张 320×320 合成图片：白底红矩形、蓝圆或空白，原点左上，无 EXIF。
预处理为方向不变、全图裁剪；偶数样本额外编码成 quality=90 JPEG，其余为 PNG。
源图和处理图均随包保留。本任务仅描述简单合成几何图形，不代表生产图片分布。
人工框使用半开像素边界除以宽高；JPEG 边缘压缩不改变标注框。

| 样本 | 两条路线的任务 | Jev 附加任务 |
| --- | --- | --- |
| v01–v08 | 完全相同的 context 公共输入，选择类别与独立 Bool 存在性 | 无 |
| v09–v10 | Decisions 为补充 context 分类，不混入可比较子集 | State 局部图片，另问 Box |
| v11–v12 | Decisions 为补充 context 分类，不混入可比较子集 | 各题 Instructions 局部图片，另问 Box |

切换前 8 个任务的路线只改变 Target 和对应 Options。每个样本只有一个搜索图片，
所以 Box 的图像归属明确；Host 根据实际发送的处理图尺寸解释坐标。Jev 保留 nullable
Box confidence 的原始结果；Decisions 的 Box 指标为 null，不虚构定位能力。

评分分组为 common、local（Jev）、supplementalContext（Decisions）。类别与存在性分别
报告全部样本分母、可评分数量、正确数、混淆、FP/FN、错误样本 ID 与逐样本终态。
可用 Choice confidence 按十个等宽区间报告 count/correct/sum；存在性另给条件 Brier，
统计数量明确，不把厂商 confidence 解释为已校准的概率。失败/空答案/拒答保留在原位置和
混淆/状态计数中；未运行分数为 null，已经运行的完整分母包含失败和未发送样本。

Box 对所有真实存在目标的合法预测计算原始 IoU，即使 Bool 低于阈值；另外报告可评分
均值和以全部存在目标为分母的均值、缺框数。Host 仅当独立 Bool probability ≥ 0.5 才消费
框；该示例策略写在配置中，不是库默认。缺目标误报是“缺目标但 Host 消费了框”，另列
所有缺目标原始框及低/缺存在性框。合法四边和 Box confidence 都不证明目标存在。
不设通用准确率门槛，不因低分更改标注、挑选样本、重复请求或补跑。

每路线每轮固定 12 个样本；最多 12 Logical Calls / 12 Attempts，零重试、48 题、
24 图片出现，累计 decoded source ≤8 MiB、展开 ≤16 MiB。实际 Jev 为 28 题/16 次图片
出现，Decisions 为 24 题/12 次。每调用 ≤30 秒，工作截止 225 秒，进程总限 240 秒，
留 15 秒写终态与审计；取消后停止新调用，所有未发送样本仍保留。
`-calls/-call-ms/-work-ms` 只能在 controlled 模式缩小上限，用于失败流程 E2E。
消耗来自实际 Attempt：缺失、部分上报分别计数，按所选目录价格快照估算，
不视为免费或厂商账单。用量不完整阻止 COMPLETE，但合法答案仍可评分。

`-formal` 将所选 Jev 或 Decisions 路线固定到 `ai.BuiltinCatalog()` 的当前正式目录；
默认仍为候选目录。正式发布分别使用两路线各自的 `-formal` 包，不能互换候选、型号
或旧目录 Hash。每个进程仍只运行一条路线，数据、评分和有限预算不变。

```sh
go run ./ai/examples/visualeval -formal -mode live -route decisions -out .evidence/my-formal-decisions-not-run
```

## 真实运行

当前两路线真实图片效果均为 **NOT_RUN**，没有使用其他任务的 key。
只有持有本票专用授权与凭据的 Host 才可执行真实评估。每进程仅一个路线，
唯一读取的凭据变量是 `BARNESS_VISUAL_KEY`，拒绝其他 credential/token/key/auth
环境变量及 endpoint override。HTTP transport 不使用环境代理、不跟随重定向，
仅使用路线固定的官方服务。

```sh
# 不提供授权/凭据时生成 12 样本、零调用的 NOT_RUN 包。
go run ./ai/examples/visualeval -mode live -route jev -out .evidence/my-jev-not-run

# 获得独立授权后，在只包含所选路线专用 key 的隔离进程中运行一次。
go run ./ai/examples/visualeval -mode live -route jev \
  -authorization /trusted/visual-authorization.json -out .evidence/my-jev-live
```

授权 JSON 必须显式匹配 route、model、完整 service、catalogHash、accountAlias、
regionAlias、reference、dedicatedCredential=true、maxCalls=12、maxAttempts=12。
Jev 为 `jev-1.14.0` / `https://api.typesafe.ai/v1/systemone`，
Decisions 为 `gpt-6-luna` / `https://api.openai.com/v1/decisions`；candidate 精确复用
`ai/live/testdata/multimodal/` 的工单 11 目录。目录内容 hash 见报告与其独立协议回执。
账户/区域/最小权限是 Host 明确担保，报告不是权限授予。

每个包保留独立时间、模型、目录、输入与来源 hash、运行配置、源码快照与 hash、
工单 11 协议回执关联、原始/处理图片、预测/标签/指标、最终 wire/请求/响应、
结果/类型化错误、Attempt/Observer、资源归零断言、完整性清单及脱敏审计。
输出目录必须不存在；失败包不可覆盖。审计失败保留材料并以非零退出，绝不变为 PASS。

独立核验会重新检查文件集合/hash/审计、固定输入/源码、候选关联、预算、全部终态、
Observer 和资源，并将捕获的响应经无网络的公开 Classify 重放，核对最终请求、图片
身份、答案、usage/估价，再重算指标。超限捕获保留 64 KiB 限额后的一个判定字节，以便重放同一个 resource_limit。取消若发生在解码前，
usage 保持零值且 unreported；若已解码，保留的用量和型号必须由捕获响应与价格快照支持。
Observer 的终态错误、用量及 Attempt 与公共结果逐项对账。Jev 仅归一化随机私有素材 key；位置、图片和问题
仍精确匹配。核验旧包应 checkout 该包自己的源码版本；新目录 Hash 不继承旧效果。
清单是完整性回执，不是来源数字签名；真实来源依赖 Host 保存原始可信包。

复现 fixture：`go run ai/examples/visualeval/testdata/generate.go` 后逐项核对 pins，
不要自动重写标签或 pins 来消除失败。公开进程 E2E：

```sh
go test ./ai/e2e -run '^TestVisualEvaluation' -count=1
BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run '^TestVisualEvaluation(FixedWorkflow|FailureRetention|CancelledAccounting|SourceSet|Authorization|NotRun)$' -count=1
```

回滚只需删除此 opt-in Host 示例及其测试；没有迁移、内部兼容层或运行时行为变更。
