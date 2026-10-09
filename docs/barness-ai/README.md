# barness-ai 发布说明

barness-ai 是 barness 连接多种 LLM Provider 的协议中间件（Go 包 `github.com/tokenbeat-lab/barness/ai`）。本目录汇总发布交付物（spec I10、Testing Decisions §6，工单 24）。

使用方式、三种模型操作、绑定与资源策略见包级 README：[English](../../ai/README.md) / [简体中文](../../ai/README.zh-CN.md)。

| 交付物 | 位置 |
| --- | --- |
| 公共契约、错误分类与装配说明 | [contract.md](contract.md) |
| 差异登记（含 D1/D2 与 DeepSeek Responses 扩展） | [differences.md](differences.md)；逐条账本 [`ai/e2e/testdata/pidiff/ledger.json`](../../ai/e2e/testdata/pidiff/ledger.json) |
| 支持矩阵 | [`ai/live/support-matrix.json`](../../ai/live/support-matrix.json)（只经 `go run ./ai/live/cmd/supportmatrix` 合并冒烟报告更新） |
| 模型/价格快照 | [`ai/release/catalog-snapshot.json`](../../ai/release/catalog-snapshot.json)（内置目录及其哈希；门禁核对它与代码一致） |
| Fixture | [`ai/e2e/testdata/`](../../ai/e2e/testdata/)；每次门禁在发布证据包中写出 `fixtures.json`（路径、大小、SHA-256） |
| 资源策略示例 | 聊天：`localassembly.LocalPolicy`、`hostintegration.CloudInteractivePolicy`、`CloudBatchPolicy`；混合操作：`localassembly.MixedPolicy`、`hostintegration.CloudMixedPolicy` |
| 最小示例 | `ai/examples/localassembly`、`hostintegration`（聊天与混合操作）、`toolloop`、`requestid`；独立中文评估见 [`chineseeval`](../../ai/examples/chineseeval/README.md) |
| 研究条目追溯 | [`ai/release/traceability.json`](../../ai/release/traceability.json)（研究 T/C/H/V/S 条目 → E/P/D 场景 → 证据用例） |
| 决策 | [ADR-0001…0024](../adr/)；门禁本身见 ADR-0017 |

## 发布门禁

在仓库根目录执行（需要 Node.js 与 `npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node` 安装的冻结 pi 副本）：

```sh
go run ./ai/release/cmd/releasegate -evaluation <独立中文效果包> [-live <live 证据包目录>]...
```

它依次执行 `go vet ./...`（含与不含 `live` tag）、离线 E2E 与 `go test -race ./...`；两次测试均打开 `BARNESS_AI_PIDIFF=1` 和 `BARNESS_AI_PRESSURE=1`。证据写入 `.evidence/barness-ai-release/<时间>/`；全部门禁通过才退出 0：

| 门禁 | 通过条件 |
| --- | --- |
| go-test / go-test-race / go-vet | 命令成功；未执行即失败 |
| p0-offline | 每个用例 PASS（或明确 UNSUPPORTED），完整 artifacts 哈希清单成立；E01–E11、P01–P09、D1、D2 与压力场景各至少一个 PASS |
| differential | 全部 PASS、0 pending；固定 pi 1.0.0 的版本/commit/model-data hash、SDK、请求/帧哈希及发现清单有效；P01–P04/P06/P07 有差分，DeepSeek Responses 与两条原生图像有独立扩展登记及 skip |
| live | 九组合自身报告与矩阵当前完整运行的能力、操作、型号、SDK、时间及正式目录/价格哈希一致，捕获/结果/Observer/图片校验完整；必交分类、生成与编辑不可 UNSUPPORTED |
| redaction-audit | 离线、race、九条 live 与中文效果包均审计，逐包与运行时 audit.json 零发现；拒绝证据不重新输出敏感字段 |
| traceability | 每个研究条目由实际证据判定为 PASS；只有映射没有证据的条目为 NO_EVIDENCE |
| snapshots | 目录/价格快照与代码一致，固定发布合同不因删追溯或路线而缩减 |
| required-artifacts | 本地/云端图像混合设计负载报告、并发/内存/75% 限额余量/资源释放有效；独立中文真实 COMPLETE 报告通过完整性核验，不设置准确率阈值 |

报告 `release-report.md`/`.json` 分别列出离线、差分与 live 结果，并附追溯表、审计结论、命令耗时；证据包另含目录快照、账本、支持矩阵、追溯表与 fixture 索引的副本。`-bundle <dir>` 只评估已有离线证据包（命令门禁记为未执行）。

### 当前九组合结论（2026-10-08）

工单 17 的完整门禁十项全部 PASS，受测提交 `0dca06c`：普通与 race 各 4,448 个
用例（离线 3,825、pi 1.0.0 差分 623，pending 0），两种 vet、九组合自身真实证据、
76 条需求追溯、正式目录/价格快照、混合设计负载与独立中文效果完整性均通过；
十二个完整包逐包审计零发现。[可重放完整证据与核验](../../.scratch/barness-ai-pi-1.0/nine-route-release-evidence/README.md)
保留原始请求/响应、消费结果、Observer、图片、资源读数和完整文件哈希。
没有调整负载或预算，中文合成任务的 24/24 结果不作为生产准确率保证。

### 后续压力基础设施修复（2026-10-08）

[工单 18](../../.scratch/barness-ai-pi-1.0/issues/18-pressure-loopback-interruption.md) 已解决，
交付提交 `2256eaf`。带关联诊断的自然失败将 15 路客户端 `unexpected EOF` 对到测试
Provider 的 TCP 写入 `ENOBUFS`，确认该轮由回环写缓冲耗尽后截断 HTTP 流引起。
测试连接现仅续写未发送字节，每次 Write 最多等待 1 秒，其他错误、持续耗尽、关闭或
取消仍失败；没有重放 HTTP 请求或 AI Attempt。另移除已完成证据对 `testing.T` 的长期
持有，释放跨轮请求捕获。原压力负载、预算和失败断言保持不变。

先红后绿的连接回归 7 PASS；同进程 100 轮、独立进程 10 轮及完整普通/race 各
4,461 用例（含 pi 差分 623）通过，两种 vet 通过。14 份归档、43 个 E2E 包的哈希、
源码匹配及脱敏审计通过，失败历史完整保留，见[压力回环证据与核验](../../.scratch/barness-ai-pi-1.0/pressure-loopback-evidence/README.md)。
旧工单 17 的九路错误无法逐路追溯具体 errno；此修复没有重新运行真实厂商调用，
上节九组合发布结论仍对应其原受测提交。

### 历史状态（2026-10-02/03）

门禁**通过**（2026-10-03 最近一次为 `.evidence/barness-ai-release/gate6`，DeepSeek 两组合在工单 33 的改动后重跑了真实冒烟；九道门禁全部 PASS）：离线 2374 个用例、差分 432 个用例（无待处理）、`-race`、vet、六组合 live、8 个证据包审计零发现、全部追溯条目、目录快照。六组合的真实冒烟用账户别名 `prod-*` 的账户运行（spec 要求独立低权限测试账户，本次由维护者决定使用生产 key）。

运行中出现过两次一次性失败，均已处理：Gemini 差分用例在全量负载下 live partial 的序列化时机不同（pi 的视图领先；已为 Gemini 与 Chat Completions 登记 live partial 扩展的这一方向，见差异登记）；本地压力场景一次有两个调用在本机回环连接上出现 transport 错误，单独与全量重跑均未复现，压力场景现记录底层网络错误以便再现时诊断。

## 资源策略示例与数值依据

数值是部署策略，不构成 pi 兼容承诺，也不是经验证的通用默认值（D1，ADR-0002）。每个示例的每个数值在源码注释中写明适用负载、依据与调整方法；准入数值由 `E08-admission-burst-pressure-*` 验证（工单 18），字节、队列与并发数值由 `E08-policy-pressure-*` 在示例声明的设计负载下验证（工单 24，`BARNESS_AI_PRESSURE=1`）：并发占满，每个调用发送设计大小的请求（长历史加图片），经 Responses（终态帧重复整段响应，最接近帧与总输出上限）流式返回设计大小的单轮输出或工具调用；随后一个停止读取的消费者必须以 `resource_limit` 结束。设计负载对每项字节与队列限额的用量不得超过 75%。

2026-10-02 两次运行测得（Apple M 系列；用量占比两次相同；堆为整个进程的可达堆增长，含测试 Provider 替身保存的请求副本，随 GC 时机在所列范围内变化）：

| 策略 | 设计负载 | 帧 | 单轮输出 | 请求 | 图片 | 工具 JSON | 可达堆增长 / 预算 | 停止读取后的积压 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| LocalPolicy | 8 并发；4 MiB 历史 + 3×2 MiB 截图；128K token 或 512 KiB 工具调用 | 13% | 44% | 38% | 40% | 50% | 0.4–0.5 / 1 GiB | 约 11 分钟（65536 事件，100 token/s） |
| CloudInteractivePolicy | 8 租户 × 4 = 32 并发；2 MiB 历史 + 2×1 MiB 图片；64K token 或 256 KiB 工具调用 | 13% | 44% | 58% | 25% | 50% | 0.7–1.1 / 2 GiB | 约 2.7 分钟（16384 事件） |
| CloudBatchPolicy | 2 租户 × 16 = 32 并发；同上请求；128K token 或 256 KiB 工具调用 | 25% | 44% | 58% | 25% | 50% | 0.9–1.2 / 2 GiB | 约 2.7 分钟 |

压力场景据实修正了三处示例数值，原因写在注释中：

- `MaxToolJSONBytes` 曾因流式工具参数在每个 delta 上全量重解析（与 pi 相同，512 KiB 单调用 23 s）收紧为云端 128 KiB、本地 256 KiB；工单 32 改为读取视图时才解析后已恢复为云端 512 KiB、本地 1 MiB，数值按内存预算确定。`E08-policy-pressure-tool-arguments-scaling` 测得单调用耗时与参数大小线性（512 KiB 约 50 ms、2 MiB 约 0.2 s）。
- `CloudBatchPolicy.MaxOutputBytes` 32 MiB → 64 MiB：128K token 单轮经 Responses 约 29 MB，占原值 88%，不足以容纳推理摘要与帧信封差异；流式响应体只读不留，加倍几乎不增内存。
- 云端注释原称请求体最坏占用 `(并发 + 等待者) × 8 MiB`；实测一次尝试期间约持有请求体的两份副本，已改为两倍。

另测得处理吞吐：16–30 万 token/s（全部并发合计），远高于厂商生成速度。

工单 15 新增独立混合策略及 `E08-mixed-pressure-{local,cloud}-design-load`：本地 4 并发、
云端 2 租户 × 4 = 8 并发，缓冲内存预算分别为 256/512 MiB。每租户同时运行 chat、
OpenAI image、Google image 和 TypeSafe；chat 为 128 KiB 历史/16 KiB 输出，图像调用各有
两张 256 KiB 参考图，OpenAI 另有 256 KiB mask，各输出两张 512 KiB 图片；classifier 为
128 KiB state 与八个约 4 KiB 问题。每张图片 base64 膨胀为 4/3，合法 unary 响应整体
大于 SSE 帧限额；请求 4 MiB、响应 8 MiB、单输入/输出图 1 MiB、总输出图 2 MiB。

证据记录 GC 可达堆增长、峰值许可、Provider 请求副本/共享响应脚本、吞吐与每项限额
占比；占比须 <=75%，可达增长须在预算内。脚本生成在测量基线前，两个强制 GC 分别
覆盖客户端 Body.Read 已交回全部字节但等待 EOF 和全部结果保留的阶段。live 堆只表示已捕获
的存活增长下界；另将调用期 TotalAlloc 累计分配作为保守上界纳入相同预算，包含 GC
采样间的全部暂存副本及测试/Observer 分配，避免漏掉解析峰值；这些数字
是合成协议负载依据，不涵盖厂商像素计算。大 body/result 不额外复制进证据，保留生成
参数、尺寸和哈希供重放。已有 `BARNESS_AI_PRESSURE=1` 显式开关/门禁启用，未开为 NOT_RUN。
可重复报告见 [混合操作证据](../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/README.md)。

## 支持矩阵

| 组合 | 操作 | 协议 | 状态 |
| --- | --- | --- | --- |
| openai-responses | chat | OpenAI × Responses（P01），gpt-5-mini | PASS |
| anthropic-messages | chat | Anthropic × Messages（P02），claude-haiku-4-5 | PASS |
| google-gemini | chat | Google × Gemini Developer API（P03），gemini-3.8-flash | PASS |
| openai-chat | chat | OpenAI × Chat Completions（P04），gpt-5-mini | PASS（reasoning-history UNSUPPORTED：该协议不返回可回放的推理） |
| deepseek-responses | chat | DeepSeek × Responses（P05，扩展路径），deepseek-flash | PASS |
| deepseek-chat | chat | DeepSeek × Chat Completions（P06），deepseek-flash / deepseek-v4-pro | PASS |
| typesafe-classifier | classifier | TypeSafe × System One（P07），jev-1.13.0 | PASS（有界上下文探针返回 400；422 形状未确认） |
| openai-images | image | OpenAI × Images（P08），gpt-image-2.5-sunburst-2026-09-08 | PASS（生成、JSON 编辑与 mask） |
| google-interactions-image | image | Google × Interactions v1beta（P09），gemini-nano-banana-2.1 | PASS（生成与参考编辑，一参考/一输出、1K/1:1） |

2026-10-08 九组合均以自己的独立真实进程重跑并合并完整报告，具体能力、型号与时间见矩阵。
账户别名为 `local-*@region-permissions-unconfirmed`，如实保留本次未确认的区域与权限属性。
完整捕获、消费结果、Observer、图像校验及独立中文效果见 [九组合发布证据](../../.scratch/barness-ai-pi-1.0/nine-route-release-evidence/README.md)。

在矩阵对应行完整通过之前，不得宣称该组合受支持（spec I10）。
