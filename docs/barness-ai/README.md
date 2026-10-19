# barness-ai 发布说明

> **当前完整验收（2026-10-19）**：目录`2026-10-18.1-minimax`包含96聊天身份、2图像和3分类；16协议组合、17自身支持和4项独立效果通过。严格来源资格11/11；来源`78843eb`完整普通/race主包各12495PASS、audit0，双vet与CLI均exit0、发布报告true，精确矩阵已发布。M3.1线上三个候选仍NOT_RUN0/0。后续文档/证据提交保持1078份选定源码。当前结论见[交付状态](../../.scratch/barness-ai-minimax/delivery-status.md)和[完整证据及最终审查](../../.scratch/barness-ai-minimax/evidence/ticket-21/final-closure/README.md)。以下保留各次历史结果。

> **历史GLM完整验收（2026-10-18）**：目录 `2026-10-17.2` / Hash `sha256:81d6d9f3b69e9231985967e1ec8e13640aa7f4fdfd9e7ccef72f7d05320a6942`，93聊天身份、2图像、3分类；中国标准GLM-5.3/5.2三协议六身份、十三协议组合/十四自身支持和四独立效果全部通过。实际完整CLI在3d3cd0f执行 normal11049/race11049、两种vet、压力/差分/撤权及20份审计，十门禁PASS。[票13结论及精确来源](../../.scratch/barness-ai-glm/evidence/ticket-13/README.md)、[可重复校验索引](../../.scratch/barness-ai-glm/evidence/ticket-13/index.json)保留原失败、282次本票实际HTTP、完整基线回滚与99项/643用例追溯。以下历次验收保留各自历史身份。

barness-ai 是 barness 连接多种模型 Provider 的模型协议模块（Go 包 `github.com/tokenbeat-lab/barness/ai`），支持 chat、image、classifier 三种操作。本目录汇总发布交付物、当前验证结论与历史证据。

使用方式、三种模型操作、绑定与资源策略见包级 README：[English](../../ai/README.md) / [简体中文](../../ai/README.zh-CN.md)。

| 交付物 | 位置 |
| --- | --- |
| 公共契约、错误分类与装配说明 | [contract.md](contract.md) |
| Decisions 分类、拒答、图片授权与撤销 | [decisions.md](decisions.md)；[Jev/Luna 迁移示例](../../ai/examples/classifiermigration/README.md) |
| 思考档位采样与 full/simple 边界 | [thinking-sampling.md](thinking-sampling.md) |
| 差异登记（含 D1/D2 与 DeepSeek Responses 扩展） | [differences.md](differences.md)；逐条账本 [`ai/e2e/testdata/pidiff/ledger.json`](../../ai/e2e/testdata/pidiff/ledger.json) |
| 支持矩阵 | [`ai/live/support-matrix.json`](../../ai/live/support-matrix.json)（只经 `go run ./ai/live/cmd/supportmatrix` 合并冒烟报告更新） |
| 模型/价格快照 | [`ai/release/catalog-snapshot.json`](../../ai/release/catalog-snapshot.json)（内置目录及其哈希；门禁核对它与代码一致） |
| Fixture | [`ai/e2e/testdata/`](../../ai/e2e/testdata/)；每次门禁在发布证据包中写出 `fixtures.json`（路径、大小、SHA-256） |
| 资源策略示例 | 聊天：`localassembly.LocalPolicy`、`hostintegration.CloudInteractivePolicy`、`CloudBatchPolicy`；混合操作：`localassembly.MixedPolicy`、`hostintegration.CloudMixedPolicy` |
| 最小示例 | `ai/examples/localassembly`、`hostintegration`（聊天与混合操作）、`toolloop`、`requestid`；独立中文评估见 [`chineseeval`](../../ai/examples/chineseeval/README.md) |
| 研究条目追溯 | [`ai/release/traceability.json`](../../ai/release/traceability.json)（研究 T/C/H/V/S 条目 → E/P/D 场景 → 证据用例） |
| 决策 | [ADR-0001…0047](../adr/)；门禁本身见 ADR-0017 |

## 历史完整验收与交付状态（2026-10-13）

正式目录为 `2026-10-13.1`，包含 87 个聊天身份、两个图像型号与两个分类型号；Hash 为
`sha256:faae843ac61fea6e3ae59bb75165dc97fcf6aa409dabfe0a2f7c9dccbba12029`。
实现使用唯一核验的 pi-ai 1.1.0 基线。Luna 已纳入正式目录，支持单选、评分、是非、逐题拒答与
显式授权的内联 PNG/JPEG/WebP；采样配置和 `durationMs` 的公共行为见上表对应契约。

[该次历史完整自动门禁报告](../../.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/release-report.md)
受测源码 `1291d10`，实际退出 0，十项 gate 全部 PASS：

| 核验 | 实际结果 |
| --- | --- |
| 完整普通 / race | 各 7,620 PASS、0 FAIL，含 861 固定 pi 差分；pending 0 |
| 两种 vet、快照与追溯 | PASS；正式目录/价格快照与固定十路线合同一致 |
| 真实冒烟 | 十组合自身 schema 3 报告符合当前合同；沿用各自原始报告的适用性核验 |
| 容量与授权撤销 | 普通/race 均验证原混合负载、分类图片负载与六种撤销；资源归零 |
| 独立中文效果 | Jev/Luna 本轮分别新采集，均 COMPLETE、24 Calls / 24 Attempts、零重试、24/24 正确 |
| 脱敏审计 | 普通、race、十个 live、两份效果包，共 14 包全部零发现 |

以下旧工单及交付状态保留该次历史记录；当前 MiniMax 完整门禁结论及此前 GLM 结论见本文开头，父 spec 的最终审查与关闭另行记录。

自动门禁 PASS 与最终交付状态分别记录。[历史交付 verdict](../../.scratch/barness-ai-openai-decisions/release-evidence/delivery-verdict.json)
仍为 `FAIL / completed=false`，[工单 13](../../.scratch/barness-ai-openai-decisions/issues/13-ten-route-release-gate.md)
仍未完成，ADR-0025/0026 保持 proposed。不安全的 socket ENOBUFS 续写已删除并通过先红后绿回归，
但原自然压力失败尚无首个偏离字节/帧证据，[缺陷 07](../../.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)
继续开放，后续门禁通过不解除这一因果证据要求。

[归档清单](../../.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/manifest.json)
与[核验收据](../../.scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/verification.json)
保留 22 个完整包、30,619 条 case 记录及全部历史失败。小样本效果不保证生产准确率，估价不等于账单。
不调用厂商的完整性核验命令：

```sh
python3 .scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/verify.py
```

此前十组合门禁（源码 `807ccba`、普通/race 各 7,616 PASS）和原最终拒绝决定，
继续保留在[工单 13 发布证据](../../.scratch/barness-ai-openai-decisions/release-evidence/README.md)。

## 发布门禁

在仓库根目录执行（需要 Node.js 与 `npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node` 安装的冻结 pi 副本）：

```sh
go run ./ai/release/cmd/releasegate -stage check -out <check-dir> -evaluation-jev <Jev-效果包> -evaluation-luna <Luna-效果包> -visual-evaluation-jev <Jev-视觉包> -visual-evaluation-decisions <Decisions-视觉包> [-live <live 证据包目录>]...
# 普通验证结束后，由人工明确选择：
go run ./ai/release/cmd/releasegate -stage race -from <check-dir> -out <race-dir>
```

默认执行 check：先核对固定映射、账本、快照、17份自身支持和四项业务效果的当前来源、完整性及审计；通过后顺序执行两种 vet、完整普通测试与原生证据消费。失败停止后续启动，全部普通验证通过则输出 `READY_FOR_RACE`、race `NOT_RUN`、发布false，退出2。`-collect-all`仅继续收集check阶段诊断，不启动race；独立`-preflight`不执行验证命令，退出0只表示外部输入就绪。

人工选择 race 后，`-from`必须指向原check目录，不能覆盖输入。工具验证原实际命令回执、来源、测试、工具身份和选定输入，复核普通原件后才运行完整race；来源或输入变化拒绝沿用。两次测试均打开`BARNESS_AI_PIDIFF=1`和`BARNESS_AI_PRESSURE=1`，分别保留原生bundle。check与race输出目录须独立且为空；命令保存执行报告、数字退出、起止时间、事件元数据及有界失败尾部。实际失败退出1，十项门禁全部通过才退出0。旧`-bundle`/`-race-bundle`入口只读重评估，命令保持NOT_RUN，不能替代认证。项目执行规则见[验证策略](../agents/verification.md)。

| 门禁 | 通过条件 |
| --- | --- |
| go-test / go-test-race / go-vet | 命令成功；未执行即失败 |
| p0-offline | 每个用例 PASS（或明确 UNSUPPORTED），完整 artifacts 哈希清单成立；E01–E11、P01–P10、D1、D2 与压力场景各至少一个 PASS |
| differential | 全部 PASS、0 pending；固定当前唯一 pi 1.1.0 的版本/commit/model-data hash、SDK、请求/帧哈希及发现清单有效；P01–P04/P06/P07/P10 有差分，DeepSeek Responses 与两条原生图像有独立扩展登记及 skip |
| live | 十三协议组合、十四自身支持身份的报告与矩阵当前完整运行的能力、操作、型号、SDK、时间及正式目录/价格哈希一致，捕获/结果/Observer/图片校验完整；必交分类、生成与编辑不可 UNSUPPORTED |
| redaction-audit | 离线、race、十四份 live 与两份中文/两份视觉效果包均审计，逐包与运行时 audit.json 零发现；拒绝证据不重新输出敏感字段 |
| traceability | 每个研究条目由实际证据判定为 PASS；只有映射没有证据的条目为 NO_EVIDENCE |
| snapshots | 目录/价格快照与代码一致，固定发布合同不因删追溯或路线而缩减 |
| required-artifacts | normal/race 各自的旧混合与分类图片设计负载、并发/内存/限额占用率 <=75%（余量 >=25%）/资源释放及六个授权撤销场景有效；Jev/Luna 中文与 Jev/Decisions 视觉四份独立当前真实报告通过完整重放与来源核验，不设置准确率阈值 |

报告 `release-report.md`/`.json` 分别列出离线、差分与 live 结果，并附追溯表、审计结论、命令耗时；证据包另含目录快照、账本、支持矩阵、追溯表与 fixture 索引的副本。`-bundle <dir>` 只评估已有离线证据包（命令门禁记为未执行），`-race-bundle <dir>` 可同时重放 race 证据。

### 历史九组合结论（2026-10-08）

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
该次交付曾按返回字节数续写 ENOBUFS；2026-10-13 的压力完整性调查确认这条测试路径
不安全并删除：Darwin 可先把数据入队再报错，续写会重复字节。现在原样返回 socket 错误，
保留连接、写入边界、调用身份与最早偏离的字节/帧诊断；不重放 HTTP 请求或 AI Attempt。
该确定性缺陷与历史自然失败的归因分别记录在[工单 07](../../.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md)。另移除已完成证据对 `testing.T` 的长期
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
是合成协议负载依据，不涵盖厂商像素计算。测量结束后才写出请求与完整结果，生成脚本保留
参数、尺寸和哈希供重放。已有 `BARNESS_AI_PRESSURE=1` 显式开关/门禁启用，未开为 NOT_RUN。
可重复报告见 [混合操作证据](../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/README.md)。

Decisions 工单 08 保留上述全部调用，增加
`E08-mixed-pressure-{local,cloud}-classifier-images-design-load`。新增每批每租户四个
Luna 图片分类调用，各有 128 KiB 状态、八个 4000 字节问题及两张 256 KiB PNG；
成功、完整读取处取消及同负载恢复三批分别占满 4/8 并发，沿用 256/512 MiB 预算。
Classifier.Images 限制四张、总 decoded 1 MiB，单图 1 MiB、完整请求 4 MiB；
设计占用率 count 50%、单图 25%、总图 50%，完整 JSON 含 base64/data URL 成本。
占用率 <=75% 表示剩余余量 >=25%。实际 artifact 导入同时要求旧/新报告并从请求捕获
核对字段、身份与容量，缺失或矛盾拒绝。该合成压力证明不替代 Luna 自己的 live。
[新容量证据与重放](../../.scratch/barness-ai-openai-decisions/classifier-load-evidence/README.md)。


## 当前支持矩阵

| 身份 | 操作 | Provider / API | 型号 | 状态 |
| --- | --- | --- | --- | --- |
| openai-responses | chat | openai / openai-responses | gpt-5-mini | PASS |
| anthropic-messages | chat | anthropic / anthropic-messages | claude-haiku-4-5 | PASS |
| google-gemini | chat | google / google-generative-ai | gemini-3.8-flash | PASS |
| openai-chat | chat | openai / openai-completions | gpt-5-mini | PASS |
| deepseek-responses | chat | deepseek / openai-responses | deepseek-flash | PASS |
| deepseek-chat | chat | deepseek / openai-completions | deepseek-flash | PASS |
| typesafe-classifier | classifier | typesafe / typesafe-system-one | jev-1.13.0 | PASS |
| openai-images | image | openai / openai-images | gpt-image-2.5-sunburst-2026-09-08 | PASS |
| google-interactions-image | image | google / google-interactions | gemini-nano-banana-2.1 | PASS |
| openai-decisions | classifier | openai / openai-decisions | gpt-6-luna | PASS |
| typesafe-classifier-jev-1.14.0 | classifier | typesafe / typesafe-system-one | jev-1.14.0 | PASS |
| zhipu-chat | chat | zhipu / openai-completions | glm-5.2 / glm-5.3 | PASS |
| zhipu-responses | chat | zhipu / openai-responses | glm-5.2 / glm-5.3 | PASS |
| zhipu-messages | chat | zhipu / anthropic-messages | glm-5.2 / glm-5.3 | PASS |

以上十四身份分别消费自身真实报告，当前完整 CLI 验收及来源适用范围见[票13校验索引](../../.scratch/barness-ai-glm/evidence/ticket-13/index.json)。三份 GLM 报告各覆盖两个型号；表中其余型号为对应矩阵主字段，工具换型等能力见原报告。账户/区域/权限仍未独立核实，固定任务效果不作为生产准确率承诺。

## 历史十组合支持矩阵（2026-10-13）

| 组合 | 操作 | 协议 | 状态 |
| --- | --- | --- | --- |
| openai-responses | chat | OpenAI × Responses（P01），gpt-5-mini | PASS |
| anthropic-messages | chat | Anthropic × Messages（P02），claude-haiku-4-5 | PASS |
| google-gemini | chat | Google × Gemini Developer API（P03），gemini-3.8-flash | PASS |
| openai-chat | chat | OpenAI × Chat Completions（P04），gpt-5-mini | PASS（reasoning-history UNSUPPORTED：该协议不返回可回放的推理） |
| deepseek-responses | chat | DeepSeek × Responses（P05，扩展路径），deepseek-flash | PASS |
| deepseek-chat | chat | DeepSeek × Chat Completions（P06），deepseek-flash | PASS |
| typesafe-classifier | classifier | TypeSafe × System One（P07），jev-1.13.0 | PASS（有界上下文探针返回 400；422 形状未确认） |
| openai-images | image | OpenAI × Images（P08），gpt-image-2.5-sunburst-2026-09-08 | PASS（生成、JSON 编辑与 mask） |
| google-interactions-image | image | Google × Interactions v1beta（P09），gemini-nano-banana-2.1 | PASS（生成与参考编辑，一参考/一输出、1K/1:1） |
| openai-decisions | classifier | OpenAI × Decisions（P10），gpt-6-luna | PASS（文本单选、评分、是非、拒答、图片及边界场景） |

2026-10-13 十组合均以自己的独立真实进程重跑，并按正式目录合并 schema 3 完整报告，
具体能力、受测型号与时间见 [`support-matrix.json`](../../ai/live/support-matrix.json)。
原九组合账户别名为 `local-*@region-permissions-unconfirmed`，Decisions 为 `live-unverified-account`，
如实保留账户、区域与权限属性未独立核实的边界。完整原始报告见
[十组合证据](../../.scratch/barness-ai-openai-decisions/release-evidence/README.md)。

在矩阵对应行完整通过之前，不得宣称该组合受支持（spec I10）。

## 历史采样实施候选（2026-10-11）

[分档采样配置与 Host 示例](thinking-sampling.md)使用唯一来源核验的 pi 1.1.0 oracle。当时六子集 814 条差分、pending 0，完整正式目录核对通过；采样的公共矩阵、授权、快照、重试和并发已验证。目录版本 2026-10-11.1，包含五处上游价格分层，没有编造采样默认。

当时真实路线与独立中文评估未全部通过；一次普通压力失败另记未确认原因的缺陷，完整发布未通过；[原候选证据](../../.scratch/barness-ai-thinking-sampling/evidence.md)保留全部成功/失败结果。该次2026-10-13历史自动门禁结论见本文历史章节；当前GLM完整结论见本文开头；历史 PASS/FAIL 仍只属于各自原受测源码与目录 hash。

## 统一图片 Host 示例（工单 10，受控验证）

[可执行示例](../../ai/examples/classifierimages/README.md)由
`go run ./ai/examples/classifierimages` 重放同一公共 PNG/JPEG 输入的 Jev/Decisions 调用、
局部图片拒绝、最终图片身份/视觉图例、Box 与独立存在性、Hook 换图和撤销。
它仅通过 Host 候选目录选择 `jev-1.14.0`，不改变正式目录、支持矩阵、发布判定或 proposed ADR。
默认无需真实凭据，完整输入/请求/响应/结果/错误/Attempt/Observer/图片 hash/断言与审计
由 manifest 封存。真实模式显式 opt-in 且要求 Host 配置和唯一命名 key 来源；真实服务协议、
视觉质量及整版发布仍独立验收。[验证与重放](../../.scratch/jev-1.14-multimodal-box/10-verification.md)。
